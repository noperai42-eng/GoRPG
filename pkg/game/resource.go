package game

import (
	"fmt"
	"rpg-game/pkg/models"
)

func HarvestResource(resourceType string, resourceStorage *map[string]models.Resource) int {
	result := 0
	resource, exists := (*resourceStorage)[resourceType]
	if exists {
		result = RollUntilSix(1, resource.RollModifier)
		resource.Stock += result
		(*resourceStorage)[resourceType] = resource
	}
	fmt.Printf("Resources: %v\n", *resourceStorage)
	fmt.Printf("Harvested %d %s\n", result, resourceType)
	return result
}
