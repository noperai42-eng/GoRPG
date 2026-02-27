package engine

import (
	"fmt"

	"rpg-game/pkg/game"
	"rpg-game/pkg/models"
)

// ─────────────────────────────────────────────────────────────────────
// NPC Quest Board
// ─────────────────────────────────────────────────────────────────────

func (e *Engine) handleTownNPCQuestBoard(session *GameSession, cmd GameCommand) GameResponse {
	town, err := e.loadOrCreateTown(session)
	if err != nil {
		session.State = StateTownMain
		return e.handleTownMain(session, GameCommand{Type: "init"})
	}
	session.SelectedTown = town
	player := session.Player

	// Refresh quest board
	game.RefreshNPCQuestBoard(town, player.Level)
	e.saveTown(town)

	if cmd.Value == "0" || cmd.Value == "back" {
		session.State = StateTownMain
		return e.handleTownMain(session, GameCommand{Type: "init"})
	}

	// Handle accept/turn-in actions
	if len(cmd.Value) > 7 && cmd.Value[:7] == "accept:" {
		questID := cmd.Value[7:]
		return e.acceptNPCQuest(session, town, questID)
	}
	if len(cmd.Value) > 7 && cmd.Value[:7] == "turnin:" {
		questID := cmd.Value[7:]
		return e.turnInNPCQuest(session, town, questID)
	}

	// Show quest board
	msgs := []GameMessage{
		Msg("============================================================", "system"),
		Msg("  NPC Quest Board", "system"),
		Msg("============================================================", "system"),
	}

	options := []MenuOption{}
	hasQuests := false

	for _, q := range town.NPCQuests {
		if q.Completed || q.Failed {
			continue
		}
		hasQuests = true

		diffTag := ""
		switch q.Difficulty {
		case "medium":
			diffTag = " [Medium]"
		case "hard":
			diffTag = " [Hard]"
		case "legendary":
			diffTag = " [Legendary]"
		}

		label := fmt.Sprintf("%s%s (from %s) - %s %d/%d",
			q.Name, diffTag, q.NPCName,
			q.Requirement.TargetName, q.Requirement.CurrentCount, q.Requirement.TargetCount)
		msgs = append(msgs, Msg(label, "system"))
		msgs = append(msgs, Msg(fmt.Sprintf("  Reward: %d XP, %d Gold, +%d Rep", q.Reward.XP, q.Reward.Gold, q.Reward.Reputation), "loot"))

		if q.AcceptedBy == "" {
			// Check rep requirement
			npcRel := 0
			for _, npc := range town.Townsfolk {
				if npc.ID == q.NPCID {
					if npc.Relationships != nil {
						npcRel = npc.Relationships[player.Name]
					}
					break
				}
			}
			if npcRel >= q.RepRequired {
				options = append(options, Opt("accept:"+q.ID, "Accept: "+q.Name))
			} else {
				options = append(options, OptDisabled("accept:"+q.ID, fmt.Sprintf("Accept: %s [need %d rep, have %d]", q.Name, q.RepRequired, npcRel)))
			}
		} else if q.AcceptedBy == player.Name {
			if q.Requirement.CurrentCount >= q.Requirement.TargetCount {
				options = append(options, Opt("turnin:"+q.ID, "Turn In: "+q.Name))
			} else {
				options = append(options, OptDisabled("turnin:"+q.ID,
					fmt.Sprintf("In Progress: %s (%d/%d)", q.Name, q.Requirement.CurrentCount, q.Requirement.TargetCount)))
			}
		} else {
			msgs = append(msgs, Msg(fmt.Sprintf("  (Accepted by %s)", q.AcceptedBy), "system"))
		}
	}

	if !hasQuests {
		msgs = append(msgs, Msg("No NPC quests available right now.", "narrative"))
	}

	options = append(options, Opt("0", "Back to Town"))

	return GameResponse{
		Type:     "menu",
		Messages: msgs,
		State:    townStateData("town_npc_quest_board", session, town),
		Options:  options,
	}
}

func (e *Engine) acceptNPCQuest(session *GameSession, town *models.Town, questID string) GameResponse {
	player := session.Player

	for i, q := range town.NPCQuests {
		if q.ID == questID && !q.Completed && !q.Failed && q.AcceptedBy == "" {
			town.NPCQuests[i].AcceptedBy = player.Name
			player.ActiveNPCQuests = append(player.ActiveNPCQuests, questID)
			e.saveTown(town)
			session.GameState.CharactersMap[player.Name] = *player

			msgs := []GameMessage{Msg(fmt.Sprintf("Accepted quest: %s", q.Name), "system")}
			session.State = StateTownNPCQuestBoard
			resp := e.handleTownNPCQuestBoard(session, GameCommand{Type: "init"})
			resp.Messages = append(msgs, resp.Messages...)
			return resp
		}
	}

	msgs := []GameMessage{Msg("Quest not available!", "error")}
	session.State = StateTownNPCQuestBoard
	resp := e.handleTownNPCQuestBoard(session, GameCommand{Type: "init"})
	resp.Messages = append(msgs, resp.Messages...)
	return resp
}

func (e *Engine) turnInNPCQuest(session *GameSession, town *models.Town, questID string) GameResponse {
	player := session.Player

	narrativeMsgs, xp, gold, rep := game.CompleteNPCQuest(player, town, questID)
	if xp == 0 && gold == 0 {
		msgs := []GameMessage{Msg("Quest not found or not ready!", "error")}
		session.State = StateTownNPCQuestBoard
		resp := e.handleTownNPCQuestBoard(session, GameCommand{Type: "init"})
		resp.Messages = append(msgs, resp.Messages...)
		return resp
	}

	e.saveTown(town)
	session.GameState.CharactersMap[player.Name] = *player

	// Level up check
	prevLevel := player.Level
	game.LevelUp(player)

	msgs := []GameMessage{}
	for _, m := range narrativeMsgs {
		msgs = append(msgs, Msg(m, "system"))
	}
	msgs = append(msgs, Msg(fmt.Sprintf("Rewards: +%d XP, +%d Gold, +%d Reputation", xp, gold, rep), "loot"))

	if player.Level > prevLevel {
		msgs = append(msgs, Msg(fmt.Sprintf("LEVEL UP! Now level %d!", player.Level), "levelup"))
	}

	session.State = StateTownNPCQuestBoard
	resp := e.handleTownNPCQuestBoard(session, GameCommand{Type: "init"})
	resp.Messages = append(msgs, resp.Messages...)
	return resp
}
