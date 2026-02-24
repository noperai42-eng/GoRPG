package game

import (
	"math/rand"
	"time"
)

func RollDice() int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(6) + 1
}

func MultiRoll(rolls int) int {
	total := 0
	for i := 0; i < rolls; i++ {
		total += RollDice()
	}
	return total
}

func RollUntilSix(minRolls int, minMod int) int {
	rolls := 0
	resultTotal := 0
	for rolls < minRolls {
		result := RollDice() + minMod
		if result > 6 {
			result = 6
		}
		resultTotal += result

		if result == 6 {
			rolls = 0
		} else {
			rolls++
		}
	}
	return resultTotal
}
