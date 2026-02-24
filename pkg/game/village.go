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

func ProcessVillageResourceCollection(village *models.Village, player *models.Character) {
	collected := false
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
			fmt.Printf("  %s collected %d %s\n", villager.Name, amount, villager.HarvestType)
			collected = true
		}
	}
	if !collected {
		fmt.Println("  No villagers are currently harvesting resources.")
	}
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

func CountVillagersByRole(village *models.Village, role string) int {
	count := 0
	for _, villager := range village.Villagers {
		if villager.Role == role {
			count++
		}
	}
	return count
}
