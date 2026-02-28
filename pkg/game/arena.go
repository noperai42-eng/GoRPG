package game

import (
	"math"
	"time"

	"rpg-game/pkg/models"
)

// ArenaMaxBattlesPerDay is the daily battle limit per player.
const ArenaMaxBattlesPerDay = 5

// CalculateArenaPoints computes ELO-style rating changes.
// Returns (winnerGain, loserLoss).
func CalculateArenaPoints(winnerRating, loserRating int) (int, int) {
	const kFactor = 32.0

	// Expected score for winner
	expected := 1.0 / (1.0 + math.Pow(10.0, float64(loserRating-winnerRating)/400.0))

	winnerGain := int(math.Round(kFactor * (1.0 - expected)))
	loserLoss := int(math.Round(kFactor * expected))

	// Clamp to [1, 50]
	if winnerGain < 1 {
		winnerGain = 1
	}
	if winnerGain > 50 {
		winnerGain = 50
	}
	if loserLoss < 1 {
		loserLoss = 1
	}
	if loserLoss > 50 {
		loserLoss = 50
	}

	return winnerGain, loserLoss
}

// GetArenaResetDate returns today's date in "2006-01-02" format (UTC).
func GetArenaResetDate() string {
	return time.Now().UTC().Format("2006-01-02")
}

// CharacterToArenaMonster converts a Character into a Monster for arena combat.
func CharacterToArenaMonster(char *models.Character) models.Monster {
	return models.Monster{
		Name:               char.Name,
		Level:              char.Level,
		Rank:               char.Level/3 + 1,
		HitpointsTotal:     char.HitpointsTotal,
		HitpointsNatural:   char.HitpointsNatural,
		HitpointsRemaining: char.HitpointsTotal,
		ManaTotal:          char.ManaTotal,
		ManaNatural:        char.ManaNatural,
		ManaRemaining:      char.ManaTotal,
		StaminaTotal:       char.StaminaTotal,
		StaminaNatural:     char.StaminaNatural,
		StaminaRemaining:   char.StaminaTotal,
		AttackRolls:        char.AttackRolls,
		DefenseRolls:       char.DefenseRolls,
		StatsMod:           char.StatsMod,
		EquipmentMap:       copyEquipmentMap(char.EquipmentMap),
		Inventory:          []models.Item{},
		LearnedSkills:      append([]models.Skill{}, char.LearnedSkills...),
		StatusEffects:      []models.StatusEffect{},
		Resistances:        copyResistances(char.Resistances),
		MonsterType:        "humanoid",
	}
}
