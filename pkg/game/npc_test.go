package game

import (
	"testing"

	"rpg-game/pkg/models"
)

func TestGenerateNPC(t *testing.T) {
	titles := []string{"Innkeeper", "Blacksmith", "Merchant", "Guard Captain", "Farmer"}

	for _, title := range titles {
		t.Run("title_"+title, func(t *testing.T) {
			npc := GenerateNPC(title)

			if npc.Name == "" {
				t.Error("expected non-empty Name")
			}
			if npc.Title != title {
				t.Errorf("expected Title %q, got %q", title, npc.Title)
			}
			if npc.ID == "" {
				t.Error("expected non-empty ID")
			}
			if npc.Personality.Archetype == "" {
				t.Error("expected non-empty Personality.Archetype")
			}
			if npc.Personality.Chattiness == 0 {
				t.Error("expected non-zero Personality.Chattiness")
			}
			if npc.Personality.Generosity == 0 {
				t.Error("expected non-zero Personality.Generosity")
			}
			if npc.Personality.Courage == 0 {
				t.Error("expected non-zero Personality.Courage")
			}
			if npc.Personality.Curiosity == 0 {
				t.Error("expected non-zero Personality.Curiosity")
			}
			if !npc.IsAlive {
				t.Error("expected IsAlive to be true")
			}
			if npc.CurrentMood != "neutral" {
				t.Errorf("expected CurrentMood %q, got %q", "neutral", npc.CurrentMood)
			}
			if npc.Level < 1 || npc.Level > 10 {
				t.Errorf("expected Level in [1,10], got %d", npc.Level)
			}
			if npc.Age < 20 || npc.Age > 69 {
				t.Errorf("expected Age in [20,69], got %d", npc.Age)
			}
			if npc.Relationships == nil {
				t.Error("expected Relationships map to be initialized")
			}
			if npc.Memory == nil {
				t.Error("expected Memory slice to be initialized")
			}
			if npc.GoldCarried < 10 || npc.GoldCarried > 59 {
				t.Errorf("expected GoldCarried in [10,59], got %d", npc.GoldCarried)
			}
		})
	}
}

func TestGenerateNPC_UniqueIDs(t *testing.T) {
	ids := make(map[string]bool)
	for i := 0; i < 20; i++ {
		npc := GenerateNPC("Merchant")
		if ids[npc.ID] {
			t.Errorf("duplicate NPC ID generated: %s", npc.ID)
		}
		ids[npc.ID] = true
	}
}

func TestGenerateDefaultTownsfolk(t *testing.T) {
	townsfolk := GenerateDefaultTownsfolk()

	if len(townsfolk) < 8 || len(townsfolk) > 12 {
		t.Errorf("expected 8-12 NPCs, got %d", len(townsfolk))
	}

	// Check that required roles are always present
	required := map[string]bool{
		"Innkeeper":    false,
		"Blacksmith":   false,
		"Merchant":     false,
		"Guard Captain": false,
		"Scholar":      false,
		"Farmer":       false,
		"Herbalist":    false,
		"Fisherman":    false,
	}

	for _, npc := range townsfolk {
		if _, ok := required[npc.Title]; ok {
			required[npc.Title] = true
		}
	}

	for title, found := range required {
		if !found {
			t.Errorf("required NPC %q not found in townsfolk", title)
		}
	}
}

func TestGenerateDefaultTownsfolk_Consistency(t *testing.T) {
	// Run multiple times to exercise the random optional NPC selection
	for i := 0; i < 10; i++ {
		townsfolk := GenerateDefaultTownsfolk()
		if len(townsfolk) < 8 {
			t.Errorf("iteration %d: expected at least 8 NPCs, got %d", i, len(townsfolk))
		}
		if len(townsfolk) > 12 {
			t.Errorf("iteration %d: expected at most 12 NPCs, got %d", i, len(townsfolk))
		}

		// Every NPC should have a valid name and title
		for _, npc := range townsfolk {
			if npc.Name == "" {
				t.Error("found NPC with empty name")
			}
			if npc.Title == "" {
				t.Error("found NPC with empty title")
			}
			if !npc.IsAlive {
				t.Errorf("NPC %q should be alive", npc.Name)
			}
		}
	}
}

func TestGetNPCDialogue(t *testing.T) {
	// Create NPCs with specific archetypes to verify different dialogue
	archetypes := []string{"friendly", "grumpy", "mysterious", "jovial", "scholarly", "cautious"}

	for _, arch := range archetypes {
		t.Run("archetype_"+arch, func(t *testing.T) {
			npc := models.NPCTownsfolk{
				Name:  "Test NPC",
				Title: "Merchant",
				Personality: models.NPCPersonality{
					Archetype: arch,
				},
				Relationships: make(map[string]int),
				CurrentMood:   "neutral",
			}

			dialogue := GetNPCDialogue(&npc, "Hero", "greeting")
			if dialogue == "" {
				t.Errorf("expected non-empty dialogue for archetype %q, context %q", arch, "greeting")
			}
		})
	}
}

func TestGetNPCDialogue_DifferentArchetypesProduceDifferentDialogue(t *testing.T) {
	// Collect dialogue samples from different archetypes. Over many runs, different
	// archetypes should produce at least some distinct dialogue lines.
	dialogueSets := make(map[string]map[string]bool)

	archetypes := []string{"friendly", "grumpy", "mysterious"}
	for _, arch := range archetypes {
		dialogueSets[arch] = make(map[string]bool)
		npc := models.NPCTownsfolk{
			Name:  "Test NPC",
			Title: "Merchant",
			Personality: models.NPCPersonality{
				Archetype: arch,
			},
			Relationships: make(map[string]int),
			CurrentMood:   "neutral",
		}

		for i := 0; i < 50; i++ {
			line := GetNPCDialogue(&npc, "Hero", "greeting")
			dialogueSets[arch][line] = true
		}
	}

	// Grumpy and friendly should have at least one unique line not shared with each other
	friendlyOnly := false
	for line := range dialogueSets["friendly"] {
		if !dialogueSets["grumpy"][line] {
			friendlyOnly = true
			break
		}
	}
	grumpyOnly := false
	for line := range dialogueSets["grumpy"] {
		if !dialogueSets["friendly"][line] {
			grumpyOnly = true
			break
		}
	}

	if !friendlyOnly && !grumpyOnly {
		t.Error("expected friendly and grumpy archetypes to have at least some different dialogue")
	}
}

func TestGetNPCDialogue_HighRelationshipIncludesPlayerName(t *testing.T) {
	npc := models.NPCTownsfolk{
		Name:  "Test NPC",
		Title: "Innkeeper",
		Personality: models.NPCPersonality{
			Archetype: "friendly",
		},
		Relationships: map[string]int{"Hero": 50},
		CurrentMood:   "neutral",
	}

	// With relationship > 30, the dialogue should include the player name
	found := false
	for i := 0; i < 50; i++ {
		dialogue := GetNPCDialogue(&npc, "Hero", "greeting")
		if len(dialogue) > 0 {
			// The dialogue should contain the player's name when rel > 30
			if contains(dialogue, "Hero") {
				found = true
				break
			}
		}
	}
	if !found {
		t.Error("expected dialogue to include player name when relationship > 30")
	}
}

func TestGetNPCDialogue_FallbackOnUnknownArchetype(t *testing.T) {
	npc := models.NPCTownsfolk{
		Name:  "Test NPC",
		Title: "Merchant",
		Personality: models.NPCPersonality{
			Archetype: "nonexistent_archetype",
		},
		Relationships: make(map[string]int),
		CurrentMood:   "neutral",
	}

	// Should fall back to "friendly" archetype dialogue
	dialogue := GetNPCDialogue(&npc, "Hero", "greeting")
	if dialogue == "" || dialogue == "..." {
		t.Error("expected fallback dialogue for unknown archetype, got empty or ellipsis")
	}
}

func TestGetNPCDialogue_MoodPrefix(t *testing.T) {
	npc := models.NPCTownsfolk{
		Name:  "Test NPC",
		Title: "Merchant",
		Personality: models.NPCPersonality{
			Archetype: "friendly",
		},
		Relationships: make(map[string]int),
		CurrentMood:   "happy",
	}

	// Happy mood should sometimes prepend a mood prefix
	foundPrefix := false
	happyPrefixes := []string{"*smiles warmly*", "*cheerfully*", "*in high spirits*"}
	for i := 0; i < 100; i++ {
		dialogue := GetNPCDialogue(&npc, "Hero", "greeting")
		for _, prefix := range happyPrefixes {
			if contains(dialogue, prefix) {
				foundPrefix = true
				break
			}
		}
		if foundPrefix {
			break
		}
	}
	if !foundPrefix {
		t.Error("expected happy mood to produce a mood prefix in dialogue")
	}
}

func TestUpdateNPCRelationship(t *testing.T) {
	t.Run("basic_increase", func(t *testing.T) {
		npc := models.NPCTownsfolk{
			Relationships: make(map[string]int),
		}
		UpdateNPCRelationship(&npc, "Hero", 10)
		if npc.Relationships["Hero"] != 10 {
			t.Errorf("expected relationship 10, got %d", npc.Relationships["Hero"])
		}
	})

	t.Run("basic_decrease", func(t *testing.T) {
		npc := models.NPCTownsfolk{
			Relationships: make(map[string]int),
		}
		UpdateNPCRelationship(&npc, "Hero", -15)
		if npc.Relationships["Hero"] != -15 {
			t.Errorf("expected relationship -15, got %d", npc.Relationships["Hero"])
		}
	})

	t.Run("accumulate", func(t *testing.T) {
		npc := models.NPCTownsfolk{
			Relationships: make(map[string]int),
		}
		UpdateNPCRelationship(&npc, "Hero", 20)
		UpdateNPCRelationship(&npc, "Hero", 30)
		if npc.Relationships["Hero"] != 50 {
			t.Errorf("expected relationship 50, got %d", npc.Relationships["Hero"])
		}
	})

	t.Run("clamp_at_100", func(t *testing.T) {
		npc := models.NPCTownsfolk{
			Relationships: make(map[string]int),
		}
		UpdateNPCRelationship(&npc, "Hero", 150)
		if npc.Relationships["Hero"] != 100 {
			t.Errorf("expected relationship clamped at 100, got %d", npc.Relationships["Hero"])
		}
	})

	t.Run("clamp_at_negative_100", func(t *testing.T) {
		npc := models.NPCTownsfolk{
			Relationships: make(map[string]int),
		}
		UpdateNPCRelationship(&npc, "Hero", -200)
		if npc.Relationships["Hero"] != -100 {
			t.Errorf("expected relationship clamped at -100, got %d", npc.Relationships["Hero"])
		}
	})

	t.Run("clamp_upper_via_accumulation", func(t *testing.T) {
		npc := models.NPCTownsfolk{
			Relationships: make(map[string]int),
		}
		UpdateNPCRelationship(&npc, "Hero", 90)
		UpdateNPCRelationship(&npc, "Hero", 20)
		if npc.Relationships["Hero"] != 100 {
			t.Errorf("expected relationship clamped at 100 after accumulation, got %d", npc.Relationships["Hero"])
		}
	})

	t.Run("clamp_lower_via_accumulation", func(t *testing.T) {
		npc := models.NPCTownsfolk{
			Relationships: make(map[string]int),
		}
		UpdateNPCRelationship(&npc, "Hero", -80)
		UpdateNPCRelationship(&npc, "Hero", -30)
		if npc.Relationships["Hero"] != -100 {
			t.Errorf("expected relationship clamped at -100 after accumulation, got %d", npc.Relationships["Hero"])
		}
	})

	t.Run("nil_relationships_map", func(t *testing.T) {
		npc := models.NPCTownsfolk{
			Relationships: nil,
		}
		UpdateNPCRelationship(&npc, "Hero", 25)
		if npc.Relationships["Hero"] != 25 {
			t.Errorf("expected relationship 25 after nil map init, got %d", npc.Relationships["Hero"])
		}
	})

	t.Run("multiple_players", func(t *testing.T) {
		npc := models.NPCTownsfolk{
			Relationships: make(map[string]int),
		}
		UpdateNPCRelationship(&npc, "Hero", 40)
		UpdateNPCRelationship(&npc, "Villain", -30)
		if npc.Relationships["Hero"] != 40 {
			t.Errorf("expected Hero relationship 40, got %d", npc.Relationships["Hero"])
		}
		if npc.Relationships["Villain"] != -30 {
			t.Errorf("expected Villain relationship -30, got %d", npc.Relationships["Villain"])
		}
	})
}

func TestAddNPCMemory(t *testing.T) {
	t.Run("add_single_memory", func(t *testing.T) {
		npc := models.NPCTownsfolk{
			Memory: []models.NPCMemory{},
		}
		mem := models.NPCMemory{
			EventType:   "trade",
			PlayerName:  "Hero",
			Description: "Traded 5 iron bars",
			Timestamp:   1000,
			Sentiment:   3,
		}
		AddNPCMemory(&npc, mem)
		if len(npc.Memory) != 1 {
			t.Errorf("expected 1 memory, got %d", len(npc.Memory))
		}
		if npc.Memory[0].EventType != "trade" {
			t.Errorf("expected EventType %q, got %q", "trade", npc.Memory[0].EventType)
		}
	})

	t.Run("add_multiple_memories", func(t *testing.T) {
		npc := models.NPCTownsfolk{
			Memory: []models.NPCMemory{},
		}
		for i := 0; i < 10; i++ {
			AddNPCMemory(&npc, models.NPCMemory{
				EventType:   "event",
				Description: "Memory",
				Timestamp:   int64(i),
				Sentiment:   1,
			})
		}
		if len(npc.Memory) != 10 {
			t.Errorf("expected 10 memories, got %d", len(npc.Memory))
		}
	})

	t.Run("fifo_cap_at_50", func(t *testing.T) {
		npc := models.NPCTownsfolk{
			Memory: []models.NPCMemory{},
		}
		for i := 0; i < 60; i++ {
			AddNPCMemory(&npc, models.NPCMemory{
				EventType:   "event",
				Description: "Memory",
				Timestamp:   int64(i),
				Sentiment:   1,
			})
		}
		if len(npc.Memory) != 50 {
			t.Errorf("expected 50 memories after FIFO cap, got %d", len(npc.Memory))
		}
		// The oldest memories (timestamps 0-9) should have been evicted
		if npc.Memory[0].Timestamp != 10 {
			t.Errorf("expected oldest remaining memory timestamp 10, got %d", npc.Memory[0].Timestamp)
		}
		// The newest memory should be the last one added
		if npc.Memory[49].Timestamp != 59 {
			t.Errorf("expected newest memory timestamp 59, got %d", npc.Memory[49].Timestamp)
		}
	})

	t.Run("exactly_50_no_eviction", func(t *testing.T) {
		npc := models.NPCTownsfolk{
			Memory: []models.NPCMemory{},
		}
		for i := 0; i < 50; i++ {
			AddNPCMemory(&npc, models.NPCMemory{
				EventType:   "event",
				Description: "Memory",
				Timestamp:   int64(i),
				Sentiment:   1,
			})
		}
		if len(npc.Memory) != 50 {
			t.Errorf("expected exactly 50 memories, got %d", len(npc.Memory))
		}
		if npc.Memory[0].Timestamp != 0 {
			t.Errorf("expected first memory timestamp 0, got %d", npc.Memory[0].Timestamp)
		}
	})

	t.Run("51st_memory_evicts_first", func(t *testing.T) {
		npc := models.NPCTownsfolk{
			Memory: []models.NPCMemory{},
		}
		for i := 0; i < 51; i++ {
			AddNPCMemory(&npc, models.NPCMemory{
				EventType:   "event",
				Description: "Memory",
				Timestamp:   int64(i),
				Sentiment:   1,
			})
		}
		if len(npc.Memory) != 50 {
			t.Errorf("expected 50 memories, got %d", len(npc.Memory))
		}
		if npc.Memory[0].Timestamp != 1 {
			t.Errorf("expected oldest memory timestamp 1 after eviction, got %d", npc.Memory[0].Timestamp)
		}
	})
}

func TestComputeNPCMood(t *testing.T) {
	t.Run("no_memories_returns_neutral", func(t *testing.T) {
		npc := models.NPCTownsfolk{
			Memory: []models.NPCMemory{},
		}
		mood := ComputeNPCMood(&npc)
		if mood != "neutral" {
			t.Errorf("expected mood %q with no memories, got %q", "neutral", mood)
		}
	})

	t.Run("positive_memories_happy", func(t *testing.T) {
		npc := models.NPCTownsfolk{
			Memory: []models.NPCMemory{},
		}
		// Add 10 very positive memories (sentiment 10 each, avg = 10 >= 5 => happy)
		for i := 0; i < 10; i++ {
			AddNPCMemory(&npc, models.NPCMemory{
				EventType: "gift",
				Sentiment: 10,
				Timestamp: int64(i),
			})
		}
		mood := ComputeNPCMood(&npc)
		if mood != "happy" {
			t.Errorf("expected mood %q with positive memories, got %q", "happy", mood)
		}
	})

	t.Run("negative_memories_sad", func(t *testing.T) {
		npc := models.NPCTownsfolk{
			Memory: []models.NPCMemory{},
		}
		// Add memories with average sentiment -3 (>= -5, < -2 => sad)
		for i := 0; i < 10; i++ {
			AddNPCMemory(&npc, models.NPCMemory{
				EventType: "insult",
				Sentiment: -3,
				Timestamp: int64(i),
			})
		}
		mood := ComputeNPCMood(&npc)
		if mood != "sad" {
			t.Errorf("expected mood %q with mildly negative memories, got %q", "sad", mood)
		}
	})

	t.Run("very_negative_memories_angry", func(t *testing.T) {
		npc := models.NPCTownsfolk{
			Memory: []models.NPCMemory{},
		}
		// Add memories with average sentiment -10 (< -5 => angry)
		for i := 0; i < 10; i++ {
			AddNPCMemory(&npc, models.NPCMemory{
				EventType: "attack",
				Sentiment: -10,
				Timestamp: int64(i),
			})
		}
		mood := ComputeNPCMood(&npc)
		if mood != "angry" {
			t.Errorf("expected mood %q with very negative memories, got %q", "angry", mood)
		}
	})

	t.Run("mildly_positive_neutral", func(t *testing.T) {
		npc := models.NPCTownsfolk{
			Memory: []models.NPCMemory{},
		}
		// Add memories with average sentiment 2 (>= 1, < 5 => neutral)
		for i := 0; i < 10; i++ {
			AddNPCMemory(&npc, models.NPCMemory{
				EventType: "chat",
				Sentiment: 2,
				Timestamp: int64(i),
			})
		}
		mood := ComputeNPCMood(&npc)
		if mood != "neutral" {
			t.Errorf("expected mood %q with mildly positive memories, got %q", "neutral", mood)
		}
	})

	t.Run("mildly_negative_neutral", func(t *testing.T) {
		npc := models.NPCTownsfolk{
			Memory: []models.NPCMemory{},
		}
		// Add memories with average sentiment -1 (>= -2, < 1 => neutral)
		for i := 0; i < 10; i++ {
			AddNPCMemory(&npc, models.NPCMemory{
				EventType: "minor_grievance",
				Sentiment: -1,
				Timestamp: int64(i),
			})
		}
		mood := ComputeNPCMood(&npc)
		if mood != "neutral" {
			t.Errorf("expected mood %q with mildly negative memories, got %q", "neutral", mood)
		}
	})

	t.Run("uses_last_10_memories_only", func(t *testing.T) {
		npc := models.NPCTownsfolk{
			Memory: []models.NPCMemory{},
		}
		// Add 20 very negative memories, then 10 very positive
		for i := 0; i < 20; i++ {
			AddNPCMemory(&npc, models.NPCMemory{
				EventType: "attack",
				Sentiment: -10,
				Timestamp: int64(i),
			})
		}
		for i := 20; i < 30; i++ {
			AddNPCMemory(&npc, models.NPCMemory{
				EventType: "gift",
				Sentiment: 10,
				Timestamp: int64(i),
			})
		}
		mood := ComputeNPCMood(&npc)
		if mood != "happy" {
			t.Errorf("expected mood %q based on last 10 positive memories, got %q", "happy", mood)
		}
	})

	t.Run("fewer_than_10_memories", func(t *testing.T) {
		npc := models.NPCTownsfolk{
			Memory: []models.NPCMemory{},
		}
		// Only 3 memories, all very positive
		for i := 0; i < 3; i++ {
			AddNPCMemory(&npc, models.NPCMemory{
				EventType: "gift",
				Sentiment: 8,
				Timestamp: int64(i),
			})
		}
		mood := ComputeNPCMood(&npc)
		if mood != "happy" {
			t.Errorf("expected mood %q with 3 positive memories, got %q", "happy", mood)
		}
	})
}

func TestRelationshipLabel(t *testing.T) {
	tests := []struct {
		value    int
		expected string
	}{
		{100, "Beloved"},
		{80, "Beloved"},
		{79, "Trusted Friend"},
		{50, "Trusted Friend"},
		{49, "Friendly"},
		{20, "Friendly"},
		{19, "Neutral"},
		{0, "Neutral"},
		{-1, "Wary"},
		{-20, "Wary"},
		{-21, "Distrustful"},
		{-50, "Distrustful"},
		{-51, "Hostile"},
		{-100, "Hostile"},
	}

	for _, tc := range tests {
		t.Run(tc.expected, func(t *testing.T) {
			label := RelationshipLabel(tc.value)
			if label != tc.expected {
				t.Errorf("RelationshipLabel(%d) = %q, want %q", tc.value, label, tc.expected)
			}
		})
	}
}

// contains checks if substr is present in s.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchSubstring(s, substr)
}

func searchSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
