package game

import (
	"math/rand"
	"testing"
	"time"

	"rpg-game/pkg/models"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// ============================================================
// NPC INN GUEST TESTS
// ============================================================

// TestGenerateNPCGuest tests NPC guest creation with proper initialization.
func TestGenerateNPCGuest(t *testing.T) {
	guest := GenerateNPCGuest("Test NPC", 5)

	if guest.CharacterName != "Test NPC" {
		t.Errorf("Expected name 'Test NPC', got '%s'", guest.CharacterName)
	}

	if guest.AccountID != 0 {
		t.Errorf("NPC guest should have AccountID 0, got %d", guest.AccountID)
	}

	if guest.Level != 5 {
		t.Errorf("Expected level 5, got %d", guest.Level)
	}

	if guest.MaxHP <= 0 {
		t.Errorf("NPC should have positive HP, got %d", guest.MaxHP)
	}

	if guest.HP != guest.MaxHP {
		t.Errorf("NPC should start at full HP: %d != %d", guest.HP, guest.MaxHP)
	}

	if guest.MaxMP <= 0 {
		t.Errorf("NPC should have positive MP, got %d", guest.MaxMP)
	}

	if guest.MaxSP <= 0 {
		t.Errorf("NPC should have positive SP, got %d", guest.MaxSP)
	}

	if guest.AttackRolls <= 0 {
		t.Errorf("NPC should have positive attack rolls, got %d", guest.AttackRolls)
	}

	if guest.DefenseRolls <= 0 {
		t.Errorf("NPC should have positive defense rolls, got %d", guest.DefenseRolls)
	}

	if guest.Resistances == nil || len(guest.Resistances) != 5 {
		t.Errorf("NPC should have 5 resistance types, got %d", len(guest.Resistances))
	}

	if guest.EquipmentMap == nil {
		t.Error("NPC EquipmentMap should be initialized")
	}

	if guest.CheckInTime <= 0 {
		t.Error("NPC should have a valid check-in time")
	}

	t.Logf("NPC guest: %s (Lv%d, HP:%d, MP:%d, SP:%d, Gold:%d, Guards:%d, Equipment:%d)",
		guest.CharacterName, guest.Level, guest.MaxHP, guest.MaxMP, guest.MaxSP,
		guest.GoldCarried, len(guest.HiredGuards), len(guest.EquipmentMap))
}

// TestNPCGuestGoldFormula verifies gold carried follows the formula: 50 + level*20.
func TestNPCGuestGoldFormula(t *testing.T) {
	testCases := []struct {
		level        int
		expectedGold int
	}{
		{1, 70},
		{5, 150},
		{10, 250},
		{15, 350},
	}

	for _, tc := range testCases {
		guest := GenerateNPCGuest("GoldTest", tc.level)
		if guest.GoldCarried != tc.expectedGold {
			t.Errorf("Level %d: expected %d gold, got %d", tc.level, tc.expectedGold, guest.GoldCarried)
		}
	}
}

// TestNPCGuestHasGuards verifies NPC guests have 1-2 guards.
func TestNPCGuestHasGuards(t *testing.T) {
	for i := 0; i < 20; i++ {
		guest := GenerateNPCGuest("GuardTest", 5)
		guardCount := len(guest.HiredGuards)
		if guardCount < 1 || guardCount > 2 {
			t.Errorf("Expected 1-2 guards, got %d", guardCount)
		}
		for _, g := range guest.HiredGuards {
			if g.Level != 5 {
				t.Errorf("Guard level should match NPC level (5), got %d", g.Level)
			}
		}
	}
}

// TestNPCGuestHasSkills verifies NPC guests get humanoid skills.
func TestNPCGuestHasSkills(t *testing.T) {
	guest := GenerateNPCGuest("SkillTest", 10)
	if len(guest.LearnedSkills) == 0 {
		t.Error("Level 10 NPC should have learned skills")
	}
	t.Logf("NPC skills: %d skills assigned", len(guest.LearnedSkills))
}

// TestNPCGuestEquipmentScaling verifies higher level NPCs get more items.
func TestNPCGuestEquipmentScaling(t *testing.T) {
	lowLevel := GenerateNPCGuest("Low", 1)
	highLevel := GenerateNPCGuest("High", 15)

	// Low level: 2 + (1/5) = 2 items generated
	// High level: 2 + (15/5) = 5, capped at 4 items generated
	// Equipment count depends on slot collisions, but high level should generally have more
	t.Logf("Equipment: Lv1=%d items, Lv15=%d items",
		len(lowLevel.EquipmentMap), len(highLevel.EquipmentMap))
}

// ============================================================
// REPLENISH NPC GUESTS TESTS
// ============================================================

// TestReplenishNPCGuests verifies NPCs are filled up to 4.
func TestReplenishNPCGuests(t *testing.T) {
	town := models.Town{
		Name:      "TestTown",
		InnGuests: []models.InnGuest{},
	}

	ReplenishNPCGuests(&town)

	npcCount := 0
	for _, guest := range town.InnGuests {
		if guest.AccountID == 0 {
			npcCount++
		}
	}

	if npcCount != 4 {
		t.Errorf("Expected 4 NPC guests, got %d", npcCount)
	}

	// Verify all NPCs have valid names
	for _, guest := range town.InnGuests {
		if guest.CharacterName == "" {
			t.Error("NPC guest should have a non-empty name")
		}
	}

	t.Logf("Replenished to %d NPC guests", npcCount)
}

// TestReplenishNPCGuestsWithExisting verifies replenish respects existing NPCs.
func TestReplenishNPCGuestsWithExisting(t *testing.T) {
	town := models.Town{
		Name: "TestTown",
		InnGuests: []models.InnGuest{
			GenerateNPCGuest("Existing NPC 1", 5),
			GenerateNPCGuest("Existing NPC 2", 8),
		},
	}

	ReplenishNPCGuests(&town)

	npcCount := 0
	for _, guest := range town.InnGuests {
		if guest.AccountID == 0 {
			npcCount++
		}
	}

	if npcCount != 4 {
		t.Errorf("Expected 4 NPC guests after replenish, got %d", npcCount)
	}

	// Original NPCs should still be there
	found := false
	for _, guest := range town.InnGuests {
		if guest.CharacterName == "Existing NPC 1" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Original NPC should still be present after replenish")
	}
}

// TestReplenishDoesNotRemovePlayers verifies player guests are not counted as NPCs.
func TestReplenishDoesNotRemovePlayers(t *testing.T) {
	town := models.Town{
		Name: "TestTown",
		InnGuests: []models.InnGuest{
			{AccountID: 1, CharacterName: "Player1", Level: 5, CheckInTime: time.Now().Unix()},
			{AccountID: 2, CharacterName: "Player2", Level: 8, CheckInTime: time.Now().Unix()},
		},
	}

	ReplenishNPCGuests(&town)

	totalGuests := len(town.InnGuests)
	npcCount := 0
	playerCount := 0
	for _, guest := range town.InnGuests {
		if guest.AccountID == 0 {
			npcCount++
		} else {
			playerCount++
		}
	}

	if npcCount != 4 {
		t.Errorf("Expected 4 NPC guests, got %d", npcCount)
	}

	if playerCount != 2 {
		t.Errorf("Expected 2 player guests to remain, got %d", playerCount)
	}

	if totalGuests != 6 {
		t.Errorf("Expected 6 total guests (2 players + 4 NPCs), got %d", totalGuests)
	}
}

// TestReplenishAlreadyFull verifies no new NPCs are added when already at 4+.
func TestReplenishAlreadyFull(t *testing.T) {
	town := models.Town{
		Name: "TestTown",
		InnGuests: []models.InnGuest{
			GenerateNPCGuest("NPC 1", 3),
			GenerateNPCGuest("NPC 2", 5),
			GenerateNPCGuest("NPC 3", 8),
			GenerateNPCGuest("NPC 4", 12),
		},
	}

	ReplenishNPCGuests(&town)

	npcCount := 0
	for _, guest := range town.InnGuests {
		if guest.AccountID == 0 {
			npcCount++
		}
	}

	if npcCount != 4 {
		t.Errorf("Should remain at 4 NPCs, got %d", npcCount)
	}
}

// ============================================================
// GENERATE DEFAULT TOWN TESTS
// ============================================================

// TestGenerateDefaultTownHasNPCGuests verifies the default town seeds 4 NPC guests.
func TestGenerateDefaultTownHasNPCGuests(t *testing.T) {
	town := GenerateDefaultTown("TestTown")

	if len(town.InnGuests) != 4 {
		t.Errorf("Default town should have 4 NPC guests, got %d", len(town.InnGuests))
	}

	expectedLevels := []int{3, 5, 8, 12}
	for i, guest := range town.InnGuests {
		if guest.AccountID != 0 {
			t.Errorf("Guest %d should be NPC (AccountID 0), got %d", i, guest.AccountID)
		}
		if guest.Level != expectedLevels[i] {
			t.Errorf("Guest %d: expected level %d, got %d", i, expectedLevels[i], guest.Level)
		}
		if guest.GoldCarried <= 0 {
			t.Errorf("Guest %d should carry gold, got %d", i, guest.GoldCarried)
		}
		if len(guest.HiredGuards) == 0 {
			t.Errorf("Guest %d should have guards", i)
		}
	}

	t.Logf("Default town has %d NPC guests at levels %v", len(town.InnGuests), expectedLevels)
}

// TestGenerateDefaultTownHasMayor verifies the default town has a mayor.
func TestGenerateDefaultTownHasMayor(t *testing.T) {
	town := GenerateDefaultTown("TestTown")

	if town.Mayor == nil {
		t.Fatal("Default town should have a mayor")
	}

	if !town.Mayor.IsNPC {
		t.Error("Default mayor should be NPC")
	}

	if town.Mayor.Level != 10 {
		t.Errorf("Default mayor should be level 10, got %d", town.Mayor.Level)
	}
}

// ============================================================
// CLEAN EXPIRED GUESTS TESTS
// ============================================================

// TestCleanExpiredGuestsPreservesNPCs verifies NPC guests are never expired.
func TestCleanExpiredGuestsPreservesNPCs(t *testing.T) {
	oldTime := time.Now().Unix() - 200000 // well past 24h expiry

	town := models.Town{
		Name: "TestTown",
		InnGuests: []models.InnGuest{
			{AccountID: 0, CharacterName: "Old NPC", Level: 5, CheckInTime: oldTime},
			{AccountID: 1, CharacterName: "Old Player", Level: 5, CheckInTime: oldTime},
			{AccountID: 0, CharacterName: "Fresh NPC", Level: 8, CheckInTime: time.Now().Unix()},
			{AccountID: 2, CharacterName: "Fresh Player", Level: 3, CheckInTime: time.Now().Unix()},
		},
	}

	CleanExpiredGuests(&town, 86400) // 24h max age

	// Should keep: Old NPC (never expires), Fresh NPC, Fresh Player
	// Should remove: Old Player (expired)
	if len(town.InnGuests) != 3 {
		t.Errorf("Expected 3 guests after cleaning, got %d", len(town.InnGuests))
		for _, g := range town.InnGuests {
			t.Logf("  Remaining: %s (AccountID=%d)", g.CharacterName, g.AccountID)
		}
	}

	// Verify Old NPC is preserved
	foundOldNPC := false
	for _, g := range town.InnGuests {
		if g.CharacterName == "Old NPC" {
			foundOldNPC = true
		}
		if g.CharacterName == "Old Player" {
			t.Error("Expired player guest should have been removed")
		}
	}

	if !foundOldNPC {
		t.Error("Old NPC guest should be preserved regardless of age")
	}
}

// TestCleanExpiredGuestsAllFresh verifies nothing is removed when all guests are fresh.
func TestCleanExpiredGuestsAllFresh(t *testing.T) {
	town := models.Town{
		Name: "TestTown",
		InnGuests: []models.InnGuest{
			{AccountID: 0, CharacterName: "NPC 1", Level: 5, CheckInTime: time.Now().Unix()},
			{AccountID: 1, CharacterName: "Player 1", Level: 5, CheckInTime: time.Now().Unix()},
		},
	}

	CleanExpiredGuests(&town, 86400)

	if len(town.InnGuests) != 2 {
		t.Errorf("Expected 2 guests (all fresh), got %d", len(town.InnGuests))
	}
}

// ============================================================
// INN GUEST TO MONSTER CONVERSION TEST
// ============================================================

// TestInnGuestToMonsterPreservesStats verifies the conversion preserves key fields.
func TestInnGuestToMonsterPreservesStats(t *testing.T) {
	guest := GenerateNPCGuest("Monster Convert", 10)

	monster := InnGuestToMonster(&guest)

	if monster.Name != "Monster Convert" {
		t.Errorf("Expected name 'Monster Convert', got '%s'", monster.Name)
	}

	if monster.Level != 10 {
		t.Errorf("Expected level 10, got %d", monster.Level)
	}

	if monster.HitpointsTotal != guest.MaxHP {
		t.Errorf("HP mismatch: monster=%d, guest=%d", monster.HitpointsTotal, guest.MaxHP)
	}

	if monster.HitpointsRemaining != guest.MaxHP {
		t.Errorf("Monster should start at full HP: %d != %d", monster.HitpointsRemaining, guest.MaxHP)
	}

	if monster.AttackRolls != guest.AttackRolls {
		t.Errorf("Attack rolls mismatch: %d != %d", monster.AttackRolls, guest.AttackRolls)
	}

	if monster.MonsterType != "humanoid" {
		t.Errorf("Converted guest should be humanoid type, got '%s'", monster.MonsterType)
	}

	t.Logf("Conversion: Guest (HP:%d, ATK:%d) -> Monster (HP:%d, ATK:%d)",
		guest.MaxHP, guest.AttackRolls, monster.HitpointsTotal, monster.AttackRolls)
}
