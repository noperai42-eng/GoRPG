package engine

import (
	"testing"
	"time"

	"rpg-game/pkg/game"
	"rpg-game/pkg/models"
)

// ============================================================
// TOWN VIEW TESTS
// ============================================================

// TestMakeTownViewInnGuestNPCFields verifies IsNPC and GoldCarried appear in the view.
func TestMakeTownViewInnGuestNPCFields(t *testing.T) {
	town := game.GenerateDefaultTown("TestTown")

	view := MakeTownView(&town, 999, "SomePlayer")

	if view == nil {
		t.Fatal("TownView should not be nil")
	}

	if len(view.Guests) != len(town.InnGuests) {
		t.Errorf("Guest count mismatch: view=%d, town=%d", len(view.Guests), len(town.InnGuests))
	}

	for i, gv := range view.Guests {
		guest := town.InnGuests[i]

		if gv.IsNPC != (guest.AccountID == 0) {
			t.Errorf("Guest %d: IsNPC should be %v", i, guest.AccountID == 0)
		}

		if gv.GoldCarried != guest.GoldCarried {
			t.Errorf("Guest %d: GoldCarried mismatch: view=%d, model=%d",
				i, gv.GoldCarried, guest.GoldCarried)
		}

		if gv.Level != guest.Level {
			t.Errorf("Guest %d: Level mismatch: view=%d, model=%d",
				i, gv.Level, guest.Level)
		}

		if gv.GuardCount != len(guest.HiredGuards) {
			t.Errorf("Guest %d: GuardCount mismatch: view=%d, model=%d",
				i, gv.GuardCount, len(guest.HiredGuards))
		}

		// NPC guests should not be flagged as "own"
		if gv.IsOwn {
			t.Errorf("Guest %d: NPC should not be flagged as IsOwn", i)
		}
	}

	t.Logf("TownView: %d guests, all NPC fields correctly populated", len(view.Guests))
}

// TestMakeTownViewPlayerGuestIsOwn verifies IsOwn flag for player guests.
func TestMakeTownViewPlayerGuestIsOwn(t *testing.T) {
	town := models.Town{
		Name: "TestTown",
		InnGuests: []models.InnGuest{
			{AccountID: 42, CharacterName: "MyHero", Level: 5, CheckInTime: time.Now().Unix()},
			{AccountID: 99, CharacterName: "OtherPlayer", Level: 8, CheckInTime: time.Now().Unix()},
			{AccountID: 0, CharacterName: "NPC Guest", Level: 3, GoldCarried: 110},
		},
		Mayor: &models.MayorData{IsNPC: true, NPCName: "TestMayor", Level: 10},
	}

	view := MakeTownView(&town, 42, "MyHero")

	if !view.Guests[0].IsOwn {
		t.Error("First guest (own character) should have IsOwn=true")
	}

	if view.Guests[1].IsOwn {
		t.Error("Second guest (other player) should have IsOwn=false")
	}

	if view.Guests[2].IsOwn {
		t.Error("Third guest (NPC) should have IsOwn=false")
	}

	if !view.Guests[2].IsNPC {
		t.Error("Third guest should have IsNPC=true")
	}

	if view.Guests[2].GoldCarried != 110 {
		t.Errorf("NPC gold carried should be 110, got %d", view.Guests[2].GoldCarried)
	}
}

// ============================================================
// TOWN ENGINE FLOW TESTS
// ============================================================

// TestTownInnShowsNPCGuests verifies navigating to Town → Inn shows NPC guests.
func TestTownInnShowsNPCGuests(t *testing.T) {
	eng, sessionID := createTestEngine(t)
	eng.ProcessCommand(sessionID, GameCommand{Type: "init"})

	// Navigate to Town (option "11")
	resp := eng.ProcessCommand(sessionID, GameCommand{Type: "select", Value: "11"})
	if resp.Type == "error" {
		t.Fatalf("Town navigation failed: %v", messagesText(resp.Messages))
	}

	// Town main should return town state
	if resp.State == nil || resp.State.Town == nil {
		t.Fatal("Expected town state data from town main menu")
	}

	townView := resp.State.Town
	npcCount := 0
	for _, guest := range townView.Guests {
		if guest.IsNPC {
			npcCount++
			if guest.GoldCarried <= 0 {
				t.Errorf("NPC %s should carry gold, got %d", guest.CharacterName, guest.GoldCarried)
			}
		}
	}

	if npcCount == 0 {
		t.Error("Town should have NPC guests")
	}

	t.Logf("Town: %d total guests, %d NPCs", len(townView.Guests), npcCount)
}

// messagesText concatenates all message texts for diagnostic output.
func messagesText(msgs []GameMessage) string {
	out := ""
	for _, m := range msgs {
		out += m.Text + " | "
	}
	return out
}
