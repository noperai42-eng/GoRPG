package game

import (
	"math"
	"math/rand"

	"rpg-game/pkg/data"
	"rpg-game/pkg/models"
)

// GenerateDungeon creates a dungeon deterministically from a template and seed.
// Each floor has 5-8 rooms with distribution: 50% combat, 15% treasure, 10% trap,
// 10% rest, 10% merchant, 5% boss. Boss floors occur every 5th floor.
// Floor N scales monster levels and rarity weights progressively.
func GenerateDungeon(template data.DungeonTemplate, seed int64) models.Dungeon {
	src := rand.NewSource(seed)
	rng := rand.New(src)

	dungeon := models.Dungeon{
		Name:         template.Name,
		CurrentFloor: 0,
		BaseLevelMin: template.MinLevel,
		BaseLevelMax: template.MaxLevel,
		BaseRankMax:  template.RankMax,
		Seed:         seed,
		Floors:       make([]models.DungeonFloor, template.Floors),
	}

	for f := 0; f < template.Floors; f++ {
		floorNum := f + 1
		isBossFloor := floorNum%5 == 0

		// Scale level and rank by floor progression
		floorScale := 1.0 + float64(floorNum)*0.2
		scaledLevelMax := int(math.Ceil(float64(template.MaxLevel) * floorScale))
		scaledRankMax := template.RankMax + (floorNum / 3)
		if scaledRankMax > 10 {
			scaledRankMax = 10
		}

		// Determine number of rooms: 5-8
		numRooms := 5 + rng.Intn(4)

		floor := models.DungeonFloor{
			FloorNumber: floorNum,
			Rooms:       make([]models.DungeonRoom, numRooms),
			CurrentRoom: 0,
			Cleared:     false,
			BossFloor:   isBossFloor,
		}

		for r := 0; r < numRooms; r++ {
			// Boss floor: last room is always a boss room
			if isBossFloor && r == numRooms-1 {
				boss := GenerateDungeonBoss(floorNum, scaledLevelMax, scaledRankMax)
				floor.Rooms[r] = models.DungeonRoom{
					Type:    "boss",
					Cleared: false,
					Monster: &boss,
				}
				continue
			}

			roomType := rollRoomType(rng)
			room := models.DungeonRoom{
				Type:    roomType,
				Cleared: false,
			}

			switch roomType {
			case "combat":
				mob := generateDungeonMonster(rng, scaledLevelMax, scaledRankMax)
				room.Monster = &mob

			case "treasure":
				numItems := rng.Intn(3) + 1
				loot := make([]models.Item, numItems)
				itemRarity := scaledRankMax
				if itemRarity < 1 {
					itemRarity = 1
				}
				for i := 0; i < numItems; i++ {
					loot[i] = GenerateItem(itemRarity)
				}
				room.Loot = loot

			case "trap":
				// Trap damage scales with floor number: base 5 + floor * 3
				room.TrapDamage = 5 + floorNum*3

			case "rest":
				// Heal amount scales with floor: base 15 + floor * 5
				room.HealAmount = 15 + floorNum*5

			case "merchant":
				// Merchants have items available as loot for purchase
				numItems := rng.Intn(3) + 2
				loot := make([]models.Item, numItems)
				for i := 0; i < numItems; i++ {
					if rng.Intn(3) == 0 {
						sizes := []string{"small", "medium", "large"}
						loot[i] = CreateHealthPotion(sizes[rng.Intn(len(sizes))])
					} else {
						loot[i] = GenerateItem(scaledRankMax)
					}
				}
				room.Loot = loot
			}

			floor.Rooms[r] = room
		}

		dungeon.Floors[f] = floor
	}

	return dungeon
}

// GenerateDungeonBoss creates a boss monster with amplified stats.
// The boss is generated using GenerateMonster, then receives IsBoss=true,
// Rarity=Legendary, HP multiplied by 5x, and attack/defense multiplied by 3x.
func GenerateDungeonBoss(floor int, baseLevel int, baseRank int) models.Monster {
	// Pick a random monster name for the boss
	name := data.MonsterNames[rand.Intn(len(data.MonsterNames))]

	level := baseLevel
	if level < 1 {
		level = 1
	}
	rank := baseRank
	if rank < 1 {
		rank = 1
	}

	boss := GenerateMonster(name, level, rank)

	boss.IsBoss = true
	boss.Rarity = models.RarityLegendary

	// Multiply HP by 5x
	boss.HitpointsNatural *= 5
	boss.HitpointsTotal = boss.HitpointsNatural
	boss.HitpointsRemaining = boss.HitpointsTotal

	// Multiply attack and defense rolls by 3x
	boss.AttackRolls *= 3
	boss.DefenseRolls *= 3

	// Scale mana and stamina for boss encounters
	boss.ManaNatural *= 3
	boss.ManaTotal = boss.ManaNatural
	boss.ManaRemaining = boss.ManaTotal
	boss.StaminaNatural *= 3
	boss.StaminaTotal = boss.StaminaNatural
	boss.StaminaRemaining = boss.StaminaTotal

	// Recalculate total HP with equipment
	boss.StatsMod = CalculateItemMods(boss.EquipmentMap)
	boss.HitpointsTotal = boss.HitpointsNatural + boss.StatsMod.HitPointMod
	boss.HitpointsRemaining = boss.HitpointsTotal

	return boss
}

// AvailableDungeons returns all dungeon templates the player qualifies for
// based on their level (template MinLevel <= playerLevel).
func AvailableDungeons(playerLevel int) []data.DungeonTemplate {
	var available []data.DungeonTemplate
	for _, t := range data.DungeonTemplates {
		if t.MinLevel <= playerLevel {
			available = append(available, t)
		}
	}
	return available
}

// rollRoomType picks a room type using weighted random selection.
// Weights: combat 50, treasure 15, trap 10, rest 10, merchant 10, boss 5.
// Boss rooms from this roll are converted to combat (actual boss rooms are
// placed explicitly on boss floors).
func rollRoomType(rng *rand.Rand) string {
	type roomWeight struct {
		roomType string
		weight   int
	}
	weights := []roomWeight{
		{"combat", 50},
		{"treasure", 15},
		{"trap", 10},
		{"rest", 10},
		{"merchant", 10},
		{"combat", 5}, // boss weight rolls into combat for non-boss placements
	}

	total := 0
	for _, w := range weights {
		total += w.weight
	}

	roll := rng.Intn(total)
	cumulative := 0
	for _, w := range weights {
		cumulative += w.weight
		if roll < cumulative {
			return w.roomType
		}
	}
	return "combat"
}

// generateDungeonMonster creates a monster for a dungeon combat room using
// the seeded RNG for deterministic generation.
func generateDungeonMonster(rng *rand.Rand, levelMax int, rankMax int) models.Monster {
	name := data.MonsterNames[rng.Intn(len(data.MonsterNames))]

	if levelMax < 1 {
		levelMax = 1
	}
	if rankMax < 1 {
		rankMax = 1
	}

	level := rng.Intn(levelMax) + 1
	rank := rng.Intn(rankMax) + 1

	mob := GenerateMonster(name, level, rank)

	// Roll and apply rarity with floor-scaled weights
	mob.Rarity = RollRarity(rankMax)
	ApplyRarity(&mob)

	mob.StatsMod = CalculateItemMods(mob.EquipmentMap)
	mob.HitpointsTotal = mob.HitpointsNatural + mob.StatsMod.HitPointMod
	mob.HitpointsRemaining = mob.HitpointsTotal

	return mob
}
