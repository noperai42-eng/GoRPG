package engine

import (
	"fmt"
	"os"
	"sync"
	"time"

	"rpg-game/pkg/data"
	"rpg-game/pkg/db"
	"rpg-game/pkg/game"
	"rpg-game/pkg/models"
)

// Engine manages game sessions and dispatches commands to handlers.
type Engine struct {
	sessions map[string]*GameSession
	store    *db.Store
	mu       sync.RWMutex
}

// NewEngine creates a new game engine (file-based persistence only).
func NewEngine() *Engine {
	return &Engine{
		sessions: make(map[string]*GameSession),
	}
}

// NewEngineWithStore creates a game engine backed by a SQLite store.
func NewEngineWithStore(store *db.Store) *Engine {
	return &Engine{
		sessions: make(map[string]*GameSession),
		store:    store,
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
		case "10":
			return e.handleMainMenu(session, cmd)
		case "11":
			return e.handleMainMenu(session, cmd)
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
	case StateHuntCountSelect:
		return e.handleHuntCountSelect(session, cmd)
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
	default:
		return ErrorResponse(fmt.Sprintf("Unknown state: %s", session.State))
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
