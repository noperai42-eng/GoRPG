package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"rpg-game/pkg/models"
)

// Account represents a user account in the database.
type Account struct {
	ID           int64
	Username     string
	PasswordHash string
	CreatedAt    time.Time
}

// Store wraps a *sql.DB and provides all database operations.
type Store struct {
	db *sql.DB
}

// NewStore opens the SQLite database at dbPath, configures it, creates tables,
// and returns a ready-to-use Store.
func NewStore(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Enable WAL mode for better concurrent read performance.
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to enable WAL mode: %w", err)
	}

	// Enable foreign key enforcement.
	if _, err := db.Exec("PRAGMA foreign_keys=ON"); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	s := &Store{db: db}
	if err := s.createTables(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return s, nil
}

// createTables creates all required tables if they do not already exist.
func (s *Store) createTables() error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS accounts (
			id            INTEGER PRIMARY KEY AUTOINCREMENT,
			username      TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			created_at    DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS characters (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			account_id  INTEGER NOT NULL REFERENCES accounts(id),
			name        TEXT NOT NULL,
			data        TEXT NOT NULL,
			created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(account_id, name)
		)`,
		`CREATE TABLE IF NOT EXISTS game_locations (
			name TEXT PRIMARY KEY,
			data TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS villages (
			id           INTEGER PRIMARY KEY AUTOINCREMENT,
			character_id INTEGER NOT NULL REFERENCES characters(id),
			name         TEXT NOT NULL,
			data         TEXT NOT NULL,
			updated_at   DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS quests (
			id   TEXT PRIMARY KEY,
			data TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS towns (
			name TEXT PRIMARY KEY,
			data TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS world_analytics (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			account_id INTEGER NOT NULL,
			character_name TEXT NOT NULL,
			event_type TEXT NOT NULL,
			event_data TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_analytics_type ON world_analytics(event_type)`,
		`CREATE TABLE IF NOT EXISTS leaderboards (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			character_name TEXT NOT NULL,
			account_id INTEGER NOT NULL,
			total_kills INTEGER DEFAULT 0,
			total_deaths INTEGER DEFAULT 0,
			bosses_killed INTEGER DEFAULT 0,
			pvp_wins INTEGER DEFAULT 0,
			player_level INTEGER DEFAULT 1,
			highest_combo INTEGER DEFAULT 0,
			dungeons_cleared INTEGER DEFAULT 0,
			floors_cleared INTEGER DEFAULT 0,
			rooms_explored INTEGER DEFAULT 0,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(account_id, character_name)
		)`,
	}

	for _, stmt := range statements {
		if _, err := s.db.Exec(stmt); err != nil {
			return fmt.Errorf("failed to execute %q: %w", stmt, err)
		}
	}

	// Migrations: add columns that may not exist in older databases.
	migrations := []string{
		`ALTER TABLE leaderboards ADD COLUMN dungeons_cleared INTEGER DEFAULT 0`,
		`ALTER TABLE leaderboards ADD COLUMN floors_cleared INTEGER DEFAULT 0`,
		`ALTER TABLE leaderboards ADD COLUMN rooms_explored INTEGER DEFAULT 0`,
	}
	for _, m := range migrations {
		s.db.Exec(m) // ignore "duplicate column" errors
	}
	return nil
}

// Close closes the underlying database connection.
func (s *Store) Close() error {
	return s.db.Close()
}

// ---------------------------------------------------------------------------
// Account methods
// ---------------------------------------------------------------------------

// CreateAccount inserts a new account and returns its ID.
func (s *Store) CreateAccount(username, passwordHash string) (int64, error) {
	result, err := s.db.Exec(
		"INSERT INTO accounts (username, password_hash) VALUES (?, ?)",
		username, passwordHash,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create account: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get account id: %w", err)
	}
	return id, nil
}

// GetAccountByUsername retrieves an account by its username.
// Returns nil and no error if the account is not found.
func (s *Store) GetAccountByUsername(username string) (*Account, error) {
	var acct Account
	err := s.db.QueryRow(
		"SELECT id, username, password_hash, created_at FROM accounts WHERE username = ?",
		username,
	).Scan(&acct.ID, &acct.Username, &acct.PasswordHash, &acct.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get account by username %q: %w", username, err)
	}
	return &acct, nil
}

// GetAccountByID retrieves an account by its ID.
// Returns nil and no error if the account is not found.
func (s *Store) GetAccountByID(id int64) (*Account, error) {
	var acct Account
	err := s.db.QueryRow(
		"SELECT id, username, password_hash, created_at FROM accounts WHERE id = ?",
		id,
	).Scan(&acct.ID, &acct.Username, &acct.PasswordHash, &acct.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get account by id %d: %w", id, err)
	}
	return &acct, nil
}

// ---------------------------------------------------------------------------
// Character methods
// ---------------------------------------------------------------------------

// SaveCharacter upserts a character for the given account. The character struct
// is serialized to JSON and stored in the data column.
func (s *Store) SaveCharacter(accountID int64, char models.Character) error {
	data, err := json.Marshal(char)
	if err != nil {
		return fmt.Errorf("failed to marshal character: %w", err)
	}

	_, err = s.db.Exec(
		`INSERT INTO characters (account_id, name, data, updated_at)
		 VALUES (?, ?, ?, CURRENT_TIMESTAMP)
		 ON CONFLICT(account_id, name)
		 DO UPDATE SET data = excluded.data, updated_at = CURRENT_TIMESTAMP`,
		accountID, char.Name, string(data),
	)
	if err != nil {
		return fmt.Errorf("failed to save character: %w", err)
	}
	return nil
}

// LoadCharacter retrieves a character by account ID and name, deserializing
// the JSON data column back into a models.Character.
func (s *Store) LoadCharacter(accountID int64, name string) (models.Character, error) {
	var data string
	err := s.db.QueryRow(
		"SELECT data FROM characters WHERE account_id = ? AND name = ?",
		accountID, name,
	).Scan(&data)
	if err != nil {
		return models.Character{}, fmt.Errorf("failed to load character %q: %w", name, err)
	}

	var char models.Character
	if err := json.Unmarshal([]byte(data), &char); err != nil {
		return models.Character{}, fmt.Errorf("failed to unmarshal character: %w", err)
	}
	return char, nil
}

// ListCharacters returns the names of all characters belonging to an account.
func (s *Store) ListCharacters(accountID int64) ([]string, error) {
	rows, err := s.db.Query(
		"SELECT name FROM characters WHERE account_id = ? ORDER BY name",
		accountID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list characters: %w", err)
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("failed to scan character name: %w", err)
		}
		names = append(names, name)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}
	return names, nil
}

// DeleteCharacter removes a character by account ID and name.
func (s *Store) DeleteCharacter(accountID int64, name string) error {
	result, err := s.db.Exec(
		"DELETE FROM characters WHERE account_id = ? AND name = ?",
		accountID, name,
	)
	if err != nil {
		return fmt.Errorf("failed to delete character: %w", err)
	}
	n, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if n == 0 {
		return fmt.Errorf("character %q not found for account %d", name, accountID)
	}
	return nil
}

// GetCharacterID returns the database row ID for a character identified by
// account ID and character name.
func (s *Store) GetCharacterID(accountID int64, name string) (int64, error) {
	var id int64
	err := s.db.QueryRow(
		"SELECT id FROM characters WHERE account_id = ? AND name = ?",
		accountID, name,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to get character id for %q: %w", name, err)
	}
	return id, nil
}

// ---------------------------------------------------------------------------
// Location methods
// ---------------------------------------------------------------------------

// SaveLocations persists a map of locations. Each location is individually
// upserted with its name as the primary key and data as a JSON blob.
func (s *Store) SaveLocations(locations map[string]models.Location) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT OR REPLACE INTO game_locations (name, data) VALUES (?, ?)")
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for name, loc := range locations {
		data, err := json.Marshal(loc)
		if err != nil {
			return fmt.Errorf("failed to marshal location %q: %w", name, err)
		}
		if _, err := stmt.Exec(name, string(data)); err != nil {
			return fmt.Errorf("failed to save location %q: %w", name, err)
		}
	}

	return tx.Commit()
}

// LoadLocations retrieves all locations from the database and returns them
// as a map keyed by location name.
func (s *Store) LoadLocations() (map[string]models.Location, error) {
	rows, err := s.db.Query("SELECT name, data FROM game_locations")
	if err != nil {
		return nil, fmt.Errorf("failed to query locations: %w", err)
	}
	defer rows.Close()

	locations := make(map[string]models.Location)
	for rows.Next() {
		var name, data string
		if err := rows.Scan(&name, &data); err != nil {
			return nil, fmt.Errorf("failed to scan location row: %w", err)
		}
		var loc models.Location
		if err := json.Unmarshal([]byte(data), &loc); err != nil {
			return nil, fmt.Errorf("failed to unmarshal location %q: %w", name, err)
		}
		locations[name] = loc
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}
	return locations, nil
}

// ---------------------------------------------------------------------------
// Village methods
// ---------------------------------------------------------------------------

// SaveVillage persists a village associated with a character row ID.
// The village struct is serialized to JSON.
func (s *Store) SaveVillage(characterID int64, village models.Village) error {
	data, err := json.Marshal(village)
	if err != nil {
		return fmt.Errorf("failed to marshal village: %w", err)
	}

	_, err = s.db.Exec(
		`INSERT INTO villages (character_id, name, data, updated_at)
		 VALUES (?, ?, ?, CURRENT_TIMESTAMP)
		 ON CONFLICT(id)
		 DO UPDATE SET data = excluded.data, updated_at = CURRENT_TIMESTAMP`,
		characterID, village.Name, string(data),
	)
	if err != nil {
		return fmt.Errorf("failed to save village: %w", err)
	}
	return nil
}

// LoadVillage retrieves a village by character row ID and village name.
func (s *Store) LoadVillage(characterID int64, villageName string) (models.Village, error) {
	var data string
	err := s.db.QueryRow(
		"SELECT data FROM villages WHERE character_id = ? AND name = ?",
		characterID, villageName,
	).Scan(&data)
	if err != nil {
		return models.Village{}, fmt.Errorf("failed to load village %q: %w", villageName, err)
	}

	var village models.Village
	if err := json.Unmarshal([]byte(data), &village); err != nil {
		return models.Village{}, fmt.Errorf("failed to unmarshal village: %w", err)
	}
	return village, nil
}

// LoadVillageByCharName retrieves a village by joining the characters and
// villages tables using account ID, character name, and village name.
func (s *Store) LoadVillageByCharName(accountID int64, charName string, villageName string) (models.Village, error) {
	var data string
	err := s.db.QueryRow(
		`SELECT v.data FROM villages v
		 JOIN characters c ON v.character_id = c.id
		 WHERE c.account_id = ? AND c.name = ? AND v.name = ?`,
		accountID, charName, villageName,
	).Scan(&data)
	if err != nil {
		return models.Village{}, fmt.Errorf("failed to load village %q for character %q: %w", villageName, charName, err)
	}

	var village models.Village
	if err := json.Unmarshal([]byte(data), &village); err != nil {
		return models.Village{}, fmt.Errorf("failed to unmarshal village: %w", err)
	}
	return village, nil
}

// ---------------------------------------------------------------------------
// Quest methods
// ---------------------------------------------------------------------------

// SaveQuests persists a map of quests. Each quest is individually upserted
// with its ID as the primary key and data as a JSON blob.
func (s *Store) SaveQuests(quests map[string]models.Quest) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT OR REPLACE INTO quests (id, data) VALUES (?, ?)")
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for id, quest := range quests {
		data, err := json.Marshal(quest)
		if err != nil {
			return fmt.Errorf("failed to marshal quest %q: %w", id, err)
		}
		if _, err := stmt.Exec(id, string(data)); err != nil {
			return fmt.Errorf("failed to save quest %q: %w", id, err)
		}
	}

	return tx.Commit()
}

// LoadQuests retrieves all quests from the database and returns them as a map
// keyed by quest ID.
func (s *Store) LoadQuests() (map[string]models.Quest, error) {
	rows, err := s.db.Query("SELECT id, data FROM quests")
	if err != nil {
		return nil, fmt.Errorf("failed to query quests: %w", err)
	}
	defer rows.Close()

	quests := make(map[string]models.Quest)
	for rows.Next() {
		var id, data string
		if err := rows.Scan(&id, &data); err != nil {
			return nil, fmt.Errorf("failed to scan quest row: %w", err)
		}
		var quest models.Quest
		if err := json.Unmarshal([]byte(data), &quest); err != nil {
			return nil, fmt.Errorf("failed to unmarshal quest %q: %w", id, err)
		}
		quests[id] = quest
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}
	return quests, nil
}

// ---------------------------------------------------------------------------
// Town methods
// ---------------------------------------------------------------------------

// SaveTown persists a town record. The town struct is serialized to JSON.
func (s *Store) SaveTown(town models.Town) error {
	data, err := json.Marshal(town)
	if err != nil {
		return fmt.Errorf("failed to marshal town: %w", err)
	}

	_, err = s.db.Exec(
		"INSERT OR REPLACE INTO towns (name, data) VALUES (?, ?)",
		town.Name, string(data),
	)
	if err != nil {
		return fmt.Errorf("failed to save town %q: %w", town.Name, err)
	}
	return nil
}

// LoadTown retrieves a town by name.
// Returns the town and a nil error, or an empty town and an error if not found.
func (s *Store) LoadTown(name string) (models.Town, error) {
	var data string
	err := s.db.QueryRow(
		"SELECT data FROM towns WHERE name = ?",
		name,
	).Scan(&data)
	if err == sql.ErrNoRows {
		return models.Town{}, fmt.Errorf("town %q not found", name)
	}
	if err != nil {
		return models.Town{}, fmt.Errorf("failed to load town %q: %w", name, err)
	}

	var town models.Town
	if err := json.Unmarshal([]byte(data), &town); err != nil {
		return models.Town{}, fmt.Errorf("failed to unmarshal town: %w", err)
	}
	return town, nil
}

// ---------------------------------------------------------------------------
// Analytics methods
// ---------------------------------------------------------------------------

// RecordAnalyticsEvent logs a world event for gossip and history.
func (s *Store) RecordAnalyticsEvent(accountID int64, charName, eventType, eventData string) error {
	_, err := s.db.Exec(
		"INSERT INTO world_analytics (account_id, character_name, event_type, event_data) VALUES (?, ?, ?, ?)",
		accountID, charName, eventType, eventData,
	)
	if err != nil {
		return fmt.Errorf("failed to record analytics event: %w", err)
	}
	return nil
}

// GetRecentEvents retrieves the most recent events of a given type.
// AnalyticsEvent represents a single analytics event with metadata.
type AnalyticsEvent struct {
	CharacterName string
	EventType     string
	EventData     string
}

func (s *Store) GetRecentEvents(eventType string, limit int) ([]AnalyticsEvent, error) {
	rows, err := s.db.Query(
		"SELECT character_name, event_type, event_data FROM world_analytics WHERE event_type = ? ORDER BY created_at DESC LIMIT ?",
		eventType, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query recent events: %w", err)
	}
	defer rows.Close()

	var events []AnalyticsEvent
	for rows.Next() {
		var evt AnalyticsEvent
		if err := rows.Scan(&evt.CharacterName, &evt.EventType, &evt.EventData); err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}
		events = append(events, evt)
	}
	return events, rows.Err()
}

// ---------------------------------------------------------------------------
// Leaderboard methods
// ---------------------------------------------------------------------------

// LeaderboardEntry represents a single leaderboard row.
type LeaderboardEntry struct {
	CharacterName   string `json:"character_name"`
	AccountID       int64  `json:"account_id"`
	TotalKills      int    `json:"total_kills"`
	TotalDeaths     int    `json:"total_deaths"`
	BossesKilled    int    `json:"bosses_killed"`
	PvPWins         int    `json:"pvp_wins"`
	PlayerLevel     int    `json:"player_level"`
	HighestCombo    int    `json:"highest_combo"`
	DungeonsCleared int    `json:"dungeons_cleared"`
	FloorsCleared   int    `json:"floors_cleared"`
	RoomsExplored   int    `json:"rooms_explored"`
}

// UpdateLeaderboard upserts the leaderboard row for a character.
func (s *Store) UpdateLeaderboard(accountID int64, charName string, stats models.CharacterStats, level int) error {
	_, err := s.db.Exec(
		`INSERT INTO leaderboards (character_name, account_id, total_kills, total_deaths, bosses_killed, pvp_wins, player_level, highest_combo, dungeons_cleared, floors_cleared, rooms_explored, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		 ON CONFLICT(account_id, character_name)
		 DO UPDATE SET total_kills=?, total_deaths=?, bosses_killed=?, pvp_wins=?, player_level=?, highest_combo=?, dungeons_cleared=?, floors_cleared=?, rooms_explored=?, updated_at=CURRENT_TIMESTAMP`,
		charName, accountID, stats.TotalKills, stats.TotalDeaths, stats.BossesKilled, stats.PvPWins, level, stats.HighestCombo, stats.DungeonsCleared, stats.FloorsCleared, stats.RoomsExplored,
		stats.TotalKills, stats.TotalDeaths, stats.BossesKilled, stats.PvPWins, level, stats.HighestCombo, stats.DungeonsCleared, stats.FloorsCleared, stats.RoomsExplored,
	)
	if err != nil {
		return fmt.Errorf("failed to update leaderboard: %w", err)
	}
	return nil
}

// GetLeaderboard returns the top entries for a given category.
func (s *Store) GetLeaderboard(category string, limit int) ([]LeaderboardEntry, error) {
	orderCol := "total_kills"
	switch category {
	case "kills":
		orderCol = "total_kills"
	case "level":
		orderCol = "player_level"
	case "bosses":
		orderCol = "bosses_killed"
	case "pvp_wins":
		orderCol = "pvp_wins"
	case "combo":
		orderCol = "highest_combo"
	case "dungeons":
		orderCol = "dungeons_cleared"
	case "floors":
		orderCol = "floors_cleared"
	case "rooms":
		orderCol = "rooms_explored"
	}

	query := fmt.Sprintf(
		"SELECT character_name, account_id, total_kills, total_deaths, bosses_killed, pvp_wins, player_level, highest_combo, dungeons_cleared, floors_cleared, rooms_explored FROM leaderboards ORDER BY %s DESC LIMIT ?",
		orderCol,
	)
	rows, err := s.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query leaderboard: %w", err)
	}
	defer rows.Close()

	var entries []LeaderboardEntry
	for rows.Next() {
		var e LeaderboardEntry
		if err := rows.Scan(&e.CharacterName, &e.AccountID, &e.TotalKills, &e.TotalDeaths, &e.BossesKilled, &e.PvPWins, &e.PlayerLevel, &e.HighestCombo, &e.DungeonsCleared, &e.FloorsCleared, &e.RoomsExplored); err != nil {
			return nil, fmt.Errorf("failed to scan leaderboard entry: %w", err)
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}
