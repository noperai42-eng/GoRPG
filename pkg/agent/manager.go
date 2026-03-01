package agent

import (
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"rpg-game/pkg/auth"
	"rpg-game/pkg/db"
	"rpg-game/pkg/engine"
)

// CreateAgentRequest is the JSON body for creating a new agent.
type CreateAgentRequest struct {
	Name     string `json:"name"`
	Strategy string `json:"strategy"`
	MinDelay int    `json:"min_delay_ms"`
	MaxDelay int    `json:"max_delay_ms"`
}

// DefaultAgents is the roster of bot agents spawned automatically on server boot.
var DefaultAgents = []CreateAgentRequest{
	{Name: "Grimjaw", Strategy: "hunter", MinDelay: 500, MaxDelay: 2000},
	{Name: "Thornveil", Strategy: "hunter", MinDelay: 500, MaxDelay: 2000},
	{Name: "Petalfoot", Strategy: "harvester", MinDelay: 500, MaxDelay: 2000},
	{Name: "Shadowdelve", Strategy: "dungeon_crawler", MinDelay: 500, MaxDelay: 2000},
	{Name: "Ironclash", Strategy: "arena_grinder", MinDelay: 500, MaxDelay: 2000},
	{Name: "Wanderlux", Strategy: "completionist", MinDelay: 500, MaxDelay: 2000},
	{Name: "Villoria", Strategy: "village_manager", MinDelay: 500, MaxDelay: 2000},
}

// Manager handles the lifecycle of AI agents.
type Manager struct {
	agents    map[string]*Agent
	engine    *engine.Engine
	store     *db.Store
	mu        sync.RWMutex
	maxAgents int
}

// NewManager creates a new agent Manager.
func NewManager(eng *engine.Engine, store *db.Store, maxAgents int) *Manager {
	if maxAgents <= 0 {
		maxAgents = 20
	}
	return &Manager{
		agents:    make(map[string]*Agent),
		engine:    eng,
		store:     store,
		maxAgents: maxAgents,
	}
}

// CreateAgent creates a bot DB account, engine session, and starts the agent goroutine.
func (m *Manager) CreateAgent(req CreateAgentRequest) (*AgentInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.agents) >= m.maxAgents {
		return nil, fmt.Errorf("maximum number of agents (%d) reached", m.maxAgents)
	}

	// Validate strategy.
	strategy := NewStrategy(req.Strategy)
	if strategy == nil {
		return nil, fmt.Errorf("unknown strategy: %q (valid: hunter, harvester, dungeon_crawler, arena_grinder, completionist, village_manager)", req.Strategy)
	}

	if req.Name == "" {
		return nil, fmt.Errorf("agent name is required")
	}

	// Check for duplicate agent name.
	agentID := fmt.Sprintf("agent_%s_%d", req.Name, rand.Intn(100000))
	for _, a := range m.agents {
		if a.Name == req.Name {
			return nil, fmt.Errorf("agent with name %q already exists", req.Name)
		}
	}

	// Set defaults for delay.
	minDelay := time.Duration(req.MinDelay) * time.Millisecond
	maxDelay := time.Duration(req.MaxDelay) * time.Millisecond
	if minDelay <= 0 {
		minDelay = 200 * time.Millisecond
	}
	if maxDelay <= 0 {
		maxDelay = 1000 * time.Millisecond
	}
	if maxDelay < minDelay {
		maxDelay = minDelay
	}

	// Create a bot DB account.
	botUsername := fmt.Sprintf("bot_%s_%d", req.Name, rand.Intn(100000))
	botPassword := fmt.Sprintf("bot_pass_%d_%d", time.Now().UnixNano(), rand.Intn(100000))
	passwordHash, err := auth.HashPassword(botPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to hash bot password: %w", err)
	}

	accountID, err := m.store.CreateAccount(botUsername, passwordHash)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot account: %w", err)
	}

	log.Printf("[AgentManager] Created bot account: %s (ID: %d)", botUsername, accountID)

	// Create an engine session for this account.
	sessionID, err := m.engine.CreateDBSession(accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to create engine session: %w", err)
	}

	log.Printf("[AgentManager] Created session: %s for agent %s", sessionID, req.Name)

	// Rename the default "Temp" character to the agent's name.
	if err := m.engine.RenameSessionCharacter(sessionID, "Temp", req.Name); err != nil {
		log.Printf("[AgentManager] Failed to rename character for agent %s: %v", req.Name, err)
		// Non-fatal: agent will still work with "Temp" name.
	}

	agent := &Agent{
		ID:        agentID,
		Name:      req.Name,
		AccountID: accountID,
		SessionID: sessionID,
		Strategy:  strategy,
		Status:    AgentStatusIdle,
		MinDelay:  minDelay,
		MaxDelay:  maxDelay,
		eng:       m.engine,
	}

	m.agents[agentID] = agent

	// Start the agent goroutine.
	go agent.Run()

	info := agent.Info()
	return &info, nil
}

// StopAgent stops an agent by ID and cleans up its resources.
func (m *Manager) StopAgent(agentID string) error {
	m.mu.Lock()
	agent, exists := m.agents[agentID]
	if !exists {
		m.mu.Unlock()
		return fmt.Errorf("agent %q not found", agentID)
	}
	delete(m.agents, agentID)
	m.mu.Unlock()

	agent.Stop()

	// Save and clean up the engine session.
	if err := m.engine.SaveSession(agent.SessionID); err != nil {
		log.Printf("[AgentManager] Failed to save session for agent %s: %v", agent.Name, err)
	}
	m.engine.RemoveSession(agent.SessionID)

	log.Printf("[AgentManager] Stopped agent %s (ID: %s)", agent.Name, agentID)
	return nil
}

// GetAgent returns info for a single agent.
func (m *Manager) GetAgent(agentID string) (*AgentInfo, error) {
	m.mu.RLock()
	agent, exists := m.agents[agentID]
	m.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("agent %q not found", agentID)
	}

	info := agent.Info()
	return &info, nil
}

// ListAgents returns info for all agents.
func (m *Manager) ListAgents() []AgentInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	agents := make([]AgentInfo, 0, len(m.agents))
	for _, a := range m.agents {
		agents = append(agents, a.Info())
	}
	return agents
}

// StopAll stops all running agents gracefully.
func (m *Manager) StopAll() {
	m.mu.Lock()
	agentsCopy := make(map[string]*Agent, len(m.agents))
	for k, v := range m.agents {
		agentsCopy[k] = v
	}
	m.agents = make(map[string]*Agent)
	m.mu.Unlock()

	for _, agent := range agentsCopy {
		agent.Stop()
		if err := m.engine.SaveSession(agent.SessionID); err != nil {
			log.Printf("[AgentManager] Failed to save session for agent %s: %v", agent.Name, err)
		}
		m.engine.RemoveSession(agent.SessionID)
	}

	log.Printf("[AgentManager] Stopped all %d agents", len(agentsCopy))
}

// SpawnDefaultAgents creates all default bot agents from the roster.
// It should be called as a goroutine during server startup.
func (m *Manager) SpawnDefaultAgents() {
	// Small delay to let the server finish initializing.
	time.Sleep(5 * time.Second)

	log.Printf("[AgentManager] Spawning default agents...")

	for _, req := range DefaultAgents {
		info, err := m.CreateAgent(req)
		if err != nil {
			log.Printf("[AgentManager] Failed to spawn default agent %s: %v", req.Name, err)
			continue
		}
		log.Printf("[AgentManager] Spawned default agent %s (ID: %s, strategy: %s)", info.Name, info.ID, info.Strategy)
	}

	log.Printf("[AgentManager] Default agent spawn complete")

	// Start the crash recovery monitor.
	go m.monitorAgents()
}

// monitorAgents periodically checks for crashed agents and respawns them.
func (m *Manager) monitorAgents() {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		m.mu.RLock()
		var crashed []CreateAgentRequest
		var crashedIDs []string
		for id, a := range m.agents {
			a.mu.RLock()
			status := a.Status
			a.mu.RUnlock()
			if status == AgentStatusError {
				crashed = append(crashed, CreateAgentRequest{
					Name:     a.Name,
					Strategy: a.Strategy.Name(),
					MinDelay: int(a.MinDelay / time.Millisecond),
					MaxDelay: int(a.MaxDelay / time.Millisecond),
				})
				crashedIDs = append(crashedIDs, id)
			}
		}
		m.mu.RUnlock()

		for i, req := range crashed {
			log.Printf("[AgentManager] Recovering crashed agent %s (ID: %s)", req.Name, crashedIDs[i])
			// StopAgent cleans up the old session and removes from map.
			if err := m.StopAgent(crashedIDs[i]); err != nil {
				log.Printf("[AgentManager] Failed to stop crashed agent %s: %v", req.Name, err)
			}
			// Respawn with the same config.
			info, err := m.CreateAgent(req)
			if err != nil {
				log.Printf("[AgentManager] Failed to respawn agent %s: %v", req.Name, err)
				continue
			}
			log.Printf("[AgentManager] Respawned agent %s (new ID: %s)", info.Name, info.ID)
		}
	}
}
