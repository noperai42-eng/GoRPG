package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"rpg-game/pkg/auth"
	"rpg-game/pkg/db"
	"rpg-game/pkg/game"
)

func main() {
	jsonFile := flag.String("json", "gamestate.json", "path to input JSON game state file")
	dbFile := flag.String("db", "game.db", "path to output SQLite database file")
	username := flag.String("username", "admin", "username for the default account")
	password := flag.String("password", "changeme", "password for the default account")
	flag.Parse()

	// Load the JSON game state.
	fmt.Printf("Loading game state from %s ...\n", *jsonFile)
	gs, err := game.LoadGameStateFromFile(*jsonFile)
	if err != nil {
		log.Fatalf("Failed to load game state: %v", err)
	}

	// Open (or create) the SQLite database.
	fmt.Printf("Opening database %s ...\n", *dbFile)
	store, err := db.NewStore(*dbFile)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer store.Close()

	// Hash the password and create a default account.
	hash, err := auth.HashPassword(*password)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}
	accountID, err := store.CreateAccount(*username, hash)
	if err != nil {
		log.Fatalf("Failed to create account %q: %v", *username, err)
	}
	fmt.Printf("Created account %q (id=%d)\n", *username, accountID)

	// Migrate characters.
	charCount := 0
	// charIDs maps character name -> database row ID for village linking.
	charIDs := make(map[string]int64)
	for name, char := range gs.CharactersMap {
		if err := store.SaveCharacter(accountID, char); err != nil {
			fmt.Fprintf(os.Stderr, "WARNING: failed to save character %q: %v\n", name, err)
			continue
		}
		cid, err := store.GetCharacterID(accountID, name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "WARNING: failed to get character ID for %q: %v\n", name, err)
			continue
		}
		charIDs[name] = cid
		charCount++
	}
	fmt.Printf("Migrated %d character(s)\n", charCount)

	// Migrate locations.
	if gs.GameLocations != nil && len(gs.GameLocations) > 0 {
		if err := store.SaveLocations(gs.GameLocations); err != nil {
			log.Fatalf("Failed to save locations: %v", err)
		}
		fmt.Printf("Migrated %d location(s)\n", len(gs.GameLocations))
	} else {
		fmt.Println("No locations to migrate")
	}

	// Migrate villages, matching each to its owning character.
	villageCount := 0
	if gs.Villages != nil {
		for vName, village := range gs.Villages {
			// Try to find the owning character by matching the village name
			// against each character's VillageName field.
			var ownerCharID int64
			found := false
			for charName, char := range gs.CharactersMap {
				if char.VillageName == vName {
					if cid, ok := charIDs[charName]; ok {
						ownerCharID = cid
						found = true
						break
					}
				}
			}
			if !found {
				// Fall back: if only one character exists, assign to it.
				if len(charIDs) == 1 {
					for _, cid := range charIDs {
						ownerCharID = cid
						found = true
					}
				}
			}
			if !found {
				fmt.Fprintf(os.Stderr, "WARNING: could not find owning character for village %q, skipping\n", vName)
				continue
			}
			if err := store.SaveVillage(ownerCharID, village); err != nil {
				fmt.Fprintf(os.Stderr, "WARNING: failed to save village %q: %v\n", vName, err)
				continue
			}
			villageCount++
		}
	}
	fmt.Printf("Migrated %d village(s)\n", villageCount)

	// Migrate quests.
	if gs.AvailableQuests != nil && len(gs.AvailableQuests) > 0 {
		if err := store.SaveQuests(gs.AvailableQuests); err != nil {
			log.Fatalf("Failed to save quests: %v", err)
		}
		fmt.Printf("Migrated %d quest(s)\n", len(gs.AvailableQuests))
	} else {
		fmt.Println("No quests to migrate")
	}

	// Summary.
	fmt.Println()
	fmt.Println("=== Migration Summary ===")
	fmt.Printf("  Account:    %s (id=%d)\n", *username, accountID)
	fmt.Printf("  Characters: %d\n", charCount)
	fmt.Printf("  Locations:  %d\n", len(gs.GameLocations))
	fmt.Printf("  Villages:   %d\n", villageCount)
	fmt.Printf("  Quests:     %d\n", len(gs.AvailableQuests))
	fmt.Printf("  Database:   %s\n", *dbFile)
	fmt.Println("Migration complete.")
}
