package game

import (
	"fmt"

	"rpg-game/pkg/models"
)

func PrintCharacter(player models.Character) {
	fmt.Printf("------------\nPlayer Stats: %s\nLevel: %d\nExperience: %d\nTotalLife: %d\nRemainingLife: %d\nAttackRolls: %d\nDefenseRolls: %d\nAttackMod: %d\nDefenseMod: %d\nHitPointMod: %d\nResurrections: %d\n------------\n", player.Name, player.Level, player.Experience, player.HitpointsTotal, player.HitpointsRemaining, player.AttackRolls, player.DefenseRolls, player.StatsMod.AttackMod, player.StatsMod.DefenseMod, player.StatsMod.HitPointMod, player.Resurrections)
	for i := 0; i < len(player.EquipmentMap); i++ {
		PrintItem(player.EquipmentMap[i])
	}
	for i := 0; i < len(player.KnownLocations); i++ {
		fmt.Println(player.KnownLocations[i])
	}
}

func PrintMonster(mob models.Monster) {
	fmt.Printf("------------\nMonster Stats: %s\nLevel: %d\nTotalLife: %d\nRemainingLife: %d\nAttackRolls: %d\nDefenseRolls: %d\n------------\n", mob.Name, mob.Level, mob.HitpointsTotal, mob.HitpointsRemaining, mob.AttackRolls, mob.DefenseRolls)
}

func PrintItem(item models.Item) {
	fmt.Printf("------------\nItem Stats: %s\nSlot: %d \nRarity: %d\nAttackMod: %d\nDefenseMod: %d\nHitPointMod: %d\n\n------------\n", item.Name, item.Slot, item.Rarity, item.StatsMod.AttackMod, item.StatsMod.DefenseMod, item.StatsMod.HitPointMod)
}

func PrintResources(resourceStorage map[string]models.Resource) {
	fmt.Println("Total resources:")
	for _, resource := range resourceStorage {
		fmt.Printf("%s : %d\n", resource.Name, resource.Stock)
	}
}

func PrintMonstersAtLocation(loc models.Location) {
	fmt.Println("Monsters At ", loc.Name)
	guardianCount := 0
	for _, mob := range loc.Monsters {
		guardianTag := ""
		if mob.IsSkillGuardian {
			guardianTag = " [GUARDIAN - " + mob.GuardedSkill.Name + "]"
			guardianCount++
		}
		fmt.Printf("%s Level: %d Exp: %d%s\n", mob.Name, mob.Level, mob.Experience, guardianTag)
	}
	if guardianCount > 0 {
		fmt.Printf("  %d Skill Guardian(s) present! Defeat them to learn skills.\n", guardianCount)
	}
}

func ShowInventory(player *models.Character) {
	fmt.Println("\n============================================================")
	fmt.Printf("INVENTORY - %s\n", player.Name)
	fmt.Println("============================================================")

	if len(player.Inventory) == 0 {
		fmt.Println("Your inventory is empty.")
	} else {
		equipment := []models.Item{}
		consumables := []models.Item{}
		skillScrolls := []models.Item{}

		for _, item := range player.Inventory {
			if item.ItemType == "consumable" {
				consumables = append(consumables, item)
			} else if item.ItemType == "skill_scroll" {
				skillScrolls = append(skillScrolls, item)
			} else {
				equipment = append(equipment, item)
			}
		}

		if len(consumables) > 0 {
			fmt.Println("\nCONSUMABLES:")
			potionCount := make(map[string]int)
			for _, item := range consumables {
				potionCount[item.Name]++
			}
			for name, count := range potionCount {
				fmt.Printf("  %s x%d\n", name, count)
			}
		}

		if len(skillScrolls) > 0 {
			fmt.Println("\nSKILL SCROLLS:")
			for i, scroll := range skillScrolls {
				fmt.Printf("  %d. %s\n", i+1, scroll.Name)
				fmt.Printf("     Skill: %s\n", scroll.SkillScroll.Skill.Name)
				fmt.Printf("     Crafting Value: %d\n", scroll.SkillScroll.CraftingValue)
				fmt.Printf("     Use: Learn skill or craft into equipment\n")
			}
		}

		if len(equipment) > 0 {
			fmt.Println("\nEQUIPMENT (Unequipped):")
			for i, item := range equipment {
				fmt.Printf("  %d. %s (Rarity %d, CP: %d)\n", i+1, item.Name, item.Rarity, item.CP)
				if item.StatsMod.AttackMod > 0 {
					fmt.Printf("     +%d Attack\n", item.StatsMod.AttackMod)
				}
				if item.StatsMod.DefenseMod > 0 {
					fmt.Printf("     +%d Defense\n", item.StatsMod.DefenseMod)
				}
				if item.StatsMod.HitPointMod > 0 {
					fmt.Printf("     +%d HP\n", item.StatsMod.HitPointMod)
				}
			}
		}
	}

	fmt.Printf("\nTotal Items: %d\n", len(player.Inventory))
	fmt.Println("============================================================")
}

func ShowSkills(player *models.Character) {
	fmt.Println("\n============================================================")
	fmt.Printf("LEARNED SKILLS - %s\n", player.Name)
	fmt.Println("============================================================")
	fmt.Printf("Level: %d | MP: %d/%d | SP: %d/%d\n\n",
		player.Level, player.ManaRemaining, player.ManaTotal,
		player.StaminaRemaining, player.StaminaTotal)

	if len(player.LearnedSkills) == 0 {
		fmt.Println("No skills learned yet.")
	} else {
		for i, skill := range player.LearnedSkills {
			fmt.Printf("%d. %s\n", i+1, skill.Name)

			costs := []string{}
			if skill.ManaCost > 0 {
				costs = append(costs, fmt.Sprintf("%d MP", skill.ManaCost))
			}
			if skill.StaminaCost > 0 {
				costs = append(costs, fmt.Sprintf("%d SP", skill.StaminaCost))
			}
			if len(costs) > 0 {
				fmt.Printf("   Cost: %s\n", costs[0])
				if len(costs) > 1 {
					fmt.Printf("        %s\n", costs[1])
				}
			}

			if skill.Damage > 0 {
				fmt.Printf("   Damage: %d %s\n", skill.Damage, skill.DamageType)
			} else if skill.Damage < 0 {
				fmt.Printf("   Healing: %d HP\n", -skill.Damage)
			}

			if skill.Effect.Type != "none" && skill.Effect.Type != "" {
				fmt.Printf("   Effect: %s (%d turns, potency %d)\n",
					skill.Effect.Type, skill.Effect.Duration, skill.Effect.Potency)
			}

			fmt.Printf("   %s\n\n", skill.Description)
		}
	}

	fmt.Printf("Total Skills: %d\n", len(player.LearnedSkills))
	fmt.Println("============================================================")
}

func ShowEquipment(player *models.Character) {
	fmt.Println("\n============================================================")
	fmt.Printf("EQUIPPED ITEMS - %s\n", player.Name)
	fmt.Println("============================================================")

	if len(player.EquipmentMap) == 0 {
		fmt.Println("No equipment equipped.")
	} else {
		slotNames := map[int]string{
			0: "Head",
			1: "Chest",
			2: "Legs",
			3: "Feet",
			4: "Hands",
			5: "Main Hand",
			6: "Off Hand",
			7: "Accessory",
		}

		for slot, item := range player.EquipmentMap {
			slotName := slotNames[slot]
			if slotName == "" {
				slotName = fmt.Sprintf("Slot %d", slot)
			}

			fmt.Printf("\n[%s]\n", slotName)
			fmt.Printf("  %s (Rarity %d, CP: %d)\n", item.Name, item.Rarity, item.CP)

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

		fmt.Printf("\n\nTotal Stats from Equipment:\n")
		fmt.Printf("  Attack:  +%d\n", player.StatsMod.AttackMod)
		fmt.Printf("  Defense: +%d\n", player.StatsMod.DefenseMod)
		fmt.Printf("  HP:      +%d\n", player.StatsMod.HitPointMod)
	}

	fmt.Printf("\nEquipped Items: %d\n", len(player.EquipmentMap))
	fmt.Println("============================================================")
}
