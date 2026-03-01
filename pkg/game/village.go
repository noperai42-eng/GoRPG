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
	totalWaves := 3 + level/5
	monstersPerWave := 5 + level/3

	defenseLevel := village.DefenseLevel
	totalDefensePower := 0
	for _, d := range village.Defenses {
		if d.Built {
			totalDefensePower += d.Defense + d.AttackPower
		}
	}

	for wave := 1; wave <= totalWaves; wave++ {
		result.WavesProcessed++
		waveMsg := fmt.Sprintf("Wave %d/%d: %d monsters attack!", wave, totalWaves, monstersPerWave)
		result.Messages = append(result.Messages, waveMsg)

		for m := 0; m < monstersPerWave; m++ {
			monsterHP := 20 + level*5 + rand.Intn(level*3+1)
			monsterAtk := 5 + level*2 + rand.Intn(level+1)

			// Phase 1: Traps
			for i := range village.Traps {
				if village.Traps[i].Remaining > 0 && rand.Intn(100) < village.Traps[i].TriggerRate {
					dmg := village.Traps[i].Damage
					monsterHP -= dmg
					result.DamageDealt += dmg
					village.Traps[i].Remaining--
				}
				if monsterHP <= 0 {
					break
				}
			}
			if monsterHP <= 0 {
				result.MonstersKilled++
				continue
			}

			// Phase 2: Towers/Defenses
			for _, d := range village.Defenses {
				if d.Built && d.AttackPower > 0 {
					dmg := d.AttackPower + rand.Intn(d.Level*2+1)
					monsterHP -= dmg
					result.DamageDealt += dmg
				}
				if monsterHP <= 0 {
					break
				}
			}
			if monsterHP <= 0 {
				result.MonstersKilled++
				continue
			}

			// Phase 3: Guards
			guardKilled := false
			for i := range village.ActiveGuards {
				if village.ActiveGuards[i].HitpointsRemaining <= 0 {
					continue
				}
				guardAtk := village.ActiveGuards[i].AttackRolls*3 + village.ActiveGuards[i].AttackBonus + village.ActiveGuards[i].StatsMod.AttackMod
				guardDef := village.ActiveGuards[i].DefenseRolls*2 + village.ActiveGuards[i].DefenseBonus + village.ActiveGuards[i].StatsMod.DefenseMod
				// Guard attacks monster
				dmg := guardAtk + rand.Intn(guardAtk/2+1)
				monsterHP -= dmg
				result.DamageDealt += dmg
				if monsterHP <= 0 {
					result.MonstersKilled++
					break
				}
				// Monster attacks guard
				mDmg := monsterAtk - guardDef
				if mDmg < 1 {
					mDmg = 1
				}
				village.ActiveGuards[i].HitpointsRemaining -= mDmg
				result.DamageTaken += mDmg
				if village.ActiveGuards[i].HitpointsRemaining <= 0 {
					village.ActiveGuards[i].Injured = true
					guardKilled = true
				}
				break // one guard engages per monster
			}
			if monsterHP <= 0 {
				if !guardKilled {
					// Already counted above
				}
				continue
			}

			// Phase 4: Breach — monster got through
			breachDmg := monsterAtk
			result.DamageTaken += breachDmg
		}

		// Decrement trap durability at end of wave
		for i := range village.Traps {
			if village.Traps[i].Remaining > 0 {
				village.Traps[i].Remaining--
			}
		}
	}

	// Determine victory/defeat
	defenseThreshold := defenseLevel * 50
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
		result.Messages = append(result.Messages, fmt.Sprintf("Victory! Gained %d XP.", result.XPReward))
		if result.BonusGold > 0 {
			result.Messages = append(result.Messages, fmt.Sprintf("Bonus: +%d Gold for strong defense!", result.BonusGold))
		}
	} else {
		// Defeat — more severe than manual
		result.Victory = false

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

		// Destroy defenses
		if len(village.Defenses) > 0 {
			destroyCount := 1 + rand.Intn(len(village.Defenses)/2+1)
			destroyed := 0
			for i := range village.Defenses {
				if destroyed >= destroyCount {
					break
				}
				if village.Defenses[i].Built {
					village.Defenses[i].Built = false
					village.Defenses[i].Level = 0
					destroyed++
				}
			}
			result.DefensesDestroyed = destroyed
			if village.DefenseLevel > 1 {
				village.DefenseLevel--
			}
		}

		// Kill villagers
		if len(village.Villagers) > 0 {
			killCount := 1 + rand.Intn(len(village.Villagers)/3+1)
			if killCount > len(village.Villagers) {
				killCount = len(village.Villagers)
			}
			village.Villagers = village.Villagers[:len(village.Villagers)-killCount]
			result.VillagersLost = killCount
		}

		// Kill guards
		if len(village.ActiveGuards) > 0 {
			guardKillCount := 1 + rand.Intn(len(village.ActiveGuards)/2+1)
			if guardKillCount > len(village.ActiveGuards) {
				guardKillCount = len(village.ActiveGuards)
			}
			village.ActiveGuards = village.ActiveGuards[:len(village.ActiveGuards)-guardKillCount]
			result.GuardsLost = guardKillCount
		}

		result.Messages = append(result.Messages, "Defeat! The monsters overwhelmed your village!")
		if result.ResourcesLost > 0 {
			result.Messages = append(result.Messages, fmt.Sprintf("Lost %d resources to looting.", result.ResourcesLost))
		}
		if result.DefensesDestroyed > 0 {
			result.Messages = append(result.Messages, fmt.Sprintf("%d defenses destroyed.", result.DefensesDestroyed))
		}
		if result.VillagersLost > 0 {
			result.Messages = append(result.Messages, fmt.Sprintf("%d villagers lost.", result.VillagersLost))
		}
		if result.GuardsLost > 0 {
			result.Messages = append(result.Messages, fmt.Sprintf("%d guards lost.", result.GuardsLost))
		}
	}

	// Clean up dead guards (already injured from combat)
	aliveGuards := []models.Guard{}
	for _, g := range village.ActiveGuards {
		if g.HitpointsRemaining > 0 {
			aliveGuards = append(aliveGuards, g)
		}
	}
	village.ActiveGuards = aliveGuards

	// Remove spent traps
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
