package game

import "rpg-game/pkg/models"

// UpgradeSkill applies a single upgrade to a skill: +5 damage (or +5 healing),
// -2 mana cost, -2 stamina cost. This matches the village crafting upgrade logic.
func UpgradeSkill(skill *models.Skill) {
	if skill.Damage > 0 {
		skill.Damage += 5
	} else if skill.Damage < 0 {
		skill.Damage -= 5 // more negative = more healing
	}
	if skill.ManaCost > 2 {
		skill.ManaCost -= 2
	}
	if skill.StaminaCost > 2 {
		skill.StaminaCost -= 2
	}
	skill.UpgradeCount++
}
