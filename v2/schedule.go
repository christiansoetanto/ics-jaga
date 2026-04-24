package main

import (
	"fmt"
	"log"
	"strings"
	"time"
)

// ── Config ────────────────────────────────────────────────────────────────────
// Edit these values to change the schedule.

var config = struct {
	Year      int
	Month     time.Month
	FirstDay  int
	Shifts    string
	CredsFile string
}{
	Year:      2026,
	Month:     time.May,
	FirstDay:  1,
	Shifts:    "Mg P Mg M LM Mg P P M LM Mg P P Mg Mg P M LM Mg Mg P Mg P M LM Mg M LM Mg M LM",
	CredsFile: "credentials.json",
}

// ── Shift definitions ─────────────────────────────────────────────────────────

type Shift struct {
	Title     string
	Start     string // "HH:MM", empty = all-day
	End       string // "HH:MM"
	NextDay   bool   // end time falls on the following date
	WithAlarm bool
}

var shiftMap = map[string]Shift{
	"P":  {Title: "Pagi", Start: "08:00", End: "16:00", NextDay: false, WithAlarm: true},
	"M":  {Title: "Malam", Start: "16:00", End: "08:00", NextDay: true, WithAlarm: true},
	"LM": {Title: "Lepas Malam", Start: "08:00", End: "23:59", NextDay: false, WithAlarm: false},
	"Mg": {Title: "OFF", Start: "", End: "", NextDay: false, WithAlarm: false},
}

// ── Event model ───────────────────────────────────────────────────────────────

type Event struct {
	Summary   string
	Date      string // YYYY-MM-DD (start date)
	EndDate   string // YYYY-MM-DD (end date; for all-day this is exclusive)
	StartTime string // HH:MM, empty if all-day
	EndTime   string // HH:MM, empty if all-day
	AllDay    bool
	WithAlarm bool
}

func (e Event) String() string {
	if e.AllDay {
		return fmt.Sprintf("[%s] %-14s  all-day", e.Date, e.Summary)
	}
	alarm := ""
	if e.WithAlarm {
		alarm = "  [alarm 12h]"
	}
	return fmt.Sprintf("[%s] %-14s  %sT%s → %sT%s%s",
		e.Date, e.Summary,
		e.Date, e.StartTime,
		e.EndDate, e.EndTime,
		alarm,
	)
}

// ── Build ─────────────────────────────────────────────────────────────────────

func calendarName() string {
	return fmt.Sprintf("Jaga %s %d", config.Month.String(), config.Year)
}

func buildEvents() []Event {
	codes := strings.Fields(config.Shifts)
	logger.Infof("building %d events for %s %d", len(codes), config.Month, config.Year)

	var events []Event
	for i, code := range codes {
		shift, ok := shiftMap[code]
		if !ok {
			log.Fatalf("unknown shift code %q — valid: P, M, LM, Mg", code)
		}

		date := time.Date(config.Year, config.Month, config.FirstDay+i, 0, 0, 0, 0, time.Local)
		next := date.AddDate(0, 0, 1)

		ev := Event{
			Summary:   shift.Title,
			Date:      date.Format("2006-01-02"),
			WithAlarm: shift.WithAlarm,
		}

		if shift.Start == "" {
			ev.AllDay = true
			ev.EndDate = next.Format("2006-01-02")
		} else {
			ev.StartTime = shift.Start
			ev.EndTime = shift.End
			if shift.NextDay {
				ev.EndDate = next.Format("2006-01-02")
			} else {
				ev.EndDate = date.Format("2006-01-02")
			}
		}

		logger.Debugf("event %d: %s", i+1, ev)
		events = append(events, ev)
	}
	return events
}

// ── Liturgical calendar ───────────────────────────────────────────────────────

// easterDate returns Easter Sunday for the given year using the
// Meeus/Jones/Butcher algorithm.
func easterDate(year int) time.Time {
	a := year % 19
	b := year / 100
	c := year % 100
	d := b / 4
	e := b % 4
	f := (b + 8) / 25
	g := (b - f + 1) / 3
	h := (19*a + b - d - g + 15) % 30
	i := c / 4
	k := c % 4
	l := (32 + 2*e + 2*i - h - k) % 7
	m := (a + 11*h + 22*l) / 451
	month := (h + l - 7*m + 114) / 31
	day := ((h+l-7*m+114)%31 + 1)
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
}

// firstSundayOfAdvent returns the first day of Advent (4th Sunday before Dec 25).
func firstSundayOfAdvent(year int) time.Time {
	christmas := time.Date(year, time.December, 25, 0, 0, 0, 0, time.UTC)
	daysSinceSunday := int(christmas.Weekday()) // Sunday=0
	nearestSunday := christmas.AddDate(0, 0, -daysSinceSunday)
	return nearestSunday.AddDate(0, 0, -21) // 3 Sundays back = 4th before Christmas
}

// LiturgicalColor holds the Google Calendar hex colors for a season.
type LiturgicalColor struct {
	BG      string
	FG      string
	Season  string
}

// liturgicalColor returns the dominant liturgical color for the given year/month,
// based on the Roman Rite calendar. It uses the color of the 1st day of the month.
//
// Season → Google Calendar color mapping:
//
//	Christmas / Easter  → Gold   #f6bf26 / #000000
//	Lent / Advent       → Grape  #7986cb / #ffffff
//	Ordinary Time       → Sage   #33b679 / #ffffff
//	Pentecost Sunday    → Tomato #d50000 / #ffffff
func liturgicalColor(year int, month time.Month) LiturgicalColor {
	easter := easterDate(year)
	ashWed := easter.AddDate(0, 0, -46)
	pentecost := easter.AddDate(0, 0, 49)
	advent := firstSundayOfAdvent(year)

	// End of Christmas season = Baptism of the Lord (Sunday after Jan 6)
	epiphany := time.Date(year, time.January, 6, 0, 0, 0, 0, time.UTC)
	daysToSun := int((7 - int(epiphany.Weekday())) % 7)
	if daysToSun == 0 {
		daysToSun = 7
	}
	baptismOfLord := epiphany.AddDate(0, 0, daysToSun)

	d := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)

	switch {
	case d.Before(baptismOfLord):
		return LiturgicalColor{"#f6bf26", "#000000", "Christmas"}
	case d.Before(ashWed):
		return LiturgicalColor{"#33b679", "#ffffff", "Ordinary Time"}
	case d.Before(easter):
		return LiturgicalColor{"#7986cb", "#ffffff", "Lent"}
	case d.Equal(pentecost):
		return LiturgicalColor{"#d50000", "#ffffff", "Pentecost"}
	case d.Before(pentecost.AddDate(0, 0, 1)):
		return LiturgicalColor{"#f6bf26", "#000000", "Easter"}
	case d.Before(advent):
		return LiturgicalColor{"#33b679", "#ffffff", "Ordinary Time"}
	default:
		return LiturgicalColor{"#7986cb", "#ffffff", "Advent"}
	}
}

// ── ANSI colors ───────────────────────────────────────────────────────────────

const (
	reset  = "\033[0m"
	bold   = "\033[1m"
	dim    = "\033[2m"
	cyan   = "\033[36m"
	green  = "\033[32m"
	blue   = "\033[34m"
	yellow = "\033[33m"
	gray   = "\033[90m"
	red    = "\033[31m"
)

var shiftColor = map[string]string{
	"Pagi":        green,
	"Malam":       blue,
	"Lepas Malam": cyan,
	"OFF":         red,
}

// ── Work hours ────────────────────────────────────────────────────────────────

// countWeekdays returns the number of Mon–Fri days in the given month.
func countWeekdays(year int, month time.Month) int {
	count := 0
	for d := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC); d.Month() == month; d = d.AddDate(0, 0, 1) {
		w := d.Weekday()
		if w != time.Saturday && w != time.Sunday {
			count++
		}
	}
	return count
}

// ── Dry-run printer ───────────────────────────────────────────────────────────

func printDryRun(calName string, events []Event) {
	width := 54
	border := strings.Repeat("─", width)

	fmt.Printf("\n%s┌%s┐%s\n", bold, strings.Repeat("─", width), reset)
	fmt.Printf("%s│%s  📅 %-*s%s│%s\n", bold, cyan, width-3, calName, reset+bold, reset)
	fmt.Printf("%s│%s  👤 %-*s%s│%s\n", bold, dim, width-3, inviteEmail+" (writer)", reset+bold, reset)
	fmt.Printf("%s└%s┘%s\n\n", bold, strings.Repeat("─", width), reset)

	// Group by week
	var weekNum int
	for i, ev := range events {
		date, _ := time.Parse("2006-01-02", ev.Date)

		// Print week header when day-of-week resets to the first weekday of first event, or Mon
		isFirstEvent := i == 0
		isMonday := date.Weekday() == time.Monday
		if isFirstEvent || isMonday {
			weekNum++
			if i > 0 {
				fmt.Println()
			}
			fmt.Printf("%s  Week %-2d%s%s\n", bold+yellow, weekNum, reset, "")
			fmt.Printf("  %s\n", border)
		}

		color := shiftColor[ev.Summary]
		dayLabel := fmt.Sprintf("%s %02d %s", date.Weekday().String()[:3], date.Day(), date.Month().String()[:3])

		if ev.AllDay {
			fmt.Printf("  %s%-14s%s │ %s%-12s%s\n",
				dim, dayLabel, reset,
				color, ev.Summary, reset,
			)
		} else {
			alarm := ""
			if ev.WithAlarm {
				alarm = " 🔔"
			}
			timeStr := fmt.Sprintf("%s → %s", ev.StartTime, ev.EndTime)
			if ev.Date != ev.EndDate {
				timeStr += " +1"
			}
			fmt.Printf("  %s%-14s%s │ %s%-12s%s  %s%s\n",
				dim, dayLabel, reset,
				color+bold, ev.Summary, reset,
				timeStr, alarm,
			)
		}
	}

	// Summary
	counts := map[string]int{}
	for _, ev := range events {
		counts[ev.Summary]++
	}
	fmt.Printf("\n  %s\n", border)
	fmt.Printf("  %sShift Summary%s\n", bold, reset)
	for _, code := range []string{"Pagi", "Malam", "Lepas Malam", "OFF"} {
		color := shiftColor[code]
		fmt.Printf("  %s%-14s%s  %s%d shifts%s\n", color+bold, code, reset, dim, counts[code], reset)
	}

	// Work hours vs Indonesian normal (UU No. 13/2003: 40 hrs/week = weekdays × 8 hrs)
	pHours := counts["Pagi"] * 8
	mHours := counts["Malam"] * 16
	actualHours := pHours + mHours

	weekdays := countWeekdays(config.Year, config.Month)
	normalHours := weekdays * 8
	diff := actualHours - normalHours

	var statusStr string
	switch {
	case diff > 8:
		statusStr = fmt.Sprintf("%s▲ OVERWORK  +%d hrs%s", red+bold, diff, reset)
	case diff < -8:
		statusStr = fmt.Sprintf("%s▼ UNDERWORK  %d hrs%s", yellow+bold, diff, reset)
	default:
		statusStr = fmt.Sprintf("%s✓ ABOUT RIGHT  %+d hrs%s", green+bold, diff, reset)
	}

	fmt.Printf("\n  %sWork Hours%s\n", bold, reset)
	fmt.Printf("  %s%-14s%s  %s%d shifts × 8h  = %d hrs%s\n",
		green+bold, "Pagi", reset, dim, counts["Pagi"], pHours, reset)
	fmt.Printf("  %s%-14s%s  %s%d shifts × 16h = %d hrs%s\n",
		blue+bold, "Malam", reset, dim, counts["Malam"], mHours, reset)
	fmt.Printf("  %s%-14s%s  %s%d hrs%s\n",
		bold, "Total", reset, bold, actualHours, reset)
	fmt.Printf("  %s%-14s%s  %s%d weekdays × 8h = %d hrs  (UU No. 13/2003)%s\n",
		dim, "Normal", reset, dim, weekdays, normalHours, reset)
	fmt.Printf("\n  %s\n\n", statusStr)
}
