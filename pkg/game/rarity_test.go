package game

import (
	"testing"

	"rpg-game/pkg/models"
)

func TestRollRarityDistribution(t *testing.T) {
	iterations := 10000
	counts := map[models.MonsterRarity]int{}

	for i := 0; i < iterations; i++ {
		r := RollRarity(5) // rankMax=5 allows all rarities up to mythic
		counts[r]++
	}

	commonPct := float64(counts[models.RarityCommon]) / float64(iterations) * 100
	legendaryPct := float64(counts[models.RarityLegendary]) / float64(iterations) * 100

	if commonPct <= 20.0 {
		t.Errorf("Common should be >20%% of rolls, got %.2f%% (%d/%d)",
			commonPct, counts[models.RarityCommon], iterations)
	}

	if legendaryPct >= 5.0 {
		t.Errorf("Legendary should be <5%% of rolls, got %.2f%% (%d/%d)",
			legendaryPct, counts[models.RarityLegendary], iterations)
	}

	// Verify common through legendary were rolled at least once in 10k iterations
	// (mythic is 1/100,000 so it almost certainly won't appear in 10k rolls)
	for _, rarity := range rarityOrder[:5] {
		if counts[rarity] == 0 {
			t.Errorf("Expected at least one roll of %s in %d iterations", rarity, iterations)
		}
	}

	// Verify rankMax caps work: rankMax=1 should only produce common/uncommon
	cappedCounts := map[models.MonsterRarity]int{}
	for i := 0; i < iterations; i++ {
		r := RollRarity(1)
		cappedCounts[r]++
	}
	for _, rarity := range rarityOrder[2:] { // rare, epic, legendary, mythic
		if cappedCounts[rarity] > 0 {
			t.Errorf("RollRarity(1) should never produce %s, got %d", rarity, cappedCounts[rarity])
		}
	}
}

func TestApplyRarity(t *testing.T) {
	tests := []struct {
		rarity     models.MonsterRarity
		wantHPMult float64
		wantAtk    int
		wantDef    int
	}{
		{models.RarityCommon, 1.0, 2, 3},       // Common: no change (ApplyRarity returns early)
		{models.RarityUncommon, 10.0, 4, 6},     // AtkRolls*2, DefRolls*2
		{models.RarityRare, 50.0, 8, 9},         // AtkRolls*4, DefRolls*3
		{models.RarityEpic, 250.0, 12, 15},      // AtkRolls*6, DefRolls*5
		{models.RarityLegendary, 1000.0, 20, 24}, // AtkRolls*10, DefRolls*8
	}

	for _, tc := range tests {
		t.Run(string(tc.rarity), func(t *testing.T) {
			mob := &models.Monster{
				HitpointsNatural:   100,
				HitpointsTotal:     100,
				HitpointsRemaining: 100,
				ManaNatural:        50,
				ManaTotal:          50,
				ManaRemaining:      50,
				StaminaNatural:     50,
				StaminaTotal:       50,
				StaminaRemaining:   50,
				AttackRolls:        2,
				DefenseRolls:       3,
				Rarity:             tc.rarity,
			}

			ApplyRarity(mob)

			if tc.rarity == models.RarityCommon {
				// Common returns early, no changes applied
				if mob.HitpointsNatural != 100 {
					t.Errorf("Common HP should be unchanged (100), got %d", mob.HitpointsNatural)
				}
				if mob.AttackRolls != 2 {
					t.Errorf("Common AttackRolls should be unchanged (2), got %d", mob.AttackRolls)
				}
				if mob.DefenseRolls != 3 {
					t.Errorf("Common DefenseRolls should be unchanged (3), got %d", mob.DefenseRolls)
				}
			} else {
				expectedHP := int(100.0 * tc.wantHPMult)
				if mob.HitpointsNatural != expectedHP {
					t.Errorf("HitpointsNatural: want %d, got %d", expectedHP, mob.HitpointsNatural)
				}
				if mob.HitpointsTotal != expectedHP {
					t.Errorf("HitpointsTotal: want %d, got %d", expectedHP, mob.HitpointsTotal)
				}
				if mob.HitpointsRemaining != expectedHP {
					t.Errorf("HitpointsRemaining: want %d, got %d", expectedHP, mob.HitpointsRemaining)
				}
				if mob.AttackRolls != tc.wantAtk {
					t.Errorf("AttackRolls: want %d, got %d", tc.wantAtk, mob.AttackRolls)
				}
				if mob.DefenseRolls != tc.wantDef {
					t.Errorf("DefenseRolls: want %d, got %d", tc.wantDef, mob.DefenseRolls)
				}
			}
		})
	}
}

func TestRarityXPMult(t *testing.T) {
	tests := []struct {
		rarity models.MonsterRarity
		want   float64
	}{
		{models.RarityCommon, 1.0},
		{models.RarityUncommon, 5.0},
		{models.RarityRare, 25.0},
		{models.RarityEpic, 100.0},
		{models.RarityLegendary, 500.0},
	}

	for _, tc := range tests {
		t.Run(string(tc.rarity), func(t *testing.T) {
			got := RarityXPMult(tc.rarity)
			if got != tc.want {
				t.Errorf("RarityXPMult(%s) = %f, want %f", tc.rarity, got, tc.want)
			}
		})
	}

	// Unknown rarity should fall back to 1.0
	got := RarityXPMult(models.MonsterRarity("mythical"))
	if got != 1.0 {
		t.Errorf("RarityXPMult(unknown) = %f, want 1.0", got)
	}
}

func TestRarityLootBonus(t *testing.T) {
	tests := []struct {
		rarity models.MonsterRarity
		want   int
	}{
		{models.RarityCommon, 0},
		{models.RarityUncommon, 2},
		{models.RarityRare, 4},
		{models.RarityEpic, 6},
		{models.RarityLegendary, 8},
	}

	for _, tc := range tests {
		t.Run(string(tc.rarity), func(t *testing.T) {
			got := RarityLootBonus(tc.rarity)
			if got != tc.want {
				t.Errorf("RarityLootBonus(%s) = %d, want %d", tc.rarity, got, tc.want)
			}
		})
	}

	// Unknown rarity should fall back to 0
	got := RarityLootBonus(models.MonsterRarity("mythical"))
	if got != 0 {
		t.Errorf("RarityLootBonus(unknown) = %d, want 0", got)
	}
}

func TestNormalizeRarity(t *testing.T) {
	// Empty string should normalize to Common
	got := NormalizeRarity("")
	if got != models.RarityCommon {
		t.Errorf("NormalizeRarity(\"\") = %q, want %q", got, models.RarityCommon)
	}

	// Valid values should pass through unchanged
	validRarities := []models.MonsterRarity{
		models.RarityCommon,
		models.RarityUncommon,
		models.RarityRare,
		models.RarityEpic,
		models.RarityLegendary,
	}

	for _, r := range validRarities {
		got := NormalizeRarity(r)
		if got != r {
			t.Errorf("NormalizeRarity(%q) = %q, want %q", r, got, r)
		}
	}

	// Non-empty unknown value should pass through as-is (no normalization)
	unknown := models.MonsterRarity("mythical")
	got = NormalizeRarity(unknown)
	if got != unknown {
		t.Errorf("NormalizeRarity(%q) = %q, want %q", unknown, got, unknown)
	}
}

func TestRarityDisplayName(t *testing.T) {
	tests := []struct {
		rarity models.MonsterRarity
		want   string
	}{
		{models.RarityCommon, "Common"},
		{models.RarityUncommon, "Uncommon"},
		{models.RarityRare, "Rare"},
		{models.RarityEpic, "Epic"},
		{models.RarityLegendary, "Legendary"},
		{"", "Common"}, // empty string should display as Common
	}

	for _, tc := range tests {
		name := string(tc.rarity)
		if name == "" {
			name = "empty"
		}
		t.Run(name, func(t *testing.T) {
			got := RarityDisplayName(tc.rarity)
			if got != tc.want {
				t.Errorf("RarityDisplayName(%q) = %q, want %q", tc.rarity, got, tc.want)
			}
		})
	}

	// Unknown rarity should default to "Common"
	got := RarityDisplayName(models.MonsterRarity("mythical"))
	if got != "Common" {
		t.Errorf("RarityDisplayName(unknown) = %q, want \"Common\"", got)
	}
}
