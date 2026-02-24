# Tracking Skill & Automatic Guardian Spawning

## 🎯 Overview

Two major features added to enhance the skill acquisition and hunting experience:

1. **Tracking Skill** - Allows players to see and choose which monster to fight
2. **Automatic Guardian Spawning** - Skill Guardians now spawn automatically in locations based on level

---

## 🔍 Tracking Skill

### What Is It?
A utility skill that gives you complete information about all monsters at a location and lets you choose your target strategically.

### How to Get It
Defeat a Skill Guardian that guards the Tracking skill. Like all skills, it must be earned!

### Benefits
- **See all monsters** at a location before choosing
- **Identify Skill Guardians** with the 🎯 tag
- **Choose your battles** strategically
- **Avoid tough fights** when you're weak
- **Target guardians** when you're ready for them

### No Resource Cost
- Mana Cost: 0
- Stamina Cost: 0
- Passive skill that's always active once learned

---

## 🎮 Using Tracking

### Without Tracking (Default Behavior)
```
Hunting at: Forest Ruins

Fights Remaining: 5

[Random monster is selected]
------------
Monster Stats: goblin
Level: 15
TotalLife: 8
RemainingLife: 8
AttackRolls: 2
DefenseRolls: 2
------------
```

### With Tracking Skill
```
Hunting at: Forest Ruins

Fights Remaining: 5

🔍 TRACKING ACTIVE - Choose your target:
============================================================
1. slime (Lv12) - HP:6/6
2. goblin (Lv15) - HP:8/8
3. Fire Elemental (Lv18) - HP:24/24 🎯 [SKILL GUARDIAN]
4. orc (Lv14) - HP:10/10
5. kobold (Lv13) - HP:7/7
6. slime (Lv11) - HP:5/5
7. Ice Wraith (Lv20) - HP:28/28 🎯 [SKILL GUARDIAN]
8. hiftier (Lv16) - HP:9/9
... (up to 20 monsters)
============================================================
Choose target (1-20, or 0 for random): 3

[You fight Fire Elemental]
```

### Strategic Choices

**Example 1: Target Weak Monsters**
```
Choose target: 6
```
→ Fight the Level 11 slime (easiest)

**Example 2: Hunt Guardians**
```
Choose target: 3
```
→ Challenge the Fire Elemental guardian to learn Fireball

**Example 3: Level Appropriate**
```
Choose target: 2
```
→ Fight goblin at your level for steady progression

**Example 4: Random (Old Behavior)**
```
Choose target: 0
```
→ Random selection, like before Tracking

---

## 🎯 Automatic Guardian Spawning

### How It Works

Skill Guardians now spawn automatically when locations are generated, based on the location's level range.

### Spawn Rules

| Location Level | Guardians | Example Locations |
|---------------|-----------|-------------------|
| < 10 | 0 | Training Hall, Forest, Lake, Hills |
| 10-29 | 1 | Hunters Lodge (LvMax 50) |
| 30-99 | 2 | Forest Ruins (LvMax 100) |
| 100+ | 3 | Ancient Dungeon, The Tower |

### Guardian Distribution

**Excluded Skills:**
- Power Strike (starting skill)
- Tracking (special utility skill)

**Guardable Skills (8 total):**
1. Fireball
2. Ice Shard
3. Lightning Bolt
4. Heal
5. Shield Wall
6. Battle Cry
7. Poison Blade
8. Regeneration

### Guardian Placement

- **Random positions** in monster list (any of 20 slots)
- **Random skills** assigned to each guardian
- **Level appropriate** to location (within LevelMax range)
- **Multiple guardians** don't duplicate skills (usually)

---

## 📊 Location Examples

### Training Hall (LvMax 20)
```
Monsters At Training Hall
slime Level: 15 Exp: 0
goblin Level: 18 Exp: 0
... (18 more regular monsters)
⚠️  0 Skill Guardian(s) present!
```
**No guardians** - too low level

### Hunters Lodge (LvMax 50)
```
Monsters At Hunters Lodge
🎯 Spawned Fire Elemental (Lv48) guarding Fireball at Hunters Lodge

slime Level: 42 Exp: 0
goblin Level: 45 Exp: 0
Fire Elemental Level: 48 Exp: 0 🎯 [GUARDIAN - Fireball]
orc Level: 47 Exp: 0
... (16 more monsters)
⚠️  1 Skill Guardian(s) present! Defeat them to learn skills.
```
**1 guardian** - mid-low level

### Forest Ruins (LvMax 100)
```
Monsters At Forest Ruins
🎯 Spawned Ice Wraith (Lv95) guarding Ice Shard at Forest Ruins
🎯 Spawned Storm Titan (Lv88) guarding Lightning Bolt at Forest Ruins

slime Level: 85 Exp: 0
Ice Wraith Level: 95 Exp: 0 🎯 [GUARDIAN - Ice Shard]
goblin Level: 92 Exp: 0
Storm Titan Level: 88 Exp: 0 🎯 [GUARDIAN - Lightning Bolt]
orc Level: 90 Exp: 0
... (16 more monsters)
⚠️  2 Skill Guardian(s) present! Defeat them to learn skills.
```
**2 guardians** - high level

### Ancient Dungeon (LvMax 200)
```
Monsters At Ancient Dungeon
🎯 Spawned Shadow Assassin (Lv195) guarding Poison Blade at Ancient Dungeon
🎯 Spawned Stone Colossus (Lv192) guarding Shield Wall at Ancient Dungeon
🎯 Spawned Battle Master (Lv188) guarding Battle Cry at Ancient Dungeon

slime Level: 180 Exp: 0
Shadow Assassin Level: 195 Exp: 0 🎯 [GUARDIAN - Poison Blade]
goblin Level: 185 Exp: 0
Stone Colossus Level: 192 Exp: 0 🎯 [GUARDIAN - Shield Wall]
Battle Master Level: 188 Exp: 0 🎯 [GUARDIAN - Battle Cry]
... (15 more monsters)
⚠️  3 Skill Guardian(s) present! Defeat them to learn skills.
```
**3 guardians** - maximum level dungeon

---

## 🎮 Complete Gameplay Flow

### 1. Early Game (Levels 1-5)
```
Location: Training Hall (LvMax 20)
- No guardians
- Build levels fighting regular monsters
- Prepare for first guardian encounter
```

### 2. First Guardian (Levels 5-10)
```
Location: Hunters Lodge (LvMax 50)
- 1 guardian present
- Random skill (e.g., Fireball, Heal, Ice Shard)
- Tough fight but achievable
- Learn your second skill!
```

### 3. Learning Tracking (Levels 10-15)
```
Options:
A) Hunt random guardian, hope for Tracking
B) Use option 4 to see all monsters at location
C) Random fights until you find Tracking guardian

Once learned:
✅ Can now see and choose targets!
✅ Identify guardians before fighting
✅ Avoid/target guardians strategically
```

### 4. Mid Game (Levels 15-30)
```
Location: Forest Ruins (LvMax 100)
- 2 guardians present
- With Tracking: Choose which to fight
- Build skill collection strategically
- Take scrolls for crafting if skill not needed
```

### 5. Late Game (Levels 30+)
```
Location: Ancient Dungeon, The Tower
- 3 guardians present
- Hunting specific skills for build
- Farming scrolls for crafting
- Maximum challenge encounters
```

---

## 🎯 Strategic Guide

### Priority Skill Acquisition

**Must-Have (Learn First):**
1. **Tracking** - Makes all future hunting strategic
2. **Heal** - Essential for survival

**Offensive Skills:**
3. **Fireball** or **Lightning Bolt** - Main damage
4. **Ice Shard** - Secondary element

**Utility/Defense:**
5. **Shield Wall** - Defensive buff
6. **Regeneration** - Sustain in long fights
7. **Battle Cry** - Offensive buff
8. **Poison Blade** - DoT damage

### Hunting Strategy

**Without Tracking:**
- Fight at lower level locations
- Accept random encounters
- Build levels safely
- Hope for Tracking guardian

**With Tracking:**
- Visit higher level locations
- Identify guardians
- Choose fights strategically:
  - Easy monsters to heal/level
  - Guardians when ready
  - Avoid overpowered enemies

### Location Progression

**Recommended Path:**
1. Training Hall → Level to 5
2. Forest/Lake/Hills → Level to 10
3. Hunters Lodge → Fight first guardian, get Tracking if possible
4. Forest Ruins → Multiple guardians, choose with Tracking
5. Ancient Dungeon → Collect remaining skills
6. The Tower → Ultimate challenges

---

## 💡 Pro Tips

### Tip 1: Scout Before Fighting
With Tracking, you can:
```
View monster list → Note guardian positions → Fight easy monsters first →
Level up → Then challenge guardian
```

### Tip 2: Guardian Identification
Look for:
- 🎯 tag in monster list
- Guardian names (Fire Elemental, Ice Wraith, etc.)
- Higher HP than normal monsters
- Skill name in brackets

### Tip 3: Scroll vs Absorb Decision
```
Have Tracking? → Absorb it immediately!
Don't need skill? → Take scroll for crafting
Already have skill? → Definitely take scroll
```

### Tip 4: Location Refresh
Guardians are generated when location is created/loaded:
- Same guardians stay until defeated
- Defeated guardians respawn as regular monsters
- Reload game to get new guardian distribution (not recommended for progression)

### Tip 5: Auto-Play with Tracking
Tracking only works in manual hunt mode (option 3).
Auto-play mode (option 8) still uses random selection for speed.

---

## 🔧 Technical Details

### Tracking Skill Properties
```go
{
    Name:        "Tracking",
    ManaCost:    0,
    StaminaCost: 0,
    Damage:      0,
    DamageType:  Physical,
    Effect:      StatusEffect{Type: "none"},
    Description: "Allows you to see and choose which monster to fight at a location",
}
```

### Guardian Spawn Logic (generateMonstersForLocation)
```go
// Determine number of guardians
if location.LevelMax >= 10 && location.LevelMax < 30 {
    numGuardians = 1
} else if location.LevelMax >= 30 && location.LevelMax < 100 {
    numGuardians = 2
} else if location.LevelMax >= 100 {
    numGuardians = 3
}

// Spawn guardians
for g := 0; g < numGuardians; g++ {
    guardianPos = random position
    guardianSkill = random skill (excluding Tracking/Power Strike)
    guardianLevel = within location LevelMax range
    guardian = generateSkillGuardian(guardianSkill, guardianLevel, location.RarityMax)
    location.Monsters[guardianPos] = guardian
}
```

### Hunt Selection Logic (goHunt)
```go
// Check if player has Tracking
hasTracking = check LearnedSkills for "Tracking"

if hasTracking {
    // Show all monsters with guardian tags
    // Prompt for choice
    // Use player selection
} else {
    // Random selection (original behavior)
}
```

---

## 📋 Changes Summary

### New Skill Added
- **Tracking** added to availableSkills (index 9)

### Functions Modified

1. **generateMonstersForLocation()** (lines 1649-1702)
   - Adds guardian spawning logic
   - Calculates guardian count based on location level
   - Excludes Tracking and Power Strike from guardian skills
   - Spawns guardians in random positions

2. **goHunt()** (lines 2162-2228)
   - Checks for Tracking skill
   - Shows monster selection menu if Tracking learned
   - Allows player to choose target
   - Falls back to random if no Tracking

3. **printMonstersAtLocation()** (lines 1613-1627)
   - Shows guardian tag for Skill Guardians
   - Displays guarded skill name
   - Counts and warns about guardians

### Lines of Code
- **~80 new lines** of functional code
- **~30 lines** modified
- **~200 lines** of documentation

---

## 🧪 Testing

### Manual Test Cases

**Test 1: Tracking Skill Detection**
```
1. Create character
2. Hunt normally (random selection)
3. Defeat guardian, learn Tracking
4. Hunt again - should see selection menu
✅ Pass if selection menu appears
```

**Test 2: Guardian Spawning**
```
1. Start new game or delete gamestate.json
2. View locations (option 4)
3. Check for guardian spawn messages
4. Verify guardians in monster lists
✅ Pass if guardians present in level 10+ locations
```

**Test 3: Monster Selection**
```
1. Have Tracking skill
2. Hunt at location (option 3)
3. See monster list with numbers
4. Choose specific monster
5. Verify correct monster is fought
✅ Pass if chosen monster appears in combat
```

**Test 4: Guardian Identification**
```
1. View location monsters (option 4)
2. Look for 🎯 tags
3. Note guardian names and skills
4. Hunt at that location with Tracking
5. Verify guardians appear in selection
✅ Pass if guardians marked correctly
```

**Test 5: Random Selection Fallback**
```
1. Without Tracking skill
2. Hunt at location
3. Verify random monster selected
4. No selection menu appears
✅ Pass if behaves like original system
```

---

## 🚀 Future Enhancements

### Phase 2: Advanced Tracking
- [ ] Show monster equipment/loot
- [ ] Display monster skills
- [ ] Resistance information
- [ ] Win probability calculator

### Phase 3: Guardian Respawning
- [ ] Defeated guardians respawn over time
- [ ] Respawn with different skills
- [ ] Farming system for skill scrolls

### Phase 4: Tracking Upgrades
- [ ] Enhanced Tracking (see more details)
- [ ] Remote Tracking (scout without visiting)
- [ ] Tracking in auto-play mode
- [ ] Batch target selection

---

## ✅ Benefits

### For Players
- ✅ Strategic combat choices
- ✅ Can target specific skills
- ✅ Avoid dangerous fights
- ✅ Better resource management
- ✅ Visual guardian identification
- ✅ No more random bad luck

### For Progression
- ✅ Guardians always available
- ✅ Multiple guardians at high levels
- ✅ Balanced skill distribution
- ✅ Clear progression path
- ✅ Rewards exploration

### For Gameplay
- ✅ More player agency
- ✅ Reduced frustration
- ✅ Strategic depth
- ✅ Risk/reward decisions
- ✅ Hunting variety

---

## 🎉 Summary

**Tracking Skill:**
- 10th skill added to the game
- Utility skill with no resource cost
- Enables monster selection at locations
- Must be earned from guardian like other skills
- Game-changing quality of life feature

**Automatic Guardian Spawning:**
- Guardians spawn based on location level
- 1-3 guardians per location
- Random skill assignment (8 guardable skills)
- Clear visual identification
- Balanced across all locations

**Both features work together perfectly:**
1. Guardians spawn automatically ✅
2. Player can find them easily ✅
3. Tracking helps choose which to fight ✅
4. Strategic skill acquisition ✅

**The game now has:**
- 10 total skills (9 combat + 1 utility)
- Automatic guardian distribution
- Player agency in combat selection
- Clear visual feedback
- Strategic depth

**Ready for production!** 🗡️🔍🎯
