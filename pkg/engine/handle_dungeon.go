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
		session.State = StateDungeonFloorMap
		return e.handleDungeonFloorMap(session, GameCommand{Type: "init"})
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

	session.State = StateDungeonFloorMap
	return GameResponse{
		Type:     "menu",
		Messages: msgs,
		State: &StateData{
			Screen:  "dungeon_floor_map",
			Player:  MakePlayerState(player),
			Dungeon: makeDungeonView(player.ActiveDungeon),
		},
		Options: dungeonFloorOptions(player.ActiveDungeon),
	}
}

// handleDungeonFloorMap shows the current floor and room options.
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
	}

	floor := &dungeon.Floors[dungeon.CurrentFloor]

	if cmd.Value == "proceed" || cmd.Value == "init" || cmd.Value == "next" {
		// Show current room
		return e.enterDungeonRoom(session)
	}

	// Default: show floor map
	return GameResponse{
		Type:     "menu",
		Messages: []GameMessage{Msg(fmt.Sprintf("Floor %d of %s", floor.FloorNumber, dungeon.Name), "system")},
		State: &StateData{
			Screen:  "dungeon_floor_map",
			Player:  MakePlayerState(player),
			Dungeon: makeDungeonView(dungeon),
		},
		Options: dungeonFloorOptions(dungeon),
	}
}

// enterDungeonRoom processes the current room on the current floor.
func (e *Engine) enterDungeonRoom(session *GameSession) GameResponse {
	player := session.Player
	dungeon := player.ActiveDungeon
	floor := &dungeon.Floors[dungeon.CurrentFloor]
	room := &floor.Rooms[floor.CurrentRoom]

	if room.Cleared {
		// Advance to next room
		floor.CurrentRoom++
		if floor.CurrentRoom >= len(floor.Rooms) {
			// Floor cleared
			floor.Cleared = true
			msgs := []GameMessage{
				Msg(fmt.Sprintf("Floor %d cleared!", floor.FloorNumber), "narrative"),
			}
			if dungeon.CurrentFloor+1 >= len(dungeon.Floors) {
				return e.completeDungeon(session)
			}
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
		room = &floor.Rooms[floor.CurrentRoom]
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
		return e.enterDungeonRoom(session)
	}
}

// startDungeonCombat initiates combat with a dungeon monster.
func (e *Engine) startDungeonCombat(session *GameSession, room *models.DungeonRoom) GameResponse {
	player := session.Player
	dungeon := player.ActiveDungeon

	if room.Monster == nil {
		room.Cleared = true
		return e.enterDungeonRoom(session)
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
		}
		fv.Rooms = make([]DungeonRoomView, len(floor.Rooms))
		for i, room := range floor.Rooms {
			fv.Rooms[i] = DungeonRoomView{
				Type:      room.Type,
				Cleared:   room.Cleared,
				RoomIndex: i,
			}
		}
		dv.Floor = fv
	}

	return dv
}

// dungeonFloorOptions returns standard options for the dungeon floor map.
func dungeonFloorOptions(dungeon *models.Dungeon) []MenuOption {
	return []MenuOption{
		Opt("proceed", "Enter next room"),
		Opt("0", "Leave dungeon"),
	}
}
