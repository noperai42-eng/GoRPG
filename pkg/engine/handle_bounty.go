package engine

import (
	"fmt"
	"strconv"

	"rpg-game/pkg/game"
)

// handleMostWantedBoard displays the Most Wanted board and lets the player select a bounty target.
func (e *Engine) handleMostWantedBoard(session *GameSession, cmd GameCommand) GameResponse {
	player := session.Player
	gs := session.GameState

	if cmd.Type == "init" || cmd.Value == "back" || cmd.Value == "" {
		// Display the Most Wanted board
		entries := game.GetMostWanted(gs.GameLocations, 10)

		if len(entries) == 0 {
			session.State = StateMainMenu
			resp := BuildMainMenuResponse(session)
			resp.Messages = append([]GameMessage{
				Msg("The Bounty Board is empty. Monsters haven't earned any notoriety yet.", "narrative"),
				Msg("Come back after monsters have fought and killed!", "system"),
			}, resp.Messages...)
			return resp
		}

		msgs := []GameMessage{
			Msg("========================================", "system"),
			Msg("BOUNTY BOARD - MOST WANTED MONSTERS", "system"),
			Msg("========================================", "system"),
			Msg("Hire a guide to hunt a specific dangerous monster.", "narrative"),
			Msg("", "system"),
		}

		options := []MenuOption{}
		for i, entry := range entries {
			totalKills := entry.PlayerKills + entry.MonsterKills
			rarityTag := ""
			rarityName := game.RarityDisplayName(entry.Rarity)
			if rarityName != "Common" {
				rarityTag = fmt.Sprintf("[%s] ", rarityName)
			}
			bossTag := ""
			if entry.IsBoss {
				bossTag = " [BOSS]"
			}
			label := fmt.Sprintf("#%d %s%s (Lv%d) - %d kills (%d players) - %s%s",
				i+1, rarityTag, entry.Name, entry.Level,
				totalKills, entry.PlayerKills, entry.LocationName, bossTag)
			options = append(options, Opt(strconv.Itoa(i), label))
		}
		options = append(options, Opt("back_menu", "Return to Hub"))

		// Store entries in a temporary way - we use the index to find the entry
		session.State = StateMostWantedBoard

		return GameResponse{
			Type:     "menu",
			Messages: msgs,
			State:    &StateData{Screen: "most_wanted_board", Player: MakePlayerState(player)},
			Options:  options,
		}
	}

	if cmd.Value == "back_menu" {
		session.State = StateMainMenu
		return BuildMainMenuResponse(session)
	}

	// Player selected a bounty target
	idx, err := strconv.Atoi(cmd.Value)
	if err != nil {
		session.State = StateMainMenu
		return BuildMainMenuResponse(session)
	}

	entries := game.GetMostWanted(gs.GameLocations, 10)
	if idx < 0 || idx >= len(entries) {
		session.State = StateMainMenu
		return BuildMainMenuResponse(session)
	}

	entry := entries[idx]

	// Verify player knows the location
	known := false
	for _, kl := range player.KnownLocations {
		if kl == entry.LocationName {
			known = true
			break
		}
	}

	if !known {
		msgs := []GameMessage{
			Msg(fmt.Sprintf("You haven't discovered %s yet! You can't hire a guide there.", entry.LocationName), "error"),
		}
		session.State = StateMostWantedBoard
		return e.handleMostWantedBoard(session, GameCommand{Type: "init"}).WithPrependedMessages(msgs)
	}

	// Calculate gold cost
	goldCost := entry.Level * 5
	rarityIdx := game.RarityIndex(entry.Rarity)
	multipliers := []int{1, 2, 5, 10, 25, 50} // common, uncommon, rare, epic, legendary, mythic
	if rarityIdx < len(multipliers) {
		goldCost *= multipliers[rarityIdx]
	}
	if goldCost < 10 {
		goldCost = 10
	}

	// Check player gold
	goldRes := player.ResourceStorageMap["Gold"]
	playerGold := goldRes.Stock

	msgs := []GameMessage{
		Msg("========================================", "system"),
		Msg("HIRE A GUIDE", "system"),
		Msg("========================================", "system"),
		Msg("", "system"),
	}

	rarityName := game.RarityDisplayName(entry.Rarity)
	if rarityName != "Common" {
		msgs = append(msgs, Msg(fmt.Sprintf("Target: [%s] %s (Level %d)", rarityName, entry.Name, entry.Level), "combat"))
	} else {
		msgs = append(msgs, Msg(fmt.Sprintf("Target: %s (Level %d)", entry.Name, entry.Level), "combat"))
	}
	msgs = append(msgs, Msg(fmt.Sprintf("Location: %s", entry.LocationName), "system"))
	msgs = append(msgs, Msg(fmt.Sprintf("Player Kills: %d | Monster Kills: %d", entry.PlayerKills, entry.MonsterKills), "system"))
	msgs = append(msgs, Msg(fmt.Sprintf("HP: %d", entry.HP), "system"))
	msgs = append(msgs, Msg("", "system"))
	msgs = append(msgs, Msg(fmt.Sprintf("Guide Cost: %d Gold (You have: %d)", goldCost, playerGold), "system"))

	if playerGold < goldCost {
		msgs = append(msgs, Msg("You don't have enough gold!", "error"))
	}

	// Store bounty selection
	session.SelectedBountyLocName = entry.LocationName
	session.SelectedBountyMobIdx = entry.LocationIdx
	session.State = StateMostWantedHunt

	options := []MenuOption{}
	if playerGold >= goldCost {
		options = append(options, Opt("confirm", fmt.Sprintf("Pay %d Gold and Hunt", goldCost)))
	}
	options = append(options, Opt("back", "Back to Bounty Board"))

	return GameResponse{
		Type:     "menu",
		Messages: msgs,
		State:    &StateData{Screen: "most_wanted_hunt", Player: MakePlayerState(player)},
		Options:  options,
	}
}

// handleMostWantedHunt processes the bounty hunt confirmation.
func (e *Engine) handleMostWantedHunt(session *GameSession, cmd GameCommand) GameResponse {
	player := session.Player
	gs := session.GameState

	if cmd.Value == "back" {
		session.State = StateMostWantedBoard
		return e.handleMostWantedBoard(session, GameCommand{Type: "init"})
	}

	if cmd.Value != "confirm" {
		session.State = StateMainMenu
		return BuildMainMenuResponse(session)
	}

	locName := session.SelectedBountyLocName
	mobIdx := session.SelectedBountyMobIdx

	loc, exists := gs.GameLocations[locName]
	if !exists || mobIdx < 0 || mobIdx >= len(loc.Monsters) {
		session.State = StateMainMenu
		resp := BuildMainMenuResponse(session)
		resp.Messages = append([]GameMessage{
			Msg("The bounty target could not be found! It may have been replaced.", "error"),
		}, resp.Messages...)
		return resp
	}

	mob := loc.Monsters[mobIdx]

	// Calculate and deduct gold cost
	goldCost := mob.Level * 5
	rarityIdx := game.RarityIndex(mob.Rarity)
	multipliers := []int{1, 2, 5, 10, 25, 50}
	if rarityIdx < len(multipliers) {
		goldCost *= multipliers[rarityIdx]
	}
	if goldCost < 10 {
		goldCost = 10
	}

	goldRes := player.ResourceStorageMap["Gold"]
	if goldRes.Stock < goldCost {
		session.State = StateMostWantedBoard
		return e.handleMostWantedBoard(session, GameCommand{Type: "init"})
	}

	goldRes.Stock -= goldCost
	player.ResourceStorageMap["Gold"] = goldRes

	return e.startCombat(session, &loc, mobIdx, mob)
}

// WithPrependedMessages is a helper to prepend messages to a GameResponse.
func (r GameResponse) WithPrependedMessages(msgs []GameMessage) GameResponse {
	r.Messages = append(msgs, r.Messages...)
	return r
}
