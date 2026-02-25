package game

import (
	"math/rand"

	"rpg-game/pkg/data"
	"rpg-game/pkg/models"
)

func GenerateItem(rarity int) models.Item {
	slot := rand.Intn(8)
	name := generateGearName(slot)
	statsMod := models.StatMod{AttackMod: 0, DefenseMod: 0, HitPointMod: 0}
	item := models.Item{Name: name, Rarity: rarity, Slot: slot, StatsMod: statsMod, CP: 0, ItemType: "equipment"}

	for i := 0; i < rarity; i++ {
		statChoice := rand.Intn(3)
		switch statChoice {
		case 0:
			item.StatsMod.AttackMod += RollUntilSix(1, 0)
		case 1:
			item.StatsMod.DefenseMod += RollUntilSix(1, 0)
		case 2:
			item.StatsMod.HitPointMod += RollUntilSix(1, 0)
		}
	}

	item.CP = item.StatsMod.AttackMod + item.StatsMod.DefenseMod + item.StatsMod.HitPointMod
	return item
}

func generateGearName(slot int) string {
	prefix := data.ItemPrefixes[rand.Intn(len(data.ItemPrefixes))]
	gearNames := data.SlotGearNames[slot]
	base := gearNames[rand.Intn(len(gearNames))]
	return prefix + " " + base
}

func CreateHealthPotion(size string) models.Item {
	var name string
	var healAmount int

	switch size {
	case "small":
		name = "Small Health Potion"
		healAmount = 15
	case "medium":
		name = "Medium Health Potion"
		healAmount = 30
	case "large":
		name = "Large Health Potion"
		healAmount = 50
	default:
		name = "Health Potion"
		healAmount = 20
	}

	return models.Item{
		Name:     name,
		ItemType: "consumable",
		Rarity:   1,
		Slot:     -1,
		CP:       0,
		Consumable: models.ConsumableEffect{
			EffectType: "heal",
			Value:      healAmount,
			Duration:   0,
		},
	}
}

func CreateSkillScroll(skill models.Skill) models.Item {
	scrollName := skill.Name + " Scroll"

	craftingValue := 10 + skill.Damage + skill.ManaCost + skill.StaminaCost
	if skill.Effect.Type != "none" {
		craftingValue += skill.Effect.Potency * skill.Effect.Duration
	}

	return models.Item{
		Name:     scrollName,
		ItemType: "skill_scroll",
		Rarity:   3,
		Slot:     -1,
		CP:       0,
		SkillScroll: models.SkillScrollData{
			Skill:         skill,
			CanBeCrafted:  true,
			CraftingValue: craftingValue,
		},
	}
}

func EquipBestItem(newItem models.Item, equipment *map[int]models.Item, inventory *[]models.Item) {
	if newItem.ItemType == "consumable" || newItem.ItemType == "skill_scroll" {
		(*inventory) = append((*inventory), newItem)
		return
	}

	currentItem, ok := (*equipment)[newItem.Slot]
	if ok {
		if newItem.CP > currentItem.CP {
			(*equipment)[newItem.Slot] = newItem
			(*inventory) = append((*inventory), currentItem)
		}
	} else {
		(*equipment)[newItem.Slot] = newItem
	}
}

func CalculateItemMods(equipment map[int]models.Item) models.StatMod {
	statMod := models.StatMod{AttackMod: 0, DefenseMod: 0, HitPointMod: 0}
	for _, item := range equipment {
		statMod.AttackMod += item.StatsMod.AttackMod
		statMod.DefenseMod += item.StatsMod.DefenseMod
		statMod.HitPointMod += item.StatsMod.HitPointMod
	}
	return statMod
}

func UseConsumableItem(item models.Item, character *models.Character) bool {
	if item.ItemType != "consumable" {
		return false
	}

	switch item.Consumable.EffectType {
	case "heal":
		healAmount := item.Consumable.Value
		character.HitpointsRemaining += healAmount
		if character.HitpointsRemaining > character.HitpointsTotal {
			character.HitpointsRemaining = character.HitpointsTotal
		}
		return true
	default:
		return false
	}
}

func RemoveItemFromInventory(inventory *[]models.Item, index int) {
	*inventory = append((*inventory)[:index], (*inventory)[index+1:]...)
}

func DropBeastMaterial(monsterType string, player *models.Character) (string, int) {
	var materials []string
	var dropChance int

	// Check per-monster override first
	if md, ok := monsterMaterialOverrides[monsterType]; ok {
		materials = md.Materials
		dropChance = md.DropChance
	} else {
		// Fall back to category
		cat := data.MonsterCategory[monsterType]
		if cat == "" {
			cat = "humanoid"
		}
		if md, ok := categoryMaterials[cat]; ok {
			materials = md.Materials
			dropChance = md.DropChance
		} else {
			materials = []string{"Beast Skin", "Beast Bone"}
			dropChance = 40
		}
	}

	if rand.Intn(100) < dropChance {
		material := materials[rand.Intn(len(materials))]
		quantity := rand.Intn(3) + 1

		resource, exists := player.ResourceStorageMap[material]
		if !exists {
			resource = models.Resource{Name: material, Stock: 0, RollModifier: 0}
		}
		resource.Stock += quantity
		player.ResourceStorageMap[material] = resource

		return material, quantity
	}
	return "", 0
}

func EquipGuardItem(newItem models.Item, equipment *map[int]models.Item, inventory *[]models.Item) {
	if newItem.ItemType == "consumable" || newItem.ItemType == "skill_scroll" {
		(*inventory) = append((*inventory), newItem)
		return
	}

	currentItem, ok := (*equipment)[newItem.Slot]
	if ok {
		if newItem.CP > currentItem.CP {
			(*equipment)[newItem.Slot] = newItem
			(*inventory) = append((*inventory), currentItem)
		} else {
			(*inventory) = append((*inventory), newItem)
		}
	} else {
		(*equipment)[newItem.Slot] = newItem
	}
}
