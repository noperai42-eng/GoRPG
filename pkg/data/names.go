package data

var SkillGuardianNames = []string{
	"Ifrit",
	"Lich King",
	"Leviathan",
	"Nightstalker",
	"Titan of Stone",
	"Plague Wraith",
	"Warlord Kael",
	"Arcane Golem",
	"Cerberus",
	"Fenrir",
}

var VillagerFirstNames = []string{
	"Aldric", "Brenna", "Cedric", "Daria", "Ewan", "Freya", "Gareth", "Hilda",
	"Ivar", "Jorunn", "Kael", "Liriel", "Magnus", "Nessa", "Orin", "Petra",
	"Rowan", "Sigrid", "Theron", "Una", "Varin", "Wren", "Ysolde", "Zarek",
	"Alaric", "Britta", "Corwin", "Dagny", "Elric", "Fenn",
}

var VillagerLastNames = []string{
	"Ashford", "Blackwood", "Cragmere", "Dunholm", "Emberfall", "Frostwind",
	"Greymoor", "Holloway", "Ironforge", "Kestrel", "Longbarrow", "Moorwen",
	"Northgate", "Oakheart", "Pinecrest", "Ravencroft", "Stonehaven", "Thornwall",
	"Underhill", "Valesong", "Whitmore", "Yarrow", "Copperfield", "Duskmantle",
}

var MayorNames = []string{
	"Lord Aldric", "Magistrate Elara", "Governor Thane", "Regent Mira",
	"Steward Fenwick", "Provost Callum", "Chancellor Isolde", "Prefect Rowan",
	"Warden Lysander", "Consul Daphne",
}

var GuardNames = []string{
	"Ser Marcus", "Captain Elena", "Knight Roland", "Dame Victoria", "Guard Captain Thorne",
	"Paladin Cedric", "Sentinel Aria", "Defender Gareth", "Shield-Bearer Lyra", "Warden Drake",
}

// ────────────────────────────────────────────────────────────────────────────
// Item naming: Prefix + Slot-specific base name
// Total unique base gear names: 25 per slot × 8 slots = 200
// ────────────────────────────────────────────────────────────────────────────

var ItemPrefixes = []string{
	// Common
	"Rusty", "Worn", "Battered", "Crude", "Tarnished",
	// Standard metals
	"Iron", "Steel", "Bronze", "Copper", "Tin",
	// Quality
	"Tempered", "Forged", "Hardened", "Reinforced", "Polished",
	// Exotic metals
	"Silver", "Gilded", "Mithril", "Adamantine", "Cobalt",
	"Orichalcum", "Electrum", "Titanium", "Darksteel", "Starmetal",
	// Enchanted
	"Runic", "Enchanted", "Arcane", "Blessed", "Cursed",
	"Hallowed", "Hexed", "Warded", "Glyphbound", "Spellforged",
	// Elemental
	"Frostforged", "Flameforged", "Stormborn", "Earthbound", "Voidtouched",
	// Heritage
	"Ancient", "Ancestral", "Weathered", "Timeworn", "Relic",
	// Ownership
	"Warden's", "Knight's", "Soldier's", "Veteran's", "Champion's",
	"King's", "Lord's", "Sentinel's", "Crusader's", "Berserker's",
	// Appearance
	"Obsidian", "Ivory", "Crimson", "Ebony", "Jade",
	"Onyx", "Amber", "Sapphire", "Ruby", "Emerald",
	// Craft
	"Bloodforged", "Dusksteel", "Shadowmeld", "Dawnforged", "Dragonscale",
	"Wyrmbone", "Demonhide", "Spiritbound", "Soulforged", "Lichbone",
}

// SlotGearNames maps equipment slot index to 25 possible gear names. (200 total)
var SlotGearNames = map[int][]string{
	// 0 = Head
	0: {
		"Helm", "Greathelm", "Coif", "Crown", "Circlet",
		"Barbute", "Skullcap", "Visor", "Sallet", "Kettle Helm",
		"Armet", "Bascinet", "Casque", "Diadem", "Hood",
		"Cowl", "Headband", "Morion", "Burgonet", "War Crown",
		"Nasal Helm", "Spangenhelm", "Tiara", "Cap", "Faceguard",
	},
	// 1 = Chest
	1: {
		"Breastplate", "Cuirass", "Hauberk", "Chestguard", "Brigandine",
		"Chainmail", "Gambeson", "Surcoat", "Plate Armor", "Scale Mail",
		"Jerkin", "Tunic", "Vest", "Doublet", "Corselet",
		"Lorica", "Tabard", "War Coat", "Splint Mail", "Ring Mail",
		"Lamellar", "Harness", "Plastron", "Haubergeon", "Cuirie",
	},
	// 2 = Legs
	2: {
		"Greaves", "Legguards", "Cuisses", "Tassets", "Leggings",
		"Chausses", "Kilt", "Legplates", "Breeches", "Trousers",
		"War Skirt", "Schynbalds", "Poleyns", "Fauld", "Hose",
		"Pantaloons", "Leg Wraps", "Thigh Guards", "Shin Guards", "Splint Greaves",
		"Plate Legs", "Scale Leggings", "Padded Legs", "War Pants", "Leg Harness",
	},
	// 3 = Feet
	3: {
		"Sabatons", "Boots", "Treads", "Sollerets", "Sandals",
		"Warboots", "Striders", "Foot Wraps", "Moccasins", "Shoes",
		"Clogs", "Buskins", "Jackboots", "Riding Boots", "Ironshods",
		"Steel Boots", "Heavy Boots", "Plated Boots", "Fur Boots", "Scout Boots",
		"Plate Sabatons", "Marching Boots", "Stompers", "Greaves", "War Shoes",
	},
	// 4 = Hands
	4: {
		"Gauntlets", "Vambraces", "Bracers", "Gloves", "Handguards",
		"Grips", "Wraps", "Mitts", "Fists", "Cestus",
		"Wrist Guards", "War Gloves", "Knuckles", "Finger Guards", "Arm Guards",
		"Plate Gloves", "Chain Gloves", "Leather Gloves", "Iron Fists", "Claws",
		"Talons", "Grasps", "Dueling Gloves", "Battle Mitts", "War Wraps",
	},
	// 5 = Main Hand (Weapons)
	5: {
		"Longsword", "Battleaxe", "Warhammer", "Mace", "Claymore",
		"Falchion", "Halberd", "Spear", "Rapier", "Scimitar",
		"Greatsword", "War Pick", "Morningstar", "Flail", "Glaive",
		"Shortsword", "Broadsword", "Katana", "Zweihander", "Pike",
		"Trident", "Saber", "Cutlass", "Maul", "Estoc",
	},
	// 6 = Off Hand (Shields / Focus)
	6: {
		"Buckler", "Tower Shield", "Kite Shield", "Round Shield", "Pavise",
		"Heater Shield", "Tome", "Ward Focus", "Lantern", "Parrying Dagger",
		"Orb", "Crystal Ball", "Targe", "Aegis", "Bulwark",
		"Rampart", "War Board", "Spell Book", "Rune Stone", "Talisman",
		"Fetish", "Totem", "Sigil", "Escutcheon", "Deflector",
	},
	// 7 = Accessory
	7: {
		"Amulet", "Ring", "Pendant", "Talisman", "Brooch",
		"Signet", "Torc", "Charm", "Locket", "Medallion",
		"Earring", "Bracelet", "Anklet", "Phylactery", "Relic",
		"Token", "Emblem", "Badge", "Seal", "Idol",
		"Effigy", "Ward", "Bead", "Cameo", "Scarab",
	},
}
