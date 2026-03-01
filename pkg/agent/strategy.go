package agent

import (
	"math/rand"
	"strings"

	"rpg-game/pkg/engine"
)

// Strategy decides which command an agent should send next.
type Strategy interface {
	Name() string
	Decide(screen string, options []engine.MenuOption) engine.GameCommand
}

// findOption returns the first option whose Key matches key, or nil.
func findOption(options []engine.MenuOption, key string) *engine.MenuOption {
	for i := range options {
		if options[i].Key == key && options[i].Enabled {
			return &options[i]
		}
	}
	return nil
}

// pickRandom returns a random enabled option, or nil if none are enabled.
func pickRandom(options []engine.MenuOption) *engine.MenuOption {
	var enabled []int
	for i, opt := range options {
		if opt.Enabled {
			enabled = append(enabled, i)
		}
	}
	if len(enabled) == 0 {
		return nil
	}
	return &options[enabled[rand.Intn(len(enabled))]]
}

// pickFirst returns the first enabled option, or nil.
func pickFirst(options []engine.MenuOption) *engine.MenuOption {
	for i := range options {
		if options[i].Enabled {
			return &options[i]
		}
	}
	return nil
}

// selectCmd creates a select-type GameCommand.
func selectCmd(value string) engine.GameCommand {
	return engine.GameCommand{Type: "select", Value: value}
}

// inputCmd creates an input-type GameCommand.
func inputCmd(value string) engine.GameCommand {
	return engine.GameCommand{Type: "input", Value: value}
}

// defaultNavigation handles common non-combat screens by picking
// the first available option or going home.
func defaultNavigation(screen string, options []engine.MenuOption) engine.GameCommand {
	// Screens with a "back" option — just go back.
	if opt := findOption(options, "back"); opt != nil {
		return selectCmd("back")
	}
	if opt := findOption(options, "0"); opt != nil {
		return selectCmd("0")
	}
	if opt := pickFirst(options); opt != nil {
		return selectCmd(opt.Key)
	}
	// Fallback: go home.
	return selectCmd("home")
}

// weightedChoice picks an index based on cumulative weights.
func weightedChoice(weights []int) int {
	total := 0
	for _, w := range weights {
		total += w
	}
	r := rand.Intn(total)
	cum := 0
	for i, w := range weights {
		cum += w
		if r < cum {
			return i
		}
	}
	return len(weights) - 1
}

// combatDecision picks a combat action key based on the given weight distribution.
// weights order: attack, skill, defend, item, auto, flee
func combatDecision(weights [6]int, options []engine.MenuOption) engine.GameCommand {
	keys := [6]string{"1", "4", "2", "3", "6", "5"}
	idx := weightedChoice(weights[:])
	key := keys[idx]
	if opt := findOption(options, key); opt != nil {
		return selectCmd(key)
	}
	// Fallback to attack.
	return selectCmd("1")
}

// mainMenuChoice picks a main menu option based on weighted keys.
func mainMenuChoice(keys []string, weights []int, options []engine.MenuOption) engine.GameCommand {
	idx := weightedChoice(weights)
	key := keys[idx]
	if opt := findOption(options, key); opt != nil {
		return selectCmd(key)
	}
	// Fallback: pick first available.
	if opt := pickFirst(options); opt != nil {
		return selectCmd(opt.Key)
	}
	return selectCmd("home")
}

// --- Hunter Strategy ---

type HunterStrategy struct{}

func (s *HunterStrategy) Name() string { return "hunter" }

func (s *HunterStrategy) Decide(screen string, options []engine.MenuOption) engine.GameCommand {
	switch {
	case screen == "main_menu":
		// 80% hunt, 10% harvest, 10% auto-play
		return mainMenuChoice(
			[]string{"3", "1", "8"},
			[]int{80, 10, 10},
			options,
		)
	case screen == "combat":
		// 50% attack, 15% skill, 10% defend, 10% item, 10% auto, 5% flee
		return combatDecision([6]int{50, 15, 10, 10, 10, 5}, options)
	case strings.HasPrefix(screen, "combat_"):
		return handleCombatSubScreen(screen, options)
	case screen == "autoplay_speed":
		return selectCmd("4") // turbo
	case screen == "autoplay_menu":
		return selectCmd("0") // return to main menu
	default:
		return handleCommonScreen(screen, options)
	}
}

// --- Harvester Strategy ---

type HarvesterStrategy struct{}

func (s *HarvesterStrategy) Name() string { return "harvester" }

func (s *HarvesterStrategy) Decide(screen string, options []engine.MenuOption) engine.GameCommand {
	switch {
	case screen == "main_menu":
		// 40% harvest, 15% village, 15% hunt, 15% town, 15% auto-play
		return mainMenuChoice(
			[]string{"1", "10", "3", "11", "8"},
			[]int{40, 15, 15, 15, 15},
			options,
		)
	case screen == "combat":
		// Auto-fight for efficiency
		return selectCmd("6")
	case strings.HasPrefix(screen, "combat_"):
		return handleCombatSubScreen(screen, options)
	case screen == "autoplay_speed":
		return selectCmd("4") // turbo
	case screen == "autoplay_menu":
		return selectCmd("0")
	default:
		return handleCommonScreen(screen, options)
	}
}

// --- Dungeon Crawler Strategy ---

type DungeonCrawlerStrategy struct{}

func (s *DungeonCrawlerStrategy) Name() string { return "dungeon_crawler" }

func (s *DungeonCrawlerStrategy) Decide(screen string, options []engine.MenuOption) engine.GameCommand {
	switch {
	case screen == "main_menu":
		// 60% dungeon, 20% hunt, 10% auto-play, 10% harvest
		return mainMenuChoice(
			[]string{"12", "3", "8", "1"},
			[]int{60, 20, 10, 10},
			options,
		)
	case screen == "combat":
		// 40% attack, 30% skill, 15% auto, 15% item (no flee)
		return combatDecision([6]int{40, 30, 0, 15, 15, 0}, options)
	case strings.HasPrefix(screen, "combat_"):
		return handleCombatSubScreen(screen, options)
	case strings.HasPrefix(screen, "dungeon_"):
		return handleDungeonScreen(screen, options)
	case screen == "autoplay_speed":
		return selectCmd("4")
	case screen == "autoplay_menu":
		return selectCmd("0")
	default:
		return handleCommonScreen(screen, options)
	}
}

// --- Arena Grinder Strategy ---

type ArenaGrinderStrategy struct {
	turnCount int
}

func (s *ArenaGrinderStrategy) Name() string { return "arena_grinder" }

func (s *ArenaGrinderStrategy) Decide(screen string, options []engine.MenuOption) engine.GameCommand {
	s.turnCount++
	switch {
	case screen == "main_menu":
		// Early game: level via auto-play first, then arena
		if s.turnCount < 100 {
			// 60% auto-play, 20% hunt, 20% harvest
			return mainMenuChoice(
				[]string{"8", "3", "1"},
				[]int{60, 20, 20},
				options,
			)
		}
		// Later: 60% arena, 20% auto-play, 20% hunt
		return mainMenuChoice(
			[]string{"14", "8", "3"},
			[]int{60, 20, 20},
			options,
		)
	case screen == "combat":
		// 45% attack, 30% skill, 10% defend, 15% item (no flee in arena)
		return combatDecision([6]int{45, 30, 10, 15, 0, 0}, options)
	case strings.HasPrefix(screen, "combat_"):
		return handleCombatSubScreen(screen, options)
	case strings.HasPrefix(screen, "arena_"):
		return handleArenaScreen(screen, options)
	case screen == "autoplay_speed":
		return selectCmd("4")
	case screen == "autoplay_menu":
		return selectCmd("0")
	default:
		return handleCommonScreen(screen, options)
	}
}

// --- Completionist Strategy ---

type CompletionistStrategy struct {
	cycle int
}

func (s *CompletionistStrategy) Name() string { return "completionist" }

func (s *CompletionistStrategy) Decide(screen string, options []engine.MenuOption) engine.GameCommand {
	switch {
	case screen == "main_menu":
		// Cycle through activities in order
		activities := []string{"3", "1", "8", "10", "11", "12", "9", "13", "14"}
		key := activities[s.cycle%len(activities)]
		s.cycle++
		if opt := findOption(options, key); opt != nil {
			return selectCmd(key)
		}
		// Skip unavailable activities
		if opt := pickFirst(options); opt != nil {
			return selectCmd(opt.Key)
		}
		return selectCmd("home")
	case screen == "combat":
		// Balanced: auto-fight
		return selectCmd("6")
	case strings.HasPrefix(screen, "combat_"):
		return handleCombatSubScreen(screen, options)
	case strings.HasPrefix(screen, "dungeon_"):
		return handleDungeonScreen(screen, options)
	case strings.HasPrefix(screen, "arena_"):
		return handleArenaScreen(screen, options)
	case screen == "autoplay_speed":
		return selectCmd("4")
	case screen == "autoplay_menu":
		return selectCmd("0")
	default:
		return handleCommonScreen(screen, options)
	}
}

// --- Shared screen handlers ---

// handleCommonScreen handles screens shared across all strategies.
func handleCommonScreen(screen string, options []engine.MenuOption) engine.GameCommand {
	switch screen {
	case "character_select":
		// Pick first character
		if opt := pickFirst(options); opt != nil {
			return selectCmd(opt.Key)
		}
		return selectCmd("home")

	case "character_create":
		// Should not happen -- agents create characters during setup
		return inputCmd("AgentChar")

	case "harvest_select":
		// Pick a random resource
		if opt := pickRandom(options); opt != nil {
			return selectCmd(opt.Key)
		}
		return selectCmd("home")

	case "hunt_location_select":
		// Pick a random non-locked location
		var nonLocked []engine.MenuOption
		for _, opt := range options {
			if opt.Enabled && !strings.HasPrefix(opt.Key, "locked:") {
				nonLocked = append(nonLocked, opt)
			}
		}
		if len(nonLocked) > 0 {
			pick := nonLocked[rand.Intn(len(nonLocked))]
			return selectCmd(pick.Key)
		}
		// Try locked locations if no unlocked ones
		if opt := pickRandom(options); opt != nil {
			return selectCmd(opt.Key)
		}
		return selectCmd("home")

	case "hunt_tracking":
		// Pick random target (option "0") or random specific target
		return selectCmd("0")

	case "quest_log", "player_stats", "discovered_locations":
		return selectCmd("back")

	case "guide_main", "guide_combat", "guide_skills", "guide_village",
		"guide_crafting", "guide_monster_drops", "guide_autoplay", "guide_quests":
		return selectCmd("back")

	// Village screens
	case "village_main":
		return handleVillageScreen(options)
	case "village_view_villagers", "village_assign_task", "village_assign_resource",
		"village_batch_assign", "village_hire_guard", "village_crafting",
		"village_craft_potion", "village_craft_armor", "village_craft_weapon",
		"village_upgrade_skill", "village_upgrade_confirm", "village_craft_scrolls",
		"village_build_defense", "village_build_walls", "village_craft_traps",
		"village_view_defenses", "village_check_tide", "village_monster_tide",
		"village_tide_wave", "village_manage_guards", "village_manage_guard",
		"village_equip_guard", "village_unequip_guard", "village_give_item",
		"village_take_item", "village_heal_guard", "village_fortifications",
		"village_training", "village_healing":
		return defaultNavigation(screen, options)

	// Town screens
	case "town_main":
		return handleTownScreen(options)
	case "town_inn", "town_inn_sleep", "town_inn_hire_guard", "town_inn_view_guests",
		"town_inn_gossip", "town_inn_gamble", "town_inn_gamble_play", "town_inn_hire_fighter",
		"town_mayor", "town_mayor_challenge", "town_mayor_menu", "town_mayor_set_tax",
		"town_mayor_create_quest", "town_mayor_create_quest_amount",
		"town_mayor_create_quest_reward", "town_mayor_hire_guard", "town_mayor_hire_monster",
		"town_fetch_quests", "town_talk_npc", "town_npc_dialogue",
		"town_npc_quest_board", "town_npc_quest_detail",
		"town_npc_quest_accept", "town_npc_quest_turn_in":
		return defaultNavigation(screen, options)

	// Bounty screens
	case "most_wanted_board", "most_wanted_hunt":
		return defaultNavigation(screen, options)

	default:
		return defaultNavigation(screen, options)
	}
}

// handleCombatSubScreen handles combat sub-menus (item select, skill select, etc.).
func handleCombatSubScreen(screen string, options []engine.MenuOption) engine.GameCommand {
	switch screen {
	case "combat_item_select":
		// Use first available item, or go back
		if opt := pickFirst(options); opt != nil {
			return selectCmd(opt.Key)
		}
		return selectCmd("back")
	case "combat_skill_select":
		// Use first available skill, or go back
		if opt := pickFirst(options); opt != nil {
			return selectCmd(opt.Key)
		}
		return selectCmd("back")
	case "combat_guard_prompt":
		// Bring guards if available
		if opt := findOption(options, "y"); opt != nil {
			return selectCmd("y")
		}
		return selectCmd("n")
	case "combat_skill_reward":
		// Accept skill reward (take scroll)
		if opt := findOption(options, "take"); opt != nil {
			return selectCmd("take")
		}
		if opt := pickFirst(options); opt != nil {
			return selectCmd(opt.Key)
		}
		return selectCmd("home")
	default:
		return defaultNavigation(screen, options)
	}
}

// handleDungeonScreen navigates dungeon screens.
func handleDungeonScreen(screen string, options []engine.MenuOption) engine.GameCommand {
	switch screen {
	case "dungeon_select":
		// Pick first dungeon
		if opt := pickFirst(options); opt != nil {
			return selectCmd(opt.Key)
		}
		return selectCmd("home")
	case "dungeon_floor_map":
		// Proceed into the dungeon
		if opt := findOption(options, "enter"); opt != nil {
			return selectCmd("enter")
		}
		if opt := pickFirst(options); opt != nil {
			return selectCmd(opt.Key)
		}
		return selectCmd("home")
	case "dungeon_grid_move":
		// Move in a random direction
		dirs := []string{"n", "s", "e", "w"}
		for _, d := range dirs {
			if opt := findOption(options, d); opt != nil {
				return selectCmd(d)
			}
		}
		if opt := pickFirst(options); opt != nil {
			return selectCmd(opt.Key)
		}
		return selectCmd("home")
	case "dungeon_room", "dungeon_treasure", "dungeon_trap", "dungeon_rest", "dungeon_merchant":
		// Pick first option (loot, proceed, etc.)
		if opt := pickFirst(options); opt != nil {
			return selectCmd(opt.Key)
		}
		return selectCmd("home")
	case "dungeon_complete", "dungeon_defeat":
		// Return to main menu
		return selectCmd("home")
	default:
		return defaultNavigation(screen, options)
	}
}

// handleArenaScreen navigates arena screens.
func handleArenaScreen(screen string, options []engine.MenuOption) engine.GameCommand {
	switch screen {
	case "arena_main":
		// Challenge someone
		if opt := findOption(options, "challenge"); opt != nil {
			return selectCmd("challenge")
		}
		if opt := pickFirst(options); opt != nil {
			return selectCmd(opt.Key)
		}
		return selectCmd("home")
	case "arena_challenge":
		// Pick first opponent
		if opt := pickFirst(options); opt != nil {
			return selectCmd(opt.Key)
		}
		return selectCmd("back")
	case "arena_confirm":
		// Confirm the fight
		if opt := findOption(options, "y"); opt != nil {
			return selectCmd("y")
		}
		if opt := pickFirst(options); opt != nil {
			return selectCmd(opt.Key)
		}
		return selectCmd("home")
	default:
		return defaultNavigation(screen, options)
	}
}

// handleVillageScreen picks a village action.
func handleVillageScreen(options []engine.MenuOption) engine.GameCommand {
	if opt := pickRandom(options); opt != nil {
		return selectCmd(opt.Key)
	}
	return selectCmd("home")
}

// handleTownScreen picks a town action.
func handleTownScreen(options []engine.MenuOption) engine.GameCommand {
	if opt := pickRandom(options); opt != nil {
		return selectCmd(opt.Key)
	}
	return selectCmd("home")
}

// --- Village Manager Strategy ---

type VillageManagerStrategy struct {
	turnCount int
}

func (s *VillageManagerStrategy) Name() string { return "village_manager" }

func (s *VillageManagerStrategy) Decide(screen string, options []engine.MenuOption) engine.GameCommand {
	s.turnCount++
	switch {
	case screen == "main_menu":
		// 50% village, 25% hunt, 15% harvest, 10% auto-play
		return mainMenuChoice(
			[]string{"10", "3", "1", "8"},
			[]int{50, 25, 15, 10},
			options,
		)
	case screen == "combat":
		// Auto-fight for efficiency
		return selectCmd("6")
	case strings.HasPrefix(screen, "combat_"):
		return handleCombatSubScreen(screen, options)
	case screen == "village_main":
		return handleVillageManagerVillageScreen(options)
	case screen == "village_hire_guard":
		// Pick first affordable guard, else back
		if opt := pickFirst(options); opt != nil {
			return selectCmd(opt.Key)
		}
		return selectCmd("back")
	case screen == "village_assign_task":
		// Pick first available villager, else back
		if opt := pickFirst(options); opt != nil {
			return selectCmd(opt.Key)
		}
		return selectCmd("back")
	case screen == "village_batch_assign":
		// Pick random resource
		if opt := pickRandom(options); opt != nil {
			return selectCmd(opt.Key)
		}
		return selectCmd("back")
	case screen == "village_assign_resource":
		// Pick random resource
		if opt := pickRandom(options); opt != nil {
			return selectCmd(opt.Key)
		}
		return selectCmd("back")
	case screen == "village_build_defense":
		// Build walls
		if opt := findOption(options, "1"); opt != nil {
			return selectCmd("1")
		}
		return selectCmd("back")
	case screen == "village_build_walls", screen == "village_craft_traps":
		// Pick first available
		if opt := pickFirst(options); opt != nil {
			return selectCmd(opt.Key)
		}
		return selectCmd("back")
	case screen == "village_monster_tide":
		// Start defense
		if opt := pickFirst(options); opt != nil {
			return selectCmd(opt.Key)
		}
		return selectCmd("back")
	case screen == "village_tide_wave":
		// Advance waves
		if opt := findOption(options, "next"); opt != nil {
			return selectCmd("next")
		}
		if opt := pickFirst(options); opt != nil {
			return selectCmd(opt.Key)
		}
		return selectCmd("back")
	case screen == "village_tide_complete":
		// Back to village
		return selectCmd("back")
	case strings.HasPrefix(screen, "village_"):
		return defaultNavigation(screen, options)
	case screen == "autoplay_speed":
		return selectCmd("4") // turbo
	case screen == "autoplay_menu":
		return selectCmd("0")
	default:
		return handleCommonScreen(screen, options)
	}
}

// handleVillageManagerVillageScreen picks village actions with weighted priorities.
func handleVillageManagerVillageScreen(options []engine.MenuOption) engine.GameCommand {
	// Weighted village actions: defend tide (25%), hire guards (20%), assign tasks (20%),
	// build defenses (15%), crafting (10%), manage guards (5%), view (3%), back (2%)
	keys := []string{"7", "3", "2", "5", "4", "8", "1", "0"}
	weights := []int{25, 20, 20, 15, 10, 5, 3, 2}
	return mainMenuChoice(keys, weights, options)
}

// NewStrategy creates a Strategy by name. Returns nil for unknown names.
func NewStrategy(name string) Strategy {
	switch name {
	case "hunter":
		return &HunterStrategy{}
	case "harvester":
		return &HarvesterStrategy{}
	case "dungeon_crawler":
		return &DungeonCrawlerStrategy{}
	case "arena_grinder":
		return &ArenaGrinderStrategy{}
	case "completionist":
		return &CompletionistStrategy{}
	case "village_manager":
		return &VillageManagerStrategy{}
	default:
		return nil
	}
}
