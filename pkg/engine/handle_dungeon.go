package engine

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"rpg-game/pkg/game"
	"rpg-game/pkg/models"
)

// handleDungeonSelect shows available dungeons for the player to enter.
func (e *Engine) handleDungeonSelect(session *GameSession, cmd GameCommand) GameResponse {
	player := session.Player

	available := game.AvailableDungeons(player.Level)
	if len(available) == 0 {
		session.State = StateMainMenu
		resp := BuildMainMenuResponse(session)
		resp.Messages = append([]GameMessage{
			Msg("No dungeons available at your level!", "error"),
		}, resp.Messages...)
		return resp
	}

	// Check if player already has an active dungeon
	if player.ActiveDungeon != nil {
		session.State = StateDungeonGridMove
		return e.showDungeonGrid(session, nil)
	}

	idx, err := strconv.Atoi(cmd.Value)
	if err != nil || idx < 1 || idx > len(available) {
		// Show dungeon selection menu
		options := []MenuOption{}
		for i, tmpl := range available {
			label := fmt.Sprintf("%s (Lv%d-%d, %d Floors)", tmpl.Name, tmpl.MinLevel, tmpl.MaxLevel, tmpl.Floors)
			options = append(options, Opt(strconv.Itoa(i+1), label))
		}
		options = append(options, Opt("0", "Return to Main Menu"))

		return GameResponse{
			Type:     "menu",
			Messages: []GameMessage{Msg("Select a dungeon to enter:", "system")},
			State:    &StateData{Screen: "dungeon_select", Player: MakePlayerState(player)},
			Options:  options,
		}
	}

	if cmd.Value == "0" {
		session.State = StateMainMenu
		return BuildMainMenuResponse(session)
	}

	// Enter selected dungeon
	tmpl := available[idx-1]
	seed := time.Now().UnixNano()
	dungeon := game.GenerateDungeon(tmpl, seed)
	player.ActiveDungeon = &dungeon

	msgs := []GameMessage{
		Msg(fmt.Sprintf("Entering %s...", dungeon.Name), "narrative"),
		Msg(fmt.Sprintf("A %d-floor dungeon awaits. Prepare yourself!", len(dungeon.Floors)), "narrative"),
	}

	session.State = StateDungeonGridMove
	return e.showDungeonGrid(session, msgs)
}

// handleDungeonFloorMap handles the floor map state (for room result "proceed" actions).
func (e *Engine) handleDungeonFloorMap(session *GameSession, cmd GameCommand) GameResponse {
	player := session.Player
	dungeon := player.ActiveDungeon

	if dungeon == nil {
		session.State = StateMainMenu
		return BuildMainMenuResponse(session)
	}

	if cmd.Value == "0" || cmd.Value == "leave" {
		// Leave dungeon
		player.ActiveDungeon = nil
		session.State = StateMainMenu
		resp := BuildMainMenuResponse(session)
		resp.Messages = append([]GameMessage{
			Msg("You leave the dungeon.", "narrative"),
		}, resp.Messages...)
		return resp
	}

	if cmd.Value == "next" {
		// Advance to next floor
		dungeon.CurrentFloor++
		if dungeon.CurrentFloor >= len(dungeon.Floors) {
			// Dungeon complete!
			return e.completeDungeon(session)
		}
		// Ensure new floor has grid
		e.ensureFloorGrid(dungeon)
		session.State = StateDungeonGridMove
		msgs := []GameMessage{
			Msg(fmt.Sprintf("Descending to floor %d...", dungeon.Floors[dungeon.CurrentFloor].FloorNumber), "narrative"),
		}
		return e.showDungeonGrid(session, msgs)
	}

	// "proceed" after room completion — return to grid movement
	if cmd.Value == "proceed" || cmd.Value == "init" {
		session.State = StateDungeonGridMove
		return e.showDungeonGrid(session, nil)
	}

	// Default: show grid map
	session.State = StateDungeonGridMove
	return e.showDungeonGrid(session, nil)
}

// handleDungeonMove processes directional movement on the dungeon grid.
func (e *Engine) handleDungeonMove(session *GameSession, cmd GameCommand) GameResponse {
	player := session.Player
	dungeon := player.ActiveDungeon

	if dungeon == nil {
		session.State = StateMainMenu
		return BuildMainMenuResponse(session)
	}

	if cmd.Value == "0" || cmd.Value == "leave" {
		player.ActiveDungeon = nil
		session.State = StateMainMenu
		resp := BuildMainMenuResponse(session)
		resp.Messages = append([]GameMessage{
			Msg("You leave the dungeon.", "narrative"),
		}, resp.Messages...)
		return resp
	}

	e.ensureFloorGrid(dungeon)
	floor := &dungeon.Floors[dungeon.CurrentFloor]

	// Calculate target position from direction
	dx, dy := 0, 0
	switch cmd.Value {
	case "n":
		dy = -1
	case "s":
		dy = 1
	case "w":
		dx = -1
	case "e":
		dx = 1
	default:
		// Unknown command, just show grid
		return e.showDungeonGrid(session, nil)
	}

	targetX := floor.PlayerPos.X + dx
	targetY := floor.PlayerPos.Y + dy

	if !game.CanMoveOnGrid(floor, targetX, targetY) {
		msgs := []GameMessage{Msg("You can't move that way.", "system")}
		return e.showDungeonGrid(session, msgs)
	}

	// Move player
	floor.PlayerPos.X = targetX
	floor.PlayerPos.Y = targetY

	// Reveal tiles within 2-tile radius
	game.RevealRadius(floor, targetX, targetY, 2)

	tile := floor.Grid[targetY][targetX]
	msgs := []GameMessage{}

	// Check if player stepped on exit tile
	if tile.Type == "exit" {
		// Check if enough rooms are cleared to use exit
		clearedCount := 0
		for _, room := range floor.Rooms {
			if room.Cleared {
				clearedCount++
			}
		}
		// Need to clear at least half the rooms (including boss) to descend
		requiredClears := len(floor.Rooms) / 2
		if requiredClears < 1 {
			requiredClears = 1
		}

		if clearedCount >= requiredClears {
			floor.Cleared = true
			if dungeon.CurrentFloor+1 >= len(dungeon.Floors) {
				return e.completeDungeon(session)
			}
			msgs = append(msgs, Msg(fmt.Sprintf("Floor %d cleared! You found the stairs down.", floor.FloorNumber), "narrative"))
			session.State = StateDungeonFloorMap
			return GameResponse{
				Type:     "menu",
				Messages: msgs,
				State: &StateData{
					Screen:  "dungeon_floor_map",
					Player:  MakePlayerState(player),
					Dungeon: makeDungeonView(dungeon),
				},
				Options: []MenuOption{
					Opt("next", "Descend to next floor"),
					Opt("0", "Leave dungeon"),
				},
			}
		}
		msgs = append(msgs, Msg(fmt.Sprintf("You found the exit, but need to clear more rooms first (%d/%d).", clearedCount, requiredClears), "system"))
		return e.showDungeonGrid(session, msgs)
	}

	// Check if target tile is an uncleared room
	if tile.RoomIdx >= 0 && tile.RoomIdx < len(floor.Rooms) {
		room := &floor.Rooms[tile.RoomIdx]
		floor.CurrentRoom = tile.RoomIdx
		if !room.Cleared {
			// Trigger room handler
			return e.enterDungeonRoom(session)
		}
		// Room already cleared, just moved through
		msgs = append(msgs, Msg("This room has already been cleared.", "system"))
	}

	return e.showDungeonGrid(session, msgs)
}

// showDungeonGrid builds the grid view response with directional movement options.
func (e *Engine) showDungeonGrid(session *GameSession, msgs []GameMessage) GameResponse {
	player := session.Player
	dungeon := player.ActiveDungeon

	e.ensureFloorGrid(dungeon)
	floor := &dungeon.Floors[dungeon.CurrentFloor]

	if msgs == nil {
		msgs = []GameMessage{}
	}
	msgs = append(msgs, Msg(fmt.Sprintf("Floor %d of %s", floor.FloorNumber, dungeon.Name), "system"))

	session.State = StateDungeonGridMove
	return GameResponse{
		Type:     "menu",
		Messages: msgs,
		State: &StateData{
			Screen:  "dungeon_floor_map",
			Player:  MakePlayerState(player),
			Dungeon: makeDungeonView(dungeon),
		},
		Options: dungeonMoveOptions(floor),
	}
}

// ensureFloorGrid regenerates the grid if it's nil (backward compat with old saves).
func (e *Engine) ensureFloorGrid(dungeon *models.Dungeon) {
	floor := &dungeon.Floors[dungeon.CurrentFloor]
	if floor.Grid == nil {
		game.RegenerateFloorGrid(floor, dungeon.Seed, dungeon.CurrentFloor)
	}
}

// enterDungeonRoom processes the current room on the current floor.
func (e *Engine) enterDungeonRoom(session *GameSession) GameResponse {
	player := session.Player
	dungeon := player.ActiveDungeon
	floor := &dungeon.Floors[dungeon.CurrentFloor]
	room := &floor.Rooms[floor.CurrentRoom]

	if room.Cleared {
		// Room already cleared, return to grid
		session.State = StateDungeonGridMove
		return e.showDungeonGrid(session, []GameMessage{Msg("This room has already been cleared.", "system")})
	}

	switch room.Type {
	case "combat", "boss":
		return e.startDungeonCombat(session, room)
	case "treasure":
		return e.handleDungeonTreasure(session, room)
	case "trap":
		return e.handleDungeonTrap(session, room)
	case "rest":
		return e.handleDungeonRest(session, room)
	case "merchant":
		return e.handleDungeonMerchant(session, room)
	default:
		room.Cleared = true
		session.State = StateDungeonGridMove
		return e.showDungeonGrid(session, nil)
	}
}

// startDungeonCombat initiates combat with a dungeon monster.
func (e *Engine) startDungeonCombat(session *GameSession, room *models.DungeonRoom) GameResponse {
	player := session.Player
	dungeon := player.ActiveDungeon

	if room.Monster == nil {
		room.Cleared = true
		session.State = StateDungeonGridMove
		return e.showDungeonGrid(session, nil)
	}

	mob := *room.Monster

	// Restore mana/stamina at combat start
	player.ManaRemaining = player.ManaTotal
	player.StaminaRemaining = player.StaminaTotal
	mob.ManaRemaining = mob.ManaTotal
	mob.StaminaRemaining = mob.StaminaTotal

	session.Combat = &CombatContext{
		Mob:       mob,
		MobLoc:    -1,
		Turn:      0,
		IsDungeon: true,
	}

	session.State = StateCombat

	bossTag := ""
	if mob.IsBoss {
		bossTag = " [DUNGEON BOSS]"
	}
	rarityTag := ""
	rarityName := game.RarityDisplayName(mob.Rarity)
	if rarityName != "Common" {
		rarityTag = fmt.Sprintf(" [%s]", rarityName)
	}

	floor := &dungeon.Floors[dungeon.CurrentFloor]
	msgs := []GameMessage{
		Msg(fmt.Sprintf("--- %s Floor %d, Room %d/%d ---", dungeon.Name, floor.FloorNumber, floor.CurrentRoom+1, len(floor.Rooms)), "system"),
		Msg(fmt.Sprintf("Lv%d %s vs Lv%d %s (%s)%s%s",
			player.Level, player.Name,
			mob.Level, mob.Name, mob.MonsterType, rarityTag, bossTag), "combat"),
	}

	return GameResponse{
		Type:     "combat",
		Messages: msgs,
		State: &StateData{
			Screen:  "combat",
			Player:  MakePlayerState(player),
			Combat:  MakeCombatView(session),
			Dungeon: makeDungeonView(dungeon),
		},
		Options: combatActionOptions(),
	}
}

// handleDungeonTreasure processes a treasure room.
func (e *Engine) handleDungeonTreasure(session *GameSession, room *models.DungeonRoom) GameResponse {
	player := session.Player
	dungeon := player.ActiveDungeon

	msgs := []GameMessage{
		Msg("You found a treasure room!", "narrative"),
	}

	for _, item := range room.Loot {
		game.EquipBestItem(item, &player.EquipmentMap, &player.Inventory)
		msgs = append(msgs, Msg(fmt.Sprintf("Found: %s (Rarity %d, CP:%d)", item.Name, item.Rarity, item.CP), "loot"))
	}

	room.Cleared = true

	session.State = StateDungeonFloorMap
	return GameResponse{
		Type:     "menu",
		Messages: msgs,
		State: &StateData{
			Screen:  "dungeon_room",
			Player:  MakePlayerState(player),
			Dungeon: makeDungeonView(dungeon),
		},
		Options: []MenuOption{Opt("proceed", "Continue")},
	}
}

// handleDungeonTrap processes a trap room.
func (e *Engine) handleDungeonTrap(session *GameSession, room *models.DungeonRoom) GameResponse {
	player := session.Player
	dungeon := player.ActiveDungeon

	// 50% chance to dodge
	dodged := rand.Intn(100) < 50
	msgs := []GameMessage{}

	if dodged {
		msgs = append(msgs, Msg("A trap springs! You dodge it nimbly.", "narrative"))
	} else {
		player.HitpointsRemaining -= room.TrapDamage
		msgs = append(msgs, Msg(fmt.Sprintf("A trap springs! You take %d damage!", room.TrapDamage), "damage"))
		if player.HitpointsRemaining <= 0 {
			return e.handleDungeonDefeat(session, msgs)
		}
	}

	room.Cleared = true

	session.State = StateDungeonFloorMap
	return GameResponse{
		Type:     "menu",
		Messages: msgs,
		State: &StateData{
			Screen:  "dungeon_room",
			Player:  MakePlayerState(player),
			Dungeon: makeDungeonView(dungeon),
		},
		Options: []MenuOption{Opt("proceed", "Continue")},
	}
}

// handleDungeonRest processes a rest room.
func (e *Engine) handleDungeonRest(session *GameSession, room *models.DungeonRoom) GameResponse {
	player := session.Player
	dungeon := player.ActiveDungeon

	healAmount := room.HealAmount
	player.HitpointsRemaining += healAmount
	if player.HitpointsRemaining > player.HitpointsTotal {
		player.HitpointsRemaining = player.HitpointsTotal
	}
	// Restore some mana and stamina
	player.ManaRemaining += healAmount / 2
	if player.ManaRemaining > player.ManaTotal {
		player.ManaRemaining = player.ManaTotal
	}
	player.StaminaRemaining += healAmount / 2
	if player.StaminaRemaining > player.StaminaTotal {
		player.StaminaRemaining = player.StaminaTotal
	}

	msgs := []GameMessage{
		Msg("You found a safe resting spot.", "narrative"),
		Msg(fmt.Sprintf("Recovered %d HP, %d MP, %d SP", healAmount, healAmount/2, healAmount/2), "heal"),
	}

	room.Cleared = true

	session.State = StateDungeonFloorMap
	return GameResponse{
		Type:     "menu",
		Messages: msgs,
		State: &StateData{
			Screen:  "dungeon_room",
			Player:  MakePlayerState(player),
			Dungeon: makeDungeonView(dungeon),
		},
		Options: []MenuOption{Opt("proceed", "Continue")},
	}
}

// handleDungeonMerchant processes a merchant room (free items for now).
func (e *Engine) handleDungeonMerchant(session *GameSession, room *models.DungeonRoom) GameResponse {
	player := session.Player
	dungeon := player.ActiveDungeon

	msgs := []GameMessage{
		Msg("A wandering merchant offers their wares!", "narrative"),
	}

	// Give all merchant items to the player
	for _, item := range room.Loot {
		if item.ItemType == "consumable" {
			player.Inventory = append(player.Inventory, item)
			msgs = append(msgs, Msg(fmt.Sprintf("Received: %s", item.Name), "loot"))
		} else {
			game.EquipBestItem(item, &player.EquipmentMap, &player.Inventory)
			msgs = append(msgs, Msg(fmt.Sprintf("Received: %s (CP:%d)", item.Name, item.CP), "loot"))
		}
	}

	room.Cleared = true

	session.State = StateDungeonFloorMap
	return GameResponse{
		Type:     "menu",
		Messages: msgs,
		State: &StateData{
			Screen:  "dungeon_room",
			Player:  MakePlayerState(player),
			Dungeon: makeDungeonView(dungeon),
		},
		Options: []MenuOption{Opt("proceed", "Continue")},
	}
}

// handleDungeonRoom dispatches to the appropriate room handler.
func (e *Engine) handleDungeonRoom(session *GameSession, cmd GameCommand) GameResponse {
	return e.handleDungeonFloorMap(session, cmd)
}

// completeDungeon handles finishing the entire dungeon.
func (e *Engine) completeDungeon(session *GameSession) GameResponse {
	player := session.Player
	dungeon := player.ActiveDungeon

	// Record dungeon clear
	game.RecordDungeonClear(&player.Stats)

	// Bonus XP for completion
	bonusXP := len(dungeon.Floors) * 100
	player.Experience += bonusXP
	game.RecordXPGained(&player.Stats, bonusXP)

	msgs := []GameMessage{
		Msg("========================================", "system"),
		Msg(fmt.Sprintf("DUNGEON COMPLETE: %s!", dungeon.Name), "narrative"),
		Msg(fmt.Sprintf("Cleared all %d floors!", len(dungeon.Floors)), "narrative"),
		Msg(fmt.Sprintf("Bonus XP: +%d", bonusXP), "levelup"),
		Msg("========================================", "system"),
	}

	// Level up check
	prevLevel := player.Level
	game.LevelUp(player)
	if player.Level > prevLevel {
		msgs = append(msgs, Msg(fmt.Sprintf("LEVEL UP! Now level %d!", player.Level), "levelup"))
	}

	player.ActiveDungeon = nil

	session.State = StateMainMenu
	return GameResponse{
		Type:     "narrative",
		Messages: msgs,
		State:    &StateData{Screen: "main_menu", Player: MakePlayerState(player)},
		Options:  BuildMainMenuResponse(session).Options,
	}
}

// handleDungeonDefeat handles the player dying in a dungeon.
func (e *Engine) handleDungeonDefeat(session *GameSession, msgs []GameMessage) GameResponse {
	player := session.Player

	msgs = append(msgs, Msg("========================================", "system"))
	msgs = append(msgs, Msg("DUNGEON DEFEAT!", "combat"))
	msgs = append(msgs, Msg("You keep all XP and loot gained, but lose dungeon progress.", "narrative"))
	msgs = append(msgs, Msg("========================================", "system"))

	game.RecordDeath(&player.Stats)

	// Resurrect
	player.HitpointsRemaining = player.HitpointsTotal
	player.ManaRemaining = player.ManaTotal
	player.StaminaRemaining = player.StaminaTotal
	player.Resurrections++
	player.StatusEffects = []models.StatusEffect{}
	player.ActiveDungeon = nil

	msgs = append(msgs, Msg(fmt.Sprintf("%s has been resurrected. (Resurrection #%d)", player.Name, player.Resurrections), "system"))

	session.State = StateMainMenu
	return GameResponse{
		Type:     "narrative",
		Messages: msgs,
		State:    &StateData{Screen: "main_menu", Player: MakePlayerState(player)},
		Options:  BuildMainMenuResponse(session).Options,
	}
}

// makeDungeonView creates a DungeonView from a dungeon model.
func makeDungeonView(dungeon *models.Dungeon) *DungeonView {
	if dungeon == nil {
		return nil
	}

	dv := &DungeonView{
		Name:         dungeon.Name,
		CurrentFloor: dungeon.CurrentFloor,
		TotalFloors:  len(dungeon.Floors),
	}

	if dungeon.CurrentFloor < len(dungeon.Floors) {
		floor := dungeon.Floors[dungeon.CurrentFloor]
		fv := &DungeonFloorView{
			FloorNumber: floor.FloorNumber,
			CurrentRoom: floor.CurrentRoom,
			TotalRooms:  len(floor.Rooms),
			Cleared:     floor.Cleared,
			BossFloor:   floor.BossFloor,
			GridW:       floor.GridW,
			GridH:       floor.GridH,
			PlayerX:     floor.PlayerPos.X,
			PlayerY:     floor.PlayerPos.Y,
			ExitX:       floor.ExitPos.X,
			ExitY:       floor.ExitPos.Y,
		}
		fv.Rooms = make([]DungeonRoomView, len(floor.Rooms))
		for i, room := range floor.Rooms {
			fv.Rooms[i] = DungeonRoomView{
				Type:      room.Type,
				Cleared:   room.Cleared,
				RoomIndex: i,
			}
		}

		// Build grid view (fog of war: unexplored tiles shown as "fog")
		if floor.Grid != nil {
			fv.Grid = make([][]DungeonTileView, floor.GridH)
			for y := 0; y < floor.GridH; y++ {
				fv.Grid[y] = make([]DungeonTileView, floor.GridW)
				for x := 0; x < floor.GridW; x++ {
					tile := floor.Grid[y][x]
					if tile.Explored {
						tv := DungeonTileView{
							Type:     tile.Type,
							RoomIdx:  tile.RoomIdx,
							Explored: true,
						}
						// Add room info for explored room tiles
						if tile.RoomIdx >= 0 && tile.RoomIdx < len(floor.Rooms) {
							room := floor.Rooms[tile.RoomIdx]
							tv.RoomType = room.Type
							tv.Cleared = room.Cleared
						}
						fv.Grid[y][x] = tv
					} else {
						fv.Grid[y][x] = DungeonTileView{
							Type:     "fog",
							RoomIdx:  -1,
							Explored: false,
						}
					}
				}
			}
		}

		dv.Floor = fv
	}

	return dv
}

// dungeonMoveOptions returns directional movement options based on walkable neighbors.
func dungeonMoveOptions(floor *models.DungeonFloor) []MenuOption {
	opts := []MenuOption{}
	pos := floor.PlayerPos

	if game.CanMoveOnGrid(floor, pos.X, pos.Y-1) {
		opts = append(opts, Opt("n", "North"))
	}
	if game.CanMoveOnGrid(floor, pos.X, pos.Y+1) {
		opts = append(opts, Opt("s", "South"))
	}
	if game.CanMoveOnGrid(floor, pos.X-1, pos.Y) {
		opts = append(opts, Opt("w", "West"))
	}
	if game.CanMoveOnGrid(floor, pos.X+1, pos.Y) {
		opts = append(opts, Opt("e", "East"))
	}
	opts = append(opts, Opt("0", "Leave Dungeon"))

	return opts
}
