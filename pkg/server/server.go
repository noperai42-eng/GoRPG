package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"strconv"

	"github.com/gorilla/websocket"
	"rpg-game/pkg/agent"
	"rpg-game/pkg/auth"
	"rpg-game/pkg/db"
	"rpg-game/pkg/engine"
	"rpg-game/pkg/game"
	"rpg-game/pkg/metrics"
	"rpg-game/pkg/models"
)

// contextKey is a private type for context keys to avoid collisions.
type contextKey string

const ctxAccountID contextKey = "accountID"
const ctxUsername contextKey = "username"

// Server is the HTTP/WebSocket server for the RPG game.
type Server struct {
	engine   *engine.Engine
	store    *db.Store
	auth     *auth.AuthService
	metrics  *metrics.MetricsCollector
	agentMgr *agent.Manager
	version  string
	upgrader websocket.Upgrader
	mux      *http.ServeMux
}

// NewServer creates a new Server wired to the given store and auth service.
// staticDir is the path to the directory containing static web assets; if empty,
// it defaults to ../../web/static relative to this source file.
// maxAgents controls the maximum number of AI agents (0 disables agents).
func NewServer(store *db.Store, authService *auth.AuthService, staticDir string, version string, mc *metrics.MetricsCollector, maxAgents int) *Server {
	eng := engine.NewEngineWithStore(store, mc)
	s := &Server{
		engine:  eng,
		store:   store,
		auth:    authService,
		metrics: mc,
		agentMgr: agent.NewManager(eng, store, maxAgents),
		version: version,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			// Permissive origin check for development.
			CheckOrigin: func(r *http.Request) bool { return true },
		},
		mux: http.NewServeMux(),
	}

	// REST endpoints
	s.mux.HandleFunc("/api/register", s.corsWrapper(s.handleRegister))
	s.mux.HandleFunc("/api/login", s.corsWrapper(s.handleLogin))
	s.mux.HandleFunc("/api/characters", s.corsWrapper(s.authMiddleware(s.handleCharacters)))
	s.mux.HandleFunc("/api/version", s.corsWrapper(s.handleVersion))
	s.mux.HandleFunc("/api/leaderboard", s.corsWrapper(s.handleLeaderboard))
	s.mux.HandleFunc("/api/mostwanted", s.corsWrapper(s.handleMostWanted))
	s.mux.HandleFunc("/api/arena", s.corsWrapper(s.handleArena))
	s.mux.HandleFunc("/api/metrics", s.corsWrapper(s.authMiddleware(s.handleMetrics)))

	// Agent API endpoints
	s.mux.HandleFunc("/api/agents", s.corsWrapper(s.authMiddleware(s.handleAgents)))
	s.mux.HandleFunc("/api/agents/", s.corsWrapper(s.authMiddleware(s.handleAgentByID)))

	// WebSocket endpoint
	s.mux.HandleFunc("/ws/game", s.handleWebSocket)

	// Static file serving
	if staticDir == "" {
		_, thisFile, _, _ := runtime.Caller(0)
		staticDir = filepath.Join(filepath.Dir(thisFile), "..", "..", "web", "static")
	}
	fs := http.FileServer(http.Dir(staticDir))
	s.mux.Handle("/", noCacheStaticHandler(fs))

	// Evolution ticker — monsters fight each other every 60 seconds.
	// Also resets arena daily battles when the date changes.
	// Flushes metrics snapshot to DB every 60 ticks (hourly).
	go func() {
		ticker := time.NewTicker(60 * time.Second)
		defer ticker.Stop()
		lastArenaReset := game.GetArenaResetDate()
		snapshotCounter := 0
		for range ticker.C {
			result := s.engine.ProcessEvolutionTick()
			if result != nil {
				for _, evt := range result.Events {
					log.Printf("[Evolution] %s at %s: %s", evt.EventType, evt.LocationName, evt.Details)
				}
			}
			// Arena daily reset check
			today := game.GetArenaResetDate()
			if today != lastArenaReset {
				if err := store.ResetArenaBattles(today); err != nil {
					log.Printf("[Arena] Failed to reset daily battles: %v", err)
				} else {
					log.Printf("[Arena] Daily battles reset for %s", today)
				}
				lastArenaReset = today
			}
			// Hourly metrics snapshot
			snapshotCounter++
			if snapshotCounter >= 60 {
				snapshotCounter = 0
				if s.metrics != nil {
					jsonData, err := s.metrics.SnapshotJSON()
					if err != nil {
						log.Printf("[Metrics] Failed to create snapshot: %v", err)
					} else if err := store.SaveMetricsSnapshot(time.Now(), jsonData); err != nil {
						log.Printf("[Metrics] Failed to save snapshot: %v", err)
					} else {
						log.Printf("[Metrics] Hourly snapshot saved")
					}
				}
			}
		}
	}()

	// Deduplicate villages table (cleans up rows from the old INSERT bug).
	if removed, err := store.DeduplicateVillages(); err != nil {
		log.Printf("[Startup] WARNING: failed to deduplicate villages: %v", err)
	} else if removed > 0 {
		log.Printf("[Startup] Deduplicated villages table: removed %d duplicate rows", removed)
	}

	// Auto-tide ticker — process automatic monster tides every 60 seconds.
	go func() {
		ticker := time.NewTicker(60 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			result := s.engine.ProcessAutoTideTick()
			if result != nil {
				log.Printf("[AutoTide] Processed %d tides", result.TidesProcessed)
			}
		}
	}()

	// Village manager ticker — automated village upkeep every 60 seconds.
	go func() {
		ticker := time.NewTicker(60 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			result := s.engine.ProcessVillageManagerTicks()
			if result != nil {
				log.Printf("[VillageManager] Managed %d villages", result.VillagesManaged)
			}
		}
	}()

	// Auto-spawn default AI agents after a brief startup delay.
	go s.agentMgr.SpawnDefaultAgents()

	return s
}

// noCacheStaticHandler wraps a file server to set no-cache headers for JS and CSS files.
func noCacheStaticHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, ".js") || strings.HasSuffix(r.URL.Path, ".css") {
			// Cache for 1 hour; ?v= query params handle cache busting on deploy
			w.Header().Set("Cache-Control", "public, max-age=3600")
		} else if strings.HasSuffix(r.URL.Path, ".html") || r.URL.Path == "/" {
			// Never cache HTML so fixes deploy immediately
			w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		}
		next.ServeHTTP(w, r)
	})
}

// ServeHTTP delegates to the internal mux so Server satisfies http.Handler.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

// ---------------------------------------------------------------------------
// JSON helpers
// ---------------------------------------------------------------------------

func jsonResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("jsonResponse encode error: %v", err)
	}
}

func jsonError(w http.ResponseWriter, status int, message string) {
	jsonResponse(w, status, map[string]string{"error": message})
}

// ---------------------------------------------------------------------------
// CORS wrapper
// ---------------------------------------------------------------------------

func (s *Server) corsWrapper(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

// ---------------------------------------------------------------------------
// Auth middleware
// ---------------------------------------------------------------------------

func (s *Server) authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if header == "" {
			jsonError(w, http.StatusUnauthorized, "missing authorization header")
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			jsonError(w, http.StatusUnauthorized, "invalid authorization header format")
			return
		}

		tokenStr := parts[1]
		accountID, username, err := s.auth.ValidateToken(tokenStr)
		if err != nil {
			jsonError(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}

		ctx := context.WithValue(r.Context(), ctxAccountID, accountID)
		ctx = context.WithValue(ctx, ctxUsername, username)
		next(w, r.WithContext(ctx))
	}
}

// ---------------------------------------------------------------------------
// REST handlers
// ---------------------------------------------------------------------------

// handleRegister handles POST /api/register.
func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	accountID, err := s.auth.Register(body.Username, body.Password)
	if err != nil {
		// Map known auth errors to appropriate status codes.
		switch err {
		case auth.ErrInvalidUsername, auth.ErrInvalidPassword:
			jsonError(w, http.StatusBadRequest, err.Error())
		case auth.ErrUsernameExists:
			jsonError(w, http.StatusConflict, err.Error())
		default:
			log.Printf("register error: %v", err)
			jsonError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	jsonResponse(w, http.StatusCreated, map[string]interface{}{
		"account_id": accountID,
		"username":   body.Username,
	})
}

// handleLogin handles POST /api/login.
func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	token, err := s.auth.Login(body.Username, body.Password)
	if err != nil {
		if err == auth.ErrInvalidCredentials {
			jsonError(w, http.StatusUnauthorized, err.Error())
		} else {
			log.Printf("login error: %v", err)
			jsonError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"token":    token,
		"username": body.Username,
	})
}

// handleVersion handles GET /api/version.
func (s *Server) handleVersion(w http.ResponseWriter, r *http.Request) {
	cal := engine.CurrentGameCalendar()
	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"version":   s.version,
		"game_time": cal.FormatGameTime(),
		"calendar":  cal,
	})
}

// handleLeaderboard handles GET /api/leaderboard?category=kills&limit=20.
func (s *Server) handleLeaderboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	category := r.URL.Query().Get("category")
	if category == "" {
		category = "kills"
	}
	limitStr := r.URL.Query().Get("limit")
	limit := 20
	if limitStr != "" {
		if n, err := fmt.Sscanf(limitStr, "%d", &limit); n != 1 || err != nil || limit < 1 {
			limit = 20
		}
		if limit > 100 {
			limit = 100
		}
	}

	entries, err := s.store.GetLeaderboard(category, limit)
	if err != nil {
		log.Printf("leaderboard error: %v", err)
		jsonError(w, http.StatusInternalServerError, "failed to get leaderboard")
		return
	}
	if entries == nil {
		entries = []db.LeaderboardEntry{}
	}

	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"category": category,
		"entries":  entries,
	})
}

// handleMostWanted handles GET /api/mostwanted?limit=10.
func (s *Server) handleMostWanted(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 10
	if limitStr != "" {
		if n, err := fmt.Sscanf(limitStr, "%d", &limit); n != 1 || err != nil || limit < 1 {
			limit = 10
		}
		if limit > 50 {
			limit = 50
		}
	}

	entries := s.engine.GetMostWanted(limit)
	if entries == nil {
		entries = []models.MostWantedEntry{}
	}

	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"entries": entries,
	})
}

// handleArena handles GET /api/arena?limit=20.
func (s *Server) handleArena(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 20
	if limitStr != "" {
		if n, err := fmt.Sscanf(limitStr, "%d", &limit); n != 1 || err != nil || limit < 1 {
			limit = 20
		}
		if limit > 100 {
			limit = 100
		}
	}

	entries, err := s.store.GetArenaLeaderboard(limit)
	if err != nil {
		log.Printf("arena leaderboard error: %v", err)
		jsonError(w, http.StatusInternalServerError, "failed to get arena leaderboard")
		return
	}
	if entries == nil {
		entries = []db.ArenaEntry{}
	}

	champion, _ := s.store.GetArenaChampion()

	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"entries":  entries,
		"champion": champion,
	})
}

// handleMetrics handles GET /api/metrics with optional ?history=N for historical snapshots.
func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	if s.metrics == nil {
		jsonError(w, http.StatusServiceUnavailable, "metrics not available")
		return
	}

	snap := s.metrics.Snapshot()
	result := map[string]interface{}{
		"uptime_seconds": snap.UptimeSeconds,
		"online_players": snap.OnlinePlayers,
		"combat":         snap.Combat,
		"progression":    snap.Progression,
		"economy":        snap.Economy,
		"arena":          snap.Arena,
		"dungeons":       snap.Dungeons,
		"engagement":     snap.Engagement,
		"village":        snap.Village,
	}

	// Optional historical data
	historyParam := r.URL.Query().Get("history")
	if historyParam != "" {
		hours, err := strconv.Atoi(historyParam)
		if err == nil && hours > 0 {
			if hours > 168 { // cap at 1 week
				hours = 168
			}
			since := time.Now().Add(-time.Duration(hours) * time.Hour)
			snapshots, err := s.store.GetMetricsHistory(since, hours)
			if err == nil {
				result["history"] = snapshots
			}
		}
	}

	jsonResponse(w, http.StatusOK, result)
}

// ---------------------------------------------------------------------------
// Agent API handlers
// ---------------------------------------------------------------------------

// handleAgents routes GET and POST /api/agents.
func (s *Server) handleAgents(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		agents := s.agentMgr.ListAgents()
		jsonResponse(w, http.StatusOK, map[string]interface{}{
			"agents": agents,
		})
	case http.MethodPost:
		var req agent.CreateAgentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			jsonError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		info, err := s.agentMgr.CreateAgent(req)
		if err != nil {
			jsonError(w, http.StatusBadRequest, err.Error())
			return
		}
		jsonResponse(w, http.StatusCreated, info)
	default:
		jsonError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// handleAgentByID routes GET and DELETE /api/agents/{id}.
func (s *Server) handleAgentByID(w http.ResponseWriter, r *http.Request) {
	// Extract agent ID from URL path: /api/agents/{id}
	agentID := strings.TrimPrefix(r.URL.Path, "/api/agents/")
	if agentID == "" {
		jsonError(w, http.StatusBadRequest, "agent ID is required")
		return
	}

	switch r.Method {
	case http.MethodGet:
		info, err := s.agentMgr.GetAgent(agentID)
		if err != nil {
			jsonError(w, http.StatusNotFound, err.Error())
			return
		}
		jsonResponse(w, http.StatusOK, info)
	case http.MethodDelete:
		if err := s.agentMgr.StopAgent(agentID); err != nil {
			jsonError(w, http.StatusNotFound, err.Error())
			return
		}
		jsonResponse(w, http.StatusOK, map[string]string{"status": "stopped"})
	default:
		jsonError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// StopAllAgents gracefully stops all running agents. Called during server shutdown.
func (s *Server) StopAllAgents() {
	if s.agentMgr != nil {
		s.agentMgr.StopAll()
	}
}

// handleCharacters routes GET and POST /api/characters to the correct handler.
func (s *Server) handleCharacters(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleListCharacters(w, r)
	case http.MethodPost:
		s.handleCreateCharacter(w, r)
	default:
		jsonError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// handleListCharacters handles GET /api/characters.
func (s *Server) handleListCharacters(w http.ResponseWriter, r *http.Request) {
	accountID := r.Context().Value(ctxAccountID).(int64)

	names, err := s.store.ListCharacters(accountID)
	if err != nil {
		log.Printf("list characters error: %v", err)
		jsonError(w, http.StatusInternalServerError, "failed to list characters")
		return
	}
	if names == nil {
		names = []string{}
	}

	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"characters": names,
	})
}

// handleCreateCharacter handles POST /api/characters.
func (s *Server) handleCreateCharacter(w http.ResponseWriter, r *http.Request) {
	accountID := r.Context().Value(ctxAccountID).(int64)

	var body struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if strings.TrimSpace(body.Name) == "" {
		jsonError(w, http.StatusBadRequest, "character name is required")
		return
	}

	// Generate a fresh level-1 character.
	char := game.GenerateCharacter(body.Name, 1, 1)
	char.EquipmentMap = map[int]models.Item{}
	char.Inventory = []models.Item{
		game.CreateHealthPotion("small"),
		game.CreateHealthPotion("small"),
		game.CreateHealthPotion("small"),
	}
	char.ResourceStorageMap = map[string]models.Resource{}
	game.GenerateLocationsForNewCharacter(&char)

	if err := s.store.SaveCharacter(accountID, char); err != nil {
		log.Printf("save character error: %v", err)
		jsonError(w, http.StatusInternalServerError, "failed to save character")
		return
	}

	jsonResponse(w, http.StatusCreated, map[string]interface{}{
		"name":  char.Name,
		"level": char.Level,
	})
}

// ---------------------------------------------------------------------------
// WebSocket handler
// ---------------------------------------------------------------------------

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 4096
)

// handleWebSocket handles GET /ws/game?token=<jwt>.
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	tokenStr := r.URL.Query().Get("token")
	if tokenStr == "" {
		jsonError(w, http.StatusUnauthorized, "missing token query parameter")
		return
	}

	accountID, username, err := s.auth.ValidateToken(tokenStr)
	if err != nil {
		jsonError(w, http.StatusUnauthorized, "invalid or expired token")
		return
	}

	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("websocket upgrade error for %s: %v", username, err)
		return
	}

	log.Printf("WebSocket connected: %s (account %d)", username, accountID)

	// Track online player count.
	if s.metrics != nil {
		s.metrics.OnlinePlayers.Add(1)
	}

	// Create a database-backed session for this account.
	sessionID, err := s.engine.CreateDBSession(accountID)
	if err != nil {
		log.Printf("create session error for %s: %v", username, err)
		conn.WriteMessage(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseInternalServerErr, "failed to create session"))
		conn.Close()
		return
	}

	// writeMu protects concurrent writes to the WebSocket connection.
	var writeMu sync.Mutex

	// writeJSON marshals data and writes it to the WebSocket connection safely.
	writeJSON := func(v interface{}) error {
		writeMu.Lock()
		defer writeMu.Unlock()
		conn.SetWriteDeadline(time.Now().Add(writeWait))
		return conn.WriteJSON(v)
	}

	// Register broadcast subscriber so this client receives server-wide events.
	s.engine.Subscribe(sessionID, func(resp engine.GameResponse) {
		writeJSON(resp)
	})

	// Send initial "init" command response so the client gets the main menu.
	initResp := s.engine.ProcessCommand(sessionID, engine.GameCommand{Type: "init", Value: ""})
	if err := writeJSON(initResp); err != nil {
		log.Printf("failed to send init response to %s: %v", username, err)
		s.engine.SaveSession(sessionID)
		s.engine.RemoveSession(sessionID)
		conn.Close()
		return
	}

	// Start ping ticker in a separate goroutine.
	done := make(chan struct{})
	go func() {
		ticker := time.NewTicker(pingPeriod)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				writeMu.Lock()
				conn.SetWriteDeadline(time.Now().Add(writeWait))
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					writeMu.Unlock()
					return
				}
				writeMu.Unlock()
			case <-done:
				return
			}
		}
	}()

	// Start harvest ticker goroutine — checks every 15 seconds.
	go func() {
		harvestTicker := time.NewTicker(15 * time.Second)
		defer harvestTicker.Stop()
		for {
			select {
			case <-harvestTicker.C:
				result := s.engine.ProcessHarvestTick(sessionID)
				if result == nil {
					continue
				}
				resp := engine.GameResponse{
					Type:     "harvest",
					Messages: result.Messages,
					State: &engine.StateData{
						Screen:  "harvest_tick",
						Player:  result.Player,
						Village: result.Village,
					},
				}
				if err := writeJSON(resp); err != nil {
					log.Printf("failed to send harvest tick to %s: %v", username, err)
					return
				}
			case <-done:
				return
			}
		}
	}()

	// Online players ticker — push every 15 seconds.
	go func() {
		presenceTicker := time.NewTicker(15 * time.Second)
		defer presenceTicker.Stop()
		for {
			select {
			case <-presenceTicker.C:
				players := s.engine.GetOnlinePlayers(sessionID)
				resp := engine.GameResponse{
					Type: "presence",
					State: &engine.StateData{
						OnlinePlayers: players,
					},
				}
				if err := writeJSON(resp); err != nil {
					return
				}
			case <-done:
				return
			}
		}
	}()

	// Configure read side.
	conn.SetReadLimit(maxMessageSize)
	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	// Read pump: process incoming commands until the connection closes.
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err,
				websocket.CloseGoingAway,
				websocket.CloseNormalClosure,
				websocket.CloseNoStatusReceived) {
				log.Printf("websocket read error for %s: %v", username, err)
			}
			break
		}

		var cmd engine.GameCommand
		if err := json.Unmarshal(message, &cmd); err != nil {
			errResp := engine.ErrorResponse(fmt.Sprintf("invalid command: %v", err))
			if writeErr := writeJSON(errResp); writeErr != nil {
				log.Printf("failed to write error response to %s: %v", username, writeErr)
				break
			}
			continue
		}

		resp := s.engine.ProcessCommand(sessionID, cmd)

		if err := writeJSON(resp); err != nil {
			log.Printf("failed to write response to %s: %v", username, err)
			break
		}
	}

	// Cleanup: stop ping goroutine, unsubscribe, save session, remove it, close connection.
	close(done)

	if s.metrics != nil {
		s.metrics.OnlinePlayers.Add(-1)
	}

	s.engine.Unsubscribe(sessionID)

	if err := s.engine.SaveSession(sessionID); err != nil {
		log.Printf("failed to save session for %s: %v", username, err)
	}
	s.engine.RemoveSession(sessionID)
	conn.Close()

	log.Printf("WebSocket disconnected: %s (account %d)", username, accountID)
}
