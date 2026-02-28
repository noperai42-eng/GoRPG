package game

import (
	"math"
	"math/rand"
	"sort"

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

		generateFloorGrid(rng, &floor, floorNum)
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

// generateFloorGrid creates a 2D grid map for a dungeon floor and places
// all rooms on it connected by corridors.
func generateFloorGrid(rng *rand.Rand, floor *models.DungeonFloor, floorNum int) {
	// Grid size scales with floor number: 15x15 min, 25x25 max
	gridSize := 15 + floorNum/5
	if gridSize > 25 {
		gridSize = 25
	}
	floor.GridW = gridSize
	floor.GridH = gridSize

	// Initialize grid with walls
	floor.Grid = make([][]models.DungeonTile, gridSize)
	for y := 0; y < gridSize; y++ {
		floor.Grid[y] = make([]models.DungeonTile, gridSize)
		for x := 0; x < gridSize; x++ {
			floor.Grid[y][x] = models.DungeonTile{
				Type:     "wall",
				RoomIdx:  -1,
				Explored: false,
			}
		}
	}

	// Place rooms on the grid using rejection sampling
	for i := range floor.Rooms {
		room := &floor.Rooms[i]
		// Room dimensions: 2x2 to 3x3
		roomW := 2 + rng.Intn(2)
		roomH := 2 + rng.Intn(2)
		room.RoomW = roomW
		room.RoomH = roomH

		placed := false
		for attempt := 0; attempt < 100; attempt++ {
			// Random position with 1-tile border margin
			rx := 1 + rng.Intn(gridSize-roomW-2)
			ry := 1 + rng.Intn(gridSize-roomH-2)

			if canPlaceRoom(floor.Grid, rx, ry, roomW, roomH, gridSize) {
				room.GridX = rx
				room.GridY = ry
				// Stamp room tiles onto grid
				for dy := 0; dy < roomH; dy++ {
					for dx := 0; dx < roomW; dx++ {
						floor.Grid[ry+dy][rx+dx] = models.DungeonTile{
							Type:     "room",
							RoomIdx:  i,
							Explored: false,
						}
					}
				}
				placed = true
				break
			}
		}

		// Fallback: if room couldn't be placed, give it a 1x1 tile in any open spot
		if !placed {
			room.RoomW = 1
			room.RoomH = 1
			for y := 1; y < gridSize-1; y++ {
				for x := 1; x < gridSize-1; x++ {
					if floor.Grid[y][x].Type == "wall" {
						room.GridX = x
						room.GridY = y
						floor.Grid[y][x] = models.DungeonTile{
							Type:     "room",
							RoomIdx:  i,
							Explored: false,
						}
						placed = true
						break
					}
				}
				if placed {
					break
				}
			}
		}
	}

	// Connect rooms with corridors via MST + extra edges
	edges := minimumSpanningTree(floor.Rooms, rng)
	for _, edge := range edges {
		fromRoom := &floor.Rooms[edge[0]]
		toRoom := &floor.Rooms[edge[1]]
		from := models.GridPosition{X: fromRoom.GridX + fromRoom.RoomW/2, Y: fromRoom.GridY + fromRoom.RoomH/2}
		to := models.GridPosition{X: toRoom.GridX + toRoom.RoomW/2, Y: toRoom.GridY + toRoom.RoomH/2}
		generateCorridor(floor.Grid, from, to, rng, gridSize)
	}

	// Mark entrance at room 0
	entranceRoom := &floor.Rooms[0]
	entranceX := entranceRoom.GridX
	entranceY := entranceRoom.GridY
	floor.Grid[entranceY][entranceX].Type = "entrance"
	floor.PlayerPos = models.GridPosition{X: entranceX, Y: entranceY}

	// Mark exit at last room (boss room on boss floors)
	exitRoom := &floor.Rooms[len(floor.Rooms)-1]
	exitX := exitRoom.GridX + exitRoom.RoomW - 1
	exitY := exitRoom.GridY + exitRoom.RoomH - 1
	floor.Grid[exitY][exitX].Type = "exit"
	floor.ExitPos = models.GridPosition{X: exitX, Y: exitY}

	// Reveal tiles within 2-tile radius of start position
	RevealRadius(floor, entranceX, entranceY, 2)
}

// canPlaceRoom checks if a room footprint fits at (rx, ry) without overlapping
// existing rooms (including a 1-tile margin around each room).
func canPlaceRoom(grid [][]models.DungeonTile, rx, ry, roomW, roomH, gridSize int) bool {
	for dy := -1; dy <= roomH; dy++ {
		for dx := -1; dx <= roomW; dx++ {
			cx := rx + dx
			cy := ry + dy
			if cx < 0 || cx >= gridSize || cy < 0 || cy >= gridSize {
				continue
			}
			if grid[cy][cx].Type != "wall" {
				return false
			}
		}
	}
	return true
}

// generateCorridor carves an L-shaped corridor between two points on the grid.
func generateCorridor(grid [][]models.DungeonTile, from, to models.GridPosition, rng *rand.Rand, gridSize int) {
	// Randomly choose horizontal-first or vertical-first
	horizontalFirst := rng.Intn(2) == 0

	if horizontalFirst {
		carveHorizontal(grid, from.X, to.X, from.Y, gridSize)
		carveVertical(grid, from.Y, to.Y, to.X, gridSize)
	} else {
		carveVertical(grid, from.Y, to.Y, from.X, gridSize)
		carveHorizontal(grid, from.X, to.X, to.Y, gridSize)
	}
}

func carveHorizontal(grid [][]models.DungeonTile, x1, x2, y, gridSize int) {
	if y < 0 || y >= gridSize {
		return
	}
	step := 1
	if x1 > x2 {
		step = -1
	}
	for x := x1; x != x2+step; x += step {
		if x < 0 || x >= gridSize {
			break
		}
		if grid[y][x].Type == "wall" {
			grid[y][x] = models.DungeonTile{
				Type:     "corridor",
				RoomIdx:  -1,
				Explored: false,
			}
		}
	}
}

func carveVertical(grid [][]models.DungeonTile, y1, y2, x, gridSize int) {
	if x < 0 || x >= gridSize {
		return
	}
	step := 1
	if y1 > y2 {
		step = -1
	}
	for y := y1; y != y2+step; y += step {
		if y < 0 || y >= gridSize {
			break
		}
		if grid[y][x].Type == "wall" {
			grid[y][x] = models.DungeonTile{
				Type:     "corridor",
				RoomIdx:  -1,
				Explored: false,
			}
		}
	}
}

// mstEdge is an edge for MST computation.
type mstEdge struct {
	from, to int
	dist     float64
}

// minimumSpanningTree computes a MST over room centers using Kruskal's algorithm
// and adds 1-2 random extra edges for loops.
func minimumSpanningTree(rooms []models.DungeonRoom, rng *rand.Rand) [][2]int {
	n := len(rooms)
	if n <= 1 {
		return nil
	}

	// Build all edges with distances
	var edges []mstEdge
	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			cx1 := float64(rooms[i].GridX) + float64(rooms[i].RoomW)/2
			cy1 := float64(rooms[i].GridY) + float64(rooms[i].RoomH)/2
			cx2 := float64(rooms[j].GridX) + float64(rooms[j].RoomW)/2
			cy2 := float64(rooms[j].GridY) + float64(rooms[j].RoomH)/2
			dist := math.Abs(cx1-cx2) + math.Abs(cy1-cy2) // Manhattan distance
			edges = append(edges, mstEdge{from: i, to: j, dist: dist})
		}
	}

	// Sort by distance
	sort.Slice(edges, func(a, b int) bool {
		return edges[a].dist < edges[b].dist
	})

	// Union-Find
	parent := make([]int, n)
	rank := make([]int, n)
	for i := range parent {
		parent[i] = i
	}
	var find func(int) int
	find = func(x int) int {
		if parent[x] != x {
			parent[x] = find(parent[x])
		}
		return parent[x]
	}
	union := func(a, b int) bool {
		ra, rb := find(a), find(b)
		if ra == rb {
			return false
		}
		if rank[ra] < rank[rb] {
			ra, rb = rb, ra
		}
		parent[rb] = ra
		if rank[ra] == rank[rb] {
			rank[ra]++
		}
		return true
	}

	var result [][2]int
	var nonMSTEdges []mstEdge

	for _, e := range edges {
		if union(e.from, e.to) {
			result = append(result, [2]int{e.from, e.to})
		} else {
			nonMSTEdges = append(nonMSTEdges, e)
		}
	}

	// Add 1-2 random extra edges for loops
	extraCount := 1
	if n > 4 {
		extraCount = 2
	}
	for i := 0; i < extraCount && i < len(nonMSTEdges); i++ {
		idx := rng.Intn(len(nonMSTEdges))
		e := nonMSTEdges[idx]
		result = append(result, [2]int{e.from, e.to})
		// Remove used edge
		nonMSTEdges[idx] = nonMSTEdges[len(nonMSTEdges)-1]
		nonMSTEdges = nonMSTEdges[:len(nonMSTEdges)-1]
	}

	return result
}

// RevealRadius marks tiles as explored within a given radius of (cx, cy).
func RevealRadius(floor *models.DungeonFloor, cx, cy, radius int) {
	for dy := -radius; dy <= radius; dy++ {
		for dx := -radius; dx <= radius; dx++ {
			nx := cx + dx
			ny := cy + dy
			if nx >= 0 && nx < floor.GridW && ny >= 0 && ny < floor.GridH {
				floor.Grid[ny][nx].Explored = true
			}
		}
	}
}

// CanMoveOnGrid checks if a position is walkable on the dungeon grid.
func CanMoveOnGrid(floor *models.DungeonFloor, x, y int) bool {
	if x < 0 || x >= floor.GridW || y < 0 || y >= floor.GridH {
		return false
	}
	tile := floor.Grid[y][x]
	return tile.Type != "wall"
}

// RegenerateFloorGrid regenerates the grid for a floor that was loaded without one.
// This handles backward compatibility with old save data.
func RegenerateFloorGrid(floor *models.DungeonFloor, seed int64, floorIdx int) {
	src := rand.NewSource(seed + int64(floorIdx)*1000)
	rng := rand.New(src)
	generateFloorGrid(rng, floor, floor.FloorNumber)
}
