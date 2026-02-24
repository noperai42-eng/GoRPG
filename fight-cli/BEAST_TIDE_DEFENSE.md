# Beast Tide Defense System

## ✅ Feature Complete

The beast tide defense system is a wave-based tower defense mechanic where your village defenses are tested against waves of attacking monsters.

---

## 🌊 How It Works

### Triggering a Tide

**Time-Based:**
- Default interval: 3600 seconds (1 hour)
- LastTideTime tracked in village
- Countdown visible in "Check Next Monster Tide" menu

**Manual Trigger:**
- Village Menu → Option 7 "Defend Against Tide"
- Only available when tide is ready (timer expired)
- Can prepare defenses before triggering

---

## 📊 Tide Scaling

### Difficulty Calculation

**Waves:**
```
Number of Waves = 3 + (Village Level / 5)
```
- Level 1-4: 3 waves
- Level 5-9: 4 waves
- Level 10-14: 5 waves
- Level 15+: 6+ waves

**Monster Level:**
```
Monster Level = Village Level ± 2
```
- Scaled to village level
- ±2 random variance per monster
- Challenges but not overwhelming

**Monsters Per Wave:**
```
Monsters Per Wave = 5 + (Village Level / 3)
```
- Level 1-2: 5 monsters
- Level 3-5: 6 monsters
- Level 6-8: 7 monsters
- Level 9-11: 8 monsters
- Level 12+: 9+ monsters

---

## ⚔️ Combat Phases

### Phase 1: Trap Triggering

**Mechanics:**
- Each trap has a TriggerRate (%)
- Random roll per monster
- If triggered: Deal trap Damage
- Kill if HP drops to 0

**Example:**
```
Spike Trap (60% trigger rate, 15 damage)
→ 60% chance to trigger
→ If triggers: 15 damage to monster
→ If monster dies: End phases
```

**Trap Types:**
- Spike Trap: 60% trigger, 15 damage
- Fire Trap: 50% trigger, 25 damage
- Ice Trap: 55% trigger, 20 damage
- Poison Trap: 65% trigger, 18 damage
- Barricade Trap: 70% trigger, 30 damage

### Phase 2: Tower Attacks

**Mechanics:**
- Calculates total Attack Power from all towers/gates
- Adds random variance (+0-4)
- Deals damage to monster
- Only if trap didn't kill it

**Tower Stats:**
- Guard Tower: 20 attack
- Arrow Tower: 35 attack
- Iron Gate: 10 attack

**Example:**
```
Total Attack = 55 (Guard + Arrow)
Tower Damage = 55 + rand(5) = 57
Monster takes 57 damage
```

### Phase 3: Guard Combat

**Mechanics:**
- Count all guards (villagers + hired)
- Each guard deals: 5-12 damage (random)
- Total damage = Guards × Individual Damage
- Deals damage to monster

**Guard Sources:**
- Villager guards (rescued, role = "guard")
- Hired guards (paid with gold)

**Example:**
```
Total Guards = 3
Guard Damage = 3 × 8 = 24 damage
Monster takes 24 damage
```

### Phase 4: Monster Breakthrough

**Mechanics:**
- Monster attacks village directly
- Base Damage = AttackRolls × 6
- Reduced by total Defense
- Minimum 1 damage always applied

**Damage Formula:**
```
Monster Attack = AttackRolls × 6
Final Damage = max(1, Monster Attack - Total Defense)
Village takes Final Damage
```

**Example:**
```
Monster: 3 attack rolls = 18 base damage
Total Defense = 50
Reduced Damage = 18 - 50 = 1 (minimum)
Village takes 1 damage
```

---

## 🏆 Victory Conditions

### Victory Formula

```
Damage Taken < Defense Level × 50
```

**Example:**
- Defense Level 5: Threshold = 250 damage
- Damage Taken 180: VICTORY
- Damage Taken 280: DEFEAT

### Victory Rewards

**Base Rewards:**
```
Village XP = 100 × Waves Defeated
```
- 3 waves: +300 XP
- 4 waves: +400 XP
- 5 waves: +500 XP

**Bonus Rewards (Perfect Defense):**
```
If Damage Taken < Threshold / 2:
  Bonus Gold = 50 + (Village Level × 10)
```
- Level 5: +100 Gold
- Level 10: +150 Gold
- Level 15: +200 Gold

**Example Victory:**
```
============================================================
🏆 TIDE DEFENSE COMPLETE! 🏆
============================================================

✨ VICTORY! Your defenses held strong! ✨

Battle Summary:
  Waves Defeated: 4/4
  Monsters Killed: 28
  Damage Dealt: 1542
  Damage Taken: 120/250
  Traps Triggered: 18 times

Rewards:
  Village XP: +400
  Bonus Gold: +150 (minimal damage taken!)
```

---

## 💀 Defeat Consequences

### Defeat Penalties

**Resource Loss:**
```
Resource Loss = Village Level × 5
```
Applied to all basic resources:
- Lumber
- Gold
- Iron
- Sand
- Stone

**Example:**
- Level 5 village: Lose 25 of each resource
- Level 10 village: Lose 50 of each resource

**Guard Casualties:**
```
Guards Lost = 1 + rand(Hired Guards / 2)
```
- Only affects hired guards
- Villager guards survive
- Can lose up to 50% of hired guards

**Example Defeat:**
```
============================================================
🏆 TIDE DEFENSE COMPLETE! 🏆
============================================================

💀 DEFEAT! The tide overwhelmed your defenses! 💀

Battle Summary:
  Waves Survived: 2/4
  Monsters Killed: 12
  Damage Taken: 310/250 (too much!)

Penalties:
  Lost 25 of each resource type
  2 hired guards were lost
```

---

## 🎮 Strategy Guide

### Early Game Defense (Village Level 1-5)

**Preparation:**
1. Build 2-3 Wooden Walls (Defense +10 each)
2. Craft 2-3 Spike Traps (easy materials)
3. Rescue 1-2 guard villagers
4. Hire 1 basic guard if possible

**Expected Results:**
- 3 waves of 5-6 monsters each
- Total Defense: ~40-50
- Damage Threshold: 150-250
- Should win with basic prep

**Tips:**
- Focus on walls first (cheap, reliable)
- Spike traps only need Beast Bone
- Guard villagers are free (rescue)
- Save gold for mid-game upgrades

### Mid Game Defense (Village Level 6-10)

**Preparation:**
1. Upgrade to Stone Walls (Defense +25)
2. Build 1 Guard Tower (Attack +20)
3. Craft Fire/Ice Traps (stronger)
4. Have 3-4 guards total
5. Stock health potions

**Expected Results:**
- 4-5 waves of 7-9 monsters each
- Total Defense: 80-120
- Attack: 20-40
- Damage Threshold: 300-500
- Can achieve bonus gold

**Tips:**
- Towers add offense (kill before damage)
- Diversify trap types
- Hire elite guards (village level +5)
- Farm beast materials in advance

### Late Game Defense (Village Level 11+)

**Preparation:**
1. Build Iron Walls (Defense +40)
2. Build Arrow Towers (Attack +35)
3. Craft all trap types (redundancy)
4. Maintain 5+ guards
5. Full equipment and skills

**Expected Results:**
- 6+ waves of 10+ monsters each
- Total Defense: 150+
- Attack: 50-70
- Damage Threshold: 550+
- Bonus gold guaranteed

**Tips:**
- Mix walls and towers
- Have 5+ traps active
- Replace consumed traps
- Guards at maximum efficiency

---

## 💡 Advanced Tactics

### Trap Management

**Duration Planning:**
```
Spike Trap: 3 waves = Good for full tide
Fire Trap: 2 waves = Use early waves
Barricade: 2 waves = High damage, short life
```

**Strategy:**
1. Craft before tide
2. Use high-damage traps early
3. Let durable traps finish tide
4. Recraft after tide

### Defense Balance

**Pure Defense (Safe):**
- 5+ Walls = High defense
- Few towers = Low attack
- Result: Survive but slow

**Balanced (Optimal):**
- 3 Walls + 2 Towers
- Good defense + Good attack
- Result: Efficient victories

**Aggressive (Risky):**
- 2 Walls + 3 Towers
- Lower defense + High attack
- Result: Fast but dangerous

### Resource Investment

**Gold Priority:**
1. Hire guards first (immediate help)
2. Build defenses second
3. Upgrade skills last

**Material Priority:**
1. Spike Traps (common materials)
2. Fire Traps (strong damage)
3. Barricade Traps (beast materials)

**Time Priority:**
1. Build walls immediately
2. Craft traps before tide
3. Hire guards when rich

---

## 📊 Statistics & Balancing

### Trap Effectiveness

| Trap | Damage | Duration | Trigger | Cost | Efficiency |
|------|--------|----------|---------|------|------------|
| Spike | 15 | 3 waves | 60% | Low | ★★★☆☆ |
| Fire | 25 | 2 waves | 50% | Med | ★★★★☆ |
| Ice | 20 | 3 waves | 55% | Med | ★★★★☆ |
| Poison | 18 | 4 waves | 65% | High | ★★★★★ |
| Barricade | 30 | 2 waves | 70% | Med | ★★★★☆ |

**Analysis:**
- Poison Trap best overall (long duration + high trigger)
- Fire Trap best damage per wave
- Spike Trap most accessible
- Barricade Trap highest single hit

### Defense ROI

| Structure | Cost | Benefit | ROI |
|-----------|------|---------|-----|
| Wooden Wall | 70 | Def +10 | ★★★☆☆ |
| Stone Wall | 100 | Def +25 | ★★★★☆ |
| Iron Wall | 140 | Def +40 | ★★★☆☆ |
| Guard Tower | 110 | Def +15, Atk +20 | ★★★★★ |
| Arrow Tower | 120 | Def +10, Atk +35 | ★★★★★ |
| Iron Gate | 120 | Def +30, Atk +10 | ★★★★☆ |

**Analysis:**
- Towers have best ROI (offense + defense)
- Stone Wall best pure defense value
- Iron Wall expensive but powerful
- Gates are balanced option

### Victory Thresholds

| Village Level | Waves | Monsters | Threshold | Easy Defense |
|---------------|-------|----------|-----------|--------------|
| 1-2 | 3 | 15-18 | 50-100 | 20+ defense |
| 3-4 | 3 | 18-24 | 150-200 | 40+ defense |
| 5-7 | 4 | 28-36 | 250-350 | 70+ defense |
| 8-10 | 5 | 40-50 | 400-500 | 100+ defense |
| 11-14 | 5-6 | 50-70 | 550-700 | 140+ defense |
| 15+ | 6+ | 70+ | 750+ | 180+ defense |

**Recommended Defense = Threshold / 5**

---

## 🔧 Technical Implementation

### Code Structure

**Main Function (lines 4161-4407):**
```go
func monsterTideDefense(gameState *GameState, player *Character, village *Village)
```

**Scaling Calculations:**
```go
numWaves := 3 + (village.Level / 5)
baseMonsterLevel := village.Level
monstersPerWave := 5 + (village.Level / 3)
```

**Victory Condition:**
```go
damageThreshold := village.DefenseLevel * 50
if damageTaken < damageThreshold {
    // VICTORY
} else {
    // DEFEAT
}
```

### Integration Points

**Village Menu (Option 7):**
```go
case "7":
    // Check if tide is ready
    if timeUntilNext <= 0 {
        monsterTideDefense(gameState, player, village)
        // Auto-save after tide
    }
```

**Tide Timer:**
```go
currentTime := time.Now().Unix()
timeSinceLastTide := currentTime - village.LastTideTime
timeUntilNext := village.TideInterval - int(timeSinceLastTide)
```

**Trap Consumption:**
```go
for j := len(village.Traps) - 1; j >= 0; j-- {
    village.Traps[j].Remaining--
    if village.Traps[j].Remaining <= 0 {
        // Remove consumed trap
        village.Traps = append(village.Traps[:j], village.Traps[j+1:]...)
    }
}
```

---

## 🧪 Testing Scenarios

### Scenario 1: Under-Prepared (Should Lose)
- Village Level 5
- 1 Wooden Wall (Defense +10)
- 0 Traps
- 0 Guards
- **Expected:** Defeat, take ~200+ damage

### Scenario 2: Adequately Prepared (Should Win)
- Village Level 5
- 3 Stone Walls (Defense +75)
- 2 Spike Traps
- 2 Guards
- **Expected:** Victory, take ~100-150 damage

### Scenario 3: Over-Prepared (Bonus Gold)
- Village Level 5
- 2 Iron Walls + 2 Towers (Defense +95, Attack +40)
- 5 Traps (mixed)
- 4 Guards
- **Expected:** Victory, take <125 damage, bonus gold

### Scenario 4: Late Game Challenge
- Village Level 15
- 4 Iron Walls + 3 Arrow Towers (Defense +265, Attack +105)
- 8 Traps active
- 8 Guards
- **Expected:** Victory, minimal damage, large rewards

---

## 📖 Player Guide

### How to Defend Your Village

**Step 1: Prepare Defenses**
```
Village Menu → Option 5 (Build Defenses)
→ Build walls and towers
→ Craft traps with beast materials
```

**Step 2: Recruit Guards**
```
Village Menu → Option 3 (Hire Guards)
→ Spend gold on guards
Hunt to rescue guard villagers (15% chance)
```

**Step 3: Check Tide Status**
```
Village Menu → Option 6 (Check Next Monster Tide)
→ See time remaining
→ Review current defenses
```

**Step 4: Defend When Ready**
```
Village Menu → Option 7 (Defend Against Tide)
→ Only available when tide is ready
→ Wave-based combat begins
→ Watch your defenses in action
```

**Step 5: Collect Rewards**
```
Victory: Massive village XP + bonus gold
Defeat: Learn and rebuild stronger
```

---

## 🚀 Future Enhancements

### Potential Additions

1. **Tide Difficulty Settings**
   - Easy/Normal/Hard modes
   - Adjustable rewards

2. **Special Tide Types**
   - Aerial tides (need towers)
   - Armored tides (bypass defenses)
   - Boss tides (one strong enemy)

3. **Trap Upgrades**
   - Upgrade existing traps
   - Combine traps for combos
   - Elemental interactions

4. **Guard Abilities**
   - Active skills for guards
   - Guard formations
   - Elite guard types

5. **Tide Achievements**
   - Perfect defense streaks
   - Speed run completions
   - Specific kill methods

---

## 📝 Summary

**Status:** ✅ Production Ready

**Features:**
- Wave-based tower defense combat
- Uses walls, towers, traps, and guards
- Scales with village level
- Victory/defeat outcomes with rewards/penalties
- Trap consumption system
- Time-based tide intervals
- Manual tide triggering

**Integration:**
- Village menu option 7
- Save/load compatible
- Beast material economy
- Village XP progression

**Balance:**
- Fair scaling curve
- Strategic depth
- Resource investment meaningful
- Multiple viable strategies

Defend your village! 🏰🌊⚔️
