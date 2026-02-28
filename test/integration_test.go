package test

import (
	"errors"
	"fmt"
	"math/rand"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"rpg-game/pkg/auth"
	"rpg-game/pkg/db"
	"rpg-game/pkg/engine"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

const jwtSecret = "integration-test-jwt-secret"

// -----------------------------------------------------------------------------
// Helper: create a store backed by a temp SQLite database
// -----------------------------------------------------------------------------

func newTestStore(t *testing.T) *db.Store {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, err := db.NewStore(dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	t.Cleanup(func() { store.Close() })
	return store
}

// -----------------------------------------------------------------------------
// Helper: navigate from main menu through hunt into combat and return the
// last response.  Handles tracking and guard-prompt intermediate states.
// -----------------------------------------------------------------------------

func navigateToCombat(t *testing.T, eng *engine.Engine, sessionID string) engine.GameResponse {
	t.Helper()

	// Select "3" (Hunt) from main menu.
	resp := eng.ProcessCommand(sessionID, engine.GameCommand{Type: "select", Value: "3"})
	if resp.Type == "error" {
		t.Fatalf("navigateToCombat: hunt select error: %v", messagesText(resp.Messages))
	}
	if len(resp.Options) == 0 {
		t.Fatal("navigateToCombat: no hunt locations available")
	}

	// Pick the first location.
	locKey := resp.Options[0].Key
	resp = eng.ProcessCommand(sessionID, engine.GameCommand{Type: "select", Value: locKey})
	if resp.Type == "error" {
		t.Fatalf("navigateToCombat: location select error: %v", messagesText(resp.Messages))
	}

	// Enter hunt count of 1.
	resp = eng.ProcessCommand(sessionID, engine.GameCommand{Type: "input", Value: "1"})
	if resp.Type == "error" {
		t.Fatalf("navigateToCombat: hunt count error: %v", messagesText(resp.Messages))
	}

	// Handle tracking state (if the player has the Tracking skill).
	if resp.State != nil && resp.State.Screen == "hunt_tracking" {
		resp = eng.ProcessCommand(sessionID, engine.GameCommand{Type: "select", Value: "0"})
	}

	// Handle guard prompt (if fighting a skill guardian or boss).
	if resp.State != nil && resp.State.Screen == "combat_guard_prompt" {
		resp = eng.ProcessCommand(sessionID, engine.GameCommand{Type: "select", Value: "n"})
	}

	return resp
}

// messagesText concatenates all message texts for diagnostic output.
func messagesText(msgs []engine.GameMessage) string {
	out := ""
	for _, m := range msgs {
		out += m.Text + " | "
	}
	return out
}

// currentScreen safely returns the screen name from a response, or "".
func currentScreen(resp engine.GameResponse) string {
	if resp.State != nil {
		return resp.State.Screen
	}
	return ""
}

// isCombatScreen returns true if the screen string indicates an active combat state.
func isCombatScreen(screen string) bool {
	switch screen {
	case "combat", "combat_item_select", "combat_skill_select", "combat_guard_prompt", "combat_skill_reward":
		return true
	}
	return false
}

// attackUntilCombatEnds sends attack commands until combat resolves.
// Returns the final response and whether combat actually ended.
func attackUntilCombatEnds(t *testing.T, eng *engine.Engine, sessionID string, maxTurns int) (engine.GameResponse, bool) {
	t.Helper()
	var resp engine.GameResponse
	for i := 0; i < maxTurns; i++ {
		resp = eng.ProcessCommand(sessionID, engine.GameCommand{Type: "select", Value: "1"})
		screen := currentScreen(resp)
		if !isCombatScreen(screen) {
			return resp, true
		}
		// Handle skill reward prompt (choose scroll).
		if screen == "combat_skill_reward" {
			resp = eng.ProcessCommand(sessionID, engine.GameCommand{Type: "select", Value: "2"})
			screen = currentScreen(resp)
			if !isCombatScreen(screen) {
				return resp, true
			}
		}
	}
	return resp, false
}

// =============================================================================
// Test 1: TestFullGameFlow
// =============================================================================

func TestFullGameFlow(t *testing.T) {
	store := newTestStore(t)
	authSvc := auth.NewAuthService(store, jwtSecret)

	// 1. Register an account.
	accountID, err := authSvc.Register("testplayer", "secret123")
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}
	if accountID <= 0 {
		t.Fatalf("expected positive account ID, got %d", accountID)
	}

	// 2. Create engine backed by SQLite.
	eng := engine.NewEngineWithStore(store, nil)

	// 3. Create DB session.
	sessionID, err := eng.CreateDBSession(accountID)
	if err != nil {
		t.Fatalf("CreateDBSession failed: %v", err)
	}
	t.Cleanup(func() { eng.RemoveSession(sessionID) })

	// 4. Init -- verify player state comes back.
	resp := eng.ProcessCommand(sessionID, engine.GameCommand{Type: "init"})
	if resp.Type == "error" {
		t.Fatalf("init error: %v", messagesText(resp.Messages))
	}
	if resp.State == nil || resp.State.Player == nil {
		t.Fatal("init did not return player state")
	}
	initialLevel := resp.State.Player.Level
	if initialLevel < 1 {
		t.Fatalf("expected level >= 1, got %d", initialLevel)
	}
	playerName := resp.State.Player.Name

	// 5. Navigate to hunt and enter combat.
	resp = navigateToCombat(t, eng, sessionID)

	// 6. If we are in combat, attack until it ends.
	screen := currentScreen(resp)
	if isCombatScreen(screen) {
		var ended bool
		resp, ended = attackUntilCombatEnds(t, eng, sessionID, 100)
		if !ended {
			t.Fatal("combat did not end after 100 turns")
		}
	}

	// 7. Verify we reached a valid post-combat state.
	screen = currentScreen(resp)
	validPostCombat := map[string]bool{
		"main_menu":            true,
		"combat":               true, // next hunt if multi-hunt
		"combat_skill_reward":  true,
		"hunt_tracking":        true,
		"combat_guard_prompt":  true,
	}
	if !validPostCombat[screen] {
		t.Errorf("unexpected screen after combat: %s", screen)
	}

	// Record experience after fight.
	var postFightXP int
	if resp.State != nil && resp.State.Player != nil {
		postFightXP = resp.State.Player.Experience
	}

	// 8. Save session.
	if err := eng.SaveSession(sessionID); err != nil {
		t.Fatalf("SaveSession failed: %v", err)
	}

	// 9. Create a brand-new session for the same account and verify persistence.
	eng2 := engine.NewEngineWithStore(store, nil)
	sessionID2, err := eng2.CreateDBSession(accountID)
	if err != nil {
		t.Fatalf("second CreateDBSession failed: %v", err)
	}
	t.Cleanup(func() { eng2.RemoveSession(sessionID2) })

	resp2 := eng2.ProcessCommand(sessionID2, engine.GameCommand{Type: "init"})
	if resp2.Type == "error" {
		t.Fatalf("second init error: %v", messagesText(resp2.Messages))
	}
	if resp2.State == nil || resp2.State.Player == nil {
		t.Fatal("second session did not return player state")
	}
	if resp2.State.Player.Name != playerName {
		t.Errorf("player name mismatch: got %q, want %q", resp2.State.Player.Name, playerName)
	}
	// After at least one fight the player should have gained some XP (or
	// stayed the same if they died and the XP was computed accordingly).
	if postFightXP > 0 && resp2.State.Player.Experience < postFightXP {
		t.Errorf("persisted XP (%d) is less than post-fight XP (%d)",
			resp2.State.Player.Experience, postFightXP)
	}
}

// =============================================================================
// Test 2: TestConcurrentPlayers
// =============================================================================

func TestConcurrentPlayers(t *testing.T) {
	store := newTestStore(t)

	const numPlayers = 10
	var wg sync.WaitGroup
	errs := make(chan error, numPlayers)

	for i := 0; i < numPlayers; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			username := fmt.Sprintf("player_%d", idx)
			hash, err := auth.HashPassword("password123")
			if err != nil {
				errs <- fmt.Errorf("[%d] hash error: %w", idx, err)
				return
			}
			accountID, err := store.CreateAccount(username, hash)
			if err != nil {
				errs <- fmt.Errorf("[%d] CreateAccount error: %w", idx, err)
				return
			}

			// Each goroutine gets its own engine to avoid any shared
			// session-map contention (the store is the shared resource).
			eng := engine.NewEngineWithStore(store, nil)

			sessionID, err := eng.CreateDBSession(accountID)
			if err != nil {
				errs <- fmt.Errorf("[%d] CreateDBSession error: %w", idx, err)
				return
			}
			defer eng.RemoveSession(sessionID)

			// Init
			resp := eng.ProcessCommand(sessionID, engine.GameCommand{Type: "init"})
			if resp.Type == "error" {
				errs <- fmt.Errorf("[%d] init error: %s", idx, messagesText(resp.Messages))
				return
			}

			// Run 3 hunts
			for hunt := 0; hunt < 3; hunt++ {
				// Navigate to hunt
				resp = eng.ProcessCommand(sessionID, engine.GameCommand{Type: "select", Value: "3"})
				if resp.Type == "error" || len(resp.Options) == 0 {
					// No huntable locations -- skip gracefully.
					continue
				}
				locKey := resp.Options[0].Key
				resp = eng.ProcessCommand(sessionID, engine.GameCommand{Type: "select", Value: locKey})
				resp = eng.ProcessCommand(sessionID, engine.GameCommand{Type: "input", Value: "1"})

				// Handle tracking
				if currentScreen(resp) == "hunt_tracking" {
					resp = eng.ProcessCommand(sessionID, engine.GameCommand{Type: "select", Value: "0"})
				}
				// Handle guard prompt
				if currentScreen(resp) == "combat_guard_prompt" {
					resp = eng.ProcessCommand(sessionID, engine.GameCommand{Type: "select", Value: "n"})
				}

				// Fight until combat ends
				for turn := 0; turn < 100; turn++ {
					screen := currentScreen(resp)
					if !isCombatScreen(screen) {
						break
					}
					if screen == "combat_skill_reward" {
						resp = eng.ProcessCommand(sessionID, engine.GameCommand{Type: "select", Value: "2"})
						continue
					}
					resp = eng.ProcessCommand(sessionID, engine.GameCommand{Type: "select", Value: "1"})
				}
			}

			// Save
			if err := eng.SaveSession(sessionID); err != nil {
				errs <- fmt.Errorf("[%d] SaveSession error: %w", idx, err)
				return
			}
		}(i)
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		t.Error(err)
	}
}

// =============================================================================
// Test 3: TestAutoPlayViaEngine
// =============================================================================

func TestAutoPlayViaEngine(t *testing.T) {
	eng := engine.NewEngine()
	tmpFile := filepath.Join(t.TempDir(), "autoplay.json")
	sessionID, err := eng.CreateLocalSession(tmpFile)
	if err != nil {
		t.Fatalf("CreateLocalSession failed: %v", err)
	}
	t.Cleanup(func() { eng.RemoveSession(sessionID) })

	// Init
	resp := eng.ProcessCommand(sessionID, engine.GameCommand{Type: "init"})
	if resp.Type == "error" {
		t.Fatalf("init error: %v", messagesText(resp.Messages))
	}

	// Select "8" (AUTO-PLAY MODE)
	resp = eng.ProcessCommand(sessionID, engine.GameCommand{Type: "select", Value: "8"})
	if resp.Type == "error" {
		// If no huntable locations exist, auto-play may fail gracefully.
		// That is acceptable -- just verify we are back at main menu.
		screen := currentScreen(resp)
		if screen != "main_menu" {
			t.Fatalf("auto-play error and not back at main_menu: screen=%s msgs=%v",
				screen, messagesText(resp.Messages))
		}
		t.Skip("no huntable locations; skipping auto-play test")
	}

	// Select speed "4" (Turbo)
	resp = eng.ProcessCommand(sessionID, engine.GameCommand{Type: "select", Value: "4"})
	screen := currentScreen(resp)
	if screen == "main_menu" {
		// Engine returned to main menu because no locations are huntable.
		t.Skip("no huntable locations for auto-play")
	}
	if screen != "autoplay_menu" {
		t.Fatalf("expected autoplay_menu screen, got %s", screen)
	}

	// Verify auto-play produced results.
	if len(resp.Messages) == 0 {
		t.Error("expected auto-play result messages")
	}

	// Select "0" to return to main menu.
	resp = eng.ProcessCommand(sessionID, engine.GameCommand{Type: "select", Value: "0"})
	screen = currentScreen(resp)
	if screen != "main_menu" {
		t.Errorf("expected main_menu after exiting auto-play, got %s", screen)
	}
}

// =============================================================================
// Test 4: TestAccountAuthFlow
// =============================================================================

func TestAccountAuthFlow(t *testing.T) {
	store := newTestStore(t)
	authSvc := auth.NewAuthService(store, jwtSecret)

	// 1. Register.
	accountID, err := authSvc.Register("testplayer", "secret123")
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}
	if accountID <= 0 {
		t.Fatalf("expected positive account ID, got %d", accountID)
	}

	// 2. Login and get JWT token.
	token, err := authSvc.Login("testplayer", "secret123")
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}

	// 3. Validate token.
	gotID, gotUsername, err := authSvc.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}
	if gotID != accountID {
		t.Errorf("account ID mismatch: got %d, want %d", gotID, accountID)
	}
	if gotUsername != "testplayer" {
		t.Errorf("username mismatch: got %q, want %q", gotUsername, "testplayer")
	}

	// 4. Duplicate registration.
	_, err = authSvc.Register("testplayer", "otherpass123")
	if err == nil {
		t.Fatal("expected error for duplicate username, got nil")
	}
	if !errors.Is(err, auth.ErrUsernameExists) {
		t.Errorf("expected ErrUsernameExists, got: %v", err)
	}

	// 5. Login with wrong password.
	_, err = authSvc.Login("testplayer", "wrongpassword")
	if err == nil {
		t.Fatal("expected error for wrong password, got nil")
	}
	if !errors.Is(err, auth.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got: %v", err)
	}

	// 6. Login with non-existent user.
	_, err = authSvc.Login("nobody", "anything1")
	if err == nil {
		t.Fatal("expected error for unknown user, got nil")
	}
	if !errors.Is(err, auth.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got: %v", err)
	}

	// 7. Invalid token string.
	_, _, err = authSvc.ValidateToken("garbage.token.value")
	if err == nil {
		t.Fatal("expected error for garbage token, got nil")
	}
	if !errors.Is(err, auth.ErrInvalidToken) {
		t.Errorf("expected ErrInvalidToken, got: %v", err)
	}
}

// =============================================================================
// Test 5: TestDBPersistence
// =============================================================================

func TestDBPersistence(t *testing.T) {
	store := newTestStore(t)

	// 1. Create account.
	hash, err := auth.HashPassword("password123")
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}
	accountID, err := store.CreateAccount("persist_user", hash)
	if err != nil {
		t.Fatalf("CreateAccount failed: %v", err)
	}

	// 2. First session: init, harvest lumber, save.
	eng1 := engine.NewEngineWithStore(store, nil)
	sid1, err := eng1.CreateDBSession(accountID)
	if err != nil {
		t.Fatalf("CreateDBSession failed: %v", err)
	}

	resp := eng1.ProcessCommand(sid1, engine.GameCommand{Type: "init"})
	if resp.Type == "error" {
		t.Fatalf("init error: %v", messagesText(resp.Messages))
	}
	if resp.State == nil || resp.State.Player == nil {
		t.Fatal("init did not return player state")
	}
	playerName := resp.State.Player.Name

	// Navigate to harvest (select "1"), then harvest Lumber.
	resp = eng1.ProcessCommand(sid1, engine.GameCommand{Type: "select", Value: "1"})
	if resp.Type == "error" {
		t.Fatalf("harvest menu error: %v", messagesText(resp.Messages))
	}
	resp = eng1.ProcessCommand(sid1, engine.GameCommand{Type: "select", Value: "Lumber"})
	if resp.Type == "error" {
		t.Fatalf("harvest lumber error: %v", messagesText(resp.Messages))
	}

	// Harvest a few more times to ensure a non-zero amount.
	for i := 0; i < 5; i++ {
		eng1.ProcessCommand(sid1, engine.GameCommand{Type: "select", Value: "1"})
		eng1.ProcessCommand(sid1, engine.GameCommand{Type: "select", Value: "Lumber"})
	}

	// Save.
	if err := eng1.SaveSession(sid1); err != nil {
		t.Fatalf("SaveSession failed: %v", err)
	}
	eng1.RemoveSession(sid1)

	// 3. Second session: verify the character still has resources.
	eng2 := engine.NewEngineWithStore(store, nil)
	sid2, err := eng2.CreateDBSession(accountID)
	if err != nil {
		t.Fatalf("second CreateDBSession failed: %v", err)
	}
	t.Cleanup(func() { eng2.RemoveSession(sid2) })

	resp2 := eng2.ProcessCommand(sid2, engine.GameCommand{Type: "init"})
	if resp2.Type == "error" {
		t.Fatalf("second init error: %v", messagesText(resp2.Messages))
	}
	if resp2.State == nil || resp2.State.Player == nil {
		t.Fatal("second session did not return player state")
	}
	if resp2.State.Player.Name != playerName {
		t.Errorf("player name mismatch after reload: got %q, want %q",
			resp2.State.Player.Name, playerName)
	}

	// To verify resources, load the character directly from the store.
	char, err := store.LoadCharacter(accountID, playerName)
	if err != nil {
		t.Fatalf("LoadCharacter failed: %v", err)
	}
	lumber, exists := char.ResourceStorageMap["Lumber"]
	if !exists {
		t.Fatal("Lumber resource not found after reload")
	}
	if lumber.Stock <= 0 {
		t.Errorf("expected positive Lumber stock, got %d", lumber.Stock)
	}
	t.Logf("Persisted Lumber stock: %d", lumber.Stock)

	// Verify the character level matches what we started with.
	if char.Level < 1 {
		t.Errorf("expected level >= 1, got %d", char.Level)
	}
}
