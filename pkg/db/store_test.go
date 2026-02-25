package db

import (
	"os"
	"path/filepath"
	"testing"

	"rpg-game/pkg/models"
)

// newTestStore creates a Store backed by a temporary SQLite file.
// The caller should defer the returned cleanup function.
func newTestStore(t *testing.T) (*Store, func()) {
	t.Helper()
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	store, err := NewStore(dbPath)
	if err != nil {
		t.Fatalf("NewStore(%q): %v", dbPath, err)
	}

	cleanup := func() {
		store.Close()
		os.Remove(dbPath)
	}
	return store, cleanup
}

func TestCreateAndGetAccount(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	// Create an account.
	id, err := store.CreateAccount("testuser", "hashed_pw_123")
	if err != nil {
		t.Fatalf("CreateAccount: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive account id, got %d", id)
	}

	// Retrieve by username.
	acct, err := store.GetAccountByUsername("testuser")
	if err != nil {
		t.Fatalf("GetAccountByUsername: %v", err)
	}
	if acct == nil {
		t.Fatal("expected account, got nil")
	}
	if acct.ID != id {
		t.Errorf("expected ID %d, got %d", id, acct.ID)
	}
	if acct.Username != "testuser" {
		t.Errorf("expected username %q, got %q", "testuser", acct.Username)
	}
	if acct.PasswordHash != "hashed_pw_123" {
		t.Errorf("expected password hash %q, got %q", "hashed_pw_123", acct.PasswordHash)
	}
	if acct.CreatedAt.IsZero() {
		t.Error("expected non-zero CreatedAt")
	}

	// Retrieve by ID.
	acct2, err := store.GetAccountByID(id)
	if err != nil {
		t.Fatalf("GetAccountByID: %v", err)
	}
	if acct2 == nil {
		t.Fatal("expected account by ID, got nil")
	}
	if acct2.Username != "testuser" {
		t.Errorf("expected username %q, got %q", "testuser", acct2.Username)
	}

	// Duplicate username should fail.
	_, err = store.CreateAccount("testuser", "other_hash")
	if err == nil {
		t.Error("expected error creating duplicate account, got nil")
	}

	// Non-existent username should return nil.
	notFound, err := store.GetAccountByUsername("noone")
	if err != nil {
		t.Fatalf("GetAccountByUsername for non-existent: %v", err)
	}
	if notFound != nil {
		t.Error("expected nil for non-existent username")
	}
}

func TestSaveLoadCharacter(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	// Create account first.
	accountID, err := store.CreateAccount("player1", "pw")
	if err != nil {
		t.Fatalf("CreateAccount: %v", err)
	}

	char := models.Character{
		Name:               "Hero",
		Level:              5,
		Experience:         400,
		HitpointsTotal:     100,
		HitpointsNatural:   80,
		HitpointsRemaining: 100,
		ManaTotal:          50,
		ManaNatural:        40,
		ManaRemaining:      50,
		StaminaTotal:       60,
		StaminaNatural:     50,
		StaminaRemaining:   60,
		AttackRolls:        2,
		DefenseRolls:       2,
		KnownLocations:     []string{"Training Hall", "Dark Forest"},
		LearnedSkills: []models.Skill{
			{Name: "Fireball", ManaCost: 10, Damage: 25, DamageType: models.Fire},
		},
		EquipmentMap:       map[int]models.Item{0: {Name: "Iron Helm", Slot: 0, CP: 5}},
		ResourceStorageMap: map[string]models.Resource{"Gold": {Name: "Gold", Stock: 100}},
		Resistances:        map[models.DamageType]float64{models.Fire: 0.5},
	}

	// Save character.
	if err := store.SaveCharacter(accountID, char); err != nil {
		t.Fatalf("SaveCharacter: %v", err)
	}

	// Load character back.
	loaded, err := store.LoadCharacter(accountID, "Hero")
	if err != nil {
		t.Fatalf("LoadCharacter: %v", err)
	}

	// Verify fields.
	if loaded.Name != "Hero" {
		t.Errorf("Name: got %q, want %q", loaded.Name, "Hero")
	}
	if loaded.Level != 5 {
		t.Errorf("Level: got %d, want 5", loaded.Level)
	}
	if loaded.HitpointsTotal != 100 {
		t.Errorf("HitpointsTotal: got %d, want 100", loaded.HitpointsTotal)
	}
	if loaded.ManaTotal != 50 {
		t.Errorf("ManaTotal: got %d, want 50", loaded.ManaTotal)
	}
	if len(loaded.KnownLocations) != 2 {
		t.Errorf("KnownLocations length: got %d, want 2", len(loaded.KnownLocations))
	}
	if len(loaded.LearnedSkills) != 1 || loaded.LearnedSkills[0].Name != "Fireball" {
		t.Errorf("LearnedSkills: unexpected value %+v", loaded.LearnedSkills)
	}
	if len(loaded.EquipmentMap) != 1 {
		t.Errorf("EquipmentMap length: got %d, want 1", len(loaded.EquipmentMap))
	}
	if loaded.ResourceStorageMap["Gold"].Stock != 100 {
		t.Errorf("Gold stock: got %d, want 100", loaded.ResourceStorageMap["Gold"].Stock)
	}
	if loaded.Resistances[models.Fire] != 0.5 {
		t.Errorf("Fire resistance: got %f, want 0.5", loaded.Resistances[models.Fire])
	}

	// List characters.
	names, err := store.ListCharacters(accountID)
	if err != nil {
		t.Fatalf("ListCharacters: %v", err)
	}
	if len(names) != 1 || names[0] != "Hero" {
		t.Errorf("ListCharacters: got %v, want [Hero]", names)
	}

	// Update character (upsert).
	char.Level = 6
	if err := store.SaveCharacter(accountID, char); err != nil {
		t.Fatalf("SaveCharacter (update): %v", err)
	}
	loaded2, err := store.LoadCharacter(accountID, "Hero")
	if err != nil {
		t.Fatalf("LoadCharacter after update: %v", err)
	}
	if loaded2.Level != 6 {
		t.Errorf("Level after update: got %d, want 6", loaded2.Level)
	}

	// GetCharacterID.
	charID, err := store.GetCharacterID(accountID, "Hero")
	if err != nil {
		t.Fatalf("GetCharacterID: %v", err)
	}
	if charID <= 0 {
		t.Errorf("expected positive character id, got %d", charID)
	}

	// Delete character.
	if err := store.DeleteCharacter(accountID, "Hero"); err != nil {
		t.Fatalf("DeleteCharacter: %v", err)
	}
	names, err = store.ListCharacters(accountID)
	if err != nil {
		t.Fatalf("ListCharacters after delete: %v", err)
	}
	if len(names) != 0 {
		t.Errorf("expected empty list after delete, got %v", names)
	}
}

func TestSaveLoadLocations(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	locations := map[string]models.Location{
		"Training Hall": {
			Name:      "Training Hall",
			Weight:    60,
			Type:      "Base",
			LevelMax:  5,
			RarityMax: 2,
			Monsters: []models.Monster{
				{
					Name:               "Ooze",
					Level:              1,
					HitpointsTotal:     20,
					HitpointsRemaining: 20,
					MonsterType:        "ooze",
					Resistances:        map[models.DamageType]float64{models.Physical: 0.5, models.Fire: 2.0},
				},
			},
		},
		"Dark Forest": {
			Name:      "Dark Forest",
			Weight:    30,
			Type:      "Mix",
			LevelMax:  15,
			RarityMax: 5,
			Monsters:  []models.Monster{},
		},
	}

	// Save locations.
	if err := store.SaveLocations(locations); err != nil {
		t.Fatalf("SaveLocations: %v", err)
	}

	// Load locations.
	loaded, err := store.LoadLocations()
	if err != nil {
		t.Fatalf("LoadLocations: %v", err)
	}

	if len(loaded) != 2 {
		t.Fatalf("expected 2 locations, got %d", len(loaded))
	}

	th, ok := loaded["Training Hall"]
	if !ok {
		t.Fatal("Training Hall not found in loaded locations")
	}
	if th.Weight != 60 {
		t.Errorf("Training Hall weight: got %d, want 60", th.Weight)
	}
	if th.Type != "Base" {
		t.Errorf("Training Hall type: got %q, want %q", th.Type, "Base")
	}
	if len(th.Monsters) != 1 {
		t.Fatalf("Training Hall monsters: got %d, want 1", len(th.Monsters))
	}
	if th.Monsters[0].MonsterType != "ooze" {
		t.Errorf("monster type: got %q, want %q", th.Monsters[0].MonsterType, "ooze")
	}
	if th.Monsters[0].Resistances[models.Fire] != 2.0 {
		t.Errorf("ooze fire resistance: got %f, want 2.0", th.Monsters[0].Resistances[models.Fire])
	}

	df, ok := loaded["Dark Forest"]
	if !ok {
		t.Fatal("Dark Forest not found in loaded locations")
	}
	if df.LevelMax != 15 {
		t.Errorf("Dark Forest LevelMax: got %d, want 15", df.LevelMax)
	}

	// Overwrite with updated data.
	locations["Training Hall"] = models.Location{
		Name:     "Training Hall",
		Weight:   60,
		Type:     "Base",
		LevelMax: 10,
	}
	if err := store.SaveLocations(locations); err != nil {
		t.Fatalf("SaveLocations (update): %v", err)
	}
	loaded2, err := store.LoadLocations()
	if err != nil {
		t.Fatalf("LoadLocations after update: %v", err)
	}
	if loaded2["Training Hall"].LevelMax != 10 {
		t.Errorf("Training Hall LevelMax after update: got %d, want 10", loaded2["Training Hall"].LevelMax)
	}
}

func TestSaveLoadQuests(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()

	quests := map[string]models.Quest{
		"quest_001": {
			ID:          "quest_001",
			Name:        "Slay the Slime King",
			Description: "Defeat the Slime King in the Training Hall.",
			Type:        "kill",
			Requirement: models.QuestRequirement{
				Type:        "kill",
				TargetValue: 1,
				TargetName:  "Slime King",
			},
			Reward: models.QuestReward{
				Type:  "item",
				Value: "Royal Slime Sword",
				XP:    500,
			},
			Completed: false,
			Active:    true,
		},
		"quest_002": {
			ID:          "quest_002",
			Name:        "Gather Resources",
			Description: "Collect 50 lumber.",
			Type:        "gather",
			Requirement: models.QuestRequirement{
				Type:         "gather",
				TargetValue:  50,
				TargetName:   "Lumber",
				CurrentValue: 10,
			},
			Reward: models.QuestReward{
				Type:  "xp",
				Value: "",
				XP:    200,
			},
			Completed: false,
			Active:    true,
		},
	}

	// Save quests.
	if err := store.SaveQuests(quests); err != nil {
		t.Fatalf("SaveQuests: %v", err)
	}

	// Load quests.
	loaded, err := store.LoadQuests()
	if err != nil {
		t.Fatalf("LoadQuests: %v", err)
	}

	if len(loaded) != 2 {
		t.Fatalf("expected 2 quests, got %d", len(loaded))
	}

	q1, ok := loaded["quest_001"]
	if !ok {
		t.Fatal("quest_001 not found")
	}
	if q1.Name != "Slay the Slime King" {
		t.Errorf("quest name: got %q, want %q", q1.Name, "Slay the Slime King")
	}
	if q1.Requirement.TargetName != "Slime King" {
		t.Errorf("quest target: got %q, want %q", q1.Requirement.TargetName, "Slime King")
	}
	if q1.Reward.XP != 500 {
		t.Errorf("quest reward XP: got %d, want 500", q1.Reward.XP)
	}
	if q1.Active != true {
		t.Error("quest_001 should be active")
	}

	q2, ok := loaded["quest_002"]
	if !ok {
		t.Fatal("quest_002 not found")
	}
	if q2.Requirement.CurrentValue != 10 {
		t.Errorf("quest_002 current value: got %d, want 10", q2.Requirement.CurrentValue)
	}

	// Update a quest.
	q1.Completed = true
	q1.Active = false
	quests["quest_001"] = q1
	if err := store.SaveQuests(quests); err != nil {
		t.Fatalf("SaveQuests (update): %v", err)
	}
	loaded2, err := store.LoadQuests()
	if err != nil {
		t.Fatalf("LoadQuests after update: %v", err)
	}
	if loaded2["quest_001"].Completed != true {
		t.Error("quest_001 should be completed after update")
	}
	if loaded2["quest_001"].Active != false {
		t.Error("quest_001 should be inactive after update")
	}
}
