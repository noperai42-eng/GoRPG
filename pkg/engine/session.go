package engine

import "rpg-game/pkg/models"

// Session states
const (
	StateInit               = "init"
	StateMainMenu           = "main_menu"
	StateCharacterCreate    = "character_create"
	StateCharacterSelect    = "character_select"
	StateHarvestSelect      = "harvest_select"
	StateHuntLocationSelect = "hunt_location_select"
	StateHuntCountSelect    = "hunt_count_select"
	StateHuntTracking       = "hunt_tracking"

	StateCombat            = "combat"
	StateCombatItemSelect  = "combat_item_select"
	StateCombatSkillSelect = "combat_skill_select"
	StateCombatGuardPrompt = "combat_guard_prompt"
	StateCombatSkillReward = "combat_skill_reward"

	StateAutoPlaySpeed = "autoplay_speed"
	StateAutoPlayMenu  = "autoplay_menu"

	StateQuestLog            = "quest_log"
	StatePlayerStats         = "player_stats"
	StateDiscoveredLocations = "discovered_locations"
	StateLoadSave            = "load_save"
	StateLoadSaveCharSelect  = "load_save_char_select"
	StateBuildSelect         = "build_select"

	StateVillageMain             = "village_main"
	StateVillageViewVillagers    = "village_view_villagers"
	StateVillageAssignTask       = "village_assign_task"
	StateVillageAssignResource   = "village_assign_resource"
	StateVillageHireGuard        = "village_hire_guard"
	StateVillageCrafting         = "village_crafting"
	StateVillageCraftPotion      = "village_craft_potion"
	StateVillageCraftArmor       = "village_craft_armor"
	StateVillageCraftWeapon      = "village_craft_weapon"
	StateVillageUpgradeSkill     = "village_upgrade_skill"
	StateVillageUpgradeConfirm   = "village_upgrade_confirm"
	StateVillageCraftScrolls     = "village_craft_scrolls"
	StateVillageBuildDefense     = "village_build_defense"
	StateVillageBuildWalls       = "village_build_walls"
	StateVillageCraftTraps       = "village_craft_traps"
	StateVillageViewDefenses     = "village_view_defenses"
	StateVillageCheckTide        = "village_check_tide"
	StateVillageMonsterTide      = "village_monster_tide"
	StateVillageTideWave         = "village_tide_wave"
	StateVillageManageGuards     = "village_manage_guards"
	StateVillageManageGuard      = "village_manage_guard"
	StateVillageEquipGuard       = "village_equip_guard"
	StateVillageUnequipGuard     = "village_unequip_guard"
	StateVillageGiveItem         = "village_give_item"
	StateVillageTakeItem         = "village_take_item"
	StateVillageHealGuard        = "village_heal_guard"

	StateGuideMain         = "guide_main"
	StateGuideCombat       = "guide_combat"
	StateGuideSkills       = "guide_skills"
	StateGuideVillage      = "guide_village"
	StateGuideCrafting     = "guide_crafting"
	StateGuideMonsterDrops = "guide_monster_drops"
	StateGuideAutoPlay     = "guide_autoplay"
	StateGuideQuests       = "guide_quests"
)

// CombatContext tracks turn-by-turn combat state.
type CombatContext struct {
	Mob            models.Monster
	MobLoc         int
	Location       *models.Location
	Turn           int
	Fled           bool
	PlayerWon      bool
	IsDefending    bool
	CombatGuards   []models.Guard
	HasGuards      bool
	GuardianLocationName string // non-empty = fighting a location guardian
	HuntsRemaining       int
	IsAutoPlay           bool
	AutoPlaySpeed  string
	AutoPlayFights int
	AutoPlayWins   int
	AutoPlayDeaths int
	AutoPlayXP     int
}

// GameSession represents a connected player's session.
type GameSession struct {
	ID        string
	AccountID int64
	State     string
	Player    *models.Character
	GameState *models.GameState
	Combat    *CombatContext
	SaveFile  string

	// Context for multi-step operations
	SelectedLocation    string
	SelectedVillage     *models.Village
	SelectedGuardIdx    int
	SelectedSkillIdx    int
	SelectedVillagerIdx int
}
