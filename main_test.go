package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"
)

// Schedule: Mg P Mg M LM Mg P P M LM Mg P P Mg Mg P M LM Mg Mg P Mg P M LM Mg M LM Mg M LM
// May 2026 (31 days)
const expectedICS = `BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//Shift Calendar//EN
BEGIN:VEVENT
SUMMARY:OFF
DTSTART;VALUE=DATE:20260501
DTEND;VALUE=DATE:20260502
END:VEVENT
BEGIN:VEVENT
SUMMARY:Pagi
DTSTART:20260502T080000
DTEND:20260502T160000
BEGIN:VALARM
TRIGGER:-PT12H
ACTION:DISPLAY
DESCRIPTION:Reminder
END:VALARM
END:VEVENT
BEGIN:VEVENT
SUMMARY:OFF
DTSTART;VALUE=DATE:20260503
DTEND;VALUE=DATE:20260504
END:VEVENT
BEGIN:VEVENT
SUMMARY:Malam
DTSTART:20260504T160000
DTEND:20260505T080000
BEGIN:VALARM
TRIGGER:-PT12H
ACTION:DISPLAY
DESCRIPTION:Reminder
END:VALARM
END:VEVENT
BEGIN:VEVENT
SUMMARY:Lepas Malam
DTSTART:20260505T080000
DTEND:20260505T235900
END:VEVENT
BEGIN:VEVENT
SUMMARY:OFF
DTSTART;VALUE=DATE:20260506
DTEND;VALUE=DATE:20260507
END:VEVENT
BEGIN:VEVENT
SUMMARY:Pagi
DTSTART:20260507T080000
DTEND:20260507T160000
BEGIN:VALARM
TRIGGER:-PT12H
ACTION:DISPLAY
DESCRIPTION:Reminder
END:VALARM
END:VEVENT
BEGIN:VEVENT
SUMMARY:Pagi
DTSTART:20260508T080000
DTEND:20260508T160000
BEGIN:VALARM
TRIGGER:-PT12H
ACTION:DISPLAY
DESCRIPTION:Reminder
END:VALARM
END:VEVENT
BEGIN:VEVENT
SUMMARY:Malam
DTSTART:20260509T160000
DTEND:20260510T080000
BEGIN:VALARM
TRIGGER:-PT12H
ACTION:DISPLAY
DESCRIPTION:Reminder
END:VALARM
END:VEVENT
BEGIN:VEVENT
SUMMARY:Lepas Malam
DTSTART:20260510T080000
DTEND:20260510T235900
END:VEVENT
BEGIN:VEVENT
SUMMARY:OFF
DTSTART;VALUE=DATE:20260511
DTEND;VALUE=DATE:20260512
END:VEVENT
BEGIN:VEVENT
SUMMARY:Pagi
DTSTART:20260512T080000
DTEND:20260512T160000
BEGIN:VALARM
TRIGGER:-PT12H
ACTION:DISPLAY
DESCRIPTION:Reminder
END:VALARM
END:VEVENT
BEGIN:VEVENT
SUMMARY:Pagi
DTSTART:20260513T080000
DTEND:20260513T160000
BEGIN:VALARM
TRIGGER:-PT12H
ACTION:DISPLAY
DESCRIPTION:Reminder
END:VALARM
END:VEVENT
BEGIN:VEVENT
SUMMARY:OFF
DTSTART;VALUE=DATE:20260514
DTEND;VALUE=DATE:20260515
END:VEVENT
BEGIN:VEVENT
SUMMARY:OFF
DTSTART;VALUE=DATE:20260515
DTEND;VALUE=DATE:20260516
END:VEVENT
BEGIN:VEVENT
SUMMARY:Pagi
DTSTART:20260516T080000
DTEND:20260516T160000
BEGIN:VALARM
TRIGGER:-PT12H
ACTION:DISPLAY
DESCRIPTION:Reminder
END:VALARM
END:VEVENT
BEGIN:VEVENT
SUMMARY:Malam
DTSTART:20260517T160000
DTEND:20260518T080000
BEGIN:VALARM
TRIGGER:-PT12H
ACTION:DISPLAY
DESCRIPTION:Reminder
END:VALARM
END:VEVENT
BEGIN:VEVENT
SUMMARY:Lepas Malam
DTSTART:20260518T080000
DTEND:20260518T235900
END:VEVENT
BEGIN:VEVENT
SUMMARY:OFF
DTSTART;VALUE=DATE:20260519
DTEND;VALUE=DATE:20260520
END:VEVENT
BEGIN:VEVENT
SUMMARY:OFF
DTSTART;VALUE=DATE:20260520
DTEND;VALUE=DATE:20260521
END:VEVENT
BEGIN:VEVENT
SUMMARY:Pagi
DTSTART:20260521T080000
DTEND:20260521T160000
BEGIN:VALARM
TRIGGER:-PT12H
ACTION:DISPLAY
DESCRIPTION:Reminder
END:VALARM
END:VEVENT
BEGIN:VEVENT
SUMMARY:OFF
DTSTART;VALUE=DATE:20260522
DTEND;VALUE=DATE:20260523
END:VEVENT
BEGIN:VEVENT
SUMMARY:Pagi
DTSTART:20260523T080000
DTEND:20260523T160000
BEGIN:VALARM
TRIGGER:-PT12H
ACTION:DISPLAY
DESCRIPTION:Reminder
END:VALARM
END:VEVENT
BEGIN:VEVENT
SUMMARY:Malam
DTSTART:20260524T160000
DTEND:20260525T080000
BEGIN:VALARM
TRIGGER:-PT12H
ACTION:DISPLAY
DESCRIPTION:Reminder
END:VALARM
END:VEVENT
BEGIN:VEVENT
SUMMARY:Lepas Malam
DTSTART:20260525T080000
DTEND:20260525T235900
END:VEVENT
BEGIN:VEVENT
SUMMARY:OFF
DTSTART;VALUE=DATE:20260526
DTEND;VALUE=DATE:20260527
END:VEVENT
BEGIN:VEVENT
SUMMARY:Malam
DTSTART:20260527T160000
DTEND:20260528T080000
BEGIN:VALARM
TRIGGER:-PT12H
ACTION:DISPLAY
DESCRIPTION:Reminder
END:VALARM
END:VEVENT
BEGIN:VEVENT
SUMMARY:Lepas Malam
DTSTART:20260528T080000
DTEND:20260528T235900
END:VEVENT
BEGIN:VEVENT
SUMMARY:OFF
DTSTART;VALUE=DATE:20260529
DTEND;VALUE=DATE:20260530
END:VEVENT
BEGIN:VEVENT
SUMMARY:Malam
DTSTART:20260530T160000
DTEND:20260531T080000
BEGIN:VALARM
TRIGGER:-PT12H
ACTION:DISPLAY
DESCRIPTION:Reminder
END:VALARM
END:VEVENT
BEGIN:VEVENT
SUMMARY:Lepas Malam
DTSTART:20260531T080000
DTEND:20260531T235900
END:VEVENT
END:VCALENDAR
`

func TestRun_FullMaySchedule(t *testing.T) {
	out := "test_schedule.ics"
	defer os.Remove(out)

	stats := run(2026, 5, 1,
		"Mg P Mg M LM Mg P P M LM Mg P P Mg Mg P M LM Mg Mg P Mg P M LM Mg M LM Mg M LM",
		out,
	)

	// Check stats
	if stats.P != 8 {
		t.Errorf("P count = %d, want 8", stats.P)
	}
	if stats.M != 6 {
		t.Errorf("M count = %d, want 6", stats.M)
	}
	if stats.LM != 6 {
		t.Errorf("LM count = %d, want 6", stats.LM)
	}
	if stats.Mg != 11 {
		t.Errorf("Mg count = %d, want 11", stats.Mg)
	}

	// Check ICS content
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}
	got := string(data)

	if got != expectedICS {
		gotLines := strings.Split(got, "\n")
		wantLines := strings.Split(expectedICS, "\n")
		for i := 0; i < len(gotLines) || i < len(wantLines); i++ {
			g, w := "", ""
			if i < len(gotLines) {
				g = gotLines[i]
			}
			if i < len(wantLines) {
				w = wantLines[i]
			}
			if g != w {
				t.Errorf("line %d:\n  got:  %q\n  want: %q", i+1, g, w)
			}
		}
	}
}

func TestGenerateDates(t *testing.T) {
	tests := []struct {
		name     string
		year     int
		month    int
		firstDay int
		count    int
		want     []string
	}{
		{
			name:     "first 5 days of May",
			year:     2026,
			month:    5,
			firstDay: 1,
			count:    5,
			want:     []string{"01/05", "02/05", "03/05", "04/05", "05/05"},
		},
		{
			name:     "single day",
			year:     2026,
			month:    5,
			firstDay: 15,
			count:    1,
			want:     []string{"15/05"},
		},
		{
			name:     "full May schedule",
			year:     2026,
			month:    5,
			firstDay: 1,
			count:    31,
			want: func() []string {
				s := make([]string, 31)
				for i := range s {
					s[i] = fmt.Sprintf("%02d/05", i+1)
				}
				return s
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateDates(tt.year, tt.month, tt.firstDay, tt.count)
			if len(got) != len(tt.want) {
				t.Fatalf("len = %d, want %d", len(got), len(tt.want))
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("dates[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestComputeDatePair(t *testing.T) {
	tests := []struct {
		name         string
		year         int
		dateStr      string
		wantDate     string
		wantNextDate string
	}{
		{
			name:         "mid month",
			year:         2026,
			dateStr:      "15/05",
			wantDate:     "20260515",
			wantNextDate: "20260516",
		},
		{
			name:         "last day of May rolls over to June",
			year:         2026,
			dateStr:      "31/05",
			wantDate:     "20260531",
			wantNextDate: "20260601",
		},
		{
			name:         "last day of April rolls over to May",
			year:         2026,
			dateStr:      "30/04",
			wantDate:     "20260430",
			wantNextDate: "20260501",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			date, nextDate := computeDatePair(tt.year, tt.dateStr)
			if date != tt.wantDate {
				t.Errorf("date = %s, want %s", date, tt.wantDate)
			}
			if nextDate != tt.wantNextDate {
				t.Errorf("nextDate = %s, want %s", nextDate, tt.wantNextDate)
			}
		})
	}
}

func TestWriteEvent_PagiShift(t *testing.T) {
	shift := Shift{Title: "Pagi", Start: "08:00", End: "16:00", NextDay: false, WithAlarm: true}
	content := captureWriteEvent(t, shift, "20260501", "20260502")

	assertContains(t, content, "BEGIN:VEVENT")
	assertContains(t, content, "SUMMARY:Pagi")
	assertContains(t, content, "DTSTART:20260501T080000")
	assertContains(t, content, "DTEND:20260501T160000")
	assertContains(t, content, "BEGIN:VALARM")
	assertContains(t, content, "TRIGGER:-PT12H")
	assertContains(t, content, "END:VALARM")
	assertContains(t, content, "END:VEVENT")
}

func TestWriteEvent_MalamShiftEndsNextDay(t *testing.T) {
	shift := Shift{Title: "Malam", Start: "16:00", End: "08:00", NextDay: true, WithAlarm: true}
	content := captureWriteEvent(t, shift, "20260504", "20260505")

	assertContains(t, content, "DTSTART:20260504T160000")
	assertContains(t, content, "DTEND:20260505T080000")
	if strings.Contains(content, "DTEND:20260504T080000") {
		t.Error("DTEND should be next day (20260505), not same day (20260504)")
	}
}

func TestWriteEvent_OffShiftAllDay(t *testing.T) {
	shift := Shift{Title: "OFF", Start: "", End: "", NextDay: false, WithAlarm: false}
	content := captureWriteEvent(t, shift, "20260501", "20260502")

	assertContains(t, content, "DTSTART;VALUE=DATE:20260501")
	assertContains(t, content, "DTEND;VALUE=DATE:20260502")
	if strings.Contains(content, "BEGIN:VALARM") {
		t.Error("OFF shift should not have VALARM")
	}
}

func TestWriteEvent_LepasMailam(t *testing.T) {
	shift := Shift{Title: "Lepas Malam", Start: "08:00", End: "23:59", NextDay: false, WithAlarm: false}
	content := captureWriteEvent(t, shift, "20260502", "20260503")

	assertContains(t, content, "DTSTART:20260502T080000")
	assertContains(t, content, "DTEND:20260502T235900")
	if strings.Contains(content, "VALUE=DATE") {
		t.Error("Lepas Malam should be a timed event, not all-day")
	}
}

func captureWriteEvent(t *testing.T, shift Shift, date, nextDay string) string {
	t.Helper()
	f, err := os.CreateTemp("", "test_ics_*.ics")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	writeEvent(f, shift, date, nextDay)

	f.Seek(0, 0)
	var buf bytes.Buffer
	buf.ReadFrom(f)
	return buf.String()
}

func assertContains(t *testing.T, content, substr string) {
	t.Helper()
	if !strings.Contains(content, substr) {
		t.Errorf("expected output to contain %q\ngot:\n%s", substr, content)
	}
}
