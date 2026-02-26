package models

// DamageType represents elemental damage categories
type DamageType string

const (
	Physical  DamageType = "physical"
	Fire      DamageType = "fire"
	Ice       DamageType = "ice"
	Lightning DamageType = "lightning"
	Poison    DamageType = "poison"
)

type GameState struct {
	CharactersMap   map[string]Character `json:"characters_map"`
	GameLocations   map[string]Location  `json:"game_locations"`
	AvailableQuests map[string]Quest     `json:"available_quests"`
	Villages        map[string]Village   `json:"villages"`
}

type Village struct {
	Name             string         `json:"name"`
	Level            int            `json:"level"`
	Experience       int            `json:"experience"`
	Villagers        []Villager     `json:"villagers"`
	Defenses         []Defense      `json:"defenses"`
	Traps            []Trap         `json:"traps"`
	ResourcePerTick  map[string]int `json:"resource_per_tick"`
	UnlockedCrafting []string       `json:"unlocked_crafting"`
	DefenseLevel     int            `json:"defense_level"`
	LastTideTime     int64          `json:"last_tide_time"`
	TideInterval     int            `json:"tide_interval"`
	ActiveGuards     []Guard        `json:"active_guards"`
}

type Villager struct {
	Name         string `json:"name"`
	Role         string `json:"role"`
	Level        int    `json:"level"`
	Efficiency   int    `json:"efficiency"`
	AssignedTask string `json:"assigned_task"`
	HarvestType  string `json:"harvest_type"`
}

type Guard struct {
	Name               string                 `json:"name"`
	Level              int                    `json:"level"`
	HitPoints          int                    `json:"hit_points"`
	HitpointsNatural   int                    `json:"hitpoints_natural"`
	HitpointsRemaining int                    `json:"hitpoints_remaining"`
	AttackBonus        int                    `json:"attack_bonus"`
	DefenseBonus       int                    `json:"defense_bonus"`
	AttackRolls        int                    `json:"attack_rolls"`
	DefenseRolls       int                    `json:"defense_rolls"`
	Hired              bool                   `json:"hired"`
	Cost               int                    `json:"cost"`
	Inventory          []Item                 `json:"inventory"`
	EquipmentMap       map[int]Item           `json:"equipment_map"`
	StatsMod           StatMod                `json:"stats_mod"`
	Injured            bool                   `json:"injured"`
	RecoveryTime       int                    `json:"recovery_time"`
	StatusEffects      []StatusEffect         `json:"status_effects"`
	Resistances        map[DamageType]float64 `json:"resistances"`
}

type Defense struct {
	Name        string `json:"name"`
	Level       int    `json:"level"`
	Defense     int    `json:"defense"`
	AttackPower int    `json:"attack_power"`
	Range       int    `json:"range"`
	Built       bool   `json:"built"`
	Type        string `json:"type"`
}

type Trap struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Damage      int    `json:"damage"`
	Duration    int    `json:"duration"`
	Remaining   int    `json:"remaining"`
	TriggerRate int    `json:"trigger_rate"`
}

type CraftingRecipe struct {
	Name              string         `json:"name"`
	Type              string         `json:"type"`
	RequiredResources map[string]int `json:"required_resources"`
	RequiredLevel     int            `json:"required_level"`
	Output            Item           `json:"output"`
	SkillUpgrade      SkillUpgrade   `json:"skill_upgrade"`
}

type SkillUpgrade struct {
	SkillName      string `json:"skill_name"`
	UpgradeLevel   int    `json:"upgrade_level"`
	DamageIncrease int    `json:"damage_increase"`
	CostReduction  int    `json:"cost_reduction"`
	Description    string `json:"description"`
}

type Character struct {
	Name               string                 `json:"name"`
	Level              int                    `json:"level"`
	Experience         int                    `json:"experience"`
	ExpSinceLevel      int                    `json:"exp_since_level"`
	HitpointsTotal     int                    `json:"hitpoints_total"`
	HitpointsNatural   int                    `json:"hitpoints_natural"`
	HitpointsRemaining int                    `json:"hitpoints_remaining"`
	ManaTotal          int                    `json:"mana_total"`
	ManaNatural        int                    `json:"mana_natural"`
	ManaRemaining      int                    `json:"mana_remaining"`
	StaminaTotal       int                    `json:"stamina_total"`
	StaminaNatural     int                    `json:"stamina_natural"`
	StaminaRemaining   int                    `json:"stamina_remaining"`
	AttackRolls        int                    `json:"attack_rolls"`
	DefenseRolls       int                    `json:"defense_rolls"`
	StatsMod           StatMod                `json:"stats_mod"`
	Resurrections      int                    `json:"resurrections"`
	Inventory          []Item                 `json:"inventory"`
	EquipmentMap       map[int]Item           `json:"equipment_map"`
	ResourceStorageMap map[string]Resource    `json:"resource_storage_map"`
	KnownLocations     []string               `json:"known_locations"`
	LockedLocations    []string               `json:"locked_locations"`
	BuiltBuildings     []Building             `json:"built_buildings"`
	LearnedSkills      []Skill                `json:"learned_skills"`
	StatusEffects      []StatusEffect         `json:"status_effects"`
	Resistances        map[DamageType]float64 `json:"resistances"`
	CompletedQuests    []string               `json:"completed_quests"`
	ActiveQuests       []string               `json:"active_quests"`
	VillageName        string                 `json:"village_name"`
}

type Quest struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Type        string           `json:"type"`
	Requirement QuestRequirement `json:"requirement"`
	Reward      QuestReward      `json:"reward"`
	Completed   bool             `json:"completed"`
	Active      bool             `json:"active"`
}

type QuestRequirement struct {
	Type         string `json:"type"`
	TargetValue  int    `json:"target_value"`
	TargetName   string `json:"target_name"`
	CurrentValue int    `json:"current_value"`
}

type QuestReward struct {
	Type  string `json:"type"`
	Value string `json:"value"`
	XP    int    `json:"xp"`
}

type Building struct {
	Name                string         `json:"name"`
	RequiredResourceMap map[string]int `json:"required_resource_map"`
	StatsMod            StatMod        `json:"stats_mod"`
}

type Location struct {
	Name      string    `json:"name"`
	Weight    int       `json:"weight"`
	Type      string    `json:"type"`
	LevelMax  int       `json:"level_max"`
	RarityMax int       `json:"rarity_max"`
	Monsters  []Monster `json:"monsters"`
}

type Resource struct {
	Name         string `json:"name"`
	Stock        int    `json:"stock"`
	RollModifier int    `json:"roll_modifier"`
}

type StatMod struct {
	AttackMod   int `json:"attack_mod"`
	DefenseMod  int `json:"defense_mod"`
	HitPointMod int `json:"hit_point_mod"`
}

type StatusEffect struct {
	Type     string `json:"type"`
	Duration int    `json:"duration"`
	Potency  int    `json:"potency"`
}

type Skill struct {
	Name        string       `json:"name"`
	ManaCost    int          `json:"mana_cost"`
	StaminaCost int          `json:"stamina_cost"`
	Damage      int          `json:"damage"`
	DamageType  DamageType   `json:"damage_type"`
	Effect      StatusEffect `json:"effect"`
	Description string       `json:"description"`
}

type Monster struct {
	Name               string                 `json:"name"`
	Level              int                    `json:"level"`
	Experience         int                    `json:"experience"`
	ExpSinceLevel      int                    `json:"exp_since_level"`
	Rank               int                    `json:"rank"`
	HitpointsTotal     int                    `json:"hitpoints_total"`
	HitpointsNatural   int                    `json:"hitpoints_natural"`
	HitpointsRemaining int                    `json:"hitpoints_remaining"`
	ManaTotal          int                    `json:"mana_total"`
	ManaNatural        int                    `json:"mana_natural"`
	ManaRemaining      int                    `json:"mana_remaining"`
	StaminaTotal       int                    `json:"stamina_total"`
	StaminaNatural     int                    `json:"stamina_natural"`
	StaminaRemaining   int                    `json:"stamina_remaining"`
	AttackRolls        int                    `json:"attack_rolls"`
	DefenseRolls       int                    `json:"defense_rolls"`
	StatsMod           StatMod                `json:"stats_mod"`
	Inventory          []Item                 `json:"inventory"`
	EquipmentMap       map[int]Item           `json:"equipment_map"`
	LearnedSkills      []Skill                `json:"learned_skills"`
	StatusEffects      []StatusEffect         `json:"status_effects"`
	Resistances        map[DamageType]float64 `json:"resistances"`
	MonsterType        string                 `json:"monster_type"`
	IsSkillGuardian    bool                   `json:"is_skill_guardian"`
	GuardedSkill       Skill                  `json:"guarded_skill"`
	IsBoss             bool                   `json:"is_boss"`
}

type Item struct {
	Name        string           `json:"name"`
	Rarity      int              `json:"rarity"`
	Slot        int              `json:"slot"`
	StatsMod    StatMod          `json:"stats_mod"`
	CP          int              `json:"cp"`
	ItemType    string           `json:"item_type"`
	Consumable  ConsumableEffect `json:"consumable"`
	SkillScroll SkillScrollData  `json:"skill_scroll"`
}

type ConsumableEffect struct {
	EffectType string `json:"effect_type"`
	Value      int    `json:"value"`
	Duration   int    `json:"duration"`
}

type SkillScrollData struct {
	Skill         Skill `json:"skill"`
	CanBeCrafted  bool  `json:"can_be_crafted"`
	CraftingValue int   `json:"crafting_value"`
}

// Town is the shared town record (one per game world).
type Town struct {
	Name        string            `json:"name"`
	InnGuests   []InnGuest        `json:"inn_guests"`
	Mayor       *MayorData        `json:"mayor"`
	Treasury    map[string]int    `json:"treasury"`
	TaxRate     int               `json:"tax_rate"`
	FetchQuests []FetchQuest      `json:"fetch_quests"`
	AttackLog   []TownAttackLog   `json:"attack_log"`
}

// InnGuest is a snapshot of a sleeping player for async PvP.
type InnGuest struct {
	AccountID     int64                  `json:"account_id"`
	CharacterName string                 `json:"character_name"`
	CheckInTime   int64                  `json:"check_in_time"`
	GoldPaid      int                    `json:"gold_paid"`
	GoldCarried   int                    `json:"gold_carried"`
	HiredGuards   []Guard                `json:"hired_guards"`
	Level         int                    `json:"level"`
	HP            int                    `json:"hp"`
	MaxHP         int                    `json:"max_hp"`
	MP            int                    `json:"mp"`
	MaxMP         int                    `json:"max_mp"`
	SP            int                    `json:"sp"`
	MaxSP         int                    `json:"max_sp"`
	AttackRolls   int                    `json:"attack_rolls"`
	DefenseRolls  int                    `json:"defense_rolls"`
	StatsMod      StatMod                `json:"stats_mod"`
	EquipmentMap  map[int]Item           `json:"equipment_map"`
	LearnedSkills []Skill                `json:"learned_skills"`
	Resistances   map[DamageType]float64 `json:"resistances"`
}

// MayorData represents the town mayor (NPC or player).
type MayorData struct {
	IsNPC         bool                   `json:"is_npc"`
	AccountID     int64                  `json:"account_id"`
	CharacterName string                 `json:"character_name"`
	NPCName       string                 `json:"npc_name"`
	Level         int                    `json:"level"`
	Guards        []Guard                `json:"guards"`
	Monsters      []Monster              `json:"monsters"`
	HP            int                    `json:"hp"`
	MaxHP         int                    `json:"max_hp"`
	AttackRolls   int                    `json:"attack_rolls"`
	DefenseRolls  int                    `json:"defense_rolls"`
	StatsMod      StatMod                `json:"stats_mod"`
	EquipmentMap  map[int]Item           `json:"equipment_map"`
	LearnedSkills []Skill                `json:"learned_skills"`
	Resistances   map[DamageType]float64 `json:"resistances"`
}

// FetchQuest is a mayor-created resource delivery quest.
type FetchQuest struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Resource    string `json:"resource"`
	Amount      int    `json:"amount"`
	RewardGold  int    `json:"reward_gold"`
	RewardXP    int    `json:"reward_xp"`
	CreatedBy   string `json:"created_by"`
	ClaimedBy   string `json:"claimed_by"`
	Completed   bool   `json:"completed"`
	Active      bool   `json:"active"`
}

// TownAttackLog records PvP and mayor challenge events.
type TownAttackLog struct {
	AttackerName string `json:"attacker_name"`
	TargetName   string `json:"target_name"`
	AttackType   string `json:"attack_type"`
	Success      bool   `json:"success"`
	Timestamp    int64  `json:"timestamp"`
	Details      string `json:"details"`
}
