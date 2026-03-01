package engine

import (
	"fmt"
	"os"
	"sync"
	"time"

	"strings"

	"rpg-game/pkg/data"
	"rpg-game/pkg/db"
	"rpg-game/pkg/game"
	"rpg-game/pkg/metrics"
	"rpg-game/pkg/models"
)

// Engine manages game sessions and dispatches commands to handlers.
type Engine struct {
	sessions    map[string]*GameSession
	store       *db.Store
	metrics     *metrics.MetricsCollector
	mu          sync.RWMutex
	subscribers map[string]func(GameResponse) // keyed by sessionID
	subMu       sync.RWMutex                  // separate mutex to avoid deadlock
}

// NewEngine creates a new game engine (file-based persistence only).
func NewEngine() *Engine {
	return &Engine{
		sessions:    make(map[string]*GameSession),
		subscribers: make(map[string]func(GameResponse)),
	}
}

// NewEngineWithStore creates a game engine backed by a SQLite store.
func NewEngineWithStore(store *db.Store, mc *metrics.MetricsCollector) *Engine {
	return &Engine{
		sessions:    make(map[string]*GameSession),
		store:       store,
		metrics:     mc,
		subscribers: make(map[string]func(GameResponse)),
	}
}

// CreateLocalSession loads or creates game state from a file and returns a session ID.
func (e *Engine) CreateLocalSession(saveFile string) (string, error) {
	gameState := models.GameState{CharactersMap: map[string]models.Character{}}

	if _, err := os.Stat(saveFile); err == nil {
		loaded, err := game.LoadGameStateFromFile(saveFile)
		if err != nil {
			return "", err
		}
		gameState = loaded
		if gameState.CharactersMap == nil {
			gameState.CharactersMap = make(map[string]models.Character)
		}
		if gameState.GameLocations == nil {
			gameState.GameLocations = make(map[string]models.Location)
		}
	} else {
		game.GenerateGameLocation(&gameState)
		player := game.GenerateCharacter("Temp", 1, 1)
		player.EquipmentMap = map[int]models.Item{}
		player.Inventory = []models.Item{
			game.CreateHealthPotion("small"),
			game.CreateHealthPotion("small"),
			game.CreateHealthPotion("small"),
		}
		player.ResourceStorageMap = map[string]models.Resource{}
		game.GenerateLocationsForNewCharacter(&player)
		gameState.CharactersMap[player.Name] = player
	}

	// Initialize quest system
	if gameState.AvailableQuests == nil {
		gameState.AvailableQuests = make(map[string]models.Quest)
		for id, quest := range data.StoryQuests {
			gameState.AvailableQuests[id] = quest
		}
	}

	sessionID := fmt.Sprintf("local-%d", time.Now().UnixNano())
	session := &GameSession{
		ID:        sessionID,
		State:     StateInit,
		GameState: &gameState,
		SaveFile:  saveFile,
	}

	e.mu.Lock()
	e.sessions[sessionID] = session
	e.mu.Unlock()

	return sessionID, nil
}

// CreateDBSession creates a session backed by SQLite for a given account.
// It loads the account's characters, locations, villages, and quests from the database.
func (e *Engine) CreateDBSession(accountID int64) (string, error) {
	if e.store == nil {
		return "", fmt.Errorf("engine has no database store configured")
	}

	gameState := models.GameState{
		CharactersMap: make(map[string]models.Character),
		GameLocations: make(map[string]models.Location),
		Villages:      make(map[string]models.Village),
	}

	// Load characters for this account.
	charNames, err := e.store.ListCharacters(accountID)
	if err != nil {
		return "", fmt.Errorf("failed to list characters: %w", err)
	}
	for _, name := range charNames {
		char, err := e.store.LoadCharacter(accountID, name)
		if err != nil {
			return "", fmt.Errorf("failed to load character %q: %w", name, err)
		}
		gameState.CharactersMap[name] = char
	}

	// Load locations.
	locations, err := e.store.LoadLocations()
	if err != nil {
		return "", fmt.Errorf("failed to load locations: %w", err)
	}
	if len(locations) > 0 {
		gameState.GameLocations = locations
		// Sync caps/types from code definitions and add any new locations.
		game.SyncLocationCaps(gameState.GameLocations, &gameState)
		// Enforce level/rarity caps on legacy monsters that should have migrated.
		game.EnforceLevelCaps(gameState.GameLocations, &gameState)
		if err := e.store.SaveLocations(gameState.GameLocations); err != nil {
			fmt.Printf("Failed to save synced locations: %v\n", err)
		}
	} else {
		// Generate initial locations if none exist.
		game.GenerateGameLocation(&gameState)
	}

	// Load quests.
	quests, err := e.store.LoadQuests()
	if err != nil {
		return "", fmt.Errorf("failed to load quests: %w", err)
	}
	if len(quests) > 0 {
		gameState.AvailableQuests = quests
	} else {
		gameState.AvailableQuests = make(map[string]models.Quest)
		for id, quest := range data.StoryQuests {
			gameState.AvailableQuests[id] = quest
		}
	}

	// If no characters exist, create a default one.
	if len(gameState.CharactersMap) == 0 {
		player := game.GenerateCharacter("Temp", 1, 1)
		player.EquipmentMap = map[int]models.Item{}
		player.Inventory = []models.Item{
			game.CreateHealthPotion("small"),
			game.CreateHealthPotion("small"),
			game.CreateHealthPotion("small"),
		}
		player.ResourceStorageMap = map[string]models.Resource{}
		game.GenerateLocationsForNewCharacter(&player)
		gameState.CharactersMap[player.Name] = player
	}

	sessionID := fmt.Sprintf("db-%d-%d", accountID, time.Now().UnixNano())
	session := &GameSession{
		ID:        sessionID,
		AccountID: accountID,
		State:     StateInit,
		GameState: &gameState,
	}

	e.mu.Lock()
	e.sessions[sessionID] = session
	e.mu.Unlock()

	return sessionID, nil
}

// ProcessCommand dispatches a command to the appropriate handler based on session state.
func (e *Engine) ProcessCommand(sessionID string, cmd GameCommand) GameResponse {
	e.mu.RLock()
	session, ok := e.sessions[sessionID]
	e.mu.RUnlock()

	if !ok {
		return ErrorResponse("Session not found")
	}

	// Navbar tab commands work regardless of current session state,
	// since the frontend tabs can send these from any screen.
	if cmd.Type == "select" {
		switch cmd.Value {
		case "home":
			// Atomic return to main menu from any state.
			// Saves village/town context if active.
			if session.SelectedVillage != nil {
				e.saveVillage(session)
				session.SelectedVillage = nil
			}
			session.SelectedTown = nil
			session.State = StateMainMenu
			if e.metrics != nil {
				e.metrics.RecordFeatureUse("home")
			}
			return BuildMainMenuResponse(session)
		case "hunt":
			// Go to hunt from any state — reset to main menu first, then route.
			if session.SelectedVillage != nil {
				e.saveVillage(session)
				session.SelectedVillage = nil
			}
			session.SelectedTown = nil
			session.State = StateMainMenu
			if e.metrics != nil {
				e.metrics.RecordFeatureUse("hunt")
			}
			return e.handleMainMenu(session, GameCommand{Type: "select", Value: "3"})
		case "harvest":
			// Go to harvest from any state.
			if session.SelectedVillage != nil {
				e.saveVillage(session)
				session.SelectedVillage = nil
			}
			session.SelectedTown = nil
			session.State = StateMainMenu
			if e.metrics != nil {
				e.metrics.RecordFeatureUse("harvest")
			}
			return e.handleMainMenu(session, GameCommand{Type: "select", Value: "1"})
		case "10":
			return e.handleMainMenu(session, cmd)
		case "11":
			return e.handleMainMenu(session, cmd)
		}
		// Direct arena challenge from leaderboard click.
		// Format: arena_challenge:<accountID>:<charName>
		if cmd.Type == "select" && strings.HasPrefix(cmd.Value, "arena_challenge:") {
			parts := strings.SplitN(cmd.Value, ":", 3)
			if len(parts) == 3 {
				if session.SelectedVillage != nil {
					e.saveVillage(session)
					session.SelectedVillage = nil
				}
				session.SelectedTown = nil
				return e.handleArenaDirectChallenge(session, parts[1], parts[2])
			}
		}
	}

	switch session.State {
	case StateInit:
		return e.handleInit(session)
	case StateMainMenu:
		return e.handleMainMenu(session, cmd)
	case StateCharacterCreate:
		return e.handleCharacterCreate(session, cmd)
	case StateCharacterSelect:
		return e.handleCharacterSelect(session, cmd)
	case StateHarvestSelect:
		return e.handleHarvestSelect(session, cmd)
	case StateHuntLocationSelect:
		return e.handleHuntLocationSelect(session, cmd)
	case StateHuntTracking:
		return e.handleHuntTracking(session, cmd)
	case StateCombat:
		return e.handleCombatAction(session, cmd)
	case StateCombatItemSelect:
		return e.handleCombatItemSelect(session, cmd)
	case StateCombatSkillSelect:
		return e.handleCombatSkillSelect(session, cmd)
	case StateCombatGuardPrompt:
		return e.handleCombatGuardPrompt(session, cmd)
	case StateCombatSkillReward:
		return e.handleCombatSkillReward(session, cmd)
	case StateAutoPlaySpeed:
		return e.handleAutoPlaySpeed(session, cmd)
	case StateAutoPlayMenu:
		return e.handleAutoPlayMenu(session, cmd)
	case StateQuestLog:
		return e.handleQuestLog(session, cmd)
	case StatePlayerStats:
		return e.handlePlayerStats(session, cmd)
	case StateDiscoveredLocations:
		return e.handleDiscoveredLocations(session, cmd)
	case StateLoadSave:
		return e.handleLoadSave(session, cmd)
	case StateLoadSaveCharSelect:
		return e.handleLoadSaveCharSelect(session, cmd)
	case StateBuildSelect:
		return e.handleBuildSelect(session, cmd)
	case StateVillageMain:
		return e.handleVillageMain(session, cmd)
	case StateVillageViewVillagers:
		return e.handleVillageViewVillagers(session, cmd)
	case StateVillageAssignTask:
		return e.handleVillageAssignTask(session, cmd)
	case StateVillageAssignResource:
		return e.handleVillageAssignResource(session, cmd)
	case StateVillageBatchAssign:
		return e.handleVillageBatchAssign(session, cmd)
	case StateVillageHireGuard:
		return e.handleVillageHireGuard(session, cmd)
	case StateVillageCrafting:
		return e.handleVillageCrafting(session, cmd)
	case StateVillageCraftPotion:
		return e.handleVillageCraftPotion(session, cmd)
	case StateVillageCraftArmor:
		return e.handleVillageCraftArmor(session, cmd)
	case StateVillageCraftWeapon:
		return e.handleVillageCraftWeapon(session, cmd)
	case StateVillageUpgradeSkill:
		return e.handleVillageUpgradeSkill(session, cmd)
	case StateVillageUpgradeConfirm:
		return e.handleVillageUpgradeConfirm(session, cmd)
	case StateVillageCraftScrolls:
		return e.handleVillageCraftScrolls(session, cmd)
	case StateVillageBuildDefense:
		return e.handleVillageBuildDefense(session, cmd)
	case StateVillageBuildWalls:
		return e.handleVillageBuildWalls(session, cmd)
	case StateVillageCraftTraps:
		return e.handleVillageCraftTraps(session, cmd)
	case StateVillageViewDefenses:
		return e.handleVillageViewDefenses(session, cmd)
	case StateVillageCheckTide:
		return e.handleVillageCheckTide(session, cmd)
	case StateVillageMonsterTide:
		return e.handleVillageMonsterTide(session, cmd)
	case StateVillageTideWave:
		return e.handleVillageTideWave(session, cmd)
	case StateVillageManageGuards:
		return e.handleVillageManageGuards(session, cmd)
	case StateVillageManageGuard:
		return e.handleVillageManageGuard(session, cmd)
	case StateVillageEquipGuard:
		return e.handleVillageEquipGuard(session, cmd)
	case StateVillageUnequipGuard:
		return e.handleVillageUnequipGuard(session, cmd)
	case StateVillageGiveItem:
		return e.handleVillageGiveItem(session, cmd)
	case StateVillageTakeItem:
		return e.handleVillageTakeItem(session, cmd)
	case StateVillageHealGuard:
		return e.handleVillageHealGuard(session, cmd)
	case StateVillageFortifications:
		return e.handleVillageFortifications(session, cmd)
	case StateVillageTraining:
		return e.handleVillageTraining(session, cmd)
	case StateVillageHealing:
		return e.handleVillageHealing(session, cmd)
	case StateGuideMain:
		return e.handleGuideMain(session, cmd)
	case StateGuideCombat, StateGuideSkills, StateGuideVillage,
		StateGuideCrafting, StateGuideMonsterDrops, StateGuideAutoPlay,
		StateGuideQuests:
		return e.handleGuideTopic(session, cmd)
	case StateTownMain:
		return e.handleTownMain(session, cmd)
	case StateTownInn:
		return e.handleTownInn(session, cmd)
	case StateTownInnSleep:
		return e.handleTownInnSleep(session, cmd)
	case StateTownInnHireGuard:
		return e.handleTownInnHireGuard(session, cmd)
	case StateTownInnViewGuests:
		return e.handleTownInnViewGuests(session, cmd)
	case StateTownInnGossip:
		return e.handleTownInnGossip(session, cmd)
	case StateTownInnGamble:
		return e.handleTownInnGamble(session, cmd)
	case StateTownInnGamblePlay:
		return e.handleTownInnGamblePlay(session, cmd)
	case StateTownInnHireFighter:
		return e.handleTownInnHireFighter(session, cmd)
	case StateTownMayor:
		return e.handleTownMayor(session, cmd)
	case StateTownMayorChallenge:
		return e.handleTownMayorChallenge(session, cmd)
	case StateTownMayorMenu:
		return e.handleTownMayorMenu(session, cmd)
	case StateTownMayorSetTax:
		return e.handleTownMayorSetTax(session, cmd)
	case StateTownMayorCreateQuest:
		return e.handleTownMayorCreateQuest(session, cmd)
	case StateTownMayorCreateQuestAmount:
		return e.handleTownMayorCreateQuestAmount(session, cmd)
	case StateTownMayorCreateQuestReward:
		return e.handleTownMayorCreateQuestReward(session, cmd)
	case StateTownMayorHireGuard:
		return e.handleTownMayorHireGuard(session, cmd)
	case StateTownMayorHireMonster:
		return e.handleTownMayorHireMonster(session, cmd)
	case StateTownFetchQuests:
		return e.handleTownFetchQuests(session, cmd)
	case StateTownTalkNPC:
		return e.handleTownTalkNPC(session, cmd)
	case StateTownNPCDialogue:
		return e.handleTownNPCDialogue(session, cmd)
	case StateTownNPCQuestBoard, StateTownNPCQuestDetail,
		StateTownNPCQuestAccept, StateTownNPCQuestTurnIn:
		return e.handleTownNPCQuestBoard(session, cmd)
	case StateMostWantedBoard:
		return e.handleMostWantedBoard(session, cmd)
	case StateMostWantedHunt:
		return e.handleMostWantedHunt(session, cmd)
	case StateArenaMain:
		return e.handleArenaMain(session, cmd)
	case StateArenaChallenge:
		return e.handleArenaChallenge(session, cmd)
	case StateArenaConfirm:
		return e.handleArenaConfirm(session, cmd)
	case StateDungeonSelect:
		return e.handleDungeonSelect(session, cmd)
	case StateDungeonFloorMap:
		return e.handleDungeonFloorMap(session, cmd)
	case StateDungeonGridMove:
		return e.handleDungeonMove(session, cmd)
	case StateDungeonRoom, StateDungeonTreasure, StateDungeonTrap,
		StateDungeonRest, StateDungeonMerchant:
		return e.handleDungeonRoom(session, cmd)
	case StateDungeonComplete, StateDungeonDefeat:
		session.State = StateMainMenu
		return BuildMainMenuResponse(session)
	default:
		return ErrorResponse(fmt.Sprintf("Unknown state: %s", session.State))
	}
}

// saveSession persists session state directly (for use by handlers that already have the session).
func (e *Engine) saveSession(session *GameSession) {
	if session.Player != nil {
		session.GameState.CharactersMap[session.Player.Name] = *session.Player
	}
	if e.store != nil && session.AccountID > 0 {
		e.saveSessionToDB(session)
		return
	}
	if session.SaveFile != "" {
		game.WriteGameStateToFile(*session.GameState, session.SaveFile)
	}
}

// SaveSession saves the current session state. Uses SQLite if available, otherwise file.
func (e *Engine) SaveSession(sessionID string) error {
	e.mu.RLock()
	session, ok := e.sessions[sessionID]
	e.mu.RUnlock()

	if !ok {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	if session.Player != nil {
		session.GameState.CharactersMap[session.Player.Name] = *session.Player
	}

	// If we have a store and an account ID, save to SQLite.
	if e.store != nil && session.AccountID > 0 {
		return e.saveSessionToDB(session)
	}

	// Otherwise fall back to file persistence.
	if session.SaveFile != "" {
		return game.WriteGameStateToFile(*session.GameState, session.SaveFile)
	}
	return nil
}

// saveSessionToDB persists all session data to the SQLite store.
func (e *Engine) saveSessionToDB(session *GameSession) error {
	// Save all characters.
	for _, char := range session.GameState.CharactersMap {
		if err := e.store.SaveCharacter(session.AccountID, char); err != nil {
			return fmt.Errorf("failed to save character %q: %w", char.Name, err)
		}
	}

	// Save locations.
	if len(session.GameState.GameLocations) > 0 {
		if err := e.store.SaveLocations(session.GameState.GameLocations); err != nil {
			return fmt.Errorf("failed to save locations: %w", err)
		}
	}

	// Save quests.
	if len(session.GameState.AvailableQuests) > 0 {
		if err := e.store.SaveQuests(session.GameState.AvailableQuests); err != nil {
			return fmt.Errorf("failed to save quests: %w", err)
		}
	}

	// Update leaderboard (always attempt even if earlier saves fail).
	if session.Player != nil {
		if err := e.store.UpdateLeaderboard(session.AccountID, session.Player.Name, session.Player.Stats, session.Player.Level); err != nil {
			fmt.Printf("[Leaderboard] Failed to update leaderboard for %s: %v\n", session.Player.Name, err)
		}
	}

	// Save villages.
	if session.GameState.Villages != nil && session.Player != nil {
		for vName, village := range session.GameState.Villages {
			charID, err := e.store.GetCharacterID(session.AccountID, session.Player.Name)
			if err != nil {
				return fmt.Errorf("failed to get character ID for village %q: %w", vName, err)
			}
			if err := e.store.SaveVillage(charID, village); err != nil {
				return fmt.Errorf("failed to save village %q: %w", vName, err)
			}
		}
	}

	return nil
}

// HarvestTickResult holds the result of a harvest tick for server push.
type HarvestTickResult struct {
	Messages []GameMessage
	Player   *PlayerState
	Village  *VillageView
}

// ProcessHarvestTick checks if harvest is due for the session and processes it.
// Returns nil if no harvest occurred.
func (e *Engine) ProcessHarvestTick(sessionID string) *HarvestTickResult {
	e.mu.RLock()
	session, ok := e.sessions[sessionID]
	e.mu.RUnlock()

	if !ok || session.Player == nil {
		return nil
	}

	villageName := session.Player.VillageName
	if villageName == "" || session.GameState.Villages == nil {
		return nil
	}

	village, exists := session.GameState.Villages[villageName]
	if !exists {
		return nil
	}

	if !game.HasActiveHarvesters(&village) || !game.ShouldHarvest(&village) {
		return nil
	}

	results := game.ProcessVillageResourceCollection(&village, session.Player)
	if len(results) == 0 {
		return nil
	}

	village.LastHarvestTime = time.Now().Unix()
	session.GameState.Villages[villageName] = village
	session.GameState.CharactersMap[session.Player.Name] = *session.Player

	msgs := []GameMessage{}
	for _, r := range results {
		msgs = append(msgs, Msg(fmt.Sprintf("%s collected %d %s", r.VillagerName, r.Amount, r.ResourceType), "loot"))
	}

	// Save
	e.SaveSession(sessionID)

	return &HarvestTickResult{
		Messages: msgs,
		Player:   MakePlayerState(session.Player),
		Village:  MakeVillageView(&village),
	}
}

// GetOnlinePlayers returns a list of online players, excluding the given session.
func (e *Engine) GetOnlinePlayers(excludeSessionID string) []OnlinePlayer {
	e.mu.RLock()
	defer e.mu.RUnlock()
	var players []OnlinePlayer
	for id, sess := range e.sessions {
		if id == excludeSessionID || sess.Player == nil {
			continue
		}
		players = append(players, OnlinePlayer{
			Name:     sess.Player.Name,
			Level:    sess.Player.Level,
			Activity: sessionActivity(sess.State),
		})
	}
	return players
}

// sessionActivity maps a session state to a friendly activity label.
func sessionActivity(state string) string {
	if strings.HasPrefix(state, "combat") {
		return "In Combat"
	}
	if strings.HasPrefix(state, "hunt") {
		return "Hunting"
	}
	if strings.HasPrefix(state, "dungeon") {
		return "Dungeon"
	}
	if strings.HasPrefix(state, "village") {
		return "Village"
	}
	if strings.HasPrefix(state, "town") {
		return "Town"
	}
	if strings.HasPrefix(state, "arena") {
		return "Arena"
	}
	return "Hub"
}

// EvolutionTickResult holds the results of a monster evolution tick.
type EvolutionTickResult struct {
	Events []game.EvolutionEvent
}

// ProcessEvolutionTick runs monster-vs-monster combat across all locations.
func (e *Engine) ProcessEvolutionTick() *EvolutionTickResult {
	if e.store == nil {
		return nil
	}

	locations, err := e.store.LoadLocations()
	if err != nil {
		fmt.Printf("[Evolution] Failed to load locations: %v\n", err)
		return nil
	}
	if len(locations) == 0 {
		return nil
	}

	// Migrate old monsters missing IDs
	game.MigrateMonsterIDs(locations)

	// Need a GameState for monster generation in ProcessLocationEvolution
	gs := &models.GameState{GameLocations: locations}

	var allEvents []game.EvolutionEvent
	for locName, loc := range locations {
		events := game.ProcessLocationEvolution(&loc, gs)
		locations[locName] = loc
		allEvents = append(allEvents, events...)
	}

	// Save updated locations
	if err := e.store.SaveLocations(locations); err != nil {
		fmt.Printf("[Evolution] Failed to save locations: %v\n", err)
		return nil
	}

	// Update all active sessions' GameLocations to keep them current
	e.mu.RLock()
	for _, sess := range e.sessions {
		sess.GameState.GameLocations = locations
	}
	e.mu.RUnlock()

	if len(allEvents) == 0 {
		return nil
	}
	return &EvolutionTickResult{Events: allEvents}
}

// AutoTideTickResult holds the results of an auto-tide processing tick.
type AutoTideTickResult struct {
	TidesProcessed int
}

// ProcessAutoTideTick checks all villages and runs auto-tides for those whose
// tide interval has elapsed.
func (e *Engine) ProcessAutoTideTick() *AutoTideTickResult {
	if e.store == nil {
		return nil
	}

	villages, err := e.store.LoadAllVillages()
	if err != nil {
		fmt.Printf("[AutoTide] Failed to load villages: %v\n", err)
		return nil
	}

	now := time.Now().Unix()
	tidesProcessed := 0

	for _, vwo := range villages {
		interval := int64(vwo.Village.TideInterval)
		if interval <= 0 {
			interval = 3600
		}
		if vwo.Village.LastTideTime+interval >= now {
			continue
		}

		// Load the owning character
		char, err := e.store.LoadCharacter(vwo.AccountID, vwo.CharacterName)
		if err != nil {
			fmt.Printf("[AutoTide] Failed to load character %s: %v\n", vwo.CharacterName, err)
			continue
		}

		// Run the auto-tide
		tideResult := game.ProcessAutoTide(&vwo.Village, &char)
		tidesProcessed++

		// Record tide outcome metric
		if e.metrics != nil {
			e.metrics.RecordTideOutcome(tideResult.Victory)
		}

		// Save village back to DB
		if err := e.store.SaveVillage(vwo.CharacterID, vwo.Village); err != nil {
			fmt.Printf("[AutoTide] Failed to save village for %s: %v\n", vwo.CharacterName, err)
		}

		// Save character back to DB
		if err := e.store.SaveCharacter(vwo.AccountID, char); err != nil {
			fmt.Printf("[AutoTide] Failed to save character %s: %v\n", vwo.CharacterName, err)
		}

		// Build broadcast messages with contextual categories
		msgs := []GameMessage{}
		for _, m := range tideResult.Messages {
			tag := "combat"
			if strings.HasPrefix(m, "---") || strings.HasPrefix(m, "-- ") || strings.HasPrefix(m, "Defenders:") {
				tag = "system"
			} else if strings.HasPrefix(m, "VICTORY") || strings.HasPrefix(m, "Strong defense") {
				tag = "loot"
			} else if strings.HasPrefix(m, "DEFEAT") || strings.HasPrefix(m, "Village level reset") || strings.HasPrefix(m, "The village must") {
				tag = "damage"
			} else if strings.Contains(m, "Guard casualties") || strings.Contains(m, "perished") || strings.Contains(m, "been lost") || strings.Contains(m, "destroyed") || strings.Contains(m, "looted") {
				tag = "debuff"
			} else if strings.Contains(m, "Wave") && strings.Contains(m, "result:") {
				tag = "narrative"
			}
			msgs = append(msgs, Msg(m, tag))
		}

		resp := GameResponse{
			Type:     "auto_tide",
			Messages: msgs,
			State: &StateData{
				Screen:  "auto_tide",
				Player:  MakePlayerState(&char),
				Village: MakeVillageView(&vwo.Village),
			},
		}

		// Send to the owning player's session if they're online
		e.broadcastToAccount(vwo.AccountID, resp)

		// Update in-memory session data for online players
		e.mu.RLock()
		for _, sess := range e.sessions {
			if sess.AccountID == vwo.AccountID && sess.Player != nil && sess.Player.Name == vwo.CharacterName {
				// Update village in session
				if sess.GameState.Villages != nil {
					sess.GameState.Villages[vwo.Village.Name] = vwo.Village
				}
				// Update character resources in session
				sess.Player.ResourceStorageMap = char.ResourceStorageMap
				sess.GameState.CharactersMap[char.Name] = char
			}
		}
		e.mu.RUnlock()
	}

	if tidesProcessed == 0 {
		return nil
	}
	if e.metrics != nil {
		e.metrics.RecordTideTick(tidesProcessed)
	}
	return &AutoTideTickResult{TidesProcessed: tidesProcessed}
}

// VillageManagerTickResult holds the results of a village manager tick.
type VillageManagerTickResult struct {
	VillagesManaged int
}

// ProcessVillageManagerTicks runs automated village upkeep for all villages.
func (e *Engine) ProcessVillageManagerTicks() *VillageManagerTickResult {
	if e.store == nil {
		return nil
	}

	villages, err := e.store.LoadAllVillages()
	if err != nil {
		fmt.Printf("[VillageManager] Failed to load villages: %v\n", err)
		return nil
	}

	villagesManaged := 0

	for _, vwo := range villages {
		// Load the owning character
		char, err := e.store.LoadCharacter(vwo.AccountID, vwo.CharacterName)
		if err != nil {
			fmt.Printf("[VillageManager] Failed to load character %s: %v\n", vwo.CharacterName, err)
			continue
		}

		// Run the village manager tick
		messages := game.ProcessVillageManagerTick(&vwo.Village, &char)
		if len(messages) == 0 {
			continue
		}
		villagesManaged++

		// Save village back to DB
		if err := e.store.SaveVillage(vwo.CharacterID, vwo.Village); err != nil {
			fmt.Printf("[VillageManager] Failed to save village for %s: %v\n", vwo.CharacterName, err)
		}

		// Save character back to DB
		if err := e.store.SaveCharacter(vwo.AccountID, char); err != nil {
			fmt.Printf("[VillageManager] Failed to save character %s: %v\n", vwo.CharacterName, err)
		}

		// Update in-memory session data for online players
		e.mu.RLock()
		for _, sess := range e.sessions {
			if sess.AccountID == vwo.AccountID && sess.Player != nil && sess.Player.Name == vwo.CharacterName {
				if sess.GameState.Villages != nil {
					sess.GameState.Villages[vwo.Village.Name] = vwo.Village
				}
				sess.Player.ResourceStorageMap = char.ResourceStorageMap
				sess.GameState.CharactersMap[char.Name] = char
			}
		}
		e.mu.RUnlock()
	}

	if villagesManaged == 0 {
		return nil
	}
	if e.metrics != nil {
		e.metrics.RecordVillageManagerTick(villagesManaged)
	}
	return &VillageManagerTickResult{VillagesManaged: villagesManaged}
}

// broadcastToAccount sends a response to all sessions belonging to the given account ID.
func (e *Engine) broadcastToAccount(accountID int64, resp GameResponse) {
	e.mu.RLock()
	var targetSessionIDs []string
	for id, sess := range e.sessions {
		if sess.AccountID == accountID {
			targetSessionIDs = append(targetSessionIDs, id)
		}
	}
	e.mu.RUnlock()

	e.subMu.RLock()
	defer e.subMu.RUnlock()
	for _, sid := range targetSessionIDs {
		if cb, ok := e.subscribers[sid]; ok {
			go cb(resp)
		}
	}
}

// GetMostWanted returns the top N most dangerous monsters across all locations.
func (e *Engine) GetMostWanted(limit int) []models.MostWantedEntry {
	if e.store == nil {
		return nil
	}
	locations, err := e.store.LoadLocations()
	if err != nil {
		return nil
	}
	return game.GetMostWanted(locations, limit)
}

// Subscribe registers a callback to receive broadcast messages for the given session.
func (e *Engine) Subscribe(sessionID string, callback func(GameResponse)) {
	e.subMu.Lock()
	e.subscribers[sessionID] = callback
	e.subMu.Unlock()
}

// Unsubscribe removes a broadcast callback for the given session.
func (e *Engine) Unsubscribe(sessionID string) {
	e.subMu.Lock()
	delete(e.subscribers, sessionID)
	e.subMu.Unlock()
}

// Broadcast sends a response to all subscribers except the excluded session.
func (e *Engine) Broadcast(excludeSessionID string, resp GameResponse) {
	e.subMu.RLock()
	defer e.subMu.RUnlock()
	for id, cb := range e.subscribers {
		if id == excludeSessionID {
			continue
		}
		go cb(resp)
	}
}

// RenameSessionCharacter renames a character within a session's game state.
// It re-keys the CharactersMap entry and updates the character's Name and VillageName.
func (e *Engine) RenameSessionCharacter(sessionID, oldName, newName string) error {
	e.mu.RLock()
	session, ok := e.sessions[sessionID]
	e.mu.RUnlock()

	if !ok {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	char, exists := session.GameState.CharactersMap[oldName]
	if !exists {
		return fmt.Errorf("character %q not found in session", oldName)
	}

	char.Name = newName
	char.VillageName = newName + "'s Village"

	delete(session.GameState.CharactersMap, oldName)
	session.GameState.CharactersMap[newName] = char

	return nil
}

// RemoveSession removes a session from the engine.
func (e *Engine) RemoveSession(sessionID string) {
	e.mu.Lock()
	delete(e.sessions, sessionID)
	e.mu.Unlock()
}

// BuildMainMenuResponse creates the main menu response.
func BuildMainMenuResponse(session *GameSession) GameResponse {
	msgs := []GameMessage{
		Msg(fmt.Sprintf("Playing as %s (Level %d)", session.Player.Name, session.Player.Level), "system"),
	}

	options := []MenuOption{
		Opt("1", "Harvest"),
		Opt("3", "Hunt"),
		Opt("4", "Discovered Locations"),
		Opt("5", "Player Stats"),
		Opt("7", "Player Guide"),
		Opt("8", "AUTO-PLAY MODE"),
		Opt("9", "Quest Log"),
		Opt("12", "Enter Dungeon"),
		Opt("13", "Bounty Board"),
		Opt("14", "Arena"),
		Opt("exit", "Exit Game"),
	}

	ps := MakePlayerStateWithLocations(session.Player, session.GameState)
	if ps != nil {
		ps.ActiveQuests = MakeQuestViews(session.Player, session.GameState)
		ps.CompletedQuests = MakeCompletedQuestViews(session.Player, session.GameState)
	}

	return GameResponse{
		Type:     "menu",
		Messages: msgs,
		State:    &StateData{Screen: "main_menu", Player: ps},
		Options:  options,
	}
}
