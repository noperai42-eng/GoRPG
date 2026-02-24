package game

import (
	"math/rand"
	"strings"

	"rpg-game/pkg/models"
)

// MakeAIDecision determines the best combat action for a player character
// based on current health, turn count, and available resources.
// Returns a string representing the chosen action.
func MakeAIDecision(player *models.Character, mob *models.Monster, turnCount int) string {
	hpPercent := float64(player.HitpointsRemaining) / float64(player.HitpointsTotal)

	// Priority 1: Heal if HP < 40%
	if hpPercent < 0.4 {
		// Check for Heal skill
		for _, skill := range player.LearnedSkills {
			if strings.EqualFold(skill.Name, "Heal") && player.ManaRemaining >= skill.ManaCost {
				return "skill_heal"
			}
		}
		// Check for Regeneration skill
		for _, skill := range player.LearnedSkills {
			if strings.EqualFold(skill.Name, "Regeneration") && player.ManaRemaining >= skill.ManaCost {
				return "skill_regeneration"
			}
		}
		// Check for any consumable in inventory
		for _, item := range player.Inventory {
			if item.ItemType == "consumable" {
				return "item"
			}
		}
	}

	// Priority 2: Use buff skills at the start of combat (turns 1-2)
	if turnCount <= 2 {
		for _, skill := range player.LearnedSkills {
			if (strings.EqualFold(skill.Name, "Battle Cry") || strings.EqualFold(skill.Name, "Shield Wall")) &&
				player.StaminaRemaining >= skill.StaminaCost {
				return "skill_" + skill.Name
			}
		}
	}

	// Priority 3: Use offensive skills if resources available (50% chance)
	if rand.Intn(100) < 50 {
		for _, skill := range player.LearnedSkills {
			if skill.Damage > 0 &&
				player.ManaRemaining >= skill.ManaCost &&
				player.StaminaRemaining >= skill.StaminaCost {
				return "skill_" + skill.Name
			}
		}
	}

	// Default: attack normally
	return "attack"
}
