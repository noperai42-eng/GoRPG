package metrics

import (
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// MetricsCollector accumulates server-wide game metrics using atomic counters
// for simple increments and a single RWMutex for distribution maps.
type MetricsCollector struct {
	startTime time.Time

	// Atomic counters (zero-lock hot path)
	TotalFights       atomic.Int64
	PlayerWins        atomic.Int64
	PlayerDeaths      atomic.Int64
	Flees             atomic.Int64
	FleeFails         atomic.Int64
	PlayerCrits       atomic.Int64
	MonsterCrits      atomic.Int64
	CombatTurns       atomic.Int64
	PlayerDamageDealt atomic.Int64
	MonsterDamageDealt atomic.Int64
	SkillUses         atomic.Int64
	ItemUses          atomic.Int64
	DefendActions     atomic.Int64
	AutoFights        atomic.Int64
	LevelUps          atomic.Int64
	XPAwarded         atomic.Int64
	Harvests          atomic.Int64
	ResourceUnits     atomic.Int64
	ItemsLooted       atomic.Int64
	ArenaFights       atomic.Int64
	DungeonEnters     atomic.Int64
	DungeonClears     atomic.Int64
	DungeonDeaths     atomic.Int64
	TideTicks           atomic.Int64
	TidesProcessed      atomic.Int64
	TideVictories       atomic.Int64
	TideDefeats         atomic.Int64
	VillageManagerTicks atomic.Int64
	VillagesManaged     atomic.Int64
	QuestCompletions    atomic.Int64
	GuardianSpawns      atomic.Int64
	GuardianDefeats     atomic.Int64
	SkillsLearned       atomic.Int64
	SkillsUpgraded      atomic.Int64
	OnlinePlayers     atomic.Int64

	// Distribution maps (single mutex)
	mu                sync.RWMutex
	WinsByLocation    map[string]int64
	LossesByLocation  map[string]int64
	WinsByMonsterType map[string]int64
	LossesByMonsterType map[string]int64
	WinsByRarity      map[string]int64
	LossesByRarity    map[string]int64
	SkillUseCounts    map[string]int64
	StatusEffects     map[string]int64
	DamageByType      map[string]int64
	LevelUpsByLevel   map[string]int64
	HarvestsByResource map[string]int64
	PotionsUsed       map[string]int64
	ItemsByRarity     map[string]int64
	ArenaWinsByGap    map[string]int64
	ArenaTotalByGap   map[string]int64
	FloorDeaths       map[string]int64
	FloorClears       map[string]int64
	FeatureUsage           map[string]int64
	QuestCompletionsByID   map[string]int64
	GuardianSpawnsBySkill  map[string]int64
	GuardianDefeatsBySkill map[string]int64
	SkillsLearnedByName    map[string]int64
	SkillsUpgradedByName   map[string]int64
}

// NewMetricsCollector creates a new MetricsCollector with initialized maps.
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		startTime:           time.Now(),
		WinsByLocation:      make(map[string]int64),
		LossesByLocation:    make(map[string]int64),
		WinsByMonsterType:   make(map[string]int64),
		LossesByMonsterType: make(map[string]int64),
		WinsByRarity:        make(map[string]int64),
		LossesByRarity:      make(map[string]int64),
		SkillUseCounts:      make(map[string]int64),
		StatusEffects:       make(map[string]int64),
		DamageByType:        make(map[string]int64),
		LevelUpsByLevel:     make(map[string]int64),
		HarvestsByResource:  make(map[string]int64),
		PotionsUsed:         make(map[string]int64),
		ItemsByRarity:       make(map[string]int64),
		ArenaWinsByGap:      make(map[string]int64),
		ArenaTotalByGap:     make(map[string]int64),
		FloorDeaths:         make(map[string]int64),
		FloorClears:         make(map[string]int64),
		FeatureUsage:           make(map[string]int64),
		QuestCompletionsByID:   make(map[string]int64),
		GuardianSpawnsBySkill:  make(map[string]int64),
		GuardianDefeatsBySkill: make(map[string]int64),
		SkillsLearnedByName:    make(map[string]int64),
		SkillsUpgradedByName:   make(map[string]int64),
	}
}

// ---------------------------------------------------------------------------
// Recording methods
// ---------------------------------------------------------------------------

// RecordCombatWin records a player victory.
func (mc *MetricsCollector) RecordCombatWin(location, monsterType, rarity string, turns int) {
	mc.TotalFights.Add(1)
	mc.PlayerWins.Add(1)
	mc.CombatTurns.Add(int64(turns))

	mc.mu.Lock()
	if location != "" {
		mc.WinsByLocation[location]++
	}
	if monsterType != "" {
		mc.WinsByMonsterType[monsterType]++
	}
	if rarity != "" {
		mc.WinsByRarity[rarity]++
	}
	mc.mu.Unlock()
}

// RecordCombatLoss records a player defeat.
func (mc *MetricsCollector) RecordCombatLoss(location, monsterType, rarity string, turns int) {
	mc.TotalFights.Add(1)
	mc.PlayerDeaths.Add(1)
	mc.CombatTurns.Add(int64(turns))

	mc.mu.Lock()
	if location != "" {
		mc.LossesByLocation[location]++
	}
	if monsterType != "" {
		mc.LossesByMonsterType[monsterType]++
	}
	if rarity != "" {
		mc.LossesByRarity[rarity]++
	}
	mc.mu.Unlock()
}

// RecordFlee records a flee attempt.
func (mc *MetricsCollector) RecordFlee(success bool) {
	if success {
		mc.Flees.Add(1)
	} else {
		mc.FleeFails.Add(1)
	}
}

// RecordCrit records a critical hit.
func (mc *MetricsCollector) RecordCrit(isPlayer bool) {
	if isPlayer {
		mc.PlayerCrits.Add(1)
	} else {
		mc.MonsterCrits.Add(1)
	}
}

// RecordDamage records damage dealt.
func (mc *MetricsCollector) RecordDamage(amount int, damageType string, isPlayer bool) {
	if isPlayer {
		mc.PlayerDamageDealt.Add(int64(amount))
	} else {
		mc.MonsterDamageDealt.Add(int64(amount))
	}
	mc.mu.Lock()
	mc.DamageByType[damageType] += int64(amount)
	mc.mu.Unlock()
}

// RecordSkillUse records a skill usage.
func (mc *MetricsCollector) RecordSkillUse(name string) {
	mc.SkillUses.Add(1)
	mc.mu.Lock()
	mc.SkillUseCounts[name]++
	mc.mu.Unlock()
}

// RecordStatusEffect records a status effect application.
func (mc *MetricsCollector) RecordStatusEffect(effectType string) {
	mc.mu.Lock()
	mc.StatusEffects[effectType]++
	mc.mu.Unlock()
}

// RecordItemUse records item consumption.
func (mc *MetricsCollector) RecordItemUse(name string) {
	mc.ItemUses.Add(1)
	mc.mu.Lock()
	mc.PotionsUsed[name]++
	mc.mu.Unlock()
}

// RecordDefend records a defend action.
func (mc *MetricsCollector) RecordDefend() {
	mc.DefendActions.Add(1)
}

// RecordAutoFight records an auto-fight initiation.
func (mc *MetricsCollector) RecordAutoFight() {
	mc.AutoFights.Add(1)
}

// RecordLevelUp records a level-up event.
func (mc *MetricsCollector) RecordLevelUp(level int) {
	mc.LevelUps.Add(1)
	mc.mu.Lock()
	mc.LevelUpsByLevel[fmt.Sprintf("%d", level)]++
	mc.mu.Unlock()
}

// RecordXP records XP awarded.
func (mc *MetricsCollector) RecordXP(amount int) {
	mc.XPAwarded.Add(int64(amount))
}

// RecordHarvest records resource harvesting.
func (mc *MetricsCollector) RecordHarvest(resource string, amount int) {
	mc.Harvests.Add(1)
	mc.ResourceUnits.Add(int64(amount))
	mc.mu.Lock()
	mc.HarvestsByResource[resource] += int64(amount)
	mc.mu.Unlock()
}

// RecordItemLooted records an item being looted, keyed by rarity.
func (mc *MetricsCollector) RecordItemLooted(rarity int) {
	mc.ItemsLooted.Add(1)
	mc.mu.Lock()
	mc.ItemsByRarity[fmt.Sprintf("%d", rarity)]++
	mc.mu.Unlock()
}

// RecordArenaFight records an arena fight outcome.
func (mc *MetricsCollector) RecordArenaFight(winnerRating, loserRating int) {
	mc.ArenaFights.Add(1)
	gap := winnerRating - loserRating
	if gap < 0 {
		gap = -gap
	}
	bucket := "0-100"
	if gap >= 200 {
		bucket = "200+"
	} else if gap >= 100 {
		bucket = "100-200"
	}
	mc.mu.Lock()
	mc.ArenaWinsByGap[bucket]++
	mc.ArenaTotalByGap[bucket]++
	mc.mu.Unlock()
}

// RecordDungeonEnter records a dungeon entry.
func (mc *MetricsCollector) RecordDungeonEnter() {
	mc.DungeonEnters.Add(1)
}

// RecordDungeonClear records a dungeon completion.
func (mc *MetricsCollector) RecordDungeonClear() {
	mc.DungeonClears.Add(1)
}

// RecordDungeonDeath records a death in a dungeon on a specific floor.
func (mc *MetricsCollector) RecordDungeonDeath(floor int) {
	mc.DungeonDeaths.Add(1)
	mc.mu.Lock()
	mc.FloorDeaths[fmt.Sprintf("%d", floor)]++
	mc.mu.Unlock()
}

// RecordFloorClear records clearing a dungeon floor.
func (mc *MetricsCollector) RecordFloorClear(floor int) {
	mc.mu.Lock()
	mc.FloorClears[fmt.Sprintf("%d", floor)]++
	mc.mu.Unlock()
}

// RecordFeatureUse records usage of a game feature.
func (mc *MetricsCollector) RecordFeatureUse(feature string) {
	mc.mu.Lock()
	mc.FeatureUsage[feature]++
	mc.mu.Unlock()
}

// RecordTideTick records a tide processing tick.
func (mc *MetricsCollector) RecordTideTick(tidesProcessed int) {
	mc.TideTicks.Add(1)
	mc.TidesProcessed.Add(int64(tidesProcessed))
}

// RecordTideOutcome records a tide victory or defeat.
func (mc *MetricsCollector) RecordTideOutcome(victory bool) {
	if victory {
		mc.TideVictories.Add(1)
	} else {
		mc.TideDefeats.Add(1)
	}
}

// RecordQuestComplete records a quest completion.
func (mc *MetricsCollector) RecordQuestComplete(questID string) {
	mc.QuestCompletions.Add(1)
	mc.mu.Lock()
	mc.QuestCompletionsByID[questID]++
	mc.mu.Unlock()
}

// RecordGuardianSpawn records a guardian spawn event.
func (mc *MetricsCollector) RecordGuardianSpawn(skillName string) {
	mc.GuardianSpawns.Add(1)
	mc.mu.Lock()
	mc.GuardianSpawnsBySkill[skillName]++
	mc.mu.Unlock()
}

// RecordGuardianDefeat records a guardian defeat event.
func (mc *MetricsCollector) RecordGuardianDefeat(skillName string) {
	mc.GuardianDefeats.Add(1)
	mc.mu.Lock()
	mc.GuardianDefeatsBySkill[skillName]++
	mc.mu.Unlock()
}

// RecordSkillLearned records a skill being learned from a guardian.
func (mc *MetricsCollector) RecordSkillLearned(skillName string) {
	mc.SkillsLearned.Add(1)
	mc.mu.Lock()
	mc.SkillsLearnedByName[skillName]++
	mc.mu.Unlock()
}

// RecordSkillUpgraded records a skill being upgraded from a duplicate guardian.
func (mc *MetricsCollector) RecordSkillUpgraded(skillName string) {
	mc.SkillsUpgraded.Add(1)
	mc.mu.Lock()
	mc.SkillsUpgradedByName[skillName]++
	mc.mu.Unlock()
}

// RecordVillageManagerTick records a village manager processing tick.
func (mc *MetricsCollector) RecordVillageManagerTick(villagesManaged int) {
	mc.VillageManagerTicks.Add(1)
	mc.VillagesManaged.Add(int64(villagesManaged))
}

// ---------------------------------------------------------------------------
// Snapshot
// ---------------------------------------------------------------------------

// MetricsSnapshot is a JSON-serializable copy of all metrics.
type MetricsSnapshot struct {
	UptimeSeconds int64                  `json:"uptime_seconds"`
	OnlinePlayers int64                  `json:"online_players"`
	Combat        CombatMetrics          `json:"combat"`
	Progression   ProgressionMetrics     `json:"progression"`
	Economy       EconomyMetrics         `json:"economy"`
	Arena         ArenaMetrics           `json:"arena"`
	Dungeons      DungeonMetrics         `json:"dungeons"`
	Engagement    EngagementMetrics      `json:"engagement"`
	Village       VillageMetrics         `json:"village"`
	Quests        QuestMetrics           `json:"quests"`
	Guardians     GuardianMetrics        `json:"guardians"`
}

// GuardianMetrics holds guardian spawn and skill acquisition data.
type GuardianMetrics struct {
	TotalSpawns        int64            `json:"total_spawns"`
	TotalDefeats       int64            `json:"total_defeats"`
	TotalSkillsLearned int64            `json:"total_skills_learned"`
	TotalSkillsUpgraded int64           `json:"total_skills_upgraded"`
	SpawnsBySkill      map[string]int64 `json:"spawns_by_skill"`
	DefeatsBySkill     map[string]int64 `json:"defeats_by_skill"`
	LearnedByName      map[string]int64 `json:"learned_by_name"`
	UpgradedByName     map[string]int64 `json:"upgraded_by_name"`
}

// CombatMetrics holds combat-related aggregates.
type CombatMetrics struct {
	TotalFights      int64              `json:"total_fights"`
	Wins             int64              `json:"wins"`
	Losses           int64              `json:"losses"`
	Flees            int64              `json:"flees"`
	FleeFails        int64              `json:"flee_fails"`
	WinRate          float64            `json:"win_rate"`
	FleeSuccessRate  float64            `json:"flee_success_rate"`
	AvgTurns         float64            `json:"avg_turns"`
	PlayerCritRate   float64            `json:"player_crit_rate"`
	MonsterCritRate  float64            `json:"monster_crit_rate"`
	AutoFightRate    float64            `json:"auto_fight_rate"`
	PlayerDamage     int64              `json:"player_damage_dealt"`
	MonsterDamage    int64              `json:"monster_damage_dealt"`
	DefendActions    int64              `json:"defend_actions"`
	ByLocation       map[string]WinLoss `json:"by_location"`
	ByMonsterType    map[string]WinLoss `json:"by_monster_type"`
	ByRarity         map[string]WinLoss `json:"by_rarity"`
	SkillUsage       map[string]int64   `json:"skill_usage"`
	StatusEffects    map[string]int64   `json:"status_effects"`
	DamageByType     map[string]int64   `json:"damage_by_type"`
}

// WinLoss holds win/loss counts for a category.
type WinLoss struct {
	W int64 `json:"w"`
	L int64 `json:"l"`
}

// ProgressionMetrics holds level-up and XP data.
type ProgressionMetrics struct {
	TotalLevelUps     int64            `json:"total_level_ups"`
	TotalXP           int64            `json:"total_xp"`
	LevelDistribution map[string]int64 `json:"level_distribution"`
}

// EconomyMetrics holds resource/item economy data.
type EconomyMetrics struct {
	HarvestsByResource map[string]int64 `json:"harvests_by_resource"`
	PotionsUsed        map[string]int64 `json:"potions_used"`
	ItemsByRarity      map[string]int64 `json:"items_by_rarity"`
	TotalItemsLooted   int64            `json:"total_items_looted"`
}

// ArenaMetrics holds PvP arena data.
type ArenaMetrics struct {
	TotalFights  int64                     `json:"total_fights"`
	ByRatingGap  map[string]ArenaGapStats  `json:"by_rating_gap"`
}

// ArenaGapStats holds arena stats for a rating gap bucket.
type ArenaGapStats struct {
	HigherWins int64 `json:"higher_wins"`
	Total      int64 `json:"total"`
}

// DungeonMetrics holds dungeon exploration data.
type DungeonMetrics struct {
	Enters        int64            `json:"enters"`
	Clears        int64            `json:"clears"`
	Deaths        int64            `json:"deaths"`
	ClearRate     float64          `json:"clear_rate"`
	DeathsByFloor map[string]int64 `json:"deaths_by_floor"`
	ClearsByFloor map[string]int64 `json:"clears_by_floor"`
}

// EngagementMetrics holds feature usage data.
type EngagementMetrics struct {
	FeatureUsage map[string]int64 `json:"feature_usage"`
}

// QuestMetrics holds quest completion data.
type QuestMetrics struct {
	TotalCompletions int64            `json:"total_completions"`
	CompletionsByID  map[string]int64 `json:"completions_by_id"`
}

// VillageMetrics holds village and tide-related aggregates.
type VillageMetrics struct {
	TideTicks           int64   `json:"tide_ticks"`
	TidesProcessed      int64   `json:"tides_processed"`
	TideVictories       int64   `json:"tide_victories"`
	TideDefeats         int64   `json:"tide_defeats"`
	TideWinRate         float64 `json:"tide_win_rate"`
	AvgTidesPerTick     float64 `json:"avg_tides_per_tick"`
	VillageManagerTicks int64   `json:"village_manager_ticks"`
	VillagesManaged     int64   `json:"villages_managed"`
}

// Snapshot returns a JSON-serializable copy of all current metrics with computed rates.
func (mc *MetricsCollector) Snapshot() MetricsSnapshot {
	totalFights := mc.TotalFights.Load()
	wins := mc.PlayerWins.Load()
	losses := mc.PlayerDeaths.Load()
	flees := mc.Flees.Load()
	fleeFails := mc.FleeFails.Load()
	playerCrits := mc.PlayerCrits.Load()
	monsterCrits := mc.MonsterCrits.Load()
	combatTurns := mc.CombatTurns.Load()
	autoFights := mc.AutoFights.Load()
	dungeonEnters := mc.DungeonEnters.Load()
	dungeonClears := mc.DungeonClears.Load()
	dungeonDeaths := mc.DungeonDeaths.Load()

	// Compute rates
	winRate := safeDiv(float64(wins), float64(wins+losses))
	fleeTotal := flees + fleeFails
	fleeSuccessRate := safeDiv(float64(flees), float64(fleeTotal))
	avgTurns := safeDiv(float64(combatTurns), float64(totalFights))

	// Total player attacks = totalFights * avgTurns is imprecise; use wins+losses as base for crit rate
	totalPlayerActions := wins + losses + flees + fleeFails
	playerCritRate := safeDiv(float64(playerCrits), float64(totalPlayerActions))
	monsterCritRate := safeDiv(float64(monsterCrits), float64(totalPlayerActions))
	autoFightRate := safeDiv(float64(autoFights), float64(totalFights))
	dungeonClearRate := safeDiv(float64(dungeonClears), float64(dungeonEnters))

	// Copy maps under lock
	mc.mu.RLock()
	byLocation := mergeWinLoss(mc.WinsByLocation, mc.LossesByLocation)
	byMonsterType := mergeWinLoss(mc.WinsByMonsterType, mc.LossesByMonsterType)
	byRarity := mergeWinLoss(mc.WinsByRarity, mc.LossesByRarity)
	skillUsage := copyMap(mc.SkillUseCounts)
	statusEffects := copyMap(mc.StatusEffects)
	damageByType := copyMap(mc.DamageByType)
	levelDist := copyMap(mc.LevelUpsByLevel)
	harvestsByRes := copyMap(mc.HarvestsByResource)
	potionsUsed := copyMap(mc.PotionsUsed)
	itemsByRarity := copyMap(mc.ItemsByRarity)
	floorDeaths := copyMap(mc.FloorDeaths)
	floorClears := copyMap(mc.FloorClears)
	featureUsage := copyMap(mc.FeatureUsage)
	questCompletions := copyMap(mc.QuestCompletionsByID)
	guardianSpawnsBySkill := copyMap(mc.GuardianSpawnsBySkill)
	guardianDefeatsBySkill := copyMap(mc.GuardianDefeatsBySkill)
	skillsLearnedByName := copyMap(mc.SkillsLearnedByName)
	skillsUpgradedByName := copyMap(mc.SkillsUpgradedByName)

	arenaByGap := make(map[string]ArenaGapStats)
	for bucket, wins := range mc.ArenaWinsByGap {
		arenaByGap[bucket] = ArenaGapStats{
			HigherWins: wins,
			Total:      mc.ArenaTotalByGap[bucket],
		}
	}
	mc.mu.RUnlock()

	return MetricsSnapshot{
		UptimeSeconds: int64(time.Since(mc.startTime).Seconds()),
		OnlinePlayers: mc.OnlinePlayers.Load(),
		Combat: CombatMetrics{
			TotalFights:     totalFights,
			Wins:            wins,
			Losses:          losses,
			Flees:           flees,
			FleeFails:       fleeFails,
			WinRate:         winRate,
			FleeSuccessRate: fleeSuccessRate,
			AvgTurns:        avgTurns,
			PlayerCritRate:  playerCritRate,
			MonsterCritRate: monsterCritRate,
			AutoFightRate:   autoFightRate,
			PlayerDamage:    mc.PlayerDamageDealt.Load(),
			MonsterDamage:   mc.MonsterDamageDealt.Load(),
			DefendActions:   mc.DefendActions.Load(),
			ByLocation:      byLocation,
			ByMonsterType:   byMonsterType,
			ByRarity:        byRarity,
			SkillUsage:      skillUsage,
			StatusEffects:   statusEffects,
			DamageByType:    damageByType,
		},
		Progression: ProgressionMetrics{
			TotalLevelUps:     mc.LevelUps.Load(),
			TotalXP:           mc.XPAwarded.Load(),
			LevelDistribution: levelDist,
		},
		Economy: EconomyMetrics{
			HarvestsByResource: harvestsByRes,
			PotionsUsed:        potionsUsed,
			ItemsByRarity:      itemsByRarity,
			TotalItemsLooted:   mc.ItemsLooted.Load(),
		},
		Arena: ArenaMetrics{
			TotalFights: mc.ArenaFights.Load(),
			ByRatingGap: arenaByGap,
		},
		Dungeons: DungeonMetrics{
			Enters:        dungeonEnters,
			Clears:        dungeonClears,
			Deaths:        dungeonDeaths,
			ClearRate:     dungeonClearRate,
			DeathsByFloor: floorDeaths,
			ClearsByFloor: floorClears,
		},
		Engagement: EngagementMetrics{
			FeatureUsage: featureUsage,
		},
		Quests: QuestMetrics{
			TotalCompletions: mc.QuestCompletions.Load(),
			CompletionsByID:  questCompletions,
		},
		Village: func() VillageMetrics {
			tideTicks := mc.TideTicks.Load()
			tidesProcessed := mc.TidesProcessed.Load()
			tideVictories := mc.TideVictories.Load()
			tideDefeats := mc.TideDefeats.Load()
			tideTotal := tideVictories + tideDefeats
			return VillageMetrics{
				TideTicks:           tideTicks,
				TidesProcessed:      tidesProcessed,
				TideVictories:       tideVictories,
				TideDefeats:         tideDefeats,
				TideWinRate:         safeDiv(float64(tideVictories), float64(tideTotal)),
				AvgTidesPerTick:     safeDiv(float64(tidesProcessed), float64(tideTicks)),
				VillageManagerTicks: mc.VillageManagerTicks.Load(),
				VillagesManaged:     mc.VillagesManaged.Load(),
			}
		}(),
		Guardians: GuardianMetrics{
			TotalSpawns:        mc.GuardianSpawns.Load(),
			TotalDefeats:       mc.GuardianDefeats.Load(),
			TotalSkillsLearned: mc.SkillsLearned.Load(),
			TotalSkillsUpgraded: mc.SkillsUpgraded.Load(),
			SpawnsBySkill:      guardianSpawnsBySkill,
			DefeatsBySkill:     guardianDefeatsBySkill,
			LearnedByName:      skillsLearnedByName,
			UpgradedByName:     skillsUpgradedByName,
		},
	}
}

// SnapshotJSON returns the snapshot as a JSON string.
func (mc *MetricsCollector) SnapshotJSON() (string, error) {
	snap := mc.Snapshot()
	data, err := json.Marshal(snap)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func safeDiv(num, denom float64) float64 {
	if denom == 0 {
		return 0
	}
	return num / denom
}

func copyMap(src map[string]int64) map[string]int64 {
	dst := make(map[string]int64, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func mergeWinLoss(wins, losses map[string]int64) map[string]WinLoss {
	keys := make(map[string]struct{})
	for k := range wins {
		keys[k] = struct{}{}
	}
	for k := range losses {
		keys[k] = struct{}{}
	}
	result := make(map[string]WinLoss, len(keys))
	for k := range keys {
		result[k] = WinLoss{W: wins[k], L: losses[k]}
	}
	return result
}
