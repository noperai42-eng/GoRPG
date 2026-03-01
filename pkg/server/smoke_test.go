package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
)

// --- Smoke-test helpers ---

// connectWS dials a WebSocket connection to the test server and returns it.
func connectWS(t *testing.T, ts *httptest.Server, token string) *websocket.Conn {
	t.Helper()
	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws/game?token=" + token
	ws, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		body := ""
		if resp != nil && resp.Body != nil {
			var b [512]byte
			n, _ := resp.Body.Read(b[:])
			body = string(b[:n])
		}
		t.Fatalf("WebSocket dial: %v (body: %s)", err, body)
	}
	return ws
}

// hasOption checks if a response contains an option with the given key.
func hasOption(resp gameResponse, key string) bool {
	for _, opt := range resp.Options {
		if opt.Key == key {
			return true
		}
	}
	return false
}

// optionKeys returns all option keys for debugging.
func optionKeys(resp gameResponse) []string {
	keys := make([]string, len(resp.Options))
	for i, opt := range resp.Options {
		keys[i] = opt.Key
	}
	return keys
}

// firstOptionKey returns the key of the first option, or "" if none.
func firstOptionKey(resp gameResponse) string {
	if len(resp.Options) > 0 {
		return resp.Options[0].Key
	}
	return ""
}

// firstNonLockedOption returns the first option key that doesn't start with "locked:".
func firstNonLockedOption(resp gameResponse) string {
	for _, opt := range resp.Options {
		if !strings.HasPrefix(opt.Key, "locked:") && opt.Key != "0" {
			return opt.Key
		}
	}
	return ""
}

// requireScreen asserts the response screen matches expected, fataling otherwise.
func requireScreen(t *testing.T, resp gameResponse, expected, context string) {
	t.Helper()
	got := screenOf(resp)
	if got != expected {
		t.Fatalf("%s: expected screen %q, got %q (options: %v, messages: %d)",
			context, expected, got, optionKeys(resp), len(resp.Messages))
	}
}

// readGameResponseSkipBroadcasts reads WS messages, skipping any broadcast-type pushes,
// and returns the first non-broadcast game response.
func readGameResponseSkipBroadcasts(t *testing.T, ws *websocket.Conn) gameResponse {
	t.Helper()
	for {
		resp := readGameResponse(t, ws)
		if resp.Type == "broadcast" {
			t.Logf("  (skipped broadcast: %v)", resp.Messages)
			continue
		}
		return resp
	}
}

func TestArenaFlow(t *testing.T) {
	_, ts := setupTestServer(t)

	// Register two separate accounts.
	token1 := registerAndLogin(t, ts, "arena_p1", "arenapass1")
	token2 := registerAndLogin(t, ts, "arena_p2", "arenapass2")

	// Connect player 1.
	ws1 := connectWS(t, ts, token1)
	defer ws1.Close()

	initResp1 := readGameResponse(t, ws1)
	requireScreen(t, initResp1, "main_menu", "p1 init")
	t.Logf("P1 init OK: screen=%s", screenOf(initResp1))

	// Player 1 enters arena (option "14") → auto-registers with 1000 rating.
	sendCommand(t, ws1, "select", "14")
	resp := readGameResponseSkipBroadcasts(t, ws1)
	requireScreen(t, resp, "arena_main", "p1 enter arena")
	t.Log("P1 entered arena, auto-registered")

	// Player 1 goes back to main menu.
	sendCommand(t, ws1, "select", "back")
	resp = readGameResponseSkipBroadcasts(t, ws1)
	requireScreen(t, resp, "main_menu", "p1 back from arena")

	// Connect player 2 (player 1 may get a login broadcast — that's fine, we skip it).
	ws2 := connectWS(t, ts, token2)
	defer ws2.Close()

	initResp2 := readGameResponse(t, ws2)
	requireScreen(t, initResp2, "main_menu", "p2 init")
	t.Logf("P2 init OK: screen=%s", screenOf(initResp2))

	// Player 2 enters arena → auto-registers.
	sendCommand(t, ws2, "select", "14")
	resp = readGameResponseSkipBroadcasts(t, ws2)
	requireScreen(t, resp, "arena_main", "p2 enter arena")
	t.Log("P2 entered arena, auto-registered")

	// Player 2 challenges (option "1") → should see player 1 in the list.
	sendCommand(t, ws2, "select", "1")
	resp = readGameResponseSkipBroadcasts(t, ws2)
	requireScreen(t, resp, "arena_challenge", "p2 challenge list")

	if len(resp.Options) < 2 { // At least 1 opponent + "back"
		t.Fatalf("expected at least 1 opponent + back, got options: %v", optionKeys(resp))
	}
	t.Logf("P2 sees opponents: %v", optionKeys(resp))

	// Select the first opponent (should be player 1).
	sendCommand(t, ws2, "select", "1")
	resp = readGameResponseSkipBroadcasts(t, ws2)
	requireScreen(t, resp, "arena_confirm", "p2 confirm challenge")

	// Confirm the fight.
	sendCommand(t, ws2, "select", "y")
	resp = readGameResponseSkipBroadcasts(t, ws2)
	requireScreen(t, resp, "combat", "p2 arena combat start")
	t.Log("Arena combat started")

	// Fight until combat ends (attack repeatedly with option "1").
	arenaResult := ""
	for i := 0; i < 100; i++ {
		screen := screenOf(resp)
		if screen == "arena_main" {
			// Combat ended, we're back at arena screen.
			arenaResult = "done"
			break
		}
		if screen == "combat_skill_reward" {
			// Handle skill reward.
			if len(resp.Options) > 0 {
				sendCommand(t, ws2, "select", resp.Options[0].Key)
				resp = readGameResponseSkipBroadcasts(t, ws2)
			}
			continue
		}
		if screen != "combat" && screen != "combat_guard_prompt" {
			t.Fatalf("unexpected screen during arena combat: %s (options: %v)", screen, optionKeys(resp))
		}
		// Attack.
		sendCommand(t, ws2, "select", "1")
		resp = readGameResponseSkipBroadcasts(t, ws2)
	}

	if arenaResult != "done" {
		t.Fatalf("arena combat did not end after 100 turns, last screen: %s", screenOf(resp))
	}

	// Verify we got victory or defeat messages.
	hasResult := false
	for _, msg := range resp.Messages {
		if strings.Contains(msg.Text, "ARENA VICTORY") || strings.Contains(msg.Text, "ARENA DEFEAT") {
			hasResult = true
			t.Logf("Arena result: %s", msg.Text)
			break
		}
	}
	if !hasResult {
		t.Error("arena combat ended without VICTORY/DEFEAT message")
		for _, msg := range resp.Messages {
			t.Logf("  [%s] %s", msg.Category, msg.Text)
		}
	}

	// Verify rating change message is present.
	hasRating := false
	for _, msg := range resp.Messages {
		if strings.Contains(msg.Text, "Rating:") {
			hasRating = true
			t.Logf("Rating update: %s", msg.Text)
			break
		}
	}
	if !hasRating {
		t.Error("expected rating change message after arena fight")
	}

	t.Log("ArenaFlow OK")
}

func TestSmoke(t *testing.T) {
	_, ts := setupTestServer(t)
	token := registerAndLogin(t, ts, "smokeuser", "smokepass1")
	ws := connectWS(t, ts, token)
	defer ws.Close()

	// 1. Init — read the initial response and verify main_menu.
	t.Run("Init", func(t *testing.T) {
		initResp := readGameResponse(t, ws)
		if initResp.Type == "error" {
			t.Fatalf("init error: %v", initResp.Messages)
		}
		requireScreen(t, initResp, "main_menu", "init")
		t.Logf("Init OK: screen=%s, options=%v", screenOf(initResp), optionKeys(initResp))
	})

	// 2. HuntFlow — use "hunt" intercept, pick location, pick count, fight to completion.
	t.Run("HuntFlow", func(t *testing.T) {
		sendCommand(t, ws, "select", "hunt")
		resp := readGameResponse(t, ws)
		requireScreen(t, resp, "hunt_location_select", "hunt intercept")

		// Select first unlocked location (keys are location names, not numbers).
		locKey := firstNonLockedOption(resp)
		if locKey == "" {
			t.Fatal("no unlocked location available")
		}
		t.Logf("Selecting location: %s", locKey)

		sendCommand(t, ws, "select", locKey)
		resp = readGameResponse(t, ws)

		// Should go directly to combat or hunt_tracking (no hunt count prompt).
		screen := screenOf(resp)
		if screen != "combat" && screen != "hunt_tracking" {
			t.Fatalf("expected combat or hunt_tracking, got %s", screen)
		}

		// If hunt_tracking, select first target.
		if screen == "hunt_tracking" {
			if len(resp.Options) > 0 {
				sendCommand(t, ws, "select", resp.Options[0].Key)
				resp = readGameResponse(t, ws)
			}
		}

		// Fight until first combat ends (attack repeatedly).
		// Hunts now chain indefinitely, so after combat we use "Stop Hunting" to exit.
		fightDone := false
		for i := 0; i < 100; i++ {
			screen = screenOf(resp)
			if screen == "main_menu" {
				fightDone = true
				break
			}
			if screen == "combat_skill_reward" {
				// Handle skill reward then continue.
				if len(resp.Options) > 0 {
					sendCommand(t, ws, "select", resp.Options[0].Key)
					resp = readGameResponse(t, ws)
				}
				continue
			}
			if screen != "combat" && screen != "combat_guard_prompt" {
				break
			}
			// Send attack (option "1").
			sendCommand(t, ws, "select", "1")
			resp = readGameResponse(t, ws)
		}

		// If still in combat (next hunt started), stop hunting.
		if !fightDone && screenOf(resp) == "combat" {
			sendCommand(t, ws, "select", "7") // Stop Hunting
			resp = readGameResponse(t, ws)
		}

		requireScreen(t, resp, "main_menu", "after hunt")
		t.Log("HuntFlow OK")
	})

	// 3. VillageFlow — send "10" → verify village_main → send "0" (back) → main_menu.
	t.Run("VillageFlow", func(t *testing.T) {
		sendCommand(t, ws, "select", "10")
		resp := readGameResponse(t, ws)
		requireScreen(t, resp, "village_main", "enter village")

		sendCommand(t, ws, "select", "0")
		resp = readGameResponse(t, ws)
		requireScreen(t, resp, "main_menu", "village back")
		t.Log("VillageFlow OK")
	})

	// 4. VillageDeepEscape — enter village → hire guards (option "3") → use "hunt" intercept → back to main.
	t.Run("VillageDeepEscape", func(t *testing.T) {
		sendCommand(t, ws, "select", "10")
		resp := readGameResponse(t, ws)
		requireScreen(t, resp, "village_main", "enter village for deep escape")

		// Navigate to hire guard (option "3").
		sendCommand(t, ws, "select", "3")
		resp = readGameResponse(t, ws)
		requireScreen(t, resp, "village_hire_guard", "hire guard screen")

		// Use "hunt" intercept to escape.
		sendCommand(t, ws, "select", "hunt")
		resp = readGameResponse(t, ws)
		requireScreen(t, resp, "hunt_location_select", "hunt intercept from village")

		// Back out to main_menu via "home" intercept.
		sendCommand(t, ws, "select", "home")
		resp = readGameResponse(t, ws)
		requireScreen(t, resp, "main_menu", "home from hunt")
		t.Log("VillageDeepEscape OK")
	})

	// 5. TownFlow — send "11" → verify town_main → send "home" → main_menu.
	t.Run("TownFlow", func(t *testing.T) {
		sendCommand(t, ws, "select", "11")
		resp := readGameResponse(t, ws)
		requireScreen(t, resp, "town_main", "enter town")

		sendCommand(t, ws, "select", "home")
		resp = readGameResponse(t, ws)
		requireScreen(t, resp, "main_menu", "home from town")
		t.Log("TownFlow OK")
	})

	// 6. HarvestIntercept — send "harvest" → verify harvest_select → send "0" → main_menu.
	t.Run("HarvestIntercept", func(t *testing.T) {
		sendCommand(t, ws, "select", "harvest")
		resp := readGameResponse(t, ws)
		requireScreen(t, resp, "harvest_select", "harvest intercept")

		sendCommand(t, ws, "select", "0")
		resp = readGameResponse(t, ws)
		requireScreen(t, resp, "main_menu", "harvest back")
		t.Log("HarvestIntercept OK")
	})

	// 7. TownDeepEscape — enter town → navigate deeper → use "harvest" intercept → back.
	t.Run("TownDeepEscape", func(t *testing.T) {
		sendCommand(t, ws, "select", "11")
		resp := readGameResponse(t, ws)
		requireScreen(t, resp, "town_main", "enter town for deep escape")

		// Navigate to inn (option "1").
		sendCommand(t, ws, "select", "1")
		resp = readGameResponse(t, ws)
		screen := screenOf(resp)
		if screen != "town_inn" && screen != "town_inn_view_guests" {
			t.Fatalf("expected town_inn or town_inn_view_guests, got %s", screen)
		}

		// Use "harvest" intercept to escape.
		sendCommand(t, ws, "select", "harvest")
		resp = readGameResponse(t, ws)
		requireScreen(t, resp, "harvest_select", "harvest intercept from town")

		// Back to main_menu.
		sendCommand(t, ws, "select", "0")
		resp = readGameResponse(t, ws)
		requireScreen(t, resp, "main_menu", "main menu after harvest escape")
		t.Log("TownDeepEscape OK")
	})
}

func TestSmokeVersion(t *testing.T) {
	_, ts := setupTestServer(t)

	resp, err := http.Get(ts.URL + "/api/version")
	if err != nil {
		t.Fatalf("GET /api/version: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if result["version"] != "test" {
		t.Errorf("expected version 'test', got %q", result["version"])
	}
	if _, ok := result["game_time"]; !ok {
		t.Error("expected game_time field in version response")
	}
	if _, ok := result["calendar"]; !ok {
		t.Error("expected calendar field in version response")
	}
}
