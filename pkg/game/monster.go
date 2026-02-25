package game

import (
	"fmt"
	"math/rand"

	"rpg-game/pkg/data"
	"rpg-game/pkg/models"
)

// ────────────────────────────────────────────────────────────────────────────
// Category-based default resistances
// ────────────────────────────────────────────────────────────────────────────

var categoryResistances = map[string]map[models.DamageType]float64{
	"beast": {
		models.Physical: 1.0, models.Fire: 1.0, models.Ice: 1.0,
		models.Lightning: 1.0, models.Poison: 1.0,
	},
	"undead": {
		models.Physical: 0.5, models.Fire: 1.5, models.Ice: 1.0,
		models.Lightning: 1.0, models.Poison: 0.25,
	},
	"elemental": {
		models.Physical: 0.5, models.Fire: 1.0, models.Ice: 1.0,
		models.Lightning: 1.0, models.Poison: 0.25,
	},
	"construct": {
		models.Physical: 0.25, models.Fire: 1.0, models.Ice: 1.0,
		models.Lightning: 2.0, models.Poison: 0.25,
	},
	"demon": {
		models.Physical: 0.8, models.Fire: 0.25, models.Ice: 1.5,
		models.Lightning: 1.0, models.Poison: 0.5,
	},
	"dragon": {
		models.Physical: 0.8, models.Fire: 0.5, models.Ice: 1.5,
		models.Lightning: 1.0, models.Poison: 1.0,
	},
	"fey": {
		models.Physical: 1.5, models.Fire: 1.0, models.Ice: 0.5,
		models.Lightning: 0.5, models.Poison: 0.5,
	},
	"aberration": {
		models.Physical: 0.8, models.Fire: 1.0, models.Ice: 1.0,
		models.Lightning: 0.5, models.Poison: 0.5,
	},
	"humanoid": {
		models.Physical: 1.0, models.Fire: 1.0, models.Ice: 1.0,
		models.Lightning: 1.0, models.Poison: 1.0,
	},
	"plant": {
		models.Physical: 0.5, models.Fire: 2.0, models.Ice: 1.5,
		models.Lightning: 1.0, models.Poison: 0.25,
	},
}

// Per-monster resistance overrides (only monsters that differ from category defaults).
var monsterResistanceOverrides = map[string]map[models.DamageType]float64{
	// ── Beast specials ──
	"giant spider":   {models.Poison: 0.25, models.Fire: 1.5},
	"giant scorpion": {models.Poison: 0.25, models.Physical: 0.5},
	"basilisk":       {models.Physical: 0.5, models.Ice: 2.0, models.Poison: 0.25},
	"hydra":          {models.Fire: 1.5, models.Physical: 0.5},
	"wyvern":         {models.Fire: 0.5, models.Lightning: 0.5, models.Ice: 1.5},
	"chimera":        {models.Fire: 0.5, models.Poison: 0.5},
	"manticore":      {models.Poison: 0.25, models.Physical: 0.8},
	"cockatrice":     {models.Physical: 0.8, models.Lightning: 1.5},

	// ── Undead specials ──
	"wraith":        {models.Physical: 0.25, models.Fire: 2.0, models.Lightning: 0.5},
	"lich":          {models.Physical: 0.25, models.Fire: 0.5, models.Lightning: 0.5, models.Ice: 0.5, models.Poison: 0.25},
	"death knight":  {models.Physical: 0.5, models.Fire: 0.8},
	"banshee":       {models.Physical: 0.25, models.Lightning: 2.0},
	"bone colossus": {models.Physical: 0.5, models.Lightning: 2.0},
	"shade":         {models.Physical: 0.25, models.Fire: 2.0},

	// ── Elemental specials ──
	"fire elemental":  {models.Fire: 0.0, models.Ice: 2.0},
	"frost phantom":   {models.Ice: 0.0, models.Fire: 2.0},
	"storm elemental": {models.Lightning: 0.0, models.Physical: 0.5},
	"earth elemental": {models.Physical: 0.25, models.Lightning: 2.0},
	"magma brute":     {models.Fire: 0.0, models.Ice: 2.0, models.Physical: 0.5},
	"thunderbird":     {models.Lightning: 0.0, models.Ice: 1.5},
	"lava golem":      {models.Fire: 0.0, models.Ice: 2.0, models.Physical: 0.25},
	"water serpent":   {models.Ice: 0.5, models.Fire: 0.5, models.Lightning: 2.0},
	"zephyr":          {models.Physical: 0.25, models.Lightning: 0.5},

	// ── Construct specials ──
	"crystal guardian":   {models.Lightning: 0.5, models.Physical: 0.5},
	"obsidian automaton": {models.Fire: 0.25, models.Ice: 1.5},
	"iron golem":        {models.Lightning: 2.0, models.Fire: 0.5},

	// ── Demon specials ──
	"balor":          {models.Fire: 0.0, models.Ice: 2.0},
	"hell hound":     {models.Fire: 0.0, models.Ice: 2.0},
	"shadow fiend":   {models.Physical: 0.25, models.Fire: 1.5},
	"pit fiend":      {models.Fire: 0.0, models.Physical: 0.5},
	"abyssal horror": {models.Poison: 0.0, models.Fire: 0.5},

	// ── Dragon specials ──
	"fire dragon":   {models.Fire: 0.0, models.Ice: 2.0},
	"frost wyrm":    {models.Ice: 0.0, models.Fire: 2.0},
	"storm dragon":  {models.Lightning: 0.0, models.Ice: 1.0},
	"shadow dragon": {models.Physical: 0.5, models.Fire: 1.0, models.Poison: 0.25},
	"elder wyrm":    {models.Physical: 0.5, models.Fire: 0.25, models.Ice: 0.5, models.Lightning: 0.5},
	"dragon turtle":  {models.Physical: 0.25, models.Ice: 0.5},
	"sea serpent":   {models.Ice: 0.5, models.Lightning: 2.0, models.Fire: 0.5},

	// ── Fey specials ──
	"treant":      {models.Fire: 2.0, models.Physical: 0.5, models.Poison: 0.25},
	"will-o-wisp": {models.Physical: 0.0, models.Fire: 0.5, models.Lightning: 0.5},
	"pixie":       {models.Physical: 2.0, models.Lightning: 0.25},
	"dryad":       {models.Fire: 2.0, models.Poison: 0.25},

	// ── Aberration specials ──
	"mind flayer": {models.Physical: 0.5, models.Lightning: 0.25},
	"beholder":    {models.Physical: 0.5, models.Lightning: 0.25, models.Fire: 0.5},
	"aboleth":     {models.Fire: 0.5, models.Ice: 0.5, models.Lightning: 2.0},

	// ── Humanoid specials ──
	"troll":    {models.Fire: 2.0, models.Poison: 0.5},
	"minotaur": {models.Fire: 0.8},
	"cyclops":  {models.Physical: 0.8, models.Lightning: 1.5},
	"ogre":     {models.Physical: 0.8},
	"harpy":    {models.Lightning: 2.0, models.Ice: 0.5},

	// ── Plant/Ooze specials ──
	"ooze":            {models.Physical: 0.5, models.Fire: 2.0},
	"shambling mound": {models.Lightning: 0.0, models.Fire: 2.0}, // absorbs lightning
	"myconid":         {models.Poison: 0.0},
}

// ────────────────────────────────────────────────────────────────────────────
// Category-based skill pools
// ────────────────────────────────────────────────────────────────────────────

type levelSkill struct {
	MinLevel int
	Skill    models.Skill
}

var categorySkillPools = map[string][]levelSkill{
	"beast": {
		{2, models.Skill{Name: "Pounce", StaminaCost: 10, Damage: 14, DamageType: models.Physical,
			Effect: models.StatusEffect{Type: "none"}, Description: "Leaps onto prey with savage force"}},
		{4, models.Skill{Name: "Frenzy", StaminaCost: 18, Damage: 20, DamageType: models.Physical,
			Effect: models.StatusEffect{Type: "none"}, Description: "A whirlwind of claws and fangs"}},
		{6, models.Skill{Name: "Venomous Bite", StaminaCost: 12, Damage: 10, DamageType: models.Poison,
			Effect: models.StatusEffect{Type: "poison", Duration: 3, Potency: 4}, Description: "Sinks envenomed fangs deep"}},
	},
	"undead": {
		{2, models.Skill{Name: "Life Drain", ManaCost: 10, Damage: 12, DamageType: models.Lightning,
			Effect: models.StatusEffect{Type: "none"}, Description: "Siphons the essence of the living"}},
		{4, models.Skill{Name: "Bone Shards", ManaCost: 8, Damage: 16, DamageType: models.Physical,
			Effect: models.StatusEffect{Type: "none"}, Description: "Launches jagged bone fragments"}},
		{7, models.Skill{Name: "Terrify", ManaCost: 15, Damage: 5, DamageType: models.Lightning,
			Effect: models.StatusEffect{Type: "stun", Duration: 1, Potency: 1}, Description: "Freezes the soul with dread"}},
	},
	"elemental": {
		{2, models.Skill{Name: "Elemental Blast", ManaCost: 8, Damage: 14, DamageType: models.Fire,
			Effect: models.StatusEffect{Type: "burn", Duration: 2, Potency: 3}, Description: "Hurls a bolt of raw elemental energy"}},
		{4, models.Skill{Name: "Elemental Shield", ManaCost: 12, Damage: 0, DamageType: models.Physical,
			Effect: models.StatusEffect{Type: "buff_defense", Duration: 3, Potency: 12}, Description: "Wraps in a shell of elemental force"}},
		{7, models.Skill{Name: "Eruption", ManaCost: 18, Damage: 22, DamageType: models.Fire,
			Effect: models.StatusEffect{Type: "burn", Duration: 2, Potency: 4}, Description: "The ground splits with elemental fury"}},
	},
	"construct": {
		{3, models.Skill{Name: "Stone Skin", ManaCost: 10, Damage: 0, DamageType: models.Physical,
			Effect: models.StatusEffect{Type: "buff_defense", Duration: 4, Potency: 15}, Description: "Hardens into near-impenetrable material"}},
		{5, models.Skill{Name: "Crushing Slam", StaminaCost: 16, Damage: 18, DamageType: models.Physical,
			Effect: models.StatusEffect{Type: "stun", Duration: 1, Potency: 1}, Description: "Brings massive fists down with shattering force"}},
		{7, models.Skill{Name: "Overclock", ManaCost: 14, Damage: 0, DamageType: models.Physical,
			Effect: models.StatusEffect{Type: "buff_attack", Duration: 3, Potency: 10}, Description: "Internal mechanisms surge with power"}},
	},
	"demon": {
		{2, models.Skill{Name: "Hellfire", ManaCost: 10, Damage: 14, DamageType: models.Fire,
			Effect: models.StatusEffect{Type: "burn", Duration: 2, Potency: 3}, Description: "Conjures flames from the infernal plane"}},
		{4, models.Skill{Name: "Soul Rend", ManaCost: 14, Damage: 18, DamageType: models.Lightning,
			Effect: models.StatusEffect{Type: "none"}, Description: "Tears at the very soul"}},
		{6, models.Skill{Name: "Corrupt", ManaCost: 12, Damage: 8, DamageType: models.Poison,
			Effect: models.StatusEffect{Type: "poison", Duration: 4, Potency: 5}, Description: "Infects with abyssal corruption"}},
		{8, models.Skill{Name: "Infernal Roar", StaminaCost: 18, Damage: 0, DamageType: models.Physical,
			Effect: models.StatusEffect{Type: "buff_attack", Duration: 3, Potency: 12}, Description: "A roar that shakes the planes"}},
	},
	"dragon": {
		{2, models.Skill{Name: "Breath Attack", ManaCost: 10, Damage: 16, DamageType: models.Fire,
			Effect: models.StatusEffect{Type: "burn", Duration: 2, Potency: 3}, Description: "Exhales a devastating torrent"}},
		{4, models.Skill{Name: "Tail Swipe", StaminaCost: 12, Damage: 14, DamageType: models.Physical,
			Effect: models.StatusEffect{Type: "none"}, Description: "A sweeping blow from a massive tail"}},
		{6, models.Skill{Name: "Wing Buffet", StaminaCost: 14, Damage: 10, DamageType: models.Physical,
			Effect: models.StatusEffect{Type: "stun", Duration: 1, Potency: 1}, Description: "Batters foes with powerful wings"}},
		{8, models.Skill{Name: "Dragon Fear", ManaCost: 16, Damage: 8, DamageType: models.Lightning,
			Effect: models.StatusEffect{Type: "stun", Duration: 2, Potency: 1}, Description: "The ancient terror of dragonkind"}},
	},
	"fey": {
		{2, models.Skill{Name: "Charm", ManaCost: 8, Damage: 4, DamageType: models.Lightning,
			Effect: models.StatusEffect{Type: "stun", Duration: 1, Potency: 1}, Description: "Beguiles the mind with fey glamour"}},
		{4, models.Skill{Name: "Nature's Wrath", ManaCost: 12, Damage: 12, DamageType: models.Poison,
			Effect: models.StatusEffect{Type: "poison", Duration: 3, Potency: 3}, Description: "The wild strikes back"}},
		{6, models.Skill{Name: "Healing Light", ManaCost: 14, Damage: -20, DamageType: models.Physical,
			Effect: models.StatusEffect{Type: "regen", Duration: 3, Potency: 4}, Description: "Bathes in restorative moonlight"}},
	},
	"aberration": {
		{2, models.Skill{Name: "Psychic Blast", ManaCost: 12, Damage: 16, DamageType: models.Lightning,
			Effect: models.StatusEffect{Type: "none"}, Description: "A searing wave of psychic energy"}},
		{4, models.Skill{Name: "Tentacle Slam", StaminaCost: 10, Damage: 14, DamageType: models.Physical,
			Effect: models.StatusEffect{Type: "none"}, Description: "Lashes out with writhing appendages"}},
		{7, models.Skill{Name: "Mind Shatter", ManaCost: 18, Damage: 12, DamageType: models.Lightning,
			Effect: models.StatusEffect{Type: "stun", Duration: 2, Potency: 1}, Description: "Fractures the psyche"}},
	},
	"humanoid": {
		{2, models.Skill{Name: "Cleave", StaminaCost: 10, Damage: 14, DamageType: models.Physical,
			Effect: models.StatusEffect{Type: "none"}, Description: "A broad sweeping strike"}},
		{4, models.Skill{Name: "Shield Bash", StaminaCost: 12, Damage: 8, DamageType: models.Physical,
			Effect: models.StatusEffect{Type: "stun", Duration: 1, Potency: 1}, Description: "Rams with a heavy shield"}},
		{6, models.Skill{Name: "War Cry", StaminaCost: 15, Damage: 0, DamageType: models.Physical,
			Effect: models.StatusEffect{Type: "buff_attack", Duration: 3, Potency: 8}, Description: "A rallying battle shout"}},
	},
	"plant": {
		{2, models.Skill{Name: "Acid Splash", ManaCost: 6, Damage: 10, DamageType: models.Poison,
			Effect: models.StatusEffect{Type: "poison", Duration: 3, Potency: 3}, Description: "Sprays corrosive fluids"}},
		{4, models.Skill{Name: "Entangle", ManaCost: 10, Damage: 4, DamageType: models.Physical,
			Effect: models.StatusEffect{Type: "stun", Duration: 1, Potency: 1}, Description: "Vines and tendrils bind the target"}},
		{6, models.Skill{Name: "Spore Cloud", ManaCost: 14, Damage: 8, DamageType: models.Poison,
			Effect: models.StatusEffect{Type: "poison", Duration: 4, Potency: 5}, Description: "Releases a choking cloud of toxic spores"}},
		{3, models.Skill{Name: "Regenerate", ManaCost: 10, Damage: 0, DamageType: models.Physical,
			Effect: models.StatusEffect{Type: "regen", Duration: 4, Potency: 5}, Description: "Regrows damaged tissue rapidly"}},
	},
}

// Per-monster skill overrides for iconic creatures that deserve unique abilities.
var monsterSkillOverrides = map[string][]levelSkill{
	"hydra": {
		{2, models.Skill{Name: "Multi-Bite", StaminaCost: 14, Damage: 20, DamageType: models.Physical,
			Effect: models.StatusEffect{Type: "none"}, Description: "Multiple heads strike simultaneously"}},
		{5, models.Skill{Name: "Head Regrowth", ManaCost: 12, Damage: 0, DamageType: models.Physical,
			Effect: models.StatusEffect{Type: "regen", Duration: 4, Potency: 6}, Description: "Severed heads regrow with terrifying speed"}},
	},
	"basilisk": {
		{3, models.Skill{Name: "Petrifying Gaze", ManaCost: 15, Damage: 5, DamageType: models.Lightning,
			Effect: models.StatusEffect{Type: "stun", Duration: 2, Potency: 1}, Description: "Turns flesh to stone with a glance"}},
		{5, models.Skill{Name: "Venomous Bite", StaminaCost: 10, Damage: 12, DamageType: models.Poison,
			Effect: models.StatusEffect{Type: "poison", Duration: 4, Potency: 5}, Description: "Sinks fangs dripping with lethal venom"}},
	},
	"lich": {
		{3, models.Skill{Name: "Necrotic Bolt", ManaCost: 12, Damage: 20, DamageType: models.Poison,
			Effect: models.StatusEffect{Type: "poison", Duration: 3, Potency: 5}, Description: "A bolt of death magic"}},
		{5, models.Skill{Name: "Power Word Stun", ManaCost: 20, Damage: 10, DamageType: models.Lightning,
			Effect: models.StatusEffect{Type: "stun", Duration: 2, Potency: 1}, Description: "Utters a word of absolute power"}},
		{7, models.Skill{Name: "Dark Pact", ManaCost: 16, Damage: 0, DamageType: models.Physical,
			Effect: models.StatusEffect{Type: "buff_attack", Duration: 4, Potency: 15}, Description: "Draws power from the realm of the dead"}},
	},
	"banshee": {
		{2, models.Skill{Name: "Wail", ManaCost: 12, Damage: 16, DamageType: models.Lightning,
			Effect: models.StatusEffect{Type: "stun", Duration: 1, Potency: 1}, Description: "A keening scream that stops the heart"}},
	},
	"death knight": {
		{3, models.Skill{Name: "Unholy Strike", StaminaCost: 14, Damage: 20, DamageType: models.Physical,
			Effect: models.StatusEffect{Type: "poison", Duration: 2, Potency: 4}, Description: "A cursed blade strike"}},
		{5, models.Skill{Name: "Death's Embrace", ManaCost: 16, Damage: 0, DamageType: models.Physical,
			Effect: models.StatusEffect{Type: "buff_defense", Duration: 4, Potency: 12}, Description: "Death itself shields the knight"}},
	},
	"balor": {
		{3, models.Skill{Name: "Flame Whip", ManaCost: 14, Damage: 22, DamageType: models.Fire,
			Effect: models.StatusEffect{Type: "burn", Duration: 3, Potency: 5}, Description: "Cracks a whip of living flame"}},
		{6, models.Skill{Name: "Demonic Roar", StaminaCost: 20, Damage: 0, DamageType: models.Physical,
			Effect: models.StatusEffect{Type: "buff_attack", Duration: 3, Potency: 15}, Description: "A roar that shakes reality"}},
	},
	"pit fiend": {
		{3, models.Skill{Name: "Inferno", ManaCost: 16, Damage: 20, DamageType: models.Fire,
			Effect: models.StatusEffect{Type: "burn", Duration: 3, Potency: 4}, Description: "Engulfs the area in hellfire"}},
		{6, models.Skill{Name: "Fear Aura", ManaCost: 14, Damage: 8, DamageType: models.Lightning,
			Effect: models.StatusEffect{Type: "stun", Duration: 2, Potency: 1}, Description: "An aura of absolute terror"}},
	},
	"fire dragon": {
		{2, models.Skill{Name: "Inferno Breath", ManaCost: 14, Damage: 22, DamageType: models.Fire,
			Effect: models.StatusEffect{Type: "burn", Duration: 3, Potency: 5}, Description: "A cone of white-hot dragonfire"}},
		{5, models.Skill{Name: "Tail Swipe", StaminaCost: 12, Damage: 16, DamageType: models.Physical,
			Effect: models.StatusEffect{Type: "none"}, Description: "A sweeping blow from a massive tail"}},
		{7, models.Skill{Name: "Dragon Fear", ManaCost: 18, Damage: 8, DamageType: models.Lightning,
			Effect: models.StatusEffect{Type: "stun", Duration: 2, Potency: 1}, Description: "The ancient terror of dragonkind"}},
	},
	"frost wyrm": {
		{2, models.Skill{Name: "Frost Breath", ManaCost: 14, Damage: 20, DamageType: models.Ice,
			Effect: models.StatusEffect{Type: "stun", Duration: 1, Potency: 1}, Description: "Exhales a blast of killing cold"}},
		{5, models.Skill{Name: "Ice Armor", ManaCost: 12, Damage: 0, DamageType: models.Physical,
			Effect: models.StatusEffect{Type: "buff_defense", Duration: 4, Potency: 14}, Description: "Encases in a shell of solid ice"}},
	},
	"elder wyrm": {
		{3, models.Skill{Name: "Ancient Breath", ManaCost: 18, Damage: 28, DamageType: models.Fire,
			Effect: models.StatusEffect{Type: "burn", Duration: 3, Potency: 6}, Description: "Primordial dragonfire that melts stone"}},
		{5, models.Skill{Name: "Tail Swipe", StaminaCost: 14, Damage: 18, DamageType: models.Physical,
			Effect: models.StatusEffect{Type: "stun", Duration: 1, Potency: 1}, Description: "A world-shaking blow"}},
		{7, models.Skill{Name: "Elder Majesty", ManaCost: 20, Damage: 10, DamageType: models.Lightning,
			Effect: models.StatusEffect{Type: "stun", Duration: 2, Potency: 1}, Description: "The weight of ages crushes the mind"}},
	},
	"mind flayer": {
		{2, models.Skill{Name: "Mind Blast", ManaCost: 14, Damage: 18, DamageType: models.Lightning,
			Effect: models.StatusEffect{Type: "stun", Duration: 1, Potency: 1}, Description: "A psychic shockwave"}},
		{5, models.Skill{Name: "Extract Brain", ManaCost: 20, Damage: 30, DamageType: models.Lightning,
			Effect: models.StatusEffect{Type: "none"}, Description: "Tentacles probe for the brain"}},
	},
	"beholder": {
		{2, models.Skill{Name: "Disintegration Ray", ManaCost: 14, Damage: 22, DamageType: models.Fire,
			Effect: models.StatusEffect{Type: "none"}, Description: "A thin beam that unmakes matter"}},
		{4, models.Skill{Name: "Charm Ray", ManaCost: 10, Damage: 4, DamageType: models.Lightning,
			Effect: models.StatusEffect{Type: "stun", Duration: 2, Potency: 1}, Description: "A beam that dominates the will"}},
		{6, models.Skill{Name: "Death Ray", ManaCost: 20, Damage: 30, DamageType: models.Lightning,
			Effect: models.StatusEffect{Type: "none"}, Description: "The dread eye of death opens"}},
	},
	"troll": {
		{3, models.Skill{Name: "Regenerate", ManaCost: 10, Damage: 0, DamageType: models.Physical,
			Effect: models.StatusEffect{Type: "regen", Duration: 4, Potency: 6}, Description: "Wounds knit shut with horrifying speed"}},
		{6, models.Skill{Name: "Frenzy", StaminaCost: 20, Damage: 22, DamageType: models.Physical,
			Effect: models.StatusEffect{Type: "none"}, Description: "A wild flurry of claws and teeth"}},
	},
	"minotaur": {
		{3, models.Skill{Name: "Gore", StaminaCost: 16, Damage: 22, DamageType: models.Physical,
			Effect: models.StatusEffect{Type: "none"}, Description: "Charges and impales with massive horns"}},
		{5, models.Skill{Name: "War Cry", StaminaCost: 15, Damage: 0, DamageType: models.Physical,
			Effect: models.StatusEffect{Type: "buff_attack", Duration: 3, Potency: 10}, Description: "A bellowing roar that stokes fury"}},
	},
	"treant": {
		{3, models.Skill{Name: "Root Slam", StaminaCost: 14, Damage: 18, DamageType: models.Physical,
			Effect: models.StatusEffect{Type: "stun", Duration: 1, Potency: 1}, Description: "Massive roots erupt from the earth"}},
		{5, models.Skill{Name: "Nature's Mend", ManaCost: 14, Damage: -25, DamageType: models.Physical,
			Effect: models.StatusEffect{Type: "regen", Duration: 4, Potency: 5}, Description: "The forest heals its guardian"}},
	},
	"succubus": {
		{2, models.Skill{Name: "Charm", ManaCost: 8, Damage: 4, DamageType: models.Lightning,
			Effect: models.StatusEffect{Type: "stun", Duration: 2, Potency: 1}, Description: "Irresistible supernatural allure"}},
		{5, models.Skill{Name: "Kiss of Death", ManaCost: 16, Damage: 24, DamageType: models.Poison,
			Effect: models.StatusEffect{Type: "none"}, Description: "Drains life with a fatal embrace"}},
	},
	"cyclops": {
		{3, models.Skill{Name: "Boulder Throw", StaminaCost: 16, Damage: 24, DamageType: models.Physical,
			Effect: models.StatusEffect{Type: "stun", Duration: 1, Potency: 1}, Description: "Hurls a massive stone"}},
	},
}

// ────────────────────────────────────────────────────────────────────────────
// Category-based material drops
// ────────────────────────────────────────────────────────────────────────────

type materialDrop struct {
	Materials  []string
	DropChance int
}

var categoryMaterials = map[string]materialDrop{
	"beast":      {[]string{"Beast Skin", "Beast Bone", "Tough Hide", "Sharp Fang"}, 50},
	"undead":     {[]string{"Beast Bone", "Ore Fragment", "Monster Claw"}, 45},
	"elemental":  {[]string{"Ore Fragment", "Beast Bone"}, 55},
	"construct":  {[]string{"Ore Fragment", "Beast Bone", "Iron"}, 65},
	"demon":      {[]string{"Monster Claw", "Sharp Fang", "Beast Bone"}, 55},
	"dragon":     {[]string{"Tough Hide", "Sharp Fang", "Beast Skin", "Monster Claw"}, 60},
	"fey":        {[]string{"Beast Skin", "Ore Fragment"}, 40},
	"aberration": {[]string{"Monster Claw", "Beast Bone", "Ore Fragment"}, 50},
	"humanoid":   {[]string{"Beast Bone", "Tough Hide", "Sharp Fang"}, 50},
	"plant":      {[]string{"Beast Skin", "Ore Fragment"}, 45},
}

// Per-monster material overrides.
var monsterMaterialOverrides = map[string]materialDrop{
	"hydra":       {[]string{"Tough Hide", "Sharp Fang", "Monster Claw"}, 75},
	"elder wyrm":  {[]string{"Tough Hide", "Sharp Fang", "Monster Claw", "Beast Bone"}, 80},
	"fire dragon": {[]string{"Tough Hide", "Sharp Fang", "Beast Skin"}, 70},
	"lich":        {[]string{"Ore Fragment", "Monster Claw", "Beast Bone"}, 70},
	"balor":       {[]string{"Monster Claw", "Sharp Fang", "Beast Bone"}, 70},
	"beholder":    {[]string{"Monster Claw", "Ore Fragment"}, 65},
	"mind flayer": {[]string{"Monster Claw", "Ore Fragment", "Beast Bone"}, 60},
	"troll":       {[]string{"Tough Hide", "Monster Claw"}, 55},
	"Guardian":    {[]string{"Tough Hide", "Sharp Fang", "Monster Claw"}, 80},
}

// ────────────────────────────────────────────────────────────────────────────
// Public API
// ────────────────────────────────────────────────────────────────────────────

func GenerateMonster(name string, level int, rank int) models.Monster {
	hitpoints := MultiRoll(rank)
	mana := MultiRoll(rank) + 10
	stamina := MultiRoll(rank) + 10

	resistances := buildResistances(name)

	monster := models.Monster{
		Name:               name,
		Level:              level,
		Experience:         0,
		Rank:               rank,
		HitpointsTotal:     hitpoints,
		HitpointsNatural:   hitpoints,
		HitpointsRemaining: hitpoints,
		ManaTotal:          mana,
		ManaRemaining:      mana,
		ManaNatural:        mana,
		StaminaTotal:       stamina,
		StaminaRemaining:   stamina,
		StaminaNatural:     stamina,
		AttackRolls:        rank,
		DefenseRolls:       rank,
		LearnedSkills:      AssignMonsterSkills(name, level),
		StatusEffects:      []models.StatusEffect{},
		Resistances:        resistances,
		MonsterType:        name,
	}
	monster.EquipmentMap = map[int]models.Item{}
	monster.Inventory = []models.Item{}
	for i := 0; i < (level/10)+rank-2; i++ {
		item := GenerateItem(rank)
		EquipBestItem(item, &monster.EquipmentMap, &monster.Inventory)
	}

	return monster
}

func GenerateBestMonster(game *models.GameState, levelMax int, rankMax int) models.Monster {
	name := data.MonsterNames[rand.Intn(len(data.MonsterNames))]
	fmt.Printf("LevelMax: %d, rankMax: %d\n", levelMax, rankMax)
	if levelMax == 0 {
		levelMax++
	}
	if rankMax == 0 {
		rankMax++
	}
	level := rand.Intn(levelMax) + 1
	rank := rand.Intn(rankMax) + 1
	var mob = GenerateMonster(name, level, rank)
	if rand.Intn(100) <= 1*rank {
		var item = GenerateItem(rank)
		EquipBestItem(item, &mob.EquipmentMap, &mob.Inventory)
		mob.EquipmentMap = map[int]models.Item{}
	}

	mob.StatsMod = CalculateItemMods(mob.EquipmentMap)
	mob.HitpointsTotal = mob.HitpointsNatural + mob.StatsMod.HitPointMod

	return mob
}

func GenerateSkillGuardian(skill models.Skill, level int, rank int) models.Monster {
	guardianName := data.SkillGuardianNames[rand.Intn(len(data.SkillGuardianNames))]
	baseMob := GenerateMonster(guardianName, level, rank)

	baseMob.HitpointsNatural = int(float64(baseMob.HitpointsNatural) * 2.0)
	baseMob.HitpointsTotal = baseMob.HitpointsNatural
	baseMob.HitpointsRemaining = baseMob.HitpointsTotal
	baseMob.AttackRolls = baseMob.AttackRolls + 2
	baseMob.DefenseRolls = baseMob.DefenseRolls + 2
	baseMob.ManaTotal = int(float64(baseMob.ManaTotal) * 1.5)
	baseMob.ManaRemaining = baseMob.ManaTotal
	baseMob.StaminaTotal = int(float64(baseMob.StaminaTotal) * 1.5)
	baseMob.StaminaRemaining = baseMob.StaminaTotal

	for i := 0; i < rank+2; i++ {
		item := GenerateItem(rank + 1)
		EquipBestItem(item, &baseMob.EquipmentMap, &baseMob.Inventory)
	}

	baseMob.IsSkillGuardian = true
	baseMob.GuardedSkill = skill
	baseMob.MonsterType = "Guardian"

	baseMob.StatsMod = CalculateItemMods(baseMob.EquipmentMap)
	baseMob.HitpointsTotal = baseMob.HitpointsNatural + baseMob.StatsMod.HitPointMod
	baseMob.HitpointsRemaining = baseMob.HitpointsTotal

	return baseMob
}

// AssignMonsterSkills returns skills for a monster based on type and level.
func AssignMonsterSkills(monsterType string, level int) []models.Skill {
	skills := []models.Skill{}

	// Check per-monster overrides first
	if pool, ok := monsterSkillOverrides[monsterType]; ok {
		for _, ls := range pool {
			if level >= ls.MinLevel {
				skills = append(skills, ls.Skill)
			}
		}
		return skills
	}

	// Fall back to category pool
	cat := data.MonsterCategory[monsterType]
	if cat == "" {
		cat = "humanoid" // safe default for guardians or unknown types
	}
	if pool, ok := categorySkillPools[cat]; ok {
		for _, ls := range pool {
			if level >= ls.MinLevel {
				skills = append(skills, ls.Skill)
			}
		}
	}
	return skills
}

func LevelUpMob(mob *models.Monster) {
	if mob.Experience >= (mob.Level * 100) {
		levelsToGrant := ((mob.Level * 100) - mob.ExpSinceLevel) / 100
		for i := 0; i < levelsToGrant; i++ {
			mob.Level++
			mob.HitpointsNatural += MultiRoll(1)
			mob.HitpointsRemaining = mob.HitpointsNatural
			mob.ManaNatural += MultiRoll(1) + 3
			mob.ManaTotal = mob.ManaNatural
			mob.ManaRemaining = mob.ManaTotal
			mob.StaminaNatural += MultiRoll(1) + 3
			mob.StaminaTotal = mob.StaminaNatural
			mob.StaminaRemaining = mob.StaminaTotal
			mob.AttackRolls = mob.Level/10 + 1
			mob.DefenseRolls = mob.Level/10 + 1
			mob.StatsMod = CalculateItemMods(mob.EquipmentMap)
			mob.HitpointsTotal = mob.HitpointsNatural + mob.StatsMod.HitPointMod
			fmt.Printf("%s leveled up to %d!\n", mob.Name, mob.Level)
		}
	}
}

// ────────────────────────────────────────────────────────────────────────────
// Internal helpers
// ────────────────────────────────────────────────────────────────────────────

func buildResistances(name string) map[models.DamageType]float64 {
	// Start with neutral defaults
	res := map[models.DamageType]float64{
		models.Physical:  1.0,
		models.Fire:      1.0,
		models.Ice:       1.0,
		models.Lightning: 1.0,
		models.Poison:    1.0,
	}

	// Apply category defaults
	cat := data.MonsterCategory[name]
	if catRes, ok := categoryResistances[cat]; ok {
		for dt, val := range catRes {
			res[dt] = val
		}
	}

	// Apply per-monster overrides
	if overrides, ok := monsterResistanceOverrides[name]; ok {
		for dt, val := range overrides {
			res[dt] = val
		}
	}

	return res
}
