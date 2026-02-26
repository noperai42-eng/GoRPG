package game

import (
	"math/rand"
	"time"

	"rpg-game/pkg/data"
	"rpg-game/pkg/models"
)

const DefaultTownName = "Crossroads"

// GenerateDefaultTown creates a new town with an NPC mayor and NPC inn guests.
func GenerateDefaultTown(name string) models.Town {
	mayor := GenerateNPCMayor(10)

	// Seed NPC inn guests at fixed levels
	npcLevels := []int{3, 5, 8, 12}
	guests := make([]models.InnGuest, 0, len(npcLevels))
	for _, lvl := range npcLevels {
		firstName := data.VillagerFirstNames[rand.Intn(len(data.VillagerFirstNames))]
		lastName := data.VillagerLastNames[rand.Intn(len(data.VillagerLastNames))]
		guests = append(guests, GenerateNPCGuest(firstName+" "+lastName, lvl))
	}

	return models.Town{
		Name:      name,
		InnGuests: guests,
		Mayor:     &mayor,
		Treasury: map[string]int{
			"Gold": 500,
		},
		TaxRate:     5,
		FetchQuests: []models.FetchQuest{},
		AttackLog:   []models.TownAttackLog{},
	}
}

// GenerateNPCMayor creates an NPC mayor with stats, guards, and a monster.
func GenerateNPCMayor(level int) models.MayorData {
	name := data.MayorNames[rand.Intn(len(data.MayorNames))]

	baseHP := 50 + (level * 10)
	attackRolls := (level / 5) + 2
	defenseRolls := (level / 5) + 2

	resistances := map[models.DamageType]float64{
		models.Physical:  1.0,
		models.Fire:      1.0,
		models.Ice:       1.0,
		models.Lightning: 1.0,
		models.Poison:    1.0,
	}

	equipMap := make(map[int]models.Item)
	inventory := []models.Item{}
	for i := 0; i < 3; i++ {
		rarity := 2 + (level / 5)
		if rarity > 5 {
			rarity = 5
		}
		item := GenerateItem(rarity)
		EquipBestItem(item, &equipMap, &inventory)
	}
	statsMod := CalculateItemMods(equipMap)

	guards := []models.Guard{
		GenerateGuard(level),
		GenerateGuard(level),
	}

	monsters := []models.Monster{
		GenerateMonster(data.MonsterNames[rand.Intn(len(data.MonsterNames))], level, level/3+1),
	}

	skills := AssignMonsterSkills("humanoid", level)

	return models.MayorData{
		IsNPC:         true,
		NPCName:       name,
		Level:         level,
		Guards:        guards,
		Monsters:      monsters,
		HP:            baseHP + statsMod.HitPointMod,
		MaxHP:         baseHP + statsMod.HitPointMod,
		AttackRolls:   attackRolls,
		DefenseRolls:  defenseRolls,
		StatsMod:      statsMod,
		EquipmentMap:  equipMap,
		LearnedSkills: skills,
		Resistances:   resistances,
	}
}

// MayorFromCharacter creates a MayorData snapshot from a player character.
func MayorFromCharacter(char *models.Character, accountID int64) models.MayorData {
	return models.MayorData{
		IsNPC:         false,
		AccountID:     accountID,
		CharacterName: char.Name,
		Level:         char.Level,
		Guards:        []models.Guard{},
		Monsters:      []models.Monster{},
		HP:            char.HitpointsTotal,
		MaxHP:         char.HitpointsTotal,
		AttackRolls:   char.AttackRolls,
		DefenseRolls:  char.DefenseRolls,
		StatsMod:      char.StatsMod,
		EquipmentMap:  copyEquipmentMap(char.EquipmentMap),
		LearnedSkills: append([]models.Skill{}, char.LearnedSkills...),
		Resistances:   copyResistances(char.Resistances),
	}
}

// InnGuestFromCharacter snapshots a character into an InnGuest record.
func InnGuestFromCharacter(char *models.Character, accountID int64, goldPaid int) models.InnGuest {
	return models.InnGuest{
		AccountID:     accountID,
		CharacterName: char.Name,
		CheckInTime:   time.Now().Unix(),
		GoldPaid:      goldPaid,
		HiredGuards:   []models.Guard{},
		Level:         char.Level,
		HP:            char.HitpointsTotal,
		MaxHP:         char.HitpointsTotal,
		MP:            char.ManaTotal,
		MaxMP:         char.ManaTotal,
		SP:            char.StaminaTotal,
		MaxSP:         char.StaminaTotal,
		AttackRolls:   char.AttackRolls,
		DefenseRolls:  char.DefenseRolls,
		StatsMod:      char.StatsMod,
		EquipmentMap:  copyEquipmentMap(char.EquipmentMap),
		LearnedSkills: append([]models.Skill{}, char.LearnedSkills...),
		Resistances:   copyResistances(char.Resistances),
	}
}

// InnGuestToMonster converts an InnGuest snapshot into a Monster for PvP combat.
func InnGuestToMonster(guest *models.InnGuest) models.Monster {
	return models.Monster{
		Name:               guest.CharacterName,
		Level:              guest.Level,
		Rank:               guest.Level/3 + 1,
		HitpointsTotal:     guest.MaxHP,
		HitpointsNatural:   guest.MaxHP,
		HitpointsRemaining: guest.MaxHP,
		ManaTotal:          guest.MaxMP,
		ManaNatural:        guest.MaxMP,
		ManaRemaining:      guest.MaxMP,
		StaminaTotal:       guest.MaxSP,
		StaminaNatural:     guest.MaxSP,
		StaminaRemaining:   guest.MaxSP,
		AttackRolls:        guest.AttackRolls,
		DefenseRolls:       guest.DefenseRolls,
		StatsMod:           guest.StatsMod,
		EquipmentMap:       copyEquipmentMap(guest.EquipmentMap),
		Inventory:          []models.Item{},
		LearnedSkills:      append([]models.Skill{}, guest.LearnedSkills...),
		StatusEffects:      []models.StatusEffect{},
		Resistances:        copyResistances(guest.Resistances),
		MonsterType:        "humanoid",
	}
}

// MayorToMonster converts a MayorData into a Monster for challenge combat.
func MayorToMonster(mayor *models.MayorData) models.Monster {
	name := mayor.NPCName
	if !mayor.IsNPC {
		name = "Mayor " + mayor.CharacterName
	}
	return models.Monster{
		Name:               name,
		Level:              mayor.Level,
		Rank:               mayor.Level/3 + 1,
		HitpointsTotal:     mayor.MaxHP,
		HitpointsNatural:   mayor.MaxHP,
		HitpointsRemaining: mayor.MaxHP,
		ManaTotal:          50 + mayor.Level*5,
		ManaNatural:        50 + mayor.Level*5,
		ManaRemaining:      50 + mayor.Level*5,
		StaminaTotal:       50 + mayor.Level*5,
		StaminaNatural:     50 + mayor.Level*5,
		StaminaRemaining:   50 + mayor.Level*5,
		AttackRolls:        mayor.AttackRolls,
		DefenseRolls:       mayor.DefenseRolls,
		StatsMod:           mayor.StatsMod,
		EquipmentMap:       copyEquipmentMap(mayor.EquipmentMap),
		Inventory:          []models.Item{},
		LearnedSkills:      append([]models.Skill{}, mayor.LearnedSkills...),
		StatusEffects:      []models.StatusEffect{},
		Resistances:        copyResistances(mayor.Resistances),
		MonsterType:        "humanoid",
		IsBoss:             true,
	}
}

// CalculateTax returns the net amount after tax and the tax amount.
func CalculateTax(amount int, taxRate int) (int, int) {
	if taxRate <= 0 || amount <= 0 {
		return amount, 0
	}
	if taxRate > 50 {
		taxRate = 50
	}
	taxAmount := (amount * taxRate) / 100
	if taxAmount < 1 && amount > 0 && taxRate > 0 {
		taxAmount = 1
	}
	return amount - taxAmount, taxAmount
}

// GenerateNPCGuest creates an NPC inn guest with equipment, skills, guards, and gold.
func GenerateNPCGuest(name string, level int) models.InnGuest {
	rank := level/3 + 1
	if rank < 1 {
		rank = 1
	}

	baseHP := MultiRoll(rank) + 20
	baseMP := MultiRoll(rank) + 10
	baseSP := MultiRoll(rank) + 10
	attackRolls := (level / 5) + 1
	defenseRolls := (level / 5) + 1

	resistances := map[models.DamageType]float64{
		models.Physical:  1.0,
		models.Fire:      1.0,
		models.Ice:       1.0,
		models.Lightning: 1.0,
		models.Poison:    1.0,
	}

	equipMap := make(map[int]models.Item)
	inventory := []models.Item{}
	numItems := 2 + (level / 5)
	if numItems > 4 {
		numItems = 4
	}
	for i := 0; i < numItems; i++ {
		rarity := 1 + (level / 5)
		if rarity > 5 {
			rarity = 5
		}
		item := GenerateItem(rarity)
		EquipBestItem(item, &equipMap, &inventory)
	}
	statsMod := CalculateItemMods(equipMap)

	skills := AssignMonsterSkills("humanoid", level)

	numGuards := 1 + rand.Intn(2)
	guards := make([]models.Guard, numGuards)
	for i := 0; i < numGuards; i++ {
		guards[i] = GenerateGuard(level)
	}

	goldCarried := 50 + level*20

	return models.InnGuest{
		AccountID:     0,
		CharacterName: name,
		CheckInTime:   time.Now().Unix(),
		GoldCarried:   goldCarried,
		HiredGuards:   guards,
		Level:         level,
		HP:            baseHP + statsMod.HitPointMod,
		MaxHP:         baseHP + statsMod.HitPointMod,
		MP:            baseMP,
		MaxMP:         baseMP,
		SP:            baseSP,
		MaxSP:         baseSP,
		AttackRolls:   attackRolls,
		DefenseRolls:  defenseRolls,
		StatsMod:      statsMod,
		EquipmentMap:  equipMap,
		LearnedSkills: skills,
		Resistances:   resistances,
	}
}

// ReplenishNPCGuests ensures the town has at least 3 NPC guests, filling up to 4.
func ReplenishNPCGuests(town *models.Town) {
	npcCount := 0
	for _, guest := range town.InnGuests {
		if guest.AccountID == 0 {
			npcCount++
		}
	}
	for npcCount < 4 {
		level := rand.Intn(15) + 1
		firstName := data.VillagerFirstNames[rand.Intn(len(data.VillagerFirstNames))]
		lastName := data.VillagerLastNames[rand.Intn(len(data.VillagerLastNames))]
		name := firstName + " " + lastName
		guest := GenerateNPCGuest(name, level)
		town.InnGuests = append(town.InnGuests, guest)
		npcCount++
	}
}

// CleanExpiredGuests removes inn guests checked in more than maxAge seconds ago.
// NPC guests (AccountID == 0) are never expired.
func CleanExpiredGuests(town *models.Town, maxAge int64) {
	now := time.Now().Unix()
	kept := []models.InnGuest{}
	for _, guest := range town.InnGuests {
		if guest.AccountID == 0 || now-guest.CheckInTime < maxAge {
			kept = append(kept, guest)
		}
	}
	town.InnGuests = kept
}

// InnSleepCost returns the gold cost for sleeping at the inn.
func InnSleepCost(level int) int {
	return 10 + level*5
}

func copyEquipmentMap(m map[int]models.Item) map[int]models.Item {
	if m == nil {
		return make(map[int]models.Item)
	}
	cp := make(map[int]models.Item, len(m))
	for k, v := range m {
		cp[k] = v
	}
	return cp
}

func copyResistances(m map[models.DamageType]float64) map[models.DamageType]float64 {
	if m == nil {
		return map[models.DamageType]float64{
			models.Physical:  1.0,
			models.Fire:      1.0,
			models.Ice:       1.0,
			models.Lightning: 1.0,
			models.Poison:    1.0,
		}
	}
	cp := make(map[models.DamageType]float64, len(m))
	for k, v := range m {
		cp[k] = v
	}
	return cp
}
