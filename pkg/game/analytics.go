package game

import "rpg-game/pkg/models"

// RecordKill increments kill-related stats on the character.
func RecordKill(stats *models.CharacterStats, monsterName string, rarity models.MonsterRarity, location string) {
	stats.TotalKills++
	stats.CurrentCombo++
	if stats.CurrentCombo > stats.HighestCombo {
		stats.HighestCombo = stats.CurrentCombo
	}

	r := string(NormalizeRarity(rarity))
	if stats.KillsByRarity == nil {
		stats.KillsByRarity = make(map[string]int)
	}
	stats.KillsByRarity[r]++

	if stats.KillsByMonster == nil {
		stats.KillsByMonster = make(map[string]int)
	}
	stats.KillsByMonster[monsterName]++

	if stats.KillsByLocation == nil {
		stats.KillsByLocation = make(map[string]int)
	}
	if location != "" {
		stats.KillsByLocation[location]++
	}
}

// RecordDeath increments death stats and resets combo.
func RecordDeath(stats *models.CharacterStats) {
	stats.TotalDeaths++
	stats.CurrentCombo = 0
}

// RecordBossKill increments the boss kill counter.
func RecordBossKill(stats *models.CharacterStats) {
	stats.BossesKilled++
}

// RecordXPGained adds to total XP earned tracking.
func RecordXPGained(stats *models.CharacterStats, xp int) {
	stats.TotalXPEarned += xp
}

// RecordPvPResult records a PvP win or loss.
func RecordPvPResult(stats *models.CharacterStats, won bool) {
	if won {
		stats.PvPWins++
	} else {
		stats.PvPLosses++
	}
}

// RecordDungeonClear increments the dungeons cleared counter.
func RecordDungeonClear(stats *models.CharacterStats) {
	stats.DungeonsCleared++
}
