package data

// DungeonTemplate defines a dungeon's base parameters.
type DungeonTemplate struct {
	Name     string
	MinLevel int
	MaxLevel int
	RankMax  int
	Floors   int
}

// DungeonTemplates contains all available dungeon templates.
var DungeonTemplates = []DungeonTemplate{
	{Name: "Goblin Warren", MinLevel: 1, MaxLevel: 10, RankMax: 2, Floors: 5},
	{Name: "Forgotten Crypt", MinLevel: 5, MaxLevel: 25, RankMax: 3, Floors: 10},
	{Name: "Dragon's Lair", MinLevel: 15, MaxLevel: 50, RankMax: 5, Floors: 15},
	{Name: "The Abyss", MinLevel: 30, MaxLevel: 100, RankMax: 7, Floors: 20},
	{Name: "Tower of Eternity", MinLevel: 50, MaxLevel: 200, RankMax: 10, Floors: 50},
}
