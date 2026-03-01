package game

import (
	"fmt"
	"rpg-game/pkg/data"
	"rpg-game/pkg/models"
)

// CheckQuestProgress evaluates all active quests for the player and completes
// any that have met their requirements, granting rewards and activating the
// next quest in the chain. Returns the list of quest IDs that were completed.
func CheckQuestProgress(player *models.Character, gameState *models.GameState) []string {
	if gameState.AvailableQuests == nil {
		gameState.AvailableQuests = make(map[string]models.Quest)
		for k, v := range data.StoryQuests {
			gameState.AvailableQuests[k] = v
		}
	}

	if player.CompletedQuests == nil {
		player.CompletedQuests = []string{}
	}

	if player.ActiveQuests == nil {
		player.ActiveQuests = []string{"quest_1_training"}
	}

	var completed []string

	for _, questID := range player.ActiveQuests {
		quest, exists := gameState.AvailableQuests[questID]
		if !exists {
			continue
		}

		switch quest.Requirement.Type {
		case "level":
			quest.Requirement.CurrentValue = player.Level
		case "location":
			// CurrentValue is incremented by combat victory at the target location;
			// no auto-update needed here.
		case "village_level":
			if gameState.Villages != nil && player.VillageName != "" {
				if village, ok := gameState.Villages[player.VillageName]; ok {
					quest.Requirement.CurrentValue = village.Level
				}
			}
		case "total_resources":
			total := 0
			for _, res := range player.ResourceStorageMap {
				total += res.Stock
			}
			quest.Requirement.CurrentValue = total
		case "skill_count":
			quest.Requirement.CurrentValue = len(player.LearnedSkills)
		case "elder_rescued":
			// CurrentValue is set directly by combat encounter; no auto-update needed.
		}

		if quest.Requirement.CurrentValue >= quest.Requirement.TargetValue {
			fmt.Printf("\n\U0001F3C6 Quest Complete: %s!\n", quest.Name)
			fmt.Printf("   \U00002728 Reward: +%d XP\n", quest.Reward.XP)

			player.Experience += quest.Reward.XP
			quest.Completed = true
			gameState.AvailableQuests[questID] = quest

			player.CompletedQuests = append(player.CompletedQuests, questID)
			completed = append(completed, questID)

			newActive := []string{}
			for _, aid := range player.ActiveQuests {
				if aid != questID {
					newActive = append(newActive, aid)
				}
			}
			player.ActiveQuests = newActive

			ActivateNextQuest(player, gameState, questID)
		} else {
			gameState.AvailableQuests[questID] = quest
		}
	}

	return completed
}

// IncrementLocationQuestProgress increments the CurrentValue for any active
// "location" type quest that matches the given location name. Called on combat
// victory at a location.
func IncrementLocationQuestProgress(player *models.Character, gameState *models.GameState, locationName string) {
	if gameState.AvailableQuests == nil || player.ActiveQuests == nil {
		return
	}
	for _, questID := range player.ActiveQuests {
		quest, exists := gameState.AvailableQuests[questID]
		if !exists {
			continue
		}
		if quest.Requirement.Type == "location" && quest.Requirement.TargetName == locationName {
			quest.Requirement.CurrentValue++
			gameState.AvailableQuests[questID] = quest
		}
	}
}

// ActivateNextQuest looks up the next quest in the story chain based on the
// completed quest ID and activates it for the player.
func ActivateNextQuest(player *models.Character, gameState *models.GameState, completedQuestID string) {
	if gameState.AvailableQuests == nil {
		gameState.AvailableQuests = make(map[string]models.Quest)
		for k, v := range data.StoryQuests {
			gameState.AvailableQuests[k] = v
		}
	}

	if player.ActiveQuests == nil {
		player.ActiveQuests = []string{}
	}

	if player.CompletedQuests == nil {
		player.CompletedQuests = []string{}
	}

	// Main story chain + village/crafting chain.
	// Some quests branch into multiple next quests (e.g. quest_1 → quest_2 + quest_v1).
	nextQuestMap := map[string][]string{
		"quest_1_training": {"quest_2_explore", "quest_v0_elder"},
		"quest_2_explore":  {"quest_3_boss"},
		"quest_3_boss":     {"quest_4_master"},
		"quest_4_master":   {"quest_5_ascension"},
		"quest_v0_elder":   {"quest_v1_village"},
		"quest_v1_village": {"quest_v2_harvest"},
		"quest_v2_harvest": {"quest_v3_potion"},
		"quest_v3_potion":  {"quest_v4_armor", "quest_v6_skills"},
		"quest_v4_armor":   {"quest_v5_weapon"},
		"quest_v5_weapon":  {"quest_v7_scrolls"},
	}

	nextQuestIDs, hasNext := nextQuestMap[completedQuestID]
	if !hasNext {
		return
	}

	for _, nextQuestID := range nextQuestIDs {
		// Skip if already completed or already active
		if Contains(player.CompletedQuests, nextQuestID) || Contains(player.ActiveQuests, nextQuestID) {
			continue
		}

		nextQuest, exists := gameState.AvailableQuests[nextQuestID]
		if !exists || nextQuest.Completed {
			continue
		}

		nextQuest.Active = true
		gameState.AvailableQuests[nextQuestID] = nextQuest

		player.ActiveQuests = append(player.ActiveQuests, nextQuestID)

		fmt.Printf("\n\U0001F4DC New Quest: %s\n", nextQuest.Name)
		fmt.Printf("   %s\n", nextQuest.Description)
	}
}

// BackfillQuests ensures existing characters have all quests they should based
// on their CompletedQuests history. When new quests are added to the chain,
// characters who already completed a prerequisite won't get the new quest
// unless this backfill runs. Returns the number of quests activated.
func BackfillQuests(player *models.Character, gameState *models.GameState) int {
	if gameState.AvailableQuests == nil {
		gameState.AvailableQuests = make(map[string]models.Quest)
		for k, v := range data.StoryQuests {
			gameState.AvailableQuests[k] = v
		}
	}

	// Ensure new quests from code are present in the game state.
	for id, quest := range data.StoryQuests {
		if _, exists := gameState.AvailableQuests[id]; !exists {
			gameState.AvailableQuests[id] = quest
		}
	}

	if player.CompletedQuests == nil {
		player.CompletedQuests = []string{}
	}
	if player.ActiveQuests == nil {
		player.ActiveQuests = []string{"quest_1_training"}
	}

	// Deduplicate CompletedQuests (fixes legacy data from pre-guard ActivateNextQuest)
	seen := make(map[string]bool)
	deduped := player.CompletedQuests[:0]
	for _, q := range player.CompletedQuests {
		if !seen[q] {
			seen[q] = true
			deduped = append(deduped, q)
		}
	}
	player.CompletedQuests = deduped

	// Deduplicate ActiveQuests
	seen = make(map[string]bool)
	dedupedActive := player.ActiveQuests[:0]
	for _, q := range player.ActiveQuests {
		if !seen[q] {
			seen[q] = true
			dedupedActive = append(dedupedActive, q)
		}
	}
	player.ActiveQuests = dedupedActive

	nextQuestMap := map[string][]string{
		"quest_1_training": {"quest_2_explore", "quest_v0_elder"},
		"quest_2_explore":  {"quest_3_boss"},
		"quest_3_boss":     {"quest_4_master"},
		"quest_4_master":   {"quest_5_ascension"},
		"quest_v0_elder":   {"quest_v1_village"},
		"quest_v1_village": {"quest_v2_harvest"},
		"quest_v2_harvest": {"quest_v3_potion"},
		"quest_v3_potion":  {"quest_v4_armor", "quest_v6_skills"},
		"quest_v4_armor":   {"quest_v5_weapon"},
		"quest_v5_weapon":  {"quest_v7_scrolls"},
	}

	activated := 0
	for _, completedID := range player.CompletedQuests {
		nextIDs, ok := nextQuestMap[completedID]
		if !ok {
			continue
		}
		for _, nextID := range nextIDs {
			// Skip if already completed or already active
			if Contains(player.CompletedQuests, nextID) || Contains(player.ActiveQuests, nextID) {
				continue
			}
			quest, exists := gameState.AvailableQuests[nextID]
			if !exists {
				continue
			}
			if quest.Completed {
				continue
			}
			quest.Active = true
			gameState.AvailableQuests[nextID] = quest
			player.ActiveQuests = append(player.ActiveQuests, nextID)
			activated++
		}
	}

	return activated
}

// ShowQuestLog displays the player's active and completed quests with
// progress indicators and completion status.
func ShowQuestLog(player *models.Character, gameState *models.GameState) {
	if gameState.AvailableQuests == nil {
		gameState.AvailableQuests = make(map[string]models.Quest)
		for k, v := range data.StoryQuests {
			gameState.AvailableQuests[k] = v
		}
	}

	if player.CompletedQuests == nil {
		player.CompletedQuests = []string{}
	}

	if player.ActiveQuests == nil {
		player.ActiveQuests = []string{"quest_1_training"}
	}

	fmt.Println("\n\U0001F4D6 ====== QUEST LOG ======")

	fmt.Println("\n\U0001F525 Active Quests:")
	if len(player.ActiveQuests) == 0 {
		fmt.Println("   No active quests.")
	}
	for _, questID := range player.ActiveQuests {
		quest, exists := gameState.AvailableQuests[questID]
		if !exists {
			continue
		}

		fmt.Printf("\n   \U000027A1 %s\n", quest.Name)
		fmt.Printf("     %s\n", quest.Description)

		switch quest.Requirement.Type {
		case "level":
			fmt.Printf("     \U0001F4CA Progress: Level %d / %d\n", quest.Requirement.CurrentValue, quest.Requirement.TargetValue)
		case "boss_kill":
			fmt.Printf("     \U0001F480 Progress: %d / %d %s defeated\n", quest.Requirement.CurrentValue, quest.Requirement.TargetValue, quest.Requirement.TargetName)
		case "location":
			fmt.Printf("     \U0001F5FA Progress: %d / %d locations explored in %s\n", quest.Requirement.CurrentValue, quest.Requirement.TargetValue, quest.Requirement.TargetName)
		}

		fmt.Printf("     \U0001F381 Reward: %d XP\n", quest.Reward.XP)
	}

	fmt.Println("\n\U00002705 Completed Quests:")
	if len(player.CompletedQuests) == 0 {
		fmt.Println("   No completed quests yet.")
	}
	for _, questID := range player.CompletedQuests {
		quest, exists := gameState.AvailableQuests[questID]
		if !exists {
			continue
		}
		fmt.Printf("   \U0001F3C6 %s - COMPLETE\n", quest.Name)
	}

	fmt.Println("\n========================")
}
