package agent

import (
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"rpg-game/pkg/engine"
)

// AgentStatus represents the current state of an agent.
type AgentStatus string

const (
	AgentStatusIdle    AgentStatus = "idle"
	AgentStatusRunning AgentStatus = "running"
	AgentStatusStopped AgentStatus = "stopped"
	AgentStatusError   AgentStatus = "error"
)

// Agent is an autonomous AI player that interacts with the game engine.
type Agent struct {
	ID        string
	Name      string
	AccountID int64
	SessionID string
	Strategy  Strategy
	Status    AgentStatus
	MinDelay  time.Duration
	MaxDelay  time.Duration
	Error     string

	eng        *engine.Engine
	done       chan struct{}
	mu         sync.RWMutex
	stuckCount int
	turnCount  int
	lastScreen string
	startedAt  time.Time
}

// AgentInfo is a read-only snapshot of agent state for the REST API.
type AgentInfo struct {
	ID            string      `json:"id"`
	Name          string      `json:"name"`
	Strategy      string      `json:"strategy"`
	Status        AgentStatus `json:"status"`
	TurnCount     int         `json:"turn_count"`
	CurrentScreen string      `json:"current_screen"`
	StartedAt     time.Time   `json:"started_at"`
	AccountID     int64       `json:"account_id"`
	Error         string      `json:"error,omitempty"`
}

// Info returns a read-only snapshot of the agent's current state.
func (a *Agent) Info() AgentInfo {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return AgentInfo{
		ID:            a.ID,
		Name:          a.Name,
		Strategy:      a.Strategy.Name(),
		Status:        a.Status,
		TurnCount:     a.turnCount,
		CurrentScreen: a.lastScreen,
		StartedAt:     a.startedAt,
		AccountID:     a.AccountID,
		Error:         a.Error,
	}
}

// Run starts the agent's main loop. It should be called as a goroutine.
func (a *Agent) Run() {
	a.mu.Lock()
	a.Status = AgentStatusRunning
	a.startedAt = time.Now()
	a.done = make(chan struct{})
	a.mu.Unlock()

	defer func() {
		if r := recover(); r != nil {
			a.mu.Lock()
			a.Status = AgentStatusError
			a.Error = fmt.Sprintf("panic: %v", r)
			a.mu.Unlock()
			log.Printf("[Agent %s] Panic: %v", a.Name, r)
		}
	}()

	log.Printf("[Agent %s] Starting with strategy %s", a.Name, a.Strategy.Name())

	// Step 1: Send init command to get to character select / main menu.
	resp := a.eng.ProcessCommand(a.SessionID, engine.GameCommand{Type: "init", Value: ""})
	screen, options := a.extractState(resp)
	a.updateState(screen)

	log.Printf("[Agent %s] Init screen: %s", a.Name, screen)

	// Step 2: If at character select, select the agent's character.
	if screen == "character_select" {
		resp = a.eng.ProcessCommand(a.SessionID, engine.GameCommand{Type: "select", Value: a.Name})
		screen, options = a.extractState(resp)
		a.updateState(screen)
	}

	// Step 3: Main loop.
	for {
		select {
		case <-a.done:
			a.mu.Lock()
			a.Status = AgentStatusStopped
			a.mu.Unlock()
			log.Printf("[Agent %s] Stopped after %d turns", a.Name, a.turnCount)
			return
		default:
		}

		// Sleep with jitter.
		jitter := a.MinDelay
		if a.MaxDelay > a.MinDelay {
			jitter += time.Duration(rand.Int63n(int64(a.MaxDelay - a.MinDelay)))
		}
		select {
		case <-time.After(jitter):
		case <-a.done:
			a.mu.Lock()
			a.Status = AgentStatusStopped
			a.mu.Unlock()
			log.Printf("[Agent %s] Stopped after %d turns", a.Name, a.turnCount)
			return
		}

		// Decide next action.
		cmd := a.Strategy.Decide(screen, options)

		// Send command.
		resp = a.eng.ProcessCommand(a.SessionID, cmd)
		screen, options = a.extractState(resp)

		a.mu.Lock()
		a.turnCount++
		a.mu.Unlock()

		// Check for errors.
		if resp.Type == "error" {
			a.recover()
			resp = a.eng.ProcessCommand(a.SessionID, engine.GameCommand{Type: "init", Value: ""})
			screen, options = a.extractState(resp)
		}

		// Stuck detection: if same screen for 10+ consecutive commands, reset.
		a.mu.Lock()
		if screen == a.lastScreen {
			a.stuckCount++
		} else {
			a.stuckCount = 0
		}
		stuck := a.stuckCount >= 10
		a.mu.Unlock()

		if stuck {
			log.Printf("[Agent %s] Stuck on screen %s for 10 turns, resetting", a.Name, screen)
			a.recover()
			resp = a.eng.ProcessCommand(a.SessionID, engine.GameCommand{Type: "init", Value: ""})
			screen, options = a.extractState(resp)
		}

		a.updateState(screen)
	}
}

// Stop signals the agent to stop its main loop.
func (a *Agent) Stop() {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.done != nil {
		select {
		case <-a.done:
			// Already closed.
		default:
			close(a.done)
		}
	}
	a.Status = AgentStatusStopped
}

// extractState pulls the screen name and options from a GameResponse.
func (a *Agent) extractState(resp engine.GameResponse) (string, []engine.MenuOption) {
	screen := ""
	if resp.State != nil {
		screen = resp.State.Screen
	}
	return screen, resp.Options
}

// updateState records the current screen and resets stuck counter if changed.
func (a *Agent) updateState(screen string) {
	a.mu.Lock()
	a.lastScreen = screen
	a.mu.Unlock()
}

// recover sends the "home" navbar command to reset the agent to the main menu.
func (a *Agent) recover() {
	a.eng.ProcessCommand(a.SessionID, engine.GameCommand{Type: "select", Value: "home"})
	a.mu.Lock()
	a.stuckCount = 0
	a.mu.Unlock()
}
