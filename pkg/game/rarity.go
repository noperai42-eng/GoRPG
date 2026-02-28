package game

import (
	"math/rand"

	"rpg-game/pkg/models"
)

// RarityMultipliers defines stat scaling per rarity tier.
type RarityMultipliers struct {
	HPMult      float64
	AtkRolls    int
	DefRolls    int
	XPMult      float64
	LootBonus   int
}

var rarityMultipliers = map[models.MonsterRarity]RarityMultipliers{
	models.RarityCommon:    {1.0, 1, 1, 1.0, 0},
	models.RarityUncommon:  {10.0, 2, 2, 5.0, 2},
	models.RarityRare:      {50.0, 4, 3, 25.0, 4},
	models.RarityEpic:      {250.0, 6, 5, 100.0, 6},
	models.RarityLegendary: {1000.0, 10, 8, 500.0, 8},
	models.RarityMythic:    {3000.0, 15, 12, 1500.0, 10},
}

// rarityWeights defines the base % chance for each rarity.
// [common, uncommon, rare, epic, legendary, mythic]
var baseRarityWeights = [6]int{60, 25, 10, 4, 1, 0}

var rarityOrder = [6]models.MonsterRarity{
	models.RarityCommon,
	models.RarityUncommon,
	models.RarityRare,
	models.RarityEpic,
	models.RarityLegendary,
	models.RarityMythic,
}

// RollRarity returns a random rarity tier. Higher rankMax shifts weights
// toward rarer tiers.
func RollRarity(rankMax int) models.MonsterRarity {
	weights := [6]int{}
	copy(weights[:], baseRarityWeights[:])

	// Shift weights based on rankMax: each point above 1 adds to rarer tiers
	shift := rankMax - 1
	if shift > 0 {
		weights[0] -= shift * 5 // reduce common
		if weights[0] < 20 {
			weights[0] = 20
		}
		weights[1] += shift * 2
		weights[2] += shift * 2
		weights[3] += shift
		weights[4] += shift / 2
		// Mythic only appears via evolution/player-kill upgrades, not random rolls
	}

	total := 0
	for _, w := range weights {
		total += w
	}

	roll := rand.Intn(total)
	cumulative := 0
	for i, w := range weights {
		cumulative += w
		if roll < cumulative {
			return rarityOrder[i]
		}
	}
	return models.RarityCommon
}

// ApplyRarity multiplies a monster's stats by its rarity tier.
func ApplyRarity(mob *models.Monster) {
	r := NormalizeRarity(mob.Rarity)
	mult, ok := rarityMultipliers[r]
	if !ok || r == models.RarityCommon {
		return
	}

	mob.HitpointsNatural = int(float64(mob.HitpointsNatural) * mult.HPMult)
	mob.HitpointsTotal = mob.HitpointsNatural
	mob.HitpointsRemaining = mob.HitpointsTotal

	mob.AttackRolls *= mult.AtkRolls
	mob.DefenseRolls *= mult.DefRolls

	// Scale mana and stamina proportionally
	mob.ManaNatural = int(float64(mob.ManaNatural) * mult.HPMult / 10)
	if mob.ManaNatural < mob.ManaTotal {
		mob.ManaNatural = mob.ManaTotal
	}
	mob.ManaTotal = mob.ManaNatural
	mob.ManaRemaining = mob.ManaTotal

	mob.StaminaNatural = int(float64(mob.StaminaNatural) * mult.HPMult / 10)
	if mob.StaminaNatural < mob.StaminaTotal {
		mob.StaminaNatural = mob.StaminaTotal
	}
	mob.StaminaTotal = mob.StaminaNatural
	mob.StaminaRemaining = mob.StaminaTotal
}

// RarityXPMult returns the XP multiplier for a given rarity.
func RarityXPMult(r models.MonsterRarity) float64 {
	r = NormalizeRarity(r)
	if mult, ok := rarityMultipliers[r]; ok {
		return mult.XPMult
	}
	return 1.0
}

// RarityLootBonus returns the extra loot rarity bonus for a given monster rarity.
func RarityLootBonus(r models.MonsterRarity) int {
	r = NormalizeRarity(r)
	if mult, ok := rarityMultipliers[r]; ok {
		return mult.LootBonus
	}
	return 0
}

// NormalizeRarity maps empty string to Common for backward compatibility.
func NormalizeRarity(r models.MonsterRarity) models.MonsterRarity {
	if r == "" {
		return models.RarityCommon
	}
	return r
}

// RarityDisplayName returns a display-friendly name for a rarity tier.
func RarityDisplayName(r models.MonsterRarity) string {
	switch NormalizeRarity(r) {
	case models.RarityCommon:
		return "Common"
	case models.RarityUncommon:
		return "Uncommon"
	case models.RarityRare:
		return "Rare"
	case models.RarityEpic:
		return "Epic"
	case models.RarityLegendary:
		return "Legendary"
	case models.RarityMythic:
		return "Mythic"
	default:
		return "Common"
	}
}

// NextRarity returns the rarity one tier above the given rarity, or "" if already max.
func NextRarity(r models.MonsterRarity) models.MonsterRarity {
	r = NormalizeRarity(r)
	for i, tier := range rarityOrder {
		if tier == r && i+1 < len(rarityOrder) {
			return rarityOrder[i+1]
		}
	}
	return ""
}

// RarityIndex returns the numeric index of a rarity (0=common, 5=mythic).
func RarityIndex(r models.MonsterRarity) int {
	r = NormalizeRarity(r)
	for i, tier := range rarityOrder {
		if tier == r {
			return i
		}
	}
	return 0
}
