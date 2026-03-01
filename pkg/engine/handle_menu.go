package engine

import (
	"fmt"
	"math/rand"
	"strconv"

	"strings"

	"rpg-game/pkg/data"
	"rpg-game/pkg/game"
	"rpg-game/pkg/models"
)

// handleInit processes the initial session state, selecting or creating a character.
func (e *Engine) handleInit(session *GameSession) GameResponse {
	gs := session.GameState

	if len(gs.CharactersMap) == 0 {
		// No characters exist -- create a default "Temp" character
		player := game.GenerateCharacter("Temp", 1, 1)
		player.EquipmentMap = map[int]models.Item{}
		player.Inventory = []models.Item{
			game.CreateHealthPotion("small"),
			game.CreateHealthPotion("small"),
			game.CreateHealthPotion("small"),
		}
		player.ResourceStorageMap = map[string]models.Resource{}
		game.GenerateLocationsForNewCharacter(&player)
		gs.CharactersMap[player.Name] = player
	}

	if len(gs.CharactersMap) == 1 {
		// Auto-select the only character
		for _, char := range gs.CharactersMap {
			c := char
			if c.CompletedQuests == nil {
				c.CompletedQuests = []string{}
			}
			if c.ActiveQuests == nil {
				c.ActiveQuests = []string{"quest_1_training"}
			}
			if c.LockedLocations == nil {
				c.LockedLocations = []string{}
			}
			gs.CharactersMap[c.Name] = c
			session.Player = &c
			break
		}
		// Ensure leaderboard entry exists for this character.
		e.saveSession(session)
		session.State = StateMainMenu
		e.Broadcast(session.ID, GameResponse{
			Type:     "broadcast",
			Messages: []GameMessage{Msg(fmt.Sprintf("%s has entered the game! (Level %d)", session.Player.Name, session.Player.Level), "system")},
			State:    &StateData{Screen: "player_joined"},
		})
		return BuildMainMenuResponse(session)
	}

	// Multiple characters -- let the player choose
	session.State = StateCharacterSelect
	msgs := []GameMessage{
		Msg("Select a character:", "system"),
	}
	options := []MenuOption{}
	for name := range gs.CharactersMap {
		options = append(options, Opt(name, name))
	}

	return GameResponse{
		Type:     "menu",
		Messages: msgs,
		State:    &StateData{Screen: "character_select"},
		Options:  options,
	}
}

// handleCharacterSelect processes the player's character selection from the list.
func (e *Engine) handleCharacterSelect(session *GameSession, cmd GameCommand) GameResponse {
	gs := session.GameState
	charName := cmd.Value

	char, exists := gs.CharactersMap[charName]
	if !exists {
		// Pick first available
		for _, c := range gs.CharactersMap {
			char = c
			break
		}
	}

	if char.CompletedQuests == nil {
		char.CompletedQuests = []string{}
	}
	if char.ActiveQuests == nil {
		char.ActiveQuests = []string{"quest_1_training"}
	}
	if char.LockedLocations == nil {
		char.LockedLocations = []string{}
	}
	gs.CharactersMap[char.Name] = char
	session.Player = &char

	// Ensure leaderboard entry exists for this character.
	e.saveSession(session)
	session.State = StateMainMenu
	e.Broadcast(session.ID, GameResponse{
		Type:     "broadcast",
		Messages: []GameMessage{Msg(fmt.Sprintf("%s has entered the game! (Level %d)", session.Player.Name, session.Player.Level), "system")},
		State:    &StateData{Screen: "player_joined"},
	})
	return BuildMainMenuResponse(session)
}

// handleCharacterCreate creates a new character with the given name.
func (e *Engine) handleCharacterCreate(session *GameSession, cmd GameCommand) GameResponse {
	gs := session.GameState
	name := cmd.Value

	if _, exists := gs.CharactersMap[name]; exists {
		return GameResponse{
			Type:     "menu",
			Messages: []GameMessage{Msg(fmt.Sprintf("Character '%s' already exists!", name), "error")},
			State:    &StateData{Screen: "character_create"},
			Prompt:   "Enter character name: ",
		}
	}

	player := game.GenerateCharacter(name, 1, 1)
	player.EquipmentMap = map[int]models.Item{}
	player.Inventory = []models.Item{
		game.CreateHealthPotion("small"),
		game.CreateHealthPotion("small"),
		game.CreateHealthPotion("small"),
	}
	player.ResourceStorageMap = map[string]models.Resource{}
	player.BuiltBuildings = []models.Building{}
	player.LockedLocations = []string{}
	game.GenerateLocationsForNewCharacter(&player)

	gs.CharactersMap[player.Name] = player
	session.Player = &player

	// Save after creation (uses DB for DB sessions, file for local sessions).
	e.saveSession(session)

	session.State = StateMainMenu
	resp := BuildMainMenuResponse(session)
	resp.Messages = append([]GameMessage{
		Msg(fmt.Sprintf("Character '%s' created!", name), "system"),
	}, resp.Messages...)
	return resp
}

// handleMainMenu processes the main menu selection.
func (e *Engine) handleMainMenu(session *GameSession, cmd GameCommand) GameResponse {
	gs := session.GameState
	player := session.Player

	switch cmd.Value {
	case "0":
		// Character Create
		session.State = StateCharacterCreate
		return GameResponse{
			Type:     "menu",
			Messages: []GameMessage{Msg("Create a new character", "system")},
			State:    &StateData{Screen: "character_create"},
			Prompt:   "Enter character name: ",
		}

	case "1":
		// Harvest
		session.State = StateHarvestSelect
		options := []MenuOption{}
		for i, res := range data.ResourceTypes {
			options = append(options, Opt(res, fmt.Sprintf("%d. %s", i+1, res)))
		}
		return GameResponse{
			Type:     "menu",
			Messages: []GameMessage{Msg("Select a resource to harvest:", "system")},
			State:    &StateData{Screen: "harvest_select", Player: MakePlayerState(player)},
			Options:  options,
		}

	case "2":
		// Locations are now discovered during hunting
		session.State = StateMainMenu
		resp := BuildMainMenuResponse(session)
		resp.Messages = append([]GameMessage{
			Msg("Locations are now discovered while hunting! Check the Hunt menu for locked locations.", "narrative"),
		}, resp.Messages...)
		return resp

	case "3":
		// Hunt -- select location
		session.State = StateHuntLocationSelect
		options := []MenuOption{}
		for _, locName := range player.KnownLocations {
			loc, exists := gs.GameLocations[locName]
			if !exists || loc.Type == "Base" {
				continue
			}
			options = append(options, Opt(locName, fmt.Sprintf("%s (%s, Lv1-%d)", locName, loc.Type, loc.LevelMax)))
		}
		// Show locked locations
		for _, locName := range player.LockedLocations {
			loc, exists := gs.GameLocations[locName]
			if !exists {
				continue
			}
			options = append(options, Opt("locked:"+locName, fmt.Sprintf("[LOCKED] %s (%s, Lv1-%d) - Defeat Guardian to Unlock", locName, loc.Type, loc.LevelMax)))
		}
		if len(options) == 0 {
			session.State = StateMainMenu
			resp := BuildMainMenuResponse(session)
			resp.Messages = append([]GameMessage{
				Msg("No huntable locations available! Keep hunting to discover new locations.", "error"),
			}, resp.Messages...)
			return resp
		}
		return GameResponse{
			Type:     "menu",
			Messages: []GameMessage{Msg("Select a hunting location:", "system")},
			State:    &StateData{Screen: "hunt_location_select", Player: MakePlayerState(player)},
			Options:  options,
		}

	case "4":
		// Discovered Locations
		session.State = StateDiscoveredLocations
		msgs := []GameMessage{
			Msg("=== Discovered Locations ===", "system"),
		}
		for _, locName := range player.KnownLocations {
			loc, exists := gs.GameLocations[locName]
			if !exists {
				msgs = append(msgs, Msg(fmt.Sprintf("%s (unknown)", locName), "narrative"))
				continue
			}
			msgs = append(msgs, Msg(fmt.Sprintf("\n--- %s (%s) ---", loc.Name, loc.Type), "narrative"))
			if loc.Type == "Base" {
				msgs = append(msgs, Msg("  Base location - no monsters", "narrative"))
				continue
			}
			guardianCount := 0
			for _, mob := range loc.Monsters {
				guardianTag := ""
				if mob.IsSkillGuardian {
					guardianTag = fmt.Sprintf(" [GUARDIAN - %s]", mob.GuardedSkill.Name)
					guardianCount++
				}
				msgs = append(msgs, Msg(fmt.Sprintf("  %s Lv%d (HP:%d/%d)%s",
					mob.Name, mob.Level, mob.HitpointsRemaining, mob.HitpointsTotal, guardianTag), "narrative"))
			}
			if guardianCount > 0 {
				msgs = append(msgs, Msg(fmt.Sprintf("  %d Skill Guardian(s) present!", guardianCount), "narrative"))
			}
		}
		// Show locked locations
		if len(player.LockedLocations) > 0 {
			msgs = append(msgs, Msg("", "system"))
			msgs = append(msgs, Msg("=== Locked Locations ===", "system"))
			for _, locName := range player.LockedLocations {
				loc, exists := gs.GameLocations[locName]
				if !exists {
					msgs = append(msgs, Msg(fmt.Sprintf("[LOCKED] %s (unknown)", locName), "narrative"))
					continue
				}
				msgs = append(msgs, Msg(fmt.Sprintf("[LOCKED] %s (%s, Lv1-%d) - Defeat Guardian to Unlock", loc.Name, loc.Type, loc.LevelMax), "narrative"))
			}
		}

		return GameResponse{
			Type:     "menu",
			Messages: msgs,
			State:    &StateData{Screen: "discovered_locations", Player: MakePlayerState(player)},
			Options:  []MenuOption{Opt("back", "Return to Main Menu")},
		}

	case "5":
		// Player Stats
		session.State = StatePlayerStats
		gs.CharactersMap[player.Name] = *player
		game.WriteGameStateToFile(*gs, session.SaveFile)

		msgs := buildPlayerStatsMessages(player, gs)
		return GameResponse{
			Type:     "menu",
			Messages: msgs,
			State:    &StateData{Screen: "player_stats", Player: MakePlayerState(player)},
			Options:  []MenuOption{Opt("back", "Return to Main Menu")},
		}

	case "6":
		// Load Save
		session.State = StateLoadSave
		return e.handleLoadSave(session, cmd)

	case "7":
		// Player Guide
		if e.metrics != nil {
			e.metrics.RecordFeatureUse("guide")
		}
		session.State = StateGuideMain
		return e.handleGuideMain(session, GameCommand{Type: "init"})

	case "8":
		// Auto-Play Speed
		if e.metrics != nil {
			e.metrics.RecordFeatureUse("autoplay")
		}
		session.State = StateAutoPlaySpeed
		return GameResponse{
			Type:     "menu",
			Messages: []GameMessage{Msg("Select auto-play speed:", "system")},
			State:    &StateData{Screen: "autoplay_speed", Player: MakePlayerState(player)},
			Options: []MenuOption{
				Opt("1", "Slow (2s per fight)"),
				Opt("2", "Normal (1s per fight)"),
				Opt("3", "Fast (0.5s per fight)"),
				Opt("4", "Turbo (0.1s per fight)"),
			},
		}

	case "9":
		// Quest Log
		if e.metrics != nil {
			e.metrics.RecordFeatureUse("quests")
		}
		session.State = StateQuestLog
		msgs := buildQuestLogMessages(player, gs)
		return GameResponse{
			Type:     "menu",
			Messages: msgs,
			State:    &StateData{Screen: "quest_log", Player: MakePlayerState(player)},
			Options:  []MenuOption{Opt("back", "Return to Main Menu")},
		}

	case "10":
		// Village Management
		if e.metrics != nil {
			e.metrics.RecordFeatureUse("village")
		}
		if gs.Villages == nil {
			gs.Villages = make(map[string]models.Village)
		}
		// Gate: village requires elder rescue quest unless player already has a village
		_, hasVillage := gs.Villages[player.VillageName]
		if !hasVillage && !game.Contains(player.CompletedQuests, "quest_v0_elder") {
			msgs := []GameMessage{
				Msg("You must first rescue a Village Elder from the Lake Ruins.", "narrative"),
				Msg("Hunt at Lake Ruins to find and rescue a captive elder.", "narrative"),
			}
			if game.Contains(player.ActiveQuests, "quest_v0_elder") {
				msgs = append(msgs, Msg("Quest Active: The Lost Elder", "narrative"))
			} else {
				msgs = append(msgs, Msg("Complete 'The First Trial' quest to unlock this quest.", "narrative"))
			}
			return GameResponse{
				Type:     "menu",
				Messages: msgs,
				State:    &StateData{Screen: "main_menu", Player: MakePlayerState(player)},
				Options:  []MenuOption{Opt("back", "Return to Main Menu")},
			}
		}
		village, exists := gs.Villages[player.VillageName]
		if !exists {
			village = game.GenerateVillage(player.Name)
			player.VillageName = player.Name + "'s Village"
			gs.Villages[player.VillageName] = village
		}
		session.SelectedVillage = &village
		session.State = StateVillageMain
		// Return a "pass-through" that the village handler will process
		return e.handleVillageMain(session, GameCommand{Type: "init", Value: ""})

	case "11":
		// Town
		if e.metrics != nil {
			e.metrics.RecordFeatureUse("town")
		}
		session.State = StateTownMain
		return e.handleTownMain(session, GameCommand{Type: "init"})

	case "12":
		// Dungeon
		if e.metrics != nil {
			e.metrics.RecordFeatureUse("dungeon")
		}
		session.State = StateDungeonSelect
		return e.handleDungeonSelect(session, GameCommand{Type: "init"})

	case "13":
		// Bounty Board
		session.State = StateMostWantedBoard
		return e.handleMostWantedBoard(session, GameCommand{Type: "init"})

	case "14":
		// Arena
		if e.metrics != nil {
			e.metrics.RecordFeatureUse("arena")
		}
		session.State = StateArenaMain
		return e.handleArenaMain(session, GameCommand{Type: "init"})

	case "exit":
		gs.CharactersMap[player.Name] = *player
		game.WriteGameStateToFile(*gs, session.SaveFile)
		return GameResponse{
			Type:     "exit",
			Messages: []GameMessage{Msg("Game saved. Goodbye!", "system")},
		}

	default:
		session.State = StateMainMenu
		return BuildMainMenuResponse(session)
	}
}

// handleHarvestSelect processes the resource harvest action.
func (e *Engine) handleHarvestSelect(session *GameSession, cmd GameCommand) GameResponse {
	player := session.Player
	gs := session.GameState
	resourceType := cmd.Value

	amount := game.HarvestResource(resourceType, &player.ResourceStorageMap)
	if e.metrics != nil {
		e.metrics.RecordHarvest(resourceType, amount)
	}

	// Apply town tax
	taxMsg := ""
	if e.store != nil && amount > 0 {
		town, err := e.store.LoadTown(game.DefaultTownName)
		if err == nil && town.TaxRate > 0 {
			netAmount, taxAmount := game.CalculateTax(amount, town.TaxRate)
			if taxAmount > 0 {
				// Reduce player's harvest by tax
				res := player.ResourceStorageMap[resourceType]
				res.Stock -= taxAmount
				player.ResourceStorageMap[resourceType] = res

				// Add to treasury
				if town.Treasury == nil {
					town.Treasury = make(map[string]int)
				}
				town.Treasury[resourceType] += taxAmount
				e.store.SaveTown(town)

				taxMsg = fmt.Sprintf("Tax collected: %d %s (%d%% tax, net: %d)", taxAmount, resourceType, town.TaxRate, netAmount)
			}
		}
	}

	msgs := []GameMessage{
		Msg(fmt.Sprintf("Harvested %d %s!", amount, resourceType), "loot"),
	}
	if taxMsg != "" {
		msgs = append(msgs, Msg(taxMsg, "system"))
	}
	msgs = append(msgs,
		Msg("", "system"),
		Msg("Current Resources:", "system"),
	)
	for _, res := range data.ResourceTypes {
		r, exists := player.ResourceStorageMap[res]
		if exists {
			msgs = append(msgs, Msg(fmt.Sprintf("  %s: %d", res, r.Stock), "system"))
		}
	}
	// Also show beast materials if any
	for _, matName := range data.BeastMaterials {
		r, exists := player.ResourceStorageMap[matName]
		if exists && r.Stock > 0 {
			msgs = append(msgs, Msg(fmt.Sprintf("  %s: %d", matName, r.Stock), "system"))
		}
	}

	gs.CharactersMap[player.Name] = *player
	session.State = StateMainMenu
	resp := BuildMainMenuResponse(session)
	resp.Messages = append(msgs, resp.Messages...)
	return resp
}

// handleHuntLocationSelect processes the selection of a hunting location.
func (e *Engine) handleHuntLocationSelect(session *GameSession, cmd GameCommand) GameResponse {
	gs := session.GameState
	player := session.Player
	locName := cmd.Value

	// Check if this is a locked location (guardian fight)
	if strings.HasPrefix(locName, "locked:") {
		actualName := strings.TrimPrefix(locName, "locked:")
		loc, exists := gs.GameLocations[actualName]
		if !exists {
			session.State = StateMainMenu
			resp := BuildMainMenuResponse(session)
			resp.Messages = append([]GameMessage{
				Msg(fmt.Sprintf("Location '%s' not found!", actualName), "error"),
			}, resp.Messages...)
			return resp
		}

		guardian := game.GenerateLocationGuardian(actualName, loc, gs)
		session.SelectedLocation = actualName
		session.Combat = &CombatContext{
			GuardianLocationName: actualName,
		}
		return e.startCombat(session, &loc, 0, guardian)
	}

	loc, exists := gs.GameLocations[locName]
	if !exists {
		session.State = StateMainMenu
		resp := BuildMainMenuResponse(session)
		resp.Messages = append([]GameMessage{
			Msg(fmt.Sprintf("Location '%s' not found!", locName), "error"),
		}, resp.Messages...)
		return resp
	}

	// Verify player knows this location
	known := false
	for _, kl := range player.KnownLocations {
		if kl == locName {
			known = true
			break
		}
	}
	if !known || loc.Type == "Base" {
		session.State = StateMainMenu
		resp := BuildMainMenuResponse(session)
		resp.Messages = append([]GameMessage{
			Msg("Cannot hunt at this location!", "error"),
		}, resp.Messages...)
		return resp
	}

	session.SelectedLocation = locName

	// Check if player has Tracking skill
	hasTracking := false
	for _, skill := range player.LearnedSkills {
		if skill.Name == "Tracking" {
			hasTracking = true
			break
		}
	}

	if hasTracking {
		session.State = StateHuntTracking
		msgs := []GameMessage{
			Msg(fmt.Sprintf("Hunting at %s (%s)", locName, loc.Type), "system"),
			Msg("TRACKING ACTIVE - Choose your target:", "system"),
			Msg("============================================================", "system"),
		}
		options := []MenuOption{
			Opt("0", "Random Target"),
		}
		for idx, monster := range loc.Monsters {
			guardianTag := ""
			if monster.IsSkillGuardian {
				guardianTag = " [SKILL GUARDIAN]"
			}
			rarityTag := ""
			rarityName := game.RarityDisplayName(monster.Rarity)
			if rarityName != "Common" {
				rarityTag = fmt.Sprintf("[%s] ", rarityName)
			}
			label := fmt.Sprintf("%s%s (Lv%d) HP:%d/%d%s",
				rarityTag, monster.Name, monster.Level,
				monster.HitpointsRemaining, monster.HitpointsTotal,
				guardianTag)
			options = append(options, Opt(strconv.Itoa(idx+1), label))
		}

		session.Combat = &CombatContext{
			ContinuousHunt: true,
			Location:       &loc,
		}

		return GameResponse{
			Type:     "menu",
			Messages: msgs,
			State:    &StateData{Screen: "hunt_tracking", Player: MakePlayerState(player)},
			Options:  options,
		}
	}

	// No tracking -- pick random monster
	mobLoc := rand.Intn(len(loc.Monsters))
	mob := loc.Monsters[mobLoc]

	// Set continuous hunt so startCombat picks it up
	session.Combat = &CombatContext{ContinuousHunt: true}
	return e.startCombat(session, &loc, mobLoc, mob)
}

// handleHuntTracking processes the player's target selection when using the Tracking skill.
func (e *Engine) handleHuntTracking(session *GameSession, cmd GameCommand) GameResponse {
	player := session.Player
	gs := session.GameState

	locName := session.SelectedLocation
	loc, exists := gs.GameLocations[locName]
	if !exists {
		session.State = StateMainMenu
		return BuildMainMenuResponse(session)
	}

	choice, err := strconv.Atoi(cmd.Value)
	if err != nil || choice < 0 || choice > len(loc.Monsters) {
		choice = 0 // Default to random
	}

	var mobLoc int
	if choice == 0 {
		mobLoc = rand.Intn(len(loc.Monsters))
	} else {
		mobLoc = choice - 1
	}

	mob := loc.Monsters[mobLoc]

	_ = player // player is used by startCombat via session
	return e.startCombat(session, &loc, mobLoc, mob)
}

// startCombat initializes a CombatContext and returns the initial combat display.
func (e *Engine) startCombat(session *GameSession, location *models.Location, mobLoc int, mob models.Monster) GameResponse {
	player := session.Player
	gs := session.GameState

	// Handle resurrection before combat
	if player.HitpointsRemaining <= 0 {
		player.HitpointsRemaining = player.HitpointsTotal
		player.Resurrections++
	}

	// Preserve fields from pre-combat context (e.g. tracking sets ContinuousHunt)
	guardianLocName := ""
	continuousHunt := false
	if session.Combat != nil {
		guardianLocName = session.Combat.GuardianLocationName
		continuousHunt = session.Combat.ContinuousHunt
	}

	// 1% chance to dynamically spawn a skill guardian (skip for location guardian fights)
	if guardianLocName == "" && !mob.IsSkillGuardian && location != nil && location.LevelMax >= 10 && rand.Intn(100) < 1 {
		guardableSkills := []models.Skill{}
		for _, skill := range data.AvailableSkills {
			if skill.Name != "Tracking" && skill.Name != "Power Strike" {
				guardableSkills = append(guardableSkills, skill)
			}
		}
		if len(guardableSkills) > 0 {
			guardianSkill := guardableSkills[rand.Intn(len(guardableSkills))]
			guardianLevel := location.LevelMax + rand.Intn(location.LevelMax/2+1) + 3
			mob = game.GenerateSkillGuardian(guardianSkill, guardianLevel, location.RarityMax)
			mob.LocationName = location.Name
			fmt.Printf("[Guardian] Spawned %s guardian at %s\n", guardianSkill.Name, location.Name)
			if e.metrics != nil {
				e.metrics.RecordGuardianSpawn(guardianSkill.Name)
			}
		}
	}

	session.Combat = &CombatContext{
		Mob:                  mob,
		MobLoc:               mobLoc,
		Location:             location,
		Turn:                 0,
		Fled:                 false,
		PlayerWon:            false,
		IsDefending:          false,
		ContinuousHunt:       continuousHunt,
		GuardianLocationName: guardianLocName,
	}

	// Check for guards for guardian/boss fights
	isSpecialFight := mob.IsSkillGuardian || mob.IsBoss
	if isSpecialFight && gs.Villages != nil {
		village, exists := gs.Villages[player.VillageName]
		if exists {
			var availableGuards []models.Guard
			for _, guard := range village.ActiveGuards {
				if !guard.Injured && guard.HitpointsRemaining > 0 {
					g := guard
					g.HitpointsRemaining = g.HitPoints
					availableGuards = append(availableGuards, g)
				}
			}

			if len(availableGuards) > 0 {
				session.Combat.CombatGuards = availableGuards
				session.State = StateCombatGuardPrompt

				msgs := []GameMessage{}
				if mob.IsBoss {
					msgs = append(msgs, Msg("WARNING: BOSS FIGHT", "combat"))
					msgs = append(msgs, Msg("Guards can DIE PERMANENTLY in boss fights!", "combat"))
				} else if mob.IsSkillGuardian {
					msgs = append(msgs, Msg(fmt.Sprintf("SKILL GUARDIAN: %s guards the skill: %s", mob.Name, mob.GuardedSkill.Name), "combat"))
				}
				msgs = append(msgs, Msg(fmt.Sprintf("Available guards: %d", len(availableGuards)), "system"))
				for _, guard := range availableGuards {
					msgs = append(msgs, Msg(fmt.Sprintf("  %s (Lv%d, HP:%d)", guard.Name, guard.Level, guard.HitPoints), "system"))
				}
				msgs = append(msgs, Msg("Bring guards to this fight?", "system"))

				return GameResponse{
					Type:     "menu",
					Messages: msgs,
					State: &StateData{
						Screen: "combat_guard_prompt",
						Player: MakePlayerState(player),
						Combat: MakeCombatView(session),
					},
					Options: []MenuOption{
						Opt("y", "Yes, bring guards"),
						Opt("n", "No, fight alone"),
					},
				}
			}
		}
	}

	// Restore mana/stamina at combat start
	player.ManaRemaining = player.ManaTotal
	player.StaminaRemaining = player.StaminaTotal
	session.Combat.Mob.ManaRemaining = session.Combat.Mob.ManaTotal
	session.Combat.Mob.StaminaRemaining = session.Combat.Mob.StaminaTotal

	session.State = StateCombat
	return buildCombatDisplay(session)
}

// buildCombatDisplay creates the combat view response with action options.
func buildCombatDisplay(session *GameSession) GameResponse {
	player := session.Player
	c := session.Combat
	mob := &c.Mob

	msgs := []GameMessage{
		Msg(fmt.Sprintf("========== TURN %d ==========", c.Turn), "combat"),
		Msg(fmt.Sprintf("[%s] HP:%d/%d | MP:%d/%d | SP:%d/%d",
			player.Name,
			player.HitpointsRemaining, player.HitpointsTotal,
			player.ManaRemaining, player.ManaTotal,
			player.StaminaRemaining, player.StaminaTotal), "combat"),
		Msg(fmt.Sprintf("[%s] HP:%d/%d | MP:%d/%d | SP:%d/%d",
			mob.Name,
			mob.HitpointsRemaining, mob.HitpointsTotal,
			mob.ManaRemaining, mob.ManaTotal,
			mob.StaminaRemaining, mob.StaminaTotal), "combat"),
	}

	if c.Turn == 0 {
		guardianTag := ""
		if mob.IsSkillGuardian {
			guardianTag = fmt.Sprintf(" [SKILL GUARDIAN - %s]", mob.GuardedSkill.Name)
		}
		if mob.IsBoss {
			guardianTag = " [BOSS]"
		}
		rarityTag := ""
		mobRarityName := game.RarityDisplayName(mob.Rarity)
		if mobRarityName != "Common" {
			rarityTag = fmt.Sprintf(" [%s]", mobRarityName)
		}
		msgs = []GameMessage{
			Msg(fmt.Sprintf("Lv%d %s vs Lv%d %s (%s)%s%s",
				player.Level, player.Name,
				mob.Level, mob.Name, mob.MonsterType, rarityTag, guardianTag), "combat"),
			Msg(fmt.Sprintf("[%s] HP:%d/%d | MP:%d/%d | SP:%d/%d",
				player.Name,
				player.HitpointsRemaining, player.HitpointsTotal,
				player.ManaRemaining, player.ManaTotal,
				player.StaminaRemaining, player.StaminaTotal), "combat"),
			Msg(fmt.Sprintf("[%s] HP:%d/%d | MP:%d/%d | SP:%d/%d",
				mob.Name,
				mob.HitpointsRemaining, mob.HitpointsTotal,
				mob.ManaRemaining, mob.ManaTotal,
				mob.StaminaRemaining, mob.StaminaTotal), "combat"),
		}
	}

	// Show status effects
	if len(player.StatusEffects) > 0 {
		effectStr := fmt.Sprintf("%s effects: ", player.Name)
		for _, eff := range player.StatusEffects {
			effectStr += fmt.Sprintf("[%s:%d] ", eff.Type, eff.Duration)
		}
		msgs = append(msgs, Msg(effectStr, "buff"))
	}
	if len(mob.StatusEffects) > 0 {
		effectStr := fmt.Sprintf("%s effects: ", mob.Name)
		for _, eff := range mob.StatusEffects {
			effectStr += fmt.Sprintf("[%s:%d] ", eff.Type, eff.Duration)
		}
		msgs = append(msgs, Msg(effectStr, "debuff"))
	}

	// Show guards if present
	if c.HasGuards && len(c.CombatGuards) > 0 {
		msgs = append(msgs, Msg("--- Guards ---", "system"))
		for _, guard := range c.CombatGuards {
			status := "Ready"
			if guard.Injured {
				status = "Injured"
			}
			msgs = append(msgs, Msg(fmt.Sprintf("  %s HP:%d/%d [%s]", guard.Name, guard.HitpointsRemaining, guard.HitPoints, status), "system"))
		}
	}

	options := combatActionOptions()

	return GameResponse{
		Type:     "combat",
		Messages: msgs,
		State: &StateData{
			Screen: "combat",
			Player: MakePlayerState(player),
			Combat: MakeCombatView(session),
		},
		Options: options,
	}
}

// handleAutoPlaySpeed selects the speed and starts auto-play.
func (e *Engine) handleAutoPlaySpeed(session *GameSession, cmd GameCommand) GameResponse {
	player := session.Player
	gs := session.GameState

	speedMap := map[string]string{
		"1": "slow",
		"2": "normal",
		"3": "fast",
		"4": "turbo",
	}
	speed := speedMap[cmd.Value]
	if speed == "" {
		speed = "normal"
	}

	// Save game before auto-play
	gs.CharactersMap[player.Name] = *player
	game.WriteGameStateToFile(*gs, session.SaveFile)

	// Find first non-Base hunt location
	var huntLocation *models.Location
	var huntLocationName string
	for _, locName := range player.KnownLocations {
		loc, exists := gs.GameLocations[locName]
		if exists && loc.Type != "Base" {
			l := loc
			huntLocation = &l
			huntLocationName = locName
			break
		}
	}

	if huntLocation == nil {
		session.State = StateMainMenu
		resp := BuildMainMenuResponse(session)
		resp.Messages = append([]GameMessage{
			Msg("No huntable locations available!", "error"),
		}, resp.Messages...)
		return resp
	}

	// Initialize combat context for auto-play
	session.Combat = &CombatContext{
		IsAutoPlay:    true,
		AutoPlaySpeed: speed,
	}

	// Determine number of fights per batch based on speed
	fightsPerBatch := 5
	switch speed {
	case "slow":
		fightsPerBatch = 3
	case "normal":
		fightsPerBatch = 5
	case "fast":
		fightsPerBatch = 10
	case "turbo":
		fightsPerBatch = 20
	}

	msgs := []GameMessage{
		Msg(fmt.Sprintf("AUTO-PLAY MODE - Speed: %s (%d fights)", speed, fightsPerBatch), "system"),
		Msg(fmt.Sprintf("Hunting at: %s", huntLocationName), "system"),
		Msg("", "system"),
	}

	// Run multiple auto-play fights
	for fight := 0; fight < fightsPerBatch; fight++ {
		// Re-fetch location in case monsters changed
		loc := gs.GameLocations[huntLocationName]
		if len(loc.Monsters) == 0 {
			break
		}
		mobLoc := rand.Intn(len(loc.Monsters))
		mobCopy := loc.Monsters[mobLoc]

		fightMsgs := e.autoPlayOneFight(session, player, gs, &mobCopy, huntLocation, mobLoc, huntLocationName)
		msgs = append(msgs, fightMsgs...)
		msgs = append(msgs, Msg("", "system"))
	}

	// Show summary
	msgs = append(msgs, Msg("", "system"))
	msgs = append(msgs, Msg("--- Auto-Play Statistics ---", "system"))
	msgs = append(msgs, Msg(fmt.Sprintf("Fights: %d | Wins: %d | Deaths: %d",
		session.Combat.AutoPlayFights, session.Combat.AutoPlayWins, session.Combat.AutoPlayDeaths), "system"))
	msgs = append(msgs, Msg(fmt.Sprintf("Level: %d | XP: %d | Total XP Gained: %d",
		player.Level, player.Experience, session.Combat.AutoPlayXP), "system"))
	msgs = append(msgs, Msg(fmt.Sprintf("HP: %d/%d | MP: %d/%d | SP: %d/%d",
		player.HitpointsRemaining, player.HitpointsTotal,
		player.ManaRemaining, player.ManaTotal,
		player.StaminaRemaining, player.StaminaTotal), "system"))

	session.State = StateAutoPlayMenu
	return GameResponse{
		Type:     "menu",
		Messages: msgs,
		State:    &StateData{Screen: "autoplay_menu", Player: MakePlayerState(player)},
		Options: []MenuOption{
			Opt("1", "View Inventory"),
			Opt("2", "View Skills"),
			Opt("3", "View Equipment"),
			Opt("4", "Quest Log"),
			Opt("5", "View Full Character Stats"),
			Opt("6", "Resume Auto-Play"),
			Opt("0", "Return to Main Menu"),
		},
	}
}

// autoPlayOneFight processes a full combat automatically and returns messages.
func (e *Engine) autoPlayOneFight(session *GameSession, player *models.Character, gs *models.GameState,
	mob *models.Monster, location *models.Location, mobLoc int, locationName string) []GameMessage {

	msgs := []GameMessage{}

	// Handle resurrection
	if player.HitpointsRemaining <= 0 {
		player.HitpointsRemaining = player.HitpointsTotal
		player.Resurrections++
		session.Combat.AutoPlayDeaths++
		msgs = append(msgs, Msg(fmt.Sprintf("RESURRECTION #%d", player.Resurrections), "system"))
	}

	// 1% chance to dynamically spawn a skill guardian
	if location != nil && location.LevelMax >= 10 && rand.Intn(100) < 1 {
		guardableSkills := []models.Skill{}
		for _, skill := range data.AvailableSkills {
			if skill.Name != "Tracking" && skill.Name != "Power Strike" {
				guardableSkills = append(guardableSkills, skill)
			}
		}
		if len(guardableSkills) > 0 {
			guardianSkill := guardableSkills[rand.Intn(len(guardableSkills))]
			guardianLevel := location.LevelMax + rand.Intn(location.LevelMax/2+1) + 3
			guardian := game.GenerateSkillGuardian(guardianSkill, guardianLevel, location.RarityMax)
			guardian.LocationName = locationName
			*mob = guardian
			fmt.Printf("[Guardian] Spawned %s guardian at %s (autoplay)\n", guardianSkill.Name, locationName)
			if e.metrics != nil {
				e.metrics.RecordGuardianSpawn(guardianSkill.Name)
			}
		}
	}

	// Restore resources at start
	player.ManaRemaining = player.ManaTotal
	player.StaminaRemaining = player.StaminaTotal
	mob.ManaRemaining = mob.ManaTotal
	mob.StaminaRemaining = mob.StaminaTotal

	// Clear status effects
	player.StatusEffects = []models.StatusEffect{}
	mob.StatusEffects = []models.StatusEffect{}

	session.Combat.AutoPlayFights++
	turnCount := 0

	msgs = append(msgs, Msg(fmt.Sprintf("Fight: %s (Lv%d) vs %s (Lv%d)",
		player.Name, player.Level, mob.Name, mob.Level), "combat"))

	startXP := player.Experience
	_ = startXP

	for player.HitpointsRemaining > 0 && mob.HitpointsRemaining > 0 {
		turnCount++

		// Safety valve to prevent infinite loops
		if turnCount > 200 {
			msgs = append(msgs, Msg("Combat timed out!", "combat"))
			break
		}

		// Process status effects for player
		for i := len(player.StatusEffects) - 1; i >= 0; i-- {
			effect := &player.StatusEffects[i]
			switch effect.Type {
			case "poison":
				player.HitpointsRemaining -= effect.Potency
			case "burn":
				player.HitpointsRemaining -= effect.Potency
			case "regen":
				player.HitpointsRemaining += effect.Potency
				if player.HitpointsRemaining > player.HitpointsTotal {
					player.HitpointsRemaining = player.HitpointsTotal
				}
			}
			effect.Duration--
			if effect.Duration <= 0 {
				switch effect.Type {
				case "buff_attack":
					player.StatsMod.AttackMod -= effect.Potency
				case "buff_defense":
					player.StatsMod.DefenseMod -= effect.Potency
				}
				player.StatusEffects = append(player.StatusEffects[:i], player.StatusEffects[i+1:]...)
			}
		}

		// Process status effects for monster
		for i := len(mob.StatusEffects) - 1; i >= 0; i-- {
			effect := &mob.StatusEffects[i]
			switch effect.Type {
			case "poison":
				mob.HitpointsRemaining -= effect.Potency
			case "burn":
				mob.HitpointsRemaining -= effect.Potency
			case "regen":
				mob.HitpointsRemaining += effect.Potency
				if mob.HitpointsRemaining > mob.HitpointsTotal {
					mob.HitpointsRemaining = mob.HitpointsTotal
				}
			}
			effect.Duration--
			if effect.Duration <= 0 {
				switch effect.Type {
				case "buff_attack":
					mob.StatsMod.AttackMod -= effect.Potency
				case "buff_defense":
					mob.StatsMod.DefenseMod -= effect.Potency
				}
				mob.StatusEffects = append(mob.StatusEffects[:i], mob.StatusEffects[i+1:]...)
			}
		}

		if player.HitpointsRemaining <= 0 || mob.HitpointsRemaining <= 0 {
			break
		}

		// Check if player is stunned
		playerStunned := false
		for _, eff := range player.StatusEffects {
			if eff.Type == "stun" {
				playerStunned = true
				break
			}
		}

		if !playerStunned {
			// AI makes decision
			decision := game.MakeAIDecision(player, mob, turnCount)

			switch decision {
			case "attack":
				playerAttack := game.MultiRoll(player.AttackRolls) + player.StatsMod.AttackMod
				if rand.Intn(100) < 15 {
					playerAttack = playerAttack * 2
				}
				mobDef := game.MultiRoll(mob.DefenseRolls) + mob.StatsMod.DefenseMod
				if playerAttack > mobDef {
					diff := game.ApplyDamage(playerAttack-mobDef, models.Physical, mob)
					mob.HitpointsRemaining -= diff
				}

			case "item":
				for idx, item := range player.Inventory {
					if item.ItemType == "consumable" {
						game.UseConsumableItem(item, player)
						game.RemoveItemFromInventory(&player.Inventory, idx)
						break
					}
				}

			default:
				// Skill usage
				if len(decision) > 6 && decision[:6] == "skill_" {
					skillName := decision[6:]
					for _, skill := range player.LearnedSkills {
						nameMatch := skill.Name == skillName
						if !nameMatch {
							// Flexible matching for common AI decisions
							switch {
							case skill.Name == "Heal" && skillName == "heal":
								nameMatch = true
							case skill.Name == "Regeneration" && skillName == "regeneration":
								nameMatch = true
							case skill.Name == "Battle Cry" && skillName == "Battle Cry":
								nameMatch = true
							case skill.Name == "Shield Wall" && skillName == "Shield Wall":
								nameMatch = true
							}
						}
						if nameMatch && skill.ManaCost <= player.ManaRemaining && skill.StaminaCost <= player.StaminaRemaining {
							player.ManaRemaining -= skill.ManaCost
							player.StaminaRemaining -= skill.StaminaCost

							if skill.Damage < 0 {
								player.HitpointsRemaining += -skill.Damage
								if player.HitpointsRemaining > player.HitpointsTotal {
									player.HitpointsRemaining = player.HitpointsTotal
								}
							} else if skill.Damage > 0 {
								finalDamage := game.ApplyDamage(skill.Damage, skill.DamageType, mob)
								mob.HitpointsRemaining -= finalDamage
							}

							if skill.Effect.Type != "none" && skill.Effect.Duration > 0 {
								if skill.Effect.Type == "buff_attack" || skill.Effect.Type == "buff_defense" || skill.Effect.Type == "regen" {
									player.StatusEffects = append(player.StatusEffects, skill.Effect)
									if skill.Effect.Type == "buff_attack" {
										player.StatsMod.AttackMod += skill.Effect.Potency
									} else if skill.Effect.Type == "buff_defense" {
										player.StatsMod.DefenseMod += skill.Effect.Potency
									}
								} else {
									mob.StatusEffects = append(mob.StatusEffects, skill.Effect)
								}
							}
							break
						}
					}
				}
			}
		}

		// Monster's turn
		if mob.HitpointsRemaining > 0 {
			mobStunned := false
			for _, eff := range mob.StatusEffects {
				if eff.Type == "stun" {
					mobStunned = true
					break
				}
			}

			if !mobStunned {
				useMonsterSkill := false
				if len(mob.LearnedSkills) > 0 && rand.Intn(100) < 40 {
					skill := mob.LearnedSkills[rand.Intn(len(mob.LearnedSkills))]
					if skill.ManaCost <= mob.ManaRemaining && skill.StaminaCost <= mob.StaminaRemaining {
						mob.ManaRemaining -= skill.ManaCost
						mob.StaminaRemaining -= skill.StaminaCost
						useMonsterSkill = true

						if skill.Damage < 0 {
							mob.HitpointsRemaining += -skill.Damage
							if mob.HitpointsRemaining > mob.HitpointsTotal {
								mob.HitpointsRemaining = mob.HitpointsTotal
							}
						} else if skill.Damage > 0 {
							finalDamage := game.ApplyDamage(skill.Damage, skill.DamageType, player)
							player.HitpointsRemaining -= finalDamage
						}

						if skill.Effect.Type != "none" && skill.Effect.Duration > 0 {
							if skill.Effect.Type == "buff_attack" || skill.Effect.Type == "buff_defense" {
								mob.StatusEffects = append(mob.StatusEffects, skill.Effect)
								if skill.Effect.Type == "buff_attack" {
									mob.StatsMod.AttackMod += skill.Effect.Potency
								} else if skill.Effect.Type == "buff_defense" {
									mob.StatsMod.DefenseMod += skill.Effect.Potency
								}
							} else {
								player.StatusEffects = append(player.StatusEffects, skill.Effect)
							}
						}
					}
				}

				if !useMonsterSkill {
					mobAttack := game.MultiRoll(mob.AttackRolls) + mob.StatsMod.AttackMod
					if rand.Intn(100) < 10 {
						mobAttack = mobAttack * 2
					}
					playerDef := game.MultiRoll(player.DefenseRolls) + player.StatsMod.DefenseMod
					if mobAttack > playerDef {
						diff := game.ApplyDamage(mobAttack-playerDef, models.Physical, player)
						player.HitpointsRemaining -= diff
					}
				}
			}
		}
	}

	// Combat resolution
	if player.HitpointsRemaining > 0 {
		xpGained := mob.Level * 10
		player.Experience += xpGained
		session.Combat.AutoPlayWins++
		session.Combat.AutoPlayXP += xpGained

		msgs = append(msgs, Msg(fmt.Sprintf("  VICTORY! (+%d XP)", xpGained), "combat"))

		// Handle skill guardian reward -- auto-learn/upgrade in auto-play
		if mob.IsSkillGuardian {
			if e.metrics != nil {
				e.metrics.RecordGuardianDefeat(mob.GuardedSkill.Name)
			}
			existingIdx := -1
			for i, s := range player.LearnedSkills {
				if s.Name == mob.GuardedSkill.Name {
					existingIdx = i
					break
				}
			}
			if existingIdx >= 0 {
				game.UpgradeSkill(&player.LearnedSkills[existingIdx])
				msgs = append(msgs, Msg(fmt.Sprintf("  SKILL GUARDIAN DEFEATED! %s upgraded! (+5 dmg, -2 cost) [+%d]", mob.GuardedSkill.Name, player.LearnedSkills[existingIdx].UpgradeCount), "loot"))
				if e.metrics != nil {
					e.metrics.RecordSkillUpgraded(mob.GuardedSkill.Name)
				}
			} else {
				player.LearnedSkills = append(player.LearnedSkills, mob.GuardedSkill)
				msgs = append(msgs, Msg(fmt.Sprintf("  SKILL GUARDIAN DEFEATED! Learned %s!", mob.GuardedSkill.Name), "loot"))
				if e.metrics != nil {
					e.metrics.RecordSkillLearned(mob.GuardedSkill.Name)
				}
			}
		}

		// Loot equipment
		for _, item := range mob.EquipmentMap {
			game.EquipBestItem(item, &player.EquipmentMap, &player.Inventory)
		}

		// Drop beast materials
		matName, matQty := game.DropBeastMaterial(mob.MonsterType, player)
		if matName != "" {
			msgs = append(msgs, Msg(fmt.Sprintf("  Dropped: %d %s", matQty, matName), "loot"))
		}

		// Chance for potion
		if rand.Intn(100) < 30 {
			potion := game.CreateHealthPotion("small")
			if rand.Intn(100) < 30 {
				potion = game.CreateHealthPotion("medium")
			}
			player.Inventory = append(player.Inventory, potion)
		}

		// 15% chance to rescue a villager (only after elder quest completed) or get a hint
		if rand.Intn(100) < 15 {
			if !game.Contains(player.CompletedQuests, "quest_v0_elder") {
				msgs = append(msgs, Msg("  You hear rumors of a Village Elder held captive in the Lake Ruins...", "narrative"))
			} else {
				if gs.Villages == nil {
					gs.Villages = make(map[string]models.Village)
				}
				village, exists := gs.Villages[player.VillageName]
				if !exists {
					village = game.GenerateVillage(player.Name)
					player.VillageName = player.Name + "'s Village"
				}
				game.RescueVillager(&village)
				village.Experience += 25
				gs.Villages[player.VillageName] = village
				msgs = append(msgs, Msg("  A villager was rescued! (+25 Village XP)", "narrative"))
			}
		}

		// Elder rescue: 20% chance at Lake Ruins if quest active
		if locationName == "Lake Ruins" && game.Contains(player.ActiveQuests, "quest_v0_elder") {
			if rand.Intn(100) < 20 {
				if q, ok := gs.AvailableQuests["quest_v0_elder"]; ok {
					q.Requirement.CurrentValue = 1
					gs.AvailableQuests["quest_v0_elder"] = q
					msgs = append(msgs, Msg("  You found a Village Elder held captive! They agree to help you establish a village.", "narrative"))
				}
			}
		}

		// 15% chance to discover a new location
		if rand.Intn(100) < 15 {
			combinedSeen := append([]string{}, player.KnownLocations...)
			combinedSeen = append(combinedSeen, player.LockedLocations...)
			discovered := game.SearchLocation(combinedSeen, data.DiscoverableLocations)
			if discovered != "" {
				locData, locExists := gs.GameLocations[discovered]
				if locExists && locData.Type == "Base" {
					if !game.Contains(player.KnownLocations, discovered) {
						player.KnownLocations = append(player.KnownLocations, discovered)
					}
					msgs = append(msgs, Msg(fmt.Sprintf("  You discovered a safe area: %s!", discovered), "narrative"))
					// Unlock village capability
					if gs.Villages == nil {
						gs.Villages = make(map[string]models.Village)
					}
					village, exists := gs.Villages[player.VillageName]
					if !exists {
						village = game.GenerateVillage(player.Name)
						player.VillageName = player.Name + "'s Village"
					}
					unlockMsg := game.UnlockBaseLocationCapability(&village, discovered)
					if unlockMsg != "" {
						msgs = append(msgs, Msg("  "+unlockMsg, "narrative"))
					}
					gs.Villages[player.VillageName] = village
				} else {
					if !game.Contains(player.LockedLocations, discovered) {
						player.LockedLocations = append(player.LockedLocations, discovered)
					}
					msgs = append(msgs, Msg(fmt.Sprintf("  You discovered a new area: %s! A powerful guardian blocks the entrance.", discovered), "narrative"))
				}
			}
		}

		player.StatsMod = game.CalculateItemMods(player.EquipmentMap)
		player.HitpointsTotal = player.HitpointsNatural + player.StatsMod.HitPointMod

		// Respawn monster at location
		loc := gs.GameLocations[locationName]
		loc.Monsters[mobLoc] = game.GenerateBestMonster(gs, location.LevelMax, location.RarityMax)
		gs.GameLocations[locationName] = loc
	} else {
		msgs = append(msgs, Msg(fmt.Sprintf("  DEFEAT! %s has fallen!", player.Name), "combat"))

		// Transfer equipment to monster
		for _, item := range player.EquipmentMap {
			game.EquipBestItem(item, &mob.EquipmentMap, &mob.Inventory)
		}

		loc := gs.GameLocations[locationName]
		loc.Monsters[mobLoc].StatsMod = game.CalculateItemMods(mob.EquipmentMap)
		loc.Monsters[mobLoc].Experience += player.Level * 100
		gs.GameLocations[locationName] = loc
	}

	// Level up
	game.LevelUp(player)
	loc := gs.GameLocations[locationName]
	game.LevelUpMob(&loc.Monsters[mobLoc])
	gs.GameLocations[locationName] = loc

	// Increment location quest progress on combat victory
	game.IncrementLocationQuestProgress(player, gs, locationName)

	// Check quest progress
	completedQuests := game.CheckQuestProgress(player, gs)
	if e.metrics != nil {
		for _, qid := range completedQuests {
			e.metrics.RecordQuestComplete(qid)
		}
	}

	// Process guard recovery
	if gs.Villages != nil {
		if village, exists := gs.Villages[player.VillageName]; exists {
			game.ProcessGuardRecovery(&village)
			gs.Villages[player.VillageName] = village
		}
	}

	// Save player state
	gs.CharactersMap[player.Name] = *player

	return msgs
}

// handleAutoPlayMenu processes the post-auto-play menu options.
func (e *Engine) handleAutoPlayMenu(session *GameSession, cmd GameCommand) GameResponse {
	player := session.Player
	gs := session.GameState

	switch cmd.Value {
	case "1":
		// View Inventory
		msgs := buildInventoryMessages(player)
		session.State = StateAutoPlayMenu
		return GameResponse{
			Type:     "menu",
			Messages: msgs,
			State:    &StateData{Screen: "autoplay_menu", Player: MakePlayerState(player)},
			Options: []MenuOption{
				Opt("1", "View Inventory"),
				Opt("2", "View Skills"),
				Opt("3", "View Equipment"),
				Opt("4", "Quest Log"),
				Opt("5", "View Full Character Stats"),
				Opt("6", "Resume Auto-Play"),
				Opt("0", "Return to Main Menu"),
			},
		}

	case "2":
		// View Skills
		msgs := buildSkillsMessages(player)
		session.State = StateAutoPlayMenu
		return GameResponse{
			Type:     "menu",
			Messages: msgs,
			State:    &StateData{Screen: "autoplay_menu", Player: MakePlayerState(player)},
			Options: []MenuOption{
				Opt("1", "View Inventory"),
				Opt("2", "View Skills"),
				Opt("3", "View Equipment"),
				Opt("4", "Quest Log"),
				Opt("5", "View Full Character Stats"),
				Opt("6", "Resume Auto-Play"),
				Opt("0", "Return to Main Menu"),
			},
		}

	case "3":
		// View Equipment
		msgs := buildEquipmentMessages(player)
		session.State = StateAutoPlayMenu
		return GameResponse{
			Type:     "menu",
			Messages: msgs,
			State:    &StateData{Screen: "autoplay_menu", Player: MakePlayerState(player)},
			Options: []MenuOption{
				Opt("1", "View Inventory"),
				Opt("2", "View Skills"),
				Opt("3", "View Equipment"),
				Opt("4", "Quest Log"),
				Opt("5", "View Full Character Stats"),
				Opt("6", "Resume Auto-Play"),
				Opt("0", "Return to Main Menu"),
			},
		}

	case "4":
		// Quest Log
		msgs := buildQuestLogMessages(player, gs)
		session.State = StateAutoPlayMenu
		return GameResponse{
			Type:     "menu",
			Messages: msgs,
			State:    &StateData{Screen: "autoplay_menu", Player: MakePlayerState(player)},
			Options: []MenuOption{
				Opt("1", "View Inventory"),
				Opt("2", "View Skills"),
				Opt("3", "View Equipment"),
				Opt("4", "Quest Log"),
				Opt("5", "View Full Character Stats"),
				Opt("6", "Resume Auto-Play"),
				Opt("0", "Return to Main Menu"),
			},
		}

	case "5":
		// View Stats
		msgs := buildPlayerStatsMessages(player, gs)
		session.State = StateAutoPlayMenu
		return GameResponse{
			Type:     "menu",
			Messages: msgs,
			State:    &StateData{Screen: "autoplay_menu", Player: MakePlayerState(player)},
			Options: []MenuOption{
				Opt("1", "View Inventory"),
				Opt("2", "View Skills"),
				Opt("3", "View Equipment"),
				Opt("4", "Quest Log"),
				Opt("5", "View Full Character Stats"),
				Opt("6", "Resume Auto-Play"),
				Opt("0", "Return to Main Menu"),
			},
		}

	case "6":
		// Resume Auto-Play at the same speed
		speed := "normal"
		if session.Combat != nil && session.Combat.AutoPlaySpeed != "" {
			speed = session.Combat.AutoPlaySpeed
		}
		speedNum := "2"
		switch speed {
		case "slow":
			speedNum = "1"
		case "normal":
			speedNum = "2"
		case "fast":
			speedNum = "3"
		case "turbo":
			speedNum = "4"
		}
		return e.handleAutoPlaySpeed(session, GameCommand{Type: "select", Value: speedNum})

	case "0":
		// Return to Main Menu
		gs.CharactersMap[player.Name] = *player
		game.WriteGameStateToFile(*gs, session.SaveFile)
		session.Combat = nil
		session.State = StateMainMenu
		resp := BuildMainMenuResponse(session)
		resp.Messages = append([]GameMessage{
			Msg("Auto-play session complete. Game saved.", "system"),
		}, resp.Messages...)
		return resp

	default:
		session.State = StateAutoPlayMenu
		return GameResponse{
			Type:     "menu",
			Messages: []GameMessage{Msg("Invalid choice.", "error")},
			State:    &StateData{Screen: "autoplay_menu", Player: MakePlayerState(player)},
			Options: []MenuOption{
				Opt("1", "View Inventory"),
				Opt("2", "View Skills"),
				Opt("3", "View Equipment"),
				Opt("4", "Quest Log"),
				Opt("5", "View Full Character Stats"),
				Opt("6", "Resume Auto-Play"),
				Opt("0", "Return to Main Menu"),
			},
		}
	}
}

// handleQuestLog displays the quest log and returns to main menu on any input.
func (e *Engine) handleQuestLog(session *GameSession, cmd GameCommand) GameResponse {
	session.State = StateMainMenu
	return BuildMainMenuResponse(session)
}

// handlePlayerStats displays the player stats and returns to main menu on any input.
func (e *Engine) handlePlayerStats(session *GameSession, cmd GameCommand) GameResponse {
	gs := session.GameState
	player := session.Player

	gs.CharactersMap[player.Name] = *player
	game.WriteGameStateToFile(*gs, session.SaveFile)

	session.State = StateMainMenu
	return BuildMainMenuResponse(session)
}

// handleDiscoveredLocations returns to main menu on any input.
func (e *Engine) handleDiscoveredLocations(session *GameSession, cmd GameCommand) GameResponse {
	session.State = StateMainMenu
	return BuildMainMenuResponse(session)
}

// handleLoadSave loads the game state from file and shows character selection.
func (e *Engine) handleLoadSave(session *GameSession, cmd GameCommand) GameResponse {
	loaded, err := game.LoadGameStateFromFile(session.SaveFile)
	if err != nil {
		session.State = StateMainMenu
		resp := BuildMainMenuResponse(session)
		resp.Messages = append([]GameMessage{
			Msg(fmt.Sprintf("Error loading save: %s", err.Error()), "error"),
		}, resp.Messages...)
		return resp
	}

	session.GameState = &loaded
	if loaded.CharactersMap == nil {
		loaded.CharactersMap = make(map[string]models.Character)
	}
	if loaded.GameLocations == nil {
		loaded.GameLocations = make(map[string]models.Location)
	}

	// Initialize quest system if needed
	if loaded.AvailableQuests == nil {
		loaded.AvailableQuests = make(map[string]models.Quest)
		for id, quest := range data.StoryQuests {
			loaded.AvailableQuests[id] = quest
		}
	}

	session.State = StateLoadSaveCharSelect
	msgs := []GameMessage{
		Msg("Game loaded! Select a character:", "system"),
	}
	options := []MenuOption{}
	for name, char := range loaded.CharactersMap {
		options = append(options, Opt(name, fmt.Sprintf("%s (Lv%d)", name, char.Level)))
	}

	if len(options) == 0 {
		session.State = StateMainMenu
		resp := BuildMainMenuResponse(session)
		resp.Messages = append([]GameMessage{
			Msg("No characters found in save file!", "error"),
		}, resp.Messages...)
		return resp
	}

	return GameResponse{
		Type:     "menu",
		Messages: msgs,
		State:    &StateData{Screen: "load_save_char_select"},
		Options:  options,
	}
}

// handleLoadSaveCharSelect loads the selected character from the save file.
func (e *Engine) handleLoadSaveCharSelect(session *GameSession, cmd GameCommand) GameResponse {
	gs := session.GameState
	charName := cmd.Value

	char, exists := gs.CharactersMap[charName]
	if !exists {
		// Pick first available
		for _, c := range gs.CharactersMap {
			char = c
			break
		}
	}

	// Ensure resource types are up to date
	if char.ResourceStorageMap == nil {
		char.ResourceStorageMap = map[string]models.Resource{}
	}
	game.GenerateMissingResourceType(&char)

	if char.CompletedQuests == nil {
		char.CompletedQuests = []string{}
	}
	if char.ActiveQuests == nil {
		char.ActiveQuests = []string{"quest_1_training"}
	}
	if char.EquipmentMap == nil {
		char.EquipmentMap = map[int]models.Item{}
	}
	if char.LockedLocations == nil {
		char.LockedLocations = []string{}
	}

	gs.CharactersMap[char.Name] = char
	session.Player = &char

	// Ensure leaderboard entry exists for this character.
	e.saveSession(session)
	session.State = StateMainMenu
	resp := BuildMainMenuResponse(session)
	resp.Messages = append([]GameMessage{
		Msg(fmt.Sprintf("Loaded character: %s (Level %d)", char.Name, char.Level), "system"),
	}, resp.Messages...)
	return resp
}

// handleBuildSelect processes the building construction.
func (e *Engine) handleBuildSelect(session *GameSession, cmd GameCommand) GameResponse {
	player := session.Player
	gs := session.GameState
	buildingName := cmd.Value

	// Find the building in available buildings
	var targetBuilding *models.Building
	for _, b := range data.AvailableBuildings {
		if b.Name == buildingName {
			bCopy := b
			targetBuilding = &bCopy
			break
		}
	}

	if targetBuilding == nil {
		session.State = StateMainMenu
		resp := BuildMainMenuResponse(session)
		resp.Messages = append([]GameMessage{
			Msg(fmt.Sprintf("Building '%s' not found!", buildingName), "error"),
		}, resp.Messages...)
		return resp
	}

	// Check if already built
	for _, b := range player.BuiltBuildings {
		if b.Name == buildingName {
			session.State = StateMainMenu
			resp := BuildMainMenuResponse(session)
			resp.Messages = append([]GameMessage{
				Msg(fmt.Sprintf("'%s' is already built!", buildingName), "error"),
			}, resp.Messages...)
			return resp
		}
	}

	// Check resources
	canBuild := true
	missingMsgs := []GameMessage{}
	for resName, required := range targetBuilding.RequiredResourceMap {
		res, exists := player.ResourceStorageMap[resName]
		if !exists || res.Stock < required {
			canBuild = false
			have := 0
			if exists {
				have = res.Stock
			}
			missingMsgs = append(missingMsgs, Msg(fmt.Sprintf("Need %d %s (have %d)", required, resName, have), "error"))
		}
	}

	if !canBuild {
		session.State = StateMainMenu
		msgs := []GameMessage{Msg(fmt.Sprintf("Not enough resources to build %s!", buildingName), "error")}
		msgs = append(msgs, missingMsgs...)
		resp := BuildMainMenuResponse(session)
		resp.Messages = append(msgs, resp.Messages...)
		return resp
	}

	// Deduct resources
	for resName, required := range targetBuilding.RequiredResourceMap {
		res := player.ResourceStorageMap[resName]
		res.Stock -= required
		player.ResourceStorageMap[resName] = res
	}

	// Add to built buildings
	if player.BuiltBuildings == nil {
		player.BuiltBuildings = []models.Building{}
	}
	player.BuiltBuildings = append(player.BuiltBuildings, *targetBuilding)

	// Recalculate player stats with building bonuses
	player.StatsMod = game.CalculateItemMods(player.EquipmentMap)
	for _, b := range player.BuiltBuildings {
		player.StatsMod.AttackMod += b.StatsMod.AttackMod
		player.StatsMod.DefenseMod += b.StatsMod.DefenseMod
		player.StatsMod.HitPointMod += b.StatsMod.HitPointMod
	}
	player.HitpointsTotal = player.HitpointsNatural + player.StatsMod.HitPointMod

	gs.CharactersMap[player.Name] = *player
	game.WriteGameStateToFile(*gs, session.SaveFile)

	session.State = StateMainMenu
	resp := BuildMainMenuResponse(session)
	resp.Messages = append([]GameMessage{
		Msg(fmt.Sprintf("Built %s!", buildingName), "system"),
	}, resp.Messages...)
	return resp
}

// --- Helper functions for building display messages ---

func buildPlayerStatsMessages(player *models.Character, gs *models.GameState) []GameMessage {
	msgs := []GameMessage{
		Msg("============ Player Stats ============", "system"),
		Msg(fmt.Sprintf("Name: %s", player.Name), "system"),
		Msg(fmt.Sprintf("Level: %d", player.Level), "system"),
		Msg(fmt.Sprintf("Experience: %d / %d", player.Experience, player.Level*100), "system"),
		Msg(fmt.Sprintf("HP: %d/%d (Natural: %d)", player.HitpointsRemaining, player.HitpointsTotal, player.HitpointsNatural), "system"),
		Msg(fmt.Sprintf("MP: %d/%d (Natural: %d)", player.ManaRemaining, player.ManaTotal, player.ManaNatural), "system"),
		Msg(fmt.Sprintf("SP: %d/%d (Natural: %d)", player.StaminaRemaining, player.StaminaTotal, player.StaminaNatural), "system"),
		Msg(fmt.Sprintf("Attack Rolls: %d", player.AttackRolls), "system"),
		Msg(fmt.Sprintf("Defense Rolls: %d", player.DefenseRolls), "system"),
		Msg(fmt.Sprintf("Attack Mod: +%d", player.StatsMod.AttackMod), "system"),
		Msg(fmt.Sprintf("Defense Mod: +%d", player.StatsMod.DefenseMod), "system"),
		Msg(fmt.Sprintf("HP Mod: +%d", player.StatsMod.HitPointMod), "system"),
		Msg(fmt.Sprintf("Resurrections: %d", player.Resurrections), "system"),
	}

	// Equipment
	if len(player.EquipmentMap) > 0 {
		msgs = append(msgs, Msg("", "system"))
		msgs = append(msgs, Msg("--- Equipped Items ---", "system"))
		for slot, item := range player.EquipmentMap {
			slotName := SlotNames[slot]
			if slotName == "" {
				slotName = fmt.Sprintf("Slot %d", slot)
			}
			msgs = append(msgs, Msg(fmt.Sprintf("  [%s] %s (Rarity %d, CP:%d)",
				slotName, item.Name, item.Rarity, item.CP), "system"))
		}
	}

	// Skills
	if len(player.LearnedSkills) > 0 {
		msgs = append(msgs, Msg("", "system"))
		msgs = append(msgs, Msg("--- Learned Skills ---", "system"))
		for _, skill := range player.LearnedSkills {
			skillLabel := skill.Name
			if skill.UpgradeCount > 0 {
				skillLabel += fmt.Sprintf(" +%d", skill.UpgradeCount)
			}
			costStr := ""
			if skill.ManaCost > 0 {
				costStr += fmt.Sprintf("%dMP ", skill.ManaCost)
			}
			if skill.StaminaCost > 0 {
				costStr += fmt.Sprintf("%dSP ", skill.StaminaCost)
			}
			msgs = append(msgs, Msg(fmt.Sprintf("  %s [%s] - %s", skillLabel, costStr, skill.Description), "system"))
		}
	}

	// Known Locations
	if len(player.KnownLocations) > 0 {
		msgs = append(msgs, Msg("", "system"))
		msgs = append(msgs, Msg("--- Known Locations ---", "system"))
		for _, loc := range player.KnownLocations {
			msgs = append(msgs, Msg(fmt.Sprintf("  %s", loc), "system"))
		}
	}

	// Resources
	msgs = append(msgs, Msg("", "system"))
	msgs = append(msgs, Msg("--- Resources ---", "system"))
	for _, res := range data.ResourceTypes {
		r, exists := player.ResourceStorageMap[res]
		if exists {
			msgs = append(msgs, Msg(fmt.Sprintf("  %s: %d", res, r.Stock), "system"))
		}
	}
	for _, matName := range data.BeastMaterials {
		r, exists := player.ResourceStorageMap[matName]
		if exists && r.Stock > 0 {
			msgs = append(msgs, Msg(fmt.Sprintf("  %s: %d", matName, r.Stock), "system"))
		}
	}

	// Built Buildings
	if len(player.BuiltBuildings) > 0 {
		msgs = append(msgs, Msg("", "system"))
		msgs = append(msgs, Msg("--- Built Buildings ---", "system"))
		for _, b := range player.BuiltBuildings {
			msgs = append(msgs, Msg(fmt.Sprintf("  %s", b.Name), "system"))
		}
	}

	msgs = append(msgs, Msg("======================================", "system"))

	return msgs
}

func buildQuestLogMessages(player *models.Character, gs *models.GameState) []GameMessage {
	if gs.AvailableQuests == nil {
		gs.AvailableQuests = make(map[string]models.Quest)
		for k, v := range data.StoryQuests {
			gs.AvailableQuests[k] = v
		}
	}
	if player.CompletedQuests == nil {
		player.CompletedQuests = []string{}
	}
	if player.ActiveQuests == nil {
		player.ActiveQuests = []string{"quest_1_training"}
	}

	msgs := []GameMessage{
		Msg("====== QUEST LOG ======", "system"),
		Msg("", "system"),
		Msg("Active Quests:", "system"),
	}

	if len(player.ActiveQuests) == 0 {
		msgs = append(msgs, Msg("  No active quests.", "narrative"))
	}
	for _, questID := range player.ActiveQuests {
		quest, exists := gs.AvailableQuests[questID]
		if !exists {
			continue
		}

		// Update current values for display
		switch quest.Requirement.Type {
		case "level":
			quest.Requirement.CurrentValue = player.Level
		case "village_level":
			if gs.Villages != nil && player.VillageName != "" {
				if village, ok := gs.Villages[player.VillageName]; ok {
					quest.Requirement.CurrentValue = village.Level
				}
			}
		case "total_resources":
			total := 0
			for _, res := range player.ResourceStorageMap {
				total += res.Stock
			}
			quest.Requirement.CurrentValue = total
		case "skill_count":
			quest.Requirement.CurrentValue = len(player.LearnedSkills)
		}
		gs.AvailableQuests[questID] = quest

		msgs = append(msgs, Msg(fmt.Sprintf("  > %s", quest.Name), "narrative"))
		msgs = append(msgs, Msg(fmt.Sprintf("    %s", quest.Description), "narrative"))

		switch quest.Requirement.Type {
		case "level":
			msgs = append(msgs, Msg(fmt.Sprintf("    Progress: Level %d / %d",
				quest.Requirement.CurrentValue, quest.Requirement.TargetValue), "narrative"))
		case "boss_kill":
			msgs = append(msgs, Msg(fmt.Sprintf("    Progress: %d / %d %s defeated",
				quest.Requirement.CurrentValue, quest.Requirement.TargetValue, quest.Requirement.TargetName), "narrative"))
		case "location":
			msgs = append(msgs, Msg(fmt.Sprintf("    Progress: %d / %d locations explored in %s",
				quest.Requirement.CurrentValue, quest.Requirement.TargetValue, quest.Requirement.TargetName), "narrative"))
		case "village_level":
			msgs = append(msgs, Msg(fmt.Sprintf("    Progress: Village Level %d / %d",
				quest.Requirement.CurrentValue, quest.Requirement.TargetValue), "narrative"))
		case "total_resources":
			msgs = append(msgs, Msg(fmt.Sprintf("    Progress: %d / %d resources gathered",
				quest.Requirement.CurrentValue, quest.Requirement.TargetValue), "narrative"))
		case "skill_count":
			msgs = append(msgs, Msg(fmt.Sprintf("    Progress: %d / %d skills learned",
				quest.Requirement.CurrentValue, quest.Requirement.TargetValue), "narrative"))
		case "elder_rescued":
			msgs = append(msgs, Msg(fmt.Sprintf("    Progress: %d / %d elders rescued",
				quest.Requirement.CurrentValue, quest.Requirement.TargetValue), "narrative"))
		}

		msgs = append(msgs, Msg(fmt.Sprintf("    Reward: %d XP", quest.Reward.XP), "narrative"))
	}

	msgs = append(msgs, Msg("", "system"))
	msgs = append(msgs, Msg("Completed Quests:", "system"))

	if len(player.CompletedQuests) == 0 {
		msgs = append(msgs, Msg("  No completed quests yet.", "narrative"))
	}
	for _, questID := range player.CompletedQuests {
		quest, exists := gs.AvailableQuests[questID]
		if !exists {
			continue
		}
		msgs = append(msgs, Msg(fmt.Sprintf("  [DONE] %s", quest.Name), "narrative"))
	}

	msgs = append(msgs, Msg("========================", "system"))

	return msgs
}

func buildInventoryMessages(player *models.Character) []GameMessage {
	msgs := []GameMessage{
		Msg("============================================================", "system"),
		Msg(fmt.Sprintf("INVENTORY - %s", player.Name), "system"),
		Msg("============================================================", "system"),
	}

	if len(player.Inventory) == 0 {
		msgs = append(msgs, Msg("Your inventory is empty.", "system"))
	} else {
		consumables := []models.Item{}
		skillScrolls := []models.Item{}
		equipment := []models.Item{}

		for _, item := range player.Inventory {
			switch item.ItemType {
			case "consumable":
				consumables = append(consumables, item)
			case "skill_scroll":
				skillScrolls = append(skillScrolls, item)
			default:
				equipment = append(equipment, item)
			}
		}

		if len(consumables) > 0 {
			msgs = append(msgs, Msg("", "system"))
			msgs = append(msgs, Msg("CONSUMABLES:", "system"))
			potionCount := make(map[string]int)
			for _, item := range consumables {
				potionCount[item.Name]++
			}
			for name, count := range potionCount {
				msgs = append(msgs, Msg(fmt.Sprintf("  %s x%d", name, count), "system"))
			}
		}

		if len(skillScrolls) > 0 {
			msgs = append(msgs, Msg("", "system"))
			msgs = append(msgs, Msg("SKILL SCROLLS:", "system"))
			for i, scroll := range skillScrolls {
				msgs = append(msgs, Msg(fmt.Sprintf("  %d. %s (Skill: %s, Crafting Value: %d)",
					i+1, scroll.Name, scroll.SkillScroll.Skill.Name, scroll.SkillScroll.CraftingValue), "system"))
			}
		}

		if len(equipment) > 0 {
			msgs = append(msgs, Msg("", "system"))
			msgs = append(msgs, Msg("EQUIPMENT (Unequipped):", "system"))
			for i, item := range equipment {
				msgs = append(msgs, Msg(fmt.Sprintf("  %d. %s (Rarity %d, CP: %d)",
					i+1, item.Name, item.Rarity, item.CP), "system"))
			}
		}
	}

	msgs = append(msgs, Msg(fmt.Sprintf("\nTotal Items: %d", len(player.Inventory)), "system"))
	msgs = append(msgs, Msg("============================================================", "system"))

	return msgs
}

func buildSkillsMessages(player *models.Character) []GameMessage {
	msgs := []GameMessage{
		Msg("============================================================", "system"),
		Msg(fmt.Sprintf("LEARNED SKILLS - %s", player.Name), "system"),
		Msg("============================================================", "system"),
		Msg(fmt.Sprintf("Level: %d | MP: %d/%d | SP: %d/%d",
			player.Level, player.ManaRemaining, player.ManaTotal,
			player.StaminaRemaining, player.StaminaTotal), "system"),
	}

	if len(player.LearnedSkills) == 0 {
		msgs = append(msgs, Msg("No skills learned yet.", "system"))
	} else {
		for i, skill := range player.LearnedSkills {
			msgs = append(msgs, Msg("", "system"))
			skillLabel := skill.Name
			if skill.UpgradeCount > 0 {
				skillLabel += fmt.Sprintf(" +%d", skill.UpgradeCount)
			}
			msgs = append(msgs, Msg(fmt.Sprintf("%d. %s", i+1, skillLabel), "system"))

			if skill.ManaCost > 0 {
				msgs = append(msgs, Msg(fmt.Sprintf("   Cost: %d MP", skill.ManaCost), "system"))
			}
			if skill.StaminaCost > 0 {
				msgs = append(msgs, Msg(fmt.Sprintf("   Cost: %d SP", skill.StaminaCost), "system"))
			}
			if skill.Damage > 0 {
				msgs = append(msgs, Msg(fmt.Sprintf("   Damage: %d %s", skill.Damage, skill.DamageType), "system"))
			} else if skill.Damage < 0 {
				msgs = append(msgs, Msg(fmt.Sprintf("   Healing: %d HP", -skill.Damage), "system"))
			}
			if skill.Effect.Type != "none" && skill.Effect.Type != "" {
				msgs = append(msgs, Msg(fmt.Sprintf("   Effect: %s (%d turns, potency %d)",
					skill.Effect.Type, skill.Effect.Duration, skill.Effect.Potency), "system"))
			}
			msgs = append(msgs, Msg(fmt.Sprintf("   %s", skill.Description), "system"))
		}
	}

	msgs = append(msgs, Msg("", "system"))
	msgs = append(msgs, Msg(fmt.Sprintf("Total Skills: %d", len(player.LearnedSkills)), "system"))
	msgs = append(msgs, Msg("============================================================", "system"))

	return msgs
}

func buildEquipmentMessages(player *models.Character) []GameMessage {
	msgs := []GameMessage{
		Msg("============================================================", "system"),
		Msg(fmt.Sprintf("EQUIPPED ITEMS - %s", player.Name), "system"),
		Msg("============================================================", "system"),
	}

	if len(player.EquipmentMap) == 0 {
		msgs = append(msgs, Msg("No equipment equipped.", "system"))
	} else {
		for slot, item := range player.EquipmentMap {
			slotName := SlotNames[slot]
			if slotName == "" {
				slotName = fmt.Sprintf("Slot %d", slot)
			}
			msgs = append(msgs, Msg(fmt.Sprintf("[%s] %s (Rarity %d, CP: %d)",
				slotName, item.Name, item.Rarity, item.CP), "system"))
			if item.StatsMod.AttackMod > 0 {
				msgs = append(msgs, Msg(fmt.Sprintf("  +%d Attack", item.StatsMod.AttackMod), "system"))
			}
			if item.StatsMod.DefenseMod > 0 {
				msgs = append(msgs, Msg(fmt.Sprintf("  +%d Defense", item.StatsMod.DefenseMod), "system"))
			}
			if item.StatsMod.HitPointMod > 0 {
				msgs = append(msgs, Msg(fmt.Sprintf("  +%d HP", item.StatsMod.HitPointMod), "system"))
			}
		}

		msgs = append(msgs, Msg("", "system"))
		msgs = append(msgs, Msg("Total Stats from Equipment:", "system"))
		msgs = append(msgs, Msg(fmt.Sprintf("  Attack:  +%d", player.StatsMod.AttackMod), "system"))
		msgs = append(msgs, Msg(fmt.Sprintf("  Defense: +%d", player.StatsMod.DefenseMod), "system"))
		msgs = append(msgs, Msg(fmt.Sprintf("  HP:      +%d", player.StatsMod.HitPointMod), "system"))
	}

	msgs = append(msgs, Msg(fmt.Sprintf("\nEquipped Items: %d", len(player.EquipmentMap)), "system"))
	msgs = append(msgs, Msg("============================================================", "system"))

	return msgs
}
