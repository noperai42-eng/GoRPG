package data

import "rpg-game/pkg/models"

var DiscoverableLocations = []models.Location{
	{Name: "Home", Weight: 0, Type: "Base"},
	{Name: "Training Hall", Weight: 0, Type: "Mix", LevelMax: 10, RarityMax: 1},
	{Name: "Forest", Weight: 0, Type: "Mix", LevelMax: 20, RarityMax: 2},
	{Name: "Lake", Weight: 0, Type: "Mix", LevelMax: 20, RarityMax: 2},
	{Name: "Hills", Weight: 0, Type: "Mix", LevelMax: 20, RarityMax: 2},
	{Name: "Hunters Lodge", Weight: 12, Type: "Ruin", LevelMax: 50, RarityMax: 3},
	{Name: "Forest Ruins", Weight: 14, Type: "Ruin", LevelMax: 100, RarityMax: 5},
	{Name: "Lake Ruins", Weight: 24, Type: "Ruin", LevelMax: 50, RarityMax: 3},
	{Name: "Quarry", Weight: 28, Type: "Resource", LevelMax: 30, RarityMax: 2},
	{Name: "Stone Keep", Weight: 33, Type: "Base"},
	{Name: "Training Hub", Weight: 37, Type: "Base"},
	{Name: "Kegger", Weight: 42, Type: "Trade", LevelMax: 20, RarityMax: 1},
	{Name: "Hospital", Weight: 53, Type: "Base"},
	{Name: "Ancient Dungeon", Weight: 55, Type: "Ruin", LevelMax: 200, RarityMax: 10},
	{Name: "The Tower", Weight: 60, Type: "Ruin", LevelMax: 2000, RarityMax: 10},
	{Name: "Godbeast Domain", Weight: 58, Type: "Ruin"},
}

var AvailableBuildings = []models.Building{
	{Name: "Training Grounds", RequiredResourceMap: map[string]int{"Lumber": 30, "Stone": 10}, StatsMod: models.StatMod{AttackMod: 0, DefenseMod: 0, HitPointMod: 0}},
	{Name: "Blacksmith", RequiredResourceMap: map[string]int{"Lumber": 10, "Stone": 30}, StatsMod: models.StatMod{AttackMod: 0, DefenseMod: 0, HitPointMod: 0}},
}
