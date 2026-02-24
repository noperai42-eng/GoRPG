package engine

import "rpg-game/pkg/models"

// GameCommand represents a player action sent to the engine.
type GameCommand struct {
	Type  string `json:"type"`  // "select", "input", "init"
	Value string `json:"value"` // option key or free text input
}

// GameResponse is returned by the engine for every processed command.
type GameResponse struct {
	Type     string        `json:"type"`     // "menu", "combat", "narrative", "error", "exit"
	Messages []GameMessage `json:"messages"` // ordered text for display
	State    *StateData    `json:"state"`    // current game state for rendering
	Options  []MenuOption  `json:"options"`  // available actions
	Prompt   string        `json:"prompt"`   // prompt text for free-text input (empty means use options)
}

// GameMessage is a single display message with category metadata.
type GameMessage struct {
	Text     string `json:"text"`
	Category string `json:"category"` // system, combat, damage, heal, loot, buff, debuff, narrative, error, levelup
}

// MenuOption represents a selectable action the player can take.
type MenuOption struct {
	Key     string `json:"key"`
	Label   string `json:"label"`
	Enabled bool   `json:"enabled"`
}

// StateData contains current visible game state for UI rendering.
type StateData struct {
	Screen string       `json:"screen"`
	Player *PlayerState `json:"player,omitempty"`
	Combat *CombatView  `json:"combat,omitempty"`
}

// PlayerState shows visible player information.
type PlayerState struct {
	Name          string `json:"name"`
	Level         int    `json:"level"`
	Experience    int    `json:"experience"`
	HP            int    `json:"hp"`
	MaxHP         int    `json:"max_hp"`
	MP            int    `json:"mp"`
	MaxMP         int    `json:"max_mp"`
	SP            int    `json:"sp"`
	MaxSP         int    `json:"max_sp"`
	AttackRolls   int    `json:"attack_rolls"`
	DefenseRolls  int    `json:"defense_rolls"`
	AttackMod     int    `json:"attack_mod"`
	DefenseMod    int    `json:"defense_mod"`
	HitPointMod   int    `json:"hitpoint_mod"`
	Resurrections int    `json:"resurrections"`
}

// CombatView contains combat-specific rendering state.
type CombatView struct {
	Turn           int          `json:"turn"`
	PlayerHP       int          `json:"player_hp"`
	PlayerMaxHP    int          `json:"player_max_hp"`
	PlayerMP       int          `json:"player_mp"`
	PlayerMaxMP    int          `json:"player_max_mp"`
	PlayerSP       int          `json:"player_sp"`
	PlayerMaxSP    int          `json:"player_max_sp"`
	PlayerEffects  []EffectView `json:"player_effects"`
	MonsterName    string       `json:"monster_name"`
	MonsterType    string       `json:"monster_type"`
	MonsterLevel   int          `json:"monster_level"`
	MonsterHP      int          `json:"monster_hp"`
	MonsterMaxHP   int          `json:"monster_max_hp"`
	MonsterMP      int          `json:"monster_mp"`
	MonsterMaxMP   int          `json:"monster_max_mp"`
	MonsterSP      int          `json:"monster_sp"`
	MonsterMaxSP   int          `json:"monster_max_sp"`
	MonsterEffects []EffectView `json:"monster_effects"`
	Guards         []GuardView  `json:"guards,omitempty"`
}

// EffectView shows a status effect for display.
type EffectView struct {
	Name     string `json:"name"`
	Duration int    `json:"duration"`
	Type     string `json:"type"` // buff, debuff, dot
}

// GuardView shows guard status in combat.
type GuardView struct {
	Name    string `json:"name"`
	HP      int    `json:"hp"`
	MaxHP   int    `json:"max_hp"`
	Injured bool   `json:"injured"`
}

// Helper constructors

func Msg(text string, category string) GameMessage {
	return GameMessage{Text: text, Category: category}
}

func Opt(key, label string) MenuOption {
	return MenuOption{Key: key, Label: label, Enabled: true}
}

func OptDisabled(key, label string) MenuOption {
	return MenuOption{Key: key, Label: label, Enabled: false}
}

func ErrorResponse(text string) GameResponse {
	return GameResponse{
		Type:     "error",
		Messages: []GameMessage{{Text: text, Category: "error"}},
	}
}

func MakePlayerState(p *models.Character) *PlayerState {
	if p == nil {
		return nil
	}
	return &PlayerState{
		Name:          p.Name,
		Level:         p.Level,
		Experience:    p.Experience,
		HP:            p.HitpointsRemaining,
		MaxHP:         p.HitpointsTotal,
		MP:            p.ManaRemaining,
		MaxMP:         p.ManaTotal,
		SP:            p.StaminaRemaining,
		MaxSP:         p.StaminaTotal,
		AttackRolls:   p.AttackRolls,
		DefenseRolls:  p.DefenseRolls,
		AttackMod:     p.StatsMod.AttackMod,
		DefenseMod:    p.StatsMod.DefenseMod,
		HitPointMod:   p.StatsMod.HitPointMod,
		Resurrections: p.Resurrections,
	}
}

func MakeCombatView(session *GameSession) *CombatView {
	if session.Combat == nil {
		return nil
	}
	c := session.Combat
	p := session.Player
	m := &c.Mob

	view := &CombatView{
		Turn:         c.Turn,
		PlayerHP:     p.HitpointsRemaining,
		PlayerMaxHP:  p.HitpointsTotal,
		PlayerMP:     p.ManaRemaining,
		PlayerMaxMP:  p.ManaTotal,
		PlayerSP:     p.StaminaRemaining,
		PlayerMaxSP:  p.StaminaTotal,
		MonsterName:  m.Name,
		MonsterType:  m.MonsterType,
		MonsterLevel: m.Level,
		MonsterHP:    m.HitpointsRemaining,
		MonsterMaxHP: m.HitpointsTotal,
		MonsterMP:    m.ManaRemaining,
		MonsterMaxMP: m.ManaTotal,
		MonsterSP:    m.StaminaRemaining,
		MonsterMaxSP: m.StaminaTotal,
	}

	for _, eff := range p.StatusEffects {
		category := "debuff"
		if eff.Type == "buff_attack" || eff.Type == "buff_defense" || eff.Type == "regen" {
			category = "buff"
		}
		view.PlayerEffects = append(view.PlayerEffects, EffectView{
			Name: eff.Type, Duration: eff.Duration, Type: category,
		})
	}

	for _, eff := range m.StatusEffects {
		category := "debuff"
		if eff.Type == "buff_attack" || eff.Type == "buff_defense" || eff.Type == "regen" {
			category = "buff"
		}
		view.MonsterEffects = append(view.MonsterEffects, EffectView{
			Name: eff.Type, Duration: eff.Duration, Type: category,
		})
	}

	for _, g := range c.CombatGuards {
		view.Guards = append(view.Guards, GuardView{
			Name: g.Name, HP: g.HitpointsRemaining, MaxHP: g.HitPoints, Injured: g.Injured,
		})
	}

	return view
}
