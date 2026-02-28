package engine

import (
	"fmt"
	"strconv"

	"rpg-game/pkg/db"
	"rpg-game/pkg/game"
)

// handleArenaMain shows the arena main screen with player rating, champion, and options.
func (e *Engine) handleArenaMain(session *GameSession, cmd GameCommand) GameResponse {
	player := session.Player
	msgs := []GameMessage{}

	// Handle back to main menu
	if cmd.Value == "back" {
		session.State = StateMainMenu
		return BuildMainMenuResponse(session)
	}

	// Handle challenge
	if cmd.Value == "1" {
		session.State = StateArenaChallenge
		return e.handleArenaChallenge(session, GameCommand{Type: "init"})
	}

	if e.store == nil {
		msgs = append(msgs, Msg("Arena requires a database connection.", "error"))
		session.State = StateMainMenu
		return GameResponse{
			Type:     "narrative",
			Messages: msgs,
			State:    &StateData{Screen: "main_menu", Player: MakePlayerState(player)},
			Options:  BuildMainMenuResponse(session).Options,
		}
	}

	// Check/apply daily reset
	today := game.GetArenaResetDate()
	e.store.ResetArenaBattles(today)

	// Get or create player's arena entry
	entry, err := e.store.GetArenaEntry(session.AccountID, player.Name)
	if err != nil {
		msgs = append(msgs, Msg("Failed to load arena data.", "error"))
		session.State = StateMainMenu
		return GameResponse{
			Type:     "narrative",
			Messages: msgs,
			State:    &StateData{Screen: "main_menu", Player: MakePlayerState(player)},
			Options:  BuildMainMenuResponse(session).Options,
		}
	}
	if entry == nil {
		// First time — register with 1000 rating
		entry = &db.ArenaEntry{
			AccountID:     session.AccountID,
			CharacterName: player.Name,
			Rating:        1000,
			Wins:          0,
			Losses:        0,
			BattlesToday:  0,
			LastReset:     today,
		}
		e.store.UpsertArenaEntry(*entry)
		// Persist character data so opponents can load it for fights
		e.saveSession(session)
	}

	// Sync stats
	player.Stats.ArenaRating = entry.Rating
	player.Stats.ArenaWins = entry.Wins
	player.Stats.ArenaLosses = entry.Losses
	player.Stats.ArenaBattlesToday = entry.BattlesToday
	player.Stats.ArenaLastReset = 0

	// Get champion
	champion, _ := e.store.GetArenaChampion()

	battlesRemaining := game.ArenaMaxBattlesPerDay - entry.BattlesToday
	if battlesRemaining < 0 {
		battlesRemaining = 0
	}

	msgs = append(msgs, Msg("=== ARENA ===", "system"))
	msgs = append(msgs, Msg(fmt.Sprintf("Your Rating: %d  |  W: %d  L: %d", entry.Rating, entry.Wins, entry.Losses), "system"))
	msgs = append(msgs, Msg(fmt.Sprintf("Battles Today: %d/%d remaining", battlesRemaining, game.ArenaMaxBattlesPerDay), "system"))
	if champion != nil {
		msgs = append(msgs, Msg(fmt.Sprintf("Arena Champion: %s (Rating: %d)", champion.CharacterName, champion.Rating), "narrative"))
	}

	session.State = StateArenaMain
	options := []MenuOption{
		Opt("1", "Challenge a Player"),
		Opt("2", "Arena Leaderboard"),
		Opt("back", "Back"),
	}

	return GameResponse{
		Type:     "menu",
		Messages: msgs,
		State:    &StateData{Screen: "arena_main", Player: MakePlayerState(player)},
		Options:  options,
	}
}

// handleArenaChallenge shows the list of arena-registered players to challenge.
func (e *Engine) handleArenaChallenge(session *GameSession, cmd GameCommand) GameResponse {
	player := session.Player
	msgs := []GameMessage{}

	if cmd.Value == "back" {
		session.State = StateArenaMain
		return e.handleArenaMain(session, GameCommand{Type: "init"})
	}

	// On init, show list
	if cmd.Type == "init" || cmd.Value == "" {
		// Check battles remaining
		entry, _ := e.store.GetArenaEntry(session.AccountID, player.Name)
		if entry != nil {
			today := game.GetArenaResetDate()
			if entry.LastReset != today {
				entry.BattlesToday = 0
				entry.LastReset = today
				e.store.UpsertArenaEntry(*entry)
			}
			remaining := game.ArenaMaxBattlesPerDay - entry.BattlesToday
			if remaining <= 0 {
				msgs = append(msgs, Msg("No arena battles remaining today! Come back tomorrow.", "system"))
				session.State = StateArenaMain
				return GameResponse{
					Type:     "menu",
					Messages: msgs,
					State:    &StateData{Screen: "arena_main", Player: MakePlayerState(player)},
					Options:  []MenuOption{Opt("back", "Back")},
				}
			}
		}

		// Load leaderboard as challenger list
		entries, _ := e.store.GetArenaLeaderboard(50)
		if len(entries) == 0 {
			msgs = append(msgs, Msg("No arena opponents registered yet.", "system"))
			session.State = StateArenaMain
			return GameResponse{
				Type:     "menu",
				Messages: msgs,
				State:    &StateData{Screen: "arena_main", Player: MakePlayerState(player)},
				Options:  []MenuOption{Opt("back", "Back")},
			}
		}

		msgs = append(msgs, Msg("=== Choose Opponent ===", "system"))
		options := []MenuOption{}
		idx := 0
		for _, e := range entries {
			if e.AccountID == session.AccountID && e.CharacterName == player.Name {
				continue // Skip self
			}
			idx++
			label := fmt.Sprintf("%s (Rating: %d, W:%d L:%d)", e.CharacterName, e.Rating, e.Wins, e.Losses)
			options = append(options, Opt(strconv.Itoa(idx), label))
		}

		if len(options) == 0 {
			msgs = append(msgs, Msg("No other players registered in the arena.", "system"))
			session.State = StateArenaMain
			return GameResponse{
				Type:     "menu",
				Messages: msgs,
				State:    &StateData{Screen: "arena_main", Player: MakePlayerState(player)},
				Options:  []MenuOption{Opt("back", "Back")},
			}
		}

		options = append(options, Opt("back", "Back"))
		session.State = StateArenaChallenge
		return GameResponse{
			Type:     "menu",
			Messages: msgs,
			State:    &StateData{Screen: "arena_challenge", Player: MakePlayerState(player)},
			Options:  options,
		}
	}

	// Selection: find the target
	selIdx, err := strconv.Atoi(cmd.Value)
	if err != nil || selIdx < 1 {
		session.State = StateArenaMain
		return e.handleArenaMain(session, GameCommand{Type: "init"})
	}

	entries, _ := e.store.GetArenaLeaderboard(50)
	// Filter out self
	filtered := []db.ArenaEntry{}
	for _, e := range entries {
		if e.AccountID == session.AccountID && e.CharacterName == player.Name {
			continue
		}
		filtered = append(filtered, e)
	}

	if selIdx > len(filtered) {
		msgs = append(msgs, Msg("Invalid selection.", "error"))
		session.State = StateArenaChallenge
		return e.handleArenaChallenge(session, GameCommand{Type: "init"})
	}

	target := filtered[selIdx-1]
	session.ArenaTargetAccountID = target.AccountID
	session.ArenaTargetCharName = target.CharacterName

	// Get own rating
	myEntry, _ := e.store.GetArenaEntry(session.AccountID, player.Name)
	myRating := 1000
	if myEntry != nil {
		myRating = myEntry.Rating
	}

	msgs = append(msgs, Msg(fmt.Sprintf("Challenge %s (Rating: %d)?", target.CharacterName, target.Rating), "system"))
	msgs = append(msgs, Msg(fmt.Sprintf("Your Rating: %d", myRating), "system"))

	session.State = StateArenaConfirm
	return GameResponse{
		Type:     "menu",
		Messages: msgs,
		State:    &StateData{Screen: "arena_confirm", Player: MakePlayerState(player)},
		Options: []MenuOption{
			Opt("y", "Fight!"),
			Opt("n", "Cancel"),
		},
	}
}

// resolveArenaWin handles arena victory: rating changes, no loot/XP.
func (e *Engine) resolveArenaWin(session *GameSession, msgs []GameMessage) GameResponse {
	combat := session.Combat
	player := session.Player

	msgs = append(msgs, Msg("========================================", "system"))
	msgs = append(msgs, Msg(fmt.Sprintf("ARENA VICTORY! %s Wins!", player.Name), "combat"))
	msgs = append(msgs, Msg("========================================", "system"))

	if e.store != nil {
		winnerEntry, _ := e.store.GetArenaEntry(session.AccountID, player.Name)
		loserEntry, _ := e.store.GetArenaEntry(combat.ArenaTargetAccountID, combat.ArenaTargetCharName)

		if winnerEntry != nil && loserEntry != nil {
			gain, loss := game.CalculateArenaPoints(winnerEntry.Rating, loserEntry.Rating)

			if e.metrics != nil {
				e.metrics.RecordArenaFight(winnerEntry.Rating, loserEntry.Rating)
			}

			winnerEntry.Rating += gain
			winnerEntry.Wins++
			winnerEntry.BattlesToday++
			winnerEntry.LastReset = game.GetArenaResetDate()

			loserEntry.Rating -= loss
			if loserEntry.Rating < 0 {
				loserEntry.Rating = 0
			}
			loserEntry.Losses++

			e.store.UpsertArenaEntry(*winnerEntry)
			e.store.UpsertArenaEntry(*loserEntry)

			player.Stats.ArenaRating = winnerEntry.Rating
			player.Stats.ArenaWins = winnerEntry.Wins
			player.Stats.ArenaBattlesToday = winnerEntry.BattlesToday

			msgs = append(msgs, Msg(fmt.Sprintf("Rating: %d (+%d)", winnerEntry.Rating, gain), "levelup"))
			msgs = append(msgs, Msg(fmt.Sprintf("Opponent loses %d rating", loss), "system"))
		}
	}

	// Restore player to full (arena fights don't kill)
	player.HitpointsRemaining = player.HitpointsTotal
	player.ManaRemaining = player.ManaTotal
	player.StaminaRemaining = player.StaminaTotal
	player.StatusEffects = nil

	e.saveSession(session)

	session.State = StateArenaMain
	return GameResponse{
		Type:     "narrative",
		Messages: msgs,
		State:    &StateData{Screen: "arena_main", Player: MakePlayerState(player)},
		Options:  []MenuOption{Opt("init", "Return to Arena")},
	}
}

// resolveArenaLoss handles arena defeat: rating changes, no death penalty.
func (e *Engine) resolveArenaLoss(session *GameSession, msgs []GameMessage) GameResponse {
	combat := session.Combat
	player := session.Player

	msgs = append(msgs, Msg("========================================", "system"))
	msgs = append(msgs, Msg(fmt.Sprintf("ARENA DEFEAT! %s lost the match.", player.Name), "combat"))
	msgs = append(msgs, Msg("========================================", "system"))

	if e.store != nil {
		loserEntry, _ := e.store.GetArenaEntry(session.AccountID, player.Name)
		winnerEntry, _ := e.store.GetArenaEntry(combat.ArenaTargetAccountID, combat.ArenaTargetCharName)

		if loserEntry != nil && winnerEntry != nil {
			gain, loss := game.CalculateArenaPoints(winnerEntry.Rating, loserEntry.Rating)

			winnerEntry.Rating += gain
			winnerEntry.Wins++

			loserEntry.Rating -= loss
			if loserEntry.Rating < 0 {
				loserEntry.Rating = 0
			}
			loserEntry.Losses++
			loserEntry.BattlesToday++
			loserEntry.LastReset = game.GetArenaResetDate()

			e.store.UpsertArenaEntry(*winnerEntry)
			e.store.UpsertArenaEntry(*loserEntry)

			player.Stats.ArenaRating = loserEntry.Rating
			player.Stats.ArenaLosses = loserEntry.Losses
			player.Stats.ArenaBattlesToday = loserEntry.BattlesToday

			msgs = append(msgs, Msg(fmt.Sprintf("Rating: %d (-%d)", loserEntry.Rating, loss), "damage"))
			msgs = append(msgs, Msg(fmt.Sprintf("Opponent gains %d rating", gain), "system"))
		}
	}

	// Restore player to full (arena fights don't kill)
	player.HitpointsRemaining = player.HitpointsTotal
	player.ManaRemaining = player.ManaTotal
	player.StaminaRemaining = player.StaminaTotal
	player.StatusEffects = nil

	e.saveSession(session)

	session.State = StateArenaMain
	return GameResponse{
		Type:     "narrative",
		Messages: msgs,
		State:    &StateData{Screen: "arena_main", Player: MakePlayerState(player)},
		Options:  []MenuOption{Opt("init", "Return to Arena")},
	}
}

// handleArenaDirectChallenge handles a direct arena challenge from the leaderboard click.
func (e *Engine) handleArenaDirectChallenge(session *GameSession, accountIDStr, charName string) GameResponse {
	player := session.Player
	msgs := []GameMessage{}

	targetAccountID, err := strconv.ParseInt(accountIDStr, 10, 64)
	if err != nil {
		msgs = append(msgs, Msg("Invalid challenge target.", "error"))
		session.State = StateMainMenu
		return BuildMainMenuResponse(session)
	}

	if e.store == nil {
		msgs = append(msgs, Msg("Arena requires a database connection.", "error"))
		session.State = StateMainMenu
		return BuildMainMenuResponse(session)
	}

	// Check/apply daily reset
	today := game.GetArenaResetDate()
	e.store.ResetArenaBattles(today)

	// Get or create player's arena entry
	entry, err := e.store.GetArenaEntry(session.AccountID, player.Name)
	if err != nil {
		msgs = append(msgs, Msg("Failed to load arena data.", "error"))
		session.State = StateMainMenu
		return BuildMainMenuResponse(session)
	}
	if entry == nil {
		entry = &db.ArenaEntry{
			AccountID:     session.AccountID,
			CharacterName: player.Name,
			Rating:        1000,
			Wins:          0,
			Losses:        0,
			BattlesToday:  0,
			LastReset:     today,
		}
		e.store.UpsertArenaEntry(*entry)
		e.saveSession(session)
	}

	// Check battles remaining
	if entry.LastReset != today {
		entry.BattlesToday = 0
		entry.LastReset = today
		e.store.UpsertArenaEntry(*entry)
	}
	remaining := game.ArenaMaxBattlesPerDay - entry.BattlesToday
	if remaining <= 0 {
		msgs = append(msgs, Msg("No arena battles remaining today! Come back tomorrow.", "system"))
		session.State = StateArenaMain
		return GameResponse{
			Type:     "menu",
			Messages: msgs,
			State:    &StateData{Screen: "arena_main", Player: MakePlayerState(player)},
			Options:  []MenuOption{Opt("back", "Back")},
		}
	}

	// Can't fight self
	if targetAccountID == session.AccountID && charName == player.Name {
		msgs = append(msgs, Msg("You can't challenge yourself!", "error"))
		session.State = StateArenaMain
		return e.handleArenaMain(session, GameCommand{Type: "init"})
	}

	// Validate target exists in arena
	targetEntry, err := e.store.GetArenaEntry(targetAccountID, charName)
	if err != nil || targetEntry == nil {
		msgs = append(msgs, Msg("Opponent not found in the arena.", "error"))
		session.State = StateArenaMain
		return e.handleArenaMain(session, GameCommand{Type: "init"})
	}

	session.ArenaTargetAccountID = targetAccountID
	session.ArenaTargetCharName = charName

	if e.metrics != nil {
		e.metrics.RecordFeatureUse("arena")
	}

	msgs = append(msgs, Msg(fmt.Sprintf("Challenge %s (Rating: %d)?", targetEntry.CharacterName, targetEntry.Rating), "system"))
	msgs = append(msgs, Msg(fmt.Sprintf("Your Rating: %d", entry.Rating), "system"))

	session.State = StateArenaConfirm
	return GameResponse{
		Type:     "menu",
		Messages: msgs,
		State:    &StateData{Screen: "arena_confirm", Player: MakePlayerState(player)},
		Options: []MenuOption{
			Opt("y", "Fight!"),
			Opt("n", "Cancel"),
		},
	}
}

// handleArenaConfirm processes the player's confirmation or cancellation of an arena challenge.
func (e *Engine) handleArenaConfirm(session *GameSession, cmd GameCommand) GameResponse {
	player := session.Player
	msgs := []GameMessage{}

	if cmd.Value != "y" {
		session.State = StateArenaMain
		return e.handleArenaMain(session, GameCommand{Type: "init"})
	}

	// Load target character snapshot
	targetChar, err := e.store.LoadCharacter(session.ArenaTargetAccountID, session.ArenaTargetCharName)
	if err != nil {
		msgs = append(msgs, Msg("Failed to load opponent data.", "error"))
		session.State = StateArenaMain
		return GameResponse{
			Type:     "narrative",
			Messages: msgs,
			State:    &StateData{Screen: "arena_main", Player: MakePlayerState(player)},
			Options:  []MenuOption{Opt("back", "Back")},
		}
	}

	// Convert to monster
	mob := game.CharacterToArenaMonster(&targetChar)

	// Restore player resources for fair fight
	player.ManaRemaining = player.ManaTotal
	player.StaminaRemaining = player.StaminaTotal
	player.StatusEffects = nil

	msgs = append(msgs, Msg("=== ARENA BATTLE ===", "system"))
	msgs = append(msgs, Msg(fmt.Sprintf("Level %d %s vs Level %d %s",
		player.Level, player.Name, mob.Level, mob.Name), "system"))

	session.Combat = &CombatContext{
		Mob:                  mob,
		MobLoc:               -1,
		Location:             nil,
		Turn:                 0,
		IsArena:              true,
		ArenaTargetAccountID: session.ArenaTargetAccountID,
		ArenaTargetCharName:  session.ArenaTargetCharName,
	}
	session.State = StateCombat

	return GameResponse{
		Type:     "combat",
		Messages: msgs,
		State: &StateData{
			Screen: "combat",
			Player: MakePlayerState(player),
			Combat: MakeCombatView(session),
		},
		Options: combatActionOptions(),
	}
}
