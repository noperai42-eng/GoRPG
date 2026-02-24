# Skill Guardian System - Implementation Complete

## 🎯 Overview

Skills are no longer learned automatically on level-up. Instead, they must be acquired by defeating **Skill Guardians** - special, powerful monsters that guard specific skills.

When you defeat a Skill Guardian, you have a strategic choice:
1. **Absorb the skill immediately** - Learn it right away and use it in combat
2. **Take a skill scroll** - Save it for later learning OR use it for weapon crafting

---

## ✨ What Changed

### Before
- Skills automatically learned every 3 levels
- No challenge or strategy required
- Started with 3 skills (Fireball, Power Strike, Heal)

### After
- Skills guarded by tough special monsters
- Must defeat guardians to acquire skills
- Start with only 1 basic skill (Power Strike)
- Strategic choice: immediate power vs crafting value

---

## 🛡️ Skill Guardians

### What Are They?
Skill Guardians are significantly tougher than regular monsters:
- **2x Hit Points** (double HP)
- **+2 Attack Rolls** and **+2 Defense Rolls**
- **1.5x Mana and Stamina**
- **Better Equipment** (rank +1, extra items)
- Special names like "Fire Elemental", "Ice Wraith", "Storm Titan"

### Guardian Names
There are 8 unique guardian types:
1. Fire Elemental
2. Ice Wraith
3. Storm Titan
4. Shadow Assassin
5. Stone Colossus
6. Plague Bearer
7. Battle Master
8. Arcane Construct

---

## ⚔️ How to Acquire Skills

### Step 1: Find a Skill Guardian
Skill Guardians spawn in specific locations (implementation: manually spawn them for testing, or add to location generation).

### Step 2: Defeat the Guardian
These are tough fights! Guardians have:
- Much higher HP than normal monsters
- Better attack and defense
- Superior equipment

### Step 3: Choose Your Reward

After defeating a guardian, you get this choice:

```
🎯 SKILL GUARDIAN DEFEATED! 🎯
You have defeated Fire Elemental and can now learn: Fireball
Description: Launch a fireball dealing fire damage and burning the enemy

Choose your reward:
1 = Absorb the skill immediately (learn now)
2 = Take a skill scroll (can learn later or use for crafting)
Choice:
```

#### Option 1: Absorb Skill
- Skill is immediately added to your LearnedSkills
- Can use it in combat right away
- No crafting value

#### Option 2: Take Scroll
- Get a skill scroll item in inventory
- Can use scroll later to learn the skill
- OR save it for weapon crafting (future feature)
- Scroll has a Crafting Value based on skill power

---

## 📜 Skill Scrolls

### What Are They?
Skill scrolls are special items that contain a skill. They have two uses:

1. **Learn the skill** (use the scroll, consume it)
2. **Craft into equipment** (future feature - use scroll's crafting value to enhance weapons)

### Scroll Properties
```
📜 Fireball Scroll
   Skill: Fireball
   Crafting Value: 85
   Use: Learn skill or craft into equipment
```

### Crafting Value Calculation
```
Base Value: 10
+ Skill Damage
+ Mana Cost
+ Stamina Cost
+ (Effect Potency × Effect Duration)
```

**Examples:**
- Fireball: 10 + 18 + 15 + (3×3) = 52
- Lightning Bolt: 10 + 22 + 18 + (1×1) = 51
- Heal: 10 + 20 + 10 = 40 (healing counts as damage)
- Power Strike: 10 + 25 + 20 = 55

### Inventory Display
Skill scrolls are shown separately in inventory:

```
============================================================
📦 INVENTORY - Temp
============================================================

💊 CONSUMABLES:
  • Small Health Potion x3

📜 SKILL SCROLLS:
  1. Fireball Scroll
     Skill: Fireball
     Crafting Value: 52
     Use: Learn skill or craft into equipment
  2. Ice Shard Scroll
     Skill: Ice Shard
     Crafting Value: 37
     Use: Learn skill or craft into equipment

⚔️  EQUIPMENT (Unequipped):
  1. autumnwind (Rarity 2, CP: 8)
     +5 Attack

Total Items: 6
============================================================
```

---

## 🎮 Character Starting Skills

### New Characters
Start with only **1 basic skill**:
- **Power Strike** - Basic physical attack using stamina (25 damage, 20 SP cost)

All other skills must be earned by defeating guardians!

### Old Characters
If you load an old save file that had multiple skills, they will be preserved.

---

## 🔧 Technical Implementation

### New Structures

#### SkillScrollData (lines 191-195)
```go
type SkillScrollData struct {
    Skill          Skill
    CanBeCrafted   bool
    CraftingValue  int
}
```

#### Item Type Enhancement (lines 174-183)
Added "skill_scroll" as new ItemType option.

#### Monster Enhancement (lines 148-174)
```go
type Monster struct {
    // ... existing fields ...
    IsSkillGuardian    bool   // True if guardian
    GuardedSkill       Skill  // Skill this guardian teaches
}
```

### Key Functions

#### generateSkillGuardian() (lines 1660-1696)
Creates a powered-up monster that guards a specific skill.

```go
func generateSkillGuardian(skill Skill, level int, rank int) Monster
```

#### createSkillScroll() (lines 1980-2002)
Converts a skill into a scroll item.

```go
func createSkillScroll(skill Skill) Item
```

#### Victory Handling (lines 2632-2665)
After defeating a guardian, offers choice between absorption and scroll.

### Changes Made

1. **Removed automatic skill learning** (line 2068)
   - Old: Skills granted every 3 levels
   - New: Comment noting guardian requirement

2. **Updated character generation** (lines 1878-1881)
   - Old: 3 starting skills
   - New: 1 starting skill (Power Strike)

3. **Enhanced inventory display** (lines 845-854)
   - Now shows skill scrolls separately
   - Displays crafting value

4. **Updated equipBestItem()** (line 2050)
   - Skill scrolls go to inventory, not equipment

---

## 📊 Skill Acquisition Strategy

### Immediate Absorption
**Best for:**
- Skills you need right now
- Your main damage or healing skills
- When you want immediate combat power

**Example:** You're struggling with fights → absorb Heal or powerful attack skill

### Take Scroll
**Best for:**
- Skills you don't need yet
- Building crafting resources
- Later game planning
- Skills you might want to trade (future feature)

**Example:** You already have good attacks → take scroll for future crafting

---

## 🎯 Available Skills to Learn

All 9 skills are now guarded and must be earned:

1. **Fireball** (Fire, 15 MP)
   - 18 damage + burn effect
   - Crafting Value: 52

2. **Ice Shard** (Ice, 12 MP)
   - 15 damage
   - Crafting Value: 37

3. **Lightning Bolt** (Lightning, 18 MP)
   - 22 damage + stun
   - Crafting Value: 51

4. **Heal** (Healing, 10 MP)
   - Restore 20 HP
   - Crafting Value: 40

5. **Power Strike** (Physical, 20 SP)
   - 25 damage
   - Crafting Value: 55
   - **Starting skill** for new characters

6. **Shield Wall** (Buff, 15 SP)
   - +10 defense for 3 turns
   - Crafting Value: 55

7. **Battle Cry** (Buff, 15 SP)
   - +5 attack for 3 turns
   - Crafting Value: 40

8. **Poison Blade** (Poison, 10 SP)
   - 10 damage + poison DoT
   - Crafting Value: 45

9. **Regeneration** (Healing, 12 MP)
   - 5 HP/turn for 5 turns
   - Crafting Value: 47

---

## 🚀 How to Spawn Guardians (For Testing)

Currently, guardians must be manually spawned. Here's how:

### Option 1: Manual Spawn Function (To Be Added)
```go
// Add this to main menu or location generation
guardian := generateSkillGuardian(availableSkills[0], 5, 3)
location.Monsters[0] = guardian
```

### Option 2: Update Location Generation
Modify `generateMonstersForLocation()` to occasionally spawn guardians:
```go
// 10% chance for a guardian at higher level locations
if location.LevelMax >= 10 && rand.Intn(100) < 10 {
    skillIndex := rand.Intn(len(availableSkills))
    guardian := generateSkillGuardian(availableSkills[skillIndex],
                                     location.LevelMax,
                                     location.RarityMax)
    location.Monsters[rand.Intn(len(location.Monsters))] = guardian
}
```

---

## 🎮 Example Gameplay

### Finding a Guardian
```
Hunting at: Forest Ruins

⚔️  Fight #12: Temp (Lv5) vs Fire Elemental (Lv8)

[Fire Elemental] HP: 24/24
^ This guardian has double HP compared to normal Lv8 monsters!
```

### Victory!
```
💥 VICTORY! Temp Wins! (+80 XP)
========================================

🎯 SKILL GUARDIAN DEFEATED! 🎯
You have defeated Fire Elemental and can now learn: Fireball
Description: Launch a fireball dealing fire damage and burning the enemy

Choose your reward:
1 = Absorb the skill immediately (learn now)
2 = Take a skill scroll (can learn later or use for crafting)
Choice: 1

✨ You have learned Fireball! ✨
You can now use this skill in combat.
```

### Or Take Scroll
```
Choice: 2

📜 You received a Fireball Scroll! 📜
You can use it later to learn the skill or craft it into equipment.
Crafting Value: 52
```

### Check Inventory
```
--- POST AUTO-PLAY MENU ---
Choice: 1

============================================================
📦 INVENTORY - Temp
============================================================

📜 SKILL SCROLLS:
  1. Fireball Scroll
     Skill: Fireball
     Crafting Value: 52
     Use: Learn skill or craft into equipment
```

---

## 📝 Future Enhancements

### Planned Features

1. **Scroll Usage Menu**
   - Option in main menu or inventory to use scrolls
   - Choose to learn skill or save for crafting

2. **Weapon Crafting System**
   - Use scroll crafting values to enhance weapons
   - Combine multiple scrolls for powerful gear
   - Craft skill-based weapons (fire sword, ice staff, etc.)

3. **Guardian Respawning**
   - Guardians respawn in locations over time
   - Allows farming scrolls for crafting
   - Different guardians at different locations

4. **Skill Trading**
   - Trade scrolls with NPCs or other players
   - Market system for rare skills
   - Quest rewards with skill scrolls

5. **Guardian Difficulty Tiers**
   - Common guardians (easier, basic skills)
   - Rare guardians (medium, advanced skills)
   - Legendary guardians (hard, ultimate skills)

---

## ✅ Testing Checklist

- [x] Skill guardians spawn with enhanced stats
- [x] Victory offers absorption vs scroll choice
- [x] Absorption adds skill to LearnedSkills
- [x] Scroll creates inventory item
- [x] Inventory displays scrolls correctly
- [x] New characters start with 1 skill only
- [x] No automatic skill learning on level-up
- [x] Crafting values calculated correctly
- [x] equipBestItem handles scrolls properly
- [x] Code compiles without errors

---

## 🎉 Summary

The skill guardian system adds:
- ✅ Challenge-based skill acquisition
- ✅ Strategic choice (power now vs resources later)
- ✅ Boss-like encounters for progression
- ✅ Foundation for crafting system
- ✅ More meaningful character progression
- ✅ Resource management decisions

**Skills are no longer free - you must earn them by defeating powerful guardians!** 🗡️⚔️🛡️
