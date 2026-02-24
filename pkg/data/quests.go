package data

import "rpg-game/pkg/models"

var StoryQuests = map[string]models.Quest{
	"quest_1_training": {
		ID:          "quest_1_training",
		Name:        "The First Trial",
		Description: "The village elder asks you to complete your training by reaching level 3.",
		Type:        "level",
		Requirement: models.QuestRequirement{Type: "level", TargetValue: 3, TargetName: "", CurrentValue: 0},
		Reward:      models.QuestReward{Type: "unlock_location", Value: "Forest Ruins", XP: 100},
		Completed:   false,
		Active:      true,
	},
	"quest_2_explore": {
		ID:          "quest_2_explore",
		Name:        "Into the Ruins",
		Description: "A mysterious force emanates from the Forest Ruins. Explore it to unlock new areas.",
		Type:        "explore",
		Requirement: models.QuestRequirement{Type: "location", TargetValue: 5, TargetName: "Forest Ruins", CurrentValue: 0},
		Reward:      models.QuestReward{Type: "unlock_location", Value: "Ancient Dungeon", XP: 250},
		Completed:   false,
		Active:      false,
	},
	"quest_3_boss": {
		ID:          "quest_3_boss",
		Name:        "The Dungeon Guardian",
		Description: "Defeat the Guardian Boss in the Ancient Dungeon to prove your worth.",
		Type:        "boss",
		Requirement: models.QuestRequirement{Type: "boss_kill", TargetValue: 1, TargetName: "Guardian", CurrentValue: 0},
		Reward:      models.QuestReward{Type: "unlock_feature", Value: "advanced_skills", XP: 500},
		Completed:   false,
		Active:      false,
	},
	"quest_4_master": {
		ID:          "quest_4_master",
		Name:        "The Master's Challenge",
		Description: "Reach level 10 and defeat the Master in combat to unlock The Tower.",
		Type:        "boss",
		Requirement: models.QuestRequirement{Type: "boss_kill", TargetValue: 1, TargetName: "The Master", CurrentValue: 0},
		Reward:      models.QuestReward{Type: "unlock_location", Value: "The Tower", XP: 1000},
		Completed:   false,
		Active:      false,
	},
	"quest_5_ascension": {
		ID:          "quest_5_ascension",
		Name:        "Tower Ascension",
		Description: "Climb The Tower and defeat the final boss to ascend to a new realm.",
		Type:        "boss",
		Requirement: models.QuestRequirement{Type: "boss_kill", TargetValue: 1, TargetName: "Tower Lord", CurrentValue: 0},
		Reward:      models.QuestReward{Type: "unlock_feature", Value: "prestige_mode", XP: 5000},
		Completed:   false,
		Active:      false,
	},
}
