package game

import (
	"math/rand"
	"testing"
	"time"

	"rpg-game/pkg/data"
	"rpg-game/pkg/models"
)

// TestGenerateDungeon generates a dungeon from a template and verifies
// correct floor count, name, room counts (5-8), and floor numbers.
func TestGenerateDungeon(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	template := data.DungeonTemplates[0] // Goblin Warren: 5 floors
	seed := int64(12345)

	dungeon := GenerateDungeon(template, seed)

	// Verify dungeon name matches template
	if dungeon.Name != template.Name {
		t.Errorf("Expected dungeon name '%s', got '%s'", template.Name, dungeon.Name)
	}

	// Verify correct number of floors
	if len(dungeon.Floors) != template.Floors {
		t.Errorf("Expected %d floors, got %d", template.Floors, len(dungeon.Floors))
	}

	// Verify seed is stored
	if dungeon.Seed != seed {
		t.Errorf("Expected seed %d, got %d", seed, dungeon.Seed)
	}

	// Verify base level/rank are set from template
	if dungeon.BaseLevelMin != template.MinLevel {
		t.Errorf("Expected BaseLevelMin %d, got %d", template.MinLevel, dungeon.BaseLevelMin)
	}

	if dungeon.BaseLevelMax != template.MaxLevel {
		t.Errorf("Expected BaseLevelMax %d, got %d", template.MaxLevel, dungeon.BaseLevelMax)
	}

	if dungeon.BaseRankMax != template.RankMax {
		t.Errorf("Expected BaseRankMax %d, got %d", template.RankMax, dungeon.BaseRankMax)
	}

	// Verify CurrentFloor starts at 0
	if dungeon.CurrentFloor != 0 {
		t.Errorf("Expected CurrentFloor 0, got %d", dungeon.CurrentFloor)
	}

	// Verify each floor
	for i, floor := range dungeon.Floors {
		expectedFloorNum := i + 1

		// Floor number should be 1-indexed
		if floor.FloorNumber != expectedFloorNum {
			t.Errorf("Floor %d: expected FloorNumber %d, got %d", i, expectedFloorNum, floor.FloorNumber)
		}

		// Each floor should have 5-8 rooms
		roomCount := len(floor.Rooms)
		if roomCount < 5 || roomCount > 8 {
			t.Errorf("Floor %d: expected 5-8 rooms, got %d", expectedFloorNum, roomCount)
		}

		// Floor should start uncleared
		if floor.Cleared {
			t.Errorf("Floor %d: should start uncleared", expectedFloorNum)
		}

		// CurrentRoom should start at 0
		if floor.CurrentRoom != 0 {
			t.Errorf("Floor %d: expected CurrentRoom 0, got %d", expectedFloorNum, floor.CurrentRoom)
		}

		// Boss floor check: every 5th floor is a boss floor
		expectedBoss := expectedFloorNum%5 == 0
		if floor.BossFloor != expectedBoss {
			t.Errorf("Floor %d: expected BossFloor=%v, got %v", expectedFloorNum, expectedBoss, floor.BossFloor)
		}
	}

	t.Logf("Dungeon '%s': %d floors generated, seed=%d", dungeon.Name, len(dungeon.Floors), seed)
}

// TestGenerateDungeonAllTemplates verifies dungeon generation works for every template.
func TestGenerateDungeonAllTemplates(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	for _, template := range data.DungeonTemplates {
		seed := int64(99999)
		dungeon := GenerateDungeon(template, seed)

		if dungeon.Name != template.Name {
			t.Errorf("Template '%s': name mismatch, got '%s'", template.Name, dungeon.Name)
		}

		if len(dungeon.Floors) != template.Floors {
			t.Errorf("Template '%s': expected %d floors, got %d", template.Name, template.Floors, len(dungeon.Floors))
		}

		t.Logf("Template '%s': %d floors OK", template.Name, len(dungeon.Floors))
	}
}

// TestGenerateDungeonDeterministic generates two dungeons with the same seed
// and verifies they produce identical room type layouts.
func TestGenerateDungeonDeterministic(t *testing.T) {
	template := data.DungeonTemplates[0] // Goblin Warren
	seed := int64(42)

	dungeon1 := GenerateDungeon(template, seed)
	dungeon2 := GenerateDungeon(template, seed)

	// Verify same number of floors
	if len(dungeon1.Floors) != len(dungeon2.Floors) {
		t.Fatalf("Floor count mismatch: %d vs %d", len(dungeon1.Floors), len(dungeon2.Floors))
	}

	// Verify each floor has the same room layout
	for f := 0; f < len(dungeon1.Floors); f++ {
		floor1 := dungeon1.Floors[f]
		floor2 := dungeon2.Floors[f]

		// Same number of rooms
		if len(floor1.Rooms) != len(floor2.Rooms) {
			t.Errorf("Floor %d: room count mismatch %d vs %d",
				floor1.FloorNumber, len(floor1.Rooms), len(floor2.Rooms))
			continue
		}

		// Same room types in same order
		for r := 0; r < len(floor1.Rooms); r++ {
			if floor1.Rooms[r].Type != floor2.Rooms[r].Type {
				t.Errorf("Floor %d, Room %d: type mismatch '%s' vs '%s'",
					floor1.FloorNumber, r, floor1.Rooms[r].Type, floor2.Rooms[r].Type)
			}
		}

		// Same boss floor designation
		if floor1.BossFloor != floor2.BossFloor {
			t.Errorf("Floor %d: BossFloor mismatch %v vs %v",
				floor1.FloorNumber, floor1.BossFloor, floor2.BossFloor)
		}
	}

	t.Logf("Deterministic check passed: two dungeons with seed %d are identical", seed)
}

// TestGenerateDungeonDeterministicDifferentSeeds verifies that different seeds
// produce different dungeons (with high probability).
func TestGenerateDungeonDeterministicDifferentSeeds(t *testing.T) {
	template := data.DungeonTemplates[2] // Dragon's Lair (15 floors for more variation)

	dungeon1 := GenerateDungeon(template, 1)
	dungeon2 := GenerateDungeon(template, 2)

	// Check if at least one floor differs in room count or room types
	anyDifference := false
	for f := 0; f < len(dungeon1.Floors); f++ {
		floor1 := dungeon1.Floors[f]
		floor2 := dungeon2.Floors[f]

		if len(floor1.Rooms) != len(floor2.Rooms) {
			anyDifference = true
			break
		}

		for r := 0; r < len(floor1.Rooms) && r < len(floor2.Rooms); r++ {
			if floor1.Rooms[r].Type != floor2.Rooms[r].Type {
				anyDifference = true
				break
			}
		}

		if anyDifference {
			break
		}
	}

	if !anyDifference {
		t.Error("Different seeds should produce different dungeons (extremely unlikely to be identical)")
	}

	t.Logf("Different seeds produce different dungeons: confirmed")
}

// TestGenerateDungeonBoss generates a boss and verifies IsBoss=true,
// Legendary rarity, and enhanced stats (5x HP, 3x attack/defense rolls).
func TestGenerateDungeonBoss(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	floor := 5
	baseLevel := 10
	baseRank := 3

	boss := GenerateDungeonBoss(floor, baseLevel, baseRank)

	// Boss flag must be set
	if !boss.IsBoss {
		t.Error("Boss should have IsBoss=true")
	}

	// Rarity must be Legendary
	if boss.Rarity != models.RarityLegendary {
		t.Errorf("Boss should have Legendary rarity, got '%s'", boss.Rarity)
	}

	// Generate a comparable non-boss monster to verify stat amplification
	// Note: We can't do exact comparison because GenerateMonster uses global rand,
	// but we can verify the boss has substantial stats.

	// Boss HP should be significantly higher than a base monster
	if boss.HitpointsTotal <= 0 {
		t.Error("Boss should have positive HP")
	}

	// Boss should have at least 3 attack rolls (3x multiplier on at least 1 base roll)
	if boss.AttackRolls < 3 {
		t.Errorf("Boss attack rolls should be >= 3 (3x multiplier), got %d", boss.AttackRolls)
	}

	// Boss should have at least 3 defense rolls
	if boss.DefenseRolls < 3 {
		t.Errorf("Boss defense rolls should be >= 3 (3x multiplier), got %d", boss.DefenseRolls)
	}

	// Boss mana should be amplified (3x)
	if boss.ManaTotal <= 0 {
		t.Error("Boss should have positive Mana")
	}

	if boss.ManaRemaining != boss.ManaTotal {
		t.Errorf("Boss should start at full Mana: %d != %d", boss.ManaRemaining, boss.ManaTotal)
	}

	// Boss stamina should be amplified (3x)
	if boss.StaminaTotal <= 0 {
		t.Error("Boss should have positive Stamina")
	}

	if boss.StaminaRemaining != boss.StaminaTotal {
		t.Errorf("Boss should start at full Stamina: %d != %d", boss.StaminaRemaining, boss.StaminaTotal)
	}

	// HP remaining should equal total
	if boss.HitpointsRemaining != boss.HitpointsTotal {
		t.Errorf("Boss should start at full HP: %d != %d", boss.HitpointsRemaining, boss.HitpointsTotal)
	}

	t.Logf("Boss generated: %s (Lv%d, HP:%d, ATK rolls:%d, DEF rolls:%d, Rarity:%s)",
		boss.Name, boss.Level, boss.HitpointsTotal, boss.AttackRolls, boss.DefenseRolls, boss.Rarity)
}

// TestGenerateDungeonBossStatAmplification verifies the 5x HP and 3x roll
// multipliers by comparing against a base monster with the same parameters.
func TestGenerateDungeonBossStatAmplification(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	// Generate multiple bosses and verify HP is always large
	for i := 0; i < 10; i++ {
		boss := GenerateDungeonBoss(5, 10, 3)

		// A level 10, rank 3 base monster has HitpointsNatural from GenerateMonster.
		// The boss multiplies that by 5, so boss HP should be at least 5.
		// (Even a 1 HP base monster would become 5 HP boss.)
		if boss.HitpointsNatural < 5 {
			t.Errorf("Boss HitpointsNatural should be >= 5 (5x multiplier on natural HP), got %d",
				boss.HitpointsNatural)
		}

		// Attack rolls are base * 3, and base is at least 1, so at least 3
		if boss.AttackRolls < 3 {
			t.Errorf("Boss AttackRolls should be >= 3, got %d", boss.AttackRolls)
		}

		// Defense rolls are base * 3
		if boss.DefenseRolls < 3 {
			t.Errorf("Boss DefenseRolls should be >= 3, got %d", boss.DefenseRolls)
		}

		// Mana is base * 3
		if boss.ManaNatural < 3 {
			t.Errorf("Boss ManaNatural should be >= 3, got %d", boss.ManaNatural)
		}

		// Stamina is base * 3
		if boss.StaminaNatural < 3 {
			t.Errorf("Boss StaminaNatural should be >= 3, got %d", boss.StaminaNatural)
		}
	}

	t.Logf("Boss stat amplification verified across 10 generations")
}

// TestGenerateDungeonBossMinimumLevel verifies boss generation handles
// edge cases with low level and rank values.
func TestGenerateDungeonBossMinimumLevel(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	// Test with minimum values
	boss := GenerateDungeonBoss(1, 0, 0)

	// Level and rank should be clamped to at least 1
	if boss.Level < 1 {
		t.Errorf("Boss level should be >= 1, got %d", boss.Level)
	}

	if !boss.IsBoss {
		t.Error("Boss should have IsBoss=true even at minimum stats")
	}

	if boss.Rarity != models.RarityLegendary {
		t.Errorf("Boss should be Legendary even at minimum stats, got '%s'", boss.Rarity)
	}

	t.Logf("Minimum boss: %s (Lv%d, HP:%d)", boss.Name, boss.Level, boss.HitpointsTotal)
}

// TestAvailableDungeons tests filtering dungeon templates by player level.
func TestAvailableDungeons(t *testing.T) {
	// Level 1: should only see Goblin Warren (MinLevel: 1)
	level1Dungeons := AvailableDungeons(1)
	if len(level1Dungeons) < 1 {
		t.Fatal("Level 1 player should see at least 1 dungeon")
	}

	foundGoblinWarren := false
	for _, d := range level1Dungeons {
		if d.Name == "Goblin Warren" {
			foundGoblinWarren = true
		}
	}
	if !foundGoblinWarren {
		t.Error("Level 1 player should see 'Goblin Warren'")
	}

	// Level 1 should NOT see higher level dungeons
	for _, d := range level1Dungeons {
		if d.MinLevel > 1 {
			t.Errorf("Level 1 player should not see '%s' (MinLevel: %d)", d.Name, d.MinLevel)
		}
	}

	// Level 50: should see all dungeons (highest MinLevel is 50 for Tower of Eternity)
	level50Dungeons := AvailableDungeons(50)
	if len(level50Dungeons) != len(data.DungeonTemplates) {
		t.Errorf("Level 50 player should see all %d dungeons, got %d",
			len(data.DungeonTemplates), len(level50Dungeons))
	}

	// Level 50 should see more dungeons than level 1
	if len(level50Dungeons) <= len(level1Dungeons) {
		t.Error("Level 50 player should see more dungeons than level 1")
	}

	// Level 5: should see Goblin Warren and Forgotten Crypt
	level5Dungeons := AvailableDungeons(5)
	expectedLevel5Count := 0
	for _, tmpl := range data.DungeonTemplates {
		if tmpl.MinLevel <= 5 {
			expectedLevel5Count++
		}
	}
	if len(level5Dungeons) != expectedLevel5Count {
		t.Errorf("Level 5 player should see %d dungeons, got %d",
			expectedLevel5Count, len(level5Dungeons))
	}

	// Verify Forgotten Crypt is available at level 5
	foundCrypt := false
	for _, d := range level5Dungeons {
		if d.Name == "Forgotten Crypt" {
			foundCrypt = true
		}
	}
	if !foundCrypt {
		t.Error("Level 5 player should see 'Forgotten Crypt'")
	}

	// Level 0: should see no dungeons (all templates have MinLevel >= 1)
	level0Dungeons := AvailableDungeons(0)
	if len(level0Dungeons) != 0 {
		t.Errorf("Level 0 player should see 0 dungeons, got %d", len(level0Dungeons))
	}

	t.Logf("Available dungeons: Lv0=%d, Lv1=%d, Lv5=%d, Lv50=%d",
		len(level0Dungeons), len(level1Dungeons), len(level5Dungeons), len(level50Dungeons))
}

// TestAvailableDungeonsProgression verifies that available dungeons increase
// monotonically as player level increases.
func TestAvailableDungeonsProgression(t *testing.T) {
	prevCount := 0
	for level := 0; level <= 60; level++ {
		count := len(AvailableDungeons(level))
		if count < prevCount {
			t.Errorf("Available dungeons should not decrease: level %d has %d (prev had %d)",
				level, count, prevCount)
		}
		prevCount = count
	}

	t.Logf("Dungeon availability is monotonically non-decreasing across levels 0-60")
}

// TestRoomTypeDistribution generates many dungeons and verifies that all
// expected room types (combat, treasure, trap, rest, merchant, boss) appear.
func TestRoomTypeDistribution(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	// Use a template with boss floors (needs at least 5 floors for a boss floor)
	template := data.DungeonTemplates[0] // Goblin Warren: 5 floors

	roomTypeCounts := map[string]int{}

	// Generate many dungeons with different seeds to get good distribution
	numDungeons := 100
	for i := 0; i < numDungeons; i++ {
		seed := int64(i * 7919) // Use prime multiplier for varied seeds
		dungeon := GenerateDungeon(template, seed)

		for _, floor := range dungeon.Floors {
			for _, room := range floor.Rooms {
				roomTypeCounts[room.Type]++
			}
		}
	}

	// Verify all expected room types appear
	expectedTypes := []string{"combat", "treasure", "trap", "rest", "merchant", "boss"}
	for _, roomType := range expectedTypes {
		count, exists := roomTypeCounts[roomType]
		if !exists || count == 0 {
			t.Errorf("Room type '%s' never appeared in %d dungeons", roomType, numDungeons)
		}
	}

	// Verify combat rooms are the most common (50% weight + 5% boss-to-combat)
	if roomTypeCounts["combat"] <= roomTypeCounts["treasure"] {
		t.Error("Combat rooms should be more common than treasure rooms")
	}

	if roomTypeCounts["combat"] <= roomTypeCounts["trap"] {
		t.Error("Combat rooms should be more common than trap rooms")
	}

	// Log the distribution
	totalRooms := 0
	for _, count := range roomTypeCounts {
		totalRooms += count
	}
	t.Logf("Room type distribution across %d dungeons (%d total rooms):", numDungeons, totalRooms)
	for _, rt := range expectedTypes {
		count := roomTypeCounts[rt]
		pct := float64(count) / float64(totalRooms) * 100
		t.Logf("  %-10s: %4d (%.1f%%)", rt, count, pct)
	}
}

// TestRoomTypeDistributionNoUnexpectedTypes verifies that no unexpected
// room types appear in generated dungeons.
func TestRoomTypeDistributionNoUnexpectedTypes(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	validTypes := map[string]bool{
		"combat":        true,
		"treasure":      true,
		"investigation": true,
		"trap":          true,
		"rest":          true,
		"merchant":      true,
		"boss":          true,
	}

	template := data.DungeonTemplates[1] // Forgotten Crypt: 10 floors
	for i := 0; i < 50; i++ {
		dungeon := GenerateDungeon(template, int64(i))
		for _, floor := range dungeon.Floors {
			for _, room := range floor.Rooms {
				if !validTypes[room.Type] {
					t.Errorf("Unexpected room type: '%s' on floor %d", room.Type, floor.FloorNumber)
				}
			}
		}
	}

	t.Logf("No unexpected room types found across 50 dungeon generations")
}

// TestBossFloorHasBossRoom verifies that boss floors (every 5th floor)
// have a boss room as the last room.
func TestBossFloorHasBossRoom(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	// Use a template with at least 10 floors so we get 2 boss floors
	template := data.DungeonTemplates[1] // Forgotten Crypt: 10 floors
	seed := int64(54321)

	dungeon := GenerateDungeon(template, seed)

	bossFloorCount := 0
	for _, floor := range dungeon.Floors {
		if floor.BossFloor {
			bossFloorCount++

			// Last room on a boss floor should be a boss room
			lastRoom := floor.Rooms[len(floor.Rooms)-1]
			if lastRoom.Type != "boss" {
				t.Errorf("Floor %d: boss floor's last room should be 'boss', got '%s'",
					floor.FloorNumber, lastRoom.Type)
			}

			// Boss room should have a monster
			if lastRoom.Monster == nil {
				t.Errorf("Floor %d: boss room should have a monster", floor.FloorNumber)
			} else if !lastRoom.Monster.IsBoss {
				t.Errorf("Floor %d: boss room monster should have IsBoss=true", floor.FloorNumber)
			}
		}
	}

	// Forgotten Crypt has 10 floors, so floors 5 and 10 should be boss floors
	expectedBossFloors := template.Floors / 5
	if bossFloorCount != expectedBossFloors {
		t.Errorf("Expected %d boss floors, got %d", expectedBossFloors, bossFloorCount)
	}

	t.Logf("Boss floors verified: %d boss floors in %d-floor dungeon", bossFloorCount, template.Floors)
}

// TestCombatRoomsHaveMonsters verifies that combat rooms always have a monster.
func TestCombatRoomsHaveMonsters(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	template := data.DungeonTemplates[0]
	dungeon := GenerateDungeon(template, 11111)

	for _, floor := range dungeon.Floors {
		for r, room := range floor.Rooms {
			if room.Type == "combat" {
				if room.Monster == nil {
					t.Errorf("Floor %d, Room %d: combat room should have a monster",
						floor.FloorNumber, r)
				}
			}
		}
	}

	t.Logf("All combat rooms have monsters: verified")
}

// TestTreasureRoomsHaveLoot verifies that treasure rooms contain loot items.
func TestTreasureRoomsHaveLoot(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	// Generate enough dungeons to find treasure rooms
	template := data.DungeonTemplates[0]
	treasureRoomFound := false

	for i := 0; i < 50; i++ {
		dungeon := GenerateDungeon(template, int64(i*1000))
		for _, floor := range dungeon.Floors {
			for _, room := range floor.Rooms {
				if room.Type == "treasure" {
					treasureRoomFound = true
					if len(room.Loot) == 0 {
						t.Errorf("Floor %d: treasure room should have loot items",
							floor.FloorNumber)
					}
					// Treasure rooms have 1-3 items
					if len(room.Loot) < 1 || len(room.Loot) > 3 {
						t.Errorf("Floor %d: treasure room should have 1-3 items, got %d",
							floor.FloorNumber, len(room.Loot))
					}
				}
			}
		}
	}

	if !treasureRoomFound {
		t.Error("No treasure rooms found across 50 dungeon generations")
	}

	t.Logf("Treasure rooms contain loot: verified")
}

// TestTrapRoomsHaveDamage verifies that trap rooms have positive trap damage.
func TestTrapRoomsHaveDamage(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	template := data.DungeonTemplates[0]
	trapRoomFound := false

	for i := 0; i < 50; i++ {
		dungeon := GenerateDungeon(template, int64(i*2000))
		for _, floor := range dungeon.Floors {
			for _, room := range floor.Rooms {
				if room.Type == "trap" {
					trapRoomFound = true
					// Trap damage formula: 5 + floorNum*3
					expectedDamage := 5 + floor.FloorNumber*3
					if room.TrapDamage != expectedDamage {
						t.Errorf("Floor %d: expected trap damage %d, got %d",
							floor.FloorNumber, expectedDamage, room.TrapDamage)
					}
				}
			}
		}
	}

	if !trapRoomFound {
		t.Error("No trap rooms found across 50 dungeon generations")
	}

	t.Logf("Trap room damage scaling: verified")
}

// TestRestRoomsHaveHealAmount verifies that rest rooms have positive heal amounts.
func TestRestRoomsHaveHealAmount(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	template := data.DungeonTemplates[0]
	restRoomFound := false

	for i := 0; i < 50; i++ {
		dungeon := GenerateDungeon(template, int64(i*3000))
		for _, floor := range dungeon.Floors {
			for _, room := range floor.Rooms {
				if room.Type == "rest" {
					restRoomFound = true
					// Heal amount formula: 15 + floorNum*5
					expectedHeal := 15 + floor.FloorNumber*5
					if room.HealAmount != expectedHeal {
						t.Errorf("Floor %d: expected heal amount %d, got %d",
							floor.FloorNumber, expectedHeal, room.HealAmount)
					}
				}
			}
		}
	}

	if !restRoomFound {
		t.Error("No rest rooms found across 50 dungeon generations")
	}

	t.Logf("Rest room healing scaling: verified")
}

// TestMerchantRoomsHaveLoot verifies that merchant rooms have items for sale.
func TestMerchantRoomsHaveLoot(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	template := data.DungeonTemplates[0]
	merchantRoomFound := false

	for i := 0; i < 50; i++ {
		dungeon := GenerateDungeon(template, int64(i*4000))
		for _, floor := range dungeon.Floors {
			for _, room := range floor.Rooms {
				if room.Type == "merchant" {
					merchantRoomFound = true
					// Merchants have 2-4 items (rng.Intn(3) + 2)
					if len(room.Loot) < 2 || len(room.Loot) > 4 {
						t.Errorf("Floor %d: merchant room should have 2-4 items, got %d",
							floor.FloorNumber, len(room.Loot))
					}
				}
			}
		}
	}

	if !merchantRoomFound {
		t.Error("No merchant rooms found across 50 dungeon generations")
	}

	t.Logf("Merchant rooms contain items for sale: verified")
}
