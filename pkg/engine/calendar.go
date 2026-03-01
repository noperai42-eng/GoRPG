package engine

import (
	"fmt"
	"time"
)

// GameCalendar represents a point in game time.
// 1 real minute = 1 game hour, 24 game hours = 1 day,
// 30 days = 1 cycle, 4 cycles = 1 year.
type GameCalendar struct {
	Day   int `json:"day"`   // 1-30
	Cycle int `json:"cycle"` // 1-4
	Year  int `json:"year"`  // 1+
	Hour  int `json:"hour"`  // 0-23
}

// gameEpoch is the fixed epoch for the game calendar.
var gameEpoch = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

// CurrentGameCalendar computes the current game calendar from real elapsed time.
func CurrentGameCalendar() GameCalendar {
	elapsed := time.Since(gameEpoch)
	totalGameHours := int(elapsed.Minutes()) // 1 real minute = 1 game hour

	hour := totalGameHours % 24
	totalGameDays := totalGameHours / 24

	day := (totalGameDays % 30) + 1       // 1-30
	totalCycles := totalGameDays / 30
	cycle := (totalCycles % 4) + 1         // 1-4
	year := int(totalCycles/4) + 1         // 1+

	return GameCalendar{
		Day:   day,
		Cycle: cycle,
		Year:  year,
		Hour:  hour,
	}
}

// FormatGameTime returns a human-readable game time string.
func (gc GameCalendar) FormatGameTime() string {
	return fmt.Sprintf("Day %d, Cycle %d, Year %d", gc.Day, gc.Cycle, gc.Year)
}

// MoonPhase represents a phase of the lunar cycle.
type MoonPhase struct {
	Name  string `json:"name"`
	Emoji string `json:"emoji"`
}

// MoonPhaseFromDay maps a day (1-30) to its lunar phase.
func MoonPhaseFromDay(day int) MoonPhase {
	switch {
	case day <= 2:
		return MoonPhase{Name: "New Moon", Emoji: "\U0001F311"}
	case day <= 5:
		return MoonPhase{Name: "Waxing Crescent", Emoji: "\U0001F312"}
	case day <= 9:
		return MoonPhase{Name: "First Quarter", Emoji: "\U0001F313"}
	case day <= 12:
		return MoonPhase{Name: "Waxing Gibbous", Emoji: "\U0001F314"}
	case day <= 17:
		return MoonPhase{Name: "Full Moon", Emoji: "\U0001F315"}
	case day <= 20:
		return MoonPhase{Name: "Waning Gibbous", Emoji: "\U0001F316"}
	case day <= 24:
		return MoonPhase{Name: "Last Quarter", Emoji: "\U0001F317"}
	default:
		return MoonPhase{Name: "Waning Crescent", Emoji: "\U0001F318"}
	}
}

// IsRaidPhase returns true when the tide leader raid is active (day 8+).
func (gc GameCalendar) IsRaidPhase() bool {
	return gc.Day >= 8
}
