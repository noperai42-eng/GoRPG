package engine

import (
	"fmt"
	"math/rand"
	"sort"
	"strconv"
	"time"

	"rpg-game/pkg/data"
	"rpg-game/pkg/game"
	"rpg-game/pkg/models"
)

// saveVillage persists the village back into the game state and writes to disk.
func (e *Engine) saveVillage(session *GameSession) {
	if session.SelectedVillage != nil && session.Player != nil {
		session.GameState.Villages[session.Player.VillageName] = *session.SelectedVillage
		session.GameState.CharactersMap[session.Player.Name] = *session.Player
		game.WriteGameStateToFile(*session.GameState, session.SaveFile)
	}
}

// buildVillageMainResponse creates the village main menu response.
func buildVillageMainResponse(session *GameSession, extraMsgs []GameMessage) GameResponse {
	village := session.SelectedVillage
	player := session.Player

	msgs := append([]GameMessage{}, extraMsgs...)
	msgs = append(msgs,
		Msg("============================================================", "system"),
		Msg(fmt.Sprintf("  %s - Level %d", village.Name, village.Level), "system"),
		Msg("============================================================", "system"),
		Msg(fmt.Sprintf("Experience: %d/%d", village.Experience, village.Level*100), "system"),
		Msg(fmt.Sprintf("Villagers: %d (Harvesters: %d, Guards: %d)",
			len(village.Villagers),
			game.CountVillagersByRole(village, "harvester"),
			game.CountVillagersByRole(village, "guard")), "system"),
		Msg(fmt.Sprintf("Hired Guards: %d", len(village.ActiveGuards)), "system"),
		Msg(fmt.Sprintf("Defenses Built: %d (Level %d)", len(village.Defenses), village.DefenseLevel), "system"),
	)

	if len(village.UnlockedCrafting) > 0 {
		craftStr := "Unlocked Crafting: "
		for i, craft := range village.UnlockedCrafting {
			if i > 0 {
				craftStr += ", "
			}
			craftStr += craft
		}
		msgs = append(msgs, Msg(craftStr, "system"))
	}

	options := []MenuOption{
		Opt("1", "View Villagers"),
		Opt("2", "Assign Harvester Tasks"),
		Opt("3", "Hire Guards"),
		Opt("4", "Crafting"),
		Opt("5", "Build Defenses"),
		Opt("6", "Check Next Monster Tide"),
		Opt("7", "Defend Against Tide (if ready)"),
		Opt("8", "Manage Guards (Equipment & Status)"),
		Opt("0", "Return to Main Menu"),
	}

	return GameResponse{
		Type:     "menu",
		Messages: msgs,
		State:    &StateData{Screen: "village_main", Player: MakePlayerState(player), Village: MakeVillageView(village)},
		Options:  options,
	}
}

func (e *Engine) handleVillageMain(session *GameSession, cmd GameCommand) GameResponse {
	village := session.SelectedVillage

	if village == nil {
		session.State = StateMainMenu
		return BuildMainMenuResponse(session)
	}

	// Process village upgrades (harvesting is now timer-based via WebSocket)
	game.UpgradeVillage(village)

	extraMsgs := []GameMessage{}

	switch cmd.Value {
	case "1":
		session.State = StateVillageViewVillagers
		return e.handleVillageViewVillagers(session, GameCommand{Type: "init"})
	case "2":
		session.State = StateVillageAssignTask
		return e.handleVillageAssignTask(session, GameCommand{Type: "init"})
	case "3":
		session.State = StateVillageHireGuard
		return e.handleVillageHireGuard(session, GameCommand{Type: "init"})
	case "4":
		session.State = StateVillageCrafting
		return e.handleVillageCrafting(session, GameCommand{Type: "init"})
	case "5":
		session.State = StateVillageBuildDefense
		return e.handleVillageBuildDefense(session, GameCommand{Type: "init"})
	case "6":
		session.State = StateVillageCheckTide
		return e.handleVillageCheckTide(session, GameCommand{Type: "init"})
	case "7":
		// Check if tide is ready
		currentTime := time.Now().Unix()
		timeSinceLastTide := currentTime - village.LastTideTime
		timeUntilNext := village.TideInterval - int(timeSinceLastTide)

		if timeUntilNext <= 0 {
			session.State = StateVillageMonsterTide
			return e.handleVillageMonsterTide(session, GameCommand{Type: "init"})
		}
		hours := timeUntilNext / 3600
		minutes := (timeUntilNext % 3600) / 60
		extraMsgs = append(extraMsgs, Msg(fmt.Sprintf("Tide not ready yet! Wait %d hours, %d minutes", hours, minutes), "system"))

	case "8":
		session.State = StateVillageManageGuards
		return e.handleVillageManageGuards(session, GameCommand{Type: "init"})
	case "0":
		e.saveVillage(session)
		session.SelectedVillage = nil
		session.State = StateMainMenu
		resp := BuildMainMenuResponse(session)
		resp.Messages = append([]GameMessage{Msg("Village saved", "system")}, resp.Messages...)
		return resp
	}

	session.State = StateVillageMain
	return buildVillageMainResponse(session, extraMsgs)
}

func (e *Engine) handleVillageViewVillagers(session *GameSession, cmd GameCommand) GameResponse {
	village := session.SelectedVillage

	if cmd.Value == "back" || cmd.Value == "0" {
		session.State = StateVillageMain
		return e.handleVillageMain(session, GameCommand{Type: "init"})
	}

	msgs := []GameMessage{
		Msg("============================================================", "system"),
		Msg("VILLAGERS", "system"),
		Msg("============================================================", "system"),
	}

	if len(village.Villagers) == 0 {
		msgs = append(msgs, Msg("No villagers yet. Rescue them during hunts!", "narrative"))
	} else {
		harvesters := []models.Villager{}
		guards := []models.Villager{}

		for _, v := range village.Villagers {
			if v.Role == "harvester" {
				harvesters = append(harvesters, v)
			} else {
				guards = append(guards, v)
			}
		}

		if len(harvesters) > 0 {
			msgs = append(msgs, Msg("", "system"))
			msgs = append(msgs, Msg("HARVESTERS:", "system"))
			for i, v := range harvesters {
				taskInfo := "Idle"
				if v.HarvestType != "" {
					taskInfo = fmt.Sprintf("Harvesting %s (+%d/visit)", v.HarvestType, v.Efficiency+(v.Level/2))
				}
				msgs = append(msgs, Msg(fmt.Sprintf("  %d. %s (Lv%d) - %s", i+1, v.Name, v.Level, taskInfo), "system"))
			}
		}

		if len(guards) > 0 {
			msgs = append(msgs, Msg("", "system"))
			msgs = append(msgs, Msg("GUARDS:", "system"))
			for i, v := range guards {
				msgs = append(msgs, Msg(fmt.Sprintf("  %d. %s (Lv%d) - Efficiency: %d", i+1, v.Name, v.Level, v.Efficiency), "system"))
			}
		}
	}

	msgs = append(msgs, Msg("============================================================", "system"))

	session.State = StateVillageViewVillagers
	return GameResponse{
		Type:     "menu",
		Messages: msgs,
		State:    &StateData{Screen: "village_view_villagers", Player: MakePlayerState(session.Player)},
		Options:  []MenuOption{Opt("back", "Back to Village")},
	}
}

func (e *Engine) handleVillageAssignTask(session *GameSession, cmd GameCommand) GameResponse {
	village := session.SelectedVillage

	if cmd.Value == "0" || cmd.Value == "back" {
		session.State = StateVillageMain
		return e.handleVillageMain(session, GameCommand{Type: "init"})
	}

	// Build list of harvesters
	harvesters := []int{}
	for i, v := range village.Villagers {
		if v.Role == "harvester" {
			harvesters = append(harvesters, i)
		}
	}

	if len(harvesters) == 0 {
		session.State = StateVillageMain
		resp := e.handleVillageMain(session, GameCommand{Type: "init"})
		resp.Messages = append([]GameMessage{Msg("No harvesters available!", "error")}, resp.Messages...)
		return resp
	}

	// Count idle harvesters for batch assign option
	idleCount := 0
	for _, idx := range harvesters {
		if village.Villagers[idx].HarvestType == "" {
			idleCount++
		}
	}

	// If cmd.Type is "init", show the harvester selection
	if cmd.Type == "init" {
		msgs := []GameMessage{
			Msg("============================================================", "system"),
			Msg("ASSIGN HARVESTER TASK", "system"),
			Msg("============================================================", "system"),
			Msg("", "system"),
			Msg("Available Harvesters:", "system"),
		}

		options := []MenuOption{}
		if idleCount > 0 {
			options = append(options, Opt("all", fmt.Sprintf("Assign All Idle Harvesters (%d)", idleCount)))
		}
		for i, idx := range harvesters {
			v := village.Villagers[idx]
			taskInfo := "Idle"
			if v.HarvestType != "" {
				taskInfo = "Currently: " + v.HarvestType
			}
			label := fmt.Sprintf("%s (Lv%d, Efficiency %d) - %s", v.Name, v.Level, v.Efficiency, taskInfo)
			options = append(options, Opt(strconv.Itoa(i+1), label))
		}
		options = append(options, Opt("0", "Cancel"))

		session.State = StateVillageAssignTask
		return GameResponse{
			Type:     "menu",
			Messages: msgs,
			State:    &StateData{Screen: "village_assign_task", Player: MakePlayerState(session.Player)},
			Options:  options,
		}
	}

	// Batch assign all idle harvesters
	if cmd.Value == "all" {
		if idleCount == 0 {
			return ErrorResponse("No idle harvesters to assign!")
		}
		session.State = StateVillageBatchAssign
		return e.handleVillageBatchAssign(session, GameCommand{Type: "init"})
	}

	// A harvester was selected - store index and move to resource selection
	idx, err := strconv.Atoi(cmd.Value)
	if err != nil || idx < 1 || idx > len(harvesters) {
		return ErrorResponse("Invalid choice!")
	}

	session.SelectedVillagerIdx = harvesters[idx-1]
	session.State = StateVillageAssignResource
	return e.handleVillageAssignResource(session, GameCommand{Type: "init"})
}

func (e *Engine) handleVillageAssignResource(session *GameSession, cmd GameCommand) GameResponse {
	village := session.SelectedVillage

	if cmd.Value == "0" || cmd.Value == "back" {
		session.State = StateVillageAssignTask
		return e.handleVillageAssignTask(session, GameCommand{Type: "init"})
	}

	if cmd.Type == "init" {
		msgs := []GameMessage{
			Msg("Assign to resource:", "system"),
		}
		options := []MenuOption{}
		for i, res := range data.ResourceTypes {
			options = append(options, Opt(strconv.Itoa(i+1), res))
		}
		options = append(options, Opt("0", "Cancel"))

		session.State = StateVillageAssignResource
		return GameResponse{
			Type:     "menu",
			Messages: msgs,
			State:    &StateData{Screen: "village_assign_resource", Player: MakePlayerState(session.Player)},
			Options:  options,
		}
	}

	resIdx, err := strconv.Atoi(cmd.Value)
	if err != nil || resIdx < 1 || resIdx > len(data.ResourceTypes) {
		return ErrorResponse("Invalid choice!")
	}

	villagerIdx := session.SelectedVillagerIdx
	village.Villagers[villagerIdx].HarvestType = data.ResourceTypes[resIdx-1]
	village.Villagers[villagerIdx].AssignedTask = "harvesting"
	village.Experience += 10

	e.saveVillage(session)

	session.State = StateVillageMain
	resp := e.handleVillageMain(session, GameCommand{Type: "init"})
	resp.Messages = append([]GameMessage{
		Msg(fmt.Sprintf("%s is now harvesting %s!", village.Villagers[villagerIdx].Name, data.ResourceTypes[resIdx-1]), "system"),
		Msg("+10 Village XP", "system"),
	}, resp.Messages...)
	return resp
}

func (e *Engine) handleVillageBatchAssign(session *GameSession, cmd GameCommand) GameResponse {
	village := session.SelectedVillage

	if cmd.Value == "0" || cmd.Value == "back" {
		session.State = StateVillageAssignTask
		return e.handleVillageAssignTask(session, GameCommand{Type: "init"})
	}

	if cmd.Type == "init" {
		msgs := []GameMessage{
			Msg("============================================================", "system"),
			Msg("ASSIGN ALL IDLE HARVESTERS", "system"),
			Msg("============================================================", "system"),
			Msg("", "system"),
			Msg("Choose a resource for all idle harvesters:", "system"),
		}
		options := []MenuOption{}
		for i, res := range data.ResourceTypes {
			options = append(options, Opt(strconv.Itoa(i+1), res))
		}
		options = append(options, Opt("0", "Cancel"))

		session.State = StateVillageBatchAssign
		return GameResponse{
			Type:     "menu",
			Messages: msgs,
			State:    &StateData{Screen: "village_batch_assign", Player: MakePlayerState(session.Player)},
			Options:  options,
		}
	}

	resIdx, err := strconv.Atoi(cmd.Value)
	if err != nil || resIdx < 1 || resIdx > len(data.ResourceTypes) {
		return ErrorResponse("Invalid choice!")
	}

	resourceName := data.ResourceTypes[resIdx-1]
	assigned := 0
	for i := range village.Villagers {
		if village.Villagers[i].Role == "harvester" && village.Villagers[i].HarvestType == "" {
			village.Villagers[i].HarvestType = resourceName
			village.Villagers[i].AssignedTask = "harvesting"
			village.Experience += 10
			assigned++
		}
	}

	e.saveVillage(session)

	session.State = StateVillageMain
	resp := e.handleVillageMain(session, GameCommand{Type: "init"})
	resp.Messages = append([]GameMessage{
		Msg(fmt.Sprintf("Assigned %d idle harvesters to %s!", assigned, resourceName), "system"),
		Msg(fmt.Sprintf("+%d Village XP", assigned*10), "system"),
	}, resp.Messages...)
	return resp
}

func (e *Engine) handleVillageHireGuard(session *GameSession, cmd GameCommand) GameResponse {
	village := session.SelectedVillage
	player := session.Player

	if cmd.Value == "0" || cmd.Value == "back" {
		session.State = StateVillageMain
		return e.handleVillageMain(session, GameCommand{Type: "init"})
	}

	if cmd.Type == "init" {
		goldResource := player.ResourceStorageMap["Gold"]

		msgs := []GameMessage{
			Msg("============================================================", "system"),
			Msg("GUARD RECRUITMENT", "system"),
			Msg("============================================================", "system"),
			Msg(fmt.Sprintf("Your Gold: %d", goldResource.Stock), "system"),
			Msg("", "system"),
			Msg("Available Guards for Hire:", "system"),
		}

		// Generate 3 guards
		guards := []models.Guard{
			game.GenerateGuard(village.Level),
			game.GenerateGuard(village.Level + 2),
			game.GenerateGuard(village.Level + 5),
		}

		// Store them on session for selection
		session.Combat = &CombatContext{}

		options := []MenuOption{}
		for i, guard := range guards {
			msgs = append(msgs, Msg("", "system"))
			msgs = append(msgs, Msg(fmt.Sprintf("%d. %s (Level %d)", i+1, guard.Name, guard.Level), "system"))
			msgs = append(msgs, Msg(fmt.Sprintf("   HP: %d | Attack Rolls: %d | Defense Rolls: %d",
				guard.HitPoints, guard.AttackRolls, guard.DefenseRolls), "system"))
			msgs = append(msgs, Msg(fmt.Sprintf("   Equipment Bonus: +%d ATK, +%d DEF, +%d HP",
				guard.StatsMod.AttackMod, guard.StatsMod.DefenseMod, guard.StatsMod.HitPointMod), "system"))
			msgs = append(msgs, Msg(fmt.Sprintf("   Starting Equipment: %d items", len(guard.EquipmentMap)), "system"))
			msgs = append(msgs, Msg(fmt.Sprintf("   Cost: %d Gold", guard.Cost), "system"))

			options = append(options, Opt(fmt.Sprintf("hire_%d_%d", i, guard.Cost), fmt.Sprintf("Hire %s (%d Gold)", guard.Name, guard.Cost)))
		}
		options = append(options, Opt("0", "Cancel"))

		// We need to store the generated guards so we can reference them on selection.
		// Store guard data encoded in the option keys.
		// Actually, regenerate on selection since guards are random anyway.
		// Store the three levels used.

		session.State = StateVillageHireGuard
		return GameResponse{
			Type:     "menu",
			Messages: msgs,
			State:    &StateData{Screen: "village_hire_guard", Player: MakePlayerState(player)},
			Options:  options,
		}
	}

	// Parse selection: "hire_<idx>_<cost>" or a number
	idx, err := strconv.Atoi(cmd.Value)
	if err == nil && idx >= 1 && idx <= 3 {
		// Simple numeric selection
		levels := []int{village.Level, village.Level + 2, village.Level + 5}
		selectedGuard := game.GenerateGuard(levels[idx-1])
		goldResource := player.ResourceStorageMap["Gold"]

		if goldResource.Stock < selectedGuard.Cost {
			session.State = StateVillageHireGuard
			return GameResponse{
				Type:     "menu",
				Messages: []GameMessage{Msg(fmt.Sprintf("Not enough gold! Need %d, have %d", selectedGuard.Cost, goldResource.Stock), "error")},
				State:    &StateData{Screen: "village_hire_guard", Player: MakePlayerState(player)},
				Options:  []MenuOption{Opt("back", "Back")},
			}
		}

		goldResource.Stock -= selectedGuard.Cost
		player.ResourceStorageMap["Gold"] = goldResource
		selectedGuard.Hired = true
		village.ActiveGuards = append(village.ActiveGuards, selectedGuard)
		village.Experience += 50

		e.saveVillage(session)

		session.State = StateVillageMain
		resp := e.handleVillageMain(session, GameCommand{Type: "init"})
		resp.Messages = append([]GameMessage{
			Msg(fmt.Sprintf("Hired %s for %d Gold!", selectedGuard.Name, selectedGuard.Cost), "system"),
			Msg("They will assist in guardian and boss fights!", "system"),
			Msg("+50 Village XP", "system"),
		}, resp.Messages...)
		return resp
	}

	// Try parsing hire_<idx>_<cost> format
	if len(cmd.Value) > 5 && cmd.Value[:5] == "hire_" {
		// Extract the index
		parts := cmd.Value[5:]
		underscoreIdx := 0
		for i, c := range parts {
			if c == '_' {
				underscoreIdx = i
				break
			}
		}
		if underscoreIdx > 0 {
			guardIdx, err2 := strconv.Atoi(parts[:underscoreIdx])
			if err2 == nil && guardIdx >= 0 && guardIdx <= 2 {
				levels := []int{village.Level, village.Level + 2, village.Level + 5}
				selectedGuard := game.GenerateGuard(levels[guardIdx])
				goldResource := player.ResourceStorageMap["Gold"]

				if goldResource.Stock < selectedGuard.Cost {
					session.State = StateVillageHireGuard
					return GameResponse{
						Type:     "menu",
						Messages: []GameMessage{Msg(fmt.Sprintf("Not enough gold! Need %d, have %d", selectedGuard.Cost, goldResource.Stock), "error")},
						State:    &StateData{Screen: "village_hire_guard", Player: MakePlayerState(player)},
						Options:  []MenuOption{Opt("back", "Back")},
					}
				}

				goldResource.Stock -= selectedGuard.Cost
				player.ResourceStorageMap["Gold"] = goldResource
				selectedGuard.Hired = true
				village.ActiveGuards = append(village.ActiveGuards, selectedGuard)
				village.Experience += 50

				e.saveVillage(session)

				session.State = StateVillageMain
				resp := e.handleVillageMain(session, GameCommand{Type: "init"})
				resp.Messages = append([]GameMessage{
					Msg(fmt.Sprintf("Hired %s for %d Gold!", selectedGuard.Name, selectedGuard.Cost), "system"),
					Msg("They will assist in guardian and boss fights!", "system"),
					Msg("+50 Village XP", "system"),
				}, resp.Messages...)
				return resp
			}
		}
	}

	session.State = StateVillageMain
	return e.handleVillageMain(session, GameCommand{Type: "init"})
}

func (e *Engine) handleVillageCrafting(session *GameSession, cmd GameCommand) GameResponse {
	village := session.SelectedVillage

	if cmd.Value == "0" || cmd.Value == "back" {
		session.State = StateVillageMain
		return e.handleVillageMain(session, GameCommand{Type: "init"})
	}

	if len(village.UnlockedCrafting) == 0 {
		msgs := []GameMessage{
			Msg("No crafting unlocked yet!", "system"),
			Msg("Level up your village to unlock crafting:", "system"),
			Msg("  Level 3  -> Potion Crafting", "system"),
			Msg("  Level 5  -> Armor Crafting", "system"),
			Msg("  Level 7  -> Weapon Crafting", "system"),
			Msg("  Level 10 -> Skill Upgrades", "system"),
			Msg("  Level 10 -> Skill Scroll Crafting", "system"),
		}
		session.State = StateVillageCrafting
		return GameResponse{
			Type:     "menu",
			Messages: msgs,
			State:    &StateData{Screen: "village_crafting", Player: MakePlayerState(session.Player)},
			Options:  []MenuOption{Opt("back", "Back to Village")},
		}
	}

	// Route to specific crafting handlers based on selection
	switch cmd.Value {
	case "potions":
		session.State = StateVillageCraftPotion
		return e.handleVillageCraftPotion(session, GameCommand{Type: "init"})
	case "armor":
		session.State = StateVillageCraftArmor
		return e.handleVillageCraftArmor(session, GameCommand{Type: "init"})
	case "weapons":
		session.State = StateVillageCraftWeapon
		return e.handleVillageCraftWeapon(session, GameCommand{Type: "init"})
	case "skill_upgrades":
		session.State = StateVillageUpgradeSkill
		return e.handleVillageUpgradeSkill(session, GameCommand{Type: "init"})
	case "skill_scrolls":
		session.State = StateVillageCraftScrolls
		return e.handleVillageCraftScrolls(session, GameCommand{Type: "init"})
	case "fortifications":
		session.State = StateVillageFortifications
		return e.handleVillageFortifications(session, GameCommand{Type: "init"})
	case "training":
		session.State = StateVillageTraining
		return e.handleVillageTraining(session, GameCommand{Type: "init"})
	case "healing":
		session.State = StateVillageHealing
		return e.handleVillageHealing(session, GameCommand{Type: "init"})
	}

	// Show crafting menu
	msgs := []GameMessage{
		Msg("============================================================", "system"),
		Msg("CRAFTING MENU", "system"),
		Msg("============================================================", "system"),
		Msg("", "system"),
		Msg("Available Crafting:", "system"),
	}

	options := []MenuOption{}
	if game.Contains(village.UnlockedCrafting, "potions") {
		options = append(options, Opt("potions", "Potion Crafting"))
	}
	if game.Contains(village.UnlockedCrafting, "armor") {
		options = append(options, Opt("armor", "Armor Crafting"))
	}
	if game.Contains(village.UnlockedCrafting, "weapons") {
		options = append(options, Opt("weapons", "Weapon Crafting"))
	}
	if game.Contains(village.UnlockedCrafting, "skill_upgrades") {
		options = append(options, Opt("skill_upgrades", "Skill Upgrades"))
	}
	if game.Contains(village.UnlockedCrafting, "skill_scrolls") {
		options = append(options, Opt("skill_scrolls", "Skill Scroll Crafting"))
	}
	if game.Contains(village.UnlockedCrafting, "fortifications") {
		options = append(options, Opt("fortifications", "Fortifications (Village Defense)"))
	}
	if game.Contains(village.UnlockedCrafting, "training") {
		options = append(options, Opt("training", "Villager Training"))
	}
	if game.Contains(village.UnlockedCrafting, "healing") {
		options = append(options, Opt("healing", "Healing Services"))
	}
	options = append(options, Opt("0", "Back"))

	session.State = StateVillageCrafting
	return GameResponse{
		Type:     "menu",
		Messages: msgs,
		State:    &StateData{Screen: "village_crafting", Player: MakePlayerState(session.Player)},
		Options:  options,
	}
}

func (e *Engine) handleVillageCraftPotion(session *GameSession, cmd GameCommand) GameResponse {
	village := session.SelectedVillage
	player := session.Player

	if cmd.Value == "0" || cmd.Value == "back" {
		session.State = StateVillageCrafting
		return e.handleVillageCrafting(session, GameCommand{Type: "init"})
	}

	type potionRecipe struct {
		name     string
		size     string
		ironCost int
		goldCost int
	}
	recipes := []potionRecipe{
		{"Small Health Potion", "small", 5, 10},
		{"Medium Health Potion", "medium", 10, 20},
		{"Large Health Potion", "large", 20, 40},
	}

	if cmd.Type != "init" {
		idx, err := strconv.Atoi(cmd.Value)
		if err == nil && idx >= 1 && idx <= len(recipes) {
			recipe := recipes[idx-1]
			iron := player.ResourceStorageMap["Iron"]
			gold := player.ResourceStorageMap["Gold"]

			if iron.Stock < recipe.ironCost {
				// Show error and re-display
				session.State = StateVillageCraftPotion
				return GameResponse{
					Type:     "menu",
					Messages: []GameMessage{Msg(fmt.Sprintf("Not enough Iron! Need %d, have %d", recipe.ironCost, iron.Stock), "error")},
					State:    &StateData{Screen: "village_craft_potion", Player: MakePlayerState(player)},
					Options:  []MenuOption{Opt("back", "Back")},
				}
			}
			if gold.Stock < recipe.goldCost {
				session.State = StateVillageCraftPotion
				return GameResponse{
					Type:     "menu",
					Messages: []GameMessage{Msg(fmt.Sprintf("Not enough Gold! Need %d, have %d", recipe.goldCost, gold.Stock), "error")},
					State:    &StateData{Screen: "village_craft_potion", Player: MakePlayerState(player)},
					Options:  []MenuOption{Opt("back", "Back")},
				}
			}

			iron.Stock -= recipe.ironCost
			gold.Stock -= recipe.goldCost
			player.ResourceStorageMap["Iron"] = iron
			player.ResourceStorageMap["Gold"] = gold

			potion := game.CreateHealthPotion(recipe.size)
			player.Inventory = append(player.Inventory, potion)
			village.Experience += 20

			e.saveVillage(session)

			session.State = StateVillageCraftPotion
			msgs := []GameMessage{
				Msg(fmt.Sprintf("Crafted %s!", recipe.name), "loot"),
				Msg("+20 Village XP", "system"),
			}
			return GameResponse{
				Type:     "menu",
				Messages: msgs,
				State:    &StateData{Screen: "village_craft_potion", Player: MakePlayerState(player)},
				Options: []MenuOption{
					Opt("1", fmt.Sprintf("Small Health Potion (Iron: %d, Gold: %d)", recipes[0].ironCost, recipes[0].goldCost)),
					Opt("2", fmt.Sprintf("Medium Health Potion (Iron: %d, Gold: %d)", recipes[1].ironCost, recipes[1].goldCost)),
					Opt("3", fmt.Sprintf("Large Health Potion (Iron: %d, Gold: %d)", recipes[2].ironCost, recipes[2].goldCost)),
					Opt("0", "Back"),
				},
			}
		}
	}

	// Show potion crafting menu
	msgs := []GameMessage{
		Msg("============================================================", "system"),
		Msg("POTION CRAFTING", "system"),
		Msg("============================================================", "system"),
		Msg("", "system"),
		Msg("Available Recipes:", "system"),
	}

	options := []MenuOption{}
	for i, recipe := range recipes {
		options = append(options, Opt(strconv.Itoa(i+1), fmt.Sprintf("%s (Iron: %d, Gold: %d)", recipe.name, recipe.ironCost, recipe.goldCost)))
	}
	options = append(options, Opt("0", "Back"))

	session.State = StateVillageCraftPotion
	return GameResponse{
		Type:     "menu",
		Messages: msgs,
		State:    &StateData{Screen: "village_craft_potion", Player: MakePlayerState(player)},
		Options:  options,
	}
}

func (e *Engine) handleVillageCraftArmor(session *GameSession, cmd GameCommand) GameResponse {
	village := session.SelectedVillage
	player := session.Player

	if cmd.Value == "0" || cmd.Value == "back" {
		session.State = StateVillageCrafting
		return e.handleVillageCrafting(session, GameCommand{Type: "init"})
	}

	type armorRecipe struct {
		name      string
		materials map[string]int
		defBonus  func(int) int
		hpBonus   func(int) int
		atkBonus  func(int) int
		rarityMin int
		rarityMax int
		xp        int
		desc      string
	}

	recipes := []armorRecipe{
		{"Enhanced Armor", map[string]int{"Iron": 30, "Stone": 20}, func(r int) int { return r * 2 }, func(r int) int { return r }, func(r int) int { return 0 }, 3, 5, 40, "Random armor, Rarity 3-5, Defense focus"},
		{"Beast Skin Armor", map[string]int{"Iron": 20, "Beast Skin": 15}, func(r int) int { return r * 2 }, func(r int) int { return r + 3 }, func(r int) int { return 0 }, 4, 6, 50, "Light armor, Rarity 4-6, Fire resistance"},
		{"Bone Plate Armor", map[string]int{"Iron": 25, "Beast Bone": 12, "Stone": 15}, func(r int) int { return r * 3 }, func(r int) int { return r * 2 }, func(r int) int { return 0 }, 5, 7, 60, "Heavy armor, Rarity 5-7, High defense"},
		{"Tough Hide Vest", map[string]int{"Tough Hide": 10, "Beast Bone": 8}, func(r int) int { return r*2 + 2 }, func(r int) int { return r + 4 }, func(r int) int { return 0 }, 4, 6, 50, "Medium armor, Rarity 4-6, Physical resistance"},
		{"Ore Fragment Mail", map[string]int{"Ore Fragment": 20, "Iron": 15}, func(r int) int { return r*2 + 3 }, func(r int) int { return r }, func(r int) int { return 0 }, 5, 7, 60, "Magic armor, Rarity 5-7, Lightning resistance"},
		{"Fang-Studded Armor", map[string]int{"Sharp Fang": 15, "Beast Skin": 10, "Iron": 20}, func(r int) int { return r*2 + 5 }, func(r int) int { return r * 2 }, func(r int) int { return r }, 6, 8, 70, "Spiked armor, Rarity 6-8, Counter-damage bonus"},
		{"Claw Guard Armor", map[string]int{"Monster Claw": 12, "Tough Hide": 8, "Iron": 15}, func(r int) int { return r*3 + 2 }, func(r int) int { return r*2 + 3 }, func(r int) int { return 0 }, 6, 8, 70, "Elite armor, Rarity 6-8, Ice resistance"},
	}

	if cmd.Type != "init" {
		idx, err := strconv.Atoi(cmd.Value)
		if err == nil && idx >= 1 && idx <= len(recipes) {
			recipe := recipes[idx-1]

			// Check resources
			for mat, qty := range recipe.materials {
				res := player.ResourceStorageMap[mat]
				if res.Stock < qty {
					session.State = StateVillageCraftArmor
					return GameResponse{
						Type:     "menu",
						Messages: []GameMessage{Msg(fmt.Sprintf("Not enough %s! Need %d, have %d", mat, qty, res.Stock), "error")},
						State:    &StateData{Screen: "village_craft_armor", Player: MakePlayerState(player)},
						Options:  []MenuOption{Opt("back", "Back")},
					}
				}
			}

			// Deduct resources
			for mat, qty := range recipe.materials {
				res := player.ResourceStorageMap[mat]
				res.Stock -= qty
				player.ResourceStorageMap[mat] = res
			}

			rarity := recipe.rarityMin + rand.Intn(recipe.rarityMax-recipe.rarityMin+1)
			armor := game.GenerateItem(rarity)
			armor.StatsMod.DefenseMod += recipe.defBonus(rarity)
			armor.StatsMod.HitPointMod += recipe.hpBonus(rarity)
			armor.StatsMod.AttackMod += recipe.atkBonus(rarity)
			armor.CP = armor.StatsMod.AttackMod + armor.StatsMod.DefenseMod + armor.StatsMod.HitPointMod

			game.EquipBestItem(armor, &player.EquipmentMap, &player.Inventory)
			player.StatsMod = game.CalculateItemMods(player.EquipmentMap)
			player.HitpointsTotal = player.HitpointsNatural + player.StatsMod.HitPointMod

			village.Experience += recipe.xp

			e.saveVillage(session)

			session.State = StateVillageCraftArmor
			extraMsgs := []GameMessage{
				Msg(fmt.Sprintf("Crafted %s (Rarity %d)!", recipe.name, rarity), "loot"),
				Msg(fmt.Sprintf("   Defense: +%d | HP: +%d | CP: %d", armor.StatsMod.DefenseMod, armor.StatsMod.HitPointMod, armor.CP), "system"),
				Msg(fmt.Sprintf("+%d Village XP", recipe.xp), "system"),
			}
			return buildArmorCraftingResponse(session, extraMsgs)
		}
	}

	return buildArmorCraftingResponse(session, nil)
}

func buildArmorCraftingResponse(session *GameSession, extraMsgs []GameMessage) GameResponse {
	msgs := append([]GameMessage{}, extraMsgs...)
	msgs = append(msgs,
		Msg("============================================================", "system"),
		Msg("ARMOR CRAFTING", "system"),
		Msg("============================================================", "system"),
		Msg("", "system"),
		Msg("Available Recipes:", "system"),
		Msg("", "system"),
		Msg("--- STANDARD ARMOR ---", "system"),
		Msg("1. Enhanced Armor (Iron: 30, Stone: 20) -> Defense focus", "system"),
		Msg("", "system"),
		Msg("--- BEAST MATERIAL ARMOR ---", "system"),
		Msg("2. Beast Skin Armor (Iron: 20, Beast Skin: 15) -> Fire resistance", "system"),
		Msg("3. Bone Plate Armor (Iron: 25, Beast Bone: 12, Stone: 15) -> High defense", "system"),
		Msg("4. Tough Hide Vest (Tough Hide: 10, Beast Bone: 8) -> Physical resistance", "system"),
		Msg("5. Ore Fragment Mail (Ore Fragment: 20, Iron: 15) -> Lightning resistance", "system"),
		Msg("6. Fang-Studded Armor (Sharp Fang: 15, Beast Skin: 10, Iron: 20) -> Counter-damage", "system"),
		Msg("7. Claw Guard Armor (Monster Claw: 12, Tough Hide: 8, Iron: 15) -> Ice resistance", "system"),
	)

	options := []MenuOption{
		Opt("1", "Enhanced Armor"),
		Opt("2", "Beast Skin Armor"),
		Opt("3", "Bone Plate Armor"),
		Opt("4", "Tough Hide Vest"),
		Opt("5", "Ore Fragment Mail"),
		Opt("6", "Fang-Studded Armor"),
		Opt("7", "Claw Guard Armor"),
		Opt("0", "Back"),
	}

	return GameResponse{
		Type:     "menu",
		Messages: msgs,
		State:    &StateData{Screen: "village_craft_armor", Player: MakePlayerState(session.Player)},
		Options:  options,
	}
}

func buildWeaponCraftingResponse(session *GameSession, extraMsgs []GameMessage) GameResponse {
	msgs := append([]GameMessage{}, extraMsgs...)
	msgs = append(msgs,
		Msg("============================================================", "system"),
		Msg("WEAPON CRAFTING", "system"),
		Msg("============================================================", "system"),
		Msg("", "system"),
		Msg("Available Recipes:", "system"),
		Msg("", "system"),
		Msg("--- STANDARD WEAPONS ---", "system"),
		Msg("1. Enhanced Weapon (Iron: 40, Gold: 30) -> Attack focus", "system"),
		Msg("", "system"),
		Msg("--- BEAST MATERIAL WEAPONS ---", "system"),
		Msg("2. Beast Claw Blade (Iron: 25, Monster Claw: 15, Sharp Fang: 10) -> Bleed bonus", "system"),
		Msg("3. Bone Crusher Mace (Iron: 30, Beast Bone: 20, Stone: 15) -> Stun chance", "system"),
		Msg("4. Hide-Wrapped Axe (Iron: 20, Tough Hide: 12, Lumber: 25) -> Balanced", "system"),
		Msg("5. Ore Fragment Sword (Iron: 35, Ore Fragment: 25, Gold: 20) -> Elemental damage", "system"),
		Msg("6. Fang Spear (Sharp Fang: 18, Beast Bone: 15, Iron: 20) -> Critical bonus", "system"),
		Msg("7. Composite War Hammer (Beast Skin: 10, Ore Fragment: 15, Iron: 25, Stone: 20) -> Elite", "system"),
	)

	options := []MenuOption{
		Opt("1", "Enhanced Weapon"),
		Opt("2", "Beast Claw Blade"),
		Opt("3", "Bone Crusher Mace"),
		Opt("4", "Hide-Wrapped Axe"),
		Opt("5", "Ore Fragment Sword"),
		Opt("6", "Fang Spear"),
		Opt("7", "Composite War Hammer"),
		Opt("0", "Back"),
	}

	return GameResponse{
		Type:     "menu",
		Messages: msgs,
		State:    &StateData{Screen: "village_craft_weapon", Player: MakePlayerState(session.Player)},
		Options:  options,
	}
}

// checkAndDeductResources verifies the player has sufficient resources and deducts them.
// Returns true on success, or false with an error message on failure.
func checkAndDeductResources(player *models.Character, materials map[string]int) (bool, string) {
	for mat, qty := range materials {
		res := player.ResourceStorageMap[mat]
		if res.Stock < qty {
			return false, fmt.Sprintf("Not enough %s! Need %d, have %d", mat, qty, res.Stock)
		}
	}
	for mat, qty := range materials {
		res := player.ResourceStorageMap[mat]
		res.Stock -= qty
		player.ResourceStorageMap[mat] = res
	}
	return true, ""
}

func (e *Engine) handleVillageCraftWeapon(session *GameSession, cmd GameCommand) GameResponse {
	village := session.SelectedVillage
	player := session.Player

	if cmd.Value == "0" || cmd.Value == "back" {
		session.State = StateVillageCrafting
		return e.handleVillageCrafting(session, GameCommand{Type: "init"})
	}

	type weaponDef struct {
		materials map[string]int
		atkBonus  func(int) int
		defBonus  func(int) int
		hpBonus   func(int) int
		rarityMin int
		rarityMax int
		xp        int
		name      string
	}

	weapons := map[string]weaponDef{
		"1": {map[string]int{"Iron": 40, "Gold": 30}, func(r int) int { return r * 3 }, func(r int) int { return 0 }, func(r int) int { return r / 2 }, 4, 6, 50, "Enhanced Weapon"},
		"2": {map[string]int{"Iron": 25, "Monster Claw": 15, "Sharp Fang": 10}, func(r int) int { return r * 4 }, func(r int) int { return 0 }, func(r int) int { return r }, 5, 7, 60, "Beast Claw Blade"},
		"3": {map[string]int{"Iron": 30, "Beast Bone": 20, "Stone": 15}, func(r int) int { return r*3 + 5 }, func(r int) int { return r }, func(r int) int { return r + 2 }, 5, 7, 60, "Bone Crusher Mace"},
		"4": {map[string]int{"Iron": 20, "Tough Hide": 12, "Lumber": 25}, func(r int) int { return r*3 + 3 }, func(r int) int { return 0 }, func(r int) int { return r*2 + 5 }, 4, 6, 55, "Hide-Wrapped Axe"},
		"5": {map[string]int{"Iron": 35, "Ore Fragment": 25, "Gold": 20}, func(r int) int { return r*4 + 5 }, func(r int) int { return 0 }, func(r int) int { return r + 3 }, 6, 8, 70, "Ore Fragment Sword"},
		"6": {map[string]int{"Sharp Fang": 18, "Beast Bone": 15, "Iron": 20}, func(r int) int { return r*3 + 7 }, func(r int) int { return 0 }, func(r int) int { return r }, 5, 7, 65, "Fang Spear"},
		"7": {map[string]int{"Beast Skin": 10, "Ore Fragment": 15, "Iron": 25, "Stone": 20}, func(r int) int { return r*5 + 3 }, func(r int) int { return r + 2 }, func(r int) int { return r*2 + 5 }, 6, 8, 80, "Composite War Hammer"},
	}

	if cmd.Type != "init" {
		if wep, ok := weapons[cmd.Value]; ok {
			ok2, errMsg := checkAndDeductResources(player, wep.materials)
			if !ok2 {
				session.State = StateVillageCraftWeapon
				return GameResponse{
					Type:     "menu",
					Messages: []GameMessage{Msg(errMsg, "error")},
					State:    &StateData{Screen: "village_craft_weapon", Player: MakePlayerState(player)},
					Options:  []MenuOption{Opt("back", "Back")},
				}
			}

			rarity := wep.rarityMin + rand.Intn(wep.rarityMax-wep.rarityMin+1)
			weapon := game.GenerateItem(rarity)
			weapon.StatsMod.AttackMod += wep.atkBonus(rarity)
			weapon.StatsMod.DefenseMod += wep.defBonus(rarity)
			weapon.StatsMod.HitPointMod += wep.hpBonus(rarity)
			weapon.CP = weapon.StatsMod.AttackMod + weapon.StatsMod.DefenseMod + weapon.StatsMod.HitPointMod

			game.EquipBestItem(weapon, &player.EquipmentMap, &player.Inventory)
			player.StatsMod = game.CalculateItemMods(player.EquipmentMap)
			player.HitpointsTotal = player.HitpointsNatural + player.StatsMod.HitPointMod

			village.Experience += wep.xp
			e.saveVillage(session)

			extraMsgs := []GameMessage{
				Msg(fmt.Sprintf("Crafted %s (Rarity %d)!", wep.name, rarity), "loot"),
				Msg(fmt.Sprintf("   Attack: +%d | Defense: +%d | HP: +%d | CP: %d",
					weapon.StatsMod.AttackMod, weapon.StatsMod.DefenseMod, weapon.StatsMod.HitPointMod, weapon.CP), "system"),
				Msg(fmt.Sprintf("+%d Village XP", wep.xp), "system"),
			}
			session.State = StateVillageCraftWeapon
			resp := buildWeaponCraftingResponse(session, extraMsgs)
			return resp
		}
	}

	session.State = StateVillageCraftWeapon
	return buildWeaponCraftingResponse(session, nil)
}

func (e *Engine) handleVillageUpgradeSkill(session *GameSession, cmd GameCommand) GameResponse {
	player := session.Player

	if cmd.Value == "0" || cmd.Value == "back" {
		session.State = StateVillageCrafting
		return e.handleVillageCrafting(session, GameCommand{Type: "init"})
	}

	if len(player.LearnedSkills) == 0 {
		session.State = StateVillageCrafting
		resp := e.handleVillageCrafting(session, GameCommand{Type: "init"})
		resp.Messages = append([]GameMessage{Msg("No skills to upgrade!", "error")}, resp.Messages...)
		return resp
	}

	if cmd.Type != "init" {
		idx, err := strconv.Atoi(cmd.Value)
		if err == nil && idx >= 1 && idx <= len(player.LearnedSkills) {
			session.SelectedSkillIdx = idx - 1
			session.State = StateVillageUpgradeConfirm
			return e.handleVillageUpgradeConfirm(session, GameCommand{Type: "init"})
		}
	}

	msgs := []GameMessage{
		Msg("============================================================", "system"),
		Msg("SKILL UPGRADES", "system"),
		Msg("============================================================", "system"),
		Msg("", "system"),
		Msg("Your Skills:", "system"),
	}

	options := []MenuOption{}
	for i, skill := range player.LearnedSkills {
		label := fmt.Sprintf("%s (Damage: %d, ManaCost: %d, StaminaCost: %d)",
			skill.Name, skill.Damage, skill.ManaCost, skill.StaminaCost)
		options = append(options, Opt(strconv.Itoa(i+1), label))
	}
	options = append(options, Opt("0", "Back"))

	session.State = StateVillageUpgradeSkill
	return GameResponse{
		Type:     "menu",
		Messages: msgs,
		State:    &StateData{Screen: "village_upgrade_skill", Player: MakePlayerState(player)},
		Options:  options,
	}
}

func (e *Engine) handleVillageUpgradeConfirm(session *GameSession, cmd GameCommand) GameResponse {
	village := session.SelectedVillage
	player := session.Player
	skillIdx := session.SelectedSkillIdx

	if skillIdx < 0 || skillIdx >= len(player.LearnedSkills) {
		session.State = StateVillageUpgradeSkill
		return e.handleVillageUpgradeSkill(session, GameCommand{Type: "init"})
	}

	skill := &player.LearnedSkills[skillIdx]

	if cmd.Value == "n" || cmd.Value == "0" || cmd.Value == "back" {
		session.State = StateVillageUpgradeSkill
		return e.handleVillageUpgradeSkill(session, GameCommand{Type: "init"})
	}

	if cmd.Type == "init" {
		msgs := []GameMessage{
			Msg(fmt.Sprintf("Upgrade %s", skill.Name), "system"),
			Msg("Cost: Gold: 50, Iron: 25", "system"),
			Msg("Effect: +5 Damage (or +5 Healing), -2 Resource Cost", "system"),
		}

		session.State = StateVillageUpgradeConfirm
		return GameResponse{
			Type:     "menu",
			Messages: msgs,
			State:    &StateData{Screen: "village_upgrade_confirm", Player: MakePlayerState(player)},
			Options: []MenuOption{
				Opt("y", "Yes, upgrade"),
				Opt("n", "No, cancel"),
			},
		}
	}

	if cmd.Value == "y" {
		iron := player.ResourceStorageMap["Iron"]
		gold := player.ResourceStorageMap["Gold"]

		if iron.Stock < 25 {
			session.State = StateVillageUpgradeSkill
			resp := e.handleVillageUpgradeSkill(session, GameCommand{Type: "init"})
			resp.Messages = append([]GameMessage{Msg(fmt.Sprintf("Not enough Iron! Need 25, have %d", iron.Stock), "error")}, resp.Messages...)
			return resp
		}
		if gold.Stock < 50 {
			session.State = StateVillageUpgradeSkill
			resp := e.handleVillageUpgradeSkill(session, GameCommand{Type: "init"})
			resp.Messages = append([]GameMessage{Msg(fmt.Sprintf("Not enough Gold! Need 50, have %d", gold.Stock), "error")}, resp.Messages...)
			return resp
		}

		iron.Stock -= 25
		gold.Stock -= 50
		player.ResourceStorageMap["Iron"] = iron
		player.ResourceStorageMap["Gold"] = gold

		resultMsgs := []GameMessage{}
		if skill.Damage > 0 {
			skill.Damage += 5
			resultMsgs = append(resultMsgs, Msg(fmt.Sprintf("%s damage increased by 5! (Now: %d)", skill.Name, skill.Damage), "system"))
		} else if skill.Damage < 0 {
			skill.Damage -= 5
			resultMsgs = append(resultMsgs, Msg(fmt.Sprintf("%s healing increased by 5! (Now: %d)", skill.Name, -skill.Damage), "system"))
		}

		if skill.ManaCost > 2 {
			skill.ManaCost -= 2
			resultMsgs = append(resultMsgs, Msg(fmt.Sprintf("Mana cost reduced by 2! (Now: %d)", skill.ManaCost), "system"))
		}
		if skill.StaminaCost > 2 {
			skill.StaminaCost -= 2
			resultMsgs = append(resultMsgs, Msg(fmt.Sprintf("Stamina cost reduced by 2! (Now: %d)", skill.StaminaCost), "system"))
		}

		village.Experience += 60
		resultMsgs = append(resultMsgs, Msg("+60 Village XP", "system"))

		e.saveVillage(session)

		session.State = StateVillageUpgradeSkill
		resp := e.handleVillageUpgradeSkill(session, GameCommand{Type: "init"})
		resp.Messages = append(resultMsgs, resp.Messages...)
		return resp
	}

	session.State = StateVillageUpgradeSkill
	return e.handleVillageUpgradeSkill(session, GameCommand{Type: "init"})
}

func (e *Engine) handleVillageCraftScrolls(session *GameSession, cmd GameCommand) GameResponse {
	village := session.SelectedVillage
	player := session.Player

	if cmd.Value == "0" || cmd.Value == "back" {
		session.State = StateVillageCrafting
		return e.handleVillageCrafting(session, GameCommand{Type: "init"})
	}

	type scrollRecipe struct {
		skillIdx  int
		materials map[string]int
		xp        int
	}

	recipes := map[string]scrollRecipe{
		"1":  {0, map[string]int{"Ore Fragment": 15, "Sharp Fang": 10, "Gold": 30}, 100},
		"2":  {1, map[string]int{"Beast Skin": 12, "Ore Fragment": 10, "Iron": 20}, 90},
		"3":  {2, map[string]int{"Ore Fragment": 20, "Monster Claw": 15, "Gold": 40}, 120},
		"4":  {4, map[string]int{"Beast Bone": 10, "Iron": 15}, 70},
		"5":  {7, map[string]int{"Beast Skin": 10, "Sharp Fang": 12, "Iron": 15}, 85},
		"6":  {3, map[string]int{"Beast Skin": 15, "Beast Bone": 10, "Gold": 25}, 95},
		"7":  {8, map[string]int{"Ore Fragment": 12, "Beast Skin": 15, "Gold": 30}, 105},
		"8":  {5, map[string]int{"Tough Hide": 15, "Beast Bone": 12, "Stone": 20}, 90},
		"9":  {6, map[string]int{"Sharp Fang": 15, "Beast Bone": 10, "Iron": 20}, 85},
		"10": {9, map[string]int{"Beast Bone": 8, "Beast Skin": 8}, 60},
	}

	if cmd.Type != "init" {
		if recipe, ok := recipes[cmd.Value]; ok {
			skillToLearn := data.AvailableSkills[recipe.skillIdx]

			// Check if already known
			for _, skill := range player.LearnedSkills {
				if skill.Name == skillToLearn.Name {
					session.State = StateVillageCraftScrolls
					resp := buildScrollCraftingResponse(session, nil)
					resp.Messages = append([]GameMessage{Msg(fmt.Sprintf("You already know %s!", skillToLearn.Name), "error")}, resp.Messages...)
					return resp
				}
			}

			// Check resources
			ok2, errMsg := checkAndDeductResources(player, recipe.materials)
			if !ok2 {
				session.State = StateVillageCraftScrolls
				resp := buildScrollCraftingResponse(session, nil)
				resp.Messages = append([]GameMessage{Msg(errMsg, "error")}, resp.Messages...)
				return resp
			}

			player.LearnedSkills = append(player.LearnedSkills, skillToLearn)
			village.Experience += recipe.xp

			e.saveVillage(session)

			extraMsgs := []GameMessage{
				Msg(fmt.Sprintf("Crafted %s Scroll!", skillToLearn.Name), "loot"),
				Msg(fmt.Sprintf("You have learned %s!", skillToLearn.Name), "system"),
				Msg(fmt.Sprintf("   %s", skillToLearn.Description), "system"),
				Msg(fmt.Sprintf("+%d Village XP", recipe.xp), "system"),
			}
			session.State = StateVillageCraftScrolls
			return buildScrollCraftingResponse(session, extraMsgs)
		}
	}

	session.State = StateVillageCraftScrolls
	return buildScrollCraftingResponse(session, nil)
}

func buildScrollCraftingResponse(session *GameSession, extraMsgs []GameMessage) GameResponse {
	msgs := append([]GameMessage{}, extraMsgs...)
	msgs = append(msgs,
		Msg("============================================================", "system"),
		Msg("SKILL SCROLL CRAFTING", "system"),
		Msg("============================================================", "system"),
		Msg("", "system"),
		Msg("Craft skill scrolls using beast materials!", "system"),
		Msg("Learn skills without defeating Skill Guardians", "system"),
		Msg("", "system"),
		Msg("--- OFFENSIVE SKILLS ---", "system"),
		Msg("1. Fireball Scroll (Ore Fragment: 15, Sharp Fang: 10, Gold: 30)", "system"),
		Msg("2. Ice Shard Scroll (Beast Skin: 12, Ore Fragment: 10, Iron: 20)", "system"),
		Msg("3. Lightning Bolt Scroll (Ore Fragment: 20, Monster Claw: 15, Gold: 40)", "system"),
		Msg("4. Power Strike Scroll (Beast Bone: 10, Iron: 15)", "system"),
		Msg("5. Poison Blade Scroll (Beast Skin: 10, Sharp Fang: 12, Iron: 15)", "system"),
		Msg("", "system"),
		Msg("--- SUPPORT SKILLS ---", "system"),
		Msg("6. Heal Scroll (Beast Skin: 15, Beast Bone: 10, Gold: 25)", "system"),
		Msg("7. Regeneration Scroll (Ore Fragment: 12, Beast Skin: 15, Gold: 30)", "system"),
		Msg("8. Shield Wall Scroll (Tough Hide: 15, Beast Bone: 12, Stone: 20)", "system"),
		Msg("9. Battle Cry Scroll (Sharp Fang: 15, Beast Bone: 10, Iron: 20)", "system"),
		Msg("", "system"),
		Msg("--- UTILITY SKILLS ---", "system"),
		Msg("10. Tracking Scroll (Beast Bone: 8, Beast Skin: 8)", "system"),
	)

	options := []MenuOption{
		Opt("1", "Fireball Scroll"),
		Opt("2", "Ice Shard Scroll"),
		Opt("3", "Lightning Bolt Scroll"),
		Opt("4", "Power Strike Scroll"),
		Opt("5", "Poison Blade Scroll"),
		Opt("6", "Heal Scroll"),
		Opt("7", "Regeneration Scroll"),
		Opt("8", "Shield Wall Scroll"),
		Opt("9", "Battle Cry Scroll"),
		Opt("10", "Tracking Scroll"),
		Opt("0", "Back"),
	}

	return GameResponse{
		Type:     "menu",
		Messages: msgs,
		State:    &StateData{Screen: "village_craft_scrolls", Player: MakePlayerState(session.Player)},
		Options:  options,
	}
}

func (e *Engine) handleVillageBuildDefense(session *GameSession, cmd GameCommand) GameResponse {
	if cmd.Value == "0" || cmd.Value == "back" {
		session.State = StateVillageMain
		return e.handleVillageMain(session, GameCommand{Type: "init"})
	}

	switch cmd.Value {
	case "1":
		session.State = StateVillageBuildWalls
		return e.handleVillageBuildWalls(session, GameCommand{Type: "init"})
	case "2":
		session.State = StateVillageCraftTraps
		return e.handleVillageCraftTraps(session, GameCommand{Type: "init"})
	case "3":
		session.State = StateVillageViewDefenses
		return e.handleVillageViewDefenses(session, GameCommand{Type: "init"})
	}

	msgs := []GameMessage{
		Msg("============================================================", "system"),
		Msg("BUILD DEFENSES & TRAPS", "system"),
		Msg("============================================================", "system"),
	}

	options := []MenuOption{
		Opt("1", "Build Walls/Towers"),
		Opt("2", "Craft Traps"),
		Opt("3", "View Current Defenses"),
		Opt("0", "Back"),
	}

	session.State = StateVillageBuildDefense
	return GameResponse{
		Type:     "menu",
		Messages: msgs,
		State:    &StateData{Screen: "village_build_defense", Player: MakePlayerState(session.Player)},
		Options:  options,
	}
}

func (e *Engine) handleVillageBuildWalls(session *GameSession, cmd GameCommand) GameResponse {
	village := session.SelectedVillage
	player := session.Player

	if cmd.Value == "0" || cmd.Value == "back" {
		session.State = StateVillageBuildDefense
		return e.handleVillageBuildDefense(session, GameCommand{Type: "init"})
	}

	type wallDef struct {
		name    string
		lumber  int
		stone   int
		iron    int
		defense int
		attack  int
		dtype   string
	}

	defenseOptions := []wallDef{
		{"Wooden Wall", 50, 20, 0, 10, 0, "wall"},
		{"Stone Wall", 30, 60, 10, 25, 0, "wall"},
		{"Iron Wall", 20, 80, 40, 40, 0, "wall"},
		{"Guard Tower", 40, 40, 30, 15, 20, "tower"},
		{"Arrow Tower", 30, 50, 40, 10, 35, "tower"},
		{"Iron Gate", 20, 50, 50, 30, 10, "wall"},
	}

	if cmd.Type != "init" {
		idx, err := strconv.Atoi(cmd.Value)
		if err == nil && idx >= 1 && idx <= len(defenseOptions) {
			selected := defenseOptions[idx-1]
			materials := map[string]int{}
			if selected.lumber > 0 {
				materials["Lumber"] = selected.lumber
			}
			if selected.stone > 0 {
				materials["Stone"] = selected.stone
			}
			if selected.iron > 0 {
				materials["Iron"] = selected.iron
			}

			ok, errMsg := checkAndDeductResources(player, materials)
			if !ok {
				session.State = StateVillageBuildWalls
				return GameResponse{
					Type:     "menu",
					Messages: []GameMessage{Msg(errMsg, "error")},
					State:    &StateData{Screen: "village_build_walls", Player: MakePlayerState(player)},
					Options:  []MenuOption{Opt("back", "Back")},
				}
			}

			newDefense := models.Defense{
				Name:        selected.name,
				Level:       1,
				Defense:     selected.defense,
				AttackPower: selected.attack,
				Range:       10,
				Built:       true,
				Type:        selected.dtype,
			}
			village.Defenses = append(village.Defenses, newDefense)
			village.DefenseLevel++
			village.Experience += 30

			e.saveVillage(session)

			extraMsgs := []GameMessage{
				Msg(fmt.Sprintf("Built %s!", selected.name), "system"),
				Msg(fmt.Sprintf("Village Defense Level increased to %d", village.DefenseLevel), "system"),
				Msg("+30 Village XP", "system"),
			}

			session.State = StateVillageBuildWalls
			resp := buildWallsResponse(session, nil)
			resp.Messages = append(extraMsgs, resp.Messages...)
			return resp
		}
	}

	session.State = StateVillageBuildWalls
	return buildWallsResponse(session, nil)
}

func buildWallsResponse(session *GameSession, extraMsgs []GameMessage) GameResponse {
	msgs := append([]GameMessage{}, extraMsgs...)
	msgs = append(msgs,
		Msg("============================================================", "system"),
		Msg("BUILD WALLS & TOWERS", "system"),
		Msg("============================================================", "system"),
		Msg("", "system"),
		Msg("Available Structures:", "system"),
	)

	type wallInfo struct {
		name    string
		lumber  int
		stone   int
		iron    int
		defense int
		attack  int
	}

	walls := []wallInfo{
		{"Wooden Wall", 50, 20, 0, 10, 0},
		{"Stone Wall", 30, 60, 10, 25, 0},
		{"Iron Wall", 20, 80, 40, 40, 0},
		{"Guard Tower", 40, 40, 30, 15, 20},
		{"Arrow Tower", 30, 50, 40, 10, 35},
		{"Iron Gate", 20, 50, 50, 30, 10},
	}

	options := []MenuOption{}
	for i, w := range walls {
		label := fmt.Sprintf("%s (Lumber:%d Stone:%d Iron:%d) -> Defense:+%d Attack:+%d",
			w.name, w.lumber, w.stone, w.iron, w.defense, w.attack)
		options = append(options, Opt(strconv.Itoa(i+1), label))
	}
	options = append(options, Opt("0", "Back"))

	return GameResponse{
		Type:     "menu",
		Messages: msgs,
		State:    &StateData{Screen: "village_build_walls", Player: MakePlayerState(session.Player)},
		Options:  options,
	}
}

func (e *Engine) handleVillageCraftTraps(session *GameSession, cmd GameCommand) GameResponse {
	village := session.SelectedVillage
	player := session.Player

	if cmd.Value == "0" || cmd.Value == "back" {
		session.State = StateVillageBuildDefense
		return e.handleVillageBuildDefense(session, GameCommand{Type: "init"})
	}

	type trapDef struct {
		name        string
		trapType    string
		materials   map[string]int
		damage      int
		duration    int
		triggerRate int
	}

	trapOptions := []trapDef{
		{"Spike Trap", "spike", map[string]int{"Iron": 10, "Beast Bone": 5}, 15, 3, 60},
		{"Fire Trap", "fire", map[string]int{"Iron": 15, "Ore Fragment": 8, "Sharp Fang": 5}, 25, 2, 50},
		{"Ice Trap", "ice", map[string]int{"Iron": 12, "Ore Fragment": 10, "Beast Skin": 8}, 20, 3, 55},
		{"Poison Trap", "poison", map[string]int{"Beast Skin": 10, "Sharp Fang": 8, "Monster Claw": 5}, 18, 4, 65},
		{"Barricade Trap", "spike", map[string]int{"Lumber": 30, "Tough Hide": 6, "Beast Bone": 8}, 30, 2, 70},
	}

	if cmd.Type != "init" {
		idx, err := strconv.Atoi(cmd.Value)
		if err == nil && idx >= 1 && idx <= len(trapOptions) {
			selected := trapOptions[idx-1]

			ok, errMsg := checkAndDeductResources(player, selected.materials)
			if !ok {
				session.State = StateVillageCraftTraps
				return GameResponse{
					Type:     "menu",
					Messages: []GameMessage{Msg(errMsg, "error")},
					State:    &StateData{Screen: "village_craft_traps", Player: MakePlayerState(player)},
					Options:  []MenuOption{Opt("back", "Back")},
				}
			}

			newTrap := models.Trap{
				Name:        selected.name,
				Type:        selected.trapType,
				Damage:      selected.damage,
				Duration:    selected.duration,
				Remaining:   selected.duration,
				TriggerRate: selected.triggerRate,
			}
			village.Traps = append(village.Traps, newTrap)
			village.Experience += 35

			e.saveVillage(session)

			extraMsgs := []GameMessage{
				Msg(fmt.Sprintf("Crafted %s!", selected.name), "loot"),
				Msg(fmt.Sprintf("Will last for %d monster tides", selected.duration), "system"),
				Msg("+35 Village XP", "system"),
			}
			session.State = StateVillageCraftTraps
			resp := buildTrapCraftingResponse(session, nil)
			resp.Messages = append(extraMsgs, resp.Messages...)
			return resp
		}
	}

	session.State = StateVillageCraftTraps
	return buildTrapCraftingResponse(session, nil)
}

func buildTrapCraftingResponse(session *GameSession, extraMsgs []GameMessage) GameResponse {
	msgs := append([]GameMessage{}, extraMsgs...)
	msgs = append(msgs,
		Msg("============================================================", "system"),
		Msg("CRAFT TRAPS", "system"),
		Msg("============================================================", "system"),
		Msg("", "system"),
		Msg("Available Traps:", "system"),
		Msg("", "system"),
		Msg("1. Spike Trap (Iron:10, Beast Bone:5) Damage:15, Duration:3, Trigger:60%", "system"),
		Msg("2. Fire Trap (Iron:15, Ore Fragment:8, Sharp Fang:5) Damage:25, Duration:2, Trigger:50%", "system"),
		Msg("3. Ice Trap (Iron:12, Ore Fragment:10, Beast Skin:8) Damage:20, Duration:3, Trigger:55%", "system"),
		Msg("4. Poison Trap (Beast Skin:10, Sharp Fang:8, Monster Claw:5) Damage:18, Duration:4, Trigger:65%", "system"),
		Msg("5. Barricade Trap (Lumber:30, Tough Hide:6, Beast Bone:8) Damage:30, Duration:2, Trigger:70%", "system"),
	)

	options := []MenuOption{
		Opt("1", "Spike Trap"),
		Opt("2", "Fire Trap"),
		Opt("3", "Ice Trap"),
		Opt("4", "Poison Trap"),
		Opt("5", "Barricade Trap"),
		Opt("0", "Back"),
	}

	return GameResponse{
		Type:     "menu",
		Messages: msgs,
		State:    &StateData{Screen: "village_craft_traps", Player: MakePlayerState(session.Player)},
		Options:  options,
	}
}

func (e *Engine) handleVillageViewDefenses(session *GameSession, cmd GameCommand) GameResponse {
	village := session.SelectedVillage

	if cmd.Value == "back" || cmd.Value == "0" {
		session.State = StateVillageBuildDefense
		return e.handleVillageBuildDefense(session, GameCommand{Type: "init"})
	}

	msgs := []GameMessage{
		Msg("============================================================", "system"),
		Msg("CURRENT DEFENSES", "system"),
		Msg("============================================================", "system"),
	}

	walls := []models.Defense{}
	towers := []models.Defense{}

	for _, def := range village.Defenses {
		if def.Type == "wall" {
			walls = append(walls, def)
		} else if def.Type == "tower" {
			towers = append(towers, def)
		}
	}

	if len(walls) > 0 {
		msgs = append(msgs, Msg("", "system"))
		msgs = append(msgs, Msg("WALLS:", "system"))
		for _, wall := range walls {
			msgs = append(msgs, Msg(fmt.Sprintf("  - %s (Defense: +%d)", wall.Name, wall.Defense), "system"))
		}
	}

	if len(towers) > 0 {
		msgs = append(msgs, Msg("", "system"))
		msgs = append(msgs, Msg("TOWERS:", "system"))
		for _, tower := range towers {
			msgs = append(msgs, Msg(fmt.Sprintf("  - %s (Defense: +%d, Attack: +%d)", tower.Name, tower.Defense, tower.AttackPower), "system"))
		}
	}

	if len(village.Traps) > 0 {
		msgs = append(msgs, Msg("", "system"))
		msgs = append(msgs, Msg("ACTIVE TRAPS:", "system"))
		for i, trap := range village.Traps {
			msgs = append(msgs, Msg(fmt.Sprintf("  %d. %s (Damage: %d, Waves left: %d/%d, Trigger: %d%%)",
				i+1, trap.Name, trap.Damage, trap.Remaining, trap.Duration, trap.TriggerRate), "system"))
		}
	}

	if len(village.Defenses) == 0 && len(village.Traps) == 0 {
		msgs = append(msgs, Msg("", "system"))
		msgs = append(msgs, Msg("No defenses built yet!", "system"))
	}

	msgs = append(msgs, Msg("", "system"))
	msgs = append(msgs, Msg(fmt.Sprintf("Total Defense Level: %d", village.DefenseLevel), "system"))
	msgs = append(msgs, Msg("============================================================", "system"))

	session.State = StateVillageViewDefenses
	return GameResponse{
		Type:     "menu",
		Messages: msgs,
		State:    &StateData{Screen: "village_view_defenses", Player: MakePlayerState(session.Player)},
		Options:  []MenuOption{Opt("back", "Back")},
	}
}

func (e *Engine) handleVillageCheckTide(session *GameSession, cmd GameCommand) GameResponse {
	village := session.SelectedVillage

	if cmd.Value == "back" || cmd.Value == "0" {
		session.State = StateVillageMain
		return e.handleVillageMain(session, GameCommand{Type: "init"})
	}

	msgs := []GameMessage{
		Msg("============================================================", "system"),
		Msg("MONSTER TIDE STATUS", "system"),
		Msg("============================================================", "system"),
	}

	currentTime := time.Now().Unix()
	timeSinceLastTide := currentTime - village.LastTideTime
	timeUntilNext := village.TideInterval - int(timeSinceLastTide)

	if timeUntilNext <= 0 {
		msgs = append(msgs, Msg("", "system"))
		msgs = append(msgs, Msg("MONSTER TIDE IS READY!", "combat"))
		msgs = append(msgs, Msg("", "system"))
		msgs = append(msgs, Msg("A wave of monsters can attack your village!", "system"))
	} else {
		hours := timeUntilNext / 3600
		minutes := (timeUntilNext % 3600) / 60
		msgs = append(msgs, Msg("", "system"))
		msgs = append(msgs, Msg(fmt.Sprintf("Next Monster Tide in: %d hours, %d minutes", hours, minutes), "system"))
	}

	msgs = append(msgs,
		Msg(fmt.Sprintf("Village Defense Level: %d", village.DefenseLevel), "system"),
		Msg(fmt.Sprintf("Defenses: %d built", len(village.Defenses)), "system"),
		Msg(fmt.Sprintf("Active Traps: %d", len(village.Traps)), "system"),
		Msg(fmt.Sprintf("Guards: %d villagers + %d hired",
			game.CountVillagersByRole(village, "guard"),
			len(village.ActiveGuards)), "system"),
		Msg("", "system"),
		Msg("Prepare your defenses by:", "system"),
		Msg("  - Building more defenses", "system"),
		Msg("  - Crafting traps", "system"),
		Msg("  - Hiring guards", "system"),
		Msg("  - Rescuing guard villagers during hunts", "system"),
		Msg("============================================================", "system"),
	)

	session.State = StateVillageCheckTide
	return GameResponse{
		Type:     "menu",
		Messages: msgs,
		State:    &StateData{Screen: "village_check_tide", Player: MakePlayerState(session.Player)},
		Options:  []MenuOption{Opt("back", "Back to Village")},
	}
}

func (e *Engine) handleVillageMonsterTide(session *GameSession, cmd GameCommand) GameResponse {
	village := session.SelectedVillage
	player := session.Player

	if cmd.Value == "back" || cmd.Value == "0" {
		session.State = StateVillageMain
		return e.handleVillageMain(session, GameCommand{Type: "init"})
	}

	// Calculate tide difficulty
	numWaves := 3 + (village.Level / 5)
	baseMonsterLevel := village.Level
	monstersPerWave := 5 + (village.Level / 3)

	// Calculate village defense stats
	totalDefense := 0
	totalAttack := 0
	for _, def := range village.Defenses {
		totalDefense += def.Defense
		totalAttack += def.AttackPower
	}

	villagerGuards := game.CountVillagersByRole(village, "guard")
	hiredGuards := len(village.ActiveGuards)
	totalGuards := villagerGuards + hiredGuards

	msgs := []GameMessage{
		Msg("============================================================", "system"),
		Msg("MONSTER TIDE DEFENSE", "system"),
		Msg("============================================================", "system"),
		Msg("", "system"),
		Msg("A Monster Tide is approaching!", "combat"),
		Msg(fmt.Sprintf("Waves: %d", numWaves), "system"),
		Msg(fmt.Sprintf("Monsters per wave: ~%d", monstersPerWave), "system"),
		Msg(fmt.Sprintf("Monster Level: ~%d", baseMonsterLevel), "system"),
		Msg("", "system"),
		Msg("YOUR DEFENSES:", "system"),
		Msg(fmt.Sprintf("  Defense Power: %d (from %d structures)", totalDefense, len(village.Defenses)), "system"),
		Msg(fmt.Sprintf("  Attack Power: %d (from towers)", totalAttack), "system"),
		Msg(fmt.Sprintf("  Active Traps: %d", len(village.Traps)), "system"),
		Msg(fmt.Sprintf("  Guards: %d total (%d villagers + %d hired)", totalGuards, villagerGuards, hiredGuards), "system"),
	}

	// Store tide parameters in combat context for wave processing
	session.Combat = &CombatContext{
		Turn:           0, // Current wave (0-indexed, will increment)
		WavesTotal: numWaves,
	}

	session.State = StateVillageTideWave
	return GameResponse{
		Type:     "menu",
		Messages: msgs,
		State:    &StateData{Screen: "village_monster_tide", Player: MakePlayerState(player)},
		Options:  []MenuOption{Opt("start", "Begin Defense!")},
	}
}

func (e *Engine) handleVillageTideWave(session *GameSession, cmd GameCommand) GameResponse {
	village := session.SelectedVillage
	player := session.Player

	if session.Combat == nil {
		session.State = StateVillageMain
		return e.handleVillageMain(session, GameCommand{Type: "init"})
	}

	numWaves := session.Combat.WavesTotal
	currentWave := session.Combat.Turn
	baseMonsterLevel := village.Level
	monstersPerWave := 5 + (village.Level / 3)

	// Calculate defenses
	totalDefense := 0
	totalAttack := 0
	for _, def := range village.Defenses {
		totalDefense += def.Defense
		totalAttack += def.AttackPower
	}

	villagerGuards := game.CountVillagersByRole(village, "guard")
	hiredGuards := len(village.ActiveGuards)
	totalGuards := villagerGuards + hiredGuards

	// Process the current wave
	currentWave++
	session.Combat.Turn = currentWave

	msgs := []GameMessage{
		Msg(fmt.Sprintf("========== WAVE %d/%d ==========", currentWave, numWaves), "combat"),
	}

	waveSize := monstersPerWave + rand.Intn(3) - 1
	if waveSize < 1 {
		waveSize = 1
	}

	msgs = append(msgs, Msg(fmt.Sprintf("%d monsters approach!", waveSize), "combat"))

	waveDamageDealt := 0
	waveDamageTaken := 0
	monstersKilled := 0
	trapsTriggered := 0

	for i := 0; i < waveSize; i++ {
		monsterLevel := baseMonsterLevel + rand.Intn(5) - 2
		if monsterLevel < 1 {
			monsterLevel = 1
		}
		rank := 1 + rand.Intn(3)
		monster := game.GenerateMonster(data.MonsterNames[rand.Intn(len(data.MonsterNames))], monsterLevel, rank)

		msgs = append(msgs, Msg(fmt.Sprintf("  %s (Lv%d, HP:%d) attacks!", monster.Name, monster.Level, monster.HitpointsRemaining), "combat"))

		// Phase 1: Traps
		for j := range village.Traps {
			trap := &village.Traps[j]
			if trap.Remaining > 0 && rand.Intn(100) < trap.TriggerRate {
				monster.HitpointsRemaining -= trap.Damage
				waveDamageDealt += trap.Damage
				trapsTriggered++
				msgs = append(msgs, Msg(fmt.Sprintf("    %s triggers! (%d damage)", trap.Name, trap.Damage), "damage"))

				if monster.HitpointsRemaining <= 0 {
					msgs = append(msgs, Msg(fmt.Sprintf("    %s killed by trap!", monster.Name), "combat"))
					monstersKilled++
					break
				}
			}
		}

		if monster.HitpointsRemaining <= 0 {
			continue
		}

		// Phase 2: Towers
		if totalAttack > 0 {
			towerDamage := totalAttack + rand.Intn(5)
			monster.HitpointsRemaining -= towerDamage
			waveDamageDealt += towerDamage
			msgs = append(msgs, Msg(fmt.Sprintf("    Towers fire! (%d damage)", towerDamage), "damage"))

			if monster.HitpointsRemaining <= 0 {
				msgs = append(msgs, Msg(fmt.Sprintf("    %s killed by towers!", monster.Name), "combat"))
				monstersKilled++
				continue
			}
		}

		// Phase 3: Guards
		if totalGuards > 0 {
			guardDamage := totalGuards * (5 + rand.Intn(8))
			monster.HitpointsRemaining -= guardDamage
			waveDamageDealt += guardDamage
			msgs = append(msgs, Msg(fmt.Sprintf("    Guards attack! (%d damage)", guardDamage), "damage"))

			if monster.HitpointsRemaining <= 0 {
				msgs = append(msgs, Msg(fmt.Sprintf("    %s killed by guards!", monster.Name), "combat"))
				monstersKilled++
				continue
			}
		}

		// Phase 4: Monster attacks village
		monsterAttack := monster.AttackRolls * 6
		reducedDamage := monsterAttack - totalDefense
		if reducedDamage < 1 {
			reducedDamage = 1
		}
		waveDamageTaken += reducedDamage
		msgs = append(msgs, Msg(fmt.Sprintf("    %s breaches defenses! (%d damage to village)", monster.Name, reducedDamage), "damage"))
	}

	// Update running totals
	session.Combat.AutoPlayWins += monstersKilled
	session.Combat.AutoPlayXP += waveDamageDealt
	session.Combat.AutoPlayDeaths += waveDamageTaken
	session.Combat.AutoPlayFights += trapsTriggered

	msgs = append(msgs,
		Msg("", "system"),
		Msg(fmt.Sprintf("Wave %d complete!", currentWave), "combat"),
		Msg(fmt.Sprintf("  Monsters killed: %d/%d", monstersKilled, waveSize), "system"),
		Msg(fmt.Sprintf("  Damage dealt: %d", waveDamageDealt), "system"),
		Msg(fmt.Sprintf("  Damage taken: %d", waveDamageTaken), "system"),
	)

	// Decrement trap durability
	for j := len(village.Traps) - 1; j >= 0; j-- {
		village.Traps[j].Remaining--
		if village.Traps[j].Remaining <= 0 {
			msgs = append(msgs, Msg(fmt.Sprintf("  %s has been consumed!", village.Traps[j].Name), "system"))
			village.Traps = append(village.Traps[:j], village.Traps[j+1:]...)
		}
	}

	// Check if more waves remain
	if currentWave < numWaves {
		session.State = StateVillageTideWave
		return GameResponse{
			Type:     "menu",
			Messages: msgs,
			State:    &StateData{Screen: "village_tide_wave", Player: MakePlayerState(player)},
			Options:  []MenuOption{Opt("next", fmt.Sprintf("Next Wave (%d/%d)", currentWave+1, numWaves))},
		}
	}

	// Tide complete - calculate results
	totalDamageTaken := session.Combat.AutoPlayDeaths
	damageThreshold := village.DefenseLevel * 50

	msgs = append(msgs,
		Msg("", "system"),
		Msg("============================================================", "system"),
		Msg("TIDE DEFENSE COMPLETE!", "system"),
		Msg("============================================================", "system"),
	)

	if totalDamageTaken < damageThreshold {
		// Victory
		xpReward := 100 * numWaves
		village.Experience += xpReward

		msgs = append(msgs,
			Msg("", "system"),
			Msg("VICTORY! Your defenses held strong!", "combat"),
			Msg("", "system"),
			Msg("Battle Summary:", "system"),
			Msg(fmt.Sprintf("  Waves Defeated: %d/%d", numWaves, numWaves), "system"),
			Msg(fmt.Sprintf("  Total Monsters Killed: %d", session.Combat.AutoPlayWins), "system"),
			Msg(fmt.Sprintf("  Total Damage Dealt: %d", session.Combat.AutoPlayXP), "system"),
			Msg(fmt.Sprintf("  Total Damage Taken: %d/%d", totalDamageTaken, damageThreshold), "system"),
			Msg(fmt.Sprintf("  Traps Triggered: %d times", session.Combat.AutoPlayFights), "system"),
			Msg("", "system"),
			Msg("Rewards:", "system"),
			Msg(fmt.Sprintf("  Village XP: +%d", xpReward), "system"),
		)

		if totalDamageTaken < damageThreshold/2 {
			bonusGold := 50 + (village.Level * 10)
			goldResource := player.ResourceStorageMap["Gold"]
			goldResource.Stock += bonusGold
			player.ResourceStorageMap["Gold"] = goldResource
			msgs = append(msgs, Msg(fmt.Sprintf("  Bonus Gold: +%d (minimal damage taken!)", bonusGold), "loot"))
		}
	} else {
		// Defeat
		msgs = append(msgs,
			Msg("", "system"),
			Msg("DEFEAT! The tide overwhelmed your defenses!", "combat"),
			Msg("", "system"),
			Msg("Battle Summary:", "system"),
			Msg(fmt.Sprintf("  Waves Survived: %d/%d", numWaves, numWaves), "system"),
			Msg(fmt.Sprintf("  Total Monsters Killed: %d", session.Combat.AutoPlayWins), "system"),
			Msg(fmt.Sprintf("  Total Damage Taken: %d/%d (too much!)", totalDamageTaken, damageThreshold), "system"),
		)

		resourceLoss := village.Level * 5
		msgs = append(msgs,
			Msg("", "system"),
			Msg("Penalties:", "system"),
			Msg(fmt.Sprintf("  Lost %d of each resource type", resourceLoss), "system"),
		)

		for _, resourceType := range data.ResourceTypes {
			resource := player.ResourceStorageMap[resourceType]
			resource.Stock -= resourceLoss
			if resource.Stock < 0 {
				resource.Stock = 0
			}
			player.ResourceStorageMap[resourceType] = resource
		}

		if len(village.ActiveGuards) > 0 {
			guardsLost := 1 + rand.Intn(len(village.ActiveGuards)/2+1)
			if guardsLost > len(village.ActiveGuards) {
				guardsLost = len(village.ActiveGuards)
			}
			village.ActiveGuards = village.ActiveGuards[:len(village.ActiveGuards)-guardsLost]
			msgs = append(msgs, Msg(fmt.Sprintf("  %d hired guards were lost", guardsLost), "system"))
		}
	}

	// Update last tide time
	village.LastTideTime = time.Now().Unix()
	game.UpgradeVillage(village)

	session.Combat = nil
	e.saveVillage(session)

	session.State = StateVillageMain
	return GameResponse{
		Type:     "menu",
		Messages: msgs,
		State:    &StateData{Screen: "village_tide_complete", Player: MakePlayerState(player)},
		Options:  []MenuOption{Opt("back", "Return to Village")},
	}
}

func (e *Engine) handleVillageManageGuards(session *GameSession, cmd GameCommand) GameResponse {
	village := session.SelectedVillage

	if cmd.Value == "0" || cmd.Value == "back" {
		session.State = StateVillageMain
		return e.handleVillageMain(session, GameCommand{Type: "init"})
	}

	if len(village.ActiveGuards) == 0 {
		msgs := []GameMessage{
			Msg("============================================================", "system"),
			Msg("GUARD MANAGEMENT", "system"),
			Msg("============================================================", "system"),
			Msg("", "system"),
			Msg("No guards hired yet!", "system"),
			Msg("Hire guards from the main village menu (option 3)", "system"),
		}
		session.State = StateVillageManageGuards
		return GameResponse{
			Type:     "menu",
			Messages: msgs,
			State:    &StateData{Screen: "village_manage_guards", Player: MakePlayerState(session.Player)},
			Options:  []MenuOption{Opt("back", "Back to Village")},
		}
	}

	// If a guard was selected by number
	if cmd.Type != "init" {
		idx, err := strconv.Atoi(cmd.Value)
		if err == nil && idx >= 1 && idx <= len(village.ActiveGuards) {
			session.SelectedGuardIdx = idx - 1
			session.State = StateVillageManageGuard
			return e.handleVillageManageGuard(session, GameCommand{Type: "init"})
		}
	}

	msgs := []GameMessage{
		Msg("============================================================", "system"),
		Msg("GUARD MANAGEMENT", "system"),
		Msg("============================================================", "system"),
		Msg("", "system"),
		Msg(fmt.Sprintf("Hired Guards: %d", len(village.ActiveGuards)), "system"),
		Msg("", "system"),
	}

	options := []MenuOption{}
	for i, guard := range village.ActiveGuards {
		statusIcon := "[OK]"
		statusText := "Ready"
		if guard.Injured {
			statusIcon = "[INJURED]"
			statusText = fmt.Sprintf("Injured (%d fights to recover)", guard.RecoveryTime)
		}

		msgs = append(msgs,
			Msg(fmt.Sprintf("%d. %s %s (Lv%d)", i+1, statusIcon, guard.Name, guard.Level), "system"),
			Msg(fmt.Sprintf("   HP: %d/%d | Status: %s", guard.HitpointsRemaining, guard.HitPoints, statusText), "system"),
			Msg(fmt.Sprintf("   Attack: %d rolls (+%d) | Defense: %d rolls (+%d)",
				guard.AttackRolls, guard.AttackBonus+guard.StatsMod.AttackMod,
				guard.DefenseRolls, guard.DefenseBonus+guard.StatsMod.DefenseMod), "system"),
			Msg(fmt.Sprintf("   Equipment: %d items | Total CP: %d",
				len(guard.EquipmentMap),
				guard.StatsMod.AttackMod+guard.StatsMod.DefenseMod+guard.StatsMod.HitPointMod), "system"),
			Msg("", "system"),
		)

		options = append(options, Opt(strconv.Itoa(i+1), fmt.Sprintf("Manage %s", guard.Name)))
	}
	options = append(options, Opt("0", "Back"))

	session.State = StateVillageManageGuards
	return GameResponse{
		Type:     "menu",
		Messages: msgs,
		State:    &StateData{Screen: "village_manage_guards", Player: MakePlayerState(session.Player)},
		Options:  options,
	}
}

func (e *Engine) handleVillageManageGuard(session *GameSession, cmd GameCommand) GameResponse {
	village := session.SelectedVillage
	player := session.Player
	guardIdx := session.SelectedGuardIdx

	if guardIdx < 0 || guardIdx >= len(village.ActiveGuards) {
		session.State = StateVillageManageGuards
		return e.handleVillageManageGuards(session, GameCommand{Type: "init"})
	}

	guard := &village.ActiveGuards[guardIdx]

	if cmd.Value == "0" || cmd.Value == "back" {
		e.saveVillage(session)
		session.State = StateVillageManageGuards
		return e.handleVillageManageGuards(session, GameCommand{Type: "init"})
	}

	switch cmd.Value {
	case "1":
		session.State = StateVillageEquipGuard
		return e.handleVillageEquipGuard(session, GameCommand{Type: "init"})
	case "2":
		session.State = StateVillageUnequipGuard
		return e.handleVillageUnequipGuard(session, GameCommand{Type: "init"})
	case "3":
		session.State = StateVillageGiveItem
		return e.handleVillageGiveItem(session, GameCommand{Type: "init"})
	case "4":
		session.State = StateVillageTakeItem
		return e.handleVillageTakeItem(session, GameCommand{Type: "init"})
	case "5":
		session.State = StateVillageHealGuard
		return e.handleVillageHealGuard(session, GameCommand{Type: "init"})
	}

	// Show guard details
	statusText := "Ready for combat"
	if guard.Injured {
		statusText = fmt.Sprintf("Injured - %d fights until recovery", guard.RecoveryTime)
	}

	msgs := []GameMessage{
		Msg("============================================================", "system"),
		Msg(fmt.Sprintf("%s (Level %d)", guard.Name, guard.Level), "system"),
		Msg("============================================================", "system"),
		Msg("", "system"),
		Msg(fmt.Sprintf("Status: %s", statusText), "system"),
		Msg(fmt.Sprintf("HP: %d/%d (Natural: %d)", guard.HitpointsRemaining, guard.HitPoints, guard.HitpointsNatural), "system"),
		Msg(fmt.Sprintf("Attack: %d rolls + %d bonus = %d total",
			guard.AttackRolls, guard.AttackBonus+guard.StatsMod.AttackMod,
			guard.AttackRolls*6+guard.AttackBonus+guard.StatsMod.AttackMod), "system"),
		Msg(fmt.Sprintf("Defense: %d rolls + %d bonus = %d total",
			guard.DefenseRolls, guard.DefenseBonus+guard.StatsMod.DefenseMod,
			guard.DefenseRolls*6+guard.DefenseBonus+guard.StatsMod.DefenseMod), "system"),
	}

	// Equipment
	msgs = append(msgs, Msg("", "system"))
	msgs = append(msgs, Msg("--- EQUIPPED ITEMS ---", "system"))
	slotNames := map[int]string{
		0: "Head", 1: "Chest", 2: "Legs", 3: "Feet",
		4: "Hands", 5: "Main Hand", 6: "Off Hand", 7: "Accessory",
	}
	if len(guard.EquipmentMap) == 0 {
		msgs = append(msgs, Msg("No equipment", "system"))
	} else {
		for slot, item := range guard.EquipmentMap {
			slotName := slotNames[slot]
			if slotName == "" {
				slotName = fmt.Sprintf("Slot %d", slot)
			}
			msgs = append(msgs, Msg(fmt.Sprintf("[%s] %s (CP:%d, Rarity:%d)", slotName, item.Name, item.CP, item.Rarity), "system"))
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
	}

	// Inventory
	msgs = append(msgs, Msg("", "system"))
	msgs = append(msgs, Msg("--- INVENTORY ---", "system"))
	if len(guard.Inventory) == 0 {
		msgs = append(msgs, Msg("Empty", "system"))
	} else {
		for i, item := range guard.Inventory {
			msgs = append(msgs, Msg(fmt.Sprintf("%d. %s (CP:%d, Rarity:%d, Slot:%d)", i+1, item.Name, item.CP, item.Rarity, item.Slot), "system"))
		}
	}

	// Equipment stats summary
	msgs = append(msgs, Msg("", "system"))
	msgs = append(msgs, Msg("--- EQUIPMENT BONUS ---", "system"))
	msgs = append(msgs, Msg(fmt.Sprintf("Total: +%d Attack, +%d Defense, +%d HP",
		guard.StatsMod.AttackMod, guard.StatsMod.DefenseMod, guard.StatsMod.HitPointMod), "system"))

	options := []MenuOption{
		Opt("1", "Equip Item from Inventory"),
		Opt("2", "Unequip Item to Inventory"),
		Opt("3", "Give Item from Player"),
		Opt("4", "Take Item to Player"),
		Opt("5", "Heal Guard (costs 1 health potion)"),
		Opt("0", "Back"),
	}

	_ = player // used via session

	session.State = StateVillageManageGuard
	return GameResponse{
		Type:     "menu",
		Messages: msgs,
		State:    &StateData{Screen: "village_manage_guard", Player: MakePlayerState(player)},
		Options:  options,
	}
}

func (e *Engine) handleVillageEquipGuard(session *GameSession, cmd GameCommand) GameResponse {
	village := session.SelectedVillage
	guardIdx := session.SelectedGuardIdx

	if guardIdx < 0 || guardIdx >= len(village.ActiveGuards) {
		session.State = StateVillageManageGuard
		return e.handleVillageManageGuard(session, GameCommand{Type: "init"})
	}

	guard := &village.ActiveGuards[guardIdx]

	if cmd.Value == "0" || cmd.Value == "back" {
		session.State = StateVillageManageGuard
		return e.handleVillageManageGuard(session, GameCommand{Type: "init"})
	}

	if len(guard.Inventory) == 0 {
		session.State = StateVillageManageGuard
		resp := e.handleVillageManageGuard(session, GameCommand{Type: "init"})
		resp.Messages = append([]GameMessage{Msg("Guard has no items in inventory!", "error")}, resp.Messages...)
		return resp
	}

	if cmd.Type != "init" {
		idx, err := strconv.Atoi(cmd.Value)
		if err == nil && idx >= 1 && idx <= len(guard.Inventory) {
			item := guard.Inventory[idx-1]
			game.EquipGuardItem(item, &guard.EquipmentMap, &guard.Inventory)

			guard.StatsMod = game.CalculateItemMods(guard.EquipmentMap)
			guard.HitPoints = guard.HitpointsNatural + guard.StatsMod.HitPointMod
			if guard.HitpointsRemaining > guard.HitPoints {
				guard.HitpointsRemaining = guard.HitPoints
			}

			e.saveVillage(session)

			session.State = StateVillageManageGuard
			resp := e.handleVillageManageGuard(session, GameCommand{Type: "init"})
			resp.Messages = append([]GameMessage{Msg(fmt.Sprintf("Equipped %s", item.Name), "system")}, resp.Messages...)
			return resp
		}
	}

	msgs := []GameMessage{
		Msg("Select item to equip:", "system"),
	}
	options := []MenuOption{}
	for i, item := range guard.Inventory {
		options = append(options, Opt(strconv.Itoa(i+1), fmt.Sprintf("%s (CP:%d, Slot:%d)", item.Name, item.CP, item.Slot)))
	}
	options = append(options, Opt("0", "Cancel"))

	session.State = StateVillageEquipGuard
	return GameResponse{
		Type:     "menu",
		Messages: msgs,
		State:    &StateData{Screen: "village_equip_guard", Player: MakePlayerState(session.Player)},
		Options:  options,
	}
}

func (e *Engine) handleVillageUnequipGuard(session *GameSession, cmd GameCommand) GameResponse {
	village := session.SelectedVillage
	guardIdx := session.SelectedGuardIdx

	if guardIdx < 0 || guardIdx >= len(village.ActiveGuards) {
		session.State = StateVillageManageGuard
		return e.handleVillageManageGuard(session, GameCommand{Type: "init"})
	}

	guard := &village.ActiveGuards[guardIdx]

	if cmd.Value == "0" || cmd.Value == "back" {
		session.State = StateVillageManageGuard
		return e.handleVillageManageGuard(session, GameCommand{Type: "init"})
	}

	if len(guard.EquipmentMap) == 0 {
		session.State = StateVillageManageGuard
		resp := e.handleVillageManageGuard(session, GameCommand{Type: "init"})
		resp.Messages = append([]GameMessage{Msg("Guard has no equipped items!", "error")}, resp.Messages...)
		return resp
	}

	// Build sorted slot list for consistent ordering
	slots := []int{}
	for slot := range guard.EquipmentMap {
		slots = append(slots, slot)
	}
	sort.Ints(slots)

	if cmd.Type != "init" {
		idx, err := strconv.Atoi(cmd.Value)
		if err == nil && idx >= 1 && idx <= len(slots) {
			slot := slots[idx-1]
			item := guard.EquipmentMap[slot]
			guard.Inventory = append(guard.Inventory, item)
			delete(guard.EquipmentMap, slot)

			guard.StatsMod = game.CalculateItemMods(guard.EquipmentMap)
			guard.HitPoints = guard.HitpointsNatural + guard.StatsMod.HitPointMod
			if guard.HitpointsRemaining > guard.HitPoints {
				guard.HitpointsRemaining = guard.HitPoints
			}

			e.saveVillage(session)

			session.State = StateVillageManageGuard
			resp := e.handleVillageManageGuard(session, GameCommand{Type: "init"})
			resp.Messages = append([]GameMessage{Msg(fmt.Sprintf("Unequipped %s", item.Name), "system")}, resp.Messages...)
			return resp
		}
	}

	slotNames := map[int]string{
		0: "Head", 1: "Chest", 2: "Legs", 3: "Feet",
		4: "Hands", 5: "Main Hand", 6: "Off Hand", 7: "Accessory",
	}

	msgs := []GameMessage{
		Msg("Select slot to unequip:", "system"),
	}
	options := []MenuOption{}
	for i, slot := range slots {
		item := guard.EquipmentMap[slot]
		slotName := slotNames[slot]
		if slotName == "" {
			slotName = fmt.Sprintf("Slot %d", slot)
		}
		options = append(options, Opt(strconv.Itoa(i+1), fmt.Sprintf("[%s] %s (CP:%d)", slotName, item.Name, item.CP)))
	}
	options = append(options, Opt("0", "Cancel"))

	session.State = StateVillageUnequipGuard
	return GameResponse{
		Type:     "menu",
		Messages: msgs,
		State:    &StateData{Screen: "village_unequip_guard", Player: MakePlayerState(session.Player)},
		Options:  options,
	}
}

func (e *Engine) handleVillageGiveItem(session *GameSession, cmd GameCommand) GameResponse {
	village := session.SelectedVillage
	player := session.Player
	guardIdx := session.SelectedGuardIdx

	if guardIdx < 0 || guardIdx >= len(village.ActiveGuards) {
		session.State = StateVillageManageGuard
		return e.handleVillageManageGuard(session, GameCommand{Type: "init"})
	}

	guard := &village.ActiveGuards[guardIdx]

	if cmd.Value == "0" || cmd.Value == "back" {
		session.State = StateVillageManageGuard
		return e.handleVillageManageGuard(session, GameCommand{Type: "init"})
	}

	// Filter equipment items
	equipment := []models.Item{}
	equipmentIndices := []int{}
	for i, item := range player.Inventory {
		if item.ItemType == "equipment" || item.ItemType == "" {
			equipment = append(equipment, item)
			equipmentIndices = append(equipmentIndices, i)
		}
	}

	if len(equipment) == 0 {
		session.State = StateVillageManageGuard
		resp := e.handleVillageManageGuard(session, GameCommand{Type: "init"})
		resp.Messages = append([]GameMessage{Msg("You have no equipment items to give!", "error")}, resp.Messages...)
		return resp
	}

	if cmd.Type != "init" {
		idx, err := strconv.Atoi(cmd.Value)
		if err == nil && idx >= 1 && idx <= len(equipment) {
			item := equipment[idx-1]
			originalIdx := equipmentIndices[idx-1]

			guard.Inventory = append(guard.Inventory, item)
			game.RemoveItemFromInventory(&player.Inventory, originalIdx)

			e.saveVillage(session)

			session.State = StateVillageManageGuard
			resp := e.handleVillageManageGuard(session, GameCommand{Type: "init"})
			resp.Messages = append([]GameMessage{Msg(fmt.Sprintf("Gave %s to %s", item.Name, guard.Name), "system")}, resp.Messages...)
			return resp
		}
	}

	msgs := []GameMessage{
		Msg("Select item to give to guard:", "system"),
	}
	options := []MenuOption{}
	for i, item := range equipment {
		options = append(options, Opt(strconv.Itoa(i+1), fmt.Sprintf("%s (CP:%d, Rarity:%d)", item.Name, item.CP, item.Rarity)))
	}
	options = append(options, Opt("0", "Cancel"))

	session.State = StateVillageGiveItem
	return GameResponse{
		Type:     "menu",
		Messages: msgs,
		State:    &StateData{Screen: "village_give_item", Player: MakePlayerState(player)},
		Options:  options,
	}
}

func (e *Engine) handleVillageTakeItem(session *GameSession, cmd GameCommand) GameResponse {
	village := session.SelectedVillage
	player := session.Player
	guardIdx := session.SelectedGuardIdx

	if guardIdx < 0 || guardIdx >= len(village.ActiveGuards) {
		session.State = StateVillageManageGuard
		return e.handleVillageManageGuard(session, GameCommand{Type: "init"})
	}

	guard := &village.ActiveGuards[guardIdx]

	if cmd.Value == "0" || cmd.Value == "back" {
		session.State = StateVillageManageGuard
		return e.handleVillageManageGuard(session, GameCommand{Type: "init"})
	}

	if len(guard.Inventory) == 0 {
		session.State = StateVillageManageGuard
		resp := e.handleVillageManageGuard(session, GameCommand{Type: "init"})
		resp.Messages = append([]GameMessage{Msg("Guard has no items to take!", "error")}, resp.Messages...)
		return resp
	}

	if cmd.Type != "init" {
		idx, err := strconv.Atoi(cmd.Value)
		if err == nil && idx >= 1 && idx <= len(guard.Inventory) {
			item := guard.Inventory[idx-1]

			player.Inventory = append(player.Inventory, item)
			game.RemoveItemFromInventory(&guard.Inventory, idx-1)

			e.saveVillage(session)

			session.State = StateVillageManageGuard
			resp := e.handleVillageManageGuard(session, GameCommand{Type: "init"})
			resp.Messages = append([]GameMessage{Msg(fmt.Sprintf("Took %s from %s", item.Name, guard.Name), "system")}, resp.Messages...)
			return resp
		}
	}

	msgs := []GameMessage{
		Msg("Select item to take from guard:", "system"),
	}
	options := []MenuOption{}
	for i, item := range guard.Inventory {
		options = append(options, Opt(strconv.Itoa(i+1), fmt.Sprintf("%s (CP:%d, Rarity:%d)", item.Name, item.CP, item.Rarity)))
	}
	options = append(options, Opt("0", "Cancel"))

	session.State = StateVillageTakeItem
	return GameResponse{
		Type:     "menu",
		Messages: msgs,
		State:    &StateData{Screen: "village_take_item", Player: MakePlayerState(player)},
		Options:  options,
	}
}

func (e *Engine) handleVillageHealGuard(session *GameSession, cmd GameCommand) GameResponse {
	village := session.SelectedVillage
	player := session.Player
	guardIdx := session.SelectedGuardIdx

	if guardIdx < 0 || guardIdx >= len(village.ActiveGuards) {
		session.State = StateVillageManageGuard
		return e.handleVillageManageGuard(session, GameCommand{Type: "init"})
	}

	guard := &village.ActiveGuards[guardIdx]

	if cmd.Value == "0" || cmd.Value == "back" || cmd.Value == "n" {
		session.State = StateVillageManageGuard
		return e.handleVillageManageGuard(session, GameCommand{Type: "init"})
	}

	if !guard.Injured && guard.HitpointsRemaining >= guard.HitPoints {
		session.State = StateVillageManageGuard
		resp := e.handleVillageManageGuard(session, GameCommand{Type: "init"})
		resp.Messages = append([]GameMessage{Msg("Guard is already at full health!", "system")}, resp.Messages...)
		return resp
	}

	// Find a health potion
	potionIdx := -1
	for i, item := range player.Inventory {
		if item.ItemType == "consumable" && item.Consumable.EffectType == "heal" {
			potionIdx = i
			break
		}
	}

	if potionIdx == -1 {
		session.State = StateVillageManageGuard
		resp := e.handleVillageManageGuard(session, GameCommand{Type: "init"})
		resp.Messages = append([]GameMessage{Msg("You don't have any health potions!", "error")}, resp.Messages...)
		return resp
	}

	if cmd.Type == "init" {
		potion := player.Inventory[potionIdx]
		msgs := []GameMessage{
			Msg(fmt.Sprintf("Use %s on %s?", potion.Name, guard.Name), "system"),
			Msg(fmt.Sprintf("Guard HP: %d/%d", guard.HitpointsRemaining, guard.HitPoints), "system"),
			Msg(fmt.Sprintf("Potion heals: %d HP", potion.Consumable.Value), "system"),
		}

		session.State = StateVillageHealGuard
		return GameResponse{
			Type:     "menu",
			Messages: msgs,
			State:    &StateData{Screen: "village_heal_guard", Player: MakePlayerState(player)},
			Options: []MenuOption{
				Opt("y", "Yes, heal"),
				Opt("n", "No, cancel"),
			},
		}
	}

	if cmd.Value == "y" {
		potion := player.Inventory[potionIdx]
		healAmount := potion.Consumable.Value

		guard.HitpointsRemaining += healAmount
		if guard.HitpointsRemaining > guard.HitPoints {
			guard.HitpointsRemaining = guard.HitPoints
		}

		if guard.HitpointsRemaining >= guard.HitPoints {
			guard.Injured = false
			guard.RecoveryTime = 0
		}

		game.RemoveItemFromInventory(&player.Inventory, potionIdx)

		e.saveVillage(session)

		resultMsgs := []GameMessage{
			Msg(fmt.Sprintf("Used %s on %s", potion.Name, guard.Name), "heal"),
			Msg(fmt.Sprintf("   %s healed for %d HP! (%d/%d)", guard.Name, healAmount, guard.HitpointsRemaining, guard.HitPoints), "heal"),
		}
		if !guard.Injured {
			resultMsgs = append(resultMsgs, Msg("   Guard is now fully recovered!", "heal"))
		}

		session.State = StateVillageManageGuard
		resp := e.handleVillageManageGuard(session, GameCommand{Type: "init"})
		resp.Messages = append(resultMsgs, resp.Messages...)
		return resp
	}

	session.State = StateVillageManageGuard
	return e.handleVillageManageGuard(session, GameCommand{Type: "init"})
}

// handleVillageFortifications lets the player craft village defense items using resources.
func (e *Engine) handleVillageFortifications(session *GameSession, cmd GameCommand) GameResponse {
	village := session.SelectedVillage
	player := session.Player

	if cmd.Value == "0" || cmd.Value == "back" {
		session.State = StateVillageCrafting
		return e.handleVillageCrafting(session, GameCommand{Type: "init"})
	}

	type fortRecipe struct {
		Name     string
		Cost     map[string]int
		Defense  int
		HitPoint int
	}
	recipes := []fortRecipe{
		{Name: "Wooden Barricade", Cost: map[string]int{"Lumber": 20}, Defense: 5, HitPoint: 50},
		{Name: "Stone Wall", Cost: map[string]int{"Stone": 30, "Lumber": 10}, Defense: 15, HitPoint: 150},
		{Name: "Guard Tower", Cost: map[string]int{"Stone": 20, "Lumber": 20, "Iron": 10}, Defense: 25, HitPoint: 200},
	}

	idx, err := strconv.Atoi(cmd.Value)
	if err == nil && idx >= 1 && idx <= len(recipes) {
		recipe := recipes[idx-1]
		canBuild := true
		for res, amount := range recipe.Cost {
			r, exists := player.ResourceStorageMap[res]
			if !exists || r.Stock < amount {
				canBuild = false
				break
			}
		}
		if !canBuild {
			msgs := []GameMessage{Msg(fmt.Sprintf("Not enough resources to build %s!", recipe.Name), "error")}
			session.State = StateVillageFortifications
			return GameResponse{
				Type:     "menu",
				Messages: msgs,
				State:    &StateData{Screen: "village_fortifications", Player: MakePlayerState(player), Village: MakeVillageView(village)},
				Options:  []MenuOption{Opt("back", "Back")},
			}
		}
		for res, amount := range recipe.Cost {
			r := player.ResourceStorageMap[res]
			r.Stock -= amount
			player.ResourceStorageMap[res] = r
		}
		village.Defenses = append(village.Defenses, models.Defense{
			Name:    recipe.Name,
			Level:   1,
			Defense: recipe.Defense,
			Built:   true,
			Type:    "fortification",
		})
		village.DefenseLevel += recipe.Defense
		e.saveVillage(session)
		msgs := []GameMessage{
			Msg(fmt.Sprintf("Built %s! (+%d defense, +%d HP)", recipe.Name, recipe.Defense, recipe.HitPoint), "system"),
		}
		session.State = StateVillageFortifications
		return e.handleVillageFortifications(session, GameCommand{Type: "init"})
		_ = msgs // messages shown via init re-render
	}

	// Show fortifications menu
	msgs := []GameMessage{
		Msg("============================================================", "system"),
		Msg("FORTIFICATIONS - Village Defense Crafting", "system"),
		Msg("============================================================", "system"),
		Msg(fmt.Sprintf("Current Defense Level: %d", village.DefenseLevel), "system"),
	}
	options := []MenuOption{}
	for i, recipe := range recipes {
		costStr := ""
		for res, amount := range recipe.Cost {
			if costStr != "" {
				costStr += ", "
			}
			costStr += fmt.Sprintf("%d %s", amount, res)
		}
		options = append(options, Opt(strconv.Itoa(i+1), fmt.Sprintf("%s (+%d DEF, +%d HP) [%s]", recipe.Name, recipe.Defense, recipe.HitPoint, costStr)))
	}
	options = append(options, Opt("0", "Back"))

	session.State = StateVillageFortifications
	return GameResponse{
		Type:     "menu",
		Messages: msgs,
		State:    &StateData{Screen: "village_fortifications", Player: MakePlayerState(player), Village: MakeVillageView(village)},
		Options:  options,
	}
}

// handleVillageTraining lets the player spend resources to level up a villager.
func (e *Engine) handleVillageTraining(session *GameSession, cmd GameCommand) GameResponse {
	village := session.SelectedVillage
	player := session.Player

	if cmd.Value == "0" || cmd.Value == "back" {
		session.State = StateVillageCrafting
		return e.handleVillageCrafting(session, GameCommand{Type: "init"})
	}

	trainingCost := map[string]int{"Gold": 10, "Lumber": 5}

	idx, err := strconv.Atoi(cmd.Value)
	if err == nil && idx >= 1 && idx <= len(village.Villagers) {
		canTrain := true
		for res, amount := range trainingCost {
			r, exists := player.ResourceStorageMap[res]
			if !exists || r.Stock < amount {
				canTrain = false
				break
			}
		}
		if !canTrain {
			msgs := []GameMessage{Msg("Not enough resources to train! (10 Gold, 5 Lumber)", "error")}
			session.State = StateVillageTraining
			return GameResponse{
				Type:     "menu",
				Messages: msgs,
				State:    &StateData{Screen: "village_training", Player: MakePlayerState(player), Village: MakeVillageView(village)},
				Options:  []MenuOption{Opt("back", "Back")},
			}
		}
		for res, amount := range trainingCost {
			r := player.ResourceStorageMap[res]
			r.Stock -= amount
			player.ResourceStorageMap[res] = r
		}
		v := &village.Villagers[idx-1]
		v.Level++
		v.Efficiency++
		e.saveVillage(session)
		msgs := []GameMessage{
			Msg(fmt.Sprintf("%s trained to Level %d! (Efficiency: %d)", v.Name, v.Level, v.Efficiency), "system"),
		}
		session.State = StateVillageTraining
		resp := e.handleVillageTraining(session, GameCommand{Type: "init"})
		resp.Messages = append(msgs, resp.Messages...)
		return resp
	}

	// Show training menu
	msgs := []GameMessage{
		Msg("============================================================", "system"),
		Msg("VILLAGER TRAINING", "system"),
		Msg("============================================================", "system"),
		Msg("Cost per training: 10 Gold, 5 Lumber", "system"),
		Msg("", "system"),
	}
	options := []MenuOption{}
	for i, v := range village.Villagers {
		options = append(options, Opt(strconv.Itoa(i+1), fmt.Sprintf("%s (Lv%d %s, Eff: %d)", v.Name, v.Level, v.Role, v.Efficiency)))
	}
	if len(village.Villagers) == 0 {
		msgs = append(msgs, Msg("No villagers to train.", "system"))
	}
	options = append(options, Opt("0", "Back"))

	session.State = StateVillageTraining
	return GameResponse{
		Type:     "menu",
		Messages: msgs,
		State:    &StateData{Screen: "village_training", Player: MakePlayerState(player), Village: MakeVillageView(village)},
		Options:  options,
	}
}

// handleVillageHealing lets the player spend resources to restore HP/MP/SP.
func (e *Engine) handleVillageHealing(session *GameSession, cmd GameCommand) GameResponse {
	village := session.SelectedVillage
	player := session.Player

	if cmd.Value == "0" || cmd.Value == "back" {
		session.State = StateVillageCrafting
		return e.handleVillageCrafting(session, GameCommand{Type: "init"})
	}

	switch cmd.Value {
	case "1": // Restore HP
		cost := 5
		r, exists := player.ResourceStorageMap["Gold"]
		if !exists || r.Stock < cost {
			msgs := []GameMessage{Msg("Not enough Gold! (5 required)", "error")}
			session.State = StateVillageHealing
			return GameResponse{
				Type:     "menu",
				Messages: msgs,
				State:    &StateData{Screen: "village_healing", Player: MakePlayerState(player), Village: MakeVillageView(village)},
				Options:  []MenuOption{Opt("back", "Back")},
			}
		}
		r.Stock -= cost
		player.ResourceStorageMap["Gold"] = r
		player.HitpointsRemaining = player.HitpointsTotal
		e.saveVillage(session)
		msgs := []GameMessage{Msg(fmt.Sprintf("HP fully restored! (%d/%d)", player.HitpointsRemaining, player.HitpointsTotal), "heal")}
		session.State = StateVillageHealing
		resp := e.handleVillageHealing(session, GameCommand{Type: "init"})
		resp.Messages = append(msgs, resp.Messages...)
		return resp

	case "2": // Restore MP
		cost := 5
		r, exists := player.ResourceStorageMap["Gold"]
		if !exists || r.Stock < cost {
			msgs := []GameMessage{Msg("Not enough Gold! (5 required)", "error")}
			session.State = StateVillageHealing
			return GameResponse{
				Type:     "menu",
				Messages: msgs,
				State:    &StateData{Screen: "village_healing", Player: MakePlayerState(player), Village: MakeVillageView(village)},
				Options:  []MenuOption{Opt("back", "Back")},
			}
		}
		r.Stock -= cost
		player.ResourceStorageMap["Gold"] = r
		player.ManaRemaining = player.ManaTotal
		e.saveVillage(session)
		msgs := []GameMessage{Msg(fmt.Sprintf("MP fully restored! (%d/%d)", player.ManaRemaining, player.ManaTotal), "heal")}
		session.State = StateVillageHealing
		resp := e.handleVillageHealing(session, GameCommand{Type: "init"})
		resp.Messages = append(msgs, resp.Messages...)
		return resp

	case "3": // Restore SP
		cost := 5
		r, exists := player.ResourceStorageMap["Gold"]
		if !exists || r.Stock < cost {
			msgs := []GameMessage{Msg("Not enough Gold! (5 required)", "error")}
			session.State = StateVillageHealing
			return GameResponse{
				Type:     "menu",
				Messages: msgs,
				State:    &StateData{Screen: "village_healing", Player: MakePlayerState(player), Village: MakeVillageView(village)},
				Options:  []MenuOption{Opt("back", "Back")},
			}
		}
		r.Stock -= cost
		player.ResourceStorageMap["Gold"] = r
		player.StaminaRemaining = player.StaminaTotal
		e.saveVillage(session)
		msgs := []GameMessage{Msg(fmt.Sprintf("SP fully restored! (%d/%d)", player.StaminaRemaining, player.StaminaTotal), "heal")}
		session.State = StateVillageHealing
		resp := e.handleVillageHealing(session, GameCommand{Type: "init"})
		resp.Messages = append(msgs, resp.Messages...)
		return resp

	case "4": // Restore All
		cost := 12
		r, exists := player.ResourceStorageMap["Gold"]
		if !exists || r.Stock < cost {
			msgs := []GameMessage{Msg("Not enough Gold! (12 required)", "error")}
			session.State = StateVillageHealing
			return GameResponse{
				Type:     "menu",
				Messages: msgs,
				State:    &StateData{Screen: "village_healing", Player: MakePlayerState(player), Village: MakeVillageView(village)},
				Options:  []MenuOption{Opt("back", "Back")},
			}
		}
		r.Stock -= cost
		player.ResourceStorageMap["Gold"] = r
		player.HitpointsRemaining = player.HitpointsTotal
		player.ManaRemaining = player.ManaTotal
		player.StaminaRemaining = player.StaminaTotal
		e.saveVillage(session)
		msgs := []GameMessage{Msg("All resources fully restored!", "heal")}
		session.State = StateVillageHealing
		resp := e.handleVillageHealing(session, GameCommand{Type: "init"})
		resp.Messages = append(msgs, resp.Messages...)
		return resp
	}

	// Show healing menu
	msgs := []GameMessage{
		Msg("============================================================", "system"),
		Msg("HEALING SERVICES", "system"),
		Msg("============================================================", "system"),
		Msg(fmt.Sprintf("HP: %d/%d | MP: %d/%d | SP: %d/%d",
			player.HitpointsRemaining, player.HitpointsTotal,
			player.ManaRemaining, player.ManaTotal,
			player.StaminaRemaining, player.StaminaTotal), "system"),
		Msg("", "system"),
	}
	options := []MenuOption{
		Opt("1", fmt.Sprintf("Restore HP (5 Gold) [%d/%d]", player.HitpointsRemaining, player.HitpointsTotal)),
		Opt("2", fmt.Sprintf("Restore MP (5 Gold) [%d/%d]", player.ManaRemaining, player.ManaTotal)),
		Opt("3", fmt.Sprintf("Restore SP (5 Gold) [%d/%d]", player.StaminaRemaining, player.StaminaTotal)),
		Opt("4", "Restore All (12 Gold)"),
		Opt("0", "Back"),
	}

	session.State = StateVillageHealing
	return GameResponse{
		Type:     "menu",
		Messages: msgs,
		State:    &StateData{Screen: "village_healing", Player: MakePlayerState(player), Village: MakeVillageView(village)},
		Options:  options,
	}
}
