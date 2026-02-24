package game

import (
	"encoding/json"
	"fmt"
	"os"
	"rpg-game/pkg/models"
)

func WriteGameStateToFile(gameState models.GameState, filename string) error {
	jsonData, err := json.MarshalIndent(gameState, "", "  ")
	if err != nil {
		return err
	}
	err = os.WriteFile(filename, jsonData, 0644)
	if err != nil {
		return err
	}
	fmt.Println("GameState written to", filename)
	return nil
}

func LoadGameStateFromFile(filename string) (models.GameState, error) {
	jsonData, err := os.ReadFile(filename)
	if err != nil {
		return models.GameState{}, err
	}
	var gameState models.GameState
	err = json.Unmarshal(jsonData, &gameState)
	if err != nil {
		return models.GameState{}, err
	}
	return gameState, nil
}
