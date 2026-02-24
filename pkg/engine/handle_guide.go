package engine

// handleGuideMain displays the player guide topic menu.
func (e *Engine) handleGuideMain(session *GameSession, cmd GameCommand) GameResponse {
	if cmd.Value == "0" || cmd.Value == "back" {
		session.State = StateMainMenu
		return BuildMainMenuResponse(session)
	}

	topicMap := map[string]string{
		"1": StateGuideCombat,
		"2": StateGuideSkills,
		"3": StateGuideVillage,
		"4": StateGuideCrafting,
		"5": StateGuideMonsterDrops,
		"6": StateGuideAutoPlay,
		"7": StateGuideQuests,
	}

	if state, ok := topicMap[cmd.Value]; ok {
		session.State = state
		return e.handleGuideTopic(session, GameCommand{Type: "init"})
	}

	// Default: show the guide menu
	session.State = StateGuideMain
	return GameResponse{
		Type:     "menu",
		Messages: []GameMessage{Msg("=== PLAYER GUIDE ===", "system"), Msg("Select a topic to learn more:", "system")},
		State:    &StateData{Screen: "guide_main", Player: MakePlayerState(session.Player)},
		Options: []MenuOption{
			Opt("1", "Combat Basics"),
			Opt("2", "Skills & Magic"),
			Opt("3", "Village & Leveling"),
			Opt("4", "Crafting Guide"),
			Opt("5", "Monster Drops"),
			Opt("6", "Auto-Play Mode"),
			Opt("7", "Quests"),
			Opt("0", "Back to Main Menu"),
		},
	}
}

// handleGuideTopic displays the selected guide topic content.
func (e *Engine) handleGuideTopic(session *GameSession, cmd GameCommand) GameResponse {
	if cmd.Value == "0" || cmd.Value == "back" {
		session.State = StateGuideMain
		return e.handleGuideMain(session, GameCommand{Type: "init"})
	}

	var msgs []GameMessage
	switch session.State {
	case StateGuideCombat:
		msgs = guideContentCombat()
	case StateGuideSkills:
		msgs = guideContentSkills()
	case StateGuideVillage:
		msgs = guideContentVillage()
	case StateGuideCrafting:
		msgs = guideContentCrafting()
	case StateGuideMonsterDrops:
		msgs = guideContentMonsterDrops()
	case StateGuideAutoPlay:
		msgs = guideContentAutoPlay()
	case StateGuideQuests:
		msgs = guideContentQuests()
	}

	return GameResponse{
		Type:     "menu",
		Messages: msgs,
		State:    &StateData{Screen: string(session.State), Player: MakePlayerState(session.Player)},
		Options:  []MenuOption{Opt("back", "Back to Guide")},
	}
}

// ---------------------------------------------------------------------------
// Guide content builders
// ---------------------------------------------------------------------------

func guideContentCombat() []GameMessage {
	return []GameMessage{
		Msg("=== COMBAT BASICS ===", "system"),
		Msg("", "system"),
		Msg("Combat is turn-based. Each turn you choose an action, then the monster acts.", "narrative"),
		Msg("", "system"),
		Msg("ACTIONS:", "system"),
		Msg("  1. Attack - Normal physical attack (15% crit chance for double damage)", "narrative"),
		Msg("  2. Defend - +50% defense but only 50% attack power", "narrative"),
		Msg("  3. Use Item - Consume a potion or other consumable from inventory", "narrative"),
		Msg("  4. Use Skill - Cast a spell or ability (costs Mana or Stamina)", "narrative"),
		Msg("  5. Flee - Attempt to escape combat", "narrative"),
		Msg("", "system"),
		Msg("FLEE CHANCE:", "system"),
		Msg("  Base 20% + bonus if your level is higher than the monster", "narrative"),
		Msg("  Up to 90% max flee chance", "narrative"),
		Msg("", "system"),
		Msg("ELEMENTAL DAMAGE:", "system"),
		Msg("  Physical - Standard melee damage", "narrative"),
		Msg("  Fire     - Strong vs ice-weak enemies", "narrative"),
		Msg("  Ice      - Can slow or freeze targets", "narrative"),
		Msg("  Lightning - Strong vs metal/golem types", "narrative"),
		Msg("  Poison   - Damage over time effects", "narrative"),
		Msg("", "system"),
		Msg("RESISTANCE MULTIPLIERS:", "system"),
		Msg("  Very Resistant: 0.25x damage", "narrative"),
		Msg("  Resistant:      0.5x damage", "narrative"),
		Msg("  Normal:         1.0x damage", "narrative"),
		Msg("  Weak:           2.0x damage", "narrative"),
		Msg("", "system"),
		Msg("NOTABLE RESISTANCES:", "system"),
		Msg("  Slimes  - Weak to Fire (2x), Resistant to Physical (0.5x)", "narrative"),
		Msg("  Golems  - Weak to Lightning (2x), Very Resistant to Physical (0.25x)", "narrative"),
		Msg("  Hiftiers - Resistant to all magic (0.5x)", "narrative"),
		Msg("", "system"),
		Msg("ON VICTORY:", "system"),
		Msg("  Gain XP (monster level x 10)", "narrative"),
		Msg("  Loot the monster's equipment (auto-equips if better)", "narrative"),
		Msg("  30% chance to find a health potion", "narrative"),
		Msg("  Chance to get beast materials for crafting", "narrative"),
		Msg("", "system"),
		Msg("ON DEFEAT:", "system"),
		Msg("  Your character is resurrected automatically", "narrative"),
		Msg("  The monster takes your equipped items", "narrative"),
		Msg("  Monster gains XP from your level", "narrative"),
	}
}

func guideContentSkills() []GameMessage {
	return []GameMessage{
		Msg("=== SKILLS & MAGIC ===", "system"),
		Msg("", "system"),
		Msg("You learn a new skill every 3 levels. 9 skills total.", "narrative"),
		Msg("Skills cost Mana (MP) and/or Stamina (SP) to use.", "narrative"),
		Msg("MP and SP fully restore at the start of each combat.", "narrative"),
		Msg("", "system"),
		Msg("PLAYER SKILLS:", "system"),
		Msg("  Fireball      - 15 MP | 20 Fire damage", "narrative"),
		Msg("  Power Strike   - 10 SP | 15 Physical damage", "narrative"),
		Msg("  Heal           - 20 MP | Restores 25 HP", "narrative"),
		Msg("  Ice Shard      - 18 MP | 22 Ice damage", "narrative"),
		Msg("  Poison Blade   - 8 SP | Applies poison (5 dmg/turn, 3 turns)", "narrative"),
		Msg("  Lightning Bolt - 25 MP | 30 Lightning damage", "narrative"),
		Msg("  Regeneration   - 22 MP | Heal 8 HP/turn for 4 turns", "narrative"),
		Msg("  Shield Wall    - 12 SP | +10 defense for 3 turns", "narrative"),
		Msg("  Battle Cry     - 15 SP | +8 attack for 3 turns", "narrative"),
		Msg("  Tracking       - Passive | Choose your target when hunting", "narrative"),
		Msg("", "system"),
		Msg("STATUS EFFECTS:", "system"),
		Msg("  Poison      - Damage each turn, duration decreases", "narrative"),
		Msg("  Burn        - Fire damage each turn", "narrative"),
		Msg("  Stun        - Skip entire turn (cannot act)", "narrative"),
		Msg("  Regen       - Heal HP each turn", "narrative"),
		Msg("  Buff Attack - Bonus attack damage for duration", "narrative"),
		Msg("  Buff Defense - Bonus defense for duration", "narrative"),
		Msg("", "system"),
		Msg("SKILL SCROLLS:", "system"),
		Msg("  Defeat Skill Guardians to earn skill scrolls", "narrative"),
		Msg("  Guardians are special monsters marked [GUARDIAN]", "narrative"),
		Msg("  You can also craft scrolls at your village (Level 10+)", "narrative"),
		Msg("", "system"),
		Msg("SKILL UPGRADES:", "system"),
		Msg("  Available at Village Level 10+", "narrative"),
		Msg("  Each upgrade: +5 damage (or healing), -2 mana/stamina cost", "narrative"),
		Msg("  Cost: 25 Iron + 50 Gold per upgrade", "narrative"),
		Msg("", "system"),
		Msg("MONSTER SKILLS:", "system"),
		Msg("  Monsters have 40% chance to use skills each turn", "narrative"),
		Msg("  Each monster type has unique abilities:", "narrative"),
		Msg("    Slimes   - Acid Spit (poison)", "narrative"),
		Msg("    Goblins  - Backstab (physical burst)", "narrative"),
		Msg("    Orcs     - War Cry (attack buff), Berserker Rage", "narrative"),
		Msg("    Golems   - Stone Skin (defense buff)", "narrative"),
		Msg("    Hiftiers - Mana Bolt (lightning), Mind Blast (stun)", "narrative"),
		Msg("    Kobolds  - Fire Breath (burn)", "narrative"),
		Msg("    Kitpods  - Regenerate (healing over time)", "narrative"),
	}
}

func guideContentVillage() []GameMessage {
	return []GameMessage{
		Msg("=== VILLAGE & LEVELING ===", "system"),
		Msg("", "system"),
		Msg("Your village is created automatically when you first visit it.", "narrative"),
		Msg("Villages gain XP from crafting, hiring, building, and defending.", "narrative"),
		Msg("", "system"),
		Msg("VILLAGE LEVEL REQUIREMENTS:", "system"),
		Msg("  Each level requires: Level x 100 XP", "narrative"),
		Msg("  Level 2: 100 XP  |  Level 5: 500 XP  |  Level 10: 1000 XP", "narrative"),
		Msg("", "system"),
		Msg("UNLOCKS BY LEVEL:", "system"),
		Msg("  Level 3  - Potion Crafting", "narrative"),
		Msg("  Level 5  - Armor Crafting", "narrative"),
		Msg("  Level 7  - Weapon Crafting", "narrative"),
		Msg("  Level 10 - Skill Scroll Crafting & Skill Upgrades", "narrative"),
		Msg("", "system"),
		Msg("XP SOURCES:", "system"),
		Msg("  Craft Potions:  20 XP each", "narrative"),
		Msg("  Craft Armor:    40-70 XP (varies by recipe)", "narrative"),
		Msg("  Craft Weapons:  50-80 XP (varies by recipe)", "narrative"),
		Msg("  Craft Scrolls:  60-120 XP (varies by skill)", "narrative"),
		Msg("  Craft Traps:    35 XP each", "narrative"),
		Msg("  Upgrade Skill:  60 XP each", "narrative"),
		Msg("  Hire Guard:     30 XP each", "narrative"),
		Msg("  Build Walls:    20 XP per section", "narrative"),
		Msg("  Monster Tide kills: 10 XP per monster", "narrative"),
		Msg("", "system"),
		Msg("VILLAGERS:", "system"),
		Msg("  15% chance to rescue a villager after each auto-play victory", "narrative"),
		Msg("  Villagers can be assigned to harvest resources automatically", "narrative"),
		Msg("  Each villager has an efficiency rating (1-5)", "narrative"),
		Msg("", "system"),
		Msg("GUARDS:", "system"),
		Msg("  Hire guards using Gold at your village", "narrative"),
		Msg("  Guards join you in Skill Guardian and Boss fights", "narrative"),
		Msg("  Equip guards with items from your inventory", "narrative"),
		Msg("  Guards recover over time when injured", "narrative"),
		Msg("  WARNING: Guards can die permanently in Boss fights!", "narrative"),
		Msg("", "system"),
		Msg("VILLAGE DEFENSES:", "system"),
		Msg("  Build walls and craft traps to defend against Monster Tides", "narrative"),
		Msg("  Monster Tides are waves of enemies that attack your village", "narrative"),
		Msg("  Guards + traps + walls all contribute to defense", "narrative"),
	}
}

func guideContentCrafting() []GameMessage {
	return []GameMessage{
		Msg("=== CRAFTING GUIDE ===", "system"),
		Msg("", "system"),
		Msg("Crafting is done at your Village. Different recipes unlock at different village levels.", "narrative"),
		Msg("Resources come from Harvesting (Iron, Gold, Stone, Lumber, Sand)", "narrative"),
		Msg("and from Monster Drops (Beast Skin, Beast Bone, etc.)", "narrative"),
		Msg("", "system"),
		Msg("--- POTIONS (Village Lv3+) ---", "system"),
		Msg("  Small Health Potion:  5 Iron, 10 Gold    (+20 XP)", "narrative"),
		Msg("  Medium Health Potion: 10 Iron, 20 Gold   (+20 XP)", "narrative"),
		Msg("  Large Health Potion:  20 Iron, 40 Gold   (+20 XP)", "narrative"),
		Msg("", "system"),
		Msg("--- ARMOR (Village Lv5+) ---", "system"),
		Msg("  Enhanced Armor:      30 Iron, 20 Stone                   (Rarity 3-5, +40 XP)", "narrative"),
		Msg("  Beast Skin Armor:    20 Iron, 15 Beast Skin              (Rarity 4-6, +50 XP)", "narrative"),
		Msg("  Bone Plate Armor:    25 Iron, 15 Stone, 12 Beast Bone   (Rarity 5-7, +60 XP)", "narrative"),
		Msg("  Tough Hide Vest:     8 Beast Bone, 10 Tough Hide         (Rarity 4-6, +50 XP)", "narrative"),
		Msg("  Ore Fragment Mail:   15 Iron, 20 Ore Fragment            (Rarity 5-7, +60 XP)", "narrative"),
		Msg("  Fang-Studded Armor:  20 Iron, 10 Beast Skin, 15 Sharp Fang (Rarity 6-8, +70 XP)", "narrative"),
		Msg("  Claw Guard Armor:    15 Iron, 8 Tough Hide               (Rarity 6-8, +70 XP)", "narrative"),
		Msg("", "system"),
		Msg("--- WEAPONS (Village Lv7+) ---", "system"),
		Msg("  Enhanced Weapon:     40 Iron, 30 Gold                    (Rarity 4-6, +50 XP)", "narrative"),
		Msg("  Beast Claw Blade:    25 Iron, 15 Monster Claw, 10 Sharp Fang (Rarity 5-7, +60 XP)", "narrative"),
		Msg("  Bone Crusher Mace:   30 Iron, 20 Beast Bone, 15 Stone   (Rarity 5-7, +60 XP)", "narrative"),
		Msg("  Hide-Wrapped Axe:    20 Iron, 25 Lumber, 12 Tough Hide  (Rarity 4-6, +55 XP)", "narrative"),
		Msg("  Ore Fragment Sword:  35 Iron, 20 Gold, 25 Ore Fragment  (Rarity 6-8, +70 XP)", "narrative"),
		Msg("  Fang Spear:          20 Iron, 15 Beast Bone, 18 Sharp Fang (Rarity 5-7, +65 XP)", "narrative"),
		Msg("  Composite War Hammer: 25 Iron, 10 Beast Skin, 20 Stone, 15 Ore Fragment (Rarity 6-8, +80 XP)", "narrative"),
		Msg("", "system"),
		Msg("--- SKILL SCROLLS (Village Lv10+) ---", "system"),
		Msg("  Fireball:      15 Ore Fragment, 10 Sharp Fang, 30 Gold   (+100 XP)", "narrative"),
		Msg("  Ice Shard:     10 Ore Fragment, 12 Beast Skin, 20 Iron   (+90 XP)", "narrative"),
		Msg("  Lightning Bolt: 20 Ore Fragment, 15 Sharp Fang, 40 Gold  (+120 XP)", "narrative"),
		Msg("  Power Strike:  15 Iron, 10 Beast Bone                    (+70 XP)", "narrative"),
		Msg("  Poison Blade:  15 Iron, 12 Sharp Fang, 10 Beast Skin    (+85 XP)", "narrative"),
		Msg("  Heal:          25 Gold, 15 Beast Skin, 10 Beast Bone    (+95 XP)", "narrative"),
		Msg("  Regeneration:  30 Gold, 12 Ore Fragment, 15 Beast Skin  (+105 XP)", "narrative"),
		Msg("  Shield Wall:   12 Beast Bone, 15 Tough Hide, 20 Stone   (+90 XP)", "narrative"),
		Msg("  Battle Cry:    20 Iron, 15 Sharp Fang, 10 Beast Bone    (+85 XP)", "narrative"),
		Msg("  Tracking:      8 Beast Skin, 8 Beast Bone                (+60 XP)", "narrative"),
		Msg("", "system"),
		Msg("--- TRAPS (Village Defenses) ---", "system"),
		Msg("  Spike Trap:    10 Iron, 5 Beast Bone         (15 dmg, 60% trigger, 3 tides)", "narrative"),
		Msg("  Fire Trap:     15 Iron, 8 Ore Fragment, 5 Sharp Fang (25 dmg, 50% trigger, 2 tides)", "narrative"),
		Msg("  Ice Trap:      12 Iron, 10 Ore Fragment, 8 Beast Skin (20 dmg, 55% trigger, 3 tides)", "narrative"),
		Msg("  Poison Trap:   10 Beast Skin, 8 Sharp Fang, 5 Monster Claw (18 dmg, 65% trigger, 4 tides)", "narrative"),
		Msg("  Barricade:     30 Lumber, 8 Beast Bone, 6 Tough Hide (30 dmg, 70% trigger, 2 tides)", "narrative"),
		Msg("", "system"),
		Msg("--- SKILL UPGRADES (Village Lv10+) ---", "system"),
		Msg("  Cost: 25 Iron + 50 Gold per upgrade", "narrative"),
		Msg("  Effect: +5 Damage (or Healing), -2 Mana/Stamina cost", "narrative"),
	}
}

func guideContentMonsterDrops() []GameMessage {
	return []GameMessage{
		Msg("=== MONSTER DROPS ===", "system"),
		Msg("", "system"),
		Msg("Monsters drop beast materials used for crafting.", "narrative"),
		Msg("Each monster type has its own drop pool and drop chance.", "narrative"),
		Msg("When a drop triggers, you get 1-3 units of one material.", "narrative"),
		Msg("", "system"),
		Msg("MONSTER TYPE        DROPS                           CHANCE", "system"),
		Msg("----------------------------------------------------------", "system"),
		Msg("Slime               Beast Skin, Ore Fragment          40%", "narrative"),
		Msg("Goblin              Beast Bone, Sharp Fang             50%", "narrative"),
		Msg("Orc                 Beast Bone, Tough Hide             55%", "narrative"),
		Msg("Kobold              Sharp Fang, Beast Skin             45%", "narrative"),
		Msg("Hiftier             Ore Fragment, Monster Claw         60%", "narrative"),
		Msg("Golem               Ore Fragment, Beast Bone           70%", "narrative"),
		Msg("Kitpod              Tough Hide, Monster Claw           50%", "narrative"),
		Msg("Skill Guardian      Tough Hide, Sharp Fang, Monster Claw  80%", "narrative"),
		Msg("Other               Beast Skin, Beast Bone             40%", "narrative"),
		Msg("----------------------------------------------------------", "system"),
		Msg("", "system"),
		Msg("MATERIAL USES:", "system"),
		Msg("  Beast Skin    - Armor, scrolls, traps", "narrative"),
		Msg("  Beast Bone    - Armor, weapons, scrolls, traps", "narrative"),
		Msg("  Ore Fragment  - Armor, weapons, scrolls, traps (high-tier)", "narrative"),
		Msg("  Tough Hide    - Armor, weapons, scrolls, traps", "narrative"),
		Msg("  Sharp Fang    - Armor, weapons, scrolls, traps", "narrative"),
		Msg("  Monster Claw  - Weapons, traps", "narrative"),
		Msg("", "system"),
		Msg("FARMING TIPS:", "system"),
		Msg("  Golems are the best source for Ore Fragment (70% drop rate)", "narrative"),
		Msg("  Goblins are reliable for Beast Bone and Sharp Fang (50%)", "narrative"),
		Msg("  Hiftiers drop Ore Fragment and Monster Claw (60%)", "narrative"),
		Msg("  Skill Guardians have the highest drop rate (80%, 3 materials)", "narrative"),
		Msg("  Use Auto-Play to farm materials efficiently", "narrative"),
	}
}

func guideContentAutoPlay() []GameMessage {
	return []GameMessage{
		Msg("=== AUTO-PLAY MODE ===", "system"),
		Msg("", "system"),
		Msg("Auto-Play fights monsters automatically at your first hunt location.", "narrative"),
		Msg("The AI makes combat decisions for you (attack, skills, items, heal).", "narrative"),
		Msg("", "system"),
		Msg("SPEED SETTINGS:", "system"),
		Msg("  Slow   - 3 fights per batch", "narrative"),
		Msg("  Normal - 5 fights per batch", "narrative"),
		Msg("  Fast   - 10 fights per batch", "narrative"),
		Msg("  Turbo  - 20 fights per batch", "narrative"),
		Msg("", "system"),
		Msg("HOW IT WORKS:", "system"),
		Msg("  1. Select Auto-Play from the main menu", "narrative"),
		Msg("  2. Choose a speed", "narrative"),
		Msg("  3. Fights run automatically at your first huntable location", "narrative"),
		Msg("  4. Results show wins, losses, XP gained, and loot", "narrative"),
		Msg("  5. Use 'Resume Auto-Play' to keep going at the same speed", "narrative"),
		Msg("", "system"),
		Msg("AUTO-PLAY MENU:", "system"),
		Msg("  Between batches you can:", "narrative"),
		Msg("  - View Inventory, Skills, Equipment, Stats", "narrative"),
		Msg("  - Check Quest Log", "narrative"),
		Msg("  - Resume Auto-Play for another batch", "narrative"),
		Msg("  - Return to Main Menu", "narrative"),
		Msg("", "system"),
		Msg("TIPS:", "system"),
		Msg("  Auto-Play is great for farming XP and beast materials", "narrative"),
		Msg("  Your character auto-resurrects on death (with a death counter)", "narrative"),
		Msg("  Equipment auto-equips if better than what you have", "narrative"),
		Msg("  15% chance to rescue a villager per victory", "narrative"),
		Msg("  Quest progress updates automatically during auto-play", "narrative"),
	}
}

func guideContentQuests() []GameMessage {
	return []GameMessage{
		Msg("=== QUESTS ===", "system"),
		Msg("", "system"),
		Msg("Quests give you goals and reward XP when completed.", "narrative"),
		Msg("New quests unlock automatically when you complete the previous one.", "narrative"),
		Msg("", "system"),
		Msg("QUEST TYPES:", "system"),
		Msg("  Level Quests    - Reach a certain character level", "narrative"),
		Msg("  Boss Kill Quests - Defeat specific boss monsters", "narrative"),
		Msg("  Location Quests - Discover and explore new areas", "narrative"),
		Msg("", "system"),
		Msg("QUEST CHAIN:", "system"),
		Msg("  The First Trial    - Reach level 5 (starting quest)", "narrative"),
		Msg("  Into the Ruins     - Explore the Forest Ruins", "narrative"),
		Msg("  Further quests unlock as you progress", "narrative"),
		Msg("", "system"),
		Msg("REWARDS:", "system"),
		Msg("  Each quest gives bonus XP on completion", "narrative"),
		Msg("  Some quests unlock new locations to explore", "narrative"),
		Msg("", "system"),
		Msg("VIEWING QUESTS:", "system"),
		Msg("  Use 'Quest Log' from the main menu to see:", "narrative"),
		Msg("  - Active quests with progress", "narrative"),
		Msg("  - Completed quests", "narrative"),
		Msg("  Quest progress updates automatically as you play", "narrative"),
	}
}
