package game

import (
	"fmt"
	"math/rand"
	"time"

	"rpg-game/pkg/models"
)

// GenerateNPCQuest creates a quest tailored to the NPC's title and scaled to
// the player's level.  Blacksmiths ask for materials, Innkeepers want monsters
// killed, Scholars need dungeon floors explored, and so on.
func GenerateNPCQuest(npc *models.NPCTownsfolk, playerLevel int) *models.NPCQuest {
	// Determine difficulty bracket and associated scaling.
	difficulty := "easy"
	var minTargets, maxTargets, multiplier, repRequired int

	switch {
	case playerLevel >= 21:
		difficulty = "hard"
		minTargets = 5
		maxTargets = 10
		multiplier = 4
		repRequired = 30
	case playerLevel >= 11:
		difficulty = "medium"
		minTargets = 3
		maxTargets = 5
		multiplier = 2
		repRequired = 10
	default:
		difficulty = "easy"
		minTargets = 1
		maxTargets = 3
		multiplier = 1
		repRequired = 0
	}

	targetCount := minTargets + rand.Intn(maxTargets-minTargets+1)

	// Choose quest type and target based on NPC title.
	var questType, targetName, questName, questDesc string

	switch npc.Title {
	case "Blacksmith":
		questType = "gather"
		materials := []string{"Iron", "Stone", "Lumber"}
		targetName = materials[rand.Intn(len(materials))]
		questName = fmt.Sprintf("Supply Run: %s", targetName)
		questDesc = fmt.Sprintf("%s needs %d %s for the forge. Can you bring some?", npc.Name, targetCount, targetName)

	case "Innkeeper":
		questType = "kill"
		targetName = "any"
		questName = "Pest Control"
		questDesc = fmt.Sprintf("%s wants the roads kept safe. Slay %d monsters.", npc.Name, targetCount)

	case "Scholar":
		questType = "explore"
		targetName = "dungeon_floor"
		questName = "Academic Expedition"
		questDesc = fmt.Sprintf("%s is researching ancient ruins. Clear %d dungeon floors.", npc.Name, targetCount)

	case "Guard Captain":
		questType = "kill_rarity"
		targetName = "rare"
		questName = "Elite Bounty"
		questDesc = fmt.Sprintf("%s has posted a bounty: eliminate %d rare or stronger monsters.", npc.Name, targetCount)

	case "Merchant":
		questType = "fetch"
		fetchItems := []string{"Iron", "Gold", "Lumber", "Stone", "Sand"}
		targetName = fetchItems[rand.Intn(len(fetchItems))]
		questName = fmt.Sprintf("Trade Goods: %s", targetName)
		questDesc = fmt.Sprintf("%s needs %d %s for a big trade deal.", npc.Name, targetCount, targetName)

	case "Farmer":
		questType = "gather"
		farmerRes := []string{"Lumber", "Stone"}
		targetName = farmerRes[rand.Intn(len(farmerRes))]
		questName = fmt.Sprintf("Farm Supplies: %s", targetName)
		questDesc = fmt.Sprintf("%s needs %d %s to repair the farm.", npc.Name, targetCount, targetName)

	default:
		// Random kill or gather for other NPC titles.
		if rand.Intn(2) == 0 {
			questType = "kill"
			targetName = "any"
			questName = "Odd Job: Monster Slaying"
			questDesc = fmt.Sprintf("%s asks you to defeat %d monsters.", npc.Name, targetCount)
		} else {
			questType = "gather"
			resources := []string{"Lumber", "Iron", "Stone", "Gold", "Sand"}
			targetName = resources[rand.Intn(len(resources))]
			questName = fmt.Sprintf("Odd Job: Gather %s", targetName)
			questDesc = fmt.Sprintf("%s needs %d %s. Can you help out?", npc.Name, targetCount, targetName)
		}
	}

	// Compute rewards.
	xp := targetCount * 50 * multiplier
	gold := targetCount * 20 * multiplier
	reputation := 5 * multiplier

	quest := &models.NPCQuest{
		ID:          fmt.Sprintf("npcq_%s_%d", npc.ID, time.Now().UnixNano()),
		NPCID:       npc.ID,
		NPCName:     npc.Name,
		Name:        questName,
		Description: questDesc,
		Type:        questType,
		Requirement: models.NPCQuestReq{
			Type:         questType,
			TargetName:   targetName,
			TargetCount:  targetCount,
			CurrentCount: 0,
		},
		Reward: models.NPCQuestReward{
			XP:         xp,
			Gold:       gold,
			Reputation: reputation,
		},
		Difficulty:  difficulty,
		RepRequired: repRequired,
		CreatedAt:   time.Now().Unix(),
		Completed:   false,
		Failed:      false,
	}

	return quest
}

// CheckNPCQuestProgress updates the progress of a quest in response to a game
// event.  It returns true when the quest transitions from incomplete to
// complete (i.e. exactly once, on the fulfilling event).
func CheckNPCQuestProgress(quest *models.NPCQuest, eventType string, eventData string) bool {
	if quest.Completed || quest.Failed {
		return false
	}

	matched := false

	switch quest.Requirement.Type {
	case "kill":
		// Any monster kill counts.
		if eventType == "kill" {
			matched = true
		}
	case "kill_rarity":
		// Only rare, epic, or legendary kills count.
		if eventType == "kill_rarity" {
			switch eventData {
			case "rare", "epic", "legendary":
				matched = true
			}
		}
	case "gather", "fetch":
		// Resource name must match the quest target.
		if eventType == "gather" && eventData == quest.Requirement.TargetName {
			matched = true
		}
	case "explore":
		// Any dungeon floor cleared counts.
		if eventType == "dungeon_floor" {
			matched = true
		}
	}

	if !matched {
		return false
	}

	quest.Requirement.CurrentCount++

	if quest.Requirement.CurrentCount >= quest.Requirement.TargetCount {
		quest.Completed = true
		return true
	}

	return false
}

// CompleteNPCQuest finalises a completed quest: awards XP and Gold to the
// player, improves the NPC relationship, and returns narrative messages along
// with the reward amounts.
func CompleteNPCQuest(player *models.Character, town *models.Town, questID string) (msgs []string, xp int, gold int, rep int) {
	// Locate the quest in the town's quest board.
	var quest *models.NPCQuest
	for i := range town.NPCQuests {
		if town.NPCQuests[i].ID == questID {
			quest = &town.NPCQuests[i]
			break
		}
	}

	if quest == nil {
		msgs = append(msgs, "Quest not found.")
		return msgs, 0, 0, 0
	}

	if !quest.Completed {
		msgs = append(msgs, fmt.Sprintf("Quest '%s' is not yet complete.", quest.Name))
		return msgs, 0, 0, 0
	}

	// Determine reward values.
	xp = quest.Reward.XP
	gold = quest.Reward.Gold
	rep = quest.Reward.Reputation

	// Grant XP.
	player.Experience += xp

	// Grant Gold via the resource storage map.
	if player.ResourceStorageMap == nil {
		player.ResourceStorageMap = make(map[string]models.Resource)
	}
	goldRes := player.ResourceStorageMap["Gold"]
	goldRes.Name = "Gold"
	goldRes.Stock += gold
	player.ResourceStorageMap["Gold"] = goldRes

	// Track the quest on the player.
	if player.CompletedNPCQuests == nil {
		player.CompletedNPCQuests = []string{}
	}
	player.CompletedNPCQuests = append(player.CompletedNPCQuests, questID)

	// Remove from active list.
	newActive := make([]string, 0, len(player.ActiveNPCQuests))
	for _, id := range player.ActiveNPCQuests {
		if id != questID {
			newActive = append(newActive, id)
		}
	}
	player.ActiveNPCQuests = newActive

	// Update NPC relationship.
	for i := range town.Townsfolk {
		if town.Townsfolk[i].ID == quest.NPCID {
			UpdateNPCRelationship(&town.Townsfolk[i], player.Name, rep)
			break
		}
	}

	// Build narrative messages.
	msgs = append(msgs, fmt.Sprintf("Quest '%s' completed!", quest.Name))
	msgs = append(msgs, fmt.Sprintf("  Reward: +%d XP, +%d Gold, +%d Reputation with %s", xp, gold, rep, quest.NPCName))

	return msgs, xp, gold, rep
}

// RefreshNPCQuestBoard ensures every quest-giving NPC that does not already
// have an active, unclaimed quest receives a freshly generated one.  The total
// number of quests on the board is capped at 20.
func RefreshNPCQuestBoard(town *models.Town, playerLevel int) {
	if town.NPCQuests == nil {
		town.NPCQuests = []models.NPCQuest{}
	}

	// Build a set of NPC IDs that already have an active (uncompleted,
	// unclaimed-or-claimed) quest on the board.
	activeNPCIDs := make(map[string]bool)
	for _, q := range town.NPCQuests {
		if !q.Completed && !q.Failed {
			activeNPCIDs[q.NPCID] = true
		}
	}

	for i := range town.Townsfolk {
		npc := &town.Townsfolk[i]
		if !npc.QuestGiver {
			continue
		}
		if !npc.IsAlive {
			continue
		}
		if activeNPCIDs[npc.ID] {
			continue
		}
		if len(town.NPCQuests) >= 20 {
			break
		}

		quest := GenerateNPCQuest(npc, playerLevel)
		town.NPCQuests = append(town.NPCQuests, *quest)
	}
}
