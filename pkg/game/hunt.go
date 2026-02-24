package game

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"

	"rpg-game/pkg/models"
)

func GoHunt(game *models.GameState, location *models.Location, huntCount int, player *models.Character) {
	for i := 0; i < huntCount; i++ {
		fmt.Println("Fights Remaining: ", huntCount-i)
		if player.HitpointsRemaining <= 0 {
			fmt.Printf("Player %s HAS RESURRECTED!!!!\n\n", player.Name)
			player.HitpointsRemaining = player.HitpointsTotal
			player.Resurrections++
		}

		var mobLoc int
		var mob models.Monster

		// Check if player has Tracking skill
		hasTracking := false
		for _, skill := range player.LearnedSkills {
			if skill.Name == "Tracking" {
				hasTracking = true
				break
			}
		}

		if hasTracking {
			fmt.Println("\nTRACKING ACTIVE - Choose your target:")
			fmt.Println("============================================================")
			for idx, monster := range location.Monsters {
				guardianTag := ""
				if monster.IsSkillGuardian {
					guardianTag = " [SKILL GUARDIAN]"
				}
				fmt.Printf("%d. %s (Lv%d) - HP:%d/%d%s\n",
					idx+1, monster.Name, monster.Level,
					monster.HitpointsRemaining, monster.HitpointsTotal,
					guardianTag)
			}
			fmt.Println("============================================================")
			fmt.Printf("Choose target (1-%d, or 0 for random): ", len(location.Monsters))

			scanner := bufio.NewScanner(os.Stdin)
			scanner.Scan()
			choice := scanner.Text()
			choiceNum, err := strconv.Atoi(choice)

			if err != nil || choiceNum < 0 || choiceNum > len(location.Monsters) {
				mobLoc = rand.Intn(len(location.Monsters))
			} else if choiceNum == 0 {
				mobLoc = rand.Intn(len(location.Monsters))
			} else {
				mobLoc = choiceNum - 1
			}
		} else {
			mobLoc = rand.Intn(len(location.Monsters))
		}

		mob = location.Monsters[mobLoc]
		PrintMonster(mob)
		FightToTheDeath(player, game, &mob, location, mobLoc)
		LevelUp(player)
		LevelUpMob(&location.Monsters[mobLoc])
		CheckQuestProgress(player, game)
	}
}

func AutoPlayMode(gameState *models.GameState, player *models.Character, speed string) {
	delays := map[string]int{
		"slow":   2000,
		"normal": 1000,
		"fast":   500,
		"turbo":  100,
	}
	delay := delays[speed]
	if delay == 0 {
		delay = 1000
	}

	fmt.Printf("\nAUTO-PLAY MODE ACTIVATED\n")
	fmt.Printf("Speed: %s (%dms delay)\n", speed, delay)
	fmt.Printf("Character: %s (Level %d)\n", player.Name, player.Level)
	fmt.Printf("Press ENTER at any time to stop\n\n")

	fightCount := 0
	totalXP := 0
	wins := 0
	deaths := 0
	startTime := time.Now()

	stopChan := make(chan bool)
	go func() {
		bufio.NewReader(os.Stdin).ReadBytes('\n')
		stopChan <- true
	}()

	var huntLocation *models.Location
	for _, locName := range player.KnownLocations {
		loc := gameState.GameLocations[locName]
		if loc.Type != "Base" {
			huntLocation = &loc
			break
		}
	}

	if huntLocation == nil {
		fmt.Println("No huntable locations available!")
		return
	}

	fmt.Printf("Hunting at: %s\n", huntLocation.Name)
	fmt.Println("=====================================")

gameLoop:
	for {
		select {
		case <-stopChan:
			fmt.Printf("\n\nAUTO-PLAY STOPPED BY USER\n\n")
			break gameLoop
		default:
		}

		fightCount++

		if player.HitpointsRemaining <= 0 {
			player.HitpointsRemaining = player.HitpointsTotal
			player.Resurrections++
			deaths++
			fmt.Printf("\nRESURRECTION #%d\n\n", player.Resurrections)
		}

		mobLoc := rand.Intn(len(huntLocation.Monsters))
		mob := huntLocation.Monsters[mobLoc]

		startXP := player.Experience
		AutoFightToTheDeath(player, gameState, &mob, huntLocation, mobLoc)

		xpGained := player.Experience - startXP
		if xpGained > 0 {
			wins++
			totalXP += xpGained
		}

		LevelUp(player)
		LevelUpMob(&huntLocation.Monsters[mobLoc])
		CheckQuestProgress(player, gameState)

		if fightCount%10 == 0 {
			fmt.Printf("\nAUTO-PLAY STATISTICS\n")
			fmt.Printf("Fights: %d | Wins: %d | Deaths: %d\n", fightCount, wins, deaths)
			fmt.Printf("Level: %d | XP: %d | Total XP Gained: %d\n", player.Level, player.Experience, totalXP)
			fmt.Printf("HP: %d/%d | MP: %d/%d | SP: %d/%d\n",
				player.HitpointsRemaining, player.HitpointsTotal,
				player.ManaRemaining, player.ManaTotal,
				player.StaminaRemaining, player.StaminaTotal)
			fmt.Printf("Skills: %d | Inventory: %d items\n", len(player.LearnedSkills), len(player.Inventory))
			fmt.Println("=====================================")
			fmt.Println("(Press ENTER to stop)")
		}

		if fightCount%50 == 0 {
			gameState.CharactersMap[player.Name] = *player
			WriteGameStateToFile(*gameState, "gamestate.json")
			fmt.Printf("Game saved (Fight #%d)\n\n", fightCount)
		}

		time.Sleep(time.Duration(delay) * time.Millisecond)
	}

	duration := time.Since(startTime)
	ShowAutoPlaySummary(fightCount, wins, deaths, totalXP, duration, player)
	ShowPostAutoPlayMenu(gameState, player)
}

func ShowAutoPlaySummary(fights int, wins int, deaths int, xp int, duration time.Duration, player *models.Character) {
	fmt.Println("\n============================================================")
	fmt.Println("AUTO-PLAY SESSION COMPLETE")
	fmt.Println("============================================================")
	fmt.Printf("Duration: %s\n", duration.Round(time.Second))
	fmt.Printf("Total Fights: %d\n", fights)
	if fights > 0 {
		fmt.Printf("Victories: %d (%.1f%%)\n", wins, float64(wins)/float64(fights)*100)
	}
	fmt.Printf("Deaths: %d\n", deaths)
	fmt.Printf("XP Gained: %d\n", xp)
	fmt.Printf("\nFinal Character State:\n")
	fmt.Printf("  Level: %d (XP: %d)\n", player.Level, player.Experience)
	fmt.Printf("  HP: %d/%d\n", player.HitpointsRemaining, player.HitpointsTotal)
	fmt.Printf("  MP: %d/%d\n", player.ManaRemaining, player.ManaTotal)
	fmt.Printf("  SP: %d/%d\n", player.StaminaRemaining, player.StaminaTotal)
	fmt.Printf("  Skills Known: %d\n", len(player.LearnedSkills))
	fmt.Printf("  Inventory Items: %d\n", len(player.Inventory))
	fmt.Println("============================================================")
}

func ShowPostAutoPlayMenu(gameState *models.GameState, player *models.Character) {
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Println("\n--- POST AUTO-PLAY MENU ---")
		fmt.Println("1 = View Inventory")
		fmt.Println("2 = View Skills")
		fmt.Println("3 = View Equipment")
		fmt.Println("4 = View Quest Log")
		fmt.Println("5 = View Full Character Stats")
		fmt.Println("6 = Resume Auto-Play")
		fmt.Println("0 = Return to Main Menu")
		fmt.Print("Choice: ")

		scanner.Scan()
		choice := scanner.Text()

		switch choice {
		case "1":
			ShowInventory(player)
		case "2":
			ShowSkills(player)
		case "3":
			ShowEquipment(player)
		case "4":
			ShowQuestLog(player, gameState)
		case "5":
			PrintCharacter(*player)
		case "6":
			fmt.Println("\nSelect speed:")
			fmt.Println("1 = Slow (2s per fight)")
			fmt.Println("2 = Normal (1s per fight)")
			fmt.Println("3 = Fast (0.5s per fight)")
			fmt.Println("4 = Turbo (0.1s per fight)")
			fmt.Print("Choice: ")
			scanner.Scan()
			speedChoice := scanner.Text()

			speedMap := map[string]string{
				"1": "slow",
				"2": "normal",
				"3": "fast",
				"4": "turbo",
			}

			speed := speedMap[speedChoice]
			if speed == "" {
				speed = "normal"
			}

			gameState.CharactersMap[player.Name] = *player
			WriteGameStateToFile(*gameState, "gamestate.json")

			AutoPlayMode(gameState, player, speed)
			return
		case "0":
			gameState.CharactersMap[player.Name] = *player
			WriteGameStateToFile(*gameState, "gamestate.json")
			fmt.Println("\nGame saved")
			return
		default:
			fmt.Println("Invalid choice")
		}
	}
}
