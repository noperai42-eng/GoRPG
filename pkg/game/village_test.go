package game

import (
	"strings"
	"testing"

	"rpg-game/pkg/models"
)

// TestAutoTideDefeatResetsVillage verifies that a lost auto-tide resets the
// village to level 1, kills all guards/villagers, and destroys all structures.
func TestAutoTideDefeatResetsVillage(t *testing.T) {
	// Level 20 with no guards/defenses: 5 waves of 12 monsters, each dealing
	// ~22-42 breach damage. Total breach damage far exceeds threshold of 80.
	village := models.Village{
		Name:         "TestVillage",
		Level:        20,
		Experience:   200,
		DefenseLevel: 1,
		Villagers: []models.Villager{
			{Name: "Alice", Role: "harvester", HarvestType: "Gold"},
			{Name: "Bob", Role: "harvester", HarvestType: "Iron"},
		},
		ActiveGuards: []models.Guard{}, // No guards = guaranteed defeat
		Defenses:     []models.Defense{},
		Traps:        []models.Trap{},
		TideInterval: 3600,
	}

	player := models.Character{
		Name: "TestPlayer",
		ResourceStorageMap: map[string]models.Resource{
			"Gold":   {Name: "Gold", Stock: 100},
			"Lumber": {Name: "Lumber", Stock: 50},
			"Iron":   {Name: "Iron", Stock: 30},
			"Sand":   {Name: "Sand", Stock: 20},
			"Stone":  {Name: "Stone", Stock: 10},
		},
	}

	result := ProcessAutoTide(&village, &player)

	// With no guards, no defenses, and no traps at level 5, defeat is certain
	if result.Victory {
		t.Fatal("expected defeat with no defenses, got victory")
	}

	// Village must be reset to level 1
	if village.Level != 1 {
		t.Errorf("expected village level 1 after defeat, got %d", village.Level)
	}
	if village.Experience != 0 {
		t.Errorf("expected village XP 0 after defeat, got %d", village.Experience)
	}
	if village.DefenseLevel != 1 {
		t.Errorf("expected defense level 1 after defeat, got %d", village.DefenseLevel)
	}

	// All villagers, guards, defenses, and traps destroyed
	if len(village.Villagers) != 0 {
		t.Errorf("expected 0 villagers after defeat, got %d", len(village.Villagers))
	}
	if len(village.ActiveGuards) != 0 {
		t.Errorf("expected 0 guards after defeat, got %d", len(village.ActiveGuards))
	}
	if len(village.Defenses) != 0 {
		t.Errorf("expected 0 defenses after defeat, got %d", len(village.Defenses))
	}
	if len(village.Traps) != 0 {
		t.Errorf("expected 0 traps after defeat, got %d", len(village.Traps))
	}

	// Result should report all losses
	if result.VillagersLost != 2 {
		t.Errorf("expected 2 villagers lost, got %d", result.VillagersLost)
	}

	// Messages should contain detailed narration
	fullLog := strings.Join(result.Messages, "\n")
	for _, expected := range []string{
		"--- Monster Tide on TestVillage",
		"Defenders:",
		"-- Wave",
		"monsters charge!",
		"-- Battle Over --",
		"DEFEAT!",
		"razed",
		"rebuilt from scratch",
	} {
		if !strings.Contains(fullLog, expected) {
			t.Errorf("expected message log to contain %q", expected)
		}
	}

	// Should have breach messages since there are no defenses
	if !strings.Contains(fullLog, "breaches undefended village") {
		t.Errorf("expected 'breaches undefended village' in log")
	}
}

// TestAutoTideVictoryMessages verifies victory path messages.
func TestAutoTideVictoryMessages(t *testing.T) {
	// Create a very strong village that will always win
	guards := []models.Guard{}
	for i := 0; i < 5; i++ {
		guards = append(guards, models.Guard{
			Name:               "Guard",
			Level:              50,
			HitPoints:          999,
			HitpointsNatural:   999,
			HitpointsRemaining: 999,
			AttackBonus:        100,
			DefenseBonus:       100,
			AttackRolls:        10,
			DefenseRolls:       10,
			StatsMod:           models.StatMod{},
		})
	}

	village := models.Village{
		Name:         "StrongVillage",
		Level:        1,
		DefenseLevel: 1,
		ActiveGuards: guards,
		Villagers:    []models.Villager{},
		Defenses:     []models.Defense{},
		Traps:        []models.Trap{},
		TideInterval: 3600,
	}

	player := models.Character{
		Name:               "TestPlayer",
		ResourceStorageMap: map[string]models.Resource{},
	}

	result := ProcessAutoTide(&village, &player)

	if !result.Victory {
		t.Fatal("expected victory with overpowered guards, got defeat")
	}

	fullLog := strings.Join(result.Messages, "\n")
	if !strings.Contains(fullLog, "VICTORY!") {
		t.Error("expected VICTORY in messages")
	}
	if !strings.Contains(fullLog, "held the line") {
		t.Error("expected 'held the line' in messages")
	}

	// Village should NOT be reset
	if village.Level < 1 {
		t.Errorf("village level should not decrease on victory, got %d", village.Level)
	}
}

// TestVillageManagerTickAssignsHarvesters verifies idle harvesters get assigned.
func TestVillageManagerTickAssignsHarvesters(t *testing.T) {
	village := models.Village{
		Name:  "ManagerTestVillage",
		Level: 1,
		Villagers: []models.Villager{
			{Name: "Alice", Role: "harvester", HarvestType: "", AssignedTask: ""},
			{Name: "Bob", Role: "harvester", HarvestType: "", AssignedTask: ""},
			{Name: "Carol", Role: "guard", HarvestType: "", AssignedTask: ""},
		},
	}
	player := models.Character{
		Name:               "TestPlayer",
		ResourceStorageMap: map[string]models.Resource{},
	}

	messages := ProcessVillageManagerTick(&village, &player)

	// Two idle harvesters should be assigned
	assignCount := 0
	for _, msg := range messages {
		if strings.Contains(msg, "assigned to harvest") {
			assignCount++
		}
	}
	if assignCount != 2 {
		t.Errorf("expected 2 harvester assignments, got %d", assignCount)
	}

	// Verify harvesters now have types
	validResources := map[string]bool{"Lumber": true, "Gold": true, "Iron": true, "Sand": true, "Stone": true}
	for _, v := range village.Villagers {
		if v.Role == "harvester" {
			if v.HarvestType == "" {
				t.Errorf("harvester %s still has empty HarvestType", v.Name)
			}
			if !validResources[v.HarvestType] {
				t.Errorf("harvester %s assigned invalid resource %q", v.Name, v.HarvestType)
			}
			if v.AssignedTask != "harvesting" {
				t.Errorf("harvester %s task should be 'harvesting', got %q", v.Name, v.AssignedTask)
			}
		}
	}

	// Guard should not be touched
	if village.Villagers[2].HarvestType != "" {
		t.Error("guard villager should not be assigned a harvest type")
	}

	// XP should be awarded (+10 per assignment)
	if village.Experience != 20 {
		t.Errorf("expected 20 XP from 2 assignments, got %d", village.Experience)
	}
}

// TestVillageManagerTickHiresGuard verifies guard hiring when resources allow.
func TestVillageManagerTickHiresGuard(t *testing.T) {
	village := models.Village{
		Name:         "GuardTestVillage",
		Level:        2,
		ActiveGuards: []models.Guard{},
		Villagers:    []models.Villager{},
		Defenses:     []models.Defense{},
		Traps:        []models.Trap{},
	}
	// Cost = 50 + 2*25 = 100 gold
	player := models.Character{
		Name: "TestPlayer",
		ResourceStorageMap: map[string]models.Resource{
			"Gold": {Name: "Gold", Stock: 100},
		},
	}

	messages := ProcessVillageManagerTick(&village, &player)

	hiredCount := 0
	for _, msg := range messages {
		if strings.Contains(msg, "Hired guard") {
			hiredCount++
		}
	}
	if hiredCount != 1 {
		t.Errorf("expected 1 guard hired, got %d", hiredCount)
	}
	if len(village.ActiveGuards) != 1 {
		t.Errorf("expected 1 active guard, got %d", len(village.ActiveGuards))
	}
	if player.ResourceStorageMap["Gold"].Stock != 0 {
		t.Errorf("expected 0 gold remaining, got %d", player.ResourceStorageMap["Gold"].Stock)
	}
}

// TestVillageManagerTickBuildsWall verifies wall building when resources allow.
func TestVillageManagerTickBuildsWall(t *testing.T) {
	village := models.Village{
		Name:         "WallTestVillage",
		Level:        2,
		DefenseLevel: 1,
		ActiveGuards: []models.Guard{{}, {}}, // At cap, no guard hiring
		Villagers:    []models.Villager{},
		Defenses:     []models.Defense{},
		Traps:        []models.Trap{},
	}
	player := models.Character{
		Name: "TestPlayer",
		ResourceStorageMap: map[string]models.Resource{
			"Lumber": {Name: "Lumber", Stock: 60},
			"Stone":  {Name: "Stone", Stock: 30},
		},
	}

	messages := ProcessVillageManagerTick(&village, &player)

	wallBuilt := false
	for _, msg := range messages {
		if strings.Contains(msg, "Built a Wooden Wall") {
			wallBuilt = true
		}
	}
	if !wallBuilt {
		t.Error("expected wall to be built")
	}
	if len(village.Defenses) != 1 {
		t.Errorf("expected 1 defense, got %d", len(village.Defenses))
	}
	if player.ResourceStorageMap["Lumber"].Stock != 10 {
		t.Errorf("expected 10 lumber remaining, got %d", player.ResourceStorageMap["Lumber"].Stock)
	}
	if player.ResourceStorageMap["Stone"].Stock != 10 {
		t.Errorf("expected 10 stone remaining, got %d", player.ResourceStorageMap["Stone"].Stock)
	}
}
