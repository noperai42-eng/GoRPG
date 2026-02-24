package game

import (
	"fmt"
	"math/rand"

	"rpg-game/pkg/data"
	"rpg-game/pkg/models"
)

func GenerateMonster(name string, level int, rank int) models.Monster {
	hitpoints := MultiRoll(rank)
	mana := MultiRoll(rank) + 10
	stamina := MultiRoll(rank) + 10

	// Set resistances based on monster type
	resistances := map[models.DamageType]float64{
		models.Physical:  1.0,
		models.Fire:      1.0,
		models.Ice:       1.0,
		models.Lightning: 1.0,
		models.Poison:    1.0,
	}

	// Customize based on monster type
	switch name {
	case "slime":
		resistances[models.Physical] = 0.5 // resistant to physical
		resistances[models.Fire] = 2.0     // weak to fire
	case "golem":
		resistances[models.Physical] = 0.25  // very resistant to physical
		resistances[models.Lightning] = 2.0  // weak to lightning
	case "orc":
		resistances[models.Fire] = 0.8 // slight fire resistance
	case "hiftier":
		resistances[models.Lightning] = 0.5 // magic resistant
		resistances[models.Ice] = 0.5
		resistances[models.Fire] = 0.5
	}

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
	// Use special guardian name
	guardianName := data.SkillGuardianNames[rand.Intn(len(data.SkillGuardianNames))]

	// Create base monster with enhanced stats
	baseMob := GenerateMonster(guardianName, level, rank)

	// Make guardian significantly tougher
	baseMob.HitpointsNatural = int(float64(baseMob.HitpointsNatural) * 2.0)
	baseMob.HitpointsTotal = baseMob.HitpointsNatural
	baseMob.HitpointsRemaining = baseMob.HitpointsTotal
	baseMob.AttackRolls = baseMob.AttackRolls + 2
	baseMob.DefenseRolls = baseMob.DefenseRolls + 2
	baseMob.ManaTotal = int(float64(baseMob.ManaTotal) * 1.5)
	baseMob.ManaRemaining = baseMob.ManaTotal
	baseMob.StaminaTotal = int(float64(baseMob.StaminaTotal) * 1.5)
	baseMob.StaminaRemaining = baseMob.StaminaTotal

	// Give guardian better equipment
	for i := 0; i < rank+2; i++ {
		item := GenerateItem(rank + 1)
		EquipBestItem(item, &baseMob.EquipmentMap, &baseMob.Inventory)
	}

	// Mark as skill guardian
	baseMob.IsSkillGuardian = true
	baseMob.GuardedSkill = skill
	baseMob.MonsterType = "Guardian"

	// Recalculate stats with equipment
	baseMob.StatsMod = CalculateItemMods(baseMob.EquipmentMap)
	baseMob.HitpointsTotal = baseMob.HitpointsNatural + baseMob.StatsMod.HitPointMod
	baseMob.HitpointsRemaining = baseMob.HitpointsTotal

	return baseMob
}

// AssignMonsterSkills assigns skills to monsters based on their type
func AssignMonsterSkills(monsterType string, level int) []models.Skill {
	skills := []models.Skill{}

	switch monsterType {
	case "slime":
		// Slimes are weak but can poison
		if level >= 3 {
			skills = append(skills, models.Skill{
				Name:        "Acid Spit",
				ManaCost:    5,
				Damage:      8,
				DamageType:  models.Poison,
				Effect:      models.StatusEffect{Type: "poison", Duration: 3, Potency: 3},
				Description: "Spit acid that poisons",
			})
		}
	case "goblin":
		// Goblins use stamina-based attacks
		if level >= 2 {
			skills = append(skills, models.Skill{
				Name:        "Backstab",
				StaminaCost: 10,
				Damage:      15,
				DamageType:  models.Physical,
				Effect:      models.StatusEffect{Type: "none"},
				Description: "Sneaky physical attack",
			})
		}
	case "orc":
		// Orcs use powerful physical attacks and buffs
		if level >= 3 {
			skills = append(skills, models.Skill{
				Name:        "War Cry",
				StaminaCost: 15,
				Damage:      0,
				DamageType:  models.Physical,
				Effect:      models.StatusEffect{Type: "buff_attack", Duration: 3, Potency: 8},
				Description: "Boost attack power",
			})
		}
		if level >= 5 {
			skills = append(skills, models.Skill{
				Name:        "Berserker Rage",
				StaminaCost: 20,
				Damage:      25,
				DamageType:  models.Physical,
				Effect:      models.StatusEffect{Type: "none"},
				Description: "Powerful rage attack",
			})
		}
	case "golem":
		// Golems are tanky with defensive skills
		if level >= 4 {
			skills = append(skills, models.Skill{
				Name:        "Stone Skin",
				ManaCost:    10,
				Damage:      0,
				DamageType:  models.Physical,
				Effect:      models.StatusEffect{Type: "buff_defense", Duration: 4, Potency: 15},
				Description: "Harden skin for defense",
			})
		}
	case "hiftier":
		// Hiftiers are magic users
		if level >= 3 {
			skills = append(skills, models.Skill{
				Name:        "Mana Bolt",
				ManaCost:    12,
				Damage:      18,
				DamageType:  models.Lightning,
				Effect:      models.StatusEffect{Type: "none"},
				Description: "Magical lightning bolt",
			})
		}
		if level >= 6 {
			skills = append(skills, models.Skill{
				Name:        "Mind Blast",
				ManaCost:    18,
				Damage:      15,
				DamageType:  models.Lightning,
				Effect:      models.StatusEffect{Type: "stun", Duration: 1, Potency: 1},
				Description: "Stun the enemy",
			})
		}
	case "kobold":
		// Kobolds use fire
		if level >= 2 {
			skills = append(skills, models.Skill{
				Name:        "Fire Breath",
				ManaCost:    8,
				Damage:      12,
				DamageType:  models.Fire,
				Effect:      models.StatusEffect{Type: "burn", Duration: 2, Potency: 2},
				Description: "Breathe fire",
			})
		}
	case "kitpod":
		// Kitpods regenerate
		if level >= 3 {
			skills = append(skills, models.Skill{
				Name:        "Regenerate",
				ManaCost:    10,
				Damage:      0,
				DamageType:  models.Physical,
				Effect:      models.StatusEffect{Type: "regen", Duration: 4, Potency: 5},
				Description: "Heal over time",
			})
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
