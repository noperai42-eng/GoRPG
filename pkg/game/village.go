package game

import (
	"fmt"
	"math/rand"
	"time"

	"rpg-game/pkg/data"
	"rpg-game/pkg/models"
)

func GenerateVillage(playerName string) models.Village {
	villageName := playerName + "'s Village"
	return models.Village{
		Name:             villageName,
		Level:            1,
		Experience:       0,
		Villagers:        []models.Villager{},
		Defenses:         []models.Defense{},
		ResourcePerTick:  make(map[string]int),
		UnlockedCrafting: []string{},
		DefenseLevel:     1,
		LastTideTime:     time.Now().Unix(),
		TideInterval:     3600,
		ActiveGuards:     []models.Guard{},
		LastHarvestTime:  time.Now().Unix(),
	}
}

func GenerateVillager(role string) models.Villager {
	firstName := data.VillagerFirstNames[rand.Intn(len(data.VillagerFirstNames))]
	lastName := data.VillagerLastNames[rand.Intn(len(data.VillagerLastNames))]
	name := firstName + " " + lastName
	efficiency := rand.Intn(3) + 1
	return models.Villager{
		Name:         name,
		Role:         role,
		Level:        1,
		Efficiency:   efficiency,
		AssignedTask: "",
		HarvestType:  "",
	}
}

func RescueVillager(village *models.Village) models.Villager {
	role := "harvester"
	if rand.Intn(100) < 30 {
		role = "guard"
	}
	villager := GenerateVillager(role)
	village.Villagers = append(village.Villagers, villager)
	fmt.Printf("🎉 You rescued %s!\n", villager.Name)
	fmt.Printf("🏘️ %s has joined your village as a %s (Efficiency: %d)\n", villager.Name, villager.Role, villager.Efficiency)
	return villager
}

// HarvestResult describes a single villager's harvest collection.
type HarvestResult struct {
	VillagerName string
	Amount       int
	ResourceType string
}

// ProcessVillageResourceCollection collects resources from active harvesters,
// updates the player's resource storage, and returns what was collected.
func ProcessVillageResourceCollection(village *models.Village, player *models.Character) []HarvestResult {
	var results []HarvestResult
	for _, villager := range village.Villagers {
		if villager.Role == "harvester" && villager.HarvestType != "" {
			amount := villager.Efficiency + villager.Level/2
			resource, exists := player.ResourceStorageMap[villager.HarvestType]
			if exists {
				resource.Stock += amount
				player.ResourceStorageMap[villager.HarvestType] = resource
			} else {
				player.ResourceStorageMap[villager.HarvestType] = models.Resource{
					Name:         villager.HarvestType,
					Stock:        amount,
					RollModifier: 0,
				}
			}
			results = append(results, HarvestResult{
				VillagerName: villager.Name,
				Amount:       amount,
				ResourceType: villager.HarvestType,
			})
		}
	}
	return results
}

// ShouldHarvest returns true if 60+ seconds have elapsed since the last harvest.
func ShouldHarvest(village *models.Village) bool {
	return time.Now().Unix()-village.LastHarvestTime >= 60
}

// HasActiveHarvesters returns true if any villager is actively harvesting.
func HasActiveHarvesters(village *models.Village) bool {
	for _, v := range village.Villagers {
		if v.Role == "harvester" && v.HarvestType != "" {
			return true
		}
	}
	return false
}

func UpgradeVillage(village *models.Village) {
	requiredXP := village.Level * 100
	if village.Experience >= requiredXP {
		village.Level++
		village.Experience -= requiredXP
		fmt.Printf("🎉 Village leveled up! Now level %d!\n", village.Level)

		if village.Level >= 3 && !Contains(village.UnlockedCrafting, "potions") {
			village.UnlockedCrafting = append(village.UnlockedCrafting, "potions")
			fmt.Println("🔓 Unlocked crafting: potions")
		}
		if village.Level >= 5 && !Contains(village.UnlockedCrafting, "armor") {
			village.UnlockedCrafting = append(village.UnlockedCrafting, "armor")
			fmt.Println("🔓 Unlocked crafting: armor")
		}
		if village.Level >= 7 && !Contains(village.UnlockedCrafting, "weapons") {
			village.UnlockedCrafting = append(village.UnlockedCrafting, "weapons")
			fmt.Println("🔓 Unlocked crafting: weapons")
		}
		if village.Level >= 10 {
			if !Contains(village.UnlockedCrafting, "skill_upgrades") {
				village.UnlockedCrafting = append(village.UnlockedCrafting, "skill_upgrades")
				fmt.Println("🔓 Unlocked crafting: skill_upgrades")
			}
			if !Contains(village.UnlockedCrafting, "skill_scrolls") {
				village.UnlockedCrafting = append(village.UnlockedCrafting, "skill_scrolls")
				fmt.Println("🔓 Unlocked crafting: skill_scrolls")
			}
		}
	}
}

func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// UnlockBaseLocationCapability grants a village crafting unlock when a Base
// location is discovered. Returns a description message, or "" if no unlock.
func UnlockBaseLocationCapability(village *models.Village, locationName string) string {
	capMap := map[string]struct {
		craft string
		desc  string
	}{
		"Stone Keep":    {"fortifications", "Village defense crafting unlocked! Build stronger walls and traps."},
		"Training Hub":  {"training", "Villager training unlocked! Boost villager efficiency and level."},
		"Hospital":      {"healing", "Healing services unlocked! Restore HP/MP/SP at the village."},
	}

	entry, ok := capMap[locationName]
	if !ok {
		return ""
	}
	if Contains(village.UnlockedCrafting, entry.craft) {
		return ""
	}
	village.UnlockedCrafting = append(village.UnlockedCrafting, entry.craft)
	return entry.desc
}

func CountVillagersByRole(village *models.Village, role string) int {
	count := 0
	for _, villager := range village.Villagers {
		if villager.Role == role {
			count++
		}
	}
	return count
}

// AutoTideResult holds the outcome of a non-interactive auto-tide.
type AutoTideResult struct {
	Victory           bool
	WavesProcessed    int
	MonstersKilled    int
	DamageDealt       int
	DamageTaken       int
	XPReward          int
	BonusGold         int
	ResourcesLost     int
	GuardsLost        int
	VillagersLost     int
	DefensesDestroyed int
	Messages          []string
}

// ProcessAutoTide runs a full non-interactive monster tide against a village.
// All waves are resolved in a single call without player input.
func ProcessAutoTide(village *models.Village, player *models.Character) AutoTideResult {
	result := AutoTideResult{}

	level := village.Level
	// Scale gently: 1 wave at level 1-4, 2 at level 5-9, etc.
	totalWaves := 1 + level/5
	// 2 monsters at level 1, scales up slowly
	monstersPerWave := 2 + level/2

	defenseLevel := village.DefenseLevel

	// Tide header
	result.Messages = append(result.Messages,
		fmt.Sprintf("--- Monster Tide on %s (Lv%d) ---", village.Name, level))

	guardCount := 0
	for _, g := range village.ActiveGuards {
		if g.HitpointsRemaining > 0 {
			guardCount++
		}
	}
	trapCount := 0
	for _, t := range village.Traps {
		if t.Remaining > 0 {
			trapCount++
		}
	}
	defenseCount := 0
	for _, d := range village.Defenses {
		if d.Built {
			defenseCount++
		}
	}
	result.Messages = append(result.Messages,
		fmt.Sprintf("Defenders: %d guards, %d traps, %d defenses", guardCount, trapCount, defenseCount))

	monsterTypes := []string{"Goblin", "Orc", "Kobold", "Slime", "Skeleton", "Wolf", "Bandit"}

	for wave := 1; wave <= totalWaves; wave++ {
		result.WavesProcessed++
		result.Messages = append(result.Messages,
			fmt.Sprintf("-- Wave %d/%d: %d monsters charge! --", wave, totalWaves, monstersPerWave))

		waveKills := 0
		waveBreaches := 0

		for m := 0; m < monstersPerWave; m++ {
			monsterHP := 8 + level*3 + rand.Intn(level*2+1)
			monsterAtk := 2 + level + rand.Intn(level+1)
			monsterName := monsterTypes[rand.Intn(len(monsterTypes))]
			monsterLabel := fmt.Sprintf("Lv%d %s", level+rand.Intn(3), monsterName)

			// Phase 1: Traps
			for i := range village.Traps {
				if village.Traps[i].Remaining > 0 && rand.Intn(100) < village.Traps[i].TriggerRate {
					dmg := village.Traps[i].Damage
					monsterHP -= dmg
					result.DamageDealt += dmg
					village.Traps[i].Remaining--
					if monsterHP <= 0 {
						result.Messages = append(result.Messages,
							fmt.Sprintf("  %s triggers %s for %d dmg - killed!", monsterLabel, village.Traps[i].Name, dmg))
					} else {
						result.Messages = append(result.Messages,
							fmt.Sprintf("  %s triggers %s for %d dmg (%d HP left)", monsterLabel, village.Traps[i].Name, dmg, monsterHP))
					}
				}
				if monsterHP <= 0 {
					break
				}
			}
			if monsterHP <= 0 {
				result.MonstersKilled++
				waveKills++
				continue
			}

			// Phase 2: Towers/Defenses
			for _, d := range village.Defenses {
				if d.Built && d.AttackPower > 0 {
					dmg := d.AttackPower + rand.Intn(d.Level*2+1)
					monsterHP -= dmg
					result.DamageDealt += dmg
					if monsterHP <= 0 {
						result.Messages = append(result.Messages,
							fmt.Sprintf("  %s hit by %s for %d dmg - killed!", monsterLabel, d.Name, dmg))
					} else {
						result.Messages = append(result.Messages,
							fmt.Sprintf("  %s hit by %s for %d dmg (%d HP left)", monsterLabel, d.Name, dmg, monsterHP))
					}
				}
				if monsterHP <= 0 {
					break
				}
			}
			if monsterHP <= 0 {
				result.MonstersKilled++
				waveKills++
				continue
			}

			// Phase 3: Guards
			guardEngaged := false
			for i := range village.ActiveGuards {
				if village.ActiveGuards[i].HitpointsRemaining <= 0 {
					continue
				}
				guardEngaged = true
				guard := &village.ActiveGuards[i]
				guardAtk := guard.AttackRolls*3 + guard.AttackBonus + guard.StatsMod.AttackMod
				guardDef := guard.DefenseRolls*2 + guard.DefenseBonus + guard.StatsMod.DefenseMod
				// Guard attacks monster
				dmg := guardAtk + rand.Intn(guardAtk/2+1)
				monsterHP -= dmg
				result.DamageDealt += dmg
				if monsterHP <= 0 {
					result.MonstersKilled++
					waveKills++
					result.Messages = append(result.Messages,
						fmt.Sprintf("  %s engages %s - deals %d dmg - killed!", guard.Name, monsterLabel, dmg))
					break
				}
				// Monster attacks guard
				mDmg := monsterAtk - guardDef
				if mDmg < 1 {
					mDmg = 1
				}
				guard.HitpointsRemaining -= mDmg
				result.DamageTaken += mDmg
				if guard.HitpointsRemaining <= 0 {
					guard.Injured = true
					result.Messages = append(result.Messages,
						fmt.Sprintf("  %s fights %s (%d dmg) but takes %d dmg and falls!", guard.Name, monsterLabel, dmg, mDmg))
				} else {
					result.Messages = append(result.Messages,
						fmt.Sprintf("  %s fights %s (%d dmg) and takes %d dmg (%d/%d HP)", guard.Name, monsterLabel, dmg, mDmg, guard.HitpointsRemaining, guard.HitPoints))
				}
				break // one guard engages per monster
			}
			if monsterHP <= 0 {
				continue
			}

			// Phase 4: Breach — monster got through
			breachDmg := monsterAtk
			result.DamageTaken += breachDmg
			waveBreaches++
			if !guardEngaged {
				result.Messages = append(result.Messages,
					fmt.Sprintf("  %s breaches undefended village! (%d breach dmg)", monsterLabel, breachDmg))
			} else {
				result.Messages = append(result.Messages,
					fmt.Sprintf("  %s breaks through defenses! (%d breach dmg)", monsterLabel, breachDmg))
			}
		}

		// Wave summary
		result.Messages = append(result.Messages,
			fmt.Sprintf("  Wave %d result: %d killed, %d breached", wave, waveKills, waveBreaches))

		// Decrement trap durability at end of wave
		for i := range village.Traps {
			if village.Traps[i].Remaining > 0 {
				village.Traps[i].Remaining--
			}
		}
	}

	// Determine victory/defeat
	// Base threshold of 30 means a level 1 village can absorb ~30 breach damage
	// before losing, which is enough to survive 1 wave of 2 weak monsters.
	defenseThreshold := 30 + defenseLevel*50

	result.Messages = append(result.Messages,
		fmt.Sprintf("-- Battle Over -- Total damage dealt: %d | Breach damage taken: %d/%d threshold",
			result.DamageDealt, result.DamageTaken, defenseThreshold))

	if result.DamageTaken < defenseThreshold {
		// Victory
		result.Victory = true
		result.XPReward = level*20 + result.MonstersKilled*5
		if result.DamageTaken < defenseThreshold/2 {
			result.BonusGold = level * 10
			if res, ok := player.ResourceStorageMap["Gold"]; ok {
				res.Stock += result.BonusGold
				player.ResourceStorageMap["Gold"] = res
			} else {
				player.ResourceStorageMap["Gold"] = models.Resource{Name: "Gold", Stock: result.BonusGold}
			}
		}
		village.Experience += result.XPReward
		result.Messages = append(result.Messages,
			fmt.Sprintf("VICTORY! %s held the line. %d/%d monsters slain. +%d XP",
				village.Name, result.MonstersKilled, result.WavesProcessed*monstersPerWave, result.XPReward))
		if result.BonusGold > 0 {
			result.Messages = append(result.Messages,
				fmt.Sprintf("Strong defense bonus: +%d Gold", result.BonusGold))
		}

		// Report surviving guard status
		injured := 0
		for _, g := range village.ActiveGuards {
			if g.HitpointsRemaining > 0 && g.HitpointsRemaining < g.HitPoints {
				injured++
			}
		}
		fallen := 0
		for _, g := range village.ActiveGuards {
			if g.HitpointsRemaining <= 0 {
				fallen++
			}
		}
		if fallen > 0 || injured > 0 {
			result.Messages = append(result.Messages,
				fmt.Sprintf("Guard casualties: %d fallen, %d injured", fallen, injured))
		}
	} else {
		// Defeat — total village destruction
		result.Victory = false
		previousLevel := village.Level

		// Count losses for reporting
		result.GuardsLost = len(village.ActiveGuards)
		result.VillagersLost = len(village.Villagers)
		result.DefensesDestroyed = len(village.Defenses)

		// Steal resources
		resourceTypes := []string{"Lumber", "Gold", "Iron", "Sand", "Stone"}
		stolen := level * 5
		for _, rt := range resourceTypes {
			if res, ok := player.ResourceStorageMap[rt]; ok && res.Stock > 0 {
				loss := stolen
				if loss > res.Stock {
					loss = res.Stock
				}
				res.Stock -= loss
				player.ResourceStorageMap[rt] = res
				result.ResourcesLost += loss
			}
		}

		// Total reset: village razed to level 1
		village.Level = 1
		village.Experience = 0
		village.DefenseLevel = 1
		village.Villagers = []models.Villager{}
		village.ActiveGuards = []models.Guard{}
		village.Defenses = []models.Defense{}
		village.Traps = []models.Trap{}

		result.Messages = append(result.Messages,
			fmt.Sprintf("DEFEAT! %s has been razed! The monsters overwhelmed all defenses.", village.Name))
		if previousLevel > 1 {
			result.Messages = append(result.Messages,
				fmt.Sprintf("Village level reset from %d to 1. All progress lost.", previousLevel))
		}
		if result.GuardsLost > 0 {
			result.Messages = append(result.Messages,
				fmt.Sprintf("All %d guards have perished.", result.GuardsLost))
		}
		if result.VillagersLost > 0 {
			result.Messages = append(result.Messages,
				fmt.Sprintf("All %d villagers have been lost.", result.VillagersLost))
		}
		if result.DefensesDestroyed > 0 {
			result.Messages = append(result.Messages,
				fmt.Sprintf("All %d defenses and structures destroyed.", result.DefensesDestroyed))
		}
		if result.ResourcesLost > 0 {
			result.Messages = append(result.Messages,
				fmt.Sprintf("%d resources looted by the invaders.", result.ResourcesLost))
		}
		result.Messages = append(result.Messages,
			"The village must be rebuilt from scratch.")
	}

	// Clean up dead guards (only matters on victory path)
	aliveGuards := []models.Guard{}
	for _, g := range village.ActiveGuards {
		if g.HitpointsRemaining > 0 {
			aliveGuards = append(aliveGuards, g)
		}
	}
	village.ActiveGuards = aliveGuards

	// Remove spent traps (only matters on victory path)
	activeTraps := []models.Trap{}
	for _, t := range village.Traps {
		if t.Remaining > 0 {
			activeTraps = append(activeTraps, t)
		}
	}
	village.Traps = activeTraps

	village.LastTideTime = time.Now().Unix()
	UpgradeVillage(village)

	return result
}

// ProcessVillageManagerTick performs automated village upkeep: assigning idle
// harvesters, hiring guards, building defenses/traps, recovering guards, and
// upgrading the village. Returns a list of action messages (empty if nothing done).
func ProcessVillageManagerTick(village *models.Village, player *models.Character) []string {
	var messages []string
	resourceTypes := []string{"Lumber", "Gold", "Iron", "Sand", "Stone"}

	// 1. Assign idle harvesters
	for i := range village.Villagers {
		v := &village.Villagers[i]
		if v.Role == "harvester" && v.HarvestType == "" {
			v.HarvestType = resourceTypes[rand.Intn(len(resourceTypes))]
			v.AssignedTask = "harvesting"
			village.Experience += 10
			messages = append(messages, fmt.Sprintf("%s assigned to harvest %s", v.Name, v.HarvestType))
		}
	}

	// 2. Hire a guard (one per tick)
	if len(village.ActiveGuards) < village.Level {
		cost := 50 + village.Level*25
		if goldRes, ok := player.ResourceStorageMap["Gold"]; ok && goldRes.Stock >= cost {
			guard := GenerateGuard(village.Level)
			guard.Hired = true
			village.ActiveGuards = append(village.ActiveGuards, guard)
			goldRes.Stock -= cost
			player.ResourceStorageMap["Gold"] = goldRes
			village.Experience += 50
			messages = append(messages, fmt.Sprintf("Hired guard %s for %d Gold", guard.Name, cost))
		}
	}

	// 3. Build a Wooden Wall (one per tick)
	if len(village.Defenses) < village.Level {
		lumberRes, hasLumber := player.ResourceStorageMap["Lumber"]
		stoneRes, hasStone := player.ResourceStorageMap["Stone"]
		if hasLumber && hasStone && lumberRes.Stock >= 50 && stoneRes.Stock >= 20 {
			wall := models.Defense{
				Name:    "Wooden Wall",
				Level:   1,
				Defense: 10,
				Built:   true,
				Type:    "wall",
			}
			village.Defenses = append(village.Defenses, wall)
			village.DefenseLevel++
			lumberRes.Stock -= 50
			stoneRes.Stock -= 20
			player.ResourceStorageMap["Lumber"] = lumberRes
			player.ResourceStorageMap["Stone"] = stoneRes
			village.Experience += 30
			messages = append(messages, "Built a Wooden Wall")
		}
	}

	// 4. Build a Spike Trap (one per tick)
	maxTraps := 2 + village.Level/3
	if len(village.Traps) < maxTraps {
		ironRes, hasIron := player.ResourceStorageMap["Iron"]
		boneRes, hasBone := player.ResourceStorageMap["Beast Bone"]
		if hasIron && hasBone && ironRes.Stock >= 10 && boneRes.Stock >= 5 {
			trap := models.Trap{
				Name:        "Spike Trap",
				Type:        "spike",
				Damage:      15,
				Duration:    3,
				Remaining:   3,
				TriggerRate: 60,
			}
			village.Traps = append(village.Traps, trap)
			ironRes.Stock -= 10
			boneRes.Stock -= 5
			player.ResourceStorageMap["Iron"] = ironRes
			player.ResourceStorageMap["Beast Bone"] = boneRes
			village.Experience += 35
			messages = append(messages, "Built a Spike Trap")
		}
	}

	// 5. Recover injured guards
	ProcessGuardRecovery(village)

	// 6. Upgrade village if enough XP
	UpgradeVillage(village)

	return messages
}
