package game

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strconv"

	"rpg-game/pkg/models"
)

func AutoFightToTheDeath(player *models.Character, game *models.GameState, mob *models.Monster, location *models.Location, mobLoc int) {
	// Restore resources at start
	player.ManaRemaining = player.ManaTotal
	player.StaminaRemaining = player.StaminaTotal
	mob.ManaRemaining = mob.ManaTotal
	mob.StaminaRemaining = mob.StaminaTotal

	turnCount := 0
	fmt.Printf("Fight #%d: %s (Lv%d) vs %s (Lv%d)\n",
		rand.Intn(10000), player.Name, player.Level, mob.Name, mob.Level)

	for player.HitpointsRemaining > 0 && mob.HitpointsRemaining > 0 {
		turnCount++

		// Process status effects
		ProcessStatusEffects(player)
		ProcessStatusEffectsMob(mob)

		// Declare decision variable before potential goto
		var decision string

		// Check if player is stunned
		if IsStunned(player) {
			fmt.Printf("  [T%d] %s is STUNNED!\n", turnCount, player.Name)
			goto MonsterTurn
		}

		// AI makes decision
		decision = MakeAIDecision(player, mob, turnCount)

		// Execute decision
		switch decision {
		case "attack":
			playerAttack := MultiRoll(player.AttackRolls) + player.StatsMod.AttackMod
			if rand.Intn(100) < 15 {
				playerAttack = playerAttack * 2
				fmt.Printf("  [T%d] %s CRITICAL HIT!\n", turnCount, player.Name)
			}
			mobDef := MultiRoll(mob.DefenseRolls) + mob.StatsMod.DefenseMod
			if playerAttack > mobDef {
				diff := ApplyDamage(playerAttack-mobDef, models.Physical, mob)
				mob.HitpointsRemaining -= diff
				fmt.Printf("  [T%d] %s attacks for %d dmg (Mob HP: %d/%d)\n",
					turnCount, player.Name, diff, mob.HitpointsRemaining, mob.HitpointsTotal)
			}

		case "item":
			// Use first available potion
			for idx, item := range player.Inventory {
				if item.ItemType == "consumable" {
					UseConsumableItem(item, player)
					RemoveItemFromInventory(&player.Inventory, idx)
					fmt.Printf("  [T%d] %s used %s\n", turnCount, player.Name, item.Name)
					break
				}
			}

		default:
			// Skill usage (decision starts with "skill_")
			if len(decision) > 6 && decision[:6] == "skill_" {
				skillName := decision[6:]
				for _, skill := range player.LearnedSkills {
					if skill.Name == skillName || skill.Name == "Heal" && skillName == "heal" ||
						skill.Name == "Regeneration" && skillName == "regeneration" ||
						skill.Name == "Battle Cry" && skillName == "Battle Cry" ||
						skill.Name == "Shield Wall" && skillName == "Shield Wall" {

						if skill.ManaCost <= player.ManaRemaining && skill.StaminaCost <= player.StaminaRemaining {
							player.ManaRemaining -= skill.ManaCost
							player.StaminaRemaining -= skill.StaminaCost

							if skill.Damage < 0 {
								// Healing
								player.HitpointsRemaining += -skill.Damage
								if player.HitpointsRemaining > player.HitpointsTotal {
									player.HitpointsRemaining = player.HitpointsTotal
								}
								fmt.Printf("  [T%d] %s used %s (+%d HP)\n", turnCount, player.Name, skill.Name, -skill.Damage)
							} else if skill.Damage > 0 {
								// Damage skill
								finalDamage := ApplyDamage(skill.Damage, skill.DamageType, mob)
								mob.HitpointsRemaining -= finalDamage
								fmt.Printf("  [T%d] %s used %s (%d %s dmg)\n",
									turnCount, player.Name, skill.Name, finalDamage, skill.DamageType)
							} else {
								// Buff
								fmt.Printf("  [T%d] %s used %s (buff)\n", turnCount, player.Name, skill.Name)
							}

							// Apply effects
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
						}
						break
					}
				}
			}
		}

	MonsterTurn:
		// Monster's turn
		if mob.HitpointsRemaining > 0 {
			if IsStunnedMob(mob) {
				// Monster stunned, skip turn
			} else {
				// Simple monster attack
				mobAttack := MultiRoll(mob.AttackRolls) + mob.StatsMod.AttackMod
				if rand.Intn(100) < 10 {
					mobAttack = mobAttack * 2
				}
				playerDef := MultiRoll(player.DefenseRolls) + player.StatsMod.DefenseMod
				if mobAttack > playerDef {
					diff := ApplyDamage(mobAttack-playerDef, models.Physical, player)
					player.HitpointsRemaining -= diff
				}
			}
		}
	}

	// Combat resolution
	if player.HitpointsRemaining > 0 {
		player.Experience += mob.Level * 10
		fmt.Printf("  VICTORY! (+%d XP)\n", mob.Level*10)

		// Loot
		for _, item := range mob.EquipmentMap {
			EquipBestItem(item, &player.EquipmentMap, &player.Inventory)
		}

		// Drop beast materials based on monster type
		DropBeastMaterial(mob.MonsterType, player)

		// Chance for potion
		if rand.Intn(100) < 30 {
			potion := CreateHealthPotion("small")
			if rand.Intn(100) < 30 {
				potion = CreateHealthPotion("medium")
			}
			player.Inventory = append(player.Inventory, potion)
		}

		// 15% chance to rescue a villager after victory
		if rand.Intn(100) < 15 {
			if game.Villages == nil {
				game.Villages = make(map[string]models.Village)
			}

			village, exists := game.Villages[player.VillageName]
			if !exists {
				village = GenerateVillage(player.Name)
				player.VillageName = player.Name + "'s Village"
			}

			RescueVillager(&village)

			village.Experience += 25
			fmt.Println("  +25 Village XP")

			game.Villages[player.VillageName] = village
		}

		player.StatsMod = CalculateItemMods(player.EquipmentMap)
		location.Monsters[mobLoc] = GenerateBestMonster(game, location.LevelMax, location.RarityMax)
	} else {
		fmt.Printf("  DEFEAT!\n")
		for _, item := range player.EquipmentMap {
			EquipBestItem(item, &mob.EquipmentMap, &mob.Inventory)
		}
		location.Monsters[mobLoc].StatsMod = CalculateItemMods(mob.EquipmentMap)
		location.Monsters[mobLoc].Experience += player.Level * 100
	}

	// Process guard recovery after combat
	if game.Villages != nil {
		if village, exists := game.Villages[player.VillageName]; exists {
			ProcessGuardRecovery(&village)
			game.Villages[player.VillageName] = village
		}
	}

	fmt.Println()
}

func FightToTheDeath(player *models.Character, game *models.GameState, mob *models.Monster, location *models.Location, mobLoc int) {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Printf("\n============================================\n")
	fmt.Printf("Level %d %s vs Level %d %s (%s)\n", player.Level, player.Name, mob.Level, mob.Name, mob.MonsterType)
	fmt.Printf("============================================\n")

	// Check if player has guards available for this fight
	var combatGuards []models.Guard
	isSpecialFight := mob.IsSkillGuardian || mob.IsBoss

	if game.Villages != nil {
		if village, exists := game.Villages[player.VillageName]; exists && isSpecialFight {
			for _, guard := range village.ActiveGuards {
				if !guard.Injured && guard.HitpointsRemaining > 0 {
					guard.HitpointsRemaining = guard.HitPoints
					combatGuards = append(combatGuards, guard)
				}
			}

			if len(combatGuards) > 0 {
				if mob.IsBoss {
					fmt.Printf("\nWARNING: BOSS FIGHT\n")
					fmt.Printf("Guards can DIE PERMANENTLY in boss fights!\n")
					fmt.Printf("Available guards: %d\n", len(combatGuards))
					fmt.Print("Bring guards to this fight? (y/n): ")
					scanner.Scan()
					bringGuards := scanner.Text()
					if bringGuards != "y" && bringGuards != "Y" {
						combatGuards = []models.Guard{}
						fmt.Println("Fighting without guards...")
					} else {
						fmt.Printf("\nGUARDS JOINING BOSS BATTLE!\n")
						for _, guard := range combatGuards {
							fmt.Printf("   %s (Lv%d, HP:%d)\n", guard.Name, guard.Level, guard.HitPoints)
						}
						fmt.Println()
					}
				} else {
					fmt.Printf("\nGUARDS JOINING BATTLE!\n")
					for _, guard := range combatGuards {
						fmt.Printf("   %s (Lv%d, HP:%d)\n", guard.Name, guard.Level, guard.HitPoints)
					}
					fmt.Println()
				}
			}
		}
	}

	// Restore mana/stamina at combat start
	player.ManaRemaining = player.ManaTotal
	player.StaminaRemaining = player.StaminaTotal
	mob.ManaRemaining = mob.ManaTotal
	mob.StaminaRemaining = mob.StaminaTotal

	playerFled := false
	turnCount := 0

	for player.HitpointsRemaining > 0 && mob.HitpointsRemaining > 0 && !playerFled {
		turnCount++
		fmt.Printf("\n========== TURN %d ==========\n", turnCount)

		ProcessStatusEffects(player)
		ProcessStatusEffectsMob(mob)

		fmt.Printf("[%s] HP:%d/%d | MP:%d/%d | SP:%d/%d\n",
			player.Name,
			player.HitpointsRemaining, player.HitpointsTotal,
			player.ManaRemaining, player.ManaTotal,
			player.StaminaRemaining, player.StaminaTotal)

		fmt.Printf("[%s] HP:%d/%d | MP:%d/%d | SP:%d/%d\n",
			mob.Name,
			mob.HitpointsRemaining, mob.HitpointsTotal,
			mob.ManaRemaining, mob.ManaTotal,
			mob.StaminaRemaining, mob.StaminaTotal)

		if len(player.StatusEffects) > 0 {
			fmt.Printf("%s effects: ", player.Name)
			for _, eff := range player.StatusEffects {
				fmt.Printf("[%s:%d] ", eff.Type, eff.Duration)
			}
			fmt.Println()
		}
		if len(mob.StatusEffects) > 0 {
			fmt.Printf("%s effects: ", mob.Name)
			for _, eff := range mob.StatusEffects {
				fmt.Printf("[%s:%d] ", eff.Type, eff.Duration)
			}
			fmt.Println()
		}

		var playerAttack int
		var playerDef int
		var defending bool
		var skipPlayerTurn bool
		var usedSkillDamage int
		var usedSkillType models.DamageType = models.Physical
		var action string

		if IsStunned(player) {
			fmt.Printf("%s is STUNNED and cannot act!\n", player.Name)
			playerDef = MultiRoll(player.DefenseRolls) + player.StatsMod.DefenseMod
			goto MonsterTurn
		}

		fmt.Println("\n--- Your Action ---")
		fmt.Println("1 = Attack (physical)")
		fmt.Println("2 = Defend (+50% defense, 50% attack)")
		fmt.Println("3 = Use Item")
		fmt.Println("4 = Use Skill")
		fmt.Println("5 = Flee")
		fmt.Print("Choice: ")

		scanner.Scan()
		action = scanner.Text()

		switch action {
		case "1": // Attack
			playerAttack = MultiRoll(player.AttackRolls) + player.StatsMod.AttackMod
			playerDef = MultiRoll(player.DefenseRolls) + player.StatsMod.DefenseMod
			if rand.Intn(100) < 15 {
				playerAttack = playerAttack * 2
				fmt.Printf("*** CRITICAL HIT! ***\n")
			}

		case "2": // Defend
			defending = true
			playerAttack = (MultiRoll(player.AttackRolls) + player.StatsMod.AttackMod) / 2
			playerDef = int(float64(MultiRoll(player.DefenseRolls)+player.StatsMod.DefenseMod) * 1.5)
			fmt.Printf("%s takes a defensive stance!\n", player.Name)

		case "3": // Use Item
			consumables := []models.Item{}
			consumableIndices := []int{}
			for idx, item := range player.Inventory {
				if item.ItemType == "consumable" {
					consumables = append(consumables, item)
					consumableIndices = append(consumableIndices, idx)
				}
			}

			if len(consumables) == 0 {
				fmt.Println("No consumable items!")
				skipPlayerTurn = true
			} else {
				fmt.Println("Available items:")
				for idx, item := range consumables {
					fmt.Printf("%d = %s (Heals %d HP)\n", idx+1, item.Name, item.Consumable.Value)
				}
				fmt.Print("Choose (0=cancel): ")
				scanner.Scan()
				itemChoice := scanner.Text()
				itemIdx, err := strconv.Atoi(itemChoice)

				if err != nil || itemIdx < 0 || itemIdx > len(consumables) {
					fmt.Println("Invalid choice!")
					skipPlayerTurn = true
				} else if itemIdx == 0 {
					fmt.Println("Cancelled.")
					skipPlayerTurn = true
				} else {
					selectedItem := consumables[itemIdx-1]
					originalIdx := consumableIndices[itemIdx-1]
					UseConsumableItem(selectedItem, player)
					RemoveItemFromInventory(&player.Inventory, originalIdx)
				}
			}
			playerDef = MultiRoll(player.DefenseRolls) + player.StatsMod.DefenseMod

		case "4": // Use Skill
			if len(player.LearnedSkills) == 0 {
				fmt.Println("No skills learned!")
				skipPlayerTurn = true
			} else {
				fmt.Println("\nAvailable Skills:")
				for idx, skill := range player.LearnedSkills {
					canAfford := "Y"
					if skill.ManaCost > player.ManaRemaining || skill.StaminaCost > player.StaminaRemaining {
						canAfford = "N"
					}
					fmt.Printf("%d [%s] %s - ", idx+1, canAfford, skill.Name)
					if skill.ManaCost > 0 {
						fmt.Printf("%dMP ", skill.ManaCost)
					}
					if skill.StaminaCost > 0 {
						fmt.Printf("%dSP ", skill.StaminaCost)
					}
					fmt.Printf("| %s\n", skill.Description)
				}
				fmt.Print("Choose skill (0=cancel): ")
				scanner.Scan()
				skillChoice := scanner.Text()
				skillIdx, err := strconv.Atoi(skillChoice)

				if err != nil || skillIdx < 0 || skillIdx > len(player.LearnedSkills) {
					fmt.Println("Invalid choice!")
					skipPlayerTurn = true
				} else if skillIdx == 0 {
					fmt.Println("Cancelled.")
					skipPlayerTurn = true
				} else {
					skill := player.LearnedSkills[skillIdx-1]

					if skill.ManaCost > player.ManaRemaining {
						fmt.Println("Not enough mana!")
						skipPlayerTurn = true
					} else if skill.StaminaCost > player.StaminaRemaining {
						fmt.Println("Not enough stamina!")
						skipPlayerTurn = true
					} else {
						player.ManaRemaining -= skill.ManaCost
						player.StaminaRemaining -= skill.StaminaCost
						fmt.Printf("%s uses %s!\n", player.Name, skill.Name)

						if skill.Damage < 0 {
							healAmount := -skill.Damage
							player.HitpointsRemaining += healAmount
							if player.HitpointsRemaining > player.HitpointsTotal {
								player.HitpointsRemaining = player.HitpointsTotal
							}
							fmt.Printf("%s heals for %d HP!\n", player.Name, healAmount)
						} else if skill.Damage > 0 {
							usedSkillDamage = skill.Damage
							usedSkillType = skill.DamageType
						}

						if skill.Effect.Type != "none" && skill.Effect.Duration > 0 {
							if skill.Effect.Type == "buff_attack" || skill.Effect.Type == "buff_defense" || skill.Effect.Type == "regen" {
								player.StatusEffects = append(player.StatusEffects, skill.Effect)
								if skill.Effect.Type == "buff_attack" {
									player.StatsMod.AttackMod += skill.Effect.Potency
								} else if skill.Effect.Type == "buff_defense" {
									player.StatsMod.DefenseMod += skill.Effect.Potency
								}
								fmt.Printf("%s gains %s effect!\n", player.Name, skill.Effect.Type)
							} else {
								mob.StatusEffects = append(mob.StatusEffects, skill.Effect)
								fmt.Printf("%s is afflicted with %s!\n", mob.Name, skill.Effect.Type)
							}
						}
					}
				}
			}
			playerDef = MultiRoll(player.DefenseRolls) + player.StatsMod.DefenseMod

		case "5": // Flee
			fleeChance := 50 + (player.Level-mob.Level)*5
			if fleeChance > 90 {
				fleeChance = 90
			}
			if fleeChance < 20 {
				fleeChance = 20
			}

			roll := rand.Intn(100)
			if roll < fleeChance {
				fmt.Printf("%s successfully fled from combat!\n", player.Name)
				playerFled = true
				continue
			} else {
				fmt.Printf("%s tried to flee but failed!\n", player.Name)
				skipPlayerTurn = true
				playerDef = MultiRoll(player.DefenseRolls) + player.StatsMod.DefenseMod
			}

		default:
			fmt.Println("Invalid action! Defaulting to Attack.")
			playerAttack = MultiRoll(player.AttackRolls) + player.StatsMod.AttackMod
			playerDef = MultiRoll(player.DefenseRolls) + player.StatsMod.DefenseMod
		}

		// Player attacks (if not skipped)
		if !skipPlayerTurn && (playerAttack > 0 || usedSkillDamage > 0) {
			mobDef := MultiRoll(mob.DefenseRolls) + mob.StatsMod.DefenseMod

			if usedSkillDamage > 0 {
				finalDamage := ApplyDamage(usedSkillDamage, usedSkillType, mob)
				mob.HitpointsRemaining -= finalDamage

				if usedSkillType != models.Physical {
					fmt.Printf("Deals %d %s damage", finalDamage, usedSkillType)
					resistance := mob.Resistances[usedSkillType]
					if resistance < 1.0 {
						fmt.Printf(" (resistant!)")
					} else if resistance > 1.0 {
						fmt.Printf(" (weak!)")
					}
					fmt.Println()
				} else {
					fmt.Printf("Deals %d damage!\n", finalDamage)
				}
			} else if playerAttack > mobDef {
				diff := playerAttack - mobDef
				finalDamage := ApplyDamage(diff, models.Physical, mob)
				mob.HitpointsRemaining -= finalDamage

				if defending {
					fmt.Printf("%s counterattacks for %d damage!\n", player.Name, finalDamage)
				} else {
					fmt.Printf("%s attacks for %d damage!\n", player.Name, finalDamage)
				}
			} else {
				fmt.Printf("%s's attack missed!\n", player.Name)
			}
		}

		// Guards attack
		if len(combatGuards) > 0 && mob.HitpointsRemaining > 0 && !skipPlayerTurn {
			fmt.Println("\n--- Guard Support ---")
			guardDamage := GuardAttack(combatGuards, mob)
			if guardDamage > 0 {
				fmt.Printf("   Total guard damage: %d\n", guardDamage)
			}
		}

	MonsterTurn:
		// Monster attacks back
		if mob.HitpointsRemaining > 0 && !playerFled {
			if IsStunnedMob(mob) {
				fmt.Printf("%s is STUNNED and cannot act!\n", mob.Name)
			} else {
				useMonsterSkill := false
				if len(mob.LearnedSkills) > 0 && rand.Intn(100) < 40 {
					skill := mob.LearnedSkills[rand.Intn(len(mob.LearnedSkills))]
					if skill.ManaCost <= mob.ManaRemaining && skill.StaminaCost <= mob.StaminaRemaining {
						mob.ManaRemaining -= skill.ManaCost
						mob.StaminaRemaining -= skill.StaminaCost
						useMonsterSkill = true

						fmt.Printf("%s uses %s!\n", mob.Name, skill.Name)

						if skill.Damage < 0 {
							healAmount := -skill.Damage
							mob.HitpointsRemaining += healAmount
							if mob.HitpointsRemaining > mob.HitpointsTotal {
								mob.HitpointsRemaining = mob.HitpointsTotal
							}
							fmt.Printf("%s heals for %d HP!\n", mob.Name, healAmount)
						} else if skill.Damage > 0 {
							finalDamage := ApplyDamage(skill.Damage, skill.DamageType, player)

							if len(combatGuards) > 0 {
								finalDamage, _ = GuardDefense(combatGuards, finalDamage)
							}

							player.HitpointsRemaining -= finalDamage
							fmt.Printf("Deals %d damage to %s!\n", finalDamage, player.Name)
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
								fmt.Printf("%s is afflicted with %s!\n", player.Name, skill.Effect.Type)
							}
						}
					}
				}

				if !useMonsterSkill {
					mobAttack := MultiRoll(mob.AttackRolls) + mob.StatsMod.AttackMod
					if rand.Intn(100) < 10 {
						mobAttack = mobAttack * 2
						fmt.Printf("*** %s CRITICAL HIT! ***\n", mob.Name)
					}

					if mobAttack > playerDef {
						diff := mobAttack - playerDef
						finalDamage := ApplyDamage(diff, models.Physical, player)

						if len(combatGuards) > 0 {
							finalDamage, _ = GuardDefense(combatGuards, finalDamage)
						}

						player.HitpointsRemaining -= finalDamage
						fmt.Printf("%s attacks for %d damage!\n", mob.Name, finalDamage)
					} else {
						fmt.Printf("%s's attack missed!\n", mob.Name)
					}
				}
			}
		}
	}

	// Combat resolution
	if playerFled {
		fmt.Println("\n========================================")
		fmt.Println("You escaped safely, but gained no rewards.")
		fmt.Println("========================================")
	} else if player.HitpointsRemaining > 0 {
		player.Experience += mob.Level * 10
		fmt.Printf("\n========================================\n")
		fmt.Printf("VICTORY! %s Wins! (+%d XP)\n", player.Name, mob.Level*10)
		fmt.Printf("========================================\n")

		// Check if this was a Skill Guardian
		if mob.IsSkillGuardian {
			fmt.Printf("\nSKILL GUARDIAN DEFEATED!\n")
			fmt.Printf("You have defeated %s and can now learn: %s\n", mob.Name, mob.GuardedSkill.Name)
			fmt.Printf("Description: %s\n\n", mob.GuardedSkill.Description)
			fmt.Println("Choose your reward:")
			fmt.Println("1 = Absorb the skill immediately (learn now)")
			fmt.Println("2 = Take a skill scroll (can learn later or use for crafting)")
			fmt.Print("Choice: ")

			scanner := bufio.NewScanner(os.Stdin)
			scanner.Scan()
			choice := scanner.Text()

			switch choice {
			case "1":
				player.LearnedSkills = append(player.LearnedSkills, mob.GuardedSkill)
				fmt.Printf("\nYou have learned %s!\n", mob.GuardedSkill.Name)
				fmt.Printf("You can now use this skill in combat.\n\n")
			case "2":
				scroll := CreateSkillScroll(mob.GuardedSkill)
				player.Inventory = append(player.Inventory, scroll)
				fmt.Printf("\nYou received a %s!\n", scroll.Name)
				fmt.Printf("You can use it later to learn the skill or craft it into equipment.\n")
				fmt.Printf("Crafting Value: %d\n\n", scroll.SkillScroll.CraftingValue)
			default:
				scroll := CreateSkillScroll(mob.GuardedSkill)
				player.Inventory = append(player.Inventory, scroll)
				fmt.Printf("\nYou received a %s!\n", scroll.Name)
			}
		}

		// Loot enemy equipment
		for _, item := range mob.EquipmentMap {
			EquipBestItem(item, &player.EquipmentMap, &player.Inventory)
		}

		// Drop beast materials
		DropBeastMaterial(mob.MonsterType, player)

		// 30% chance to get a health potion
		if rand.Intn(100) < 30 {
			potionSize := "small"
			roll := rand.Intn(100)
			if roll < 50 {
				potionSize = "small"
			} else if roll < 85 {
				potionSize = "medium"
			} else {
				potionSize = "large"
			}
			potion := CreateHealthPotion(potionSize)
			player.Inventory = append(player.Inventory, potion)
			fmt.Printf("Found a %s!\n", potion.Name)
		}

		// 15% chance to rescue a villager after victory
		if rand.Intn(100) < 15 {
			if game.Villages == nil {
				game.Villages = make(map[string]models.Village)
			}

			village, exists := game.Villages[player.VillageName]
			if !exists {
				village = GenerateVillage(player.Name)
				player.VillageName = player.Name + "'s Village"
			}

			RescueVillager(&village)
			village.Experience += 25
			fmt.Println("+25 Village XP")
			game.Villages[player.VillageName] = village
		}

		player.StatsMod = CalculateItemMods(player.EquipmentMap)
		location.Monsters[mobLoc] = GenerateBestMonster(game, location.LevelMax, location.RarityMax)

		// Update guard states in village after combat
		if len(combatGuards) > 0 {
			if village, exists := game.Villages[player.VillageName]; exists {
				deadGuards := []string{}

				for i := len(village.ActiveGuards) - 1; i >= 0; i-- {
					guard := village.ActiveGuards[i]
					for _, combatGuard := range combatGuards {
						if guard.Name == combatGuard.Name {
							if mob.IsBoss && combatGuard.HitpointsRemaining <= 0 {
								deadGuards = append(deadGuards, combatGuard.Name)
								village.ActiveGuards = append(village.ActiveGuards[:i], village.ActiveGuards[i+1:]...)
							} else {
								village.ActiveGuards[i].HitpointsRemaining = combatGuard.HitpointsRemaining
								village.ActiveGuards[i].Injured = combatGuard.Injured
								village.ActiveGuards[i].RecoveryTime = combatGuard.RecoveryTime
							}
							break
						}
					}
				}

				if len(deadGuards) > 0 {
					fmt.Println("\nGUARDS FALLEN")
					for _, guardName := range deadGuards {
						fmt.Printf("   %s has died in battle! (PERMANENT LOSS)\n", guardName)
					}
					fmt.Println()
				}

				game.Villages[player.VillageName] = village
			}
		}
	} else {
		fmt.Printf("\n========================================\n")
		fmt.Printf("DEFEAT! %s HAS DIED!\n", player.Name)
		fmt.Printf("%s Wins!\n", mob.Name)
		fmt.Printf("========================================\n")

		for _, item := range player.EquipmentMap {
			EquipBestItem(item, &mob.EquipmentMap, &mob.Inventory)
		}

		location.Monsters[mobLoc].StatsMod = CalculateItemMods(mob.EquipmentMap)
		location.Monsters[mobLoc].Experience += player.Level * 100
	}

	// Process guard recovery after combat
	if game.Villages != nil {
		if village, exists := game.Villages[player.VillageName]; exists {
			ProcessGuardRecovery(&village)
			game.Villages[player.VillageName] = village
		}
	}
}
