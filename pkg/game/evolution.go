package game

import (
	"fmt"
	"math/rand"
	"sort"
	"time"

	"rpg-game/pkg/models"
)

// EvolutionEvent records something notable that happened during evolution.
type EvolutionEvent struct {
	EventType    string // "fight", "upgrade", "level_up"
	LocationName string
	Details      string
}

// MonsterVsMonsterCombat runs a simplified auto-fight between two monsters.
// Returns a pointer to the winner (a or b). Both are modified in place.
func MonsterVsMonsterCombat(a, b *models.Monster) *models.Monster {
	// Restore resources and clear status effects
	a.HitpointsRemaining = a.HitpointsTotal
	a.ManaRemaining = a.ManaTotal
	a.StaminaRemaining = a.StaminaTotal
	a.StatusEffects = []models.StatusEffect{}
	b.HitpointsRemaining = b.HitpointsTotal
	b.ManaRemaining = b.ManaTotal
	b.StaminaRemaining = b.StaminaTotal
	b.StatusEffects = []models.StatusEffect{}

	for turn := 0; turn < 200; turn++ {
		// Process status effects for both
		ProcessStatusEffectsMob(a)
		ProcessStatusEffectsMob(b)

		if a.HitpointsRemaining <= 0 {
			return b
		}
		if b.HitpointsRemaining <= 0 {
			return a
		}

		// Monster A attacks B
		if !IsStunnedMob(a) {
			monsterAttack(a, b)
		}
		if b.HitpointsRemaining <= 0 {
			return a
		}

		// Monster B attacks A
		if !IsStunnedMob(b) {
			monsterAttack(b, a)
		}
		if a.HitpointsRemaining <= 0 {
			return b
		}
	}

	// Tie-breaker: whoever has more HP% remaining wins
	aPct := float64(a.HitpointsRemaining) / float64(a.HitpointsTotal)
	bPct := float64(b.HitpointsRemaining) / float64(b.HitpointsTotal)
	if aPct >= bPct {
		return a
	}
	return b
}

// monsterAttack executes a single monster's attack on a target monster.
// 40% chance to use a skill (matching existing monster AI).
func monsterAttack(attacker, target *models.Monster) {
	// 40% chance to use a skill if available
	if rand.Intn(100) < 40 && len(attacker.LearnedSkills) > 0 {
		skill := attacker.LearnedSkills[rand.Intn(len(attacker.LearnedSkills))]
		canUse := true
		if skill.ManaCost > 0 && attacker.ManaRemaining < skill.ManaCost {
			canUse = false
		}
		if skill.StaminaCost > 0 && attacker.StaminaRemaining < skill.StaminaCost {
			canUse = false
		}
		if canUse {
			attacker.ManaRemaining -= skill.ManaCost
			attacker.StaminaRemaining -= skill.StaminaCost

			if skill.Damage > 0 {
				finalDmg := ApplyDamage(skill.Damage, skill.DamageType, target)
				target.HitpointsRemaining -= finalDmg
			} else if skill.Damage < 0 {
				// Healing skill
				attacker.HitpointsRemaining -= skill.Damage // subtracting negative = adding
				if attacker.HitpointsRemaining > attacker.HitpointsTotal {
					attacker.HitpointsRemaining = attacker.HitpointsTotal
				}
			}

			// Apply status effect
			if skill.Effect.Type != "" && skill.Effect.Type != "none" {
				effect := models.StatusEffect{
					Type:     skill.Effect.Type,
					Duration: skill.Effect.Duration,
					Potency:  skill.Effect.Potency,
				}
				if skill.Effect.Type == "buff_attack" || skill.Effect.Type == "buff_defense" ||
					skill.Effect.Type == "regen" {
					attacker.StatusEffects = append(attacker.StatusEffects, effect)
					if skill.Effect.Type == "buff_attack" {
						attacker.StatsMod.AttackMod += effect.Potency
					} else if skill.Effect.Type == "buff_defense" {
						attacker.StatsMod.DefenseMod += effect.Potency
					}
				} else {
					target.StatusEffects = append(target.StatusEffects, effect)
				}
			}
			return
		}
	}

	// Normal attack
	atkRoll := MultiRoll(attacker.AttackRolls) + attacker.StatsMod.AttackMod
	defRoll := MultiRoll(target.DefenseRolls) + target.StatsMod.DefenseMod
	damage := atkRoll - defRoll
	if damage < 1 {
		damage = 1
	}
	// 10% crit chance for monsters
	if rand.Intn(100) < 10 {
		damage = int(float64(damage) * 1.5)
	}
	finalDmg := ApplyDamage(damage, models.Physical, target)
	target.HitpointsRemaining -= finalDmg
}

// TryUpgradeRarity attempts to upgrade a monster's rarity based on its monster kills.
// Returns true if an upgrade occurred.
func TryUpgradeRarity(mob *models.Monster) bool {
	chance := mob.MonsterKills * 2
	if chance > 50 {
		chance = 50
	}
	if rand.Intn(100) >= chance {
		return false
	}

	next := NextRarity(mob.Rarity)
	if next == "" {
		return false // already max rarity
	}

	mob.Rarity = next
	ApplyRarity(mob)
	mob.HitpointsRemaining = mob.HitpointsTotal
	return true
}

// ProcessLocationEvolution runs one evolution tick for a single location.
// Two random monsters fight; winner gets XP, equipment, and a chance to upgrade.
func ProcessLocationEvolution(loc *models.Location, gs *models.GameState) []EvolutionEvent {
	if loc.Type == "Base" || len(loc.Monsters) < 2 {
		return nil
	}

	events := []EvolutionEvent{}

	// Pick 2 random different indices
	idxA := rand.Intn(len(loc.Monsters))
	idxB := rand.Intn(len(loc.Monsters) - 1)
	if idxB >= idxA {
		idxB++
	}

	// Make copies for the fight
	a := loc.Monsters[idxA]
	b := loc.Monsters[idxB]

	winner := MonsterVsMonsterCombat(&a, &b)

	var winnerIdx, loserIdx int
	var loser *models.Monster
	if winner == &a {
		winnerIdx = idxA
		loserIdx = idxB
		loser = &b
	} else {
		winnerIdx = idxB
		loserIdx = idxA
		loser = &a
	}

	// Winner: increment kills, gain equipment, XP, level up, try rarity upgrade
	winner.MonsterKills++

	// Transfer loser's equipment
	for _, item := range loser.EquipmentMap {
		EquipBestItem(item, &winner.EquipmentMap, &winner.Inventory)
	}

	// XP gain
	winner.Experience += loser.Level * 100
	LevelUpMob(winner)

	// Restore winner to full HP
	winner.StatsMod = CalculateItemMods(winner.EquipmentMap)
	winner.HitpointsTotal = winner.HitpointsNatural + winner.StatsMod.HitPointMod
	winner.HitpointsRemaining = winner.HitpointsTotal
	winner.ManaRemaining = winner.ManaTotal
	winner.StaminaRemaining = winner.StaminaTotal
	winner.StatusEffects = []models.StatusEffect{}

	events = append(events, EvolutionEvent{
		EventType:    "fight",
		LocationName: loc.Name,
		Details:      fmt.Sprintf("%s (Lv%d) defeated %s (Lv%d)", winner.Name, winner.Level, loser.Name, loser.Level),
	})

	// Try rarity upgrade
	upgraded := TryUpgradeRarity(winner)
	if upgraded {
		events = append(events, EvolutionEvent{
			EventType:    "upgrade",
			LocationName: loc.Name,
			Details:      fmt.Sprintf("%s upgraded to %s rarity!", winner.Name, RarityDisplayName(winner.Rarity)),
		})

		// Check if the upgraded monster should migrate to a harder zone
		winnerCopy := *winner
		if evt := MigrateMonster(&winnerCopy, loc, gs); evt != nil {
			events = append(events, *evt)
			// Replace the migrated monster's old slot with a fresh one
			freshMob := GenerateBestMonster(gs, loc.LevelMax, loc.RarityMax)
			freshMob.LocationName = loc.Name
			loc.Monsters[winnerIdx] = freshMob
			// The migrated monster was placed in the target location by MigrateMonster
			winner = nil // signal that winner was already handled
		}
	}

	// Check level-based migration (if not already rarity-migrated)
	if winner != nil {
		winnerCopy := *winner
		if evt := MigrateMonsterByLevel(&winnerCopy, loc, gs); evt != nil {
			events = append(events, *evt)
			// Replace the migrated monster's old slot with a fresh one
			freshMob := GenerateBestMonster(gs, loc.LevelMax, loc.RarityMax)
			freshMob.LocationName = loc.Name
			loc.Monsters[winnerIdx] = freshMob
			winner = nil
		}
	}

	// Place winner back (if not migrated)
	if winner != nil {
		// Safety net: if winner exceeds location caps despite migration attempts,
		// replace with a fresh monster to prevent cap violations.
		overLevel := loc.LevelMax > 0 && winner.Level > loc.LevelMax
		overRarity := loc.RarityMax > 0 && RarityIndex(winner.Rarity) > loc.RarityMax
		if overLevel || overRarity {
			freshMob := GenerateBestMonster(gs, loc.LevelMax, loc.RarityMax)
			freshMob.LocationName = loc.Name
			loc.Monsters[winnerIdx] = freshMob
			events = append(events, EvolutionEvent{
				EventType:    "cap_enforce",
				LocationName: loc.Name,
				Details:      fmt.Sprintf("%s (Lv%d, %s) replaced — exceeded %s caps (Lv%d/R%d)", winner.Name, winner.Level, RarityDisplayName(winner.Rarity), loc.Name, loc.LevelMax, loc.RarityMax),
			})
		} else {
			loc.Monsters[winnerIdx] = *winner
		}
	}

	// Replace loser with a fresh monster
	newMob := GenerateBestMonster(gs, loc.LevelMax, loc.RarityMax)
	newMob.LocationName = loc.Name
	loc.Monsters[loserIdx] = newMob

	return events
}

// MigrateMonster checks if a monster has outgrown its location's rarity cap
// and migrates it to a suitable harder zone. Returns an EvolutionEvent if
// migration occurred, or nil if no migration was needed/possible.
func MigrateMonster(monster *models.Monster, fromLoc *models.Location, gs *models.GameState) *EvolutionEvent {
	rarityIdx := RarityIndex(monster.Rarity)

	// No migration needed if location is uncapped or monster fits
	if fromLoc.RarityMax == 0 || rarityIdx <= fromLoc.RarityMax {
		return nil
	}

	// Find candidate locations that can hold this rarity
	var candidates []string
	for name, loc := range gs.GameLocations {
		if name == fromLoc.Name {
			continue
		}
		if loc.Type == "Base" {
			continue
		}
		if len(loc.Monsters) == 0 {
			continue
		}
		// Location must allow this rarity (uncapped or cap >= monster's rarity index)
		if loc.RarityMax == 0 || loc.RarityMax >= rarityIdx {
			candidates = append(candidates, name)
		}
	}

	if len(candidates) == 0 {
		return nil
	}

	// Pick a random target location
	targetName := candidates[rand.Intn(len(candidates))]
	targetLoc := gs.GameLocations[targetName]

	// Replace a random monster in the target location
	replaceIdx := rand.Intn(len(targetLoc.Monsters))
	monster.LocationName = targetName
	targetLoc.Monsters[replaceIdx] = *monster
	gs.GameLocations[targetName] = targetLoc

	return &EvolutionEvent{
		EventType:    "migration",
		LocationName: fromLoc.Name,
		Details:      fmt.Sprintf("%s (%s) migrated from %s to %s", monster.Name, RarityDisplayName(monster.Rarity), fromLoc.Name, targetName),
	}
}

// MigrateMonsterByLevel checks if a monster has outgrown its location's level cap
// and migrates it to a suitable higher-level zone. Returns an EvolutionEvent if
// migration occurred, or nil if no migration was needed/possible.
func MigrateMonsterByLevel(monster *models.Monster, fromLoc *models.Location, gs *models.GameState) *EvolutionEvent {
	// No migration needed if location is uncapped or monster fits
	if fromLoc.LevelMax == 0 || monster.Level <= fromLoc.LevelMax {
		return nil
	}

	// Find candidate locations that can hold this level
	var candidates []string
	for name, loc := range gs.GameLocations {
		if name == fromLoc.Name {
			continue
		}
		if loc.Type == "Base" {
			continue
		}
		if len(loc.Monsters) == 0 {
			continue
		}
		// Location must allow this level (uncapped or cap >= monster's level)
		if loc.LevelMax == 0 || loc.LevelMax >= monster.Level {
			candidates = append(candidates, name)
		}
	}

	if len(candidates) == 0 {
		return nil
	}

	// Pick a random target location
	targetName := candidates[rand.Intn(len(candidates))]
	targetLoc := gs.GameLocations[targetName]

	// Replace a random monster in the target location
	replaceIdx := rand.Intn(len(targetLoc.Monsters))
	monster.LocationName = targetName
	targetLoc.Monsters[replaceIdx] = *monster
	gs.GameLocations[targetName] = targetLoc

	return &EvolutionEvent{
		EventType:    "migration",
		LocationName: fromLoc.Name,
		Details:      fmt.Sprintf("%s (Lv%d) migrated from %s to %s (exceeded level cap %d)", monster.Name, monster.Level, fromLoc.Name, targetName, fromLoc.LevelMax),
	}
}

// MigrateMonsterIDs assigns IDs and LocationNames to any monsters missing them.
func MigrateMonsterIDs(locations map[string]models.Location) {
	for locName, loc := range locations {
		changed := false
		for i := range loc.Monsters {
			if loc.Monsters[i].ID == "" {
				loc.Monsters[i].ID = fmt.Sprintf("mob-%d-%d", time.Now().UnixNano(), rand.Intn(100000))
				changed = true
			}
			if loc.Monsters[i].LocationName == "" {
				loc.Monsters[i].LocationName = locName
				changed = true
			}
		}
		if changed {
			locations[locName] = loc
		}
	}
}

// GetMostWanted scans all locations and returns the top N monsters ranked by total kills.
func GetMostWanted(locations map[string]models.Location, limit int) []models.MostWantedEntry {
	var entries []models.MostWantedEntry

	for _, loc := range locations {
		for idx, mob := range loc.Monsters {
			if mob.PlayerKills > 0 {
				entries = append(entries, models.MostWantedEntry{
					MonsterID:    mob.ID,
					Name:         mob.Name,
					MonsterType:  mob.MonsterType,
					Level:        mob.Level,
					Rarity:       mob.Rarity,
					PlayerKills:  mob.PlayerKills,
					MonsterKills: mob.MonsterKills,
					LocationName: mob.LocationName,
					LocationIdx:  idx,
					IsBoss:       mob.IsBoss,
					HP:           mob.HitpointsTotal,
				})
			}
		}
	}

	// Sort by player kills descending (only player/guard/NPC kills count for Most Wanted)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].PlayerKills > entries[j].PlayerKills
	})

	if limit > 0 && len(entries) > limit {
		entries = entries[:limit]
	}

	return entries
}
