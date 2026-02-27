package game

import (
	"fmt"
	"math/rand"

	"rpg-game/pkg/db"
	"rpg-game/pkg/models"
)

// GenerateGossip pulls recent analytics events and NPC memories to create gossip strings.
func GenerateGossip(town *models.Town, store *db.Store) []string {
	gossip := []string{}

	// Pull recent events from analytics if store is available
	if store != nil {
		events, err := store.GetRecentEvents("kill", 5)
		if err == nil {
			for _, evt := range events {
				gossip = append(gossip, fmt.Sprintf("Word is that %s slew a monster recently!", evt.CharacterName))
			}
		}

		pvpEvents, err := store.GetRecentEvents("pvp_win", 3)
		if err == nil {
			for _, evt := range pvpEvents {
				gossip = append(gossip, fmt.Sprintf("Did you hear? %s won a PvP battle!", evt.CharacterName))
			}
		}

		dungeonEvents, err := store.GetRecentEvents("dungeon_clear", 3)
		if err == nil {
			for _, evt := range dungeonEvents {
				gossip = append(gossip, fmt.Sprintf("%s conquered a dungeon! Impressive feat!", evt.CharacterName))
			}
		}
	}

	// Pull from NPC memories
	for _, npc := range town.Townsfolk {
		if !npc.IsAlive || len(npc.Memory) == 0 {
			continue
		}
		// Last memory from each NPC
		last := npc.Memory[len(npc.Memory)-1]
		if last.EventType == "met" {
			gossip = append(gossip, fmt.Sprintf("%s the %s was seen chatting with %s.", npc.Name, npc.Title, last.PlayerName))
		}
	}

	// Static fallback rumors
	if len(gossip) == 0 {
		gossip = staticRumors()
	}

	// Cap at 10 entries
	if len(gossip) > 10 {
		gossip = gossip[:10]
	}

	return gossip
}

func staticRumors() []string {
	return []string{
		"They say a dragon sleeps beneath the deepest dungeon floor.",
		"The blacksmith has been forging something unusual late at night.",
		"A merchant caravan is expected to arrive any day now.",
		"Strange sounds echo from the mines after dark.",
		"The herbalist found a rare moonpetal in the forest.",
		"An old map was discovered in the library archives.",
		"The guards have doubled patrols near the eastern road.",
		"A legendary warrior was spotted heading toward the tower.",
	}
}

// ResolveGamble resolves a dice gambling game. Returns won, narrative, payout.
func ResolveGamble(playerLevel int, bet int) (bool, string, int) {
	// Player rolls 3d6
	playerTotal := 0
	playerRolls := []int{}
	for i := 0; i < 3; i++ {
		r := rand.Intn(6) + 1
		playerTotal += r
		playerRolls = append(playerRolls, r)
	}

	// House rolls 3d6
	houseTotal := 0
	houseRolls := []int{}
	for i := 0; i < 3; i++ {
		r := rand.Intn(6) + 1
		houseTotal += r
		houseRolls = append(houseRolls, r)
	}

	narrative := fmt.Sprintf("You rolled %v = %d. House rolled %v = %d.", playerRolls, playerTotal, houseRolls, houseTotal)

	if playerTotal > houseTotal {
		payout := bet * 2
		return true, narrative + " You win!", payout
	} else if playerTotal == houseTotal {
		return false, narrative + " It's a tie! House wins ties.", 0
	}
	return false, narrative + " You lose!", 0
}

// GenerateNPCFighters creates 3-5 hireable NPC fighters scaled to average player level.
func GenerateNPCFighters(avgLevel int) []models.NPCFighter {
	count := 3 + rand.Intn(3)
	fighters := make([]models.NPCFighter, 0, count)

	specialties := []string{"tank", "dps", "healer"}

	for i := 0; i < count; i++ {
		level := avgLevel + rand.Intn(5) - 2
		if level < 1 {
			level = 1
		}
		specialty := specialties[rand.Intn(len(specialties))]
		cost := level * 25

		name := fmt.Sprintf("%s %s",
			[]string{"Iron", "Swift", "Shadow", "Storm", "Fire", "Frost", "Stone", "Wild"}[rand.Intn(8)],
			[]string{"Blade", "Fist", "Shield", "Arrow", "Axe", "Mace", "Staff", "Spear"}[rand.Intn(8)])

		fighters = append(fighters, models.NPCFighter{
			NPCID:     fmt.Sprintf("fighter_%d_%d", i, rand.Intn(100000)),
			Name:      name,
			Level:     level,
			HireCost:  cost,
			Specialty: specialty,
		})
	}

	return fighters
}
