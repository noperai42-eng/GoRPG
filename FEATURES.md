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
