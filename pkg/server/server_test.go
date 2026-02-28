package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"rpg-game/pkg/auth"
	"rpg-game/pkg/db"
)

func setupTestServer(t *testing.T) (*Server, *httptest.Server) {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, err := db.NewStore(dbPath)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	t.Cleanup(func() { store.Close() })

	authSvc := auth.NewAuthService(store, "test-secret-key")
	srv := NewServer(store, authSvc, "", "test", nil, 0)
	ts := httptest.NewServer(srv)
	t.Cleanup(ts.Close)
	return srv, ts
}

func TestRegisterEndpoint(t *testing.T) {
	_, ts := setupTestServer(t)

	// Register a new account.
	body := `{"username":"hero123","password":"secret99"}`
	resp, err := http.Post(ts.URL+"/api/register", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("POST /api/register: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		var errBody map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errBody)
		t.Fatalf("expected 201, got %d: %v", resp.StatusCode, errBody)
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	if result["username"] != "hero123" {
		t.Errorf("expected username hero123, got %v", result["username"])
	}
	if result["account_id"] == nil || result["account_id"].(float64) <= 0 {
		t.Errorf("expected positive account_id, got %v", result["account_id"])
	}

	// Duplicate should return 409.
	resp2, err := http.Post(ts.URL+"/api/register", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("duplicate register: %v", err)
	}
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusConflict {
		t.Errorf("expected 409 for duplicate, got %d", resp2.StatusCode)
	}
}

func TestLoginEndpoint(t *testing.T) {
	_, ts := setupTestServer(t)

	// Register first.
	regBody := `{"username":"warrior","password":"strongpass"}`
	resp, err := http.Post(ts.URL+"/api/register", "application/json", strings.NewReader(regBody))
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("register failed: %d", resp.StatusCode)
	}

	// Login with correct credentials.
	loginBody := `{"username":"warrior","password":"strongpass"}`
	resp2, err := http.Post(ts.URL+"/api/login", "application/json", strings.NewReader(loginBody))
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusOK {
		var errBody map[string]interface{}
		json.NewDecoder(resp2.Body).Decode(&errBody)
		t.Fatalf("expected 200, got %d: %v", resp2.StatusCode, errBody)
	}

	var loginResult map[string]interface{}
	json.NewDecoder(resp2.Body).Decode(&loginResult)

	if loginResult["token"] == nil || loginResult["token"].(string) == "" {
		t.Error("expected non-empty token")
	}
	if loginResult["username"] != "warrior" {
		t.Errorf("expected username warrior, got %v", loginResult["username"])
	}

	// Login with wrong password.
	badBody := `{"username":"warrior","password":"wrong"}`
	resp3, err := http.Post(ts.URL+"/api/login", "application/json", strings.NewReader(badBody))
	if err != nil {
		t.Fatalf("bad login: %v", err)
	}
	resp3.Body.Close()
	if resp3.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401 for wrong password, got %d", resp3.StatusCode)
	}
}

func TestCreateCharacterAndList(t *testing.T) {
	_, ts := setupTestServer(t)

	// Register + login.
	token := registerAndLogin(t, ts, "mage_player", "password1")

	// Create a character.
	req, _ := http.NewRequest("POST", ts.URL+"/api/characters", strings.NewReader(`{"name":"Gandalf"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("create char: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		var errBody map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errBody)
		t.Fatalf("expected 201, got %d: %v", resp.StatusCode, errBody)
	}

	// List characters.
	req2, _ := http.NewRequest("GET", ts.URL+"/api/characters", nil)
	req2.Header.Set("Authorization", "Bearer "+token)
	resp2, err := http.DefaultClient.Do(req2)
	if err != nil {
		t.Fatalf("list chars: %v", err)
	}
	defer resp2.Body.Close()

	var listResult map[string]interface{}
	json.NewDecoder(resp2.Body).Decode(&listResult)

	chars, ok := listResult["characters"].([]interface{})
	if !ok || len(chars) == 0 {
		t.Fatalf("expected characters list, got %v", listResult)
	}
	if chars[0].(string) != "Gandalf" {
		t.Errorf("expected Gandalf, got %v", chars[0])
	}
}

func TestWebSocketAutoHunt(t *testing.T) {
	_, ts := setupTestServer(t)

	// Register, login, get token.
	token := registerAndLogin(t, ts, "hunter99", "huntpass1")

	// Connect WebSocket.
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
	defer ws.Close()

	// Read the init response (server sends it automatically on connect).
	initResp := readGameResponse(t, ws)
	if initResp.Type == "error" {
		t.Fatalf("init error: %v", initResp.Messages)
	}
	t.Logf("Init screen: %s", screenOf(initResp))

	// Should be at main menu now (engine auto-creates a Temp character).
	if screenOf(initResp) != "main_menu" {
		t.Fatalf("expected main_menu, got %s", screenOf(initResp))
	}

	// Select auto-play (option "8").
	sendCommand(t, ws, "select", "8")
	speedResp := readGameResponse(t, ws)
	t.Logf("Auto-play speed screen: %s", screenOf(speedResp))

	// Select turbo speed (option "4").
	sendCommand(t, ws, "select", "4")
	autoResp := readGameResponse(t, ws)
	t.Logf("Auto-play result screen: %s, messages: %d", screenOf(autoResp), len(autoResp.Messages))

	// Should be on autoplay_menu with fight results.
	if screenOf(autoResp) != "autoplay_menu" {
		t.Errorf("expected autoplay_menu, got %s", screenOf(autoResp))
	}
	if len(autoResp.Messages) < 3 {
		t.Errorf("expected multiple fight messages, got %d", len(autoResp.Messages))
	}

	// Check that fights actually happened by looking for combat messages.
	hasVictoryOrDefeat := false
	for _, msg := range autoResp.Messages {
		if strings.Contains(msg.Text, "VICTORY") || strings.Contains(msg.Text, "DEFEAT") {
			hasVictoryOrDefeat = true
			break
		}
	}
	if !hasVictoryOrDefeat {
		t.Error("auto-play did not produce any fight results")
		for _, msg := range autoResp.Messages {
			t.Logf("  [%s] %s", msg.Category, msg.Text)
		}
	}

	// Return to main menu.
	sendCommand(t, ws, "select", "0")
	menuResp := readGameResponse(t, ws)
	if screenOf(menuResp) != "main_menu" {
		t.Errorf("expected main_menu after return, got %s", screenOf(menuResp))
	}
}

// --- Helpers ---

func registerAndLogin(t *testing.T, ts *httptest.Server, username, password string) string {
	t.Helper()

	regBody := `{"username":"` + username + `","password":"` + password + `"}`
	resp, err := http.Post(ts.URL+"/api/register", "application/json", strings.NewReader(regBody))
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("register status: %d", resp.StatusCode)
	}

	loginBody := `{"username":"` + username + `","password":"` + password + `"}`
	resp2, err := http.Post(ts.URL+"/api/login", "application/json", strings.NewReader(loginBody))
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusOK {
		t.Fatalf("login status: %d", resp2.StatusCode)
	}

	var result map[string]interface{}
	json.NewDecoder(resp2.Body).Decode(&result)
	token, ok := result["token"].(string)
	if !ok || token == "" {
		t.Fatal("no token in login response")
	}
	return token
}

type gameResponse struct {
	Type     string        `json:"type"`
	Messages []gameMessage `json:"messages"`
	State    *stateData    `json:"state"`
	Options  []menuOption  `json:"options"`
	Prompt   string        `json:"prompt"`
}

type gameMessage struct {
	Text     string `json:"text"`
	Category string `json:"category"`
}

type stateData struct {
	Screen string `json:"screen"`
}

type menuOption struct {
	Key     string `json:"key"`
	Label   string `json:"label"`
	Enabled bool   `json:"enabled"`
}

func readGameResponse(t *testing.T, ws *websocket.Conn) gameResponse {
	t.Helper()
	ws.SetReadDeadline(time.Now().Add(10 * time.Second))
	_, msg, err := ws.ReadMessage()
	if err != nil {
		t.Fatalf("read ws: %v", err)
	}
	var resp gameResponse
	if err := json.Unmarshal(msg, &resp); err != nil {
		t.Fatalf("unmarshal response: %v\nraw: %s", err, string(msg))
	}
	return resp
}

func sendCommand(t *testing.T, ws *websocket.Conn, cmdType, value string) {
	t.Helper()
	msg := `{"type":"` + cmdType + `","value":"` + value + `"}`
	ws.SetWriteDeadline(time.Now().Add(5 * time.Second))
	if err := ws.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
		t.Fatalf("write ws: %v", err)
	}
}

func screenOf(resp gameResponse) string {
	if resp.State != nil {
		return resp.State.Screen
	}
	return ""
}
