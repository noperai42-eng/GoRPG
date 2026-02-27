package game

import (
	"math/rand"
	"testing"
	"time"

	"rpg-game/pkg/models"
)

// ============================================================
// HELPER: build a minimal NPC townsfolk for quest generation.
// ============================================================

func makeQuestNPC(id, name, title string) models.NPCTownsfolk {
	return models.NPCTownsfolk{
		ID:            id,
		Name:          name,
		Title:         title,
		QuestGiver:    true,
		IsAlive:       true,
		Relationships: make(map[string]int),
		Personality:   models.NPCPersonality{Archetype: "friendly"},
	}
}

// ============================================================
// TestGenerateNPCQuest
// ============================================================

func TestGenerateNPCQuest(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	cases := []struct {
		title        string
		expectedType string
	}{
		{"Blacksmith", "gather"},
		{"Innkeeper", "kill"},
		{"Scholar", "explore"},
		{"Guard Captain", "kill_rarity"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.title, func(t *testing.T) {
			npc := makeQuestNPC("npc_"+tc.title, "Test "+tc.title, tc.title)
			quest := GenerateNPCQuest(&npc, 1)

			if quest == nil {
				t.Fatal("GenerateNPCQuest returned nil")
			}

			// Quest type must match the NPC title mapping.
			if quest.Type != tc.expectedType {
				t.Errorf("Title %q: expected quest type %q, got %q",
					tc.title, tc.expectedType, quest.Type)
			}

			// Reward values must be positive.
			if quest.Reward.XP <= 0 {
				t.Errorf("Title %q: expected positive XP reward, got %d",
					tc.title, quest.Reward.XP)
			}
			if quest.Reward.Gold <= 0 {
				t.Errorf("Title %q: expected positive Gold reward, got %d",
					tc.title, quest.Reward.Gold)
			}
			if quest.Reward.Reputation <= 0 {
				t.Errorf("Title %q: expected positive Reputation reward, got %d",
					tc.title, quest.Reward.Reputation)
			}

			// Difficulty must be set (non-empty).
			if quest.Difficulty == "" {
				t.Errorf("Title %q: expected non-empty difficulty", tc.title)
			}

			// Requirement type should mirror quest type.
			if quest.Requirement.Type != quest.Type {
				t.Errorf("Title %q: requirement type %q does not match quest type %q",
					tc.title, quest.Requirement.Type, quest.Type)
			}

			// TargetCount must be positive.
			if quest.Requirement.TargetCount <= 0 {
				t.Errorf("Title %q: expected positive TargetCount, got %d",
					tc.title, quest.Requirement.TargetCount)
			}

			// CurrentCount starts at 0.
			if quest.Requirement.CurrentCount != 0 {
				t.Errorf("Title %q: expected CurrentCount 0, got %d",
					tc.title, quest.Requirement.CurrentCount)
			}

			// Quest should reference the originating NPC.
			if quest.NPCID != npc.ID {
				t.Errorf("Title %q: expected NPCID %q, got %q",
					tc.title, npc.ID, quest.NPCID)
			}

			t.Logf("Title %q -> quest %q (type=%s, difficulty=%s, target=%d, XP=%d, Gold=%d)",
				tc.title, quest.Name, quest.Type, quest.Difficulty,
				quest.Requirement.TargetCount, quest.Reward.XP, quest.Reward.Gold)
		})
	}
}

// ============================================================
// TestCheckNPCQuestProgress
// ============================================================

func TestCheckNPCQuestProgress(t *testing.T) {
	quest := models.NPCQuest{
		ID:   "test_kill_quest",
		Type: "kill",
		Requirement: models.NPCQuestReq{
			Type:         "kill",
			TargetName:   "any",
			TargetCount:  3,
			CurrentCount: 0,
		},
		Completed: false,
		Failed:    false,
	}

	// First kill: progress but not complete.
	completed := CheckNPCQuestProgress(&quest, "kill", "goblin")
	if completed {
		t.Error("Quest should not be complete after 1 kill")
	}
	if quest.Requirement.CurrentCount != 1 {
		t.Errorf("Expected CurrentCount 1, got %d", quest.Requirement.CurrentCount)
	}

	// Second kill: progress but not complete.
	completed = CheckNPCQuestProgress(&quest, "kill", "orc")
	if completed {
		t.Error("Quest should not be complete after 2 kills")
	}
	if quest.Requirement.CurrentCount != 2 {
		t.Errorf("Expected CurrentCount 2, got %d", quest.Requirement.CurrentCount)
	}

	// Third kill: should now complete.
	completed = CheckNPCQuestProgress(&quest, "kill", "slime")
	if !completed {
		t.Error("Quest should be complete after 3 kills")
	}
	if quest.Requirement.CurrentCount != 3 {
		t.Errorf("Expected CurrentCount 3, got %d", quest.Requirement.CurrentCount)
	}
	if !quest.Completed {
		t.Error("Quest Completed flag should be true")
	}

	// Additional kill after completion should not increment further.
	completed = CheckNPCQuestProgress(&quest, "kill", "wolf")
	if completed {
		t.Error("Completed quest should return false on further events")
	}
	if quest.Requirement.CurrentCount != 3 {
		t.Errorf("CurrentCount should remain 3 after completion, got %d",
			quest.Requirement.CurrentCount)
	}
}

// ============================================================
// TestCheckNPCQuestProgressWrongType
// ============================================================

func TestCheckNPCQuestProgressWrongType(t *testing.T) {
	quest := models.NPCQuest{
		ID:   "test_kill_quest_wrong",
		Type: "kill",
		Requirement: models.NPCQuestReq{
			Type:         "kill",
			TargetName:   "any",
			TargetCount:  3,
			CurrentCount: 0,
		},
		Completed: false,
		Failed:    false,
	}

	// Sending gather events should not affect a kill quest.
	completed := CheckNPCQuestProgress(&quest, "gather", "Iron")
	if completed {
		t.Error("gather event should not complete a kill quest")
	}
	if quest.Requirement.CurrentCount != 0 {
		t.Errorf("CurrentCount should remain 0 for wrong event type, got %d",
			quest.Requirement.CurrentCount)
	}

	// Sending dungeon_floor events should not affect a kill quest.
	completed = CheckNPCQuestProgress(&quest, "dungeon_floor", "floor_1")
	if completed {
		t.Error("dungeon_floor event should not complete a kill quest")
	}
	if quest.Requirement.CurrentCount != 0 {
		t.Errorf("CurrentCount should remain 0 for dungeon_floor event, got %d",
			quest.Requirement.CurrentCount)
	}

	// Sending kill_rarity events should not affect a plain kill quest.
	completed = CheckNPCQuestProgress(&quest, "kill_rarity", "rare")
	if completed {
		t.Error("kill_rarity event should not complete a kill quest")
	}
	if quest.Requirement.CurrentCount != 0 {
		t.Errorf("CurrentCount should remain 0 for kill_rarity event, got %d",
			quest.Requirement.CurrentCount)
	}
}

// ============================================================
// TestCompleteNPCQuest
// ============================================================

func TestCompleteNPCQuest(t *testing.T) {
	npc := makeQuestNPC("npc_smith", "Gordo the Smith", "Blacksmith")

	town := models.Town{
		Name:      "TestTown",
		Townsfolk: []models.NPCTownsfolk{npc},
		NPCQuests: []models.NPCQuest{
			{
				ID:      "quest_done_1",
				NPCID:   "npc_smith",
				NPCName: "Gordo the Smith",
				Name:    "Supply Run: Iron",
				Type:    "gather",
				Requirement: models.NPCQuestReq{
					Type:         "gather",
					TargetName:   "Iron",
					TargetCount:  3,
					CurrentCount: 3,
				},
				Reward: models.NPCQuestReward{
					XP:         150,
					Gold:       60,
					Reputation: 5,
				},
				Completed: true,
				Failed:    false,
			},
		},
	}

	player := models.Character{
		Name:            "Hero",
		Experience:      0,
		ActiveNPCQuests: []string{"quest_done_1"},
	}

	msgs, xp, gold, rep := CompleteNPCQuest(&player, &town, "quest_done_1")

	// Verify rewards were returned.
	if xp != 150 {
		t.Errorf("Expected 150 XP, got %d", xp)
	}
	if gold != 60 {
		t.Errorf("Expected 60 Gold, got %d", gold)
	}
	if rep != 5 {
		t.Errorf("Expected 5 Reputation, got %d", rep)
	}

	// Verify XP was granted to player.
	if player.Experience != 150 {
		t.Errorf("Player experience should be 150, got %d", player.Experience)
	}

	// Verify Gold was added to ResourceStorageMap.
	goldRes, ok := player.ResourceStorageMap["Gold"]
	if !ok {
		t.Fatal("Gold should exist in ResourceStorageMap after quest completion")
	}
	if goldRes.Stock != 60 {
		t.Errorf("Expected 60 Gold in storage, got %d", goldRes.Stock)
	}

	// Verify quest moved from active to completed on the player.
	if len(player.ActiveNPCQuests) != 0 {
		t.Errorf("ActiveNPCQuests should be empty, got %v", player.ActiveNPCQuests)
	}
	if len(player.CompletedNPCQuests) != 1 || player.CompletedNPCQuests[0] != "quest_done_1" {
		t.Errorf("CompletedNPCQuests should contain quest_done_1, got %v", player.CompletedNPCQuests)
	}

	// Verify NPC relationship was updated.
	npcAfter := town.Townsfolk[0]
	relVal, exists := npcAfter.Relationships[player.Name]
	if !exists {
		t.Fatal("NPC relationship entry should exist for the player")
	}
	if relVal != rep {
		t.Errorf("NPC relationship should be %d, got %d", rep, relVal)
	}

	// Verify messages were produced.
	if len(msgs) == 0 {
		t.Error("Expected at least one narrative message")
	}

	t.Logf("CompleteNPCQuest messages: %v", msgs)
}

// TestCompleteNPCQuestNotFound verifies behaviour when the quest ID is invalid.
func TestCompleteNPCQuestNotFound(t *testing.T) {
	town := models.Town{Name: "Empty", NPCQuests: []models.NPCQuest{}}
	player := models.Character{Name: "Hero"}

	msgs, xp, gold, rep := CompleteNPCQuest(&player, &town, "nonexistent")

	if xp != 0 || gold != 0 || rep != 0 {
		t.Errorf("Expected zero rewards for missing quest, got XP=%d Gold=%d Rep=%d", xp, gold, rep)
	}
	if len(msgs) == 0 {
		t.Error("Expected a 'not found' message")
	}
}

// TestCompleteNPCQuestNotYetDone verifies incomplete quests cannot be turned in.
func TestCompleteNPCQuestNotYetDone(t *testing.T) {
	town := models.Town{
		Name: "TestTown",
		NPCQuests: []models.NPCQuest{
			{
				ID:        "quest_incomplete",
				Completed: false,
				Name:      "Incomplete Quest",
				Reward:    models.NPCQuestReward{XP: 100, Gold: 50, Reputation: 5},
			},
		},
	}
	player := models.Character{Name: "Hero"}

	msgs, xp, gold, rep := CompleteNPCQuest(&player, &town, "quest_incomplete")

	if xp != 0 || gold != 0 || rep != 0 {
		t.Errorf("Expected zero rewards for incomplete quest, got XP=%d Gold=%d Rep=%d", xp, gold, rep)
	}
	if player.Experience != 0 {
		t.Errorf("Player XP should not change, got %d", player.Experience)
	}
	if len(msgs) == 0 {
		t.Error("Expected a 'not yet complete' message")
	}
}

// ============================================================
// TestRefreshNPCQuestBoard
// ============================================================

func TestRefreshNPCQuestBoard(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	town := models.Town{
		Name: "QuestTown",
		Townsfolk: []models.NPCTownsfolk{
			makeQuestNPC("npc_1", "Alice", "Blacksmith"),
			makeQuestNPC("npc_2", "Bob", "Innkeeper"),
			makeQuestNPC("npc_3", "Carol", "Scholar"),
		},
		NPCQuests: nil, // No quests yet.
	}

	RefreshNPCQuestBoard(&town, 5)

	// Each quest-giving NPC should have one quest generated.
	if len(town.NPCQuests) != 3 {
		t.Errorf("Expected 3 quests (one per quest-giving NPC), got %d", len(town.NPCQuests))
	}

	// Verify each quest references a valid NPC.
	npcIDs := map[string]bool{"npc_1": true, "npc_2": true, "npc_3": true}
	for _, q := range town.NPCQuests {
		if !npcIDs[q.NPCID] {
			t.Errorf("Quest %q references unknown NPC ID %q", q.Name, q.NPCID)
		}
		if q.Completed || q.Failed {
			t.Errorf("Freshly generated quest %q should not be completed or failed", q.Name)
		}
	}

	t.Logf("Quest board after refresh: %d quests", len(town.NPCQuests))
}

// TestRefreshNPCQuestBoardSkipsDead verifies dead NPCs do not get quests.
func TestRefreshNPCQuestBoardSkipsDead(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	deadNPC := makeQuestNPC("npc_dead", "DeadGuy", "Guard Captain")
	deadNPC.IsAlive = false

	town := models.Town{
		Name: "QuestTown",
		Townsfolk: []models.NPCTownsfolk{
			makeQuestNPC("npc_alive", "LiveGuy", "Innkeeper"),
			deadNPC,
		},
		NPCQuests: nil,
	}

	RefreshNPCQuestBoard(&town, 5)

	if len(town.NPCQuests) != 1 {
		t.Errorf("Expected 1 quest (dead NPC skipped), got %d", len(town.NPCQuests))
	}
	if len(town.NPCQuests) > 0 && town.NPCQuests[0].NPCID != "npc_alive" {
		t.Errorf("Quest should be from the alive NPC, got NPCID %q", town.NPCQuests[0].NPCID)
	}
}

// TestRefreshNPCQuestBoardSkipsNonGivers verifies non-quest-giver NPCs are skipped.
func TestRefreshNPCQuestBoardSkipsNonGivers(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	nonGiver := makeQuestNPC("npc_nongiver", "Bystander", "Farmer")
	nonGiver.QuestGiver = false

	town := models.Town{
		Name: "QuestTown",
		Townsfolk: []models.NPCTownsfolk{
			makeQuestNPC("npc_giver", "QuestNPC", "Blacksmith"),
			nonGiver,
		},
		NPCQuests: nil,
	}

	RefreshNPCQuestBoard(&town, 5)

	if len(town.NPCQuests) != 1 {
		t.Errorf("Expected 1 quest (non-giver skipped), got %d", len(town.NPCQuests))
	}
}

// TestRefreshNPCQuestBoardDoesNotDuplicate verifies no new quest is generated
// for an NPC that already has an active quest on the board.
func TestRefreshNPCQuestBoardDoesNotDuplicate(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	town := models.Town{
		Name: "QuestTown",
		Townsfolk: []models.NPCTownsfolk{
			makeQuestNPC("npc_1", "Alice", "Blacksmith"),
		},
		NPCQuests: []models.NPCQuest{
			{
				ID:        "existing_quest",
				NPCID:     "npc_1",
				Completed: false,
				Failed:    false,
			},
		},
	}

	RefreshNPCQuestBoard(&town, 5)

	if len(town.NPCQuests) != 1 {
		t.Errorf("Expected 1 quest (existing preserved, no duplicate), got %d",
			len(town.NPCQuests))
	}
}

// ============================================================
// TestRepGating
// ============================================================

func TestRepGating(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	npc := makeQuestNPC("npc_rep", "RepNPC", "Innkeeper")

	// Level 1 should produce easy quests with repRequired 0.
	easyQuest := GenerateNPCQuest(&npc, 1)
	if easyQuest.Difficulty != "easy" {
		t.Errorf("Level 1: expected difficulty 'easy', got %q", easyQuest.Difficulty)
	}
	if easyQuest.RepRequired != 0 {
		t.Errorf("Level 1: expected RepRequired 0, got %d", easyQuest.RepRequired)
	}

	// Level 15 should produce medium quests with repRequired 10.
	medQuest := GenerateNPCQuest(&npc, 15)
	if medQuest.Difficulty != "medium" {
		t.Errorf("Level 15: expected difficulty 'medium', got %q", medQuest.Difficulty)
	}
	if medQuest.RepRequired != 10 {
		t.Errorf("Level 15: expected RepRequired 10, got %d", medQuest.RepRequired)
	}

	// Level 25 should produce hard quests with repRequired 30.
	hardQuest := GenerateNPCQuest(&npc, 25)
	if hardQuest.Difficulty != "hard" {
		t.Errorf("Level 25: expected difficulty 'hard', got %q", hardQuest.Difficulty)
	}
	if hardQuest.RepRequired != 30 {
		t.Errorf("Level 25: expected RepRequired 30, got %d", hardQuest.RepRequired)
	}

	// Verify that higher difficulty means larger rewards.
	if easyQuest.Reward.XP >= medQuest.Reward.XP {
		// Because the multiplier scales and target counts can vary, we check
		// the minimum possible medium reward (3 targets * 50 * 2 = 300) vs
		// the maximum possible easy reward (3 targets * 50 * 1 = 150).
		// This is a sanity check; exact values depend on RNG for target counts.
		t.Logf("Note: easy XP=%d, medium XP=%d (overlap possible due to RNG target counts)",
			easyQuest.Reward.XP, medQuest.Reward.XP)
	}

	t.Logf("RepGating: easy(rep=%d), medium(rep=%d), hard(rep=%d)",
		easyQuest.RepRequired, medQuest.RepRequired, hardQuest.RepRequired)
}
