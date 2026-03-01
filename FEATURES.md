# Feature Roadmap

This document lists potential features to enhance the game, inspired by classic BBS door games like Usurper, Legend of the Red Dragon (LORD), Trade Wars, and modern roguelikes.

## Already Planned (from code comments)

These features are mentioned in TODO comments in fight-cli.go:

1. **Client/Server Setup** (line 15)
   - Multi-player capability
   - Network play support
   - Persistent world state across multiple players

2. **Location Type System** (line 17)
   - Map location names to types
   - Spawn appropriate mob types per location (Beasts, Monsters, Animals)
   - Location-specific encounter tables

3. **NPC Tiers** (line 23)
   - Common, Rare, Elite, Mini Boss, Boss, King, Emperor
   - Each tier grants additional stats and difficulty multipliers
   - Special loot tables per tier

4. **Skills System** (line 26)
   - Recruit: Hire NPCs to join your party
   - Tame: Capture monsters as companions
   - Allow bases to have defensive NPC groups

5. **Gate Keepers** (line 31)
   - NPCs that block access to high-level areas
   - Require defeating them or meeting conditions to progress

6. **Merchants** (line 33)
   - Buy/sell items and resources
   - Trading economy
   - Special/unique items for sale

7. **Location Mob Combat Loop** (line 21)
   - Monsters fight each other at locations
   - Ecosystem simulation where mobs level up independently
   - Dynamic difficulty based on mob interactions

## Core Gameplay Enhancements

### Combat System

8. **Magic/Skills System**
   - Spells with mana cost
   - Special abilities (critical hits, dodges, parries)
   - Cooldown-based abilities
   - Status effects (poison, burn, freeze, stun)

9. **Combat Tactics**
   - Attack, Defend, Use Item, Run Away options
   - Defending provides temporary defense bonus
   - Running has success chance based on level difference

10. **Party System**
   - Recruit multiple companions
   - Formation/positioning mechanics
   - Party-wide buffs and abilities
   - NPC AI for party members

11. **Enemy Variety**
   - Special enemy abilities (healing, summoning, ranged)
   - Boss mechanics with multiple phases
   - Elite packs with affixes
   - Rare spawns with better loot

### Character Development

12. **Class System**
   - Multiple character classes (Warrior, Mage, Rogue, Cleric)
   - Class-specific abilities and stat growth
   - Multiclassing or prestige classes

13. **Skill Trees**
   - Unlock abilities as you level
   - Specialization paths
   - Passive bonuses

14. **Attributes System**
   - Strength, Dexterity, Constitution, Intelligence, Wisdom, Charisma
   - Stat points to allocate on level up
   - Attribute-based skill checks

15. **Reputation/Fame System**
   - Track victories and achievements
   - Unlock special NPCs and quests at milestones
   - Leaderboards for competitive play

### Economy & Crafting

16. **Enhanced Merchant System**
   - Multiple merchants with different inventories
   - Reputation affects prices
   - Rare item rotations (daily/weekly)
   - Black market for special goods

17. **Crafting System**
   - Combine resources to create items
   - Crafting recipes to discover
   - Quality tiers (Normal, Fine, Masterwork, Legendary)
   - Enchanting/upgrading equipment

18. **Resource Improvements**
   - Resource nodes that deplete and regenerate
   - Gathering skills that level up
   - Special resources from specific locations
   - Resource conversion/refining

19. **Banking System**
   - Store gold safely
   - Interest accrual
   - Loans for early-game progression

### World & Exploration

20. **Quest System**
   - Story quests with dialogue
   - Daily quests for resources/experience
   - Repeatable bounties
   - Quest chains with progressive difficulty

21. **Random Events**
   - Encounters while traveling
   - Treasure chests
   - Ambushes
   - Friendly travelers offering trades/information

22. **Dungeon Crawling**
   - Multi-level dungeons
   - Random generation or hand-crafted
   - Keys, locked doors, traps
   - Dungeon-specific boss encounters

23. **World Map Improvements**
   - Travel time between locations
   - Random encounters during travel
   - Fast travel after discovering waypoints
   - Hidden locations requiring special conditions

24. **Town Features**
   - Inn (rest to restore HP/MP, limited uses per day)
   - Tavern (rumors, quests, gambling)
   - Training Hall (respec skills/stats for a cost)
   - Arena (PvP or vs progressively harder enemies)

### Social & Competitive

25. **PvP System**
   - Challenge other players
   - Rankings and seasons
   - Betting on matches
   - Safe zones vs PvP zones

26. **Guilds/Clans**
   - Player groups with shared resources
   - Guild halls with upgradeable facilities
   - Guild wars and territory control
   - Cooperative guild quests

27. **Mail/Messaging**
   - Send messages to other players
   - Trade items via mail
   - Gift system

28. **Leaderboards**
   - Highest level
   - Most gold
   - Arena victories
   - Fastest dungeon completions

### Base Building (Expansion)

29. **Enhanced Building System**
   - Visual base layout
   - Multiple building types (Barracks, Workshop, Farm, Mine)
   - Buildings generate passive resources
   - Upgrade buildings for better benefits

30. **Base Defense**
   - Defend against monster raids
   - NPC defenders
   - Tower defense mini-game
   - Rewards for successful defenses

31. **Base Visitors**
   - Random NPCs visit offering quests/trades
   - Recruit NPCs as defenders or companions
   - Special events at your base

### Quality of Life

32. **Daily Turns System** (BBS Door Game Classic)
   - Limited actions per day (e.g., 100 forest fights)
   - Encourages daily engagement
   - Turns refresh at midnight
   - Purchase extra turns with in-game currency

33. **Auto-Combat Options**
   - Auto-battle weaker enemies
   - Speed up repeated fights
   - Battle logs/summaries

34. **Inventory Management**
   - Sort by type, rarity, CP
   - Sell multiple items at once
   - Item comparison tooltips
   - Inventory limits with storage solutions

35. **Better Save System**
   - Multiple save slots
   - Auto-save intervals
   - Cloud save support
   - Character export/import

36. **Tutorial System**
   - New player guidance
   - Contextual tips
   - Achievement-based teaching

### Progression & Endgame

37. **Prestige/Rebirth System**
   - Reset to level 1 with permanent bonuses
   - New game+ difficulty
   - Prestige-only content and items

38. **Seasonal Events**
   - Holiday-themed content
   - Limited-time challenges
   - Exclusive rewards

39. **Achievement System**
   - Track accomplishments
   - Reward titles, cosmetics, bonuses
   - Steam-style achievement unlocks

40. **Endless Modes**
   - Survival mode (how long can you last?)
   - Wave defense
   - Procedurally generated endless dungeon

### Technical Improvements

41. **Better RNG System**
   - Use single seeded random generator
   - Loot roll improvements
   - Luck stat affecting random outcomes

42. **Configuration System**
   - Difficulty settings
   - Gameplay modifiers
   - Customizable game rules

43. **Modding Support**
   - External data files for monsters/items/locations
   - Plugin system for features
   - Community content support

44. **Statistics Tracking**
   - Detailed combat logs
   - Session statistics
   - Historical tracking (total monsters killed, etc.)

### Shipped in v0.4.16

- **Lunar Calendar** — Moon phases (8 phases across 30-day cycle), raid phase indicator (day >= 8), moon emoji in navbar
- **Global Tide Leader** — Persistent world boss raided by all villages, scales with undefeated streak, defeat rewards distributed to participants
- **Village Elder Rescue Quest** — Village access gated behind `quest_v0_elder`, elder found at Lake Ruins (20% per combat win), villager rescue also gated behind elder completion

### Data-Driven Fixes (from v0.4.12 metrics review)

45. **Fix: Wire up `RecordItemLooted` in combat resolution** [HIGH]
   - `items_looted` is always 0 because `RecordItemLooted()` is never called in the engine
   - `resolveCombatWin()` loops over `mob.EquipmentMap` and equips items, but the metrics call is missing
   - Without this, all item economy analysis is blind — cannot tune drop rates or rarity curves
   - Data: `items_looted: 0, items_by_rarity: {}` despite thousands of fights

46. **Fix: Add metrics instrumentation to `autoResolveCombat`** [HIGH]
   - 59% of fights use auto-resolve, which has zero calls to `RecordDamage`, `RecordSkillUse`, or `RecordStatusEffect`
   - Current metrics only reflect 41% of fights (manual combat path)
   - This means "only Power Strike used" and "only physical damage" may be inaccurate
   - Data: `auto_fight_rate: 0.5926`, `skill_usage: {"Power Strike": 1595}` (manual path only)

47. **Fix: Record actual damage types in `RecordDamage` calls** [HIGH]
   - Both `RecordDamage` calls hardcode `"physical"` regardless of actual damage type
   - Player skill damage and monster skill damage have no `RecordDamage` call at all
   - The 3:1 monster-to-player damage asymmetry (572k vs 185k) is partly explained by missing skill damage tracking
   - Data: `damage_by_type: {"physical": 757974}`, zero elemental damage recorded

48. **Fix: Reduce barrier to elemental skill acquisition** [HIGH]
   - Players start with only Power Strike (physical). All other skills gated behind Skill Guardian defeats
   - This is the root cause of three anomalies: only Power Strike used, zero elemental damage, and empty status effects
   - Options: (a) grant an elemental skill at creation, (b) guarantee early Skill Guardian encounter, (c) tutorial quest for first elemental skill
   - Data: `skill_usage: {"Power Strike": 1595}`, `status_effects: {}`, `damage_by_type: {"physical": 757974}`

49. **Fix: Rebalance will-o-wisp physical immunity** [HIGH]
   - Will-o-wisp has `Physical: 0.0` resistance (immune to all physical damage)
   - Combined with players only having physical skills, it is literally invincible
   - Options: (a) change physical resistance from 0.0 to 0.25, or (b) ensure elemental skills are available first (item 48)
   - Data: 1W / 411L (0.24% win rate across 412 fights)

### Data-Driven Fixes (from v0.4.16 metrics review)

50. **Fix: Bot account duplication on server restart** [MEDIUM]
   - Server creates new bot accounts with random suffixes on every restart
   - 79 total accounts (77 bot, 2 human) for only 7 bot names — ~11 restarts worth of duplicates
   - Creates 69 villages (most orphaned duplicates), polluting leaderboards and arena
   - Fix: Query existing `bot_%` accounts before creating new ones
   - Data: `accounts: 79, bot: 77, human: 2, villages: 69`

51. **Fix: Metrics snapshot frequency** [MEDIUM]
   - Only 1 unique metrics snapshot exists (from Feb 28), despite multiple deploys since
   - Cannot track trends or measure impact of balance changes
   - Server may not be writing periodic snapshots, or the ticker interval is too long
   - Data: `snapshots_in_gamedb: 1, last_snapshot: 2026-02-28`

52. **Fix: Rarity curve too steep — Epic/Legendary/Mythic unwinnable** [HIGH]
   - Mythic: 0% win rate (0W/34L), Legendary: 10.1% (7W/62L), Epic: 17.7% (23W/107L)
   - Combined with item 48 (no elemental skills), high-rarity monsters are effectively impossible
   - Options: (a) reduce stat scaling per rarity tier, (b) cap rarity based on player level, (c) add flee-on-death protection
   - Data: `mythic_wr: 0%, legendary_wr: 10.1%, epic_wr: 17.7%`

53. **Fix: Crit rates below design targets** [LOW]
   - Player crit rate: 8.5% (target 15%), Monster crit rate: 4.8% (target 10%)
   - May be due to defend actions or auto-resolve not tracking crits
   - Data: `player_crit_rate: 0.085, monster_crit_rate: 0.048`

## Priority Recommendations

Based on the current game state and common player expectations, here are suggested priorities:

### Phase 1: Core Mechanics Polish
- Quest system (basic fetch/kill quests)
- Merchant system (buy/sell items)
- Magic/skills system (at least 3-5 abilities per character)
- Better inventory management

### Phase 2: Player Engagement
- Daily turns system
- Achievement system
- Random events during exploration
- Inn/Tavern town features

### Phase 3: Depth & Variety
- Class system (3-4 classes)
- Crafting system
- Dungeon crawling with multi-level dungeons
- Enhanced NPC tiers

### Phase 4: Social & Longevity
- PvP system
- Leaderboards
- Prestige/rebirth system
- Seasonal events

### Phase 5: Multiplayer (Long-term)
- Client/Server implementation
- Guild system
- Persistent world state
- PvP arenas and rankings
