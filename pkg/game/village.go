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
