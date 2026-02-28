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
		return nil, fmt.Errorf("unknown strategy: %q (valid: hunter, harvester, dungeon_crawler, arena_grinder, completionist)", req.Strategy)
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
