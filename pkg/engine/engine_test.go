package engine

import (
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func createTestEngine(t *testing.T) (*Engine, string) {
	t.Helper()
	eng := NewEngine()
	// Use a unique temp file per test to avoid stale state
	tmpFile := fmt.Sprintf("/tmp/test_engine_%d.json", time.Now().UnixNano())
	t.Cleanup(func() { os.Remove(tmpFile) })
	sessionID, err := eng.CreateLocalSession(tmpFile)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	return eng, sessionID
}

func TestEngineInit(t *testing.T) {
	eng, sessionID := createTestEngine(t)

	resp := eng.ProcessCommand(sessionID, GameCommand{Type: "init"})
	if resp.Type == "error" {
		t.Fatalf("Init returned error: %v", resp.Messages)
	}

	// Should have set up a player and returned main menu
	if resp.State == nil {
		t.Fatal("Response state is nil")
	}
	if resp.State.Player == nil {
		t.Fatal("Player state is nil")
	}
	if resp.State.Player.Level != 1 {
		t.Errorf("Expected level 1, got %d", resp.State.Player.Level)
	}
	if len(resp.Options) == 0 {
		t.Error("Expected menu options")
	}
}

func TestMainMenuNavigation(t *testing.T) {
	eng, sessionID := createTestEngine(t)
	eng.ProcessCommand(sessionID, GameCommand{Type: "init"})

	// Test character stats display
	resp := eng.ProcessCommand(sessionID, GameCommand{Type: "select", Value: "5"})
	if resp.Type == "error" {
		t.Fatalf("Player stats returned error: %v", resp.Messages)
	}
	if len(resp.Messages) == 0 {
		t.Error("Expected stat messages")
	}

	// Should return to main menu on any input
	resp = eng.ProcessCommand(sessionID, GameCommand{Type: "select", Value: "0"})
	if resp.Type == "error" {
		t.Fatalf("Return to menu failed: %v", resp.Messages)
	}
}

func TestCharacterCreation(t *testing.T) {
	eng, sessionID := createTestEngine(t)
	eng.ProcessCommand(sessionID, GameCommand{Type: "init"})

	// Go to character create
	resp := eng.ProcessCommand(sessionID, GameCommand{Type: "select", Value: "0"})
	if resp.Prompt == "" {
		t.Error("Expected name prompt for character creation")
	}

	// Create a new character
	resp = eng.ProcessCommand(sessionID, GameCommand{Type: "input", Value: "TestHero"})
	if resp.Type == "error" {
		t.Fatalf("Character creation failed: %v", resp.Messages)
	}

	// Verify character was created
	eng.mu.RLock()
	session := eng.sessions[sessionID]
	eng.mu.RUnlock()

	_, exists := session.GameState.CharactersMap["TestHero"]
	if !exists {
		t.Error("TestHero not found in character map")
	}
}

func TestHarvestResource(t *testing.T) {
	eng, sessionID := createTestEngine(t)
	eng.ProcessCommand(sessionID, GameCommand{Type: "init"})

	// Go to harvest
	resp := eng.ProcessCommand(sessionID, GameCommand{Type: "select", Value: "1"})
	if len(resp.Options) == 0 {
		t.Error("Expected resource type options")
	}

	// Harvest lumber
	resp = eng.ProcessCommand(sessionID, GameCommand{Type: "select", Value: "Lumber"})
	if resp.Type == "error" {
		t.Fatalf("Harvest failed: %v", resp.Messages)
	}
}

func TestSearchLocation(t *testing.T) {
	eng, sessionID := createTestEngine(t)
	eng.ProcessCommand(sessionID, GameCommand{Type: "init"})

	// Search for location
	resp := eng.ProcessCommand(sessionID, GameCommand{Type: "select", Value: "2"})
	if resp.Type == "error" {
		t.Fatalf("Search failed: %v", resp.Messages)
	}
	// Should return to main menu
	if len(resp.Options) == 0 {
		t.Error("Expected menu options after search")
	}
}

func TestHuntFlow(t *testing.T) {
	eng, sessionID := createTestEngine(t)
	eng.ProcessCommand(sessionID, GameCommand{Type: "init"})

	// Go to hunt
	resp := eng.ProcessCommand(sessionID, GameCommand{Type: "select", Value: "3"})
	if len(resp.Options) == 0 {
		t.Fatal("Expected location options for hunt")
	}

	// Pick first available location
	locationKey := resp.Options[0].Key
	resp = eng.ProcessCommand(sessionID, GameCommand{Type: "select", Value: locationKey})
	if resp.Type == "error" {
		t.Fatalf("Location select failed: %v", resp.Messages)
	}

	// Enter hunt count
	resp = eng.ProcessCommand(sessionID, GameCommand{Type: "input", Value: "1"})
	if resp.Type == "error" {
		t.Fatalf("Hunt count failed: %v", resp.Messages)
	}

	// Should now be in combat (or tracking)
	eng.mu.RLock()
	session := eng.sessions[sessionID]
	eng.mu.RUnlock()

	if session.State != StateCombat && session.State != StateHuntTracking && session.State != StateCombatGuardPrompt {
		t.Errorf("Expected combat or tracking state, got: %s", session.State)
	}
}

func TestCombatFlow(t *testing.T) {
	eng, sessionID := createTestEngine(t)
	eng.ProcessCommand(sessionID, GameCommand{Type: "init"})

	// Navigate to combat
	resp := eng.ProcessCommand(sessionID, GameCommand{Type: "select", Value: "3"})
	if len(resp.Options) == 0 {
		t.Fatal("No hunt locations")
	}

	locationKey := resp.Options[0].Key
	eng.ProcessCommand(sessionID, GameCommand{Type: "select", Value: locationKey})
	resp = eng.ProcessCommand(sessionID, GameCommand{Type: "input", Value: "1"})

	// Handle tracking state if player has tracking
	eng.mu.RLock()
	session := eng.sessions[sessionID]
	eng.mu.RUnlock()
	if session.State == StateHuntTracking {
		resp = eng.ProcessCommand(sessionID, GameCommand{Type: "select", Value: "0"})
	}

	// If guard prompt, decline
	if session.State == StateCombatGuardPrompt {
		resp = eng.ProcessCommand(sessionID, GameCommand{Type: "select", Value: "n"})
	}

	// Should be in combat now
	if session.State != StateCombat {
		t.Skipf("Not in combat state (state=%s), skipping combat test", session.State)
	}

	// Attack until combat ends (max 100 turns to prevent infinite loop)
	for i := 0; i < 100; i++ {
		resp = eng.ProcessCommand(sessionID, GameCommand{Type: "select", Value: "1"})
		if session.State != StateCombat &&
			session.State != StateCombatItemSelect &&
			session.State != StateCombatSkillSelect {
			break
		}
	}

	// Combat should have ended
	if session.State == StateCombat {
		t.Error("Combat didn't end after 100 turns")
	}
}

func TestQuestLog(t *testing.T) {
	eng, sessionID := createTestEngine(t)
	eng.ProcessCommand(sessionID, GameCommand{Type: "init"})

	resp := eng.ProcessCommand(sessionID, GameCommand{Type: "select", Value: "9"})
	if resp.Type == "error" {
		t.Fatalf("Quest log failed: %v", resp.Messages)
	}
	if len(resp.Messages) == 0 {
		t.Error("Expected quest log messages")
	}

	// Return to main menu (send "back" which is the option key)
	resp = eng.ProcessCommand(sessionID, GameCommand{Type: "select", Value: "back"})

	eng.mu.RLock()
	session := eng.sessions[sessionID]
	eng.mu.RUnlock()

	if session.State != StateMainMenu {
		t.Errorf("Expected main menu, got: %s", session.State)
	}
}

func TestExitSaves(t *testing.T) {
	eng, sessionID := createTestEngine(t)
	eng.ProcessCommand(sessionID, GameCommand{Type: "init"})

	resp := eng.ProcessCommand(sessionID, GameCommand{Type: "select", Value: "exit"})
	if resp.Type != "exit" {
		t.Errorf("Expected exit response type, got: %s", resp.Type)
	}
}

func TestDiscoveredLocations(t *testing.T) {
	eng, sessionID := createTestEngine(t)
	eng.ProcessCommand(sessionID, GameCommand{Type: "init"})

	resp := eng.ProcessCommand(sessionID, GameCommand{Type: "select", Value: "4"})
	if resp.Type == "error" {
		t.Fatalf("Discovered locations failed: %v", resp.Messages)
	}
	if len(resp.Messages) == 0 {
		t.Error("Expected location messages")
	}
}
