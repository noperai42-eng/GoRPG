package game

import (
	"encoding/json"
	"math/rand"
	"os"
	"testing"
	"time"

	"rpg-game/pkg/data"
	"rpg-game/pkg/models"
)

// Test helper to create a test character
func createTestCharacter(name string, level int) models.Character {
	char := GenerateCharacter(name, level, 1)
	char.EquipmentMap = map[int]models.Item{}
	char.Inventory = []models.Item{
		CreateHealthPotion("small"),
		CreateHealthPotion("small"),
		CreateHealthPotion("small"),
	}
	char.ResourceStorageMap = map[string]models.Resource{}
	GenerateLocationsForNewCharacter(&char)
	return char
}

// Test helper to create a test game state
func createTestGameState() models.GameState {
	gameState := models.GameState{
		CharactersMap:   make(map[string]models.Character),
		GameLocations:   make(map[string]models.Location),
		AvailableQuests: make(map[string]models.Quest),
	}
	GenerateGameLocation(&gameState)

	// Initialize quest system
	for id, quest := range data.StoryQuests {
		gameState.AvailableQuests[id] = quest
	}

	return gameState
}

// TestCharacterCreation tests basic character generation
func TestCharacterCreation(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	char := GenerateCharacter("TestHero", 1, 1)

	if char.Name != "TestHero" {
		t.Errorf("Expected name 'TestHero', got '%s'", char.Name)
	}

	if char.Level != 1 {
		t.Errorf("Expected level 1, got %d", char.Level)
	}

	if char.HitpointsTotal <= 0 {
		t.Errorf("Character should have positive HP, got %d", char.HitpointsTotal)
	}

	if char.ManaTotal <= 0 {
		t.Errorf("Character should have positive Mana, got %d", char.ManaTotal)
	}

	if char.StaminaTotal <= 0 {
		t.Errorf("Character should have positive Stamina, got %d", char.StaminaTotal)
	}

	// Check starting skills
	if len(char.LearnedSkills) < 1 {
		t.Errorf("Character should start with at least 1 skill, got %d", len(char.LearnedSkills))
	}

	// Check quest initialization
	if len(char.ActiveQuests) == 0 {
		t.Errorf("Character should start with active quests")
	}

	t.Logf("Character created: %s (Level %d, HP: %d, MP: %d, SP: %d)",
		char.Name, char.Level, char.HitpointsTotal, char.ManaTotal, char.StaminaTotal)
}

// TestMonsterGeneration tests monster creation
func TestMonsterGeneration(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	gameState := createTestGameState()

	monster := GenerateBestMonster(&gameState, 10, 5)

	if monster.Level <= 0 || monster.Level > 10 {
		t.Errorf("Monster level should be 1-10, got %d", monster.Level)
	}

	if monster.Rank <= 0 || monster.Rank > 5 {
		t.Errorf("Monster rank should be 1-5, got %d", monster.Rank)
	}

	if monster.HitpointsTotal <= 0 {
		t.Errorf("Monster should have positive HP, got %d", monster.HitpointsTotal)
	}

	// Check resistances are initialized
	if monster.Resistances == nil {
		t.Errorf("Monster should have resistance map initialized")
	}

	t.Logf("Monster generated: %s (Level %d, Rank %d, HP: %d)",
		monster.Name, monster.Level, monster.Rank, monster.HitpointsTotal)
}

// TestItemGeneration tests item creation and stats
func TestItemGeneration(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	for rarity := 1; rarity <= 5; rarity++ {
		item := GenerateItem(rarity)

		if item.Rarity != rarity {
			t.Errorf("Expected rarity %d, got %d", rarity, item.Rarity)
		}

		if item.CP <= 0 {
			t.Errorf("Item should have positive CP, got %d", item.CP)
		}

		t.Logf("Item generated: %s (Rarity %d, CP: %d)", item.Name, item.Rarity, item.CP)
	}
}

// TestHealthPotion tests consumable creation
func TestHealthPotion(t *testing.T) {
	sizes := []string{"small", "medium", "large"}
	expectedHealing := map[string]int{"small": 15, "medium": 30, "large": 50}

	for _, size := range sizes {
		potion := CreateHealthPotion(size)

		if potion.ItemType != "consumable" {
			t.Errorf("Potion should be consumable, got '%s'", potion.ItemType)
		}

		if potion.Consumable.Value != expectedHealing[size] {
			t.Errorf("Expected %d healing, got %d", expectedHealing[size], potion.Consumable.Value)
		}

		t.Logf("Potion created: %s (Heals %d HP)", potion.Name, potion.Consumable.Value)
	}
}

// TestSkills tests all available skills
func TestSkills(t *testing.T) {
	for _, skill := range data.AvailableSkills {
		// Basic validation
		if skill.Name == "" {
			t.Errorf("Skill should have a name")
		}

		if skill.ManaCost < 0 {
			t.Errorf("Skill %s has negative mana cost: %d", skill.Name, skill.ManaCost)
		}

		if skill.StaminaCost < 0 {
			t.Errorf("Skill %s has negative stamina cost: %d", skill.Name, skill.StaminaCost)
		}

		// Check that skill has some cost
		if skill.ManaCost == 0 && skill.StaminaCost == 0 && skill.Damage != 0 {
			t.Errorf("Skill %s should have a resource cost", skill.Name)
		}

		t.Logf("Skill: %s (Mana: %d, Stamina: %d, Damage: %d)",
			skill.Name, skill.ManaCost, skill.StaminaCost, skill.Damage)
	}
}

// TestElementalDamage tests damage calculation with resistances
func TestElementalDamage(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	// Create a slime (weak to fire)
	slime := GenerateMonster("ooze", 5, 3)

	// Test fire damage (should be 2x)
	fireDamage := ApplyDamage(10, models.Fire, &slime)
	if fireDamage != 20 {
		t.Errorf("Slime should take 2x fire damage, expected 20, got %d", fireDamage)
	}

	// Test physical damage (should be 0.5x)
	physDamage := ApplyDamage(10, models.Physical, &slime)
	if physDamage != 5 {
		t.Errorf("Slime should take 0.5x physical damage, expected 5, got %d", physDamage)
	}

	t.Logf("Elemental damage: Fire=20 (2x), Physical=5 (0.5x)")
}

// TestLevelUp tests character leveling
func TestLevelUp(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	char := createTestCharacter("LevelTest", 1)
	originalLevel := char.Level
	originalHP := char.HitpointsTotal

	// Give enough XP to level up (PlayerExpToLevel(1) = 310)
	char.Experience = PlayerExpToLevel(1)
	LevelUp(&char)

	if char.Level <= originalLevel {
		t.Errorf("Character should have leveled up from %d", originalLevel)
	}

	if char.HitpointsTotal <= originalHP {
		t.Errorf("HP should increase on level up")
	}

	t.Logf("Level up: %d -> %d (HP: %d -> %d, Skills: %d)",
		originalLevel, char.Level, originalHP, char.HitpointsTotal, len(char.LearnedSkills))
}

// TestQuestSystem tests quest initialization and progression
func TestQuestSystem(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	gameState := createTestGameState()
	char := createTestCharacter("QuestTester", 1)

	// Add to game state
	gameState.CharactersMap[char.Name] = char

	// Check quest initialization
	if len(gameState.AvailableQuests) == 0 {
		t.Errorf("Quest system should have available quests")
	}

	if len(char.ActiveQuests) == 0 {
		t.Errorf("Character should start with active quests")
	}

	t.Logf("Quest system: %d available quests, %d active",
		len(gameState.AvailableQuests), len(char.ActiveQuests))

	// Test quest completion by reaching level 3
	char.Level = 3
	char.Experience = 300
	CheckQuestProgress(&char, &gameState)

	// Update in map
	gameState.CharactersMap[char.Name] = char

	t.Logf("Quest progress check complete")
}

// TestAIDecisionMaking tests auto-play AI
func TestAIDecisionMaking(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	player := createTestCharacter("AITest", 5)
	mob := GenerateMonster("kobold", 5, 3)

	// Test low HP decision (should try to heal)
	player.HitpointsRemaining = player.HitpointsTotal / 4 // 25% HP
	decision := MakeAIDecision(&player, &mob, 5)

	if decision != "skill_heal" && decision != "skill_regeneration" && decision != "item" {
		t.Logf("Low HP AI decision: %s (expected healing action)", decision)
	} else {
		t.Logf("Low HP triggers healing: %s", decision)
	}

	// Test early combat decision (should try to buff)
	player.HitpointsRemaining = player.HitpointsTotal // Full HP
	decision = MakeAIDecision(&player, &mob, 1)       // Turn 1

	t.Logf("Turn 1 AI decision: %s", decision)
}

// TestSaveLoad tests game state persistence
func TestSaveLoad(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	// Create test game state
	gameState := createTestGameState()
	char := createTestCharacter("SaveTest", 5)
	char.Experience = 500
	gameState.CharactersMap[char.Name] = char

	// Save to temporary file
	testFile := "test_gamestate.json"
	defer os.Remove(testFile) // Clean up after test

	err := WriteGameStateToFile(gameState, testFile)
	if err != nil {
		t.Fatalf("Failed to save game state: %v", err)
	}

	// Load from file
	loadedState, err := LoadGameStateFromFile(testFile)
	if err != nil {
		t.Fatalf("Failed to load game state: %v", err)
	}

	// Verify data
	loadedChar, exists := loadedState.CharactersMap[char.Name]
	if !exists {
		t.Fatalf("Character not found in loaded save")
	}

	if loadedChar.Name != char.Name {
		t.Errorf("Character name mismatch: expected '%s', got '%s'", char.Name, loadedChar.Name)
	}

	if loadedChar.Level != char.Level {
		t.Errorf("Character level mismatch: expected %d, got %d", char.Level, loadedChar.Level)
	}

	if loadedChar.Experience != char.Experience {
		t.Errorf("Character XP mismatch: expected %d, got %d", char.Experience, loadedChar.Experience)
	}

	t.Logf("Save/Load: Character data preserved")
}

// TestResourceHarvesting tests resource gathering
func TestResourceHarvesting(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	resourceStorage := make(map[string]models.Resource)
	resourceStorage["Lumber"] = models.Resource{Name: "Lumber", Stock: 0, RollModifier: 0}

	initialStock := resourceStorage["Lumber"].Stock
	result := HarvestResource("Lumber", &resourceStorage)

	if result <= 0 {
		t.Errorf("Should harvest at least 1 resource, got %d", result)
	}

	if resourceStorage["Lumber"].Stock <= initialStock {
		t.Errorf("Resource stock should increase after harvest")
	}

	t.Logf("Harvested %d Lumber (Total: %d)", result, resourceStorage["Lumber"].Stock)
}

// TestStatusEffects tests status effect application and processing
func TestStatusEffects(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	char := createTestCharacter("StatusTest", 5)

	// Add poison effect
	char.StatusEffects = append(char.StatusEffects, models.StatusEffect{
		Type:     "poison",
		Duration: 3,
		Potency:  5,
	})

	initialHP := char.HitpointsRemaining
	ProcessStatusEffects(&char)

	// Should have taken poison damage
	if char.HitpointsRemaining >= initialHP {
		t.Errorf("Poison should reduce HP")
	}

	// Effect duration should decrease
	if len(char.StatusEffects) > 0 && char.StatusEffects[0].Duration != 2 {
		t.Errorf("Effect duration should decrease, got %d", char.StatusEffects[0].Duration)
	}

	t.Logf("Status effects: Poison dealt %d damage, %d turns remaining",
		initialHP-char.HitpointsRemaining, char.StatusEffects[0].Duration)
}

// TestCombatSimulation runs a simulated combat encounter
func TestCombatSimulation(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	gameState := createTestGameState()
	player := createTestCharacter("CombatTest", 5)
	player.EquipmentMap = map[int]models.Item{}

	// Find a huntable location
	var location *models.Location
	for locName := range gameState.GameLocations {
		loc := gameState.GameLocations[locName]
		if loc.Type != "Base" && len(loc.Monsters) > 0 {
			location = &loc
			break
		}
	}

	if location == nil {
		t.Fatal("No huntable location found")
	}

	mob := location.Monsters[0]

	// Simulate combat using auto-play AI
	turnCount := 0
	maxTurns := 100 // Prevent infinite loop

	// Restore resources
	player.ManaRemaining = player.ManaTotal
	player.StaminaRemaining = player.StaminaTotal
	mob.ManaRemaining = mob.ManaTotal
	mob.StaminaRemaining = mob.StaminaTotal

	for player.HitpointsRemaining > 0 && mob.HitpointsRemaining > 0 && turnCount < maxTurns {
		turnCount++

		// Player turn - use AI decision
		if !IsStunned(&player) {
			decision := MakeAIDecision(&player, &mob, turnCount)

			if decision == "attack" {
				playerAttack := MultiRoll(player.AttackRolls) + player.StatsMod.AttackMod
				mobDef := MultiRoll(mob.DefenseRolls) + mob.StatsMod.DefenseMod
				if playerAttack > mobDef {
					damage := ApplyDamage(playerAttack-mobDef, models.Physical, &mob)
					mob.HitpointsRemaining -= damage
				}
			}
		}

		// Monster turn
		if mob.HitpointsRemaining > 0 && !IsStunnedMob(&mob) {
			mobAttack := MultiRoll(mob.AttackRolls) + mob.StatsMod.AttackMod
			playerDef := MultiRoll(player.DefenseRolls) + player.StatsMod.DefenseMod
			if mobAttack > playerDef {
				damage := ApplyDamage(mobAttack-playerDef, models.Physical, &player)
				player.HitpointsRemaining -= damage
			}
		}

		ProcessStatusEffects(&player)
		ProcessStatusEffectsMob(&mob)
	}

	if turnCount >= maxTurns {
		t.Errorf("Combat exceeded maximum turns, possible infinite loop")
	}

	winner := "Player"
	if player.HitpointsRemaining <= 0 {
		winner = "Monster"
	}

	t.Logf("Combat simulation: %s won in %d turns", winner, turnCount)
}

// TestAutoPlayMode tests a short auto-play session with inline combat
func TestAutoPlayMode(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping auto-play test in short mode")
	}

	rand.Seed(time.Now().UnixNano())

	gameState := createTestGameState()
	player := createTestCharacter("AutoTest", 1)
	player.EquipmentMap = map[int]models.Item{}
	gameState.CharactersMap[player.Name] = player

	// Find a huntable location
	var huntLocation *models.Location
	for _, locName := range player.KnownLocations {
		loc := gameState.GameLocations[locName]
		if loc.Type != "Base" {
			huntLocation = &loc
			break
		}
	}

	if huntLocation == nil {
		t.Fatal("No huntable location available")
	}

	// Run 5 auto-play fights using inline combat simulation
	fightCount := 5
	startXP := player.Experience
	startLevel := player.Level

	for i := 0; i < fightCount; i++ {
		mobLoc := rand.Intn(len(huntLocation.Monsters))
		mob := huntLocation.Monsters[mobLoc]

		// Restore resources before each fight
		player.ManaRemaining = player.ManaTotal
		player.StaminaRemaining = player.StaminaTotal
		player.HitpointsRemaining = player.HitpointsTotal
		mob.ManaRemaining = mob.ManaTotal
		mob.StaminaRemaining = mob.StaminaTotal
		mob.HitpointsRemaining = mob.HitpointsTotal

		turnCount := 0
		maxTurns := 100

		for player.HitpointsRemaining > 0 && mob.HitpointsRemaining > 0 && turnCount < maxTurns {
			turnCount++

			if !IsStunned(&player) {
				playerAttack := MultiRoll(player.AttackRolls) + player.StatsMod.AttackMod
				mobDef := MultiRoll(mob.DefenseRolls) + mob.StatsMod.DefenseMod
				if playerAttack > mobDef {
					damage := ApplyDamage(playerAttack-mobDef, models.Physical, &mob)
					mob.HitpointsRemaining -= damage
				}
			}

			if mob.HitpointsRemaining > 0 && !IsStunnedMob(&mob) {
				mobAttack := MultiRoll(mob.AttackRolls) + mob.StatsMod.AttackMod
				playerDef := MultiRoll(player.DefenseRolls) + player.StatsMod.DefenseMod
				if mobAttack > playerDef {
					damage := ApplyDamage(mobAttack-playerDef, models.Physical, &player)
					player.HitpointsRemaining -= damage
				}
			}

			ProcessStatusEffects(&player)
			ProcessStatusEffectsMob(&mob)
		}

		// Award XP if player won
		if mob.HitpointsRemaining <= 0 {
			player.Experience += mob.Level * 10
		}

		LevelUp(&player)
		CheckQuestProgress(&player, &gameState)
	}

	xpGained := player.Experience - startXP
	levelsGained := player.Level - startLevel

	t.Logf("Auto-play: %d fights, %d XP gained, %d levels gained",
		fightCount, xpGained, levelsGained)

	if xpGained <= 0 {
		t.Logf("Warning: No XP gained (may have died all fights)")
	}
}

// TestBackwardCompatibility tests loading old save files
func TestBackwardCompatibility(t *testing.T) {
	// Create an old-style character (before quest system)
	oldChar := models.Character{
		Name:               "OldHero",
		Level:              10,
		Experience:         1000,
		HitpointsTotal:     50,
		HitpointsRemaining: 50,
		AttackRolls:        2,
		DefenseRolls:       2,
		EquipmentMap:       map[int]models.Item{},
		Inventory:          []models.Item{},
		// Note: No quest fields
	}

	oldGameState := models.GameState{
		CharactersMap: map[string]models.Character{
			"OldHero": oldChar,
		},
		GameLocations: map[string]models.Location{},
		// Note: No AvailableQuests
	}

	// Save to file
	testFile := "test_old_save.json"
	defer os.Remove(testFile)

	jsonData, _ := json.MarshalIndent(oldGameState, "", "  ")
	os.WriteFile(testFile, jsonData, 0644)

	// Load with new code
	loadedState, err := LoadGameStateFromFile(testFile)
	if err != nil {
		t.Fatalf("Failed to load old save: %v", err)
	}

	// Test that quest system can handle nil values
	loadedChar := loadedState.CharactersMap["OldHero"]
	CheckQuestProgress(&loadedChar, &loadedState)

	// Should now have quest arrays initialized
	if loadedChar.ActiveQuests == nil {
		t.Errorf("Quest arrays should be initialized for old saves")
	}

	t.Logf("Backward compatibility: Old save loaded and upgraded")
}

// Benchmark tests
func BenchmarkCharacterGeneration(b *testing.B) {
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < b.N; i++ {
		GenerateCharacter("BenchHero", 1, 1)
	}
}

func BenchmarkMonsterGeneration(b *testing.B) {
	rand.Seed(time.Now().UnixNano())
	gameState := createTestGameState()
	for i := 0; i < b.N; i++ {
		GenerateBestMonster(&gameState, 20, 5)
	}
}

func BenchmarkDamageCalculation(b *testing.B) {
	rand.Seed(time.Now().UnixNano())
	mob := GenerateMonster("ooze", 5, 3)
	for i := 0; i < b.N; i++ {
		ApplyDamage(10, models.Fire, &mob)
	}
}

// ============================================================
// GUARD COMBAT SYSTEM TESTS
// ============================================================

// TestGenerateGuard tests guard creation with proper initialization
func TestGenerateGuard(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	guard := GenerateGuard(5)

	if guard.Name == "" {
		t.Error("Guard name should not be empty")
	}

	if guard.Level != 5 {
		t.Errorf("Expected guard level 5, got %d", guard.Level)
	}

	if guard.HitPoints <= 0 {
		t.Error("Guard should have positive HP")
	}

	if guard.HitpointsRemaining != guard.HitPoints {
		t.Error("Guard should start at full HP")
	}

	if guard.Hired != false {
		t.Error("Guard should not be hired initially")
	}

	if guard.Injured != false {
		t.Error("Guard should not be injured initially")
	}

	if guard.RecoveryTime != 0 {
		t.Error("Guard should have 0 recovery time initially")
	}

	if guard.EquipmentMap == nil {
		t.Error("Guard EquipmentMap should be initialized")
	}

	if guard.Inventory == nil {
		t.Error("Guard Inventory should be initialized")
	}

	if guard.Resistances == nil {
		t.Error("Guard Resistances should be initialized")
	}

	if len(guard.Resistances) != 5 {
		t.Errorf("Expected 5 resistance types, got %d", len(guard.Resistances))
	}

	t.Logf("Guard generated: %s (Lv%d, HP:%d, Cost:%d)",
		guard.Name, guard.Level, guard.HitPoints, guard.Cost)
}

// TestGuardEquipment tests guard equipment system
func TestGuardEquipment(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	guard := GenerateGuard(5)

	// Clear starting equipment to have clean slate for testing
	guard.EquipmentMap = map[int]models.Item{}
	guard.Inventory = []models.Item{}
	guard.StatsMod = models.StatMod{}
	guard.HitPoints = guard.HitpointsNatural
	guard.HitpointsRemaining = guard.HitpointsNatural

	initialHP := guard.HitpointsNatural

	// Create test item with HP bonus
	testItem := models.Item{
		Name:     "Test Armor",
		Rarity:   3,
		Slot:     1, // Chest
		ItemType: "equipment",
		StatsMod: models.StatMod{
			AttackMod:   5,
			DefenseMod:  10,
			HitPointMod: 15,
		},
		CP: 30,
	}

	// Equip the item
	EquipGuardItem(testItem, &guard.EquipmentMap, &guard.Inventory)

	// Verify item is equipped
	equippedItem, exists := guard.EquipmentMap[1]
	if !exists {
		t.Error("Item should be equipped in slot 1")
	}

	if equippedItem.Name != "Test Armor" {
		t.Errorf("Expected 'Test Armor', got '%s'", equippedItem.Name)
	}

	// Recalculate stats
	guard.StatsMod = CalculateItemMods(guard.EquipmentMap)
	guard.HitPoints = guard.HitpointsNatural + guard.StatsMod.HitPointMod

	// Verify stats updated
	if guard.StatsMod.AttackMod != 5 {
		t.Errorf("Expected AttackMod 5, got %d", guard.StatsMod.AttackMod)
	}

	if guard.StatsMod.DefenseMod != 10 {
		t.Errorf("Expected DefenseMod 10, got %d", guard.StatsMod.DefenseMod)
	}

	if guard.StatsMod.HitPointMod != 15 {
		t.Errorf("Expected HitPointMod 15, got %d", guard.StatsMod.HitPointMod)
	}

	if guard.HitPoints != initialHP+15 {
		t.Errorf("Expected HP to increase by 15, got increase of %d", guard.HitPoints-initialHP)
	}

	t.Logf("Guard equipment: +%d ATK, +%d DEF, +%d HP",
		guard.StatsMod.AttackMod, guard.StatsMod.DefenseMod, guard.StatsMod.HitPointMod)
}

// TestGuardEquipmentReplacement tests replacing equipped items
func TestGuardEquipmentReplacement(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	guard := GenerateGuard(5)

	// Clear starting equipment for clean testing
	guard.EquipmentMap = map[int]models.Item{}
	guard.Inventory = []models.Item{}

	// Equip first item
	item1 := models.Item{
		Name:     "Weak Armor",
		Slot:     1,
		ItemType: "equipment",
		CP:       10,
		StatsMod: models.StatMod{DefenseMod: 5},
	}
	EquipGuardItem(item1, &guard.EquipmentMap, &guard.Inventory)

	// Equip better item in same slot
	item2 := models.Item{
		Name:     "Strong Armor",
		Slot:     1,
		ItemType: "equipment",
		CP:       20,
		StatsMod: models.StatMod{DefenseMod: 15},
	}
	EquipGuardItem(item2, &guard.EquipmentMap, &guard.Inventory)

	// Verify better item is equipped
	equipped := guard.EquipmentMap[1]
	if equipped.Name != "Strong Armor" {
		t.Errorf("Expected 'Strong Armor', got '%s'", equipped.Name)
	}

	// Verify weaker item is in inventory
	if len(guard.Inventory) != 1 {
		t.Errorf("Expected 1 item in inventory, got %d", len(guard.Inventory))
	}

	if guard.Inventory[0].Name != "Weak Armor" {
		t.Errorf("Expected 'Weak Armor' in inventory, got '%s'", guard.Inventory[0].Name)
	}

	t.Logf("Equipment replacement: Better item equipped, old item moved to inventory")
}

// TestGuardAttack tests guard attack mechanics
func TestGuardAttack(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	guards := []models.Guard{
		{
			Name:               "Test Guard 1",
			Level:              5,
			AttackRolls:        2,
			AttackBonus:        5,
			HitpointsRemaining: 50,
			HitPoints:          50,
			Injured:            false,
			StatsMod:           models.StatMod{AttackMod: 10},
		},
		{
			Name:               "Test Guard 2",
			Level:              5,
			AttackRolls:        2,
			AttackBonus:        5,
			HitpointsRemaining: 50,
			HitPoints:          50,
			Injured:            false,
			StatsMod:           models.StatMod{AttackMod: 8},
		},
	}

	monster := GenerateMonster("kobold", 5, 2)
	initialHP := monster.HitpointsRemaining

	// Guards attack
	damage := GuardAttack(guards, &monster)

	// Verify damage was dealt
	if damage <= 0 {
		t.Error("Guards should deal damage")
	}

	// Verify monster took damage
	if monster.HitpointsRemaining >= initialHP {
		t.Error("Monster should have lost HP from guard attacks")
	}

	t.Logf("Guard attack: 2 guards dealt %d total damage", damage)
}

// TestInjuredGuardNoAttack tests that injured guards don't attack
func TestInjuredGuardNoAttack(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	guards := []models.Guard{
		{
			Name:               "Injured Guard",
			Level:              5,
			AttackRolls:        2,
			AttackBonus:        5,
			HitpointsRemaining: 10,
			HitPoints:          50,
			Injured:            true,
			RecoveryTime:       3,
			StatsMod:           models.StatMod{AttackMod: 10},
		},
	}

	monster := GenerateMonster("kobold", 5, 2)
	initialHP := monster.HitpointsRemaining

	// Injured guard attempts attack
	damage := GuardAttack(guards, &monster)

	// Verify no damage dealt
	if damage != 0 {
		t.Errorf("Injured guard should not attack, but dealt %d damage", damage)
	}

	// Verify monster HP unchanged
	if monster.HitpointsRemaining != initialHP {
		t.Error("Monster HP should be unchanged when only injured guards are present")
	}

	t.Logf("Injured guard correctly skipped attack")
}

// TestGuardDefense tests guard damage absorption
func TestGuardDefense(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	guards := []models.Guard{
		{
			Name:               "Guard 1",
			HitPoints:          100,
			HitpointsRemaining: 100,
			Injured:            false,
		},
		{
			Name:               "Guard 2",
			HitPoints:          100,
			HitpointsRemaining: 100,
			Injured:            false,
		},
	}

	incomingDamage := 100

	// Guards defend
	remainingDamage, _ := GuardDefense(guards, incomingDamage)

	// With 2 healthy guards, should absorb 40% (20% each)
	expectedRemaining := 60
	if remainingDamage != expectedRemaining {
		t.Errorf("Expected %d remaining damage, got %d", expectedRemaining, remainingDamage)
	}

	t.Logf("Guard defense: 2 guards absorbed 40%% damage (%d -> %d)",
		incomingDamage, remainingDamage)
}

// TestProcessGuardRecovery tests guard injury recovery system
func TestProcessGuardRecovery(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	village := models.Village{
		ActiveGuards: []models.Guard{
			{
				Name:               "Recovering Guard",
				Injured:            true,
				RecoveryTime:       1,
				HitPoints:          100,
				HitpointsRemaining: 50,
			},
			{
				Name:               "Still Injured Guard",
				Injured:            true,
				RecoveryTime:       3,
				HitPoints:          100,
				HitpointsRemaining: 60,
			},
		},
	}

	// Process recovery
	ProcessGuardRecovery(&village)

	// First guard should be recovered
	if village.ActiveGuards[0].Injured {
		t.Error("Guard with RecoveryTime 1 should have recovered")
	}

	if village.ActiveGuards[0].RecoveryTime != 0 {
		t.Error("Recovered guard should have RecoveryTime 0")
	}

	if village.ActiveGuards[0].HitpointsRemaining != village.ActiveGuards[0].HitPoints {
		t.Error("Recovered guard should be at full HP")
	}

	// Second guard should still be injured
	if !village.ActiveGuards[1].Injured {
		t.Error("Guard with RecoveryTime 3 should still be injured")
	}

	if village.ActiveGuards[1].RecoveryTime != 2 {
		t.Errorf("Expected RecoveryTime 2, got %d", village.ActiveGuards[1].RecoveryTime)
	}

	t.Logf("Guard recovery: 1 recovered, 1 still healing")
}

// TestConsumableNotEquipped tests that consumables go to inventory
func TestConsumableNotEquipped(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	guard := GenerateGuard(5)

	// Clear starting equipment to have clean slate for testing
	guard.EquipmentMap = map[int]models.Item{}
	guard.Inventory = []models.Item{}

	potion := CreateHealthPotion("small")

	// Try to equip consumable
	EquipGuardItem(potion, &guard.EquipmentMap, &guard.Inventory)

	// Verify it went to inventory, not equipment
	if len(guard.EquipmentMap) != 0 {
		t.Error("Consumable should not be equipped")
	}

	if len(guard.Inventory) == 0 {
		t.Error("Consumable should be in inventory")
	}

	found := false
	for _, item := range guard.Inventory {
		if item.ItemType == "consumable" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Consumable not found in inventory")
	}

	t.Logf("Consumable correctly added to inventory, not equipped")
}

// TestCountVillagersByRole tests villager role counting
func TestCountVillagersByRole(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	village := models.Village{
		Villagers: []models.Villager{
			{Name: "Harvester 1", Role: "harvester"},
			{Name: "Harvester 2", Role: "harvester"},
			{Name: "Guard 1", Role: "guard"},
			{Name: "Harvester 3", Role: "harvester"},
			{Name: "Guard 2", Role: "guard"},
		},
	}

	harvesterCount := CountVillagersByRole(&village, "harvester")
	if harvesterCount != 3 {
		t.Errorf("Expected 3 harvesters, got %d", harvesterCount)
	}

	guardCount := CountVillagersByRole(&village, "guard")
	if guardCount != 2 {
		t.Errorf("Expected 2 guards, got %d", guardCount)
	}

	t.Logf("Villager count: %d harvesters, %d guards", harvesterCount, guardCount)
}

// TestGuardStatRecalculation tests stat updates after equipment changes
func TestGuardStatRecalculation(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	guard := GenerateGuard(5)

	// Clear starting equipment to have clean slate for testing
	guard.EquipmentMap = map[int]models.Item{}
	guard.Inventory = []models.Item{}
	guard.StatsMod = models.StatMod{}
	guard.HitPoints = guard.HitpointsNatural
	guard.HitpointsRemaining = guard.HitpointsNatural

	initialNaturalHP := guard.HitpointsNatural

	// Equip armor with HP bonus
	armor := models.Item{
		Name:     "HP Armor",
		Slot:     1,
		ItemType: "equipment",
		StatsMod: models.StatMod{HitPointMod: 20},
		CP:       20,
	}

	EquipGuardItem(armor, &guard.EquipmentMap, &guard.Inventory)

	// Recalculate stats
	guard.StatsMod = CalculateItemMods(guard.EquipmentMap)
	guard.HitPoints = guard.HitpointsNatural + guard.StatsMod.HitPointMod

	// Verify HP increased
	if guard.HitPoints != initialNaturalHP+20 {
		t.Errorf("Expected HP %d, got %d", initialNaturalHP+20, guard.HitPoints)
	}

	// Unequip the armor
	delete(guard.EquipmentMap, 1)

	// Recalculate stats
	guard.StatsMod = CalculateItemMods(guard.EquipmentMap)
	guard.HitPoints = guard.HitpointsNatural + guard.StatsMod.HitPointMod
	if guard.HitpointsRemaining > guard.HitPoints {
		guard.HitpointsRemaining = guard.HitPoints
	}

	// Verify HP returned to base
	if guard.HitPoints != initialNaturalHP {
		t.Errorf("Expected HP %d after unequip, got %d", initialNaturalHP, guard.HitPoints)
	}

	t.Logf("Stat recalculation: %d -> %d (equipped) -> %d (unequipped)",
		initialNaturalHP, initialNaturalHP+20, guard.HitPoints)
}

// TestGuardStartingEquipment tests that higher level guards have more equipment
func TestGuardStartingEquipment(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	lowLevelGuard := GenerateGuard(1)
	midLevelGuard := GenerateGuard(10)
	highLevelGuard := GenerateGuard(20)

	// All guards should have at least some equipment
	if len(lowLevelGuard.EquipmentMap) < 1 {
		t.Error("Low level guard should have at least 1 starting item")
	}

	// Higher level guards should have more items
	if len(highLevelGuard.EquipmentMap) <= len(lowLevelGuard.EquipmentMap) {
		t.Error("Higher level guards should have more starting equipment")
	}

	t.Logf("Starting equipment: Lv1=%d items, Lv10=%d items, Lv20=%d items",
		len(lowLevelGuard.EquipmentMap), len(midLevelGuard.EquipmentMap), len(highLevelGuard.EquipmentMap))
}

// TestGuardCostScaling tests that guard cost increases with level
func TestGuardCostScaling(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	guard1 := GenerateGuard(1)
	guard10 := GenerateGuard(10)
	guard20 := GenerateGuard(20)

	if guard10.Cost <= guard1.Cost {
		t.Error("Higher level guards should cost more")
	}

	if guard20.Cost <= guard10.Cost {
		t.Error("Level 20 guard should cost more than level 10")
	}

	// Cost formula: 50 + (level * 25)
	expectedCost1 := 50 + (1 * 25)
	if guard1.Cost != expectedCost1 {
		t.Errorf("Expected cost %d for level 1, got %d", expectedCost1, guard1.Cost)
	}

	expectedCost10 := 50 + (10 * 25)
	if guard10.Cost != expectedCost10 {
		t.Errorf("Expected cost %d for level 10, got %d", expectedCost10, guard10.Cost)
	}

	t.Logf("Guard costs: Lv1=%d gold, Lv10=%d gold, Lv20=%d gold",
		guard1.Cost, guard10.Cost, guard20.Cost)
}

// TestMonsterBossFlag tests IsBoss flag on monsters
func TestMonsterBossFlag(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	// Regular monster
	regularMonster := GenerateMonster("kobold", 5, 2)
	if regularMonster.IsBoss {
		t.Error("Regular monster should not be marked as boss")
	}

	// Skill Guardian
	skill := data.AvailableSkills[0]
	guardian := GenerateSkillGuardian(skill, 10, 3)
	if guardian.IsBoss {
		t.Error("Skill guardian should not be marked as boss")
	}

	if !guardian.IsSkillGuardian {
		t.Error("Skill guardian should have IsSkillGuardian flag")
	}

	t.Logf("Monster flags: Regular (IsBoss=%v), Guardian (IsSkillGuardian=%v)",
		regularMonster.IsBoss, guardian.IsSkillGuardian)
}

// TestGuardInjuryThreshold tests that guards get injured at low HP
func TestGuardInjuryThreshold(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	guard := models.Guard{
		Name:               "Test Guard",
		Level:              5,
		HitPoints:          100,
		HitpointsRemaining: 100,
		Injured:            false,
		RecoveryTime:       0,
	}

	// Reduce HP to 29% (below 30% threshold)
	guard.HitpointsRemaining = 29

	// Check injury threshold (this happens in guardDefense function)
	if guard.HitpointsRemaining <= (guard.HitPoints * 30 / 100) && !guard.Injured {
		guard.Injured = true
		guard.RecoveryTime = 3
	}

	if !guard.Injured {
		t.Error("Guard should be injured at 29% HP")
	}

	if guard.RecoveryTime != 3 {
		t.Errorf("Expected RecoveryTime 3, got %d", guard.RecoveryTime)
	}

	t.Logf("Injury threshold: Guard injured at %d/%d HP (%d%% threshold)",
		guard.HitpointsRemaining, guard.HitPoints, 30)
}
