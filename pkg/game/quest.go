package game

import (
	"fmt"
	"rpg-game/pkg/data"
	"rpg-game/pkg/models"
)

// CheckQuestProgress evaluates all active quests for the player and completes
// any that have met their requirements, granting rewards and activating the
// next quest in the chain.
func CheckQuestProgress(player *models.Character, gameState *models.GameState) {
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

	for _, questID := range player.ActiveQuests {
		quest, exists := gameState.AvailableQuests[questID]
		if !exists {
			continue
		}

		switch quest.Requirement.Type {
		case "level":
			quest.Requirement.CurrentValue = player.Level
		}

		if quest.Requirement.CurrentValue >= quest.Requirement.TargetValue {
			fmt.Printf("\n\U0001F3C6 Quest Complete: %s!\n", quest.Name)
			fmt.Printf("   \U00002728 Reward: +%d XP\n", quest.Reward.XP)

			player.Experience += quest.Reward.XP
			quest.Completed = true
			gameState.AvailableQuests[questID] = quest

			player.CompletedQuests = append(player.CompletedQuests, questID)

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

	nextQuestMap := map[string]string{
		"quest_1_training": "quest_2_explore",
		"quest_2_explore":  "quest_3_boss",
		"quest_3_boss":     "quest_4_master",
		"quest_4_master":   "quest_5_ascension",
	}

	nextQuestID, hasNext := nextQuestMap[completedQuestID]
	if !hasNext {
		return
	}

	nextQuest, exists := gameState.AvailableQuests[nextQuestID]
	if !exists {
		return
	}

	nextQuest.Active = true
	gameState.AvailableQuests[nextQuestID] = nextQuest

	player.ActiveQuests = append(player.ActiveQuests, nextQuestID)

	fmt.Printf("\n\U0001F4DC New Quest: %s\n", nextQuest.Name)
	fmt.Printf("   %s\n", nextQuest.Description)
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
