package game

import (
	"fmt"
	"math/rand"
	"time"

	"rpg-game/pkg/data"
	"rpg-game/pkg/models"
)

// GenerateNPC creates a random NPC with the given title.
func GenerateNPC(title string) models.NPCTownsfolk {
	firstName := data.NPCFirstNames[rand.Intn(len(data.NPCFirstNames))]
	lastName := data.NPCLastNames[rand.Intn(len(data.NPCLastNames))]
	archetype := data.NPCArchetypes[rand.Intn(len(data.NPCArchetypes))]

	return models.NPCTownsfolk{
		ID:    fmt.Sprintf("npc_%s_%d", title, time.Now().UnixNano()+int64(rand.Intn(100000))),
		Name:  firstName + " " + lastName,
		Title: title,
		Personality: models.NPCPersonality{
			Archetype:  archetype,
			Chattiness: 0.2 + rand.Float64()*0.8,
			Generosity: 0.1 + rand.Float64()*0.9,
			Courage:    0.1 + rand.Float64()*0.9,
			Curiosity:  0.1 + rand.Float64()*0.9,
		},
		Memory:        []models.NPCMemory{},
		Relationships: make(map[string]int),
		Level:         rand.Intn(10) + 1,
		Age:           20 + rand.Intn(50),
		IsAlive:       true,
		CurrentMood:   "neutral",
		QuestGiver:    title == "Blacksmith" || title == "Innkeeper" || title == "Scholar" || title == "Guard Captain" || title == "Merchant" || title == "Farmer",
		LocationName:  npcDefaultLocation(title),
		GoldCarried:   rand.Intn(50) + 10,
	}
}

// npcDefaultLocation returns the default location name for a given NPC title.
func npcDefaultLocation(title string) string {
	switch title {
	case "Innkeeper":
		return "inn"
	case "Blacksmith":
		return "forge"
	case "Merchant":
		return "market"
	case "Guard Captain":
		return "barracks"
	case "Scholar":
		return "library"
	case "Farmer":
		return "fields"
	case "Herbalist":
		return "apothecary"
	case "Fisherman":
		return "docks"
	case "Baker":
		return "bakery"
	case "Weaver":
		return "workshop"
	case "Hunter":
		return "outskirts"
	case "Healer":
		return "temple"
	case "Scribe":
		return "library"
	case "Miner":
		return "mines"
	case "Woodcutter":
		return "forest"
	default:
		return "town_square"
	}
}

// GenerateDefaultTownsfolk creates 8-12 starter NPCs for a new town.
func GenerateDefaultTownsfolk() []models.NPCTownsfolk {
	// Required roles that always appear
	required := []string{"Innkeeper", "Blacksmith", "Merchant", "Guard Captain", "Scholar", "Farmer", "Herbalist", "Fisherman"}

	townsfolk := make([]models.NPCTownsfolk, 0, 12)
	for _, title := range required {
		townsfolk = append(townsfolk, GenerateNPC(title))
	}

	// Add 2-4 random additional NPCs
	optional := []string{"Baker", "Weaver", "Hunter", "Healer", "Scribe", "Miner", "Woodcutter"}
	extras := 2 + rand.Intn(3)
	rand.Shuffle(len(optional), func(i, j int) { optional[i], optional[j] = optional[j], optional[i] })
	for i := 0; i < extras && i < len(optional); i++ {
		townsfolk = append(townsfolk, GenerateNPC(optional[i]))
	}

	return townsfolk
}

// GetNPCDialogue returns a contextual dialogue line from an NPC.
func GetNPCDialogue(npc *models.NPCTownsfolk, playerName string, context string) string {
	archetype := npc.Personality.Archetype

	// Get mood prefix
	moodPrefix := ""
	if moods, ok := data.NPCMoodDialogue[npc.CurrentMood]; ok && len(moods) > 0 {
		moodPrefix = moods[rand.Intn(len(moods))]
	}

	// Get dialogue line
	line := ""
	if archetypeDialogue, ok := data.NPCDialogue[archetype]; ok {
		if contextLines, ok := archetypeDialogue[context]; ok && len(contextLines) > 0 {
			line = contextLines[rand.Intn(len(contextLines))]
		}
	}

	// Fallback to friendly archetype if no dialogue found
	if line == "" {
		if archetypeDialogue, ok := data.NPCDialogue["friendly"]; ok {
			if contextLines, ok := archetypeDialogue[context]; ok && len(contextLines) > 0 {
				line = contextLines[rand.Intn(len(contextLines))]
			}
		}
	}

	if line == "" {
		line = "..."
	}

	// Personalize with player name based on relationship
	rel := npc.Relationships[playerName]
	if rel > 30 {
		line = fmt.Sprintf("Ah, %s! %s", playerName, line)
	}

	if moodPrefix != "" {
		return moodPrefix + " " + line
	}
	return line
}

// UpdateNPCRelationship adjusts an NPC's relationship with a player.
func UpdateNPCRelationship(npc *models.NPCTownsfolk, playerName string, delta int) {
	if npc.Relationships == nil {
		npc.Relationships = make(map[string]int)
	}
	val := npc.Relationships[playerName] + delta
	if val > 100 {
		val = 100
	}
	if val < -100 {
		val = -100
	}
	npc.Relationships[playerName] = val
}

// AddNPCMemory adds a memory to an NPC, FIFO capped at 50.
func AddNPCMemory(npc *models.NPCTownsfolk, memory models.NPCMemory) {
	npc.Memory = append(npc.Memory, memory)
	if len(npc.Memory) > 50 {
		npc.Memory = npc.Memory[len(npc.Memory)-50:]
	}
}

// ComputeNPCMood derives mood from recent memory sentiment.
func ComputeNPCMood(npc *models.NPCTownsfolk) string {
	if len(npc.Memory) == 0 {
		return "neutral"
	}

	// Average sentiment of last 10 memories
	count := 10
	if len(npc.Memory) < count {
		count = len(npc.Memory)
	}
	total := 0
	for i := len(npc.Memory) - count; i < len(npc.Memory); i++ {
		total += npc.Memory[i].Sentiment
	}
	avg := total / count

	switch {
	case avg >= 5:
		return "happy"
	case avg >= 1:
		return "neutral"
	case avg >= -2:
		return "neutral"
	case avg >= -5:
		return "sad"
	default:
		return "angry"
	}
}

// RelationshipLabel returns a human-readable label for a relationship value.
func RelationshipLabel(value int) string {
	switch {
	case value >= 80:
		return "Beloved"
	case value >= 50:
		return "Trusted Friend"
	case value >= 20:
		return "Friendly"
	case value >= 0:
		return "Neutral"
	case value >= -20:
		return "Wary"
	case value >= -50:
		return "Distrustful"
	default:
		return "Hostile"
	}
}
