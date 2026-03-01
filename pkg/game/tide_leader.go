package game

import (
	"fmt"
	"math/rand"

	"rpg-game/pkg/models"
)

// Tide leader names — one is picked randomly on spawn.
var tideLeaderNames = []string{
	"Abyssal Warden",
	"The Hollow King",
	"Tidecaller Gor'thax",
	"Moonbane Sentinel",
	"Riftclaw Behemoth",
	"The Drowned Colossus",
	"Stormfang Elder",
	"Cragtide Leviathan",
}

// GenerateTideLeader creates a new tide leader for the given cycle.
// Stats scale with the undefeated streak: base level 10 + 5 per streak,
// 25% more HP per streak.
func GenerateTideLeader(cycleYear, cycleSeason, timesUndefeated int) models.TideLeader {
	name := tideLeaderNames[rand.Intn(len(tideLeaderNames))]
	level := 10 + 5*timesUndefeated
	baseHP := 500 + level*50
	// 25% more HP per undefeated streak
	hp := baseHP + (baseHP * 25 * timesUndefeated / 100)
	attackRolls := 3 + timesUndefeated
	defenseRolls := 2 + timesUndefeated

	return models.TideLeader{
		Name:               name,
		Level:              level,
		HitpointsTotal:     hp,
		HitpointsRemaining: hp,
		AttackRolls:        attackRolls,
		DefenseRolls:       defenseRolls,
		CycleYear:          cycleYear,
		CycleSeason:        cycleSeason,
		TimesUndefeated:    timesUndefeated,
		Defeated:           false,
		RaidParticipants:   []string{},
		LastRaidDay:        0,
	}
}

// RaidResult holds the outcome of a single village's raid contribution.
type RaidResult struct {
	VillageName    string
	DamageDealt    int
	DamageTaken    int
	Messages       []string
	LeaderDefeated bool
}

// ProcessTideLeaderRaid calculates a village's attack against the tide leader.
// The village contributes damage based on guards + defenses + player level.
// The leader retaliates against village defenses.
func ProcessTideLeaderRaid(leader *models.TideLeader, village *models.Village, playerLevel int) RaidResult {
	result := RaidResult{VillageName: village.Name}

	// Calculate village attack power
	guardPower := 0
	for _, g := range village.ActiveGuards {
		if !g.Injured {
			guardPower += g.Level*2 + g.AttackBonus
		}
	}
	defensePower := 0
	for _, d := range village.Defenses {
		if d.Built {
			defensePower += d.AttackPower
		}
	}
	villageDamage := guardPower + defensePower + playerLevel*3

	// Minimum damage of 5
	if villageDamage < 5 {
		villageDamage = 5
	}

	// Add some randomness (80-120%)
	villageDamage = villageDamage * (80 + rand.Intn(41)) / 100

	leader.HitpointsRemaining -= villageDamage
	result.DamageDealt = villageDamage

	result.Messages = append(result.Messages,
		fmt.Sprintf("%s contributes %d damage to %s!", village.Name, villageDamage, leader.Name))

	// Leader retaliates — damages village defenses
	leaderDamage := leader.Level*2 + leader.AttackRolls*3
	leaderDamage = leaderDamage * (80 + rand.Intn(41)) / 100

	// Defenses absorb damage first
	totalDefense := 0
	for _, d := range village.Defenses {
		if d.Built {
			totalDefense += d.Defense
		}
	}
	actualDamage := leaderDamage - totalDefense/2
	if actualDamage < 0 {
		actualDamage = 0
	}
	result.DamageTaken = actualDamage

	// Chance to injure a random guard
	if actualDamage > 0 && len(village.ActiveGuards) > 0 && rand.Intn(100) < 30 {
		idx := rand.Intn(len(village.ActiveGuards))
		if !village.ActiveGuards[idx].Injured {
			village.ActiveGuards[idx].Injured = true
			village.ActiveGuards[idx].RecoveryTime = 3600
			result.Messages = append(result.Messages,
				fmt.Sprintf("  %s was injured by %s's retaliation!", village.ActiveGuards[idx].Name, leader.Name))
		}
	}

	// Track participant
	if !containsString(leader.RaidParticipants, village.Name) {
		leader.RaidParticipants = append(leader.RaidParticipants, village.Name)
	}

	leader.TotalDamageDealt += villageDamage

	if leader.HitpointsRemaining <= 0 {
		leader.HitpointsRemaining = 0
		leader.Defeated = true
		result.LeaderDefeated = true
		result.Messages = append(result.Messages,
			fmt.Sprintf("The %s has been DEFEATED! All villages celebrate!", leader.Name))
	}

	return result
}

// TideLeaderDefeatReward grants rewards to a player whose village participated.
func TideLeaderDefeatReward(player *models.Character, leader *models.TideLeader) (int, int) {
	xp := leader.Level*50 + leader.TimesUndefeated*100
	gold := leader.Level*10 + leader.TimesUndefeated*20

	player.Experience += xp
	if player.ResourceStorageMap == nil {
		player.ResourceStorageMap = make(map[string]models.Resource)
	}
	goldRes := player.ResourceStorageMap["Gold"]
	goldRes.Name = "Gold"
	goldRes.Stock += gold
	player.ResourceStorageMap["Gold"] = goldRes

	return xp, gold
}

// ScaleTidesForUndefeated reduces the tide interval by 10% per undefeated
// cycle (minimum 1800 seconds = 30 minutes).
func ScaleTidesForUndefeated(village *models.Village, timesUndefeated int) {
	if village.TideInterval <= 0 {
		village.TideInterval = 3600
	}
	reduction := village.TideInterval * 10 * timesUndefeated / 100
	village.TideInterval -= reduction
	if village.TideInterval < 1800 {
		village.TideInterval = 1800
	}
}

func containsString(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}
