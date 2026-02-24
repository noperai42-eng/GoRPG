# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a text-based RPG/roguelike game written in Go featuring an advanced combat system with skills, magic, status effects, and elemental damage types. The game features character creation, location exploration, tactical turn-based combat with monsters, resource harvesting, and persistent save/load functionality through JSON serialization.

## Commands

### Building
```bash
cd fight-cli
go build -o fight-cli fight-cli.go
```

### Running
```bash
cd fight-cli
./fight-cli
```

Or directly with Go:
```bash
cd fight-cli
go run fight-cli.go
```

### Testing
No formal test suite. To test, run the application and interact with the menu system. Recommended test flow:
1. Create new character (option 0)
2. Go hunting at Training Hall (option 3)
3. Test all combat actions (Attack, Defend, Use Item, Use Skill, Flee)
4. Verify skills work correctly
5. Check status effects apply properly

## Architecture

### Core Game Loop
The game runs a single-threaded CLI event loop in `main()` that:
1. Loads game state from `gamestate.json` (or initializes new state)
2. Presents a menu of options (0-7)
3. Processes user input via switch statement
4. Auto-saves on certain actions ("exit", character creation, location discovery)

### Data Model

**GameState** (fight-cli.go:35-38)
- Central state container holding all game data
- `CharactersMap`: All player characters indexed by name
- `GameLocations`: All discoverable locations indexed by name

**Character** (fight-cli.go:40-66)
- Player entity with stats (level, HP, attack/defense rolls)
- **Resources:** HP, Mana (MP), Stamina (SP)
- Equipment system with 8 slots (0-7) stored in `EquipmentMap`
- `StatsMod`: Aggregate bonuses from equipped items
- `ResourceStorageMap`: Harvested resources (Lumber, Gold, Iron, Sand, Stone)
- `KnownLocations`: Discovered location names
- `builtBuildings`: Constructed buildings that provide stat bonuses
- `LearnedSkills`: Combat skills unlocked through leveling
- `StatusEffects`: Active buffs/debuffs
- `Resistances`: Elemental damage modifiers

**Monster** (fight-cli.go:121-145)
- Similar structure to Character but includes `Rank` (difficulty tier: 1-10)
- `MonsterType`: Base type for special ability assignment
- Higher rank = more attack/defense rolls, more equipment, more HP
- Spawned at locations and can level up from combat victories
- Each monster type has unique resistances and skills

**StatusEffect** (fight-cli.go:95-99)
- Temporary combat effects
- Types: "poison", "burn", "stun", "regen", "buff_attack", "buff_defense"
- Duration in turns, potency for damage/healing

**DamageType** (fight-cli.go:101-109)
- Enumeration: Physical, Fire, Ice, Lightning, Poison
- Used for elemental damage and resistance calculations

**Skill** (fight-cli.go:111-119)
- Combat abilities for players and monsters
- Cost in mana or stamina (or both)
- Can deal damage and/or apply status effects
- 9 player skills + 15+ monster-specific skills

**Location** (fight-cli.go:65-72)
- Contains 20 pre-generated monsters
- `Type`: "Base", "Mix", "Ruin", "Resource", "Trade"
- `LevelMax`/`RarityMax`: Caps for monster generation
- `Weight`: Used in weighted random discovery system

**Item** (fight-cli.go:147-156)
- Procedurally generated with adjective+noun names
- `ItemType`: "equipment" or "consumable"
- `Slot`: Equipment slot (0-7) for equipment, -1 for consumables
- `Rarity`: Determines number of stat bonuses
- `CP` (Combat Power): Sum of all stat mods, used for "best item" comparison
- `ConsumableEffect`: For healing potions and buff consumables

### Key Systems

**Combat** (fight-cli.go:1301-1651)
- **Turn-based tactical system** with player choices each turn:
  - **Attack**: Normal physical attack (15% crit chance)
  - **Defend**: +50% defense, 50% attack power (defensive stance)
  - **Use Item**: Consume health potions or other consumables
  - **Use Skill**: Cast spells or use abilities (costs mana/stamina)
  - **Flee**: Attempt to escape (success chance: 20-90% based on level difference)
- Each turn processes status effects first
- Displays HP/MP/SP for both combatants
- Shows active status effects with duration
- Elemental damage with resistance modifiers
- Monster AI (40% chance to use skills if available)
- Winner takes loser's equipment
- Players resurrect on death, monsters respawn with new random stats
- Experience gained = loser's level × 10 (for players) or × 100 (for monsters)
- 30% chance to find health potions after victory

**Status Effects System** (fight-cli.go:1077-1172)
- Processed at start of each turn
- Damage-over-time effects: poison, burn
- Healing-over-time: regen
- Buff effects: buff_attack, buff_defense
- Control effects: stun (skips turn entirely)
- Duration decreases each turn, removed when expired
- Visual indicators show active effects

**Elemental Damage** (fight-cli.go:1058-1074)
- `applyDamage()` function calculates final damage
- Resistance multipliers: 0.25x-0.5x (resistant), 1.0x (normal), 2.0x (weak)
- Slimes: weak to fire (2x), resistant to physical (0.5x)
- Golems: weak to lightning (2x), very resistant to physical (0.25x)
- Hiftiers: resistant to all magic (0.5x)
- Combat displays resistance feedback

**Skills System** (fight-cli.go:168-250 for definitions)
- 9 player skills learned progressively
- Starting skills: Fireball, Power Strike, Heal
- New skill every 3 levels
- Each monster type has 1-3 unique skills based on level
- Skills assigned via `assignMonsterSkills()` (fight-cli.go:750-861)

**Monster Special Abilities** (fight-cli.go:750-861)
- Slimes: Acid Spit (poison damage)
- Goblins: Backstab (physical burst)
- Orcs: War Cry (attack buff), Berserker Rage (high damage)
- Golems: Stone Skin (defense buff)
- Hiftiers: Mana Bolt (lightning), Mind Blast (stun)
- Kobolds: Fire Breath (burn effect)
- Kitpods: Regenerate (healing over time)

**Location Discovery** (fight-cli.go:577-595)
- Weighted random system using `discoverableLocations` array
- Each location has a `Weight` value (0-60)
- Random number generated, exact match triggers discovery
- Note: Current implementation only finds on exact match, not cumulative weight

**Item Generation** (fight-cli.go:949-1013)
- Equipment: Rarity determines number of stat rolls
- Each roll randomly adds to Attack/Defense/HitPoint mod
- `rollUntilSix()` mechanic: Keeps rolling on 6s, accumulating total
- Items auto-equip if better CP than current slot item
- Consumables: Health potions (small/medium/large) with fixed heal amounts
- `createHealthPotion()`: Creates potions with 15/30/50 HP healing
- New characters start with 3 small health potions

**Resource Harvesting** (fight-cli.go:1043-1055)
- 5 resource types defined in `resourceTypes` global
- Uses `rollUntilSix()` for harvest amount
- Resources stored in Character's `ResourceStorageMap`
- Used for building construction (case "7" in main loop)

### Important Globals

- `discoverableLocations` (fight-cli.go:272-286): All game locations with weights
- `monsterNames` (fight-cli.go:165): Available monster types
- `availableSkills` (fight-cli.go:168-250): 9 player skills with full definitions
- `ADJECTIVES` / `NOUNS` (fight-cli.go:252-270): Item name generation pools
- `resourceTypes` (fight-cli.go:163): "Lumber", "Gold", "Iron", "Sand", "Stone"
- `availableBuildings` (fight-cli.go:288-291): Buildings that can be constructed

### Save System

**Persistence** (fight-cli.go:526-560)
- `writeGameStateToFile()`: JSON marshals entire GameState
- `loadGameStateFromFile()`: Restores GameState from JSON
- Default filename: `gamestate.json`
- Auto-saves on exit and after certain actions
- Manual save via option 5 (Player Stats)

**Note:** Old save files (pre-combat update) will load but characters won't have new features like skills, mana, stamina. Recommend creating new character to experience full feature set.

### Common Patterns

**Stat Calculation**
- Natural stats (base) + Item mods (from equipment) = Total stats
- `calculateItemMods()` (fight-cli.go:966-980) aggregates equipment bonuses
- Applied on level up, equipment change, or monster generation

**Level Up Logic** (fight-cli.go:982-1012)
- Characters level when experience ≥ (level × 100)
- Grants HP/MP/SP roll, recalculates attack/defense rolls (level/10 + 1)
- Learns new skill every 3 levels
- Full HP/MP/SP restoration on level up
- Separate functions: `levelUp()` for players, `levelUpMob()` for monsters

**Equipment Management** (fight-cli.go:938-955)
- `equipBestItem()`: Auto-equips if new item CP > current CP
- Consumables bypass equipment logic and go directly to inventory
- Replaced equipment moves to inventory
- 8 equipment slots (head, body, legs, feet, hands, ring, amulet, weapon - implied)

**Combat AI** (fight-cli.go:1540-1600)
- Monsters have 40% chance to use skills instead of attacking
- Checks resource availability before using skill
- Simple AI: random skill selection if multiple available
- Falls back to normal attack if no resources or random roll fails

## Development Notes

- The game state is mutable throughout, passed by pointer to most functions
- Random number generation uses `rand.Seed(time.Now().UnixNano())` repeatedly
- No formal error handling for invalid user input in menu system
- Building system (option 7) is implemented but not included in main menu display
- Location discovery uses exact weight matching rather than cumulative ranges
- Combat system uses goto for stunned state handling (necessary for proper flow)
- Status effects are processed before actions each turn
- Mana and stamina fully restore at combat start (not between combats)
- Critical hit chances: 15% for players, 10% for monsters
- Monster skill usage is probabilistic (40% chance per turn if capable)

## Recent Major Changes (v2.0 - Combat Overhaul)

**Added Systems:**
- Complete tactical combat with 5 action types
- 9 player skills + 15+ monster-specific skills
- Mana and stamina resource management
- 6 types of status effects (poison, burn, stun, regen, buff_attack, buff_defense)
- 5 elemental damage types with resistance system
- Critical hit system (10-15% chance)
- Enhanced combat display with turn tracking and resource bars
- Monster AI with skill usage

**Breaking Changes:**
- Character and Monster structs significantly expanded
- Combat function completely rewritten (~350 lines)
- Level up functions updated to grant MP/SP
- generateMonster now assigns skills based on type
- Old save files compatible but won't have new features

**Performance:**
- No performance impact (turn-based, no loops)
- Compiled size: ~2.7MB
- Clean compile with no warnings
