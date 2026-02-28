package metrics

import (
	"encoding/json"
	"sync"
	"testing"
)

func TestRecordCombatWin(t *testing.T) {
	mc := NewMetricsCollector()
	mc.RecordCombatWin("Forest", "Goblin", "Common", 5)

	if mc.TotalFights.Load() != 1 {
		t.Errorf("TotalFights = %d, want 1", mc.TotalFights.Load())
	}
	if mc.PlayerWins.Load() != 1 {
		t.Errorf("PlayerWins = %d, want 1", mc.PlayerWins.Load())
	}
	if mc.CombatTurns.Load() != 5 {
		t.Errorf("CombatTurns = %d, want 5", mc.CombatTurns.Load())
	}
	if mc.WinsByLocation["Forest"] != 1 {
		t.Errorf("WinsByLocation[Forest] = %d, want 1", mc.WinsByLocation["Forest"])
	}
	if mc.WinsByMonsterType["Goblin"] != 1 {
		t.Errorf("WinsByMonsterType[Goblin] = %d, want 1", mc.WinsByMonsterType["Goblin"])
	}
	if mc.WinsByRarity["Common"] != 1 {
		t.Errorf("WinsByRarity[Common] = %d, want 1", mc.WinsByRarity["Common"])
	}
}

func TestRecordCombatLoss(t *testing.T) {
	mc := NewMetricsCollector()
	mc.RecordCombatLoss("Cave", "Orc", "Rare", 10)

	if mc.TotalFights.Load() != 1 {
		t.Errorf("TotalFights = %d, want 1", mc.TotalFights.Load())
	}
	if mc.PlayerDeaths.Load() != 1 {
		t.Errorf("PlayerDeaths = %d, want 1", mc.PlayerDeaths.Load())
	}
	if mc.LossesByLocation["Cave"] != 1 {
		t.Errorf("LossesByLocation[Cave] = %d, want 1", mc.LossesByLocation["Cave"])
	}
}

func TestRecordFlee(t *testing.T) {
	mc := NewMetricsCollector()
	mc.RecordFlee(true)
	mc.RecordFlee(false)
	mc.RecordFlee(true)

	if mc.Flees.Load() != 2 {
		t.Errorf("Flees = %d, want 2", mc.Flees.Load())
	}
	if mc.FleeFails.Load() != 1 {
		t.Errorf("FleeFails = %d, want 1", mc.FleeFails.Load())
	}
}

func TestRecordCrit(t *testing.T) {
	mc := NewMetricsCollector()
	mc.RecordCrit(true)
	mc.RecordCrit(true)
	mc.RecordCrit(false)

	if mc.PlayerCrits.Load() != 2 {
		t.Errorf("PlayerCrits = %d, want 2", mc.PlayerCrits.Load())
	}
	if mc.MonsterCrits.Load() != 1 {
		t.Errorf("MonsterCrits = %d, want 1", mc.MonsterCrits.Load())
	}
}

func TestRecordDamage(t *testing.T) {
	mc := NewMetricsCollector()
	mc.RecordDamage(50, "physical", true)
	mc.RecordDamage(30, "fire", true)
	mc.RecordDamage(20, "physical", false)

	if mc.PlayerDamageDealt.Load() != 80 {
		t.Errorf("PlayerDamageDealt = %d, want 80", mc.PlayerDamageDealt.Load())
	}
	if mc.MonsterDamageDealt.Load() != 20 {
		t.Errorf("MonsterDamageDealt = %d, want 20", mc.MonsterDamageDealt.Load())
	}
	if mc.DamageByType["physical"] != 70 {
		t.Errorf("DamageByType[physical] = %d, want 70", mc.DamageByType["physical"])
	}
	if mc.DamageByType["fire"] != 30 {
		t.Errorf("DamageByType[fire] = %d, want 30", mc.DamageByType["fire"])
	}
}

func TestRecordSkillUse(t *testing.T) {
	mc := NewMetricsCollector()
	mc.RecordSkillUse("Fireball")
	mc.RecordSkillUse("Fireball")
	mc.RecordSkillUse("Heal")

	if mc.SkillUses.Load() != 3 {
		t.Errorf("SkillUses = %d, want 3", mc.SkillUses.Load())
	}
	if mc.SkillUseCounts["Fireball"] != 2 {
		t.Errorf("SkillUseCounts[Fireball] = %d, want 2", mc.SkillUseCounts["Fireball"])
	}
}

func TestRecordItemUse(t *testing.T) {
	mc := NewMetricsCollector()
	mc.RecordItemUse("Small Health Potion")
	mc.RecordItemUse("Small Health Potion")

	if mc.ItemUses.Load() != 2 {
		t.Errorf("ItemUses = %d, want 2", mc.ItemUses.Load())
	}
	if mc.PotionsUsed["Small Health Potion"] != 2 {
		t.Errorf("PotionsUsed = %d, want 2", mc.PotionsUsed["Small Health Potion"])
	}
}

func TestRecordLevelUp(t *testing.T) {
	mc := NewMetricsCollector()
	mc.RecordLevelUp(5)
	mc.RecordLevelUp(5)
	mc.RecordLevelUp(10)

	if mc.LevelUps.Load() != 3 {
		t.Errorf("LevelUps = %d, want 3", mc.LevelUps.Load())
	}
	if mc.LevelUpsByLevel["5"] != 2 {
		t.Errorf("LevelUpsByLevel[5] = %d, want 2", mc.LevelUpsByLevel["5"])
	}
}

func TestRecordHarvest(t *testing.T) {
	mc := NewMetricsCollector()
	mc.RecordHarvest("Gold", 15)
	mc.RecordHarvest("Iron", 10)

	if mc.Harvests.Load() != 2 {
		t.Errorf("Harvests = %d, want 2", mc.Harvests.Load())
	}
	if mc.ResourceUnits.Load() != 25 {
		t.Errorf("ResourceUnits = %d, want 25", mc.ResourceUnits.Load())
	}
	if mc.HarvestsByResource["Gold"] != 15 {
		t.Errorf("HarvestsByResource[Gold] = %d, want 15", mc.HarvestsByResource["Gold"])
	}
}

func TestRecordArenaFight(t *testing.T) {
	mc := NewMetricsCollector()
	mc.RecordArenaFight(1200, 1150) // gap 50 -> "0-100"
	mc.RecordArenaFight(1500, 1350) // gap 150 -> "100-200"
	mc.RecordArenaFight(1800, 1500) // gap 300 -> "200+"

	if mc.ArenaFights.Load() != 3 {
		t.Errorf("ArenaFights = %d, want 3", mc.ArenaFights.Load())
	}
	if mc.ArenaWinsByGap["0-100"] != 1 {
		t.Errorf("ArenaWinsByGap[0-100] = %d, want 1", mc.ArenaWinsByGap["0-100"])
	}
	if mc.ArenaWinsByGap["100-200"] != 1 {
		t.Errorf("ArenaWinsByGap[100-200] = %d, want 1", mc.ArenaWinsByGap["100-200"])
	}
	if mc.ArenaWinsByGap["200+"] != 1 {
		t.Errorf("ArenaWinsByGap[200+] = %d, want 1", mc.ArenaWinsByGap["200+"])
	}
}

func TestRecordDungeon(t *testing.T) {
	mc := NewMetricsCollector()
	mc.RecordDungeonEnter()
	mc.RecordDungeonEnter()
	mc.RecordFloorClear(1)
	mc.RecordFloorClear(2)
	mc.RecordDungeonClear()
	mc.RecordDungeonDeath(3)

	if mc.DungeonEnters.Load() != 2 {
		t.Errorf("DungeonEnters = %d, want 2", mc.DungeonEnters.Load())
	}
	if mc.DungeonClears.Load() != 1 {
		t.Errorf("DungeonClears = %d, want 1", mc.DungeonClears.Load())
	}
	if mc.DungeonDeaths.Load() != 1 {
		t.Errorf("DungeonDeaths = %d, want 1", mc.DungeonDeaths.Load())
	}
	if mc.FloorClears["1"] != 1 {
		t.Errorf("FloorClears[1] = %d, want 1", mc.FloorClears["1"])
	}
	if mc.FloorDeaths["3"] != 1 {
		t.Errorf("FloorDeaths[3] = %d, want 1", mc.FloorDeaths["3"])
	}
}

func TestRecordFeatureUse(t *testing.T) {
	mc := NewMetricsCollector()
	mc.RecordFeatureUse("hunt")
	mc.RecordFeatureUse("hunt")
	mc.RecordFeatureUse("arena")

	if mc.FeatureUsage["hunt"] != 2 {
		t.Errorf("FeatureUsage[hunt] = %d, want 2", mc.FeatureUsage["hunt"])
	}
	if mc.FeatureUsage["arena"] != 1 {
		t.Errorf("FeatureUsage[arena] = %d, want 1", mc.FeatureUsage["arena"])
	}
}

func TestSnapshotComputedRates(t *testing.T) {
	mc := NewMetricsCollector()
	// 7 wins, 3 losses = 0.7 win rate
	for i := 0; i < 7; i++ {
		mc.RecordCombatWin("Forest", "Goblin", "Common", 10)
	}
	for i := 0; i < 3; i++ {
		mc.RecordCombatLoss("Forest", "Goblin", "Common", 10)
	}
	// 6 flees, 4 fails = 0.6 flee rate
	for i := 0; i < 6; i++ {
		mc.RecordFlee(true)
	}
	for i := 0; i < 4; i++ {
		mc.RecordFlee(false)
	}

	snap := mc.Snapshot()

	if snap.Combat.TotalFights != 10 {
		t.Errorf("TotalFights = %d, want 10", snap.Combat.TotalFights)
	}
	if snap.Combat.WinRate < 0.69 || snap.Combat.WinRate > 0.71 {
		t.Errorf("WinRate = %f, want ~0.7", snap.Combat.WinRate)
	}
	if snap.Combat.FleeSuccessRate < 0.59 || snap.Combat.FleeSuccessRate > 0.61 {
		t.Errorf("FleeSuccessRate = %f, want ~0.6", snap.Combat.FleeSuccessRate)
	}
	if snap.Combat.AvgTurns < 9.9 || snap.Combat.AvgTurns > 10.1 {
		t.Errorf("AvgTurns = %f, want ~10", snap.Combat.AvgTurns)
	}

	// Verify location win/loss merge
	wl, ok := snap.Combat.ByLocation["Forest"]
	if !ok {
		t.Fatal("ByLocation missing Forest")
	}
	if wl.W != 7 || wl.L != 3 {
		t.Errorf("ByLocation[Forest] = {W:%d, L:%d}, want {W:7, L:3}", wl.W, wl.L)
	}
}

func TestSnapshotJSON(t *testing.T) {
	mc := NewMetricsCollector()
	mc.RecordCombatWin("Forest", "Goblin", "Common", 5)

	jsonStr, err := mc.SnapshotJSON()
	if err != nil {
		t.Fatalf("SnapshotJSON error: %v", err)
	}

	var snap MetricsSnapshot
	if err := json.Unmarshal([]byte(jsonStr), &snap); err != nil {
		t.Fatalf("Failed to unmarshal snapshot JSON: %v", err)
	}
	if snap.Combat.Wins != 1 {
		t.Errorf("JSON snapshot wins = %d, want 1", snap.Combat.Wins)
	}
}

func TestConcurrentRecording(t *testing.T) {
	mc := NewMetricsCollector()
	var wg sync.WaitGroup

	// 100 goroutines each recording 100 events
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				mc.RecordCombatWin("Forest", "Goblin", "Common", 1)
				mc.RecordCombatLoss("Cave", "Orc", "Rare", 2)
				mc.RecordFlee(j%2 == 0)
				mc.RecordCrit(j%3 == 0)
				mc.RecordDamage(10, "physical", true)
				mc.RecordSkillUse("Fireball")
				mc.RecordItemUse("Potion")
				mc.RecordLevelUp(5)
				mc.RecordHarvest("Gold", 1)
				mc.RecordFeatureUse("hunt")
				mc.RecordDungeonEnter()
				mc.RecordArenaFight(1200, 1100)
			}
		}()
	}
	wg.Wait()

	// Verify totals
	if mc.TotalFights.Load() != 20000 { // 100*100*2
		t.Errorf("TotalFights = %d, want 20000", mc.TotalFights.Load())
	}
	if mc.PlayerWins.Load() != 10000 {
		t.Errorf("PlayerWins = %d, want 10000", mc.PlayerWins.Load())
	}
	if mc.PlayerDeaths.Load() != 10000 {
		t.Errorf("PlayerDeaths = %d, want 10000", mc.PlayerDeaths.Load())
	}
	if mc.SkillUses.Load() != 10000 {
		t.Errorf("SkillUses = %d, want 10000", mc.SkillUses.Load())
	}

	// Verify snapshot doesn't panic under concurrent access
	snap := mc.Snapshot()
	if snap.Combat.TotalFights != 20000 {
		t.Errorf("Snapshot TotalFights = %d, want 20000", snap.Combat.TotalFights)
	}
}

func TestZeroDivisionSafety(t *testing.T) {
	mc := NewMetricsCollector()
	snap := mc.Snapshot()

	if snap.Combat.WinRate != 0 {
		t.Errorf("WinRate = %f, want 0", snap.Combat.WinRate)
	}
	if snap.Combat.FleeSuccessRate != 0 {
		t.Errorf("FleeSuccessRate = %f, want 0", snap.Combat.FleeSuccessRate)
	}
	if snap.Combat.AvgTurns != 0 {
		t.Errorf("AvgTurns = %f, want 0", snap.Combat.AvgTurns)
	}
	if snap.Dungeons.ClearRate != 0 {
		t.Errorf("ClearRate = %f, want 0", snap.Dungeons.ClearRate)
	}
}
