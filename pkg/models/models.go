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

// MonsterRarity represents rarity tiers for monsters
type MonsterRarity string

const (
	RarityCommon    MonsterRarity = "common"
	RarityUncommon  MonsterRarity = "uncommon"
	RarityRare      MonsterRarity = "rare"
	RarityEpic      MonsterRarity = "epic"
	RarityLegendary MonsterRarity = "legendary"
	RarityMythic    MonsterRarity = "mythic"
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
	LastHarvestTime  int64          `json:"last_harvest_time"`
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
	Stats              CharacterStats         `json:"stats"`
	ActiveDungeon      *Dungeon               `json:"active_dungeon,omitempty"`
	ActiveNPCQuests    []string               `json:"active_npc_quests"`
	CompletedNPCQuests []string               `json:"completed_npc_quests"`
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
	Rarity             MonsterRarity          `json:"rarity"`
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
	ID                 string                 `json:"id,omitempty"`
	LocationName       string                 `json:"location_name,omitempty"`
	PlayerKills        int                    `json:"player_kills,omitempty"`
	MonsterKills       int                    `json:"monster_kills,omitempty"`
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
	Townsfolk   []NPCTownsfolk    `json:"townsfolk"`
	GossipBoard []string          `json:"gossip_board"`
	NPCFighters []NPCFighter      `json:"npc_fighters"`
	NPCQuests   []NPCQuest        `json:"npc_quests"`
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

// CharacterStats tracks kill/death/progression analytics.
type CharacterStats struct {
	TotalKills      int            `json:"total_kills"`
	TotalDeaths     int            `json:"total_deaths"`
	KillsByRarity   map[string]int `json:"kills_by_rarity"`
	KillsByMonster  map[string]int `json:"kills_by_monster"`
	KillsByLocation map[string]int `json:"kills_by_location"`
	BossesKilled    int            `json:"bosses_killed"`
	TotalXPEarned   int            `json:"total_xp_earned"`
	HighestCombo    int            `json:"highest_combo"`
	CurrentCombo    int            `json:"current_combo"`
	PvPWins         int            `json:"pvp_wins"`
	PvPLosses       int            `json:"pvp_losses"`
	DungeonsCleared int            `json:"dungeons_cleared"`
	DungeonsEntered int            `json:"dungeons_entered"`
	FloorsCleared   int            `json:"floors_cleared"`
	RoomsExplored   int            `json:"rooms_explored"`
	TreasuresFound  int            `json:"treasures_found"`
	TrapsTriggered  int            `json:"traps_triggered"`
	HiddenChests      int            `json:"hidden_chests_found"`
	ArenaRating       int            `json:"arena_rating"`
	ArenaWins         int            `json:"arena_wins"`
	ArenaLosses       int            `json:"arena_losses"`
	ArenaBattlesToday int            `json:"arena_battles_today"`
	ArenaLastReset    int64          `json:"arena_last_reset"`
}

// ArenaChallenge holds the target of an arena challenge.
type ArenaChallenge struct {
	TargetAccountID int64  `json:"target_account_id"`
	TargetCharName  string `json:"target_char_name"`
}

// GridPosition represents a 2D coordinate on the dungeon grid.
type GridPosition struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// DungeonTile represents a single tile on the dungeon grid.
type DungeonTile struct {
	Type     string `json:"type"`      // "room", "corridor", "wall", "entrance", "exit"
	RoomIdx  int    `json:"room_idx"`  // index into DungeonFloor.Rooms, -1 if not a room
	Explored bool   `json:"explored"`
}

// Dungeon represents an active dungeon run.
type Dungeon struct {
	Name         string         `json:"name"`
	Floors       []DungeonFloor `json:"floors"`
	CurrentFloor int            `json:"current_floor"`
	BaseLevelMin int            `json:"base_level_min"`
	BaseLevelMax int            `json:"base_level_max"`
	BaseRankMax  int            `json:"base_rank_max"`
	Seed         int64          `json:"seed"`
}

// DungeonFloor represents a single floor within a dungeon.
type DungeonFloor struct {
	FloorNumber int             `json:"floor_number"`
	Rooms       []DungeonRoom   `json:"rooms"`
	CurrentRoom int             `json:"current_room"`
	Cleared     bool            `json:"cleared"`
	BossFloor   bool            `json:"boss_floor"`
	Grid        [][]DungeonTile `json:"grid,omitempty"`
	GridW       int             `json:"grid_w"`
	GridH       int             `json:"grid_h"`
	PlayerPos   GridPosition    `json:"player_pos"`
	ExitPos     GridPosition    `json:"exit_pos"`
}

// DungeonRoom represents a single room on a dungeon floor.
type DungeonRoom struct {
	Type       string   `json:"type"` // combat, treasure, trap, rest, boss, merchant
	Cleared    bool     `json:"cleared"`
	Monster    *Monster `json:"monster,omitempty"`
	Loot       []Item   `json:"loot,omitempty"`
	TrapDamage int      `json:"trap_damage,omitempty"`
	HealAmount int      `json:"heal_amount,omitempty"`
	GridX      int      `json:"grid_x"`
	GridY      int      `json:"grid_y"`
	RoomW      int      `json:"room_w"`
	RoomH      int      `json:"room_h"`
}

// NPCTownsfolk represents a persistent NPC in the town.
type NPCTownsfolk struct {
	ID            string          `json:"id"`
	Name          string          `json:"name"`
	Title         string          `json:"title"`
	Personality   NPCPersonality  `json:"personality"`
	Memory        []NPCMemory     `json:"memory"`
	Relationships map[string]int  `json:"relationships"`
	Level         int             `json:"level"`
	Age           int             `json:"age"`
	IsAlive       bool            `json:"is_alive"`
	CurrentMood   string          `json:"current_mood"`
	QuestGiver    bool            `json:"quest_giver"`
	LocationName  string          `json:"location_name"`
	GoldCarried   int             `json:"gold_carried"`
}

// NPCPersonality defines the behavioral traits of an NPC.
type NPCPersonality struct {
	Archetype  string  `json:"archetype"`
	Chattiness float64 `json:"chattiness"`
	Generosity float64 `json:"generosity"`
	Courage    float64 `json:"courage"`
	Curiosity  float64 `json:"curiosity"`
}

// NPCMemory records a single event an NPC remembers.
type NPCMemory struct {
	EventType   string `json:"event_type"`
	PlayerName  string `json:"player_name"`
	Description string `json:"description"`
	Timestamp   int64  `json:"timestamp"`
	Sentiment   int    `json:"sentiment"`
}

// NPCFighter is an NPC that can be hired for combat at the inn.
type NPCFighter struct {
	NPCID     string `json:"npc_id"`
	Name      string `json:"name"`
	Level     int    `json:"level"`
	HireCost  int    `json:"hire_cost"`
	Specialty string `json:"specialty"`
}

// NPCQuest is a dynamically generated quest from an NPC townsfolk.
type NPCQuest struct {
	ID          string         `json:"id"`
	NPCID       string         `json:"npc_id"`
	NPCName     string         `json:"npc_name"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Type        string         `json:"type"`
	Requirement NPCQuestReq    `json:"requirement"`
	Reward      NPCQuestReward `json:"reward"`
	Difficulty  string         `json:"difficulty"`
	RepRequired int            `json:"rep_required"`
	CreatedAt   int64          `json:"created_at"`
	AcceptedBy  string         `json:"accepted_by,omitempty"`
	Completed   bool           `json:"completed"`
	Failed      bool           `json:"failed"`
}

// NPCQuestReq defines the requirement to complete an NPC quest.
type NPCQuestReq struct {
	Type         string `json:"type"`
	TargetName   string `json:"target_name"`
	TargetCount  int    `json:"target_count"`
	CurrentCount int    `json:"current_count"`
}

// NPCQuestReward defines the rewards for completing an NPC quest.
type NPCQuestReward struct {
	XP         int `json:"xp"`
	Gold       int `json:"gold"`
	Reputation int `json:"reputation"`
	ItemRarity int `json:"item_rarity,omitempty"`
}

// MostWantedEntry represents a monster on the Most Wanted board.
type MostWantedEntry struct {
	MonsterID    string        `json:"monster_id"`
	Name         string        `json:"name"`
	MonsterType  string        `json:"monster_type"`
	Level        int           `json:"level"`
	Rarity       MonsterRarity `json:"rarity"`
	PlayerKills  int           `json:"player_kills"`
	MonsterKills int           `json:"monster_kills"`
	LocationName string        `json:"location_name"`
	LocationIdx  int           `json:"location_idx"`
	IsBoss       bool          `json:"is_boss"`
	HP           int           `json:"hp"`
}
