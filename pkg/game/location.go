package game

import (
	"fmt"
	"math/rand"

	"rpg-game/pkg/data"
	"rpg-game/pkg/models"
)

func GenerateGameLocation(game *models.GameState) {
	game.GameLocations = map[string]models.Location{}
	for _, locationValue := range data.DiscoverableLocations {
		GenerateMonstersForLocation(&locationValue, game)
		game.GameLocations[locationValue.Name] = locationValue
	}
}

func GenerateMonstersForLocation(location *models.Location, game *models.GameState) {
	if location.Type == "Base" {
		return
	}

	location.Monsters = make([]models.Monster, 20)
	for i := 0; i < 20; i++ {
		location.Monsters[i] = GenerateBestMonster(game, location.LevelMax, location.RarityMax)
		location.Monsters[i].LocationName = location.Name
	}

	// Add Skill Guardians based on location level
	// Higher level locations have more guardians
	numGuardians := 0
	if location.LevelMax >= 10 && location.LevelMax < 30 {
		numGuardians = 1 // Low level areas: 1 guardian
	} else if location.LevelMax >= 30 && location.LevelMax < 100 {
		numGuardians = 2 // Mid level areas: 2 guardians
	} else if location.LevelMax >= 100 {
		numGuardians = 3 // High level areas: 3 guardians
	}

	// Spawn guardians if location is suitable
	if numGuardians > 0 && len(data.AvailableSkills) > 0 {
		// Create a list of available skills (excluding Tracking and Power Strike)
		guardableSkills := []models.Skill{}
		for _, skill := range data.AvailableSkills {
			if skill.Name != "Tracking" && skill.Name != "Power Strike" {
				guardableSkills = append(guardableSkills, skill)
			}
		}

		// Randomly place guardians
		for g := 0; g < numGuardians && len(guardableSkills) > 0; g++ {
			// Pick random position
			guardianPos := rand.Intn(len(location.Monsters))

			// Pick random skill from guardable skills
			skillIndex := rand.Intn(len(guardableSkills))
			guardianSkill := guardableSkills[skillIndex]

			// Guardians spawn above location level — they're elite fights
			guardianLevel := location.LevelMax + rand.Intn(location.LevelMax/2+1) + 3

			guardian := GenerateSkillGuardian(guardianSkill, guardianLevel, location.RarityMax)
			guardian.LocationName = location.Name
			location.Monsters[guardianPos] = guardian

			fmt.Printf("Spawned %s (Lv%d) guarding %s at %s\n",
				guardian.Name, guardian.Level, guardianSkill.Name, location.Name)
		}
	}
}

func SearchLocation(discoveredLocations []string, undiscoveredLocations []models.Location) string {
	// Filter candidates: skip Weight <= 0 and already-discovered locations
	candidates := []models.Location{}
	for _, location := range undiscoveredLocations {
		if location.Weight <= 0 {
			continue
		}
		alreadyDiscovered := false
		for _, discovered := range discoveredLocations {
			if discovered == location.Name {
				alreadyDiscovered = true
				break
			}
		}
		if !alreadyDiscovered {
			candidates = append(candidates, location)
		}
	}

	if len(candidates) == 0 {
		fmt.Println("No new locations to discover!")
		return ""
	}

	totalWeight := CalculateTotalWeight(candidates)
	randomNum := rand.Intn(totalWeight)
	fmt.Printf("TotalWeight: %d\n", totalWeight)
	fmt.Printf("RandNum: %d\n", randomNum)

	cumulative := 0
	for _, location := range candidates {
		cumulative += location.Weight
		if randomNum < cumulative {
			fmt.Printf("Discovered location: %s\n", location.Name)
			return location.Name
		}
	}

	return ""
}

func GenerateLocationGuardian(locationName string, loc models.Location, gs *models.GameState) models.Monster {
	levelMax := loc.LevelMax
	if levelMax == 0 {
		levelMax = 30
	}
	rarityMax := loc.RarityMax
	if rarityMax == 0 {
		rarityMax = 3
	}
	guardian := GenerateBestMonster(gs, levelMax, rarityMax)
	guardian.IsBoss = true
	guardian.Name = "Guardian of " + locationName
	guardian.HitpointsNatural = int(float64(guardian.HitpointsNatural) * 1.5)
	guardian.HitpointsTotal = int(float64(guardian.HitpointsTotal) * 1.5)
	guardian.HitpointsRemaining = guardian.HitpointsTotal
	return guardian
}

func CalculateTotalWeight(locations []models.Location) int {
	total := 0
	for _, location := range locations {
		total += location.Weight
	}
	return total
}

func RemoveLocation(locations []models.Location, index int) []models.Location {
	return append(locations[:index], locations[index+1:]...)
}
