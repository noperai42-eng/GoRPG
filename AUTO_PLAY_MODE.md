# AUTO-PLAY MODE - Progress Quest Style!

## 🎮 Watch the Game Play Itself!

Inspired by **Progress Quest**, the famous "zero-player RPG", I've implemented an AUTO-PLAY MODE where you can sit back and watch an AI play your game automatically!

---

## What is Auto-Play Mode?

Auto-Play Mode turns your RPG into a spectator experience. The AI takes control of your character and:
- **Automatically hunts monsters** at discovered locations
- **Makes smart combat decisions** (attacks, uses skills, heals, uses items)
- **Manages resources** intelligently (HP, MP, SP)
- **Levels up** and learns new skills
- **Collects loot** and equipment
- **Displays live statistics** every 10 fights
- **Auto-saves** progress every 50 fights

It's like watching a livestream of your character adventuring!

---

## How to Use

### Starting Auto-Play

1. Run the game: `./fight-cli`
2. Select option **8** from the main menu
3. Choose your playback speed:
   - **Slow** (2s per fight) - Watch every detail
   - **Normal** (1s per fight) - Balanced viewing
   - **Fast** (0.5s per fight) - Speed run mode
   - **Turbo** (0.1s per fight) - MAXIMUM SPEED!

4. Press **Ctrl+C** to stop at any time

### Speed Comparison

| Speed | Delay | Fights/Minute | Best For |
|-------|-------|---------------|----------|
| Slow | 2000ms | ~30 | Watching tactics, learning AI decisions |
| Normal | 1000ms | ~60 | General viewing, seeing progression |
| Fast | 500ms | ~120 | Leveling up quickly |
| Turbo | 100ms | ~600 | Maximum XP grinding, overnight runs |

---

## AI Decision Making

The AI uses intelligent priority-based decision making:

### Priority 1: Survival (HP < 40%)
- Use **Heal** spell if available
- Use **Regeneration** if available
- Drink **health potions** if no healing spells

### Priority 2: Buff at Combat Start (Turns 1-2)
- Cast **Battle Cry** (+5 attack for 3 turns)
- Cast **Shield Wall** (+10 defense for 3 turns)
- Sets up advantages for the fight

### Priority 3: Offensive Skills (50% chance)
- **Fireball** - Fire damage + burn effect
- **Lightning Bolt** - High damage + stun chance
- **Ice Shard** - Cold damage
- **Power Strike** - Stamina-based physical attack
- **Poison Blade** - Poison damage over time

### Priority 4: Normal Attack
- Basic physical attack with 15% crit chance
- Fallback if no skills available or resources depleted

---

## Live Statistics Display

Every 10 fights, you'll see:

```
📊 AUTO-PLAY STATISTICS 📊
Fights: 50 | Wins: 47 | Deaths: 3
Level: 5 | XP: 1250 | Total XP Gained: 1200
HP: 18/18 | MP: 35/35 | SP: 35/35
Skills: 6 | Inventory: 12 items
=====================================
```

### What the Stats Mean:
- **Fights** - Total combats completed
- **Wins** - Victories (XP gained)
- **Deaths** - Times character died (auto-resurrects)
- **Level** - Current character level
- **XP** - Total experience points
- **Total XP Gained** - XP earned during this auto-play session
- **HP/MP/SP** - Current/Max resources
- **Skills** - Number of learned skills
- **Inventory** - Number of items carried

---

## Combat Display

Each fight shows a condensed but informative view:

```
⚔️  Fight #4251: Temp (Lv3) vs goblin (Lv2)
  [T1] Temp used Battle Cry (buff)
  [T2] Temp attacks for 8 dmg (Mob HP: 2/10)
  [T3] Temp used Fireball (12 fire dmg)
  ✅ VICTORY! (+20 XP)

⚔️  Fight #4252: Temp (Lv3) vs slime (Lv1)
  [T1] Temp CRITICAL HIT!
  [T1] Temp attacks for 14 dmg (Mob HP: -8/6)
  ✅ VICTORY! (+10 XP)
```

### Combat Indicators:
- `⚔️` - Fight start
- `[T#]` - Turn number
- `CRITICAL HIT!` - 2x damage
- `STUNNED!` - Skipped turn
- `✅ VICTORY!` - Win + XP gained
- `❌ DEFEAT!` - Loss (auto-resurrect)
- `💀 RESURRECTION #X` - Death counter

---

## Features

### Auto-Save System
- Saves progress every **50 fights**
- Prevents data loss
- Resume anytime by restarting auto-play

### Smart Resource Management
- AI conserves mana/stamina for important moments
- Uses healing efficiently
- Prioritizes survival over damage

### Skill Usage
- AI uses learned skills strategically
- Adapts to available resources
- Learns new skills every 3 levels automatically

### Equipment & Loot
- Auto-equips better items
- Collects health potions (30% drop rate)
- Inventory managed automatically

### Level Progression
- Automatic level up
- Skills learned at levels 3, 6, 9, 12, etc.
- Stats increase (HP, MP, SP, Attack, Defense)

---

## Use Cases

### 1. **Idle/AFK Grinding**
Leave it running while you:
- Work or study
- Sleep (use Turbo mode!)
- Watch TV
- Do other things

### 2. **Testing & Balancing**
Watch the AI to:
- See which skills are most effective
- Identify balance issues
- Test monster difficulty
- Observe progression curves

### 3. **Entertainment**
Just enjoy watching:
- Like watching a Twitch stream
- See if your character can survive
- Check back periodically for stats
- Relaxing background activity

### 4. **Speedrunning**
Use Turbo mode to:
- Level up quickly
- Collect lots of loot
- Test maximum progression rate
- Compete for highest level in X time

---

## Tips & Tricks

### Maximize XP Gain
1. Start with a **new character** (low-level enemies = consistent wins)
2. Use **Normal or Fast** speed for good XP rate
3. Let it run for **several hours** for best results
4. **Check stats** every 100 fights

### Avoid Deaths
1. Make sure character has **health potions** in inventory
2. Character learns **Heal** spell at level 1 (starting skill)
3. **Lower difficulty locations** = fewer deaths

### Overnight Grinding
1. Use **Turbo mode** (100ms delay)
2. Save your game first
3. Let it run overnight
4. Wake up to a high-level character!

### Statistics Tracking
- **Wins/Deaths ratio** shows survival rate
- **XP per fight** indicates efficiency
- **Level progression** tracks growth rate

---

## Technical Details

### AI Implementation
- Priority-based decision tree
- Resource availability checks
- Turn-based tactical decisions
- Skill-specific targeting

### Performance
- Minimal CPU usage (turn delays)
- No memory leaks
- Stable for long runs
- Auto-save prevents data loss

### Compatibility
- Works with existing save files
- No impact on manual play mode
- Separate from regular combat system
- Safe to interrupt anytime

---

## Comparison to Progress Quest

| Feature | Progress Quest | This Game |
|---------|---------------|-----------|
| Zero-player | ✅ Yes | ✅ Yes |
| Auto-combat | ✅ Yes | ✅ Yes |
| Live stats | ✅ Yes | ✅ Yes |
| Speed control | ❌ No | ✅ Yes (4 speeds) |
| Skill decisions | ❌ Random | ✅ Intelligent AI |
| Watch combat | ❌ No details | ✅ Turn-by-turn |
| Pause/Resume | ❌ No | ✅ Ctrl+C anytime |
| Auto-save | ❌ No | ✅ Every 50 fights |

---

## Examples

### Example 1: Slow Speed (Learning)
```
⚔️  Fight #1: Temp (Lv1) vs goblin (Lv1)
  [T1] Temp used Battle Cry (buff)
  [T2] Temp used Fireball (18 fire dmg)
  ✅ VICTORY! (+10 XP)

⚔️  Fight #2: Temp (Lv1) vs slime (Lv1)
  [T1] Temp attacks for 6 dmg (Mob HP: 0/6)
  ✅ VICTORY! (+10 XP)

... (watches every detail at 2s per fight)
```

### Example 2: Turbo Speed (Grinding)
```
⚔️  Fight #7891: Temp (Lv42) vs orc (Lv38)
  ✅ VICTORY! (+380 XP)
⚔️  Fight #7892: Temp (Lv42) vs goblin (Lv35)
  ✅ VICTORY! (+350 XP)
⚔️  Fight #7893: Temp (Lv42) vs slime (Lv40)
  ❌ DEFEAT!
💀 RESURRECTION #15

📊 AUTO-PLAY STATISTICS 📊
Fights: 7900 | Wins: 7624 | Deaths: 276
Level: 42 | XP: 289450 | Total XP Gained: 289450
HP: 52/52 | MP: 137/137 | SP: 137/137
Skills: 9 | Inventory: 45 items
=====================================
```

---

## Stopping Auto-Play

**Press Ctrl+C** at any time to stop.

The game will:
- Stop combat immediately
- Save current state (if at save interval)
- Return to main menu
- Character progress is preserved

---

## 🎯 Perfect For

- **Idle gaming fans** - Love incremental/idle games
- **Busy people** - Want progress without active play
- **Testers** - Need to see long-term balance
- **Streamers** - Background content for streams
- **Curious players** - Want to see AI play

---

## Future Enhancements

Potential additions:
- Multiple characters auto-playing simultaneously
- Real-time graphs of progression
- Custom AI strategies
- Boss fight notifications
- Discord/webhook notifications for milestones
- Web dashboard to monitor remotely

---

## 🏆 Achievement Ideas

Track your auto-play sessions:
- **AFKer** - 100 auto-play fights
- **Observer** - 1,000 auto-play fights
- **Spectator** - 10,000 auto-play fights
- **Zero-Player Legend** - Reach level 50 via auto-play
- **Marathon Runner** - 24-hour continuous auto-play session

---

**Enjoy watching your character become a legend... automatically!** 🎮✨
