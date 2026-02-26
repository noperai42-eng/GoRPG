package game

import (
	"fmt"
	"math/rand"

	"rpg-game/pkg/data"
	"rpg-game/pkg/models"
)

func GenerateCharacter(name string, level int, rank int) models.Character {
	hitpoints := MultiRoll(rank)
	mana := MultiRoll(rank) + 20
	stamina := MultiRoll(rank) + 20

	// Start with only 1 basic skill - others must be earned from Skill Guardians
	learnedSkills := []models.Skill{
		data.AvailableSkills[4], // Power Strike - basic physical attack
	}

	resistances := map[models.DamageType]float64{
		models.Physical:  1.0,
		models.Fire:      1.0,
		models.Ice:       1.0,
		models.Lightning: 1.0,
		models.Poison:    1.0,
	}

	return models.Character{
		Name:               name,
		Level:              level,
		Experience:         0,
		HitpointsTotal:     hitpoints,
		HitpointsNatural:   hitpoints,
		HitpointsRemaining: hitpoints,
		ManaTotal:          mana,
		ManaRemaining:      mana,
		ManaNatural:        mana,
		StaminaTotal:       stamina,
		StaminaRemaining:   stamina,
		StaminaNatural:     stamina,
		AttackRolls:        rank,
		DefenseRolls:       rank,
		LearnedSkills:      learnedSkills,
		StatusEffects:      []models.StatusEffect{},
		Resistances:        resistances,
		CompletedQuests:    []string{},
		ActiveQuests:       []string{"quest_1_training"},
		VillageName:        name + "'s Village",
	}
}

// PlayerExpToLevel returns the total XP needed to reach the next level.
// Scaled so ~30 even-level kills at low levels, increasing with level.
func PlayerExpToLevel(level int) int {
	return level * (300 + level*10)
}

func LevelUp(player *models.Character) {
	for player.Experience >= PlayerExpToLevel(player.Level) {
		player.Level++
		player.HitpointsNatural += MultiRoll(1)
		player.HitpointsRemaining = player.HitpointsNatural
		player.ManaNatural += MultiRoll(1) + 5
		player.ManaTotal = player.ManaNatural
		player.ManaRemaining = player.ManaTotal
		player.StaminaNatural += MultiRoll(1) + 5
		player.StaminaTotal = player.StaminaNatural
		player.StaminaRemaining = player.StaminaTotal
		player.AttackRolls = player.Level/10 + 1
		player.DefenseRolls = player.Level/10 + 1
		player.StatsMod = CalculateItemMods(player.EquipmentMap)
		player.HitpointsTotal = player.HitpointsNatural + player.StatsMod.HitPointMod
		fmt.Printf("LEVEL UP!!! Now level %d!\n", player.Level)
		fmt.Printf("HP: %d, MP: %d, SP: %d\n", player.HitpointsTotal, player.ManaTotal, player.StaminaTotal)

		// Skills are now learned from defeating Skill Guardians, not automatic
	}
}

func GenerateLocationsForNewCharacter(char *models.Character) {
	GenerateMissingResourceType(char)
	char.KnownLocations = []string{"Home", "Training Hall", "Forest", "Lake", "Hills"}
}

func GenerateMissingResourceType(char *models.Character) {
	for resource := range data.ResourceTypes {
		_, exists := char.ResourceStorageMap[data.ResourceTypes[resource]]
		if !exists {
			char.ResourceStorageMap[data.ResourceTypes[resource]] = models.Resource{Name: data.ResourceTypes[resource], Stock: 0, RollModifier: 0}
		}
	}
}

// Suppress unused import warnings
var _ = rand.Intn
