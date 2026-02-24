package game

import (
	"fmt"

	"rpg-game/pkg/models"
)

// ProcessStatusEffects iterates through a character's active status effects,
// applies per-turn damage/healing, decrements durations, and removes expired effects.
func ProcessStatusEffects(character *models.Character) {
	for i := len(character.StatusEffects) - 1; i >= 0; i-- {
		effect := &character.StatusEffects[i]

		switch effect.Type {
		case "poison":
			character.HitpointsRemaining -= effect.Potency
			fmt.Printf("%s takes %d poison damage!\n", character.Name, effect.Potency)
		case "burn":
			character.HitpointsRemaining -= effect.Potency
			fmt.Printf("%s takes %d burn damage!\n", character.Name, effect.Potency)
		case "regen":
			character.HitpointsRemaining += effect.Potency
			if character.HitpointsRemaining > character.HitpointsTotal {
				character.HitpointsRemaining = character.HitpointsTotal
			}
			fmt.Printf("%s regenerates %d HP!\n", character.Name, effect.Potency)
		case "buff_attack":
			// already applied when effect was first added
		case "buff_defense":
			// already applied when effect was first added
		}

		effect.Duration--

		if effect.Duration <= 0 {
			switch effect.Type {
			case "buff_attack":
				character.StatsMod.AttackMod -= effect.Potency
			case "buff_defense":
				character.StatsMod.DefenseMod -= effect.Potency
			}
			fmt.Printf("%s's %s effect has worn off.\n", character.Name, effect.Type)
			character.StatusEffects = append(character.StatusEffects[:i], character.StatusEffects[i+1:]...)
		}
	}
}

// ProcessStatusEffectsMob iterates through a monster's active status effects,
// applies per-turn damage/healing, decrements durations, and removes expired effects.
func ProcessStatusEffectsMob(mob *models.Monster) {
	for i := len(mob.StatusEffects) - 1; i >= 0; i-- {
		effect := &mob.StatusEffects[i]

		switch effect.Type {
		case "poison":
			mob.HitpointsRemaining -= effect.Potency
			fmt.Printf("%s takes %d poison damage!\n", mob.Name, effect.Potency)
		case "burn":
			mob.HitpointsRemaining -= effect.Potency
			fmt.Printf("%s takes %d burn damage!\n", mob.Name, effect.Potency)
		case "regen":
			mob.HitpointsRemaining += effect.Potency
			if mob.HitpointsRemaining > mob.HitpointsTotal {
				mob.HitpointsRemaining = mob.HitpointsTotal
			}
			fmt.Printf("%s regenerates %d HP!\n", mob.Name, effect.Potency)
		case "buff_attack":
			// already applied when effect was first added
		case "buff_defense":
			// already applied when effect was first added
		}

		effect.Duration--

		if effect.Duration <= 0 {
			switch effect.Type {
			case "buff_attack":
				mob.StatsMod.AttackMod -= effect.Potency
			case "buff_defense":
				mob.StatsMod.DefenseMod -= effect.Potency
			}
			fmt.Printf("%s's %s effect has worn off.\n", mob.Name, effect.Type)
			mob.StatusEffects = append(mob.StatusEffects[:i], mob.StatusEffects[i+1:]...)
		}
	}
}

// ApplyDamage calculates final damage after applying elemental resistance modifiers.
// The target can be either a *models.Character or a *models.Monster.
func ApplyDamage(damage int, damageType models.DamageType, target interface{}) int {
	resistance := 1.0

	switch t := target.(type) {
	case *models.Character:
		if res, ok := t.Resistances[damageType]; ok {
			resistance = res
		}
	case *models.Monster:
		if res, ok := t.Resistances[damageType]; ok {
			resistance = res
		}
	}

	return int(float64(damage) * resistance)
}

// IsStunned checks whether a character has an active stun status effect.
func IsStunned(character *models.Character) bool {
	for _, effect := range character.StatusEffects {
		if effect.Type == "stun" {
			return true
		}
	}
	return false
}

// IsStunnedMob checks whether a monster has an active stun status effect.
func IsStunnedMob(mob *models.Monster) bool {
	for _, effect := range mob.StatusEffects {
		if effect.Type == "stun" {
			return true
		}
	}
	return false
}
