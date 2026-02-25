package data

// MonsterCategory maps each monster name to its archetype category.
// Categories drive default resistances, skill pools, and material drops.
var MonsterCategory = map[string]string{
	// ── Beast (12) ──
	"wolf":           "beast",
	"dire bear":      "beast",
	"giant spider":   "beast",
	"giant scorpion": "beast",
	"griffin":        "beast",
	"manticore":      "beast",
	"chimera":        "beast",
	"wyvern":         "beast",
	"basilisk":       "beast",
	"cockatrice":     "beast",
	"hydra":          "beast",
	"dire boar":      "beast",

	// ── Undead (12) ──
	"skeleton":      "undead",
	"zombie":        "undead",
	"ghoul":         "undead",
	"wraith":        "undead",
	"revenant":      "undead",
	"banshee":       "undead",
	"death knight":  "undead",
	"mummy":         "undead",
	"shade":         "undead",
	"wight":         "undead",
	"bone colossus": "undead",
	"lich":          "undead",

	// ── Elemental (10) ──
	"fire elemental":  "elemental",
	"frost phantom":   "elemental",
	"storm elemental": "elemental",
	"earth elemental": "elemental",
	"magma brute":     "elemental",
	"thunderbird":     "elemental",
	"dust devil":      "elemental",
	"water serpent":   "elemental",
	"lava golem":      "elemental",
	"zephyr":          "elemental",

	// ── Construct (8) ──
	"gargoyle":           "construct",
	"iron golem":         "construct",
	"stone sentinel":     "construct",
	"clockwork soldier":  "construct",
	"bronze colossus":    "construct",
	"crystal guardian":   "construct",
	"obsidian automaton": "construct",
	"runic construct":    "construct",

	// ── Demon (12) ──
	"imp":            "demon",
	"succubus":       "demon",
	"pit fiend":      "demon",
	"hell hound":     "demon",
	"shadow fiend":   "demon",
	"balor":          "demon",
	"dretch":         "demon",
	"infernal brute": "demon",
	"abyssal horror": "demon",
	"pain devil":     "demon",
	"chaos spawn":    "demon",
	"tormentor":      "demon",

	// ── Dragon (10) ──
	"drake":          "dragon",
	"fire dragon":    "dragon",
	"frost wyrm":     "dragon",
	"storm dragon":   "dragon",
	"shadow dragon":  "dragon",
	"elder wyrm":     "dragon",
	"sea serpent":    "dragon",
	"lindworm":       "dragon",
	"amphisbaena":    "dragon",
	"dragon turtle":  "dragon",

	// ── Fey (8) ──
	"pixie":       "fey",
	"dryad":       "fey",
	"treant":      "fey",
	"satyr":       "fey",
	"will-o-wisp": "fey",
	"redcap":      "fey",
	"boggart":     "fey",
	"spriggan":    "fey",

	// ── Aberration (10) ──
	"mind flayer":       "aberration",
	"beholder":          "aberration",
	"gibbering mouther": "aberration",
	"aboleth":           "aberration",
	"nothic":            "aberration",
	"umber hulk":        "aberration",
	"carrion crawler":   "aberration",
	"otyugh":            "aberration",
	"hook horror":       "aberration",
	"brain eater":       "aberration",

	// ── Humanoid (10) ──
	"kobold":   "humanoid",
	"goblin":   "humanoid",
	"orc":      "humanoid",
	"troll":    "humanoid",
	"ogre":     "humanoid",
	"minotaur": "humanoid",
	"harpy":    "humanoid",
	"gnoll":    "humanoid",
	"bugbear":  "humanoid",
	"cyclops":  "humanoid",

	// ── Plant / Ooze (8) ──
	"ooze":            "plant",
	"shambling mound": "plant",
	"myconid":         "plant",
	"blight":          "plant",
	"fungal horror":   "plant",
	"vine lurker":     "plant",
	"corpse flower":   "plant",
	"rot grub":        "plant",
}

// MonsterNames is the ordered list of all available monster types.
var MonsterNames = []string{
	// Beast
	"wolf", "dire bear", "giant spider", "giant scorpion",
	"griffin", "manticore", "chimera", "wyvern",
	"basilisk", "cockatrice", "hydra", "dire boar",
	// Undead
	"skeleton", "zombie", "ghoul", "wraith",
	"revenant", "banshee", "death knight", "mummy",
	"shade", "wight", "bone colossus", "lich",
	// Elemental
	"fire elemental", "frost phantom", "storm elemental", "earth elemental",
	"magma brute", "thunderbird", "dust devil", "water serpent",
	"lava golem", "zephyr",
	// Construct
	"gargoyle", "iron golem", "stone sentinel", "clockwork soldier",
	"bronze colossus", "crystal guardian", "obsidian automaton", "runic construct",
	// Demon
	"imp", "succubus", "pit fiend", "hell hound",
	"shadow fiend", "balor", "dretch", "infernal brute",
	"abyssal horror", "pain devil", "chaos spawn", "tormentor",
	// Dragon
	"drake", "fire dragon", "frost wyrm", "storm dragon",
	"shadow dragon", "elder wyrm", "sea serpent", "lindworm",
	"amphisbaena", "dragon turtle",
	// Fey
	"pixie", "dryad", "treant", "satyr",
	"will-o-wisp", "redcap", "boggart", "spriggan",
	// Aberration
	"mind flayer", "beholder", "gibbering mouther", "aboleth",
	"nothic", "umber hulk", "carrion crawler", "otyugh",
	"hook horror", "brain eater",
	// Humanoid
	"kobold", "goblin", "orc", "troll",
	"ogre", "minotaur", "harpy", "gnoll",
	"bugbear", "cyclops",
	// Plant / Ooze
	"ooze", "shambling mound", "myconid", "blight",
	"fungal horror", "vine lurker", "corpse flower", "rot grub",
}
