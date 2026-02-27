package engine

import (
	"fmt"

	"rpg-game/pkg/data"
	"rpg-game/pkg/game"
	"rpg-game/pkg/models"
)

// SlotNames maps equipment slot indices to human-readable names.
var SlotNames = map[int]string{
	0: "Head", 1: "Chest", 2: "Legs", 3: "Feet",
	4: "Hands", 5: "Main Hand", 6: "Off Hand", 7: "Accessory",
}

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
	Screen  string       `json:"screen"`
	Player  *PlayerState `json:"player,omitempty"`
	Combat  *CombatView  `json:"combat,omitempty"`
	Village *VillageView `json:"village,omitempty"`
	Town    *TownView    `json:"town,omitempty"`
}

// --- View structs for enriched frontend state ---

// ItemView represents an item for the frontend.
type ItemView struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	SlotName  string `json:"slot_name"`
	Slot      int    `json:"slot"`
	Rarity    int    `json:"rarity"`
	CP        int    `json:"cp"`
	Attack    int    `json:"attack"`
	Defense   int    `json:"defense"`
	HitPoint  int    `json:"hitpoint"`
	HealValue int    `json:"heal_value,omitempty"`
	SkillName string `json:"skill_name,omitempty"`
}

// SkillView represents a skill for the frontend.
type SkillView struct {
	Name        string `json:"name"`
	ManaCost    int    `json:"mana_cost"`
	StaminaCost int    `json:"stamina_cost"`
	Damage      int    `json:"damage"`
	DamageType  string `json:"damage_type"`
	Effect      string `json:"effect"`
	Duration    int    `json:"duration,omitempty"`
	Description string `json:"description"`
}

// LocationView represents a location for the frontend.
type LocationView struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	LevelMax  int    `json:"level_max"`
	RarityMax int    `json:"rarity_max"`
	Locked    bool   `json:"locked"`
}

// QuestView represents a quest for the frontend.
type QuestView struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Progress    int    `json:"progress"`
	Target      int    `json:"target"`
	RewardXP    int    `json:"reward_xp"`
	Completed   bool   `json:"completed"`
}

// VillageView represents village state for the frontend.
type VillageView struct {
	Name             string              `json:"name"`
	Level            int                 `json:"level"`
	Experience       int                 `json:"experience"`
	ExpToLevel       int                 `json:"exp_to_level"`
	Villagers        []VillagerView      `json:"villagers"`
	Guards           []VillageGuardView  `json:"guards"`
	Defenses         []DefenseView       `json:"defenses"`
	Traps            []TrapView          `json:"traps"`
	DefenseLevel     int                 `json:"defense_level"`
	UnlockedCrafting []string            `json:"unlocked_crafting"`
	ResourcePerTick  map[string]int      `json:"resource_per_tick"`
	LastHarvestTime  int64               `json:"last_harvest_time"`
}

// VillagerView represents a villager for the frontend.
type VillagerView struct {
	Name         string `json:"name"`
	Role         string `json:"role"`
	Level        int    `json:"level"`
	Efficiency   int    `json:"efficiency"`
	AssignedTask string `json:"assigned_task"`
	HarvestType  string `json:"harvest_type"`
}

// VillageGuardView represents a guard for the frontend.
type VillageGuardView struct {
	Name         string `json:"name"`
	Level        int    `json:"level"`
	HP           int    `json:"hp"`
	MaxHP        int    `json:"max_hp"`
	AttackBonus  int    `json:"attack_bonus"`
	DefenseBonus int    `json:"defense_bonus"`
	Hired        bool   `json:"hired"`
	Cost         int    `json:"cost"`
	Injured      bool   `json:"injured"`
	RecoveryTime int    `json:"recovery_time"`
}

// DefenseView represents a defense structure for the frontend.
type DefenseView struct {
	Name  string `json:"name"`
	Level int    `json:"level"`
	Type  string `json:"type"`
	Built bool   `json:"built"`
}

// TrapView represents a trap for the frontend.
type TrapView struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	Damage    int    `json:"damage"`
	Remaining int    `json:"remaining"`
}

// PlayerState shows visible player information.
type PlayerState struct {
	Name          string `json:"name"`
	Level         int    `json:"level"`
	Experience    int    `json:"experience"`
	ExpToLevel    int    `json:"exp_to_level"`
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

	Inventory       []ItemView          `json:"inventory"`
	Equipment       map[string]ItemView `json:"equipment"`
	Skills          []SkillView         `json:"skills"`
	Resources       map[string]int      `json:"resources"`
	KnownLocations  []LocationView      `json:"known_locations"`
	LockedLocations []LocationView      `json:"locked_locations"`
	ActiveQuests    []QuestView         `json:"active_quests"`
	CompletedQuests []QuestView         `json:"completed_quests"`
	VillageName     string              `json:"village_name"`
	Buildings       []string            `json:"buildings"`
}

// CombatView contains combat-specific rendering state.
type CombatView struct {
	Turn              int          `json:"turn"`
	PlayerHP          int          `json:"player_hp"`
	PlayerMaxHP       int          `json:"player_max_hp"`
	PlayerMP          int          `json:"player_mp"`
	PlayerMaxMP       int          `json:"player_max_mp"`
	PlayerSP          int          `json:"player_sp"`
	PlayerMaxSP       int          `json:"player_max_sp"`
	PlayerEffects     []EffectView `json:"player_effects"`
	MonsterName       string       `json:"monster_name"`
	MonsterType       string       `json:"monster_type"`
	MonsterLevel      int          `json:"monster_level"`
	MonsterHP         int          `json:"monster_hp"`
	MonsterMaxHP      int          `json:"monster_max_hp"`
	MonsterMP         int          `json:"monster_mp"`
	MonsterMaxMP      int          `json:"monster_max_mp"`
	MonsterSP         int          `json:"monster_sp"`
	MonsterMaxSP      int          `json:"monster_max_sp"`
	MonsterEffects    []EffectView `json:"monster_effects"`
	MonsterIsBoss     bool         `json:"monster_is_boss"`
	MonsterIsGuardian bool         `json:"monster_is_guardian"`
	GuardedSkillName  string       `json:"guarded_skill_name,omitempty"`
	Guards            []GuardView  `json:"guards,omitempty"`
	HuntsRemaining    int          `json:"hunts_remaining"`
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

// makeItemView converts a models.Item to an ItemView for the frontend.
func makeItemView(item models.Item) ItemView {
	slotName := ""
	if item.ItemType == "equipment" {
		slotName = SlotNames[item.Slot]
		if slotName == "" {
			slotName = fmt.Sprintf("Slot %d", item.Slot)
		}
	}
	v := ItemView{
		Name:     item.Name,
		Type:     item.ItemType,
		SlotName: slotName,
		Slot:     item.Slot,
		Rarity:   item.Rarity,
		CP:       item.CP,
		Attack:   item.StatsMod.AttackMod,
		Defense:  item.StatsMod.DefenseMod,
		HitPoint: item.StatsMod.HitPointMod,
	}
	if item.ItemType == "consumable" {
		v.HealValue = item.Consumable.Value
	}
	if item.ItemType == "skill_scroll" {
		v.SkillName = item.SkillScroll.Skill.Name
	}
	return v
}

// makeSkillView converts a models.Skill to a SkillView for the frontend.
func makeSkillView(skill models.Skill) SkillView {
	effectStr := ""
	if skill.Effect.Type != "" && skill.Effect.Type != "none" {
		effectStr = skill.Effect.Type
	}
	return SkillView{
		Name:        skill.Name,
		ManaCost:    skill.ManaCost,
		StaminaCost: skill.StaminaCost,
		Damage:      skill.Damage,
		DamageType:  string(skill.DamageType),
		Effect:      effectStr,
		Duration:    skill.Effect.Duration,
		Description: skill.Description,
	}
}

func MakePlayerState(p *models.Character) *PlayerState {
	if p == nil {
		return nil
	}
	ps := &PlayerState{
		Name:          p.Name,
		Level:         p.Level,
		Experience:    p.Experience,
		ExpToLevel:    game.PlayerExpToLevel(p.Level),
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
		VillageName:   p.VillageName,
	}

	// Inventory
	ps.Inventory = make([]ItemView, 0, len(p.Inventory))
	for _, item := range p.Inventory {
		ps.Inventory = append(ps.Inventory, makeItemView(item))
	}

	// Equipment
	ps.Equipment = make(map[string]ItemView)
	for slot, item := range p.EquipmentMap {
		slotName := SlotNames[slot]
		if slotName == "" {
			slotName = fmt.Sprintf("Slot %d", slot)
		}
		ps.Equipment[slotName] = makeItemView(item)
	}

	// Skills
	ps.Skills = make([]SkillView, 0, len(p.LearnedSkills))
	for _, skill := range p.LearnedSkills {
		ps.Skills = append(ps.Skills, makeSkillView(skill))
	}

	// Resources
	ps.Resources = make(map[string]int)
	for _, resName := range data.ResourceTypes {
		if r, exists := p.ResourceStorageMap[resName]; exists {
			ps.Resources[resName] = r.Stock
		} else {
			ps.Resources[resName] = 0
		}
	}
	for _, matName := range data.BeastMaterials {
		if r, exists := p.ResourceStorageMap[matName]; exists && r.Stock > 0 {
			ps.Resources[matName] = r.Stock
		}
	}

	// Known locations
	ps.KnownLocations = make([]LocationView, 0)
	for _, locName := range p.KnownLocations {
		lv := LocationView{Name: locName, Locked: false}
		ps.KnownLocations = append(ps.KnownLocations, lv)
	}

	// Locked locations
	ps.LockedLocations = make([]LocationView, 0)
	for _, locName := range p.LockedLocations {
		lv := LocationView{Name: locName, Locked: true}
		ps.LockedLocations = append(ps.LockedLocations, lv)
	}

	// Completed quests (populated by MakeCompletedQuestViews where GameState is available)
	ps.CompletedQuests = []QuestView{}

	// Buildings
	ps.Buildings = make([]string, 0, len(p.BuiltBuildings))
	for _, b := range p.BuiltBuildings {
		ps.Buildings = append(ps.Buildings, b.Name)
	}

	// ActiveQuests will be populated by MakeQuestViews where GameState is available
	ps.ActiveQuests = []QuestView{}

	return ps
}

// MakePlayerStateWithLocations creates a PlayerState with location type info from GameState.
func MakePlayerStateWithLocations(p *models.Character, gs *models.GameState) *PlayerState {
	ps := MakePlayerState(p)
	if ps == nil || gs == nil {
		return ps
	}

	// Enrich known locations with type info
	ps.KnownLocations = make([]LocationView, 0)
	for _, locName := range p.KnownLocations {
		lv := LocationView{Name: locName, Locked: false}
		if loc, exists := gs.GameLocations[locName]; exists {
			lv.Type = loc.Type
			lv.LevelMax = loc.LevelMax
			lv.RarityMax = loc.RarityMax
		}
		ps.KnownLocations = append(ps.KnownLocations, lv)
	}

	// Enrich locked locations with type info
	ps.LockedLocations = make([]LocationView, 0)
	for _, locName := range p.LockedLocations {
		lv := LocationView{Name: locName, Locked: true}
		if loc, exists := gs.GameLocations[locName]; exists {
			lv.Type = loc.Type
			lv.LevelMax = loc.LevelMax
			lv.RarityMax = loc.RarityMax
		}
		ps.LockedLocations = append(ps.LockedLocations, lv)
	}

	return ps
}

// MakeQuestViews builds quest views for the frontend. Requires GameState for quest data.
func MakeQuestViews(p *models.Character, gs *models.GameState) []QuestView {
	if p == nil || gs == nil {
		return []QuestView{}
	}

	if gs.AvailableQuests == nil {
		gs.AvailableQuests = make(map[string]models.Quest)
		for k, v := range data.StoryQuests {
			gs.AvailableQuests[k] = v
		}
	}

	views := make([]QuestView, 0)
	for _, questID := range p.ActiveQuests {
		quest, exists := gs.AvailableQuests[questID]
		if !exists {
			continue
		}

		// Update current values
		switch quest.Requirement.Type {
		case "level":
			quest.Requirement.CurrentValue = p.Level
		case "village_level":
			if gs.Villages != nil && p.VillageName != "" {
				if village, ok := gs.Villages[p.VillageName]; ok {
					quest.Requirement.CurrentValue = village.Level
				}
			}
		case "total_resources":
			total := 0
			for _, res := range p.ResourceStorageMap {
				total += res.Stock
			}
			quest.Requirement.CurrentValue = total
		case "skill_count":
			quest.Requirement.CurrentValue = len(p.LearnedSkills)
		}

		views = append(views, QuestView{
			ID:          questID,
			Name:        quest.Name,
			Description: quest.Description,
			Progress:    quest.Requirement.CurrentValue,
			Target:      quest.Requirement.TargetValue,
			RewardXP:    quest.Reward.XP,
			Completed:   false,
		})
	}

	return views
}

// MakeCompletedQuestViews builds completed quest views for the frontend.
func MakeCompletedQuestViews(p *models.Character, gs *models.GameState) []QuestView {
	if p == nil || gs == nil {
		return []QuestView{}
	}

	views := make([]QuestView, 0)
	for _, questID := range p.CompletedQuests {
		quest, exists := gs.AvailableQuests[questID]
		if !exists {
			continue
		}
		views = append(views, QuestView{
			ID:          questID,
			Name:        quest.Name,
			Description: quest.Description,
			Progress:    quest.Requirement.TargetValue,
			Target:      quest.Requirement.TargetValue,
			RewardXP:    quest.Reward.XP,
			Completed:   true,
		})
	}
	return views
}

// MakeVillageView builds a VillageView from a models.Village.
func MakeVillageView(village *models.Village) *VillageView {
	if village == nil {
		return nil
	}

	vv := &VillageView{
		Name:             village.Name,
		Level:            village.Level,
		Experience:       village.Experience,
		ExpToLevel:       village.Level * 100,
		DefenseLevel:     village.DefenseLevel,
		UnlockedCrafting: village.UnlockedCrafting,
		ResourcePerTick:  village.ResourcePerTick,
		LastHarvestTime:  village.LastHarvestTime,
	}

	if vv.UnlockedCrafting == nil {
		vv.UnlockedCrafting = []string{}
	}
	if vv.ResourcePerTick == nil {
		vv.ResourcePerTick = make(map[string]int)
	}

	// Villagers
	vv.Villagers = make([]VillagerView, 0, len(village.Villagers))
	for _, v := range village.Villagers {
		vv.Villagers = append(vv.Villagers, VillagerView{
			Name:         v.Name,
			Role:         v.Role,
			Level:        v.Level,
			Efficiency:   v.Efficiency,
			AssignedTask: v.AssignedTask,
			HarvestType:  v.HarvestType,
		})
	}

	// Guards
	vv.Guards = make([]VillageGuardView, 0, len(village.ActiveGuards))
	for _, g := range village.ActiveGuards {
		vv.Guards = append(vv.Guards, VillageGuardView{
			Name:         g.Name,
			Level:        g.Level,
			HP:           g.HitpointsRemaining,
			MaxHP:        g.HitPoints,
			AttackBonus:  g.AttackBonus,
			DefenseBonus: g.DefenseBonus,
			Hired:        g.Hired,
			Cost:         g.Cost,
			Injured:      g.Injured,
			RecoveryTime: g.RecoveryTime,
		})
	}

	// Defenses
	vv.Defenses = make([]DefenseView, 0, len(village.Defenses))
	for _, d := range village.Defenses {
		vv.Defenses = append(vv.Defenses, DefenseView{
			Name:  d.Name,
			Level: d.Level,
			Type:  d.Type,
			Built: d.Built,
		})
	}

	// Traps
	vv.Traps = make([]TrapView, 0, len(village.Traps))
	for _, t := range village.Traps {
		vv.Traps = append(vv.Traps, TrapView{
			Name:      t.Name,
			Type:      t.Type,
			Damage:    t.Damage,
			Remaining: t.Remaining,
		})
	}

	return vv
}

// --- Town view structs ---

// TownView represents town state for the frontend.
type TownView struct {
	Name           string           `json:"name"`
	TaxRate        int              `json:"tax_rate"`
	Treasury       map[string]int   `json:"treasury,omitempty"`
	Guests         []InnGuestView   `json:"guests"`
	Mayor          *MayorView       `json:"mayor"`
	FetchQuests    []FetchQuestView `json:"fetch_quests"`
	IsCurrentMayor bool             `json:"is_current_mayor"`
	AttackLog      []AttackLogView  `json:"attack_log"`
}

// InnGuestView represents an inn guest for the frontend.
type InnGuestView struct {
	CharacterName string `json:"character_name"`
	Level         int    `json:"level"`
	GuardCount    int    `json:"guard_count"`
	IsOwn         bool   `json:"is_own"`
	IsNPC         bool   `json:"is_npc"`
	GoldCarried   int    `json:"gold_carried"`
}

// MayorView represents the mayor for the frontend.
type MayorView struct {
	Name       string `json:"name"`
	IsNPC      bool   `json:"is_npc"`
	Level      int    `json:"level"`
	GuardCount int    `json:"guard_count"`
	MonsterCount int  `json:"monster_count"`
}

// FetchQuestView represents a fetch quest for the frontend.
type FetchQuestView struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Resource    string `json:"resource"`
	Amount      int    `json:"amount"`
	RewardGold  int    `json:"reward_gold"`
	RewardXP    int    `json:"reward_xp"`
	ClaimedBy   string `json:"claimed_by"`
	Completed   bool   `json:"completed"`
}

// AttackLogView represents an attack log entry for the frontend.
type AttackLogView struct {
	AttackerName string `json:"attacker_name"`
	TargetName   string `json:"target_name"`
	AttackType   string `json:"attack_type"`
	Success      bool   `json:"success"`
	Details      string `json:"details"`
}

// MakeTownView builds a TownView from a models.Town for the frontend.
func MakeTownView(town *models.Town, accountID int64, charName string) *TownView {
	if town == nil {
		return nil
	}

	tv := &TownView{
		Name:    town.Name,
		TaxRate: town.TaxRate,
	}

	// Check if current player is mayor
	isMayor := false
	if town.Mayor != nil && !town.Mayor.IsNPC && town.Mayor.AccountID == accountID && town.Mayor.CharacterName == charName {
		isMayor = true
		tv.Treasury = town.Treasury
	}
	tv.IsCurrentMayor = isMayor

	// Mayor view
	if town.Mayor != nil {
		name := town.Mayor.NPCName
		if !town.Mayor.IsNPC {
			name = town.Mayor.CharacterName
		}
		tv.Mayor = &MayorView{
			Name:         name,
			IsNPC:        town.Mayor.IsNPC,
			Level:        town.Mayor.Level,
			GuardCount:   len(town.Mayor.Guards),
			MonsterCount: len(town.Mayor.Monsters),
		}
	}

	// Guests
	tv.Guests = make([]InnGuestView, 0, len(town.InnGuests))
	for _, guest := range town.InnGuests {
		tv.Guests = append(tv.Guests, InnGuestView{
			CharacterName: guest.CharacterName,
			Level:         guest.Level,
			GuardCount:    len(guest.HiredGuards),
			IsOwn:         guest.AccountID == accountID && guest.CharacterName == charName,
			IsNPC:         guest.AccountID == 0,
			GoldCarried:   guest.GoldCarried,
		})
	}

	// Fetch quests
	tv.FetchQuests = make([]FetchQuestView, 0)
	for _, fq := range town.FetchQuests {
		if !fq.Active {
			continue
		}
		tv.FetchQuests = append(tv.FetchQuests, FetchQuestView{
			ID:          fq.ID,
			Name:        fq.Name,
			Description: fq.Description,
			Resource:    fq.Resource,
			Amount:      fq.Amount,
			RewardGold:  fq.RewardGold,
			RewardXP:    fq.RewardXP,
			ClaimedBy:   fq.ClaimedBy,
			Completed:   fq.Completed,
		})
	}

	// Attack log (last 20)
	tv.AttackLog = make([]AttackLogView, 0)
	start := 0
	if len(town.AttackLog) > 20 {
		start = len(town.AttackLog) - 20
	}
	for i := len(town.AttackLog) - 1; i >= start; i-- {
		log := town.AttackLog[i]
		tv.AttackLog = append(tv.AttackLog, AttackLogView{
			AttackerName: log.AttackerName,
			TargetName:   log.TargetName,
			AttackType:   log.AttackType,
			Success:      log.Success,
			Details:      log.Details,
		})
	}

	return tv
}

func MakeCombatView(session *GameSession) *CombatView {
	if session.Combat == nil {
		return nil
	}
	c := session.Combat
	p := session.Player
	m := &c.Mob

	view := &CombatView{
		Turn:              c.Turn,
		PlayerHP:          p.HitpointsRemaining,
		PlayerMaxHP:       p.HitpointsTotal,
		PlayerMP:          p.ManaRemaining,
		PlayerMaxMP:       p.ManaTotal,
		PlayerSP:          p.StaminaRemaining,
		PlayerMaxSP:       p.StaminaTotal,
		MonsterName:       m.Name,
		MonsterType:       m.MonsterType,
		MonsterLevel:      m.Level,
		MonsterHP:         m.HitpointsRemaining,
		MonsterMaxHP:      m.HitpointsTotal,
		MonsterMP:         m.ManaRemaining,
		MonsterMaxMP:      m.ManaTotal,
		MonsterSP:         m.StaminaRemaining,
		MonsterMaxSP:      m.StaminaTotal,
		MonsterIsBoss:     m.IsBoss,
		MonsterIsGuardian: m.IsSkillGuardian,
		HuntsRemaining:    c.HuntsRemaining,
	}

	if m.IsSkillGuardian {
		view.GuardedSkillName = m.GuardedSkill.Name
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
