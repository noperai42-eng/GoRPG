package game

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"

	"rpg-game/pkg/data"
	"rpg-game/pkg/models"
)

// ShowVillageMenu displays the main village management menu and handles user input.
func ShowVillageMenu(gameState *models.GameState, player *models.Character, village *models.Village) {
	scanner := bufio.NewScanner(os.Stdin)

	for {
		// Process auto-collection when entering village menu
		ProcessVillageResourceCollection(village, player)
		UpgradeVillage(village)

		fmt.Println("\n============================================================")
		fmt.Printf("  %s - Level %d\n", village.Name, village.Level)
		fmt.Println("============================================================")
		fmt.Printf("Experience: %d/%d\n", village.Experience, village.Level*100)
		fmt.Printf("Villagers: %d (Harvesters: %d, Guards: %d)\n",
			len(village.Villagers),
			CountVillagersByRole(village, "harvester"),
			CountVillagersByRole(village, "guard"))
		fmt.Printf("Hired Guards: %d\n", len(village.ActiveGuards))
		fmt.Printf("Defenses Built: %d (Level %d)\n", len(village.Defenses), village.DefenseLevel)

		// Show unlocked crafting
		if len(village.UnlockedCrafting) > 0 {
			fmt.Print("Unlocked Crafting: ")
			for i, craft := range village.UnlockedCrafting {
				if i > 0 {
					fmt.Print(", ")
				}
				fmt.Print(craft)
			}
			fmt.Println()
		}

		fmt.Println("\n--- Village Management ---")
		fmt.Println("1 = View Villagers")
		fmt.Println("2 = Assign Harvester Tasks")
		fmt.Println("3 = Hire Guards")
		fmt.Println("4 = Crafting")
		fmt.Println("5 = Build Defenses")
		fmt.Println("6 = Check Next Monster Tide")
		fmt.Println("7 = Defend Against Tide (if ready)")
		fmt.Println("8 = Manage Guards (Equipment & Status)")
		fmt.Println("0 = Return to Main Menu")
		fmt.Print("Choice: ")

		scanner.Scan()
		choice := scanner.Text()

		switch choice {
		case "1":
			viewVillagers(village)
		case "2":
			assignVillagerTask(village, player)
		case "3":
			hireGuardMenu(village, player)
		case "4":
			craftingMenu(village, player)
		case "5":
			buildDefenseMenu(village, player)
		case "6":
			checkMonsterTide(village)
		case "7":
			// Check if tide is ready
			currentTime := time.Now().Unix()
			timeSinceLastTide := currentTime - village.LastTideTime
			timeUntilNext := village.TideInterval - int(timeSinceLastTide)

			if timeUntilNext <= 0 {
				MonsterTideDefense(gameState, player, village)
				// Save after tide
				gameState.CharactersMap[player.Name] = *player
				gameState.Villages[player.VillageName] = *village
				WriteGameStateToFile(*gameState, "gamestate.json")
			} else {
				hours := timeUntilNext / 3600
				minutes := (timeUntilNext % 3600) / 60
				fmt.Printf("\nTide not ready yet! Wait %d hours, %d minutes\n", hours, minutes)
			}
		case "8":
			ManageGuardsMenu(gameState, player, village)
		case "0":
			// Save and return
			gameState.CharactersMap[player.Name] = *player
			gameState.Villages[player.VillageName] = *village
			WriteGameStateToFile(*gameState, "gamestate.json")
			fmt.Println("\nVillage saved")
			return
		default:
			fmt.Println("Invalid choice")
		}
	}
}

// ManageGuardsMenu displays the guard management interface, listing all hired
// guards and allowing the player to select one to manage individually.
func ManageGuardsMenu(gameState *models.GameState, player *models.Character, village *models.Village) {
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Println("\n============================================================")
		fmt.Println("GUARD MANAGEMENT")
		fmt.Println("============================================================")

		if len(village.ActiveGuards) == 0 {
			fmt.Println("\nNo guards hired yet!")
			fmt.Println("Hire guards from the main village menu (option 3)")
			fmt.Print("\nPress ENTER to return...")
			scanner.Scan()
			return
		}

		fmt.Printf("\nHired Guards: %d\n\n", len(village.ActiveGuards))

		// Display all guards with status
		for i, guard := range village.ActiveGuards {
			statusIcon := "[OK]"
			statusText := "Ready"
			if guard.Injured {
				statusIcon = "[INJURED]"
				statusText = fmt.Sprintf("Injured (%d fights to recover)", guard.RecoveryTime)
			}

			fmt.Printf("%d. %s %s (Lv%d)\n", i+1, statusIcon, guard.Name, guard.Level)
			fmt.Printf("   HP: %d/%d | Status: %s\n",
				guard.HitpointsRemaining, guard.HitPoints, statusText)
			fmt.Printf("   Attack: %d rolls (+%d) | Defense: %d rolls (+%d)\n",
				guard.AttackRolls, guard.AttackBonus+guard.StatsMod.AttackMod,
				guard.DefenseRolls, guard.DefenseBonus+guard.StatsMod.DefenseMod)
			fmt.Printf("   Equipment: %d items | Total CP: %d\n",
				len(guard.EquipmentMap),
				guard.StatsMod.AttackMod+guard.StatsMod.DefenseMod+guard.StatsMod.HitPointMod)
			fmt.Println()
		}

		fmt.Println("Select a guard to manage (0=back):")
		fmt.Print("Choice: ")
		scanner.Scan()
		choice := scanner.Text()
		idx, err := strconv.Atoi(choice)

		if err != nil || idx < 0 || idx > len(village.ActiveGuards) {
			fmt.Println("Invalid choice!")
			continue
		}
		if idx == 0 {
			return
		}

		// Manage selected guard
		manageIndividualGuard(&village.ActiveGuards[idx-1], player)

		// Save village after any changes
		gameState.Villages[player.VillageName] = *village
	}
}

func manageIndividualGuard(guard *models.Guard, player *models.Character) {
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Println("\n============================================================")
		fmt.Printf("%s (Level %d)\n", guard.Name, guard.Level)
		fmt.Println("============================================================")

		// Show status
		statusText := "Ready for combat"
		if guard.Injured {
			statusText = fmt.Sprintf("Injured - %d fights until recovery", guard.RecoveryTime)
		}

		fmt.Printf("\nStatus: %s\n", statusText)
		fmt.Printf("HP: %d/%d (Natural: %d)\n", guard.HitpointsRemaining, guard.HitPoints, guard.HitpointsNatural)
		fmt.Printf("Attack: %d rolls + %d bonus = %d total\n",
			guard.AttackRolls, guard.AttackBonus+guard.StatsMod.AttackMod,
			guard.AttackRolls*6+guard.AttackBonus+guard.StatsMod.AttackMod)
		fmt.Printf("Defense: %d rolls + %d bonus = %d total\n",
			guard.DefenseRolls, guard.DefenseBonus+guard.StatsMod.DefenseMod,
			guard.DefenseRolls*6+guard.DefenseBonus+guard.StatsMod.DefenseMod)

		// Show equipment
		fmt.Println("\n--- EQUIPPED ITEMS ---")
		if len(guard.EquipmentMap) == 0 {
			fmt.Println("No equipment")
		} else {
			slotNames := map[int]string{
				0: "Head", 1: "Chest", 2: "Legs", 3: "Feet",
				4: "Hands", 5: "Main Hand", 6: "Off Hand", 7: "Accessory",
			}
			for slot, item := range guard.EquipmentMap {
				slotName := slotNames[slot]
				if slotName == "" {
					slotName = fmt.Sprintf("Slot %d", slot)
				}
				fmt.Printf("[%s] %s (CP:%d, Rarity:%d)\n", slotName, item.Name, item.CP, item.Rarity)
				if item.StatsMod.AttackMod > 0 {
					fmt.Printf("  +%d Attack\n", item.StatsMod.AttackMod)
				}
				if item.StatsMod.DefenseMod > 0 {
					fmt.Printf("  +%d Defense\n", item.StatsMod.DefenseMod)
				}
				if item.StatsMod.HitPointMod > 0 {
					fmt.Printf("  +%d HP\n", item.StatsMod.HitPointMod)
				}
			}
		}

		// Show inventory
		fmt.Println("\n--- INVENTORY ---")
		if len(guard.Inventory) == 0 {
			fmt.Println("Empty")
		} else {
			for i, item := range guard.Inventory {
				fmt.Printf("%d. %s (CP:%d, Rarity:%d, Slot:%d)\n", i+1, item.Name, item.CP, item.Rarity, item.Slot)
			}
		}

		// Show equipment stats summary
		fmt.Println("\n--- EQUIPMENT BONUS ---")
		fmt.Printf("Total: +%d Attack, +%d Defense, +%d HP\n",
			guard.StatsMod.AttackMod, guard.StatsMod.DefenseMod, guard.StatsMod.HitPointMod)

		// Menu options
		fmt.Println("\n--- OPTIONS ---")
		fmt.Println("1 = Equip Item from Inventory")
		fmt.Println("2 = Unequip Item to Inventory")
		fmt.Println("3 = Give Item from Player")
		fmt.Println("4 = Take Item to Player")
		fmt.Println("5 = Heal Guard (costs 1 health potion)")
		fmt.Println("0 = Back")
		fmt.Print("Choice: ")

		scanner.Scan()
		choice := scanner.Text()

		switch choice {
		case "1":
			equipGuardItemFromInventory(guard)
		case "2":
			unequipGuardItem(guard)
		case "3":
			giveItemToGuard(guard, player)
		case "4":
			takeItemFromGuard(guard, player)
		case "5":
			healGuard(guard, player)
		case "0":
			return
		default:
			fmt.Println("Invalid choice")
		}
	}
}

func equipGuardItemFromInventory(guard *models.Guard) {
	if len(guard.Inventory) == 0 {
		fmt.Println("\nGuard has no items in inventory!")
		return
	}

	fmt.Println("\nSelect item to equip:")
	for i, item := range guard.Inventory {
		fmt.Printf("%d. %s (CP:%d, Slot:%d)\n", i+1, item.Name, item.CP, item.Slot)
	}
	fmt.Print("Choice (0=cancel): ")

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	choice := scanner.Text()
	idx, err := strconv.Atoi(choice)

	if err != nil || idx < 0 || idx > len(guard.Inventory) {
		fmt.Println("Invalid choice!")
		return
	}
	if idx == 0 {
		return
	}

	item := guard.Inventory[idx-1]
	EquipGuardItem(item, &guard.EquipmentMap, &guard.Inventory)

	// Recalculate stats
	guard.StatsMod = CalculateItemMods(guard.EquipmentMap)
	guard.HitPoints = guard.HitpointsNatural + guard.StatsMod.HitPointMod
	if guard.HitpointsRemaining > guard.HitPoints {
		guard.HitpointsRemaining = guard.HitPoints
	}

	fmt.Printf("\nEquipped %s\n", item.Name)
}

func unequipGuardItem(guard *models.Guard) {
	if len(guard.EquipmentMap) == 0 {
		fmt.Println("\nGuard has no equipped items!")
		return
	}

	fmt.Println("\nSelect slot to unequip:")
	slotNames := map[int]string{
		0: "Head", 1: "Chest", 2: "Legs", 3: "Feet",
		4: "Hands", 5: "Main Hand", 6: "Off Hand", 7: "Accessory",
	}

	slots := []int{}
	for slot := range guard.EquipmentMap {
		slots = append(slots, slot)
	}

	for i, slot := range slots {
		item := guard.EquipmentMap[slot]
		slotName := slotNames[slot]
		if slotName == "" {
			slotName = fmt.Sprintf("Slot %d", slot)
		}
		fmt.Printf("%d. [%s] %s (CP:%d)\n", i+1, slotName, item.Name, item.CP)
	}
	fmt.Print("Choice (0=cancel): ")

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	choice := scanner.Text()
	idx, err := strconv.Atoi(choice)

	if err != nil || idx < 0 || idx > len(slots) {
		fmt.Println("Invalid choice!")
		return
	}
	if idx == 0 {
		return
	}

	slot := slots[idx-1]
	item := guard.EquipmentMap[slot]
	guard.Inventory = append(guard.Inventory, item)
	delete(guard.EquipmentMap, slot)

	// Recalculate stats
	guard.StatsMod = CalculateItemMods(guard.EquipmentMap)
	guard.HitPoints = guard.HitpointsNatural + guard.StatsMod.HitPointMod
	if guard.HitpointsRemaining > guard.HitPoints {
		guard.HitpointsRemaining = guard.HitPoints
	}

	fmt.Printf("\nUnequipped %s\n", item.Name)
}

func giveItemToGuard(guard *models.Guard, player *models.Character) {
	if len(player.Inventory) == 0 {
		fmt.Println("\nYou have no items to give!")
		return
	}

	// Filter only equipment items
	equipment := []models.Item{}
	equipmentIndices := []int{}
	for i, item := range player.Inventory {
		if item.ItemType == "equipment" || item.ItemType == "" {
			equipment = append(equipment, item)
			equipmentIndices = append(equipmentIndices, i)
		}
	}

	if len(equipment) == 0 {
		fmt.Println("\nYou have no equipment items to give!")
		return
	}

	fmt.Println("\nSelect item to give to guard:")
	for i, item := range equipment {
		fmt.Printf("%d. %s (CP:%d, Rarity:%d)\n", i+1, item.Name, item.CP, item.Rarity)
	}
	fmt.Print("Choice (0=cancel): ")

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	choice := scanner.Text()
	idx, err := strconv.Atoi(choice)

	if err != nil || idx < 0 || idx > len(equipment) {
		fmt.Println("Invalid choice!")
		return
	}
	if idx == 0 {
		return
	}

	item := equipment[idx-1]
	originalIdx := equipmentIndices[idx-1]

	// Give to guard
	guard.Inventory = append(guard.Inventory, item)

	// Remove from player
	RemoveItemFromInventory(&player.Inventory, originalIdx)

	fmt.Printf("\nGave %s to %s\n", item.Name, guard.Name)
}

func takeItemFromGuard(guard *models.Guard, player *models.Character) {
	if len(guard.Inventory) == 0 {
		fmt.Println("\nGuard has no items to take!")
		return
	}

	fmt.Println("\nSelect item to take from guard:")
	for i, item := range guard.Inventory {
		fmt.Printf("%d. %s (CP:%d, Rarity:%d)\n", i+1, item.Name, item.CP, item.Rarity)
	}
	fmt.Print("Choice (0=cancel): ")

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	choice := scanner.Text()
	idx, err := strconv.Atoi(choice)

	if err != nil || idx < 0 || idx > len(guard.Inventory) {
		fmt.Println("Invalid choice!")
		return
	}
	if idx == 0 {
		return
	}

	item := guard.Inventory[idx-1]

	// Give to player
	player.Inventory = append(player.Inventory, item)

	// Remove from guard
	RemoveItemFromInventory(&guard.Inventory, idx-1)

	fmt.Printf("\nTook %s from %s\n", item.Name, guard.Name)
}

func healGuard(guard *models.Guard, player *models.Character) {
	if !guard.Injured && guard.HitpointsRemaining >= guard.HitPoints {
		fmt.Println("\nGuard is already at full health!")
		return
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
		fmt.Println("\nYou don't have any health potions!")
		return
	}

	potion := player.Inventory[potionIdx]
	healAmount := potion.Consumable.Value

	// Heal guard
	guard.HitpointsRemaining += healAmount
	if guard.HitpointsRemaining > guard.HitPoints {
		guard.HitpointsRemaining = guard.HitPoints
	}

	// Clear injury status if fully healed
	if guard.HitpointsRemaining >= guard.HitPoints {
		guard.Injured = false
		guard.RecoveryTime = 0
	}

	// Remove potion from player
	RemoveItemFromInventory(&player.Inventory, potionIdx)

	fmt.Printf("\nUsed %s on %s\n", potion.Name, guard.Name)
	fmt.Printf("   %s healed for %d HP! (%d/%d)\n", guard.Name, healAmount, guard.HitpointsRemaining, guard.HitPoints)

	if !guard.Injured {
		fmt.Println("   Guard is now fully recovered!")
	}
}

func viewVillagers(village *models.Village) {
	fmt.Println("\n============================================================")
	fmt.Println("VILLAGERS")
	fmt.Println("============================================================")

	if len(village.Villagers) == 0 {
		fmt.Println("No villagers yet. Rescue them during hunts!")
		return
	}

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
		fmt.Println("\nHARVESTERS:")
		for i, v := range harvesters {
			taskInfo := "Idle"
			if v.HarvestType != "" {
				taskInfo = "Harvesting " + v.HarvestType + " (+" + fmt.Sprint(v.Efficiency+(v.Level/2)) + "/visit)"
			}
			fmt.Printf("  %d. %s (Lv%d) - %s\n", i+1, v.Name, v.Level, taskInfo)
		}
	}

	if len(guards) > 0 {
		fmt.Println("\nGUARDS:")
		for i, v := range guards {
			fmt.Printf("  %d. %s (Lv%d) - Efficiency: %d\n", i+1, v.Name, v.Level, v.Efficiency)
		}
	}

	fmt.Println("============================================================")
}

func assignVillagerTask(village *models.Village, player *models.Character) {
	harvesters := []int{}
	for i, v := range village.Villagers {
		if v.Role == "harvester" {
			harvesters = append(harvesters, i)
		}
	}

	if len(harvesters) == 0 {
		fmt.Println("\nNo harvesters available!")
		return
	}

	fmt.Println("\n============================================================")
	fmt.Println("ASSIGN HARVESTER TASK")
	fmt.Println("============================================================")
	fmt.Println("\nAvailable Harvesters:")
	for i, idx := range harvesters {
		v := village.Villagers[idx]
		taskInfo := "Idle"
		if v.HarvestType != "" {
			taskInfo = "Currently: " + v.HarvestType
		}
		fmt.Printf("%d. %s (Lv%d, Efficiency %d) - %s\n", i+1, v.Name, v.Level, v.Efficiency, taskInfo)
	}

	fmt.Print("\nSelect harvester (0=cancel): ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	choice := scanner.Text()
	idx, err := strconv.Atoi(choice)

	if err != nil || idx < 0 || idx > len(harvesters) {
		fmt.Println("Invalid choice!")
		return
	}
	if idx == 0 {
		return
	}

	villagerIdx := harvesters[idx-1]

	fmt.Println("\nAssign to resource:")
	for i, res := range data.ResourceTypes {
		fmt.Printf("%d = %s\n", i+1, res)
	}
	fmt.Print("Choice (0=cancel): ")
	scanner.Scan()
	resChoice := scanner.Text()
	resIdx, err := strconv.Atoi(resChoice)

	if err != nil || resIdx < 0 || resIdx > len(data.ResourceTypes) {
		fmt.Println("Invalid choice!")
		return
	}
	if resIdx == 0 {
		return
	}

	village.Villagers[villagerIdx].HarvestType = data.ResourceTypes[resIdx-1]
	village.Villagers[villagerIdx].AssignedTask = "harvesting"

	fmt.Printf("\n%s is now harvesting %s!\n",
		village.Villagers[villagerIdx].Name,
		data.ResourceTypes[resIdx-1])

	// Grant village XP for task assignment
	village.Experience += 10
	fmt.Println("+10 Village XP")
}

func hireGuardMenu(village *models.Village, player *models.Character) {
	fmt.Println("\n============================================================")
	fmt.Println("GUARD RECRUITMENT")
	fmt.Println("============================================================")

	// Get player's gold
	goldResource, hasGold := player.ResourceStorageMap["Gold"]
	if !hasGold {
		goldResource = models.Resource{Name: "Gold", Stock: 0, RollModifier: 0}
	}

	fmt.Printf("Your Gold: %d\n\n", goldResource.Stock)

	fmt.Println("Available Guards for Hire:")

	// Generate 3 guards at different levels
	availableGuards := []models.Guard{
		GenerateGuard(village.Level),
		GenerateGuard(village.Level + 2),
		GenerateGuard(village.Level + 5),
	}

	for i, guard := range availableGuards {
		fmt.Printf("\n%d. %s (Level %d)\n", i+1, guard.Name, guard.Level)
		fmt.Printf("   HP: %d | Attack Rolls: %d | Defense Rolls: %d\n",
			guard.HitPoints, guard.AttackRolls, guard.DefenseRolls)
		fmt.Printf("   Equipment Bonus: +%d ATK, +%d DEF, +%d HP\n",
			guard.StatsMod.AttackMod, guard.StatsMod.DefenseMod, guard.StatsMod.HitPointMod)
		fmt.Printf("   Starting Equipment: %d items\n", len(guard.EquipmentMap))
		fmt.Printf("   Cost: %d Gold\n", guard.Cost)
	}

	fmt.Print("\nHire guard (0=cancel): ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	choice := scanner.Text()
	idx, err := strconv.Atoi(choice)

	if err != nil || idx < 0 || idx > len(availableGuards) {
		fmt.Println("Invalid choice!")
		return
	}
	if idx == 0 {
		return
	}

	selectedGuard := availableGuards[idx-1]

	if goldResource.Stock < selectedGuard.Cost {
		fmt.Printf("\nNot enough gold! Need %d, have %d\n", selectedGuard.Cost, goldResource.Stock)
		return
	}

	// Deduct gold
	goldResource.Stock -= selectedGuard.Cost
	player.ResourceStorageMap["Gold"] = goldResource

	// Add guard to village
	selectedGuard.Hired = true
	village.ActiveGuards = append(village.ActiveGuards, selectedGuard)

	fmt.Printf("\nHired %s for %d Gold!\n", selectedGuard.Name, selectedGuard.Cost)
	fmt.Printf("They will assist in guardian and boss fights!\n")

	// Grant village XP
	village.Experience += 50
	fmt.Println("+50 Village XP")
}

func craftingMenu(village *models.Village, player *models.Character) {
	if len(village.UnlockedCrafting) == 0 {
		fmt.Println("\nNo crafting unlocked yet!")
		fmt.Println("Level up your village to unlock crafting:")
		fmt.Println("  Level 3  -> Potion Crafting")
		fmt.Println("  Level 5  -> Armor Crafting")
		fmt.Println("  Level 7  -> Weapon Crafting")
		fmt.Println("  Level 10 -> Skill Upgrades")
		fmt.Println("  Level 10 -> Skill Scroll Crafting")
		return
	}

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Println("\n============================================================")
		fmt.Println("CRAFTING MENU")
		fmt.Println("============================================================")
		fmt.Println("\nAvailable Crafting:")

		optionNum := 1
		optionMap := make(map[int]string)

		if Contains(village.UnlockedCrafting, "potions") {
			fmt.Printf("%d = Potion Crafting\n", optionNum)
			optionMap[optionNum] = "potions"
			optionNum++
		}
		if Contains(village.UnlockedCrafting, "armor") {
			fmt.Printf("%d = Armor Crafting\n", optionNum)
			optionMap[optionNum] = "armor"
			optionNum++
		}
		if Contains(village.UnlockedCrafting, "weapons") {
			fmt.Printf("%d = Weapon Crafting\n", optionNum)
			optionMap[optionNum] = "weapons"
			optionNum++
		}
		if Contains(village.UnlockedCrafting, "skill_upgrades") {
			fmt.Printf("%d = Skill Upgrades\n", optionNum)
			optionMap[optionNum] = "skill_upgrades"
			optionNum++
		}
		if Contains(village.UnlockedCrafting, "skill_scrolls") {
			fmt.Printf("%d = Skill Scroll Crafting\n", optionNum)
			optionMap[optionNum] = "skill_scrolls"
			optionNum++
		}

		fmt.Println("0 = Back")
		fmt.Print("Choice: ")

		scanner.Scan()
		choice := scanner.Text()
		idx, err := strconv.Atoi(choice)

		if err != nil || idx < 0 || idx >= optionNum {
			fmt.Println("Invalid choice!")
			continue
		}
		if idx == 0 {
			return
		}

		craftType := optionMap[idx]
		switch craftType {
		case "potions":
			craftPotion(village, player)
		case "armor":
			craftArmor(village, player)
		case "weapons":
			craftWeapon(village, player)
		case "skill_upgrades":
			upgradeSkillMenu(village, player)
		case "skill_scrolls":
			craftSkillScrolls(village, player)
		}
	}
}

func craftPotion(village *models.Village, player *models.Character) {
	fmt.Println("\n============================================================")
	fmt.Println("POTION CRAFTING")
	fmt.Println("============================================================")

	potionRecipes := []struct {
		name     string
		size     string
		ironCost int
		goldCost int
	}{
		{"Small Health Potion", "small", 5, 10},
		{"Medium Health Potion", "medium", 10, 20},
		{"Large Health Potion", "large", 20, 40},
	}

	fmt.Println("\nAvailable Recipes:")
	for i, recipe := range potionRecipes {
		fmt.Printf("%d. %s (Iron: %d, Gold: %d)\n",
			i+1, recipe.name, recipe.ironCost, recipe.goldCost)
	}

	fmt.Print("\nCraft (0=cancel): ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	choice := scanner.Text()
	idx, err := strconv.Atoi(choice)

	if err != nil || idx < 0 || idx > len(potionRecipes) {
		fmt.Println("Invalid choice!")
		return
	}
	if idx == 0 {
		return
	}

	recipe := potionRecipes[idx-1]

	// Check resources
	iron := player.ResourceStorageMap["Iron"]
	gold := player.ResourceStorageMap["Gold"]

	if iron.Stock < recipe.ironCost {
		fmt.Printf("Not enough Iron! Need %d, have %d\n", recipe.ironCost, iron.Stock)
		return
	}
	if gold.Stock < recipe.goldCost {
		fmt.Printf("Not enough Gold! Need %d, have %d\n", recipe.goldCost, gold.Stock)
		return
	}

	// Deduct resources
	iron.Stock -= recipe.ironCost
	gold.Stock -= recipe.goldCost
	player.ResourceStorageMap["Iron"] = iron
	player.ResourceStorageMap["Gold"] = gold

	// Create potion
	potion := CreateHealthPotion(recipe.size)
	player.Inventory = append(player.Inventory, potion)

	fmt.Printf("\nCrafted %s!\n", recipe.name)

	// Grant village XP
	village.Experience += 20
	fmt.Println("+20 Village XP")
}

func craftArmor(village *models.Village, player *models.Character) {
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Println("\n============================================================")
		fmt.Println("ARMOR CRAFTING")
		fmt.Println("============================================================")

		fmt.Println("\nAvailable Recipes:")
		fmt.Println("\n--- STANDARD ARMOR ---")
		fmt.Println("1. Enhanced Armor (Iron: 30, Stone: 20)")
		fmt.Println("   -> Random armor, Rarity 3-5, Defense focus")

		fmt.Println("\n--- BEAST MATERIAL ARMOR ---")
		fmt.Println("2. Beast Skin Armor (Iron: 20, Beast Skin: 15)")
		fmt.Println("   -> Light armor, Rarity 4-6, Fire resistance")

		fmt.Println("3. Bone Plate Armor (Iron: 25, Beast Bone: 12, Stone: 15)")
		fmt.Println("   -> Heavy armor, Rarity 5-7, High defense")

		fmt.Println("4. Tough Hide Vest (Tough Hide: 10, Beast Bone: 8)")
		fmt.Println("   -> Medium armor, Rarity 4-6, Physical resistance")

		fmt.Println("5. Ore Fragment Mail (Ore Fragment: 20, Iron: 15)")
		fmt.Println("   -> Magic armor, Rarity 5-7, Lightning resistance")

		fmt.Println("6. Fang-Studded Armor (Sharp Fang: 15, Beast Skin: 10, Iron: 20)")
		fmt.Println("   -> Spiked armor, Rarity 6-8, Counter-damage bonus")

		fmt.Println("7. Claw Guard Armor (Monster Claw: 12, Tough Hide: 8, Iron: 15)")
		fmt.Println("   -> Elite armor, Rarity 6-8, Ice resistance")

		fmt.Println("\n0 = Back")
		fmt.Print("Choice: ")

		scanner.Scan()
		choice := scanner.Text()

		switch choice {
		case "1":
			// Standard armor
			iron := player.ResourceStorageMap["Iron"]
			stone := player.ResourceStorageMap["Stone"]

			if iron.Stock < 30 {
				fmt.Printf("Not enough Iron! Need 30, have %d\n", iron.Stock)
				continue
			}
			if stone.Stock < 20 {
				fmt.Printf("Not enough Stone! Need 20, have %d\n", stone.Stock)
				continue
			}

			iron.Stock -= 30
			stone.Stock -= 20
			player.ResourceStorageMap["Iron"] = iron
			player.ResourceStorageMap["Stone"] = stone

			rarity := 3 + rand.Intn(3)
			armor := GenerateItem(rarity)
			armor.StatsMod.DefenseMod += rarity * 2
			armor.StatsMod.HitPointMod += rarity
			armor.CP = armor.StatsMod.AttackMod + armor.StatsMod.DefenseMod + armor.StatsMod.HitPointMod

			EquipBestItem(armor, &player.EquipmentMap, &player.Inventory)
			fmt.Printf("\nCrafted %s (Rarity %d)!\n", armor.Name, armor.Rarity)
			fmt.Printf("   Defense: +%d | HP: +%d | CP: %d\n",
				armor.StatsMod.DefenseMod, armor.StatsMod.HitPointMod, armor.CP)

			village.Experience += 40
			fmt.Println("+40 Village XP")

		case "2":
			// Beast Skin Armor
			iron := player.ResourceStorageMap["Iron"]
			beastSkin := player.ResourceStorageMap["Beast Skin"]

			if iron.Stock < 20 || beastSkin.Stock < 15 {
				fmt.Printf("Insufficient materials!\n")
				fmt.Printf("   Need: Iron 20 (have %d), Beast Skin 15 (have %d)\n", iron.Stock, beastSkin.Stock)
				continue
			}

			iron.Stock -= 20
			beastSkin.Stock -= 15
			player.ResourceStorageMap["Iron"] = iron
			player.ResourceStorageMap["Beast Skin"] = beastSkin

			rarity := 4 + rand.Intn(3)
			armor := GenerateItem(rarity)
			armor.StatsMod.DefenseMod += rarity * 2
			armor.StatsMod.HitPointMod += rarity + 3
			armor.CP = armor.StatsMod.AttackMod + armor.StatsMod.DefenseMod + armor.StatsMod.HitPointMod

			EquipBestItem(armor, &player.EquipmentMap, &player.Inventory)
			fmt.Printf("\nCrafted Beast Skin Armor (Rarity %d)!\n", rarity)
			fmt.Printf("   Defense: +%d | HP: +%d | Fire Resistance\n",
				armor.StatsMod.DefenseMod, armor.StatsMod.HitPointMod)

			village.Experience += 50
			fmt.Println("+50 Village XP")

		case "3":
			// Bone Plate Armor
			iron := player.ResourceStorageMap["Iron"]
			beastBone := player.ResourceStorageMap["Beast Bone"]
			stone := player.ResourceStorageMap["Stone"]

			if iron.Stock < 25 || beastBone.Stock < 12 || stone.Stock < 15 {
				fmt.Printf("Insufficient materials!\n")
				continue
			}

			iron.Stock -= 25
			beastBone.Stock -= 12
			stone.Stock -= 15
			player.ResourceStorageMap["Iron"] = iron
			player.ResourceStorageMap["Beast Bone"] = beastBone
			player.ResourceStorageMap["Stone"] = stone

			rarity := 5 + rand.Intn(3)
			armor := GenerateItem(rarity)
			armor.StatsMod.DefenseMod += rarity * 3
			armor.StatsMod.HitPointMod += rarity * 2
			armor.CP = armor.StatsMod.AttackMod + armor.StatsMod.DefenseMod + armor.StatsMod.HitPointMod

			EquipBestItem(armor, &player.EquipmentMap, &player.Inventory)
			fmt.Printf("\nCrafted Bone Plate Armor (Rarity %d)!\n", rarity)
			fmt.Printf("   Defense: +%d | HP: +%d | Heavy Armor\n",
				armor.StatsMod.DefenseMod, armor.StatsMod.HitPointMod)

			village.Experience += 60
			fmt.Println("+60 Village XP")

		case "4":
			// Tough Hide Vest
			toughHide := player.ResourceStorageMap["Tough Hide"]
			beastBone := player.ResourceStorageMap["Beast Bone"]

			if toughHide.Stock < 10 || beastBone.Stock < 8 {
				fmt.Printf("Insufficient materials!\n")
				continue
			}

			toughHide.Stock -= 10
			beastBone.Stock -= 8
			player.ResourceStorageMap["Tough Hide"] = toughHide
			player.ResourceStorageMap["Beast Bone"] = beastBone

			rarity := 4 + rand.Intn(3)
			armor := GenerateItem(rarity)
			armor.StatsMod.DefenseMod += rarity*2 + 2
			armor.StatsMod.HitPointMod += rarity + 4
			armor.CP = armor.StatsMod.AttackMod + armor.StatsMod.DefenseMod + armor.StatsMod.HitPointMod

			EquipBestItem(armor, &player.EquipmentMap, &player.Inventory)
			fmt.Printf("\nCrafted Tough Hide Vest (Rarity %d)!\n", rarity)
			fmt.Printf("   Defense: +%d | HP: +%d | Physical Resistance\n",
				armor.StatsMod.DefenseMod, armor.StatsMod.HitPointMod)

			village.Experience += 50
			fmt.Println("+50 Village XP")

		case "5":
			// Ore Fragment Mail
			oreFragment := player.ResourceStorageMap["Ore Fragment"]
			iron := player.ResourceStorageMap["Iron"]

			if oreFragment.Stock < 20 || iron.Stock < 15 {
				fmt.Printf("Insufficient materials!\n")
				continue
			}

			oreFragment.Stock -= 20
			iron.Stock -= 15
			player.ResourceStorageMap["Ore Fragment"] = oreFragment
			player.ResourceStorageMap["Iron"] = iron

			rarity := 5 + rand.Intn(3)
			armor := GenerateItem(rarity)
			armor.StatsMod.DefenseMod += rarity*2 + 3
			armor.StatsMod.HitPointMod += rarity
			armor.CP = armor.StatsMod.AttackMod + armor.StatsMod.DefenseMod + armor.StatsMod.HitPointMod

			EquipBestItem(armor, &player.EquipmentMap, &player.Inventory)
			fmt.Printf("\nCrafted Ore Fragment Mail (Rarity %d)!\n", rarity)
			fmt.Printf("   Defense: +%d | HP: +%d | Lightning Resistance\n",
				armor.StatsMod.DefenseMod, armor.StatsMod.HitPointMod)

			village.Experience += 60
			fmt.Println("+60 Village XP")

		case "6":
			// Fang-Studded Armor
			sharpFang := player.ResourceStorageMap["Sharp Fang"]
			beastSkin := player.ResourceStorageMap["Beast Skin"]
			iron := player.ResourceStorageMap["Iron"]

			if sharpFang.Stock < 15 || beastSkin.Stock < 10 || iron.Stock < 20 {
				fmt.Printf("Insufficient materials!\n")
				continue
			}

			sharpFang.Stock -= 15
			beastSkin.Stock -= 10
			iron.Stock -= 20
			player.ResourceStorageMap["Sharp Fang"] = sharpFang
			player.ResourceStorageMap["Beast Skin"] = beastSkin
			player.ResourceStorageMap["Iron"] = iron

			rarity := 6 + rand.Intn(3)
			armor := GenerateItem(rarity)
			armor.StatsMod.DefenseMod += rarity*2 + 5
			armor.StatsMod.HitPointMod += rarity * 2
			armor.StatsMod.AttackMod += rarity // Spike bonus
			armor.CP = armor.StatsMod.AttackMod + armor.StatsMod.DefenseMod + armor.StatsMod.HitPointMod

			EquipBestItem(armor, &player.EquipmentMap, &player.Inventory)
			fmt.Printf("\nCrafted Fang-Studded Armor (Rarity %d)!\n", rarity)
			fmt.Printf("   Defense: +%d | HP: +%d | Attack: +%d (Counter-damage)\n",
				armor.StatsMod.DefenseMod, armor.StatsMod.HitPointMod, armor.StatsMod.AttackMod)

			village.Experience += 70
			fmt.Println("+70 Village XP")

		case "7":
			// Claw Guard Armor
			monsterClaw := player.ResourceStorageMap["Monster Claw"]
			toughHide := player.ResourceStorageMap["Tough Hide"]
			iron := player.ResourceStorageMap["Iron"]

			if monsterClaw.Stock < 12 || toughHide.Stock < 8 || iron.Stock < 15 {
				fmt.Printf("Insufficient materials!\n")
				continue
			}

			monsterClaw.Stock -= 12
			toughHide.Stock -= 8
			iron.Stock -= 15
			player.ResourceStorageMap["Monster Claw"] = monsterClaw
			player.ResourceStorageMap["Tough Hide"] = toughHide
			player.ResourceStorageMap["Iron"] = iron

			rarity := 6 + rand.Intn(3)
			armor := GenerateItem(rarity)
			armor.StatsMod.DefenseMod += rarity*3 + 2
			armor.StatsMod.HitPointMod += rarity*2 + 3
			armor.CP = armor.StatsMod.AttackMod + armor.StatsMod.DefenseMod + armor.StatsMod.HitPointMod

			EquipBestItem(armor, &player.EquipmentMap, &player.Inventory)
			fmt.Printf("\nCrafted Claw Guard Armor (Rarity %d)!\n", rarity)
			fmt.Printf("   Defense: +%d | HP: +%d | Ice Resistance (Elite)\n",
				armor.StatsMod.DefenseMod, armor.StatsMod.HitPointMod)

			village.Experience += 70
			fmt.Println("+70 Village XP")

		case "0":
			return

		default:
			fmt.Println("Invalid choice")
		}
	}
}

func craftWeapon(village *models.Village, player *models.Character) {
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Println("\n============================================================")
		fmt.Println("WEAPON CRAFTING")
		fmt.Println("============================================================")

		fmt.Println("\nAvailable Recipes:")
		fmt.Println("\n--- STANDARD WEAPONS ---")
		fmt.Println("1. Enhanced Weapon (Iron: 40, Gold: 30)")
		fmt.Println("   -> Random weapon, Rarity 4-6, Attack focus")

		fmt.Println("\n--- BEAST MATERIAL WEAPONS ---")
		fmt.Println("2. Beast Claw Blade (Iron: 25, Monster Claw: 15, Sharp Fang: 10)")
		fmt.Println("   -> Slashing weapon, Rarity 5-7, High attack + bleed bonus")

		fmt.Println("3. Bone Crusher Mace (Iron: 30, Beast Bone: 20, Stone: 15)")
		fmt.Println("   -> Crushing weapon, Rarity 5-7, Attack + stun chance")

		fmt.Println("4. Hide-Wrapped Axe (Iron: 20, Tough Hide: 12, Lumber: 25)")
		fmt.Println("   -> Balanced weapon, Rarity 4-6, Attack + HP bonus")

		fmt.Println("5. Ore Fragment Sword (Iron: 35, Ore Fragment: 25, Gold: 20)")
		fmt.Println("   -> Magical weapon, Rarity 6-8, High attack + elemental damage")

		fmt.Println("6. Fang Spear (Sharp Fang: 18, Beast Bone: 15, Iron: 20)")
		fmt.Println("   -> Piercing weapon, Rarity 5-7, Attack + critical bonus")

		fmt.Println("7. Composite War Hammer (Beast Skin: 10, Ore Fragment: 15, Iron: 25, Stone: 20)")
		fmt.Println("   -> Elite weapon, Rarity 6-8, Massive attack + durability")

		fmt.Println("\n0 = Back")
		fmt.Print("Choice: ")

		scanner.Scan()
		choice := scanner.Text()

		switch choice {
		case "1":
			// Standard Enhanced Weapon
			iron := player.ResourceStorageMap["Iron"]
			gold := player.ResourceStorageMap["Gold"]

			if iron.Stock < 40 {
				fmt.Printf("Not enough Iron! Need 40, have %d\n", iron.Stock)
				continue
			}
			if gold.Stock < 30 {
				fmt.Printf("Not enough Gold! Need 30, have %d\n", gold.Stock)
				continue
			}

			iron.Stock -= 40
			gold.Stock -= 30
			player.ResourceStorageMap["Iron"] = iron
			player.ResourceStorageMap["Gold"] = gold

			rarity := 4 + rand.Intn(3)
			weapon := GenerateItem(rarity)
			weapon.StatsMod.AttackMod += rarity * 3
			weapon.StatsMod.HitPointMod += rarity / 2
			weapon.CP = weapon.StatsMod.AttackMod + weapon.StatsMod.DefenseMod + weapon.StatsMod.HitPointMod

			EquipBestItem(weapon, &player.EquipmentMap, &player.Inventory)
			fmt.Printf("\nCrafted Enhanced %s (Rarity %d)!\n", weapon.Name, rarity)
			fmt.Printf("   Attack: +%d | HP: +%d | CP: %d\n",
				weapon.StatsMod.AttackMod, weapon.StatsMod.HitPointMod, weapon.CP)

			village.Experience += 50
			fmt.Println("+50 Village XP")

		case "2":
			// Beast Claw Blade
			iron := player.ResourceStorageMap["Iron"]
			monsterClaw := player.ResourceStorageMap["Monster Claw"]
			sharpFang := player.ResourceStorageMap["Sharp Fang"]

			if iron.Stock < 25 || monsterClaw.Stock < 15 || sharpFang.Stock < 10 {
				fmt.Printf("Insufficient materials!\n")
				fmt.Printf("   Need: Iron 25 (have %d), Monster Claw 15 (have %d), Sharp Fang 10 (have %d)\n",
					iron.Stock, monsterClaw.Stock, sharpFang.Stock)
				continue
			}

			iron.Stock -= 25
			monsterClaw.Stock -= 15
			sharpFang.Stock -= 10
			player.ResourceStorageMap["Iron"] = iron
			player.ResourceStorageMap["Monster Claw"] = monsterClaw
			player.ResourceStorageMap["Sharp Fang"] = sharpFang

			rarity := 5 + rand.Intn(3)
			weapon := GenerateItem(rarity)
			weapon.StatsMod.AttackMod += rarity * 4
			weapon.StatsMod.HitPointMod += rarity
			weapon.CP = weapon.StatsMod.AttackMod + weapon.StatsMod.DefenseMod + weapon.StatsMod.HitPointMod

			EquipBestItem(weapon, &player.EquipmentMap, &player.Inventory)
			fmt.Printf("\nCrafted Beast Claw Blade (Rarity %d)!\n", rarity)
			fmt.Printf("   Attack: +%d | HP: +%d | Bleed Bonus\n",
				weapon.StatsMod.AttackMod, weapon.StatsMod.HitPointMod)

			village.Experience += 60
			fmt.Println("+60 Village XP")

		case "3":
			// Bone Crusher Mace
			iron := player.ResourceStorageMap["Iron"]
			beastBone := player.ResourceStorageMap["Beast Bone"]
			stone := player.ResourceStorageMap["Stone"]

			if iron.Stock < 30 || beastBone.Stock < 20 || stone.Stock < 15 {
				fmt.Printf("Insufficient materials!\n")
				continue
			}

			iron.Stock -= 30
			beastBone.Stock -= 20
			stone.Stock -= 15
			player.ResourceStorageMap["Iron"] = iron
			player.ResourceStorageMap["Beast Bone"] = beastBone
			player.ResourceStorageMap["Stone"] = stone

			rarity := 5 + rand.Intn(3)
			weapon := GenerateItem(rarity)
			weapon.StatsMod.AttackMod += rarity*3 + 5
			weapon.StatsMod.DefenseMod += rarity
			weapon.StatsMod.HitPointMod += rarity + 2
			weapon.CP = weapon.StatsMod.AttackMod + weapon.StatsMod.DefenseMod + weapon.StatsMod.HitPointMod

			EquipBestItem(weapon, &player.EquipmentMap, &player.Inventory)
			fmt.Printf("\nCrafted Bone Crusher Mace (Rarity %d)!\n", rarity)
			fmt.Printf("   Attack: +%d | Defense: +%d | HP: +%d | Stun Chance\n",
				weapon.StatsMod.AttackMod, weapon.StatsMod.DefenseMod, weapon.StatsMod.HitPointMod)

			village.Experience += 60
			fmt.Println("+60 Village XP")

		case "4":
			// Hide-Wrapped Axe
			iron := player.ResourceStorageMap["Iron"]
			toughHide := player.ResourceStorageMap["Tough Hide"]
			lumber := player.ResourceStorageMap["Lumber"]

			if iron.Stock < 20 || toughHide.Stock < 12 || lumber.Stock < 25 {
				fmt.Printf("Insufficient materials!\n")
				continue
			}

			iron.Stock -= 20
			toughHide.Stock -= 12
			lumber.Stock -= 25
			player.ResourceStorageMap["Iron"] = iron
			player.ResourceStorageMap["Tough Hide"] = toughHide
			player.ResourceStorageMap["Lumber"] = lumber

			rarity := 4 + rand.Intn(3)
			weapon := GenerateItem(rarity)
			weapon.StatsMod.AttackMod += rarity*3 + 3
			weapon.StatsMod.HitPointMod += rarity*2 + 5
			weapon.CP = weapon.StatsMod.AttackMod + weapon.StatsMod.DefenseMod + weapon.StatsMod.HitPointMod

			EquipBestItem(weapon, &player.EquipmentMap, &player.Inventory)
			fmt.Printf("\nCrafted Hide-Wrapped Axe (Rarity %d)!\n", rarity)
			fmt.Printf("   Attack: +%d | HP: +%d | Balanced\n",
				weapon.StatsMod.AttackMod, weapon.StatsMod.HitPointMod)

			village.Experience += 55
			fmt.Println("+55 Village XP")

		case "5":
			// Ore Fragment Sword
			iron := player.ResourceStorageMap["Iron"]
			oreFragment := player.ResourceStorageMap["Ore Fragment"]
			gold := player.ResourceStorageMap["Gold"]

			if iron.Stock < 35 || oreFragment.Stock < 25 || gold.Stock < 20 {
				fmt.Printf("Insufficient materials!\n")
				continue
			}

			iron.Stock -= 35
			oreFragment.Stock -= 25
			gold.Stock -= 20
			player.ResourceStorageMap["Iron"] = iron
			player.ResourceStorageMap["Ore Fragment"] = oreFragment
			player.ResourceStorageMap["Gold"] = gold

			rarity := 6 + rand.Intn(3)
			weapon := GenerateItem(rarity)
			weapon.StatsMod.AttackMod += rarity*4 + 5
			weapon.StatsMod.HitPointMod += rarity + 3
			weapon.CP = weapon.StatsMod.AttackMod + weapon.StatsMod.DefenseMod + weapon.StatsMod.HitPointMod

			EquipBestItem(weapon, &player.EquipmentMap, &player.Inventory)
			fmt.Printf("\nCrafted Ore Fragment Sword (Rarity %d)!\n", rarity)
			fmt.Printf("   Attack: +%d | HP: +%d | Elemental Damage (Magical)\n",
				weapon.StatsMod.AttackMod, weapon.StatsMod.HitPointMod)

			village.Experience += 70
			fmt.Println("+70 Village XP")

		case "6":
			// Fang Spear
			sharpFang := player.ResourceStorageMap["Sharp Fang"]
			beastBone := player.ResourceStorageMap["Beast Bone"]
			iron := player.ResourceStorageMap["Iron"]

			if sharpFang.Stock < 18 || beastBone.Stock < 15 || iron.Stock < 20 {
				fmt.Printf("Insufficient materials!\n")
				continue
			}

			sharpFang.Stock -= 18
			beastBone.Stock -= 15
			iron.Stock -= 20
			player.ResourceStorageMap["Sharp Fang"] = sharpFang
			player.ResourceStorageMap["Beast Bone"] = beastBone
			player.ResourceStorageMap["Iron"] = iron

			rarity := 5 + rand.Intn(3)
			weapon := GenerateItem(rarity)
			weapon.StatsMod.AttackMod += rarity*3 + 7
			weapon.StatsMod.HitPointMod += rarity
			weapon.CP = weapon.StatsMod.AttackMod + weapon.StatsMod.DefenseMod + weapon.StatsMod.HitPointMod

			EquipBestItem(weapon, &player.EquipmentMap, &player.Inventory)
			fmt.Printf("\nCrafted Fang Spear (Rarity %d)!\n", rarity)
			fmt.Printf("   Attack: +%d | HP: +%d | Critical Bonus (Piercing)\n",
				weapon.StatsMod.AttackMod, weapon.StatsMod.HitPointMod)

			village.Experience += 65
			fmt.Println("+65 Village XP")

		case "7":
			// Composite War Hammer
			beastSkin := player.ResourceStorageMap["Beast Skin"]
			oreFragment := player.ResourceStorageMap["Ore Fragment"]
			iron := player.ResourceStorageMap["Iron"]
			stone := player.ResourceStorageMap["Stone"]

			if beastSkin.Stock < 10 || oreFragment.Stock < 15 || iron.Stock < 25 || stone.Stock < 20 {
				fmt.Printf("Insufficient materials!\n")
				continue
			}

			beastSkin.Stock -= 10
			oreFragment.Stock -= 15
			iron.Stock -= 25
			stone.Stock -= 20
			player.ResourceStorageMap["Beast Skin"] = beastSkin
			player.ResourceStorageMap["Ore Fragment"] = oreFragment
			player.ResourceStorageMap["Iron"] = iron
			player.ResourceStorageMap["Stone"] = stone

			rarity := 6 + rand.Intn(3)
			weapon := GenerateItem(rarity)
			weapon.StatsMod.AttackMod += rarity*5 + 3
			weapon.StatsMod.DefenseMod += rarity + 2
			weapon.StatsMod.HitPointMod += rarity*2 + 5
			weapon.CP = weapon.StatsMod.AttackMod + weapon.StatsMod.DefenseMod + weapon.StatsMod.HitPointMod

			EquipBestItem(weapon, &player.EquipmentMap, &player.Inventory)
			fmt.Printf("\nCrafted Composite War Hammer (Rarity %d)!\n", rarity)
			fmt.Printf("   Attack: +%d | Defense: +%d | HP: +%d | Elite (Durability)\n",
				weapon.StatsMod.AttackMod, weapon.StatsMod.DefenseMod, weapon.StatsMod.HitPointMod)

			village.Experience += 80
			fmt.Println("+80 Village XP")

		case "0":
			return

		default:
			fmt.Println("Invalid choice")
		}
	}
}

func upgradeSkillMenu(village *models.Village, player *models.Character) {
	fmt.Println("\n============================================================")
	fmt.Println("SKILL UPGRADES")
	fmt.Println("============================================================")

	if len(player.LearnedSkills) == 0 {
		fmt.Println("\nNo skills to upgrade!")
		return
	}

	fmt.Println("\nYour Skills:")
	for i, skill := range player.LearnedSkills {
		fmt.Printf("%d. %s (Damage: %d, ManaCost: %d, StaminaCost: %d)\n",
			i+1, skill.Name, skill.Damage, skill.ManaCost, skill.StaminaCost)
	}

	fmt.Print("\nUpgrade skill (0=cancel): ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	choice := scanner.Text()
	idx, err := strconv.Atoi(choice)

	if err != nil || idx < 0 || idx > len(player.LearnedSkills) {
		fmt.Println("Invalid choice!")
		return
	}
	if idx == 0 {
		return
	}

	skillIdx := idx - 1
	skill := &player.LearnedSkills[skillIdx]

	fmt.Printf("\nUpgrade %s\n", skill.Name)
	fmt.Println("Cost: Gold: 50, Iron: 25")
	fmt.Println("Effect: +5 Damage (or +5 Healing), -2 Resource Cost")

	fmt.Print("\nUpgrade? (y/n): ")
	scanner.Scan()
	confirm := scanner.Text()

	if confirm != "y" && confirm != "Y" {
		return
	}

	// Check resources
	iron := player.ResourceStorageMap["Iron"]
	gold := player.ResourceStorageMap["Gold"]

	if iron.Stock < 25 {
		fmt.Printf("Not enough Iron! Need 25, have %d\n", iron.Stock)
		return
	}
	if gold.Stock < 50 {
		fmt.Printf("Not enough Gold! Need 50, have %d\n", gold.Stock)
		return
	}

	// Deduct resources
	iron.Stock -= 25
	gold.Stock -= 50
	player.ResourceStorageMap["Iron"] = iron
	player.ResourceStorageMap["Gold"] = gold

	// Upgrade skill
	if skill.Damage > 0 {
		skill.Damage += 5
		fmt.Printf("%s damage increased by 5! (Now: %d)\n", skill.Name, skill.Damage)
	} else if skill.Damage < 0 {
		skill.Damage -= 5
		fmt.Printf("%s healing increased by 5! (Now: %d)\n", skill.Name, -skill.Damage)
	}

	if skill.ManaCost > 2 {
		skill.ManaCost -= 2
		fmt.Printf("Mana cost reduced by 2! (Now: %d)\n", skill.ManaCost)
	}
	if skill.StaminaCost > 2 {
		skill.StaminaCost -= 2
		fmt.Printf("Stamina cost reduced by 2! (Now: %d)\n", skill.StaminaCost)
	}

	// Grant village XP
	village.Experience += 60
	fmt.Println("+60 Village XP")
}

func craftSkillScrolls(village *models.Village, player *models.Character) {
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Println("\n============================================================")
		fmt.Println("SKILL SCROLL CRAFTING")
		fmt.Println("============================================================")

		fmt.Println("\nCraft skill scrolls using beast materials!")
		fmt.Println("Learn skills without defeating Skill Guardians")

		fmt.Println("Available Skill Scrolls:")
		fmt.Println("\n--- OFFENSIVE SKILLS ---")
		fmt.Println("1. Fireball Scroll (Ore Fragment: 15, Sharp Fang: 10, Gold: 30)")
		fmt.Println("   -> Fire damage + burn effect")

		fmt.Println("2. Ice Shard Scroll (Beast Skin: 12, Ore Fragment: 10, Iron: 20)")
		fmt.Println("   -> Ice damage")

		fmt.Println("3. Lightning Bolt Scroll (Ore Fragment: 20, Monster Claw: 15, Gold: 40)")
		fmt.Println("   -> Lightning damage + stun (EXPENSIVE)")

		fmt.Println("4. Power Strike Scroll (Beast Bone: 10, Iron: 15)")
		fmt.Println("   -> Physical stamina attack")

		fmt.Println("5. Poison Blade Scroll (Beast Skin: 10, Sharp Fang: 12, Iron: 15)")
		fmt.Println("   -> Poison damage over time")

		fmt.Println("\n--- SUPPORT SKILLS ---")
		fmt.Println("6. Heal Scroll (Beast Skin: 15, Beast Bone: 10, Gold: 25)")
		fmt.Println("   -> Restore HP")

		fmt.Println("7. Regeneration Scroll (Ore Fragment: 12, Beast Skin: 15, Gold: 30)")
		fmt.Println("   -> Heal over time")

		fmt.Println("8. Shield Wall Scroll (Tough Hide: 15, Beast Bone: 12, Stone: 20)")
		fmt.Println("   -> Defense buff")

		fmt.Println("9. Battle Cry Scroll (Sharp Fang: 15, Beast Bone: 10, Iron: 20)")
		fmt.Println("   -> Attack buff")

		fmt.Println("\n--- UTILITY SKILLS ---")
		fmt.Println("10. Tracking Scroll (Beast Bone: 8, Beast Skin: 8)")
		fmt.Println("    -> Choose targets in combat")

		fmt.Println("\n0 = Back")
		fmt.Print("Choice: ")

		scanner.Scan()
		choice := scanner.Text()

		var skillToLearn models.Skill
		var materialsNeeded map[string]int
		var villageXP int

		switch choice {
		case "1":
			// Fireball
			skillToLearn = data.AvailableSkills[0]
			materialsNeeded = map[string]int{
				"Ore Fragment": 15,
				"Sharp Fang":   10,
				"Gold":         30,
			}
			villageXP = 100

		case "2":
			// Ice Shard
			skillToLearn = data.AvailableSkills[1]
			materialsNeeded = map[string]int{
				"Beast Skin":   12,
				"Ore Fragment": 10,
				"Iron":         20,
			}
			villageXP = 90

		case "3":
			// Lightning Bolt
			skillToLearn = data.AvailableSkills[2]
			materialsNeeded = map[string]int{
				"Ore Fragment": 20,
				"Monster Claw": 15,
				"Gold":         40,
			}
			villageXP = 120

		case "4":
			// Power Strike
			skillToLearn = data.AvailableSkills[4]
			materialsNeeded = map[string]int{
				"Beast Bone": 10,
				"Iron":       15,
			}
			villageXP = 70

		case "5":
			// Poison Blade
			skillToLearn = data.AvailableSkills[7]
			materialsNeeded = map[string]int{
				"Beast Skin": 10,
				"Sharp Fang": 12,
				"Iron":       15,
			}
			villageXP = 85

		case "6":
			// Heal
			skillToLearn = data.AvailableSkills[3]
			materialsNeeded = map[string]int{
				"Beast Skin": 15,
				"Beast Bone": 10,
				"Gold":       25,
			}
			villageXP = 95

		case "7":
			// Regeneration
			skillToLearn = data.AvailableSkills[8]
			materialsNeeded = map[string]int{
				"Ore Fragment": 12,
				"Beast Skin":   15,
				"Gold":         30,
			}
			villageXP = 105

		case "8":
			// Shield Wall
			skillToLearn = data.AvailableSkills[5]
			materialsNeeded = map[string]int{
				"Tough Hide": 15,
				"Beast Bone": 12,
				"Stone":      20,
			}
			villageXP = 90

		case "9":
			// Battle Cry
			skillToLearn = data.AvailableSkills[6]
			materialsNeeded = map[string]int{
				"Sharp Fang": 15,
				"Beast Bone": 10,
				"Iron":       20,
			}
			villageXP = 85

		case "10":
			// Tracking
			skillToLearn = data.AvailableSkills[9]
			materialsNeeded = map[string]int{
				"Beast Bone": 8,
				"Beast Skin": 8,
			}
			villageXP = 60

		case "0":
			return

		default:
			fmt.Println("Invalid choice")
			continue
		}

		// Check if player already has this skill
		hasSkill := false
		for _, skill := range player.LearnedSkills {
			if skill.Name == skillToLearn.Name {
				hasSkill = true
				break
			}
		}

		if hasSkill {
			fmt.Printf("\nYou already know %s!\n", skillToLearn.Name)
			continue
		}

		// Check if player has all required materials
		canCraft := true
		for material, qty := range materialsNeeded {
			resource, exists := player.ResourceStorageMap[material]
			if !exists || resource.Stock < qty {
				canCraft = false
				have := 0
				if exists {
					have = resource.Stock
				}
				fmt.Printf("Not enough %s! Need %d, have %d\n", material, qty, have)
			}
		}

		if !canCraft {
			continue
		}

		// Deduct materials
		for material, qty := range materialsNeeded {
			resource := player.ResourceStorageMap[material]
			resource.Stock -= qty
			player.ResourceStorageMap[material] = resource
		}

		// Learn the skill
		player.LearnedSkills = append(player.LearnedSkills, skillToLearn)

		fmt.Printf("\nCrafted %s Scroll!\n", skillToLearn.Name)
		fmt.Printf("You have learned %s!\n", skillToLearn.Name)
		fmt.Printf("   %s\n", skillToLearn.Description)

		// Grant village XP
		village.Experience += villageXP
		fmt.Printf("+%d Village XP\n", villageXP)
	}
}

func buildDefenseMenu(village *models.Village, player *models.Character) {
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Println("\n============================================================")
		fmt.Println("BUILD DEFENSES & TRAPS")
		fmt.Println("============================================================")
		fmt.Println("\n1 = Build Walls/Towers")
		fmt.Println("2 = Craft Traps")
		fmt.Println("3 = View Current Defenses")
		fmt.Println("0 = Back")
		fmt.Print("Choice: ")

		scanner.Scan()
		choice := scanner.Text()

		switch choice {
		case "1":
			buildWallsMenu(village, player)
		case "2":
			craftTrapsMenu(village, player)
		case "3":
			viewDefenses(village)
		case "0":
			return
		default:
			fmt.Println("Invalid choice")
		}
	}
}

func buildWallsMenu(village *models.Village, player *models.Character) {
	fmt.Println("\n============================================================")
	fmt.Println("BUILD WALLS & TOWERS")
	fmt.Println("============================================================")

	defenseOptions := []struct {
		name    string
		lumber  int
		stone   int
		iron    int
		defense int
		attack  int
		dtype   string
	}{
		{"Wooden Wall", 50, 20, 0, 10, 0, "wall"},
		{"Stone Wall", 30, 60, 10, 25, 0, "wall"},
		{"Iron Wall", 20, 80, 40, 40, 0, "wall"},
		{"Guard Tower", 40, 40, 30, 15, 20, "tower"},
		{"Arrow Tower", 30, 50, 40, 10, 35, "tower"},
		{"Iron Gate", 20, 50, 50, 30, 10, "wall"},
	}

	fmt.Println("\nAvailable Structures:")
	for i, def := range defenseOptions {
		fmt.Printf("%d. %s (Lumber:%d Stone:%d Iron:%d) -> Defense:+%d Attack:+%d\n",
			i+1, def.name, def.lumber, def.stone, def.iron, def.defense, def.attack)
	}

	fmt.Print("\nBuild (0=cancel): ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	choice := scanner.Text()
	idx, err := strconv.Atoi(choice)

	if err != nil || idx < 0 || idx > len(defenseOptions) {
		fmt.Println("Invalid choice!")
		return
	}
	if idx == 0 {
		return
	}

	selected := defenseOptions[idx-1]

	// Check resources
	lumber := player.ResourceStorageMap["Lumber"]
	stone := player.ResourceStorageMap["Stone"]
	iron := player.ResourceStorageMap["Iron"]

	if lumber.Stock < selected.lumber {
		fmt.Printf("Not enough Lumber! Need %d, have %d\n", selected.lumber, lumber.Stock)
		return
	}
	if stone.Stock < selected.stone {
		fmt.Printf("Not enough Stone! Need %d, have %d\n", selected.stone, stone.Stock)
		return
	}
	if iron.Stock < selected.iron {
		fmt.Printf("Not enough Iron! Need %d, have %d\n", selected.iron, iron.Stock)
		return
	}

	// Deduct resources
	lumber.Stock -= selected.lumber
	stone.Stock -= selected.stone
	iron.Stock -= selected.iron
	player.ResourceStorageMap["Lumber"] = lumber
	player.ResourceStorageMap["Stone"] = stone
	player.ResourceStorageMap["Iron"] = iron

	// Add defense
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
	village.DefenseLevel += 1

	fmt.Printf("\nBuilt %s!\n", selected.name)
	fmt.Printf("Village Defense Level increased to %d\n", village.DefenseLevel)

	// Grant village XP
	village.Experience += 30
	fmt.Println("+30 Village XP")
}

func craftTrapsMenu(village *models.Village, player *models.Character) {
	fmt.Println("\n============================================================")
	fmt.Println("CRAFT TRAPS")
	fmt.Println("============================================================")

	trapOptions := []struct {
		name        string
		trapType    string
		materials   map[string]int
		damage      int
		duration    int
		triggerRate int
	}{
		{
			name:     "Spike Trap",
			trapType: "spike",
			materials: map[string]int{
				"Iron":       10,
				"Beast Bone": 5,
			},
			damage:      15,
			duration:    3,
			triggerRate: 60,
		},
		{
			name:     "Fire Trap",
			trapType: "fire",
			materials: map[string]int{
				"Iron":         15,
				"Ore Fragment": 8,
				"Sharp Fang":   5,
			},
			damage:      25,
			duration:    2,
			triggerRate: 50,
		},
		{
			name:     "Ice Trap",
			trapType: "ice",
			materials: map[string]int{
				"Iron":         12,
				"Ore Fragment": 10,
				"Beast Skin":   8,
			},
			damage:      20,
			duration:    3,
			triggerRate: 55,
		},
		{
			name:     "Poison Trap",
			trapType: "poison",
			materials: map[string]int{
				"Beast Skin":   10,
				"Sharp Fang":   8,
				"Monster Claw": 5,
			},
			damage:      18,
			duration:    4,
			triggerRate: 65,
		},
		{
			name:     "Barricade Trap",
			trapType: "spike",
			materials: map[string]int{
				"Lumber":     30,
				"Tough Hide": 6,
				"Beast Bone": 8,
			},
			damage:      30,
			duration:    2,
			triggerRate: 70,
		},
	}

	fmt.Println("\nAvailable Traps:")
	for i, trap := range trapOptions {
		fmt.Printf("\n%d. %s (Damage: %d, Duration: %d waves, Trigger: %d%%)\n",
			i+1, trap.name, trap.damage, trap.duration, trap.triggerRate)
		fmt.Print("   Costs: ")
		first := true
		for mat, qty := range trap.materials {
			if !first {
				fmt.Print(", ")
			}
			fmt.Printf("%s:%d", mat, qty)
			first = false
		}
		fmt.Println()
	}

	fmt.Print("\nCraft (0=cancel): ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	choice := scanner.Text()
	idx, err := strconv.Atoi(choice)

	if err != nil || idx < 0 || idx > len(trapOptions) {
		fmt.Println("Invalid choice!")
		return
	}
	if idx == 0 {
		return
	}

	selected := trapOptions[idx-1]

	// Check if player has all required materials
	for material, qty := range selected.materials {
		resource, exists := player.ResourceStorageMap[material]
		if !exists || resource.Stock < qty {
			have := 0
			if exists {
				have = resource.Stock
			}
			fmt.Printf("Not enough %s! Need %d, have %d\n", material, qty, have)
			return
		}
	}

	// Deduct materials
	for material, qty := range selected.materials {
		resource := player.ResourceStorageMap[material]
		resource.Stock -= qty
		player.ResourceStorageMap[material] = resource
	}

	// Create trap
	newTrap := models.Trap{
		Name:        selected.name,
		Type:        selected.trapType,
		Damage:      selected.damage,
		Duration:    selected.duration,
		Remaining:   selected.duration,
		TriggerRate: selected.triggerRate,
	}
	village.Traps = append(village.Traps, newTrap)

	fmt.Printf("\nCrafted %s!\n", selected.name)
	fmt.Printf("Will last for %d monster tides\n", selected.duration)

	// Grant village XP
	village.Experience += 35
	fmt.Println("+35 Village XP")
}

func viewDefenses(village *models.Village) {
	fmt.Println("\n============================================================")
	fmt.Println("CURRENT DEFENSES")
	fmt.Println("============================================================")

	// Show walls and towers
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
		fmt.Println("\nWALLS:")
		for _, wall := range walls {
			fmt.Printf("  - %s (Defense: +%d)\n", wall.Name, wall.Defense)
		}
	}

	if len(towers) > 0 {
		fmt.Println("\nTOWERS:")
		for _, tower := range towers {
			fmt.Printf("  - %s (Defense: +%d, Attack: +%d)\n",
				tower.Name, tower.Defense, tower.AttackPower)
		}
	}

	// Show traps
	if len(village.Traps) > 0 {
		fmt.Println("\nACTIVE TRAPS:")
		for i, trap := range village.Traps {
			fmt.Printf("  %d. %s (Damage: %d, Waves left: %d/%d, Trigger: %d%%)\n",
				i+1, trap.Name, trap.Damage, trap.Remaining, trap.Duration, trap.TriggerRate)
		}
	}

	if len(village.Defenses) == 0 && len(village.Traps) == 0 {
		fmt.Println("\nNo defenses built yet!")
	}

	fmt.Printf("\nTotal Defense Level: %d\n", village.DefenseLevel)
	fmt.Println("============================================================")
}

func checkMonsterTide(village *models.Village) {
	fmt.Println("\n============================================================")
	fmt.Println("MONSTER TIDE STATUS")
	fmt.Println("============================================================")

	currentTime := time.Now().Unix()
	timeSinceLastTide := currentTime - village.LastTideTime
	timeUntilNext := village.TideInterval - int(timeSinceLastTide)

	if timeUntilNext <= 0 {
		fmt.Println("\nMONSTER TIDE IS READY!")
		fmt.Println("\nA wave of monsters can attack your village!")
		fmt.Printf("Village Defense Level: %d\n", village.DefenseLevel)
		fmt.Printf("Defenses: %d built\n", len(village.Defenses))
		fmt.Printf("Active Traps: %d\n", len(village.Traps))
		fmt.Printf("Guards: %d villagers + %d hired\n",
			CountVillagersByRole(village, "guard"),
			len(village.ActiveGuards))
		fmt.Println("\nNote: You can trigger the tide defense from this menu")
		fmt.Println("or it will happen automatically when you check again.")
	} else {
		hours := timeUntilNext / 3600
		minutes := (timeUntilNext % 3600) / 60
		fmt.Printf("\nNext Monster Tide in: %d hours, %d minutes\n", hours, minutes)
		fmt.Printf("Village Defense Level: %d\n", village.DefenseLevel)
		fmt.Printf("Defenses: %d built\n", len(village.Defenses))
		fmt.Printf("Active Traps: %d\n", len(village.Traps))
		fmt.Printf("Guards: %d villagers + %d hired\n",
			CountVillagersByRole(village, "guard"),
			len(village.ActiveGuards))
	}

	fmt.Println("\nPrepare your defenses by:")
	fmt.Println("  - Building more defenses")
	fmt.Println("  - Crafting traps")
	fmt.Println("  - Hiring guards")
	fmt.Println("  - Rescuing guard villagers during hunts")
	fmt.Println("============================================================")
}

// MonsterTideDefense runs the active wave-based monster tide defense event.
func MonsterTideDefense(gameState *models.GameState, player *models.Character, village *models.Village) {
	fmt.Println("\n============================================================")
	fmt.Println("MONSTER TIDE DEFENSE")
	fmt.Println("============================================================")

	// Calculate tide difficulty based on village level
	numWaves := 3 + (village.Level / 5) // 3-5 waves typically
	baseMonsterLevel := village.Level
	monstersPerWave := 5 + (village.Level / 3)

	fmt.Printf("\nA Monster Tide is approaching!\n")
	fmt.Printf("Waves: %d\n", numWaves)
	fmt.Printf("Monsters per wave: ~%d\n", monstersPerWave)
	fmt.Printf("Monster Level: ~%d\n\n", baseMonsterLevel)

	// Calculate village defense stats
	totalDefense := 0
	totalAttack := 0

	for _, def := range village.Defenses {
		totalDefense += def.Defense
		totalAttack += def.AttackPower
	}

	// Count guards
	villagerGuards := CountVillagersByRole(village, "guard")
	hiredGuards := len(village.ActiveGuards)
	totalGuards := villagerGuards + hiredGuards

	fmt.Println("YOUR DEFENSES:")
	fmt.Printf("  Defense Power: %d (from %d structures)\n", totalDefense, len(village.Defenses))
	fmt.Printf("  Attack Power: %d (from towers)\n", totalAttack)
	fmt.Printf("  Active Traps: %d\n", len(village.Traps))
	fmt.Printf("  Guards: %d total (%d villagers + %d hired)\n\n", totalGuards, villagerGuards, hiredGuards)

	fmt.Print("Press ENTER to begin defense...")
	bufio.NewScanner(os.Stdin).Scan()

	// Battle statistics
	wavesDefeated := 0
	totalMonstersKilled := 0
	damageDealt := 0
	damageTaken := 0
	trapsTriggered := 0

	// Main wave loop
	for wave := 1; wave <= numWaves; wave++ {
		fmt.Printf("\n\n========== WAVE %d/%d ==========\n", wave, numWaves)

		// Generate wave of monsters
		waveSize := monstersPerWave + rand.Intn(3) - 1 // +/-1 variance
		monsters := make([]models.Monster, waveSize)

		for i := 0; i < waveSize; i++ {
			monsterLevel := baseMonsterLevel + rand.Intn(5) - 2 // +/-2 level variance
			if monsterLevel < 1 {
				monsterLevel = 1
			}
			rank := 1 + rand.Intn(3)
			monsters[i] = GenerateMonster(data.MonsterNames[rand.Intn(len(data.MonsterNames))], monsterLevel, rank)
		}

		fmt.Printf("\n%d monsters approach!\n", len(monsters))
		time.Sleep(1 * time.Second)

		// Process each monster in the wave
		monstersAlive := len(monsters)

		for i := 0; i < len(monsters) && monstersAlive > 0; i++ {
			monster := &monsters[i]

			if monster.HitpointsRemaining <= 0 {
				continue // Already dead
			}

			fmt.Printf("\n  %s (Lv%d, HP:%d) attacks!\n", monster.Name, monster.Level, monster.HitpointsRemaining)

			// PHASE 1: Trap triggering
			trapTriggered := false
			for j := range village.Traps {
				trap := &village.Traps[j]
				if trap.Remaining > 0 {
					// Check if trap triggers
					if rand.Intn(100) < trap.TriggerRate {
						damage := trap.Damage
						monster.HitpointsRemaining -= damage
						fmt.Printf("    %s triggers! (%d damage)\n", trap.Name, damage)
						damageDealt += damage
						trapsTriggered++
						trapTriggered = true

						if monster.HitpointsRemaining <= 0 {
							fmt.Printf("    %s killed by trap!\n", monster.Name)
							monstersAlive--
							totalMonstersKilled++
							break
						}
					}
				}
			}

			if monster.HitpointsRemaining <= 0 {
				continue
			}

			// PHASE 2: Tower attacks
			if totalAttack > 0 && !trapTriggered {
				towerDamage := totalAttack + rand.Intn(5)
				monster.HitpointsRemaining -= towerDamage
				fmt.Printf("    Towers fire! (%d damage)\n", towerDamage)
				damageDealt += towerDamage

				if monster.HitpointsRemaining <= 0 {
					fmt.Printf("    %s killed by towers!\n", monster.Name)
					monstersAlive--
					totalMonstersKilled++
					continue
				}
			}

			// PHASE 3: Guard combat
			if totalGuards > 0 {
				guardDamage := totalGuards * (5 + rand.Intn(8))
				monster.HitpointsRemaining -= guardDamage
				fmt.Printf("    Guards attack! (%d damage)\n", guardDamage)
				damageDealt += guardDamage

				if monster.HitpointsRemaining <= 0 {
					fmt.Printf("    %s killed by guards!\n", monster.Name)
					monstersAlive--
					totalMonstersKilled++
					continue
				}
			}

			// PHASE 4: Monster attacks village
			monsterAttack := monster.AttackRolls * 6
			reducedDamage := monsterAttack - totalDefense
			if reducedDamage < 1 {
				reducedDamage = 1 // Minimum 1 damage
			}

			damageTaken += reducedDamage
			fmt.Printf("    %s breaches defenses! (%d damage to village)\n", monster.Name, reducedDamage)

			// Small delay for readability
			time.Sleep(500 * time.Millisecond)
		}

		// Wave complete
		fmt.Printf("\nWave %d complete!\n", wave)
		fmt.Printf("  Monsters killed: %d/%d\n", waveSize-monstersAlive, waveSize)
		fmt.Printf("  Damage dealt: %d\n", damageDealt)
		fmt.Printf("  Damage taken: %d\n", damageTaken)

		wavesDefeated++

		// Decrement trap durability at end of wave
		for j := len(village.Traps) - 1; j >= 0; j-- {
			village.Traps[j].Remaining--
			if village.Traps[j].Remaining <= 0 {
				fmt.Printf("\n  %s has been consumed!\n", village.Traps[j].Name)
				village.Traps = append(village.Traps[:j], village.Traps[j+1:]...)
			}
		}

		if wave < numWaves {
			fmt.Print("\n  Press ENTER for next wave...")
			bufio.NewScanner(os.Stdin).Scan()
		}
	}

	// Tide complete - calculate results
	fmt.Println("\n\n============================================================")
	fmt.Println("TIDE DEFENSE COMPLETE!")
	fmt.Println("============================================================")

	damageThreshold := village.DefenseLevel * 50 // Higher defense = more damage tolerance

	if damageTaken < damageThreshold {
		// VICTORY
		xpReward := 100 * wavesDefeated
		village.Experience += xpReward

		fmt.Println("\nVICTORY! Your defenses held strong!")
		fmt.Printf("\nBattle Summary:\n")
		fmt.Printf("  Waves Defeated: %d/%d\n", wavesDefeated, numWaves)
		fmt.Printf("  Monsters Killed: %d\n", totalMonstersKilled)
		fmt.Printf("  Damage Dealt: %d\n", damageDealt)
		fmt.Printf("  Damage Taken: %d/%d\n", damageTaken, damageThreshold)
		fmt.Printf("  Traps Triggered: %d times\n", trapsTriggered)
		fmt.Printf("\nRewards:\n")
		fmt.Printf("  Village XP: +%d\n", xpReward)

		// Bonus rewards for perfect defense
		if damageTaken < damageThreshold/2 {
			bonusGold := 50 + (village.Level * 10)
			goldResource := player.ResourceStorageMap["Gold"]
			goldResource.Stock += bonusGold
			player.ResourceStorageMap["Gold"] = goldResource
			fmt.Printf("  Bonus Gold: +%d (minimal damage taken!)\n", bonusGold)
		}

	} else {
		// DEFEAT
		fmt.Println("\nDEFEAT! The tide overwhelmed your defenses!")
		fmt.Printf("\nBattle Summary:\n")
		fmt.Printf("  Waves Survived: %d/%d\n", wavesDefeated, numWaves)
		fmt.Printf("  Monsters Killed: %d\n", totalMonstersKilled)
		fmt.Printf("  Damage Taken: %d/%d (too much!)\n", damageTaken, damageThreshold)

		// Penalties
		resourceLoss := village.Level * 5
		fmt.Printf("\nPenalties:\n")
		fmt.Printf("  Lost %d of each resource type\n", resourceLoss)

		// Deduct resources
		for _, resourceType := range data.ResourceTypes {
			resource := player.ResourceStorageMap[resourceType]
			resource.Stock -= resourceLoss
			if resource.Stock < 0 {
				resource.Stock = 0
			}
			player.ResourceStorageMap[resourceType] = resource
		}

		// Injure some guards
		if len(village.ActiveGuards) > 0 {
			guardsLost := 1 + rand.Intn(len(village.ActiveGuards)/2+1)
			if guardsLost > len(village.ActiveGuards) {
				guardsLost = len(village.ActiveGuards)
			}
			village.ActiveGuards = village.ActiveGuards[:len(village.ActiveGuards)-guardsLost]
			fmt.Printf("  %d hired guards were lost\n", guardsLost)
		}
	}

	// Update last tide time
	village.LastTideTime = time.Now().Unix()

	// Upgrade village
	UpgradeVillage(village)

	fmt.Println("\n============================================================")
	fmt.Print("Press ENTER to continue...")
	bufio.NewScanner(os.Stdin).Scan()
}
