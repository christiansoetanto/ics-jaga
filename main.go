package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

type Shift struct {
	Title     string
	Start     string
	End       string
	NextDay   bool
	WithAlarm bool
}

func generateDates(year int, month int, firstDay int, count int) []string {
	dates := make([]string, count)
	for i := 0; i < count; i++ {
		day := firstDay + i
		dates[i] = fmt.Sprintf("%02d/%02d", day, month)
	}
	return dates
}

// computeDatePair returns (date, nextDate) as "YYYYMMDD" strings for a given dateStr "DD/MM".
func computeDatePair(year int, dateStr string) (string, string) {
	t, _ := time.Parse("02/01", dateStr)
	date := fmt.Sprintf("%d%02d%02d", year, t.Month(), t.Day())
	next := t.Add(24 * time.Hour)
	nextDate := fmt.Sprintf("%d%02d%02d", year, next.Month(), next.Day())
	return date, nextDate
}

func writeEvent(f *os.File, shift Shift, date string, nextDay string) {
	if shift.Start != "" {
		start := date + "T" + strings.ReplaceAll(shift.Start, ":", "") + "00"
		end := date
		if shift.NextDay {
			t, _ := time.Parse("20060102", date)
			end = t.Add(24 * time.Hour).Format("20060102")
		}
		end += "T" + strings.ReplaceAll(shift.End, ":", "") + "00"
		fmt.Fprintln(f, "BEGIN:VEVENT")
		fmt.Fprintf(f, "SUMMARY:%s\n", shift.Title)
		fmt.Fprintf(f, "DTSTART:%s\n", start)
		fmt.Fprintf(f, "DTEND:%s\n", end)
		if shift.WithAlarm {
			fmt.Fprintln(f, "BEGIN:VALARM")
			fmt.Fprintln(f, "TRIGGER:-PT12H")
			fmt.Fprintln(f, "ACTION:DISPLAY")
			fmt.Fprintln(f, "DESCRIPTION:Reminder")
			fmt.Fprintln(f, "END:VALARM")
		}
		fmt.Fprintln(f, "END:VEVENT")
	} else {
		fmt.Fprintln(f, "BEGIN:VEVENT")
		fmt.Fprintf(f, "SUMMARY:%s\n", shift.Title)
		fmt.Fprintf(f, "DTSTART;VALUE=DATE:%s\n", date)
		fmt.Fprintf(f, "DTEND;VALUE=DATE:%s\n", nextDay)
		if shift.WithAlarm {
			fmt.Fprintln(f, "BEGIN:VALARM")
			fmt.Fprintln(f, "TRIGGER:-PT12H")
			fmt.Fprintln(f, "ACTION:DISPLAY")
			fmt.Fprintln(f, "DESCRIPTION:Reminder")
			fmt.Fprintln(f, "END:VALARM")
		}
		fmt.Fprintln(f, "END:VEVENT")
	}
}

type Stats struct {
	P  int `json:"p"`
	M  int `json:"m"`
	Mg int `json:"mg"`
	LM int `json:"lm"`
}

func main() {
	run(2026, 5, 1, "Mg P Mg M LM Mg P P M LM Mg P P Mg Mg P M LM Mg Mg P Mg P M LM Mg M LM Mg M LM", "schedule.ics")
}

func run(year, month, firstDay int, shiftsStr, outputPath string) Stats {
	shifts := strings.Split(shiftsStr, " ")

	shiftMap := map[string]Shift{
		"P":  {Title: "Pagi", Start: "08:00", End: "16:00", NextDay: false, WithAlarm: true},
		"M":  {Title: "Malam", Start: "16:00", End: "08:00", NextDay: true, WithAlarm: true},
		"LM": {Title: "Lepas Malam", Start: "08:00", End: "23:59", NextDay: false, WithAlarm: false},
		"Mg": {Title: "OFF", Start: "", End: "", NextDay: false, WithAlarm: false},
	}

	dates := generateDates(year, month, firstDay, len(shifts))

	f, err := os.Create(outputPath)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	fmt.Fprintln(f, "BEGIN:VCALENDAR")
	fmt.Fprintln(f, "VERSION:2.0")
	fmt.Fprintln(f, "PRODID:-//Shift Calendar//EN")

	var stats Stats
	for i, shiftCode := range shifts {
		shift := shiftMap[shiftCode]
		date, nextDate := computeDatePair(year, dates[i])
		writeEvent(f, shift, date, nextDate)

		switch shiftCode {
		case "P":
			stats.P++
		case "M":
			stats.M++
		case "LM":
			stats.LM++
		case "Mg":
			stats.Mg++
		}
	}

	fmt.Fprintln(f, "END:VCALENDAR")

	_stats, _ := json.MarshalIndent(stats, "", "\t")
	fmt.Printf("%s\n", string(_stats))
	return stats
}
