package game

import (
	"testing"

	"rpg-game/pkg/models"
)

func TestRecordKill(t *testing.T) {
	stats := &models.CharacterStats{}

	RecordKill(stats, "Goblin", models.RarityCommon, "Dark Forest")

	if stats.TotalKills != 1 {
		t.Errorf("expected TotalKills=1, got %d", stats.TotalKills)
	}
	if stats.CurrentCombo != 1 {
		t.Errorf("expected CurrentCombo=1, got %d", stats.CurrentCombo)
	}
	if stats.HighestCombo != 1 {
		t.Errorf("expected HighestCombo=1, got %d", stats.HighestCombo)
	}
	if stats.KillsByRarity["common"] != 1 {
		t.Errorf("expected KillsByRarity[\"common\"]=1, got %d", stats.KillsByRarity["common"])
	}
	if stats.KillsByMonster["Goblin"] != 1 {
		t.Errorf("expected KillsByMonster[\"Goblin\"]=1, got %d", stats.KillsByMonster["Goblin"])
	}
	if stats.KillsByLocation["Dark Forest"] != 1 {
		t.Errorf("expected KillsByLocation[\"Dark Forest\"]=1, got %d", stats.KillsByLocation["Dark Forest"])
	}

	// Record a second kill with different parameters.
	RecordKill(stats, "Orc", models.RarityRare, "Dark Forest")

	if stats.TotalKills != 2 {
		t.Errorf("expected TotalKills=2 after second kill, got %d", stats.TotalKills)
	}
	if stats.CurrentCombo != 2 {
		t.Errorf("expected CurrentCombo=2 after second kill, got %d", stats.CurrentCombo)
	}
	if stats.KillsByRarity["rare"] != 1 {
		t.Errorf("expected KillsByRarity[\"rare\"]=1, got %d", stats.KillsByRarity["rare"])
	}
	if stats.KillsByMonster["Orc"] != 1 {
		t.Errorf("expected KillsByMonster[\"Orc\"]=1, got %d", stats.KillsByMonster["Orc"])
	}
	if stats.KillsByLocation["Dark Forest"] != 2 {
		t.Errorf("expected KillsByLocation[\"Dark Forest\"]=2, got %d", stats.KillsByLocation["Dark Forest"])
	}

	// Empty location should not be recorded.
	RecordKill(stats, "Slime", models.RarityCommon, "")
	if _, exists := stats.KillsByLocation[""]; exists {
		t.Error("expected empty location not to be recorded in KillsByLocation")
	}

	// Empty rarity should normalize to common.
	RecordKill(stats, "Slime", "", "Swamp")
	if stats.KillsByRarity["common"] != 3 {
		t.Errorf("expected empty rarity to normalize to common, KillsByRarity[\"common\"]=%d", stats.KillsByRarity["common"])
	}
}

func TestRecordDeath(t *testing.T) {
	stats := &models.CharacterStats{}

	RecordDeath(stats)

	if stats.TotalDeaths != 1 {
		t.Errorf("expected TotalDeaths=1, got %d", stats.TotalDeaths)
	}
	if stats.CurrentCombo != 0 {
		t.Errorf("expected CurrentCombo=0 after death, got %d", stats.CurrentCombo)
	}

	// Build up a combo then die.
	RecordKill(stats, "Goblin", models.RarityCommon, "Forest")
	RecordKill(stats, "Goblin", models.RarityCommon, "Forest")
	if stats.CurrentCombo != 2 {
		t.Errorf("expected CurrentCombo=2, got %d", stats.CurrentCombo)
	}

	RecordDeath(stats)

	if stats.TotalDeaths != 2 {
		t.Errorf("expected TotalDeaths=2, got %d", stats.TotalDeaths)
	}
	if stats.CurrentCombo != 0 {
		t.Errorf("expected CurrentCombo=0 after second death, got %d", stats.CurrentCombo)
	}
}

func TestRecordBossKill(t *testing.T) {
	stats := &models.CharacterStats{}

	RecordBossKill(stats)
	if stats.BossesKilled != 1 {
		t.Errorf("expected BossesKilled=1, got %d", stats.BossesKilled)
	}

	RecordBossKill(stats)
	RecordBossKill(stats)
	if stats.BossesKilled != 3 {
		t.Errorf("expected BossesKilled=3, got %d", stats.BossesKilled)
	}
}

func TestRecordXPGained(t *testing.T) {
	stats := &models.CharacterStats{}

	RecordXPGained(stats, 100)
	if stats.TotalXPEarned != 100 {
		t.Errorf("expected TotalXPEarned=100, got %d", stats.TotalXPEarned)
	}

	RecordXPGained(stats, 250)
	if stats.TotalXPEarned != 350 {
		t.Errorf("expected TotalXPEarned=350, got %d", stats.TotalXPEarned)
	}

	// Zero XP should not change the total.
	RecordXPGained(stats, 0)
	if stats.TotalXPEarned != 350 {
		t.Errorf("expected TotalXPEarned=350 after adding 0, got %d", stats.TotalXPEarned)
	}
}

func TestRecordPvPResult(t *testing.T) {
	stats := &models.CharacterStats{}

	RecordPvPResult(stats, true)
	if stats.PvPWins != 1 {
		t.Errorf("expected PvPWins=1, got %d", stats.PvPWins)
	}
	if stats.PvPLosses != 0 {
		t.Errorf("expected PvPLosses=0, got %d", stats.PvPLosses)
	}

	RecordPvPResult(stats, false)
	if stats.PvPWins != 1 {
		t.Errorf("expected PvPWins=1 after loss, got %d", stats.PvPWins)
	}
	if stats.PvPLosses != 1 {
		t.Errorf("expected PvPLosses=1, got %d", stats.PvPLosses)
	}

	RecordPvPResult(stats, true)
	RecordPvPResult(stats, true)
	RecordPvPResult(stats, false)

	if stats.PvPWins != 3 {
		t.Errorf("expected PvPWins=3, got %d", stats.PvPWins)
	}
	if stats.PvPLosses != 2 {
		t.Errorf("expected PvPLosses=2, got %d", stats.PvPLosses)
	}
}

func TestRecordDungeonClear(t *testing.T) {
	stats := &models.CharacterStats{}

	RecordDungeonClear(stats)
	if stats.DungeonsCleared != 1 {
		t.Errorf("expected DungeonsCleared=1, got %d", stats.DungeonsCleared)
	}

	RecordDungeonClear(stats)
	RecordDungeonClear(stats)
	if stats.DungeonsCleared != 3 {
		t.Errorf("expected DungeonsCleared=3, got %d", stats.DungeonsCleared)
	}
}

func TestComboTracking(t *testing.T) {
	stats := &models.CharacterStats{}

	// Build a 5-kill combo.
	for i := 0; i < 5; i++ {
		RecordKill(stats, "Goblin", models.RarityCommon, "Forest")
	}
	if stats.CurrentCombo != 5 {
		t.Errorf("expected CurrentCombo=5 after 5 kills, got %d", stats.CurrentCombo)
	}
	if stats.HighestCombo != 5 {
		t.Errorf("expected HighestCombo=5, got %d", stats.HighestCombo)
	}

	// Die -- combo resets but highest is preserved.
	RecordDeath(stats)
	if stats.CurrentCombo != 0 {
		t.Errorf("expected CurrentCombo=0 after death, got %d", stats.CurrentCombo)
	}
	if stats.HighestCombo != 5 {
		t.Errorf("expected HighestCombo=5 preserved after death, got %d", stats.HighestCombo)
	}

	// Build a smaller combo -- highest should remain 5.
	RecordKill(stats, "Slime", models.RarityCommon, "Swamp")
	RecordKill(stats, "Slime", models.RarityCommon, "Swamp")
	if stats.CurrentCombo != 2 {
		t.Errorf("expected CurrentCombo=2, got %d", stats.CurrentCombo)
	}
	if stats.HighestCombo != 5 {
		t.Errorf("expected HighestCombo=5 unchanged when current combo is lower, got %d", stats.HighestCombo)
	}

	// Build past the old highest combo -- highest should update.
	for i := 0; i < 5; i++ {
		RecordKill(stats, "Orc", models.RarityRare, "Mountain")
	}
	if stats.CurrentCombo != 7 {
		t.Errorf("expected CurrentCombo=7, got %d", stats.CurrentCombo)
	}
	if stats.HighestCombo != 7 {
		t.Errorf("expected HighestCombo=7 after surpassing old record, got %d", stats.HighestCombo)
	}

	// Die again -- highest still 7.
	RecordDeath(stats)
	if stats.CurrentCombo != 0 {
		t.Errorf("expected CurrentCombo=0 after second death, got %d", stats.CurrentCombo)
	}
	if stats.HighestCombo != 7 {
		t.Errorf("expected HighestCombo=7 preserved after second death, got %d", stats.HighestCombo)
	}
}
