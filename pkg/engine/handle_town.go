package engine

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"rpg-game/pkg/data"
	"rpg-game/pkg/game"
	"rpg-game/pkg/models"
)

// loadOrCreateTown loads the town from DB, or creates a default if none exists.
func (e *Engine) loadOrCreateTown(session *GameSession) (*models.Town, error) {
	if e.store == nil {
		// Local mode: create in-memory town
		town := game.GenerateDefaultTown(game.DefaultTownName)
		return &town, nil
	}
	town, err := e.store.LoadTown(game.DefaultTownName)
	if err != nil {
		// Town doesn't exist yet, create it
		town = game.GenerateDefaultTown(game.DefaultTownName)
		if saveErr := e.store.SaveTown(town); saveErr != nil {
			return nil, saveErr
		}
	}
	// Clean expired guests (24h)
	game.CleanExpiredGuests(&town, 86400)
	return &town, nil
}

// saveTown persists the town to DB.
func (e *Engine) saveTown(town *models.Town) error {
	if e.store == nil {
		return nil
	}
	return e.store.SaveTown(*town)
}

// townStateData builds StateData with town view.
func townStateData(screen string, session *GameSession, town *models.Town) *StateData {
	return &StateData{
		Screen: screen,
		Player: MakePlayerState(session.Player),
		Town:   MakeTownView(town, session.AccountID, session.Player.Name),
	}
}

// ─────────────────────────────────────────────────────────────────────
// Town Main Menu
// ─────────────────────────────────────────────────────────────────────

func (e *Engine) handleTownMain(session *GameSession, cmd GameCommand) GameResponse {
	town, err := e.loadOrCreateTown(session)
	if err != nil {
		session.State = StateMainMenu
		return BuildMainMenuResponse(session)
	}
	session.SelectedTown = town

	switch cmd.Value {
	case "1": // Enter Inn
		session.State = StateTownInn
		return e.handleTownInn(session, GameCommand{Type: "init"})
	case "2": // View Mayor
		session.State = StateTownMayor
		return e.handleTownMayor(session, GameCommand{Type: "init"})
	case "3": // Fetch Quests
		session.State = StateTownFetchQuests
		return e.handleTownFetchQuests(session, GameCommand{Type: "init"})
	case "4": // Challenge Mayor
		session.State = StateTownMayorChallenge
		return e.handleTownMayorChallenge(session, GameCommand{Type: "init"})
	case "0", "back":
		session.SelectedTown = nil
		session.State = StateMainMenu
		return BuildMainMenuResponse(session)
	}

	// Default: show town main menu
	msgs := []GameMessage{
		Msg("============================================================", "system"),
		Msg(fmt.Sprintf("  Town of %s", town.Name), "system"),
		Msg("============================================================", "system"),
		Msg(fmt.Sprintf("Tax Rate: %d%%", town.TaxRate), "system"),
		Msg(fmt.Sprintf("Inn Guests: %d", len(town.InnGuests)), "system"),
	}

	if town.Mayor != nil {
		mayorName := town.Mayor.NPCName
		if !town.Mayor.IsNPC {
			mayorName = town.Mayor.CharacterName
		}
		msgs = append(msgs, Msg(fmt.Sprintf("Mayor: %s (Level %d)", mayorName, town.Mayor.Level), "system"))
	}

	activeFetchQuests := 0
	for _, fq := range town.FetchQuests {
		if fq.Active && !fq.Completed {
			activeFetchQuests++
		}
	}
	msgs = append(msgs, Msg(fmt.Sprintf("Active Fetch Quests: %d", activeFetchQuests), "system"))

	options := []MenuOption{
		Opt("1", "Enter Inn"),
		Opt("2", "View Mayor"),
		Opt("3", "Fetch Quests"),
		Opt("4", "Challenge Mayor"),
		Opt("0", "Return to Main Menu"),
	}

	return GameResponse{
		Type:     "menu",
		Messages: msgs,
		State:    townStateData("town_main", session, town),
		Options:  options,
	}
}

// ─────────────────────────────────────────────────────────────────────
// Inn
// ─────────────────────────────────────────────────────────────────────

func (e *Engine) handleTownInn(session *GameSession, cmd GameCommand) GameResponse {
	town, err := e.loadOrCreateTown(session)
	if err != nil {
		session.State = StateTownMain
		return e.handleTownMain(session, GameCommand{Type: "init"})
	}
	session.SelectedTown = town

	switch cmd.Value {
	case "1": // Sleep
		session.State = StateTownInnSleep
		return e.handleTownInnSleep(session, GameCommand{Type: "init"})
	case "2": // Hire Guard
		session.State = StateTownInnHireGuard
		return e.handleTownInnHireGuard(session, GameCommand{Type: "init"})
	case "3": // View Guests
		session.State = StateTownInnViewGuests
		return e.handleTownInnViewGuests(session, GameCommand{Type: "init"})
	case "0", "back":
		session.State = StateTownMain
		return e.handleTownMain(session, GameCommand{Type: "init"})
	}

	// Default: show inn menu
	player := session.Player
	cost := game.InnSleepCost(player.Level)

	msgs := []GameMessage{
		Msg("============================================================", "system"),
		Msg("  The Crossroads Inn", "system"),
		Msg("============================================================", "system"),
		Msg(fmt.Sprintf("Sleep Cost: %d Gold (restores HP/MP/SP, leaves snapshot for PvP)", cost), "system"),
		Msg(fmt.Sprintf("Current Guests: %d", len(town.InnGuests)), "system"),
	}

	// Check if already a guest
	isGuest := false
	for _, g := range town.InnGuests {
		if g.AccountID == session.AccountID && g.CharacterName == player.Name {
			isGuest = true
			break
		}
	}
	if isGuest {
		msgs = append(msgs, Msg("You are currently registered as a guest.", "system"))
	}

	goldRes, hasGold := player.ResourceStorageMap["Gold"]
	canAfford := hasGold && goldRes.Stock >= cost

	options := []MenuOption{}
	if canAfford {
		options = append(options, Opt("1", fmt.Sprintf("Sleep (%d Gold)", cost)))
	} else {
		options = append(options, OptDisabled("1", fmt.Sprintf("Sleep (%d Gold) [not enough gold]", cost)))
	}
	options = append(options,
		Opt("2", "Hire Inn Guard"),
		Opt("3", "View Guests"),
		Opt("0", "Back to Town"),
	)

	return GameResponse{
		Type:     "menu",
		Messages: msgs,
		State:    townStateData("town_inn", session, town),
		Options:  options,
	}
}

func (e *Engine) handleTownInnSleep(session *GameSession, cmd GameCommand) GameResponse {
	town, err := e.loadOrCreateTown(session)
	if err != nil {
		session.State = StateTownInn
		return e.handleTownInn(session, GameCommand{Type: "init"})
	}
	session.SelectedTown = town
	player := session.Player

	cost := game.InnSleepCost(player.Level)
	goldRes, hasGold := player.ResourceStorageMap["Gold"]
	if !hasGold || goldRes.Stock < cost {
		session.State = StateTownInn
		resp := e.handleTownInn(session, GameCommand{Type: "init"})
		resp.Messages = append([]GameMessage{Msg("Not enough gold!", "error")}, resp.Messages...)
		return resp
	}

	// Deduct gold
	goldRes.Stock -= cost
	player.ResourceStorageMap["Gold"] = goldRes

	// Restore stats
	player.HitpointsRemaining = player.HitpointsTotal
	player.ManaRemaining = player.ManaTotal
	player.StaminaRemaining = player.StaminaTotal

	// Remove existing guest entry for this player, if any
	kept := []models.InnGuest{}
	for _, g := range town.InnGuests {
		if !(g.AccountID == session.AccountID && g.CharacterName == player.Name) {
			kept = append(kept, g)
		}
	}
	town.InnGuests = kept

	// Snapshot and add as guest
	guest := game.InnGuestFromCharacter(player, session.AccountID, cost)
	town.InnGuests = append(town.InnGuests, guest)

	// Add gold to treasury
	if town.Treasury == nil {
		town.Treasury = make(map[string]int)
	}
	town.Treasury["Gold"] += cost

	e.saveTown(town)

	// Save player state
	session.GameState.CharactersMap[player.Name] = *player

	msgs := []GameMessage{
		Msg(fmt.Sprintf("You rest at the inn for %d gold. HP/MP/SP fully restored!", cost), "heal"),
		Msg("Your snapshot has been registered at the inn.", "system"),
	}

	session.State = StateTownInn
	resp := e.handleTownInn(session, GameCommand{Type: "init"})
	resp.Messages = append(msgs, resp.Messages...)
	return resp
}

func (e *Engine) handleTownInnHireGuard(session *GameSession, cmd GameCommand) GameResponse {
	town, err := e.loadOrCreateTown(session)
	if err != nil {
		session.State = StateTownInn
		return e.handleTownInn(session, GameCommand{Type: "init"})
	}
	session.SelectedTown = town
	player := session.Player

	// Find player's inn guest entry
	guestIdx := -1
	for i, g := range town.InnGuests {
		if g.AccountID == session.AccountID && g.CharacterName == player.Name {
			guestIdx = i
			break
		}
	}

	if guestIdx == -1 {
		session.State = StateTownInn
		resp := e.handleTownInn(session, GameCommand{Type: "init"})
		resp.Messages = append([]GameMessage{Msg("You must sleep at the inn first before hiring guards!", "error")}, resp.Messages...)
		return resp
	}

	if cmd.Value == "0" || cmd.Value == "back" {
		session.State = StateTownInn
		return e.handleTownInn(session, GameCommand{Type: "init"})
	}

	// If a guard selection was made
	guardIdx, parseErr := strconv.Atoi(cmd.Value)
	if parseErr == nil && guardIdx >= 1 && guardIdx <= 3 {
		guard := game.GenerateGuard(player.Level)
		goldRes, hasGold := player.ResourceStorageMap["Gold"]
		if !hasGold || goldRes.Stock < guard.Cost {
			msgs := []GameMessage{Msg(fmt.Sprintf("Not enough gold! Need %d", guard.Cost), "error")}
			session.State = StateTownInn
			resp := e.handleTownInn(session, GameCommand{Type: "init"})
			resp.Messages = append(msgs, resp.Messages...)
			return resp
		}

		// Deduct gold and add guard
		goldRes.Stock -= guard.Cost
		player.ResourceStorageMap["Gold"] = goldRes
		guard.Hired = true
		town.InnGuests[guestIdx].HiredGuards = append(town.InnGuests[guestIdx].HiredGuards, guard)
		e.saveTown(town)
		session.GameState.CharactersMap[player.Name] = *player

		msgs := []GameMessage{
			Msg(fmt.Sprintf("Hired %s (Lv%d) for %d gold to guard your inn stay!", guard.Name, guard.Level, guard.Cost), "system"),
		}
		session.State = StateTownInn
		resp := e.handleTownInn(session, GameCommand{Type: "init"})
		resp.Messages = append(msgs, resp.Messages...)
		return resp
	}

	// Show guard options
	msgs := []GameMessage{
		Msg("============================================================", "system"),
		Msg("  Hire Inn Guards", "system"),
		Msg("============================================================", "system"),
		Msg(fmt.Sprintf("Current inn guards: %d", len(town.InnGuests[guestIdx].HiredGuards)), "system"),
		Msg("Guards protect your sleeping character from attackers.", "system"),
	}

	options := []MenuOption{}
	for i := 1; i <= 3; i++ {
		g := game.GenerateGuard(player.Level)
		goldRes, hasGold := player.ResourceStorageMap["Gold"]
		canAfford := hasGold && goldRes.Stock >= g.Cost
		label := fmt.Sprintf("%s (Lv%d, ATK:+%d, DEF:+%d) - %d Gold", g.Name, g.Level, g.AttackBonus, g.DefenseBonus, g.Cost)
		if canAfford {
			options = append(options, Opt(strconv.Itoa(i), label))
		} else {
			options = append(options, OptDisabled(strconv.Itoa(i), label+" [not enough gold]"))
		}
	}
	options = append(options, Opt("0", "Back"))

	return GameResponse{
		Type:     "menu",
		Messages: msgs,
		State:    townStateData("town_inn_hire_guard", session, town),
		Options:  options,
	}
}

func (e *Engine) handleTownInnViewGuests(session *GameSession, cmd GameCommand) GameResponse {
	town, err := e.loadOrCreateTown(session)
	if err != nil {
		session.State = StateTownInn
		return e.handleTownInn(session, GameCommand{Type: "init"})
	}
	session.SelectedTown = town
	player := session.Player

	if cmd.Value == "0" || cmd.Value == "back" {
		session.State = StateTownInn
		return e.handleTownInn(session, GameCommand{Type: "init"})
	}

	// Check if attacking a guest
	if cmd.Value != "" && cmd.Value != "init" {
		guestIdx, parseErr := strconv.Atoi(cmd.Value)
		if parseErr == nil && guestIdx >= 1 && guestIdx <= len(town.InnGuests) {
			target := &town.InnGuests[guestIdx-1]
			if target.AccountID == session.AccountID && target.CharacterName == player.Name {
				msgs := []GameMessage{Msg("You can't attack yourself!", "error")}
				// Re-show guest list
				session.State = StateTownInnViewGuests
				resp := e.showGuestList(session, town)
				resp.Messages = append(msgs, resp.Messages...)
				return resp
			}
			// Start PvP combat
			return e.startInnPvP(session, town, target, guestIdx-1)
		}
	}

	return e.showGuestList(session, town)
}

func (e *Engine) showGuestList(session *GameSession, town *models.Town) GameResponse {
	player := session.Player
	msgs := []GameMessage{
		Msg("============================================================", "system"),
		Msg("  Inn Guests", "system"),
		Msg("============================================================", "system"),
	}

	if len(town.InnGuests) == 0 {
		msgs = append(msgs, Msg("The inn is empty.", "narrative"))
	}

	options := []MenuOption{}
	for i, guest := range town.InnGuests {
		guardInfo := ""
		if len(guest.HiredGuards) > 0 {
			guardInfo = fmt.Sprintf(" [%d guards]", len(guest.HiredGuards))
		}
		label := fmt.Sprintf("%s (Lv%d)%s", guest.CharacterName, guest.Level, guardInfo)

		isOwn := guest.AccountID == session.AccountID && guest.CharacterName == player.Name
		if isOwn {
			label += " (You)"
			msgs = append(msgs, Msg(fmt.Sprintf("  %d. %s", i+1, label), "system"))
			options = append(options, OptDisabled(strconv.Itoa(i+1), "Attack "+label))
		} else {
			msgs = append(msgs, Msg(fmt.Sprintf("  %d. %s", i+1, label), "system"))
			options = append(options, Opt(strconv.Itoa(i+1), "Attack "+label))
		}
	}

	options = append(options, Opt("0", "Back"))

	return GameResponse{
		Type:     "menu",
		Messages: msgs,
		State:    townStateData("town_inn_view_guests", session, town),
		Options:  options,
	}
}

func (e *Engine) startInnPvP(session *GameSession, town *models.Town, target *models.InnGuest, guestIdx int) GameResponse {
	player := session.Player

	// Build synthetic monster from guest snapshot
	mob := game.InnGuestToMonster(target)

	// Set up combat guards from the target's hired guards
	var combatGuards []models.Guard
	for _, g := range target.HiredGuards {
		gc := g
		gc.HitpointsRemaining = gc.HitPoints
		combatGuards = append(combatGuards, gc)
	}

	// Resurrect player if needed
	if player.HitpointsRemaining <= 0 {
		player.HitpointsRemaining = player.HitpointsTotal
		player.Resurrections++
	}

	session.PvPTargetGuest = target
	session.Combat = &CombatContext{
		Mob:            mob,
		MobLoc:         -1,
		Turn:           0,
		Fled:           false,
		PlayerWon:      false,
		IsDefending:    false,
		IsPvP:          true,
		PvPTargetGuest: target,
		CombatGuards:   combatGuards,
		HasGuards:      len(combatGuards) > 0,
	}

	// Restore mana/stamina
	player.ManaRemaining = player.ManaTotal
	player.StaminaRemaining = player.StaminaTotal

	msgs := []GameMessage{
		Msg("============================================================", "system"),
		Msg(fmt.Sprintf("PVP ATTACK: %s vs %s (sleeping at the inn)", player.Name, target.CharacterName), "combat"),
		Msg("============================================================", "system"),
	}

	if len(combatGuards) > 0 {
		msgs = append(msgs, Msg(fmt.Sprintf("The target has %d guards defending them!", len(combatGuards)), "combat"))
		// Guards fight alongside the target in PvP
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

// ─────────────────────────────────────────────────────────────────────
// Mayor
// ─────────────────────────────────────────────────────────────────────

func (e *Engine) handleTownMayor(session *GameSession, cmd GameCommand) GameResponse {
	town, err := e.loadOrCreateTown(session)
	if err != nil {
		session.State = StateTownMain
		return e.handleTownMain(session, GameCommand{Type: "init"})
	}
	session.SelectedTown = town
	player := session.Player

	// Check if current player is mayor
	isMayor := town.Mayor != nil && !town.Mayor.IsNPC &&
		town.Mayor.AccountID == session.AccountID && town.Mayor.CharacterName == player.Name

	if isMayor {
		// Redirect to mayor management menu
		session.State = StateTownMayorMenu
		return e.handleTownMayorMenu(session, cmd)
	}

	if cmd.Value == "0" || cmd.Value == "back" {
		session.State = StateTownMain
		return e.handleTownMain(session, GameCommand{Type: "init"})
	}

	// Show mayor info
	msgs := []GameMessage{
		Msg("============================================================", "system"),
		Msg("  Town Mayor", "system"),
		Msg("============================================================", "system"),
	}

	if town.Mayor != nil {
		name := town.Mayor.NPCName
		if !town.Mayor.IsNPC {
			name = town.Mayor.CharacterName
		}
		msgs = append(msgs,
			Msg(fmt.Sprintf("Mayor: %s (Level %d)", name, town.Mayor.Level), "system"),
			Msg(fmt.Sprintf("Guards: %d", len(town.Mayor.Guards)), "system"),
			Msg(fmt.Sprintf("Monsters: %d", len(town.Mayor.Monsters)), "system"),
			Msg(fmt.Sprintf("Tax Rate: %d%%", town.TaxRate), "system"),
		)
	}

	options := []MenuOption{
		Opt("0", "Back to Town"),
	}

	return GameResponse{
		Type:     "menu",
		Messages: msgs,
		State:    townStateData("town_mayor", session, town),
		Options:  options,
	}
}

// ─────────────────────────────────────────────────────────────────────
// Mayor Challenge
// ─────────────────────────────────────────────────────────────────────

func (e *Engine) handleTownMayorChallenge(session *GameSession, cmd GameCommand) GameResponse {
	town, err := e.loadOrCreateTown(session)
	if err != nil {
		session.State = StateTownMain
		return e.handleTownMain(session, GameCommand{Type: "init"})
	}
	session.SelectedTown = town
	player := session.Player

	if town.Mayor == nil {
		session.State = StateTownMain
		resp := e.handleTownMain(session, GameCommand{Type: "init"})
		resp.Messages = append([]GameMessage{Msg("No mayor to challenge!", "error")}, resp.Messages...)
		return resp
	}

	// Check if player is already mayor
	if !town.Mayor.IsNPC && town.Mayor.AccountID == session.AccountID && town.Mayor.CharacterName == player.Name {
		session.State = StateTownMain
		resp := e.handleTownMain(session, GameCommand{Type: "init"})
		resp.Messages = append([]GameMessage{Msg("You are already the mayor!", "error")}, resp.Messages...)
		return resp
	}

	if cmd.Value == "0" || cmd.Value == "back" {
		session.State = StateTownMain
		return e.handleTownMain(session, GameCommand{Type: "init"})
	}

	if cmd.Value == "1" || cmd.Type == "init" {
		// Start challenge - Phase 0: fight guards
		return e.startMayorChallengePhase(session, town, 0)
	}

	// Show challenge info
	mayorName := town.Mayor.NPCName
	if !town.Mayor.IsNPC {
		mayorName = town.Mayor.CharacterName
	}

	msgs := []GameMessage{
		Msg("============================================================", "system"),
		Msg("  Challenge the Mayor", "system"),
		Msg("============================================================", "system"),
		Msg(fmt.Sprintf("You are challenging %s (Level %d)!", mayorName, town.Mayor.Level), "combat"),
		Msg(fmt.Sprintf("Phase 1: Defeat %d guards", len(town.Mayor.Guards)), "system"),
		Msg(fmt.Sprintf("Phase 2: Defeat %d monsters", len(town.Mayor.Monsters)), "system"),
		Msg("Phase 3: Defeat the Mayor in single combat", "system"),
		Msg("", "system"),
		Msg("If you win, you become the new mayor!", "narrative"),
	}

	options := []MenuOption{
		Opt("1", "Begin Challenge!"),
		Opt("0", "Back to Town"),
	}

	return GameResponse{
		Type:     "menu",
		Messages: msgs,
		State:    townStateData("town_mayor_challenge", session, town),
		Options:  options,
	}
}

func (e *Engine) startMayorChallengePhase(session *GameSession, town *models.Town, phase int) GameResponse {
	player := session.Player

	if player.HitpointsRemaining <= 0 {
		player.HitpointsRemaining = player.HitpointsTotal
		player.Resurrections++
	}

	switch phase {
	case 0: // Fight guards
		if len(town.Mayor.Guards) == 0 {
			// Skip to phase 1
			return e.startMayorChallengePhase(session, town, 1)
		}
		// Create a combined guard monster
		guard := town.Mayor.Guards[0]
		mob := models.Monster{
			Name:               guard.Name + " (Mayor's Guard)",
			Level:              guard.Level,
			Rank:               guard.Level/3 + 1,
			HitpointsTotal:     guard.HitPoints,
			HitpointsNatural:   guard.HitPoints,
			HitpointsRemaining: guard.HitPoints,
			ManaTotal:          30,
			ManaNatural:        30,
			ManaRemaining:      30,
			StaminaTotal:       30,
			StaminaNatural:     30,
			StaminaRemaining:   30,
			AttackRolls:        guard.AttackRolls,
			DefenseRolls:       guard.DefenseRolls,
			StatsMod:           guard.StatsMod,
			EquipmentMap:       guard.EquipmentMap,
			Inventory:          guard.Inventory,
			LearnedSkills:      game.AssignMonsterSkills("humanoid", guard.Level),
			StatusEffects:      []models.StatusEffect{},
			Resistances:        guard.Resistances,
			MonsterType:        "humanoid",
		}

		session.Combat = &CombatContext{
			Mob:                 mob,
			MobLoc:              -1,
			Turn:                0,
			IsMayorChallenge:    true,
			MayorChallengePhase: 0,
		}

		player.ManaRemaining = player.ManaTotal
		player.StaminaRemaining = player.StaminaTotal

		remainingGuards := len(town.Mayor.Guards) - 1
		msgs := []GameMessage{
			Msg(fmt.Sprintf("MAYOR CHALLENGE - Phase 1: Guards (%d remaining after this)", remainingGuards), "combat"),
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

	case 1: // Fight monsters
		if len(town.Mayor.Monsters) == 0 {
			return e.startMayorChallengePhase(session, town, 2)
		}
		mob := town.Mayor.Monsters[0]
		mob.HitpointsRemaining = mob.HitpointsTotal
		mob.ManaRemaining = mob.ManaTotal
		mob.StaminaRemaining = mob.StaminaTotal

		session.Combat = &CombatContext{
			Mob:                 mob,
			MobLoc:              -1,
			Turn:                0,
			IsMayorChallenge:    true,
			MayorChallengePhase: 1,
		}

		player.ManaRemaining = player.ManaTotal
		player.StaminaRemaining = player.StaminaTotal

		remainingMonsters := len(town.Mayor.Monsters) - 1
		msgs := []GameMessage{
			Msg(fmt.Sprintf("MAYOR CHALLENGE - Phase 2: Monsters (%d remaining after this)", remainingMonsters), "combat"),
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

	case 2: // Fight mayor
		mob := game.MayorToMonster(town.Mayor)

		session.Combat = &CombatContext{
			Mob:                 mob,
			MobLoc:              -1,
			Turn:                0,
			IsMayorChallenge:    true,
			MayorChallengePhase: 2,
		}

		player.ManaRemaining = player.ManaTotal
		player.StaminaRemaining = player.StaminaTotal

		msgs := []GameMessage{
			Msg("MAYOR CHALLENGE - Final Phase: The Mayor!", "combat"),
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

	session.State = StateTownMain
	return e.handleTownMain(session, GameCommand{Type: "init"})
}

// ─────────────────────────────────────────────────────────────────────
// Mayor Menu (for the current mayor)
// ─────────────────────────────────────────────────────────────────────

func (e *Engine) handleTownMayorMenu(session *GameSession, cmd GameCommand) GameResponse {
	town, err := e.loadOrCreateTown(session)
	if err != nil {
		session.State = StateTownMain
		return e.handleTownMain(session, GameCommand{Type: "init"})
	}
	session.SelectedTown = town

	switch cmd.Value {
	case "1": // Set Tax
		session.State = StateTownMayorSetTax
		return GameResponse{
			Type:     "menu",
			Messages: []GameMessage{Msg(fmt.Sprintf("Current tax rate: %d%%. Enter new rate (0-50):", town.TaxRate), "system")},
			State:    townStateData("town_mayor_set_tax", session, town),
			Prompt:   "New tax rate (0-50): ",
		}
	case "2": // Create Fetch Quest
		session.State = StateTownMayorCreateQuest
		return e.handleTownMayorCreateQuest(session, GameCommand{Type: "init"})
	case "3": // Hire Guard
		session.State = StateTownMayorHireGuard
		return e.handleTownMayorHireGuard(session, GameCommand{Type: "init"})
	case "4": // Hire Monster
		session.State = StateTownMayorHireMonster
		return e.handleTownMayorHireMonster(session, GameCommand{Type: "init"})
	case "5": // View Treasury
		msgs := []GameMessage{
			Msg("============================================================", "system"),
			Msg("  Town Treasury", "system"),
			Msg("============================================================", "system"),
		}
		for res, amount := range town.Treasury {
			msgs = append(msgs, Msg(fmt.Sprintf("  %s: %d", res, amount), "system"))
		}
		if len(town.Treasury) == 0 {
			msgs = append(msgs, Msg("  Treasury is empty.", "narrative"))
		}
		return GameResponse{
			Type:     "menu",
			Messages: msgs,
			State:    townStateData("town_mayor_menu", session, town),
			Options:  e.mayorMenuOptions(),
		}
	case "6": // Abdicate
		town.Mayor = nil
		npcMayor := game.GenerateNPCMayor(10)
		town.Mayor = &npcMayor
		e.saveTown(town)
		msgs := []GameMessage{Msg("You have abdicated your position. A new NPC mayor has been appointed.", "narrative")}
		session.State = StateTownMain
		resp := e.handleTownMain(session, GameCommand{Type: "init"})
		resp.Messages = append(msgs, resp.Messages...)
		return resp
	case "0", "back":
		session.State = StateTownMain
		return e.handleTownMain(session, GameCommand{Type: "init"})
	}

	// Show mayor menu
	msgs := []GameMessage{
		Msg("============================================================", "system"),
		Msg("  Mayor's Office", "system"),
		Msg("============================================================", "system"),
		Msg(fmt.Sprintf("Tax Rate: %d%%", town.TaxRate), "system"),
		Msg(fmt.Sprintf("Guards: %d", len(town.Mayor.Guards)), "system"),
		Msg(fmt.Sprintf("Monsters: %d", len(town.Mayor.Monsters)), "system"),
	}
	treasuryGold := town.Treasury["Gold"]
	msgs = append(msgs, Msg(fmt.Sprintf("Treasury Gold: %d", treasuryGold), "system"))

	return GameResponse{
		Type:     "menu",
		Messages: msgs,
		State:    townStateData("town_mayor_menu", session, town),
		Options:  e.mayorMenuOptions(),
	}
}

func (e *Engine) mayorMenuOptions() []MenuOption {
	return []MenuOption{
		Opt("1", "Set Tax Rate"),
		Opt("2", "Create Fetch Quest"),
		Opt("3", "Hire Guard (from Treasury)"),
		Opt("4", "Hire Monster (from Treasury)"),
		Opt("5", "View Treasury"),
		Opt("6", "Abdicate"),
		Opt("0", "Back to Town"),
	}
}

func (e *Engine) handleTownMayorSetTax(session *GameSession, cmd GameCommand) GameResponse {
	town, err := e.loadOrCreateTown(session)
	if err != nil {
		session.State = StateTownMayorMenu
		return e.handleTownMayorMenu(session, GameCommand{Type: "init"})
	}
	session.SelectedTown = town

	rate, parseErr := strconv.Atoi(cmd.Value)
	if parseErr != nil || rate < 0 || rate > 50 {
		session.State = StateTownMayorMenu
		resp := e.handleTownMayorMenu(session, GameCommand{Type: "init"})
		resp.Messages = append([]GameMessage{Msg("Invalid tax rate! Must be 0-50.", "error")}, resp.Messages...)
		return resp
	}

	town.TaxRate = rate
	e.saveTown(town)

	msgs := []GameMessage{Msg(fmt.Sprintf("Tax rate set to %d%%!", rate), "system")}
	session.State = StateTownMayorMenu
	resp := e.handleTownMayorMenu(session, GameCommand{Type: "init"})
	resp.Messages = append(msgs, resp.Messages...)
	return resp
}

// ─────────────────────────────────────────────────────────────────────
// Mayor: Create Fetch Quest (multi-step)
// ─────────────────────────────────────────────────────────────────────

func (e *Engine) handleTownMayorCreateQuest(session *GameSession, cmd GameCommand) GameResponse {
	town, err := e.loadOrCreateTown(session)
	if err != nil {
		session.State = StateTownMayorMenu
		return e.handleTownMayorMenu(session, GameCommand{Type: "init"})
	}
	session.SelectedTown = town

	if cmd.Value == "0" || cmd.Value == "back" {
		session.State = StateTownMayorMenu
		return e.handleTownMayorMenu(session, GameCommand{Type: "init"})
	}

	// Check if a resource was selected
	for _, res := range data.ResourceTypes {
		if cmd.Value == res {
			session.TownQuestResource = res
			session.State = StateTownMayorCreateQuestAmount
			return GameResponse{
				Type:     "menu",
				Messages: []GameMessage{Msg(fmt.Sprintf("Creating fetch quest for %s. How much?", res), "system")},
				State:    townStateData("town_mayor_create_quest_amount", session, town),
				Prompt:   "Amount required: ",
			}
		}
	}

	// Show resource selection
	msgs := []GameMessage{
		Msg("Create Fetch Quest - Select Resource:", "system"),
	}
	options := []MenuOption{}
	for _, res := range data.ResourceTypes {
		options = append(options, Opt(res, res))
	}
	options = append(options, Opt("0", "Cancel"))

	return GameResponse{
		Type:     "menu",
		Messages: msgs,
		State:    townStateData("town_mayor_create_quest", session, town),
		Options:  options,
	}
}

func (e *Engine) handleTownMayorCreateQuestAmount(session *GameSession, cmd GameCommand) GameResponse {
	town := session.SelectedTown
	amount, parseErr := strconv.Atoi(cmd.Value)
	if parseErr != nil || amount <= 0 || amount > 1000 {
		session.State = StateTownMayorMenu
		resp := e.handleTownMayorMenu(session, GameCommand{Type: "init"})
		resp.Messages = append([]GameMessage{Msg("Invalid amount! Must be 1-1000.", "error")}, resp.Messages...)
		return resp
	}

	session.TownQuestAmount = amount
	session.State = StateTownMayorCreateQuestReward

	return GameResponse{
		Type:     "menu",
		Messages: []GameMessage{Msg(fmt.Sprintf("Quest: Deliver %d %s. Set gold reward (from treasury, max %d):", amount, session.TownQuestResource, town.Treasury["Gold"]), "system")},
		State:    townStateData("town_mayor_create_quest_reward", session, town),
		Prompt:   "Gold reward: ",
	}
}

func (e *Engine) handleTownMayorCreateQuestReward(session *GameSession, cmd GameCommand) GameResponse {
	town, err := e.loadOrCreateTown(session)
	if err != nil {
		session.State = StateTownMayorMenu
		return e.handleTownMayorMenu(session, GameCommand{Type: "init"})
	}
	session.SelectedTown = town

	reward, parseErr := strconv.Atoi(cmd.Value)
	if parseErr != nil || reward <= 0 {
		session.State = StateTownMayorMenu
		resp := e.handleTownMayorMenu(session, GameCommand{Type: "init"})
		resp.Messages = append([]GameMessage{Msg("Invalid reward amount!", "error")}, resp.Messages...)
		return resp
	}

	treasuryGold := town.Treasury["Gold"]
	if reward > treasuryGold {
		session.State = StateTownMayorMenu
		resp := e.handleTownMayorMenu(session, GameCommand{Type: "init"})
		resp.Messages = append([]GameMessage{Msg(fmt.Sprintf("Not enough gold in treasury! Have %d", treasuryGold), "error")}, resp.Messages...)
		return resp
	}

	// Create the quest
	questID := fmt.Sprintf("fq_%d", time.Now().UnixNano())
	rewardXP := session.TownQuestAmount * 5

	fq := models.FetchQuest{
		ID:          questID,
		Name:        fmt.Sprintf("Deliver %d %s", session.TownQuestAmount, session.TownQuestResource),
		Description: fmt.Sprintf("The mayor requests %d %s for the town.", session.TownQuestAmount, session.TownQuestResource),
		Resource:    session.TownQuestResource,
		Amount:      session.TownQuestAmount,
		RewardGold:  reward,
		RewardXP:    rewardXP,
		CreatedBy:   session.Player.Name,
		Active:      true,
	}

	// Deduct gold from treasury
	town.Treasury["Gold"] -= reward
	town.FetchQuests = append(town.FetchQuests, fq)
	e.saveTown(town)

	msgs := []GameMessage{
		Msg(fmt.Sprintf("Fetch quest created! Deliver %d %s for %d gold + %d XP", session.TownQuestAmount, session.TownQuestResource, reward, rewardXP), "system"),
	}

	session.State = StateTownMayorMenu
	resp := e.handleTownMayorMenu(session, GameCommand{Type: "init"})
	resp.Messages = append(msgs, resp.Messages...)
	return resp
}

// ─────────────────────────────────────────────────────────────────────
// Mayor: Hire Guard / Monster from Treasury
// ─────────────────────────────────────────────────────────────────────

func (e *Engine) handleTownMayorHireGuard(session *GameSession, cmd GameCommand) GameResponse {
	town, err := e.loadOrCreateTown(session)
	if err != nil {
		session.State = StateTownMayorMenu
		return e.handleTownMayorMenu(session, GameCommand{Type: "init"})
	}
	session.SelectedTown = town

	if cmd.Value == "0" || cmd.Value == "back" {
		session.State = StateTownMayorMenu
		return e.handleTownMayorMenu(session, GameCommand{Type: "init"})
	}

	if cmd.Value == "1" {
		guard := game.GenerateGuard(town.Mayor.Level)
		treasuryGold := town.Treasury["Gold"]
		if treasuryGold < guard.Cost {
			msgs := []GameMessage{Msg(fmt.Sprintf("Not enough treasury gold! Need %d, have %d", guard.Cost, treasuryGold), "error")}
			session.State = StateTownMayorMenu
			resp := e.handleTownMayorMenu(session, GameCommand{Type: "init"})
			resp.Messages = append(msgs, resp.Messages...)
			return resp
		}

		town.Treasury["Gold"] -= guard.Cost
		guard.Hired = true
		town.Mayor.Guards = append(town.Mayor.Guards, guard)
		e.saveTown(town)

		msgs := []GameMessage{
			Msg(fmt.Sprintf("Hired %s (Lv%d) for %d gold from treasury!", guard.Name, guard.Level, guard.Cost), "system"),
		}
		session.State = StateTownMayorMenu
		resp := e.handleTownMayorMenu(session, GameCommand{Type: "init"})
		resp.Messages = append(msgs, resp.Messages...)
		return resp
	}

	// Show hire guard info
	guard := game.GenerateGuard(town.Mayor.Level)
	treasuryGold := town.Treasury["Gold"]
	msgs := []GameMessage{
		Msg("Hire a guard for the mayor's defense.", "system"),
		Msg(fmt.Sprintf("Cost: ~%d gold (Treasury: %d gold)", guard.Cost, treasuryGold), "system"),
	}

	options := []MenuOption{}
	if treasuryGold >= guard.Cost {
		options = append(options, Opt("1", fmt.Sprintf("Hire Guard (~%d gold)", guard.Cost)))
	} else {
		options = append(options, OptDisabled("1", "Hire Guard [insufficient treasury gold]"))
	}
	options = append(options, Opt("0", "Back"))

	return GameResponse{
		Type:     "menu",
		Messages: msgs,
		State:    townStateData("town_mayor_hire_guard", session, town),
		Options:  options,
	}
}

func (e *Engine) handleTownMayorHireMonster(session *GameSession, cmd GameCommand) GameResponse {
	town, err := e.loadOrCreateTown(session)
	if err != nil {
		session.State = StateTownMayorMenu
		return e.handleTownMayorMenu(session, GameCommand{Type: "init"})
	}
	session.SelectedTown = town

	if cmd.Value == "0" || cmd.Value == "back" {
		session.State = StateTownMayorMenu
		return e.handleTownMayorMenu(session, GameCommand{Type: "init"})
	}

	monsterCost := 100 + town.Mayor.Level*20

	if cmd.Value == "1" {
		treasuryGold := town.Treasury["Gold"]
		if treasuryGold < monsterCost {
			msgs := []GameMessage{Msg(fmt.Sprintf("Not enough treasury gold! Need %d, have %d", monsterCost, treasuryGold), "error")}
			session.State = StateTownMayorMenu
			resp := e.handleTownMayorMenu(session, GameCommand{Type: "init"})
			resp.Messages = append(msgs, resp.Messages...)
			return resp
		}

		monsterName := data.MonsterNames[rand.Intn(len(data.MonsterNames))]
		rank := town.Mayor.Level/3 + 1
		if rank > 5 {
			rank = 5
		}
		mob := game.GenerateMonster(monsterName, town.Mayor.Level, rank)
		mob.StatsMod = game.CalculateItemMods(mob.EquipmentMap)
		mob.HitpointsTotal = mob.HitpointsNatural + mob.StatsMod.HitPointMod
		mob.HitpointsRemaining = mob.HitpointsTotal

		town.Treasury["Gold"] -= monsterCost
		town.Mayor.Monsters = append(town.Mayor.Monsters, mob)
		e.saveTown(town)

		msgs := []GameMessage{
			Msg(fmt.Sprintf("Hired %s (Lv%d) for %d gold from treasury!", mob.Name, mob.Level, monsterCost), "system"),
		}
		session.State = StateTownMayorMenu
		resp := e.handleTownMayorMenu(session, GameCommand{Type: "init"})
		resp.Messages = append(msgs, resp.Messages...)
		return resp
	}

	treasuryGold := town.Treasury["Gold"]
	msgs := []GameMessage{
		Msg("Hire a monster for the mayor's defense.", "system"),
		Msg(fmt.Sprintf("Cost: %d gold (Treasury: %d gold)", monsterCost, treasuryGold), "system"),
	}

	options := []MenuOption{}
	if treasuryGold >= monsterCost {
		options = append(options, Opt("1", fmt.Sprintf("Hire Monster (%d gold)", monsterCost)))
	} else {
		options = append(options, OptDisabled("1", "Hire Monster [insufficient treasury gold]"))
	}
	options = append(options, Opt("0", "Back"))

	return GameResponse{
		Type:     "menu",
		Messages: msgs,
		State:    townStateData("town_mayor_hire_monster", session, town),
		Options:  options,
	}
}

// ─────────────────────────────────────────────────────────────────────
// Fetch Quests
// ─────────────────────────────────────────────────────────────────────

func (e *Engine) handleTownFetchQuests(session *GameSession, cmd GameCommand) GameResponse {
	town, err := e.loadOrCreateTown(session)
	if err != nil {
		session.State = StateTownMain
		return e.handleTownMain(session, GameCommand{Type: "init"})
	}
	session.SelectedTown = town
	player := session.Player

	if cmd.Value == "0" || cmd.Value == "back" {
		session.State = StateTownMain
		return e.handleTownMain(session, GameCommand{Type: "init"})
	}

	// Handle accept/complete actions
	if len(cmd.Value) > 7 && cmd.Value[:7] == "accept:" {
		questID := cmd.Value[7:]
		return e.acceptFetchQuest(session, town, questID)
	}
	if len(cmd.Value) > 9 && cmd.Value[:9] == "complete:" {
		questID := cmd.Value[9:]
		return e.completeFetchQuest(session, town, questID)
	}

	// Show quest list
	msgs := []GameMessage{
		Msg("============================================================", "system"),
		Msg("  Fetch Quests", "system"),
		Msg("============================================================", "system"),
	}

	options := []MenuOption{}
	hasQuests := false

	for _, fq := range town.FetchQuests {
		if !fq.Active || fq.Completed {
			continue
		}
		hasQuests = true
		label := fmt.Sprintf("%s - %d %s for %d Gold + %d XP", fq.Name, fq.Amount, fq.Resource, fq.RewardGold, fq.RewardXP)
		msgs = append(msgs, Msg(label, "system"))

		if fq.ClaimedBy == "" {
			options = append(options, Opt("accept:"+fq.ID, "Accept: "+fq.Name))
		} else if fq.ClaimedBy == player.Name {
			// Check if player has enough resources
			res, exists := player.ResourceStorageMap[fq.Resource]
			if exists && res.Stock >= fq.Amount {
				options = append(options, Opt("complete:"+fq.ID, "Complete: "+fq.Name))
			} else {
				have := 0
				if exists {
					have = res.Stock
				}
				options = append(options, OptDisabled("complete:"+fq.ID, fmt.Sprintf("Complete: %s (need %d %s, have %d)", fq.Name, fq.Amount, fq.Resource, have)))
			}
		} else {
			msgs = append(msgs, Msg(fmt.Sprintf("  (Claimed by %s)", fq.ClaimedBy), "system"))
		}
	}

	if !hasQuests {
		msgs = append(msgs, Msg("No active fetch quests.", "narrative"))
	}

	options = append(options, Opt("0", "Back to Town"))

	return GameResponse{
		Type:     "menu",
		Messages: msgs,
		State:    townStateData("town_fetch_quests", session, town),
		Options:  options,
	}
}

func (e *Engine) acceptFetchQuest(session *GameSession, town *models.Town, questID string) GameResponse {
	player := session.Player

	for i, fq := range town.FetchQuests {
		if fq.ID == questID && fq.Active && !fq.Completed && fq.ClaimedBy == "" {
			town.FetchQuests[i].ClaimedBy = player.Name
			e.saveTown(town)

			msgs := []GameMessage{Msg(fmt.Sprintf("Accepted quest: %s", fq.Name), "system")}
			session.State = StateTownFetchQuests
			resp := e.handleTownFetchQuests(session, GameCommand{Type: "init"})
			resp.Messages = append(msgs, resp.Messages...)
			return resp
		}
	}

	msgs := []GameMessage{Msg("Quest not found or already claimed!", "error")}
	session.State = StateTownFetchQuests
	resp := e.handleTownFetchQuests(session, GameCommand{Type: "init"})
	resp.Messages = append(msgs, resp.Messages...)
	return resp
}

func (e *Engine) completeFetchQuest(session *GameSession, town *models.Town, questID string) GameResponse {
	player := session.Player

	for i, fq := range town.FetchQuests {
		if fq.ID == questID && fq.Active && !fq.Completed && fq.ClaimedBy == player.Name {
			// Check resources
			res, exists := player.ResourceStorageMap[fq.Resource]
			if !exists || res.Stock < fq.Amount {
				msgs := []GameMessage{Msg("Not enough resources!", "error")}
				session.State = StateTownFetchQuests
				resp := e.handleTownFetchQuests(session, GameCommand{Type: "init"})
				resp.Messages = append(msgs, resp.Messages...)
				return resp
			}

			// Deduct resources
			res.Stock -= fq.Amount
			player.ResourceStorageMap[fq.Resource] = res

			// Award gold
			goldRes, hasGold := player.ResourceStorageMap["Gold"]
			if !hasGold {
				goldRes = models.Resource{Name: "Gold", Stock: 0}
			}
			goldRes.Stock += fq.RewardGold
			player.ResourceStorageMap["Gold"] = goldRes

			// Award XP
			player.Experience += fq.RewardXP

			// Mark complete
			town.FetchQuests[i].Completed = true
			e.saveTown(town)
			session.GameState.CharactersMap[player.Name] = *player

			// Level up check
			prevLevel := player.Level
			game.LevelUp(player)

			msgs := []GameMessage{
				Msg(fmt.Sprintf("Quest complete! Delivered %d %s", fq.Amount, fq.Resource), "system"),
				Msg(fmt.Sprintf("Reward: %d Gold, %d XP", fq.RewardGold, fq.RewardXP), "loot"),
			}
			if player.Level > prevLevel {
				msgs = append(msgs, Msg(fmt.Sprintf("LEVEL UP! Now level %d!", player.Level), "levelup"))
			}

			session.State = StateTownFetchQuests
			resp := e.handleTownFetchQuests(session, GameCommand{Type: "init"})
			resp.Messages = append(msgs, resp.Messages...)
			return resp
		}
	}

	msgs := []GameMessage{Msg("Quest not found!", "error")}
	session.State = StateTownFetchQuests
	resp := e.handleTownFetchQuests(session, GameCommand{Type: "init"})
	resp.Messages = append(msgs, resp.Messages...)
	return resp
}

// ─────────────────────────────────────────────────────────────────────
// Combat resolution helpers for PvP and Mayor Challenge
// ─────────────────────────────────────────────────────────────────────

// resolvePvPWin handles PvP victory at the inn.
func (e *Engine) resolvePvPWin(session *GameSession, msgs []GameMessage) {
	town, err := e.loadOrCreateTown(session)
	if err != nil {
		return
	}

	combat := session.Combat
	target := combat.PvPTargetGuest
	player := session.Player

	if target == nil {
		return
	}

	// Transfer 1-3 equipment items from target
	transferred := 0
	maxTransfer := rand.Intn(3) + 1
	for slot, item := range target.EquipmentMap {
		if transferred >= maxTransfer {
			break
		}
		game.EquipBestItem(item, &player.EquipmentMap, &player.Inventory)
		delete(target.EquipmentMap, slot)
		transferred++
	}

	// Remove target from inn
	kept := []models.InnGuest{}
	for _, g := range town.InnGuests {
		if !(g.AccountID == target.AccountID && g.CharacterName == target.CharacterName) {
			kept = append(kept, g)
		}
	}
	town.InnGuests = kept

	// Log the attack
	town.AttackLog = append(town.AttackLog, models.TownAttackLog{
		AttackerName: player.Name,
		TargetName:   target.CharacterName,
		AttackType:   "inn_guest",
		Success:      true,
		Timestamp:    time.Now().Unix(),
		Details:      fmt.Sprintf("%s defeated %s at the inn and looted %d items", player.Name, target.CharacterName, transferred),
	})

	e.saveTown(town)
}

// resolvePvPLoss handles PvP defeat at the inn.
func (e *Engine) resolvePvPLoss(session *GameSession, msgs []GameMessage) {
	town, err := e.loadOrCreateTown(session)
	if err != nil {
		return
	}

	combat := session.Combat
	target := combat.PvPTargetGuest
	player := session.Player

	if target == nil {
		return
	}

	town.AttackLog = append(town.AttackLog, models.TownAttackLog{
		AttackerName: player.Name,
		TargetName:   target.CharacterName,
		AttackType:   "inn_guest",
		Success:      false,
		Timestamp:    time.Now().Unix(),
		Details:      fmt.Sprintf("%s was defeated attacking %s at the inn", player.Name, target.CharacterName),
	})

	e.saveTown(town)
}

// resolveMayorChallengeWin handles winning a mayor challenge phase.
func (e *Engine) resolveMayorChallengeWin(session *GameSession, msgs []GameMessage) ([]GameMessage, bool) {
	town, err := e.loadOrCreateTown(session)
	if err != nil {
		return msgs, true // done
	}
	session.SelectedTown = town
	combat := session.Combat

	switch combat.MayorChallengePhase {
	case 0: // Beat a guard
		if len(town.Mayor.Guards) > 0 {
			town.Mayor.Guards = town.Mayor.Guards[1:]
			e.saveTown(town)
		}
		if len(town.Mayor.Guards) > 0 {
			// More guards - continue phase 0
			return msgs, false
		}
		// Move to phase 1
		combat.MayorChallengePhase = 1
		return msgs, false

	case 1: // Beat a monster
		if len(town.Mayor.Monsters) > 0 {
			town.Mayor.Monsters = town.Mayor.Monsters[1:]
			e.saveTown(town)
		}
		if len(town.Mayor.Monsters) > 0 {
			// More monsters
			return msgs, false
		}
		// Move to phase 2
		combat.MayorChallengePhase = 2
		return msgs, false

	case 2: // Beat the mayor!
		player := session.Player

		// Install player as new mayor
		newMayor := game.MayorFromCharacter(player, session.AccountID)
		town.Mayor = &newMayor

		// Log the challenge
		town.AttackLog = append(town.AttackLog, models.TownAttackLog{
			AttackerName: player.Name,
			TargetName:   "Mayor",
			AttackType:   "mayor_challenge",
			Success:      true,
			Timestamp:    time.Now().Unix(),
			Details:      fmt.Sprintf("%s defeated the mayor and seized control!", player.Name),
		})

		e.saveTown(town)

		msgs = append(msgs,
			Msg("============================================================", "system"),
			Msg("YOU ARE NOW THE MAYOR!", "narrative"),
			Msg("============================================================", "system"),
			Msg("You can now set taxes, create quests, and hire guards.", "system"),
		)
		return msgs, true
	}

	return msgs, true
}

// resolveMayorChallengeLoss handles losing a mayor challenge.
func (e *Engine) resolveMayorChallengeLoss(session *GameSession, msgs []GameMessage) {
	town, err := e.loadOrCreateTown(session)
	if err != nil {
		return
	}

	player := session.Player
	town.AttackLog = append(town.AttackLog, models.TownAttackLog{
		AttackerName: player.Name,
		TargetName:   "Mayor",
		AttackType:   "mayor_challenge",
		Success:      false,
		Timestamp:    time.Now().Unix(),
		Details:      fmt.Sprintf("%s was defeated challenging the mayor", player.Name),
	})

	e.saveTown(town)
}
