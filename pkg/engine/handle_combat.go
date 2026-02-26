package engine

import (
	"fmt"
	"math/rand"
	"strconv"

	"rpg-game/pkg/data"
	"rpg-game/pkg/game"
	"rpg-game/pkg/models"
)

// handleCombatGuardPrompt processes the player's decision on whether to bring guards into combat.
func (e *Engine) handleCombatGuardPrompt(session *GameSession, cmd GameCommand) GameResponse {
	combat := session.Combat
	player := session.Player
	mob := &combat.Mob
	msgs := []GameMessage{}

	if cmd.Value == "y" || cmd.Value == "Y" {
		msgs = append(msgs, Msg("Guards are joining the battle!", "system"))
		for _, g := range combat.CombatGuards {
			msgs = append(msgs, Msg(fmt.Sprintf("  %s (Lv%d, HP:%d) ready for combat", g.Name, g.Level, g.HitPoints), "system"))
		}
	} else {
		combat.CombatGuards = []models.Guard{}
		combat.HasGuards = false
		msgs = append(msgs, Msg("Fighting without guards...", "system"))
	}

	// Restore mana and stamina to full for both combatants
	player.ManaRemaining = player.ManaTotal
	player.StaminaRemaining = player.StaminaTotal
	mob.ManaRemaining = mob.ManaTotal
	mob.StaminaRemaining = mob.StaminaTotal

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

// handleCombatAction processes one full combat turn based on the player's chosen action.
func (e *Engine) handleCombatAction(session *GameSession, cmd GameCommand) GameResponse {
	combat := session.Combat
	player := session.Player
	mob := &combat.Mob
	msgs := []GameMessage{}

	// Increment turn
	combat.Turn++
	combat.IsDefending = false
	msgs = append(msgs, Msg(fmt.Sprintf("--- Turn %d ---", combat.Turn), "system"))

	// =====================================================================
	// Process status effects for player (inline, no calls to game.ProcessStatusEffects)
	// =====================================================================
	for i := len(player.StatusEffects) - 1; i >= 0; i-- {
		effect := &player.StatusEffects[i]
		switch effect.Type {
		case "poison":
			player.HitpointsRemaining -= effect.Potency
			msgs = append(msgs, Msg(fmt.Sprintf("%s takes %d poison damage!", player.Name, effect.Potency), "damage"))
		case "burn":
			player.HitpointsRemaining -= effect.Potency
			msgs = append(msgs, Msg(fmt.Sprintf("%s takes %d burn damage!", player.Name, effect.Potency), "damage"))
		case "regen":
			player.HitpointsRemaining += effect.Potency
			if player.HitpointsRemaining > player.HitpointsTotal {
				player.HitpointsRemaining = player.HitpointsTotal
			}
			msgs = append(msgs, Msg(fmt.Sprintf("%s regenerates %d HP!", player.Name, effect.Potency), "heal"))
		case "buff_attack":
			// already applied when effect was first added
		case "buff_defense":
			// already applied when effect was first added
		}
		effect.Duration--
		if effect.Duration <= 0 {
			switch effect.Type {
			case "buff_attack":
				player.StatsMod.AttackMod -= effect.Potency
			case "buff_defense":
				player.StatsMod.DefenseMod -= effect.Potency
			}
			msgs = append(msgs, Msg(fmt.Sprintf("%s's %s effect has worn off.", player.Name, effect.Type), "debuff"))
			player.StatusEffects = append(player.StatusEffects[:i], player.StatusEffects[i+1:]...)
		}
	}

	// =====================================================================
	// Process status effects for mob (inline)
	// =====================================================================
	for i := len(mob.StatusEffects) - 1; i >= 0; i-- {
		effect := &mob.StatusEffects[i]
		switch effect.Type {
		case "poison":
			mob.HitpointsRemaining -= effect.Potency
			msgs = append(msgs, Msg(fmt.Sprintf("%s takes %d poison damage!", mob.Name, effect.Potency), "damage"))
		case "burn":
			mob.HitpointsRemaining -= effect.Potency
			msgs = append(msgs, Msg(fmt.Sprintf("%s takes %d burn damage!", mob.Name, effect.Potency), "damage"))
		case "regen":
			mob.HitpointsRemaining += effect.Potency
			if mob.HitpointsRemaining > mob.HitpointsTotal {
				mob.HitpointsRemaining = mob.HitpointsTotal
			}
			msgs = append(msgs, Msg(fmt.Sprintf("%s regenerates %d HP!", mob.Name, effect.Potency), "heal"))
		case "buff_attack":
			// already applied
		case "buff_defense":
			// already applied
		}
		effect.Duration--
		if effect.Duration <= 0 {
			switch effect.Type {
			case "buff_attack":
				mob.StatsMod.AttackMod -= effect.Potency
			case "buff_defense":
				mob.StatsMod.DefenseMod -= effect.Potency
			}
			msgs = append(msgs, Msg(fmt.Sprintf("%s's %s effect has worn off.", mob.Name, effect.Type), "debuff"))
			mob.StatusEffects = append(mob.StatusEffects[:i], mob.StatusEffects[i+1:]...)
		}
	}

	// Check if mob died from effects
	if mob.HitpointsRemaining <= 0 {
		return e.resolveCombatWin(session, msgs)
	}

	// Check if player died from effects
	if player.HitpointsRemaining <= 0 {
		return e.resolveCombatLoss(session, msgs)
	}

	// =====================================================================
	// Check if player is stunned
	// =====================================================================
	if game.IsStunned(player) {
		msgs = append(msgs, Msg(fmt.Sprintf("%s is STUNNED and cannot act!", player.Name), "debuff"))
		playerDef := game.MultiRoll(player.DefenseRolls) + player.StatsMod.DefenseMod
		monsterMsgs := e.processMonsterTurnMsgs(session, playerDef)
		msgs = append(msgs, monsterMsgs...)

		if player.HitpointsRemaining <= 0 {
			return e.resolveCombatLoss(session, msgs)
		}

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

	// =====================================================================
	// Process player action
	// =====================================================================
	var playerDef int
	skipMonsterTurn := false

	switch cmd.Value {
	case "1": // Attack
		playerAttack := game.MultiRoll(player.AttackRolls) + player.StatsMod.AttackMod
		playerDef = game.MultiRoll(player.DefenseRolls) + player.StatsMod.DefenseMod
		isCrit := rand.Intn(100) < 15
		if isCrit {
			playerAttack *= 2
			msgs = append(msgs, Msg("*** CRITICAL HIT! ***", "combat"))
		}
		mobDef := game.MultiRoll(mob.DefenseRolls) + mob.StatsMod.DefenseMod
		if playerAttack > mobDef {
			diff := game.ApplyDamage(playerAttack-mobDef, models.Physical, mob)
			mob.HitpointsRemaining -= diff
			msgs = append(msgs, Msg(fmt.Sprintf("%s attacks for %d damage!", player.Name, diff), "damage"))
		} else {
			msgs = append(msgs, Msg(fmt.Sprintf("%s's attack missed!", player.Name), "combat"))
		}

	case "2": // Defend
		combat.IsDefending = true
		playerAttack := (game.MultiRoll(player.AttackRolls) + player.StatsMod.AttackMod) / 2
		playerDef = int(float64(game.MultiRoll(player.DefenseRolls)+player.StatsMod.DefenseMod) * 1.5)
		msgs = append(msgs, Msg(fmt.Sprintf("%s takes a defensive stance!", player.Name), "combat"))
		mobDef := game.MultiRoll(mob.DefenseRolls) + mob.StatsMod.DefenseMod
		if playerAttack > mobDef {
			diff := game.ApplyDamage(playerAttack-mobDef, models.Physical, mob)
			mob.HitpointsRemaining -= diff
			msgs = append(msgs, Msg(fmt.Sprintf("%s counterattacks for %d damage!", player.Name, diff), "damage"))
		} else {
			msgs = append(msgs, Msg(fmt.Sprintf("%s's counterattack missed!", player.Name), "combat"))
		}

	case "3": // Use Item - switch to item select
		consumables := []models.Item{}
		for _, item := range player.Inventory {
			if item.ItemType == "consumable" {
				consumables = append(consumables, item)
			}
		}
		if len(consumables) == 0 {
			msgs = append(msgs, Msg("No consumable items available!", "system"))
			// Undo turn increment since no action taken
			combat.Turn--
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
		session.State = StateCombatItemSelect
		// Undo turn increment - turn will be consumed when item is actually used
		combat.Turn--
		options := []MenuOption{}
		idx := 0
		for _, item := range player.Inventory {
			if item.ItemType == "consumable" {
				idx++
				label := fmt.Sprintf("%s (Heals %d HP)", item.Name, item.Consumable.Value)
				options = append(options, Opt(strconv.Itoa(idx), label))
			}
		}
		options = append(options, Opt("0", "Cancel"))
		return GameResponse{
			Type:     "combat",
			Messages: []GameMessage{Msg("Choose an item to use:", "system")},
			State: &StateData{
				Screen: "combat_item_select",
				Player: MakePlayerState(player),
				Combat: MakeCombatView(session),
			},
			Options: options,
		}

	case "4": // Use Skill - switch to skill select
		if len(player.LearnedSkills) == 0 {
			msgs = append(msgs, Msg("No skills learned!", "system"))
			combat.Turn--
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
		session.State = StateCombatSkillSelect
		combat.Turn--
		options := []MenuOption{}
		for idx, skill := range player.LearnedSkills {
			canAfford := skill.ManaCost <= player.ManaRemaining && skill.StaminaCost <= player.StaminaRemaining
			label := skill.Name
			if skill.ManaCost > 0 {
				label += fmt.Sprintf(" %dMP", skill.ManaCost)
			}
			if skill.StaminaCost > 0 {
				label += fmt.Sprintf(" %dSP", skill.StaminaCost)
			}
			label += " | " + skill.Description
			if canAfford {
				options = append(options, Opt(strconv.Itoa(idx+1), label))
			} else {
				options = append(options, OptDisabled(strconv.Itoa(idx+1), label+" [insufficient resources]"))
			}
		}
		options = append(options, Opt("0", "Cancel"))
		return GameResponse{
			Type:     "combat",
			Messages: []GameMessage{Msg("Choose a skill:", "system")},
			State: &StateData{
				Screen: "combat_skill_select",
				Player: MakePlayerState(player),
				Combat: MakeCombatView(session),
			},
			Options: options,
		}

	case "6": // Auto-fight: resolve rest of combat automatically
		combat.Turn-- // undo increment, autoResolveCombat manages its own turns
		return e.autoResolveCombat(session, msgs)

	case "5": // Flee
		fleeChance := 50 + (player.Level-mob.Level)*5
		if fleeChance > 90 {
			fleeChance = 90
		}
		if fleeChance < 20 {
			fleeChance = 20
		}
		roll := rand.Intn(100)
		if roll < fleeChance {
			combat.Fled = true
			msgs = append(msgs, Msg(fmt.Sprintf("%s successfully fled from combat!", player.Name), "system"))
			msgs = append(msgs, Msg("You escaped safely, but gained no rewards.", "system"))

			// Process guard recovery
			if session.GameState.Villages != nil {
				if village, exists := session.GameState.Villages[player.VillageName]; exists {
					game.ProcessGuardRecovery(&village)
					session.GameState.Villages[player.VillageName] = village
				}
			}

			// Check hunts remaining
			if combat.HuntsRemaining > 0 {
				return e.startNextHunt(session, msgs)
			}

			session.State = StateMainMenu
			return GameResponse{
				Type:     "narrative",
				Messages: msgs,
				State:    &StateData{Screen: "main_menu", Player: MakePlayerState(player)},
				Options:  BuildMainMenuResponse(session).Options,
			}
		}
		msgs = append(msgs, Msg(fmt.Sprintf("%s tried to flee but failed!", player.Name), "combat"))
		playerDef = game.MultiRoll(player.DefenseRolls) + player.StatsMod.DefenseMod
		// Skip to monster turn (failed flee = no player action)
		skipMonsterTurn = false

	default: // Invalid action, default to attack
		msgs = append(msgs, Msg("Invalid action! Defaulting to Attack.", "system"))
		playerAttack := game.MultiRoll(player.AttackRolls) + player.StatsMod.AttackMod
		playerDef = game.MultiRoll(player.DefenseRolls) + player.StatsMod.DefenseMod
		mobDef := game.MultiRoll(mob.DefenseRolls) + mob.StatsMod.DefenseMod
		if playerAttack > mobDef {
			diff := game.ApplyDamage(playerAttack-mobDef, models.Physical, mob)
			mob.HitpointsRemaining -= diff
			msgs = append(msgs, Msg(fmt.Sprintf("%s attacks for %d damage!", player.Name, diff), "damage"))
		} else {
			msgs = append(msgs, Msg(fmt.Sprintf("%s's attack missed!", player.Name), "combat"))
		}
	}

	if skipMonsterTurn {
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

	// Check if mob died from player attack
	if mob.HitpointsRemaining <= 0 {
		return e.resolveCombatWin(session, msgs)
	}

	// Guard attacks (if guards present and mob still alive)
	if combat.HasGuards && len(combat.CombatGuards) > 0 && mob.HitpointsRemaining > 0 {
		msgs = append(msgs, Msg("--- Guard Support ---", "system"))
		// Call game.GuardAttack for state mutation (it prints to stdout, which is fine)
		guardDamage := game.GuardAttack(combat.CombatGuards, mob)
		if guardDamage > 0 {
			msgs = append(msgs, Msg(fmt.Sprintf("Guards deal %d total damage!", guardDamage), "damage"))
		}
	}

	// Check if mob died from guard attacks
	if mob.HitpointsRemaining <= 0 {
		return e.resolveCombatWin(session, msgs)
	}

	// =====================================================================
	// Monster turn
	// =====================================================================
	monsterMsgs := e.processMonsterTurnMsgs(session, playerDef)
	msgs = append(msgs, monsterMsgs...)

	// Check if player died from monster attack
	if player.HitpointsRemaining <= 0 {
		return e.resolveCombatLoss(session, msgs)
	}

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

// handleCombatItemSelect processes the player's item selection during combat.
func (e *Engine) handleCombatItemSelect(session *GameSession, cmd GameCommand) GameResponse {
	combat := session.Combat
	player := session.Player
	mob := &combat.Mob
	msgs := []GameMessage{}

	// Build consumable list with original indices
	consumables := []models.Item{}
	consumableIndices := []int{}
	for idx, item := range player.Inventory {
		if item.ItemType == "consumable" {
			consumables = append(consumables, item)
			consumableIndices = append(consumableIndices, idx)
		}
	}

	itemIdx, err := strconv.Atoi(cmd.Value)
	if err != nil || itemIdx < 0 || itemIdx > len(consumables) || cmd.Value == "0" {
		// Cancel or invalid - return to combat without consuming turn
		session.State = StateCombat
		return GameResponse{
			Type:     "combat",
			Messages: []GameMessage{Msg("Cancelled.", "system")},
			State: &StateData{
				Screen: "combat",
				Player: MakePlayerState(player),
				Combat: MakeCombatView(session),
			},
			Options: combatActionOptions(),
		}
	}

	// Use the selected item
	selectedItem := consumables[itemIdx-1]
	originalIdx := consumableIndices[itemIdx-1]
	game.UseConsumableItem(selectedItem, player)
	game.RemoveItemFromInventory(&player.Inventory, originalIdx)

	// Consume the turn now
	combat.Turn++
	msgs = append(msgs, Msg(fmt.Sprintf("--- Turn %d ---", combat.Turn), "system"))
	msgs = append(msgs, Msg(fmt.Sprintf("%s uses %s! (Heals %d HP)", player.Name, selectedItem.Name, selectedItem.Consumable.Value), "heal"))

	session.State = StateCombat

	// Process status effects for player (inline)
	for i := len(player.StatusEffects) - 1; i >= 0; i-- {
		effect := &player.StatusEffects[i]
		switch effect.Type {
		case "poison":
			player.HitpointsRemaining -= effect.Potency
			msgs = append(msgs, Msg(fmt.Sprintf("%s takes %d poison damage!", player.Name, effect.Potency), "damage"))
		case "burn":
			player.HitpointsRemaining -= effect.Potency
			msgs = append(msgs, Msg(fmt.Sprintf("%s takes %d burn damage!", player.Name, effect.Potency), "damage"))
		case "regen":
			player.HitpointsRemaining += effect.Potency
			if player.HitpointsRemaining > player.HitpointsTotal {
				player.HitpointsRemaining = player.HitpointsTotal
			}
			msgs = append(msgs, Msg(fmt.Sprintf("%s regenerates %d HP!", player.Name, effect.Potency), "heal"))
		}
		effect.Duration--
		if effect.Duration <= 0 {
			switch effect.Type {
			case "buff_attack":
				player.StatsMod.AttackMod -= effect.Potency
			case "buff_defense":
				player.StatsMod.DefenseMod -= effect.Potency
			}
			msgs = append(msgs, Msg(fmt.Sprintf("%s's %s effect has worn off.", player.Name, effect.Type), "debuff"))
			player.StatusEffects = append(player.StatusEffects[:i], player.StatusEffects[i+1:]...)
		}
	}

	// Process status effects for mob (inline)
	for i := len(mob.StatusEffects) - 1; i >= 0; i-- {
		effect := &mob.StatusEffects[i]
		switch effect.Type {
		case "poison":
			mob.HitpointsRemaining -= effect.Potency
			msgs = append(msgs, Msg(fmt.Sprintf("%s takes %d poison damage!", mob.Name, effect.Potency), "damage"))
		case "burn":
			mob.HitpointsRemaining -= effect.Potency
			msgs = append(msgs, Msg(fmt.Sprintf("%s takes %d burn damage!", mob.Name, effect.Potency), "damage"))
		case "regen":
			mob.HitpointsRemaining += effect.Potency
			if mob.HitpointsRemaining > mob.HitpointsTotal {
				mob.HitpointsRemaining = mob.HitpointsTotal
			}
			msgs = append(msgs, Msg(fmt.Sprintf("%s regenerates %d HP!", mob.Name, effect.Potency), "heal"))
		}
		effect.Duration--
		if effect.Duration <= 0 {
			switch effect.Type {
			case "buff_attack":
				mob.StatsMod.AttackMod -= effect.Potency
			case "buff_defense":
				mob.StatsMod.DefenseMod -= effect.Potency
			}
			msgs = append(msgs, Msg(fmt.Sprintf("%s's %s effect has worn off.", mob.Name, effect.Type), "debuff"))
			mob.StatusEffects = append(mob.StatusEffects[:i], mob.StatusEffects[i+1:]...)
		}
	}

	// Check if mob died from effects
	if mob.HitpointsRemaining <= 0 {
		return e.resolveCombatWin(session, msgs)
	}

	// Check if player died from effects
	if player.HitpointsRemaining <= 0 {
		return e.resolveCombatLoss(session, msgs)
	}

	// Guard attacks
	if combat.HasGuards && len(combat.CombatGuards) > 0 && mob.HitpointsRemaining > 0 {
		msgs = append(msgs, Msg("--- Guard Support ---", "system"))
		guardDamage := game.GuardAttack(combat.CombatGuards, mob)
		if guardDamage > 0 {
			msgs = append(msgs, Msg(fmt.Sprintf("Guards deal %d total damage!", guardDamage), "damage"))
		}
	}

	// Check if mob died from guard attacks
	if mob.HitpointsRemaining <= 0 {
		return e.resolveCombatWin(session, msgs)
	}

	// Monster turn - player defense is normal (item usage doesn't boost defense)
	playerDef := game.MultiRoll(player.DefenseRolls) + player.StatsMod.DefenseMod
	monsterMsgs := e.processMonsterTurnMsgs(session, playerDef)
	msgs = append(msgs, monsterMsgs...)

	// Check if player died
	if player.HitpointsRemaining <= 0 {
		return e.resolveCombatLoss(session, msgs)
	}

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

// handleCombatSkillSelect processes the player's skill selection during combat.
func (e *Engine) handleCombatSkillSelect(session *GameSession, cmd GameCommand) GameResponse {
	combat := session.Combat
	player := session.Player
	mob := &combat.Mob
	msgs := []GameMessage{}

	skillIdx, err := strconv.Atoi(cmd.Value)
	if err != nil || skillIdx < 0 || skillIdx > len(player.LearnedSkills) || cmd.Value == "0" {
		// Cancel - return to combat without consuming turn
		session.State = StateCombat
		return GameResponse{
			Type:     "combat",
			Messages: []GameMessage{Msg("Cancelled.", "system")},
			State: &StateData{
				Screen: "combat",
				Player: MakePlayerState(player),
				Combat: MakeCombatView(session),
			},
			Options: combatActionOptions(),
		}
	}

	skill := player.LearnedSkills[skillIdx-1]

	// Check affordability
	if skill.ManaCost > player.ManaRemaining {
		msgs = append(msgs, Msg("Not enough mana!", "error"))
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
	if skill.StaminaCost > player.StaminaRemaining {
		msgs = append(msgs, Msg("Not enough stamina!", "error"))
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

	// Consume the turn
	combat.Turn++
	session.State = StateCombat

	msgs = append(msgs, Msg(fmt.Sprintf("--- Turn %d ---", combat.Turn), "system"))

	// Process status effects for player (inline)
	for i := len(player.StatusEffects) - 1; i >= 0; i-- {
		effect := &player.StatusEffects[i]
		switch effect.Type {
		case "poison":
			player.HitpointsRemaining -= effect.Potency
			msgs = append(msgs, Msg(fmt.Sprintf("%s takes %d poison damage!", player.Name, effect.Potency), "damage"))
		case "burn":
			player.HitpointsRemaining -= effect.Potency
			msgs = append(msgs, Msg(fmt.Sprintf("%s takes %d burn damage!", player.Name, effect.Potency), "damage"))
		case "regen":
			player.HitpointsRemaining += effect.Potency
			if player.HitpointsRemaining > player.HitpointsTotal {
				player.HitpointsRemaining = player.HitpointsTotal
			}
			msgs = append(msgs, Msg(fmt.Sprintf("%s regenerates %d HP!", player.Name, effect.Potency), "heal"))
		}
		effect.Duration--
		if effect.Duration <= 0 {
			switch effect.Type {
			case "buff_attack":
				player.StatsMod.AttackMod -= effect.Potency
			case "buff_defense":
				player.StatsMod.DefenseMod -= effect.Potency
			}
			msgs = append(msgs, Msg(fmt.Sprintf("%s's %s effect has worn off.", player.Name, effect.Type), "debuff"))
			player.StatusEffects = append(player.StatusEffects[:i], player.StatusEffects[i+1:]...)
		}
	}

	// Process status effects for mob (inline)
	for i := len(mob.StatusEffects) - 1; i >= 0; i-- {
		effect := &mob.StatusEffects[i]
		switch effect.Type {
		case "poison":
			mob.HitpointsRemaining -= effect.Potency
			msgs = append(msgs, Msg(fmt.Sprintf("%s takes %d poison damage!", mob.Name, effect.Potency), "damage"))
		case "burn":
			mob.HitpointsRemaining -= effect.Potency
			msgs = append(msgs, Msg(fmt.Sprintf("%s takes %d burn damage!", mob.Name, effect.Potency), "damage"))
		case "regen":
			mob.HitpointsRemaining += effect.Potency
			if mob.HitpointsRemaining > mob.HitpointsTotal {
				mob.HitpointsRemaining = mob.HitpointsTotal
			}
			msgs = append(msgs, Msg(fmt.Sprintf("%s regenerates %d HP!", mob.Name, effect.Potency), "heal"))
		}
		effect.Duration--
		if effect.Duration <= 0 {
			switch effect.Type {
			case "buff_attack":
				mob.StatsMod.AttackMod -= effect.Potency
			case "buff_defense":
				mob.StatsMod.DefenseMod -= effect.Potency
			}
			msgs = append(msgs, Msg(fmt.Sprintf("%s's %s effect has worn off.", mob.Name, effect.Type), "debuff"))
			mob.StatusEffects = append(mob.StatusEffects[:i], mob.StatusEffects[i+1:]...)
		}
	}

	// Check if mob died from effects
	if mob.HitpointsRemaining <= 0 {
		return e.resolveCombatWin(session, msgs)
	}

	// Check if player died from effects
	if player.HitpointsRemaining <= 0 {
		return e.resolveCombatLoss(session, msgs)
	}

	// Deduct skill costs
	player.ManaRemaining -= skill.ManaCost
	player.StaminaRemaining -= skill.StaminaCost
	msgs = append(msgs, Msg(fmt.Sprintf("%s uses %s!", player.Name, skill.Name), "combat"))

	// Apply skill effects
	if skill.Damage < 0 {
		// Healing skill
		healAmount := -skill.Damage
		player.HitpointsRemaining += healAmount
		if player.HitpointsRemaining > player.HitpointsTotal {
			player.HitpointsRemaining = player.HitpointsTotal
		}
		msgs = append(msgs, Msg(fmt.Sprintf("%s heals for %d HP!", player.Name, healAmount), "heal"))
	} else if skill.Damage > 0 {
		// Damage skill
		finalDamage := game.ApplyDamage(skill.Damage, skill.DamageType, mob)
		mob.HitpointsRemaining -= finalDamage
		if skill.DamageType != models.Physical {
			resistance := mob.Resistances[skill.DamageType]
			resistMsg := ""
			if resistance < 1.0 {
				resistMsg = " (resistant!)"
			} else if resistance > 1.0 {
				resistMsg = " (weak!)"
			}
			msgs = append(msgs, Msg(fmt.Sprintf("Deals %d %s damage%s", finalDamage, skill.DamageType, resistMsg), "damage"))
		} else {
			msgs = append(msgs, Msg(fmt.Sprintf("Deals %d damage!", finalDamage), "damage"))
		}
	}

	// Apply status effects from skill
	if skill.Effect.Type != "none" && skill.Effect.Duration > 0 {
		if skill.Effect.Type == "buff_attack" || skill.Effect.Type == "buff_defense" || skill.Effect.Type == "regen" {
			// Buff on player
			player.StatusEffects = append(player.StatusEffects, skill.Effect)
			if skill.Effect.Type == "buff_attack" {
				player.StatsMod.AttackMod += skill.Effect.Potency
			} else if skill.Effect.Type == "buff_defense" {
				player.StatsMod.DefenseMod += skill.Effect.Potency
			}
			msgs = append(msgs, Msg(fmt.Sprintf("%s gains %s effect!", player.Name, skill.Effect.Type), "buff"))
		} else {
			// Debuff/DoT on mob
			mob.StatusEffects = append(mob.StatusEffects, skill.Effect)
			msgs = append(msgs, Msg(fmt.Sprintf("%s is afflicted with %s!", mob.Name, skill.Effect.Type), "debuff"))
		}
	}

	// Check if mob died from skill damage
	if mob.HitpointsRemaining <= 0 {
		return e.resolveCombatWin(session, msgs)
	}

	// Guard attacks
	if combat.HasGuards && len(combat.CombatGuards) > 0 && mob.HitpointsRemaining > 0 {
		msgs = append(msgs, Msg("--- Guard Support ---", "system"))
		guardDamage := game.GuardAttack(combat.CombatGuards, mob)
		if guardDamage > 0 {
			msgs = append(msgs, Msg(fmt.Sprintf("Guards deal %d total damage!", guardDamage), "damage"))
		}
	}

	// Check if mob died from guard attacks
	if mob.HitpointsRemaining <= 0 {
		return e.resolveCombatWin(session, msgs)
	}

	// Monster turn
	playerDef := game.MultiRoll(player.DefenseRolls) + player.StatsMod.DefenseMod
	monsterMsgs := e.processMonsterTurnMsgs(session, playerDef)
	msgs = append(msgs, monsterMsgs...)

	// Check if player died
	if player.HitpointsRemaining <= 0 {
		return e.resolveCombatLoss(session, msgs)
	}

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

// handleCombatSkillReward processes the player's choice after defeating a skill guardian.
func (e *Engine) handleCombatSkillReward(session *GameSession, cmd GameCommand) GameResponse {
	combat := session.Combat
	player := session.Player
	msgs := []GameMessage{}

	guardedSkill := combat.Mob.GuardedSkill

	switch cmd.Value {
	case "1": // Learn skill immediately
		player.LearnedSkills = append(player.LearnedSkills, guardedSkill)
		msgs = append(msgs, Msg(fmt.Sprintf("You have learned %s!", guardedSkill.Name), "levelup"))
		msgs = append(msgs, Msg("You can now use this skill in combat.", "system"))
	case "2": // Take skill scroll
		scroll := game.CreateSkillScroll(guardedSkill)
		player.Inventory = append(player.Inventory, scroll)
		msgs = append(msgs, Msg(fmt.Sprintf("You received a %s!", scroll.Name), "loot"))
		msgs = append(msgs, Msg(fmt.Sprintf("Crafting Value: %d", scroll.SkillScroll.CraftingValue), "system"))
	default: // Default to scroll
		scroll := game.CreateSkillScroll(guardedSkill)
		player.Inventory = append(player.Inventory, scroll)
		msgs = append(msgs, Msg(fmt.Sprintf("You received a %s!", scroll.Name), "loot"))
	}

	// Check if more hunts remain
	if combat.HuntsRemaining > 0 {
		return e.startNextHunt(session, msgs)
	}

	session.State = StateMainMenu
	return GameResponse{
		Type:     "narrative",
		Messages: msgs,
		State:    &StateData{Screen: "main_menu", Player: MakePlayerState(player)},
		Options:  BuildMainMenuResponse(session).Options,
	}
}

// resolveCombatWin handles all post-victory logic: XP, loot, materials, potions, villagers, level ups.
func (e *Engine) resolveCombatWin(session *GameSession, msgs []GameMessage) GameResponse {
	combat := session.Combat
	player := session.Player
	mob := &combat.Mob
	combat.PlayerWon = true

	xpGained := scaledXP(player.Level, mob.Level)
	player.Experience += xpGained

	msgs = append(msgs, Msg("========================================", "system"))
	if xpGained > 0 {
		msgs = append(msgs, Msg(fmt.Sprintf("VICTORY! %s Wins! (+%d XP)", player.Name, xpGained), "combat"))
	} else {
		msgs = append(msgs, Msg(fmt.Sprintf("VICTORY! %s Wins! (No XP - enemy too weak)", player.Name), "combat"))
	}
	msgs = append(msgs, Msg("========================================", "system"))

	// Track autoplay stats
	if combat.IsAutoPlay {
		combat.AutoPlayWins++
		combat.AutoPlayXP += xpGained
	}

	// Loot enemy equipment
	for _, item := range mob.EquipmentMap {
		game.EquipBestItem(item, &player.EquipmentMap, &player.Inventory)
		msgs = append(msgs, Msg(fmt.Sprintf("Looted: %s", item.Name), "loot"))
	}

	// Drop beast materials
	materialName, materialQty := game.DropBeastMaterial(mob.MonsterType, player)
	if materialName != "" {
		msgs = append(msgs, Msg(fmt.Sprintf("Obtained %d %s!", materialQty, materialName), "loot"))
	}

	// 30% chance to get a health potion
	if rand.Intn(100) < 30 {
		potionSize := "small"
		roll := rand.Intn(100)
		if roll < 50 {
			potionSize = "small"
		} else if roll < 85 {
			potionSize = "medium"
		} else {
			potionSize = "large"
		}
		potion := game.CreateHealthPotion(potionSize)
		player.Inventory = append(player.Inventory, potion)
		msgs = append(msgs, Msg(fmt.Sprintf("Found a %s!", potion.Name), "loot"))
	}

	// 15% chance to rescue a villager
	if rand.Intn(100) < 15 {
		if session.GameState.Villages == nil {
			session.GameState.Villages = make(map[string]models.Village)
		}
		village, exists := session.GameState.Villages[player.VillageName]
		if !exists {
			village = game.GenerateVillage(player.Name)
			player.VillageName = player.Name + "'s Village"
		}
		// Call for side effects (prints to stdout)
		game.RescueVillager(&village)
		village.Experience += 25
		msgs = append(msgs, Msg("You rescued a villager! +25 Village XP", "narrative"))
		session.GameState.Villages[player.VillageName] = village
	}

	// Recalculate player stats from equipment
	player.StatsMod = game.CalculateItemMods(player.EquipmentMap)
	player.HitpointsTotal = player.HitpointsNatural + player.StatsMod.HitPointMod

	// Replace monster at location
	if combat.Location != nil && combat.MobLoc >= 0 && combat.MobLoc < len(combat.Location.Monsters) {
		combat.Location.Monsters[combat.MobLoc] = game.GenerateBestMonster(session.GameState, combat.Location.LevelMax, combat.Location.RarityMax)
	}

	// Update guard states in village after combat
	if combat.HasGuards && len(combat.CombatGuards) > 0 {
		if village, exists := session.GameState.Villages[player.VillageName]; exists {
			deadGuards := []string{}

			for i := len(village.ActiveGuards) - 1; i >= 0; i-- {
				guard := village.ActiveGuards[i]
				for _, combatGuard := range combat.CombatGuards {
					if guard.Name == combatGuard.Name {
						if mob.IsBoss && combatGuard.HitpointsRemaining <= 0 {
							// Permanent death in boss fights
							deadGuards = append(deadGuards, combatGuard.Name)
							village.ActiveGuards = append(village.ActiveGuards[:i], village.ActiveGuards[i+1:]...)
						} else {
							village.ActiveGuards[i].HitpointsRemaining = combatGuard.HitpointsRemaining
							village.ActiveGuards[i].Injured = combatGuard.Injured
							village.ActiveGuards[i].RecoveryTime = combatGuard.RecoveryTime
						}
						break
					}
				}
			}

			if len(deadGuards) > 0 {
				msgs = append(msgs, Msg("GUARDS FALLEN", "combat"))
				for _, guardName := range deadGuards {
					msgs = append(msgs, Msg(fmt.Sprintf("  %s has died in battle! (PERMANENT LOSS)", guardName), "combat"))
				}
			}

			session.GameState.Villages[player.VillageName] = village
		}
	}

	// Process guard recovery
	if session.GameState.Villages != nil {
		if village, exists := session.GameState.Villages[player.VillageName]; exists {
			game.ProcessGuardRecovery(&village)
			session.GameState.Villages[player.VillageName] = village
		}
	}

	// Level up - call game.LevelUp for state mutation side effects
	prevLevel := player.Level
	game.LevelUp(player)
	if player.Level > prevLevel {
		msgs = append(msgs, Msg(fmt.Sprintf("LEVEL UP! Now level %d!", player.Level), "levelup"))
		msgs = append(msgs, Msg(fmt.Sprintf("HP: %d, MP: %d, SP: %d", player.HitpointsTotal, player.ManaTotal, player.StaminaTotal), "levelup"))
	}

	// Level up the replaced monster
	if combat.Location != nil && combat.MobLoc >= 0 && combat.MobLoc < len(combat.Location.Monsters) {
		game.LevelUpMob(&combat.Location.Monsters[combat.MobLoc])
	}

	// Check quest progress for side effects
	game.CheckQuestProgress(player, session.GameState)

	// 15% chance to discover a new location during manual combat victory
	if combat.GuardianLocationName == "" && rand.Intn(100) < 15 {
		combinedSeen := append([]string{}, player.KnownLocations...)
		combinedSeen = append(combinedSeen, player.LockedLocations...)
		discovered := game.SearchLocation(combinedSeen, data.DiscoverableLocations)
		if discovered != "" {
			player.LockedLocations = append(player.LockedLocations, discovered)
			msgs = append(msgs, Msg(fmt.Sprintf("You discovered a new area: %s! A powerful guardian blocks the entrance.", discovered), "narrative"))
		}
	}

	// If location guardian, unlock location and return to main menu
	if combat.GuardianLocationName != "" {
		locName := combat.GuardianLocationName
		// Remove from LockedLocations
		for i, l := range player.LockedLocations {
			if l == locName {
				player.LockedLocations = append(player.LockedLocations[:i], player.LockedLocations[i+1:]...)
				break
			}
		}
		// Add to KnownLocations
		player.KnownLocations = append(player.KnownLocations, locName)
		msgs = append(msgs, Msg(fmt.Sprintf("The guardian has been slain! %s is now unlocked!", locName), "narrative"))

		session.GameState.CharactersMap[player.Name] = *player
		game.WriteGameStateToFile(*session.GameState, session.SaveFile)

		session.State = StateMainMenu
		return GameResponse{
			Type:     "narrative",
			Messages: msgs,
			State:    &StateData{Screen: "main_menu", Player: MakePlayerState(player)},
			Options:  BuildMainMenuResponse(session).Options,
		}
	}

	// PvP victory
	if combat.IsPvP {
		goldLooted := e.resolvePvPWin(session, msgs)
		if goldLooted > 0 {
			msgs = append(msgs, Msg(fmt.Sprintf("PvP Victory! You looted %d gold and items!", goldLooted), "loot"))
		} else {
			msgs = append(msgs, Msg("PvP Victory! You looted items from the sleeping player.", "loot"))
		}

		session.State = StateTownMain
		return GameResponse{
			Type:     "narrative",
			Messages: msgs,
			State:    &StateData{Screen: "town_main", Player: MakePlayerState(player)},
			Options:  []MenuOption{Opt("init", "Return to Town")},
		}
	}

	// Mayor challenge victory
	if combat.IsMayorChallenge {
		var done bool
		msgs, done = e.resolveMayorChallengeWin(session, msgs)
		if !done {
			// Advance to next phase
			town := session.SelectedTown
			if town == nil {
				t, loadErr := e.loadOrCreateTown(session)
				if loadErr == nil {
					town = t
				}
			}
			if town != nil {
				return e.startMayorChallengePhase(session, town, combat.MayorChallengePhase)
			}
		}

		session.State = StateTownMain
		return GameResponse{
			Type:     "narrative",
			Messages: msgs,
			State:    &StateData{Screen: "town_main", Player: MakePlayerState(player)},
			Options:  []MenuOption{Opt("init", "Return to Town")},
		}
	}

	// If skill guardian, offer reward choice
	if mob.IsSkillGuardian {
		msgs = append(msgs, Msg("SKILL GUARDIAN DEFEATED!", "narrative"))
		msgs = append(msgs, Msg(fmt.Sprintf("You defeated %s and can now learn: %s", mob.Name, mob.GuardedSkill.Name), "narrative"))
		msgs = append(msgs, Msg(fmt.Sprintf("Description: %s", mob.GuardedSkill.Description), "narrative"))
		msgs = append(msgs, Msg("Choose your reward:", "system"))

		session.State = StateCombatSkillReward
		return GameResponse{
			Type:     "combat",
			Messages: msgs,
			State: &StateData{
				Screen: "combat_skill_reward",
				Player: MakePlayerState(player),
				Combat: MakeCombatView(session),
			},
			Options: []MenuOption{
				Opt("1", "Absorb the skill immediately (learn now)"),
				Opt("2", "Take a skill scroll (can learn later or use for crafting)"),
			},
		}
	}

	// Check if more hunts remain
	if combat.HuntsRemaining > 0 {
		return e.startNextHunt(session, msgs)
	}

	session.State = StateMainMenu
	return GameResponse{
		Type:     "narrative",
		Messages: msgs,
		State:    &StateData{Screen: "main_menu", Player: MakePlayerState(player)},
		Options:  BuildMainMenuResponse(session).Options,
	}
}

// resolveCombatLoss handles post-defeat logic: equipment transfer, XP to mob, resurrection.
func (e *Engine) resolveCombatLoss(session *GameSession, msgs []GameMessage) GameResponse {
	combat := session.Combat
	player := session.Player
	mob := &combat.Mob

	msgs = append(msgs, Msg("========================================", "system"))
	msgs = append(msgs, Msg(fmt.Sprintf("DEFEAT! %s HAS DIED!", player.Name), "combat"))
	msgs = append(msgs, Msg(fmt.Sprintf("%s Wins!", mob.Name), "combat"))
	msgs = append(msgs, Msg("========================================", "system"))

	// Track autoplay stats
	if combat.IsAutoPlay {
		combat.AutoPlayDeaths++
	}

	// Transfer player equipment to mob
	for _, item := range player.EquipmentMap {
		game.EquipBestItem(item, &mob.EquipmentMap, &mob.Inventory)
	}

	// Give mob XP
	if combat.Location != nil && combat.MobLoc >= 0 && combat.MobLoc < len(combat.Location.Monsters) {
		combat.Location.Monsters[combat.MobLoc].StatsMod = game.CalculateItemMods(mob.EquipmentMap)
		combat.Location.Monsters[combat.MobLoc].Experience += player.Level * 100
	}

	// Process guard recovery
	if session.GameState.Villages != nil {
		if village, exists := session.GameState.Villages[player.VillageName]; exists {
			game.ProcessGuardRecovery(&village)
			session.GameState.Villages[player.VillageName] = village
		}
	}

	// PvP defeat
	if combat.IsPvP {
		e.resolvePvPLoss(session, msgs)
		player.HitpointsRemaining = player.HitpointsTotal
		player.ManaRemaining = player.ManaTotal
		player.StaminaRemaining = player.StaminaTotal
		player.Resurrections++
		player.StatusEffects = []models.StatusEffect{}
		msgs = append(msgs, Msg(fmt.Sprintf("%s has been resurrected. (Resurrection #%d)", player.Name, player.Resurrections), "system"))
		msgs = append(msgs, Msg("Your inn attack failed. The target's defenses held.", "narrative"))

		session.State = StateTownMain
		return GameResponse{
			Type:     "narrative",
			Messages: msgs,
			State:    &StateData{Screen: "town_main", Player: MakePlayerState(player)},
			Options:  []MenuOption{Opt("init", "Return to Town")},
		}
	}

	// Mayor challenge defeat
	if combat.IsMayorChallenge {
		e.resolveMayorChallengeLoss(session, msgs)
		player.HitpointsRemaining = player.HitpointsTotal
		player.ManaRemaining = player.ManaTotal
		player.StaminaRemaining = player.StaminaTotal
		player.Resurrections++
		player.StatusEffects = []models.StatusEffect{}
		msgs = append(msgs, Msg(fmt.Sprintf("%s has been resurrected. (Resurrection #%d)", player.Name, player.Resurrections), "system"))
		msgs = append(msgs, Msg("Your challenge against the mayor has failed.", "narrative"))

		session.State = StateTownMain
		return GameResponse{
			Type:     "narrative",
			Messages: msgs,
			State:    &StateData{Screen: "town_main", Player: MakePlayerState(player)},
			Options:  []MenuOption{Opt("init", "Return to Town")},
		}
	}

	// If location guardian, stay locked and return to main menu
	if combat.GuardianLocationName != "" {
		msgs = append(msgs, Msg(fmt.Sprintf("The guardian proved too powerful. %s remains locked.", combat.GuardianLocationName), "narrative"))

		player.HitpointsRemaining = player.HitpointsTotal
		player.ManaRemaining = player.ManaTotal
		player.StaminaRemaining = player.StaminaTotal
		player.Resurrections++
		player.StatusEffects = []models.StatusEffect{}
		msgs = append(msgs, Msg(fmt.Sprintf("%s has been resurrected. (Resurrection #%d)", player.Name, player.Resurrections), "system"))

		session.GameState.CharactersMap[player.Name] = *player
		session.State = StateMainMenu
		return GameResponse{
			Type:     "narrative",
			Messages: msgs,
			State:    &StateData{Screen: "main_menu", Player: MakePlayerState(player)},
			Options:  BuildMainMenuResponse(session).Options,
		}
	}

	// Check if more hunts remain
	if combat.HuntsRemaining > 0 {
		// Resurrect player
		player.HitpointsRemaining = player.HitpointsTotal
		player.ManaRemaining = player.ManaTotal
		player.StaminaRemaining = player.StaminaTotal
		player.Resurrections++
		player.StatusEffects = []models.StatusEffect{}
		msgs = append(msgs, Msg(fmt.Sprintf("%s HAS RESURRECTED! (Resurrection #%d)", player.Name, player.Resurrections), "system"))

		return e.startNextHunt(session, msgs)
	}

	// Resurrect for main menu return
	player.HitpointsRemaining = player.HitpointsTotal
	player.ManaRemaining = player.ManaTotal
	player.StaminaRemaining = player.StaminaTotal
	player.Resurrections++
	player.StatusEffects = []models.StatusEffect{}
	msgs = append(msgs, Msg(fmt.Sprintf("%s has been resurrected. (Resurrection #%d)", player.Name, player.Resurrections), "system"))

	session.State = StateMainMenu
	return GameResponse{
		Type:     "narrative",
		Messages: msgs,
		State:    &StateData{Screen: "main_menu", Player: MakePlayerState(player)},
		Options:  BuildMainMenuResponse(session).Options,
	}
}

// processMonsterTurnMsgs generates messages for the monster's turn and applies state changes.
func (e *Engine) processMonsterTurnMsgs(session *GameSession, playerDef int) []GameMessage {
	combat := session.Combat
	player := session.Player
	mob := &combat.Mob
	msgs := []GameMessage{}

	if mob.HitpointsRemaining <= 0 {
		return msgs
	}

	// Check if mob is stunned
	if game.IsStunnedMob(mob) {
		msgs = append(msgs, Msg(fmt.Sprintf("%s is STUNNED and cannot act!", mob.Name), "debuff"))
		return msgs
	}

	// 40% chance to use skill if mob has skills and resources
	usedSkill := false
	if len(mob.LearnedSkills) > 0 && rand.Intn(100) < 40 {
		skill := mob.LearnedSkills[rand.Intn(len(mob.LearnedSkills))]
		if skill.ManaCost <= mob.ManaRemaining && skill.StaminaCost <= mob.StaminaRemaining {
			mob.ManaRemaining -= skill.ManaCost
			mob.StaminaRemaining -= skill.StaminaCost
			usedSkill = true

			msgs = append(msgs, Msg(fmt.Sprintf("%s uses %s!", mob.Name, skill.Name), "combat"))

			if skill.Damage < 0 {
				// Healing
				healAmount := -skill.Damage
				mob.HitpointsRemaining += healAmount
				if mob.HitpointsRemaining > mob.HitpointsTotal {
					mob.HitpointsRemaining = mob.HitpointsTotal
				}
				msgs = append(msgs, Msg(fmt.Sprintf("%s heals for %d HP!", mob.Name, healAmount), "heal"))
			} else if skill.Damage > 0 {
				// Damage skill
				finalDamage := game.ApplyDamage(skill.Damage, skill.DamageType, player)

				// Guard defense if guards present
				if combat.HasGuards && len(combat.CombatGuards) > 0 {
					remainingDamage, _ := game.GuardDefense(combat.CombatGuards, finalDamage)
					absorbedDamage := finalDamage - remainingDamage
					if absorbedDamage > 0 {
						msgs = append(msgs, Msg(fmt.Sprintf("Guards absorbed %d of %d incoming damage!", absorbedDamage, finalDamage), "system"))
					}
					finalDamage = remainingDamage
				}

				player.HitpointsRemaining -= finalDamage
				msgs = append(msgs, Msg(fmt.Sprintf("Deals %d damage to %s!", finalDamage, player.Name), "damage"))
			}

			// Apply skill status effects
			if skill.Effect.Type != "none" && skill.Effect.Duration > 0 {
				if skill.Effect.Type == "buff_attack" || skill.Effect.Type == "buff_defense" {
					// Buff on mob
					mob.StatusEffects = append(mob.StatusEffects, skill.Effect)
					if skill.Effect.Type == "buff_attack" {
						mob.StatsMod.AttackMod += skill.Effect.Potency
					} else if skill.Effect.Type == "buff_defense" {
						mob.StatsMod.DefenseMod += skill.Effect.Potency
					}
					msgs = append(msgs, Msg(fmt.Sprintf("%s gains %s effect!", mob.Name, skill.Effect.Type), "buff"))
				} else {
					// Debuff/DoT on player
					player.StatusEffects = append(player.StatusEffects, skill.Effect)
					msgs = append(msgs, Msg(fmt.Sprintf("%s is afflicted with %s!", player.Name, skill.Effect.Type), "debuff"))
				}
			}
		}
	}

	// Normal attack if no skill used
	if !usedSkill {
		mobAttack := game.MultiRoll(mob.AttackRolls) + mob.StatsMod.AttackMod
		isCrit := rand.Intn(100) < 10
		if isCrit {
			mobAttack *= 2
			msgs = append(msgs, Msg(fmt.Sprintf("*** %s CRITICAL HIT! ***", mob.Name), "combat"))
		}

		if mobAttack > playerDef {
			diff := mobAttack - playerDef
			finalDamage := game.ApplyDamage(diff, models.Physical, player)

			// Guard defense if guards present
			if combat.HasGuards && len(combat.CombatGuards) > 0 {
				remainingDamage, _ := game.GuardDefense(combat.CombatGuards, finalDamage)
				absorbedDamage := finalDamage - remainingDamage
				if absorbedDamage > 0 {
					msgs = append(msgs, Msg(fmt.Sprintf("Guards absorbed %d of %d incoming damage!", absorbedDamage, finalDamage), "system"))
				}
				finalDamage = remainingDamage
			}

			player.HitpointsRemaining -= finalDamage
			msgs = append(msgs, Msg(fmt.Sprintf("%s attacks for %d damage!", mob.Name, finalDamage), "damage"))
		} else {
			msgs = append(msgs, Msg(fmt.Sprintf("%s's attack missed!", mob.Name), "combat"))
		}
	}

	return msgs
}

// startNextHunt begins the next hunt in a multi-hunt session.
func (e *Engine) startNextHunt(session *GameSession, msgs []GameMessage) GameResponse {
	combat := session.Combat
	player := session.Player
	location := combat.Location

	combat.HuntsRemaining--
	msgs = append(msgs, Msg(fmt.Sprintf("Hunts remaining: %d", combat.HuntsRemaining+1), "system"))

	// Resurrect if dead
	if player.HitpointsRemaining <= 0 {
		player.HitpointsRemaining = player.HitpointsTotal
		player.ManaRemaining = player.ManaTotal
		player.StaminaRemaining = player.StaminaTotal
		player.Resurrections++
		player.StatusEffects = []models.StatusEffect{}
		msgs = append(msgs, Msg(fmt.Sprintf("%s HAS RESURRECTED!", player.Name), "system"))
	}

	// Clear status effects for fresh fight
	player.StatusEffects = []models.StatusEffect{}

	// Pick a random mob from location
	mobLoc := rand.Intn(len(location.Monsters))
	mob := location.Monsters[mobLoc]

	msgs = append(msgs, Msg("========================================", "system"))
	msgs = append(msgs, Msg(fmt.Sprintf("Level %d %s vs Level %d %s (%s)",
		player.Level, player.Name, mob.Level, mob.Name, mob.MonsterType), "system"))
	msgs = append(msgs, Msg("========================================", "system"))

	// Restore mana and stamina for new fight
	player.ManaRemaining = player.ManaTotal
	player.StaminaRemaining = player.StaminaTotal
	mob.ManaRemaining = mob.ManaTotal
	mob.StaminaRemaining = mob.StaminaTotal

	// Preserve hunt tracking and autoplay fields from previous combat
	guardianLocationName := combat.GuardianLocationName
	isAutoPlay := combat.IsAutoPlay
	autoPlaySpeed := combat.AutoPlaySpeed
	autoPlayFights := combat.AutoPlayFights + 1
	autoPlayWins := combat.AutoPlayWins
	autoPlayDeaths := combat.AutoPlayDeaths
	autoPlayXP := combat.AutoPlayXP
	huntsRemaining := combat.HuntsRemaining
	combatGuards := combat.CombatGuards
	hasGuards := combat.HasGuards

	// Set up new combat context
	session.Combat = &CombatContext{
		Mob:                  mob,
		MobLoc:               mobLoc,
		Location:             location,
		Turn:                 0,
		Fled:                 false,
		PlayerWon:            false,
		IsDefending:          false,
		GuardianLocationName: guardianLocationName,
		CombatGuards:         combatGuards,
		HasGuards:            hasGuards,
		HuntsRemaining:       huntsRemaining,
		IsAutoPlay:           isAutoPlay,
		AutoPlaySpeed:        autoPlaySpeed,
		AutoPlayFights:       autoPlayFights,
		AutoPlayWins:         autoPlayWins,
		AutoPlayDeaths:       autoPlayDeaths,
		AutoPlayXP:           autoPlayXP,
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

// scaledXP calculates XP gained from defeating a monster based on level difference.
// Monsters 10+ levels below the player grant 0 XP. Equal or higher level grants full XP.
// In between, XP scales linearly.
func scaledXP(playerLevel, mobLevel int) int {
	diff := playerLevel - mobLevel // positive means player is higher
	if diff >= 10 {
		return 0
	}
	baseXP := mobLevel * 10
	if diff <= 0 {
		// Monster is equal or higher level: full XP + bonus
		bonus := (-diff) * 5
		return baseXP + bonus
	}
	// Monster is 1-9 levels below: scale down linearly (90% at -1, 80% at -2, ... 10% at -9)
	pct := 100 - (diff * 10)
	return baseXP * pct / 100
}

// combatActionOptions returns the standard set of combat action menu options.
func combatActionOptions() []MenuOption {
	return []MenuOption{
		Opt("1", "Attack"),
		Opt("2", "Defend"),
		Opt("3", "Use Item"),
		Opt("4", "Use Skill"),
		Opt("5", "Flee"),
		Opt("6", "Auto Fight"),
	}
}

// autoResolveCombat runs the current fight to completion using AI for the player.
func (e *Engine) autoResolveCombat(session *GameSession, msgs []GameMessage) GameResponse {
	combat := session.Combat
	player := session.Player
	mob := &combat.Mob

	msgs = append(msgs, Msg("--- AUTO FIGHT ---", "system"))

	for player.HitpointsRemaining > 0 && mob.HitpointsRemaining > 0 {
		combat.Turn++

		// Safety valve
		if combat.Turn > 200 {
			msgs = append(msgs, Msg("Combat timed out!", "combat"))
			break
		}

		// Process player status effects
		for i := len(player.StatusEffects) - 1; i >= 0; i-- {
			effect := &player.StatusEffects[i]
			switch effect.Type {
			case "poison":
				player.HitpointsRemaining -= effect.Potency
				msgs = append(msgs, Msg(fmt.Sprintf("%s takes %d poison damage!", player.Name, effect.Potency), "damage"))
			case "burn":
				player.HitpointsRemaining -= effect.Potency
				msgs = append(msgs, Msg(fmt.Sprintf("%s takes %d burn damage!", player.Name, effect.Potency), "damage"))
			case "regen":
				player.HitpointsRemaining += effect.Potency
				if player.HitpointsRemaining > player.HitpointsTotal {
					player.HitpointsRemaining = player.HitpointsTotal
				}
			}
			effect.Duration--
			if effect.Duration <= 0 {
				switch effect.Type {
				case "buff_attack":
					player.StatsMod.AttackMod -= effect.Potency
				case "buff_defense":
					player.StatsMod.DefenseMod -= effect.Potency
				}
				player.StatusEffects = append(player.StatusEffects[:i], player.StatusEffects[i+1:]...)
			}
		}

		// Process mob status effects
		for i := len(mob.StatusEffects) - 1; i >= 0; i-- {
			effect := &mob.StatusEffects[i]
			switch effect.Type {
			case "poison":
				mob.HitpointsRemaining -= effect.Potency
			case "burn":
				mob.HitpointsRemaining -= effect.Potency
			case "regen":
				mob.HitpointsRemaining += effect.Potency
				if mob.HitpointsRemaining > mob.HitpointsTotal {
					mob.HitpointsRemaining = mob.HitpointsTotal
				}
			}
			effect.Duration--
			if effect.Duration <= 0 {
				switch effect.Type {
				case "buff_attack":
					mob.StatsMod.AttackMod -= effect.Potency
				case "buff_defense":
					mob.StatsMod.DefenseMod -= effect.Potency
				}
				mob.StatusEffects = append(mob.StatusEffects[:i], mob.StatusEffects[i+1:]...)
			}
		}

		if player.HitpointsRemaining <= 0 || mob.HitpointsRemaining <= 0 {
			break
		}

		// Player AI turn (skip if stunned)
		if !game.IsStunned(player) {
			decision := game.MakeAIDecision(player, mob, combat.Turn)

			switch decision {
			case "attack":
				playerAttack := game.MultiRoll(player.AttackRolls) + player.StatsMod.AttackMod
				if rand.Intn(100) < 15 {
					playerAttack *= 2
					msgs = append(msgs, Msg("*** CRITICAL HIT! ***", "combat"))
				}
				mobDef := game.MultiRoll(mob.DefenseRolls) + mob.StatsMod.DefenseMod
				if playerAttack > mobDef {
					diff := game.ApplyDamage(playerAttack-mobDef, models.Physical, mob)
					mob.HitpointsRemaining -= diff
					msgs = append(msgs, Msg(fmt.Sprintf("%s attacks for %d damage!", player.Name, diff), "damage"))
				} else {
					msgs = append(msgs, Msg(fmt.Sprintf("%s's attack missed!", player.Name), "combat"))
				}
			case "item":
				for idx, item := range player.Inventory {
					if item.ItemType == "consumable" {
						game.UseConsumableItem(item, player)
						game.RemoveItemFromInventory(&player.Inventory, idx)
						msgs = append(msgs, Msg(fmt.Sprintf("%s uses %s!", player.Name, item.Name), "heal"))
						break
					}
				}
			default:
				if len(decision) > 6 && decision[:6] == "skill_" {
					skillName := decision[6:]
					for _, skill := range player.LearnedSkills {
						if skill.Name == skillName && skill.ManaCost <= player.ManaRemaining && skill.StaminaCost <= player.StaminaRemaining {
							player.ManaRemaining -= skill.ManaCost
							player.StaminaRemaining -= skill.StaminaCost
							if skill.Damage < 0 {
								player.HitpointsRemaining += -skill.Damage
								if player.HitpointsRemaining > player.HitpointsTotal {
									player.HitpointsRemaining = player.HitpointsTotal
								}
								msgs = append(msgs, Msg(fmt.Sprintf("%s uses %s! Heals %d HP!", player.Name, skill.Name, -skill.Damage), "heal"))
							} else if skill.Damage > 0 {
								finalDamage := game.ApplyDamage(skill.Damage, skill.DamageType, mob)
								mob.HitpointsRemaining -= finalDamage
								msgs = append(msgs, Msg(fmt.Sprintf("%s uses %s for %d damage!", player.Name, skill.Name, finalDamage), "damage"))
							}
							if skill.Effect.Type != "none" && skill.Effect.Duration > 0 {
								if skill.Effect.Type == "buff_attack" || skill.Effect.Type == "buff_defense" || skill.Effect.Type == "regen" {
									player.StatusEffects = append(player.StatusEffects, skill.Effect)
									if skill.Effect.Type == "buff_attack" {
										player.StatsMod.AttackMod += skill.Effect.Potency
									} else if skill.Effect.Type == "buff_defense" {
										player.StatsMod.DefenseMod += skill.Effect.Potency
									}
								} else {
									mob.StatusEffects = append(mob.StatusEffects, skill.Effect)
								}
							}
							break
						}
					}
				}
			}
		} else {
			msgs = append(msgs, Msg(fmt.Sprintf("%s is STUNNED!", player.Name), "debuff"))
		}

		if mob.HitpointsRemaining <= 0 || player.HitpointsRemaining <= 0 {
			break
		}

		// Monster turn (skip if stunned)
		if !game.IsStunnedMob(mob) {
			useSkill := len(mob.LearnedSkills) > 0 && rand.Intn(100) < 40
			if useSkill {
				skill := mob.LearnedSkills[rand.Intn(len(mob.LearnedSkills))]
				if skill.ManaCost <= mob.ManaRemaining && skill.StaminaCost <= mob.StaminaRemaining {
					mob.ManaRemaining -= skill.ManaCost
					mob.StaminaRemaining -= skill.StaminaCost
					if skill.Damage < 0 {
						mob.HitpointsRemaining += -skill.Damage
						if mob.HitpointsRemaining > mob.HitpointsTotal {
							mob.HitpointsRemaining = mob.HitpointsTotal
						}
					} else if skill.Damage > 0 {
						playerDef := game.MultiRoll(player.DefenseRolls) + player.StatsMod.DefenseMod
						finalDamage := game.ApplyDamage(skill.Damage, skill.DamageType, player)
						if finalDamage > playerDef {
							player.HitpointsRemaining -= (finalDamage - playerDef)
						}
					}
					if skill.Effect.Type != "none" && skill.Effect.Duration > 0 {
						if skill.Effect.Type == "buff_attack" || skill.Effect.Type == "buff_defense" || skill.Effect.Type == "regen" {
							mob.StatusEffects = append(mob.StatusEffects, skill.Effect)
						} else {
							player.StatusEffects = append(player.StatusEffects, skill.Effect)
						}
					}
					msgs = append(msgs, Msg(fmt.Sprintf("%s uses %s!", mob.Name, skill.Name), "combat"))
					continue
				}
			}
			// Normal attack
			mobAttack := game.MultiRoll(mob.AttackRolls) + mob.StatsMod.AttackMod
			if rand.Intn(100) < 10 {
				mobAttack *= 2
			}
			playerDef := game.MultiRoll(player.DefenseRolls) + player.StatsMod.DefenseMod
			if mobAttack > playerDef {
				diff := game.ApplyDamage(mobAttack-playerDef, models.Physical, player)
				player.HitpointsRemaining -= diff
				msgs = append(msgs, Msg(fmt.Sprintf("%s attacks for %d damage!", mob.Name, diff), "damage"))
			}
		}
	}

	msgs = append(msgs, Msg(fmt.Sprintf("--- Auto fight ended (Turn %d) ---", combat.Turn), "system"))

	// Resolve outcome
	if mob.HitpointsRemaining <= 0 {
		return e.resolveCombatWin(session, msgs)
	}
	return e.resolveCombatLoss(session, msgs)
}
