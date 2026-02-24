package main

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"rpg-game/pkg/engine"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	eng := engine.NewEngine()
	sessionID, err := eng.CreateLocalSession("gamestate.json")
	if err != nil {
		log.Fatal(err)
	}

	// Send init command to set up character selection / main menu
	resp := eng.ProcessCommand(sessionID, engine.GameCommand{Type: "init"})
	renderResponse(resp)

	scanner := bufio.NewScanner(os.Stdin)

	for {
		// Show prompt
		if resp.Prompt != "" {
			fmt.Print(resp.Prompt)
		} else if len(resp.Options) > 0 {
			fmt.Print("Enter input:\n")
		}

		if resp.Type == "exit" {
			break
		}

		if !scanner.Scan() {
			break
		}
		input := scanner.Text()

		// Determine command type based on whether we have a prompt or options
		cmdType := "select"
		if resp.Prompt != "" {
			cmdType = "input"
		}

		resp = eng.ProcessCommand(sessionID, engine.GameCommand{Type: cmdType, Value: input})
		renderResponse(resp)
	}

	if err := eng.SaveSession(sessionID); err != nil {
		log.Printf("Error saving: %v\n", err)
	}
	fmt.Println("Exiting the application.")
}

func renderResponse(resp engine.GameResponse) {
	// Print all messages
	for _, m := range resp.Messages {
		fmt.Println(m.Text)
	}

	// Show options if present and no text prompt
	if len(resp.Options) > 0 && resp.Prompt == "" {
		fmt.Println()
		for _, o := range resp.Options {
			marker := ""
			if !o.Enabled {
				marker = " (unavailable)"
			}
			fmt.Printf("%s = %s%s\n", o.Key, o.Label, marker)
		}
	}

	// Show combat HUD if in combat
	if resp.State != nil && resp.State.Combat != nil {
		renderCombatHUD(resp.State.Combat)
	}
}

func renderCombatHUD(cv *engine.CombatView) {
	fmt.Printf("\n========== TURN %d ==========\n", cv.Turn)
	fmt.Printf("[Player] HP:%d/%d | MP:%d/%d | SP:%d/%d\n",
		cv.PlayerHP, cv.PlayerMaxHP,
		cv.PlayerMP, cv.PlayerMaxMP,
		cv.PlayerSP, cv.PlayerMaxSP)

	fmt.Printf("[%s] HP:%d/%d | MP:%d/%d | SP:%d/%d\n",
		cv.MonsterName,
		cv.MonsterHP, cv.MonsterMaxHP,
		cv.MonsterMP, cv.MonsterMaxMP,
		cv.MonsterSP, cv.MonsterMaxSP)

	if len(cv.PlayerEffects) > 0 {
		effects := []string{}
		for _, eff := range cv.PlayerEffects {
			effects = append(effects, fmt.Sprintf("[%s:%d]", eff.Name, eff.Duration))
		}
		fmt.Printf("Player effects: %s\n", strings.Join(effects, " "))
	}

	if len(cv.MonsterEffects) > 0 {
		effects := []string{}
		for _, eff := range cv.MonsterEffects {
			effects = append(effects, fmt.Sprintf("[%s:%d]", eff.Name, eff.Duration))
		}
		fmt.Printf("%s effects: %s\n", cv.MonsterName, strings.Join(effects, " "))
	}

	if len(cv.Guards) > 0 {
		fmt.Println("Guards:")
		for _, g := range cv.Guards {
			status := ""
			if g.Injured {
				status = " [INJURED]"
			}
			fmt.Printf("  %s HP:%d/%d%s\n", g.Name, g.HP, g.MaxHP, status)
		}
	}
}
