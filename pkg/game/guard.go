package game

import (
	"fmt"
	"math/rand"

	"rpg-game/pkg/data"
	"rpg-game/pkg/models"
)

// GenerateGuard creates a new guard with stats scaled to the given level,
// starting equipment, and default resistances.
func GenerateGuard(level int) models.Guard {
	name := data.GuardNames[rand.Intn(len(data.GuardNames))]

	baseHP := 20 + (level * 5)
	attackRolls := (level / 5) + 1
	defenseRolls := (level / 5) + 1

	resistances := map[models.DamageType]float64{
		models.Physical:  1.0,
		models.Fire:      1.0,
		models.Ice:       1.0,
		models.Lightning: 1.0,
		models.Poison:    1.0,
	}

	guard := models.Guard{
		Name:               name,
		Level:              level,
		HitPoints:          baseHP,
		HitpointsNatural:   baseHP,
		HitpointsRemaining: baseHP,
		AttackBonus:        2 + level,
		DefenseBonus:       2 + level,
		AttackRolls:        attackRolls,
		DefenseRolls:       defenseRolls,
		Hired:              false,
		Cost:               50 + (level * 25),
		Inventory:          []models.Item{},
		EquipmentMap:       make(map[int]models.Item),
		StatsMod:           models.StatMod{},
		Injured:            false,
		RecoveryTime:       0,
		StatusEffects:      []models.StatusEffect{},
		Resistances:        resistances,
	}

	// Generate starting equipment
	numItems := 1 + (level / 3)
	if numItems > 3 {
		numItems = 3
	}

	for i := 0; i < numItems; i++ {
		rarity := 1 + (level / 5)
		if rarity > 5 {
			rarity = 5
		}
		item := GenerateItem(rarity)
		EquipGuardItem(item, &guard.EquipmentMap, &guard.Inventory)
	}

	// Recalculate stats from equipped items
	guard.StatsMod = CalculateItemMods(guard.EquipmentMap)
	guard.HitPoints = guard.HitpointsNatural + guard.StatsMod.HitPointMod
	guard.HitpointsRemaining = guard.HitPoints

	return guard
}

// GuardAttack processes attacks from all healthy guards against a monster,
// applying critical hits and elemental resistance. Returns total damage dealt.
func GuardAttack(guards []models.Guard, mob *models.Monster) int {
	totalDamage := 0

	for i := range guards {
		guard := &guards[i]

		if guard.Injured || guard.HitpointsRemaining <= 0 {
			continue
		}

		guardAttack := MultiRoll(guard.AttackRolls) + guard.StatsMod.AttackMod + guard.AttackBonus

		// 10% critical hit chance
		if rand.Intn(100) < 10 {
			guardAttack *= 2
			fmt.Printf("⚔️  %s lands a CRITICAL HIT!\n", guard.Name)
		} else {
			fmt.Printf("🗡️  %s attacks %s.\n", guard.Name, mob.Name)
		}

		mobDef := MultiRoll(mob.DefenseRolls) + mob.StatsMod.DefenseMod

		if guardAttack > mobDef {
			damage := guardAttack - mobDef
			finalDamage := ApplyDamage(damage, models.Physical, mob)
			mob.HitpointsRemaining -= finalDamage
			totalDamage += finalDamage
			fmt.Printf("🛡️  %s deals %d damage to %s!\n", guard.Name, finalDamage, mob.Name)
		} else {
			fmt.Printf("🛡️  %s's attack was blocked by %s!\n", guard.Name, mob.Name)
		}
	}

	return totalDamage
}

// GuardDefense distributes incoming damage among healthy guards, absorbing a
// percentage based on the number of active guards. Returns the remaining damage
// that passes through to the player and the indices of guards that absorbed damage.
func GuardDefense(guards []models.Guard, incomingDamage int) (int, []int) {
	healthyGuards := 0
	healthyIndices := []int{}

	for i := range guards {
		if !guards[i].Injured && guards[i].HitpointsRemaining > 0 {
			healthyGuards++
			healthyIndices = append(healthyIndices, i)
		}
	}

	if healthyGuards == 0 {
		return incomingDamage, nil
	}

	absorbPercent := healthyGuards * 20
	if absorbPercent > 60 {
		absorbPercent = 60
	}

	absorbedDamage := (incomingDamage * absorbPercent) / 100
	remainingDamage := incomingDamage - absorbedDamage

	// Distribute absorbed damage evenly among healthy guards
	damagePerGuard := absorbedDamage / healthyGuards
	extraDamage := absorbedDamage % healthyGuards

	damagedIndices := []int{}

	for idx, guardIndex := range healthyIndices {
		guard := &guards[guardIndex]
		dmg := damagePerGuard
		if idx < extraDamage {
			dmg++
		}

		if dmg > 0 {
			guard.HitpointsRemaining -= dmg
			damagedIndices = append(damagedIndices, guardIndex)
			fmt.Printf("🛡️  %s absorbs %d damage!\n", guard.Name, dmg)

			// Check if guard HP dropped below 30%
			hpThreshold := (guard.HitPoints * 30) / 100
			if guard.HitpointsRemaining <= hpThreshold {
				guard.Injured = true
				guard.RecoveryTime = 3
				fmt.Printf("🚑  %s has been seriously injured and needs recovery!\n", guard.Name)
			}
		}
	}

	fmt.Printf("🛡️  Guards absorbed %d of %d incoming damage!\n", absorbedDamage, incomingDamage)

	return remainingDamage, damagedIndices
}

// ProcessGuardRecovery handles recovery for injured guards in a village,
// decrementing recovery timers and restoring guards to full health when ready.
func ProcessGuardRecovery(village *models.Village) {
	for i := range village.ActiveGuards {
		guard := &village.ActiveGuards[i]

		if !guard.Injured {
			continue
		}

		guard.RecoveryTime--

		if guard.RecoveryTime <= 0 {
			guard.Injured = false
			guard.RecoveryTime = 0
			guard.HitpointsRemaining = guard.HitPoints
			fmt.Printf("💚  %s has fully recovered and is ready for duty!\n", guard.Name)
		} else {
			fmt.Printf("🏥  %s is recovering... %d turns remaining.\n", guard.Name, guard.RecoveryTime)
		}
	}
}
