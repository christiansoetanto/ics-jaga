package main

import (
	"context"
	"errors"
	"math/rand/v2"
	"strings"
	"time"

	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/googleapi"
)

const (
	inviteEmail = "nadiae.nurtanto@gmail.com"
	timeZone    = "Asia/Jakarta"
)

func runExecute(ctx context.Context, srv *calendar.Service, calName string, events []Event, reset bool) {
	calID := findCalendar(srv, calName)

	if reset && calID != "" {
		logger.Infof("reset: deleting calendar %q (id: %s)", calName, calID)
		if err := srv.Calendars.Delete(calID).Do(); err != nil {
			logger.Fatalf("delete calendar: %v", err)
		}
		logger.Infof("calendar deleted")
		calID = ""
	}

	if calID == "" {
		calID = createCalendar(srv, calName, workHoursSummary(events))
		shareCalendar(srv, calID)
		setCalendarColor(srv, calID)
	} else {
		logger.Infof("calendar %q already exists (id: %s), checking for existing events...", calName, calID)
		guardNoOverlap(srv, calID, calName)
	}

	insertEvents(srv, calID, events)
	logger.Infof("all done — https://calendar.google.com/calendar/r?cid=%s", calID)
}

func findCalendar(srv *calendar.Service, name string) string {
	logger.Infof("looking for calendar %q...", name)
	list, err := srv.CalendarList.List().Do()
	if err != nil {
		logger.Fatalf("list calendars: %v", err)
	}

	var matches []string
	for _, c := range list.Items {
		logger.Debugf("found calendar: %q (id: %s)", c.Summary, c.Id)
		if c.Summary == name {
			matches = append(matches, c.Id)
		}
	}

	switch len(matches) {
	case 0:
		logger.Infof("calendar not found")
		return ""
	case 1:
		logger.Infof("found calendar (id: %s)", matches[0])
		return matches[0]
	default:
		logger.Fatalf(
			"found %d calendars named %q — please delete the duplicates manually and retry:\n%s",
			len(matches), name, strings.Join(matches, "\n"),
		)
		return ""
	}
}

func createCalendar(srv *calendar.Service, name, description string) string {
	logger.Infof("creating calendar %q...", name)
	cal, err := srv.Calendars.Insert(&calendar.Calendar{
		Summary:     name,
		Description: description,
		TimeZone:    timeZone,
	}).Do()
	if err != nil {
		logger.Fatalf("create calendar: %v", err)
	}
	logger.Infof("calendar created (id: %s)", cal.Id)
	return cal.Id
}

func shareCalendar(srv *calendar.Service, calID string) {
	logger.Infof("sharing calendar with %s as writer...", inviteEmail)
	rule, err := srv.Acl.Insert(calID, &calendar.AclRule{
		Scope: &calendar.AclRuleScope{Type: "user", Value: inviteEmail},
		Role:  "writer",
	}).Do()
	if err != nil {
		logger.Fatalf("share calendar: %v", err)
	}
	logger.Infof("shared — acl rule id: %s", rule.Id)
}

func setCalendarColor(srv *calendar.Service, calID string) {
	lc := liturgicalColor(config.Year, config.Month)
	logger.Infof("setting calendar color for %s (web: %s, mobile colorId: %s)...", lc.Season, lc.BG, lc.ColorID)
	// Send both: web renders BG hex, mobile renders colorId.
	_, err := srv.CalendarList.Patch(calID, &calendar.CalendarListEntry{
		ColorId:         lc.ColorID,
		BackgroundColor: lc.BG,
		ForegroundColor: lc.FG,
	}).ColorRgbFormat(true).Do()
	if err != nil {
		logger.Warnf("set calendar color: %v", err)
	}
}

func guardNoOverlap(srv *calendar.Service, calID, calName string) {
	existing, err := srv.Events.List(calID).MaxResults(1).Do()
	if err != nil {
		logger.Fatalf("check existing events: %v", err)
	}
	if len(existing.Items) > 0 {
		logger.Fatalf(
			"calendar %q already has events — use 'reset' command to wipe and repopulate",
			calName,
		)
	}
	logger.Infof("no existing events, safe to insert")
}

func insertEvents(srv *calendar.Service, calID string, events []Event) {
	logger.Infof("inserting %d events...", len(events))
	for i, ev := range events {
		gcalEv := toGCalEvent(ev)
		created, err := insertWithRetry(srv, calID, gcalEv)
		if err != nil {
			logger.Fatalf("insert event [%s] %s: %v", ev.Date, ev.Summary, err)
		}
		logger.Infof("(%d/%d) inserted: %s — event id: %s", i+1, len(events), ev, created.Id)
	}
}

// insertWithRetry inserts a single calendar event with exponential backoff.
// It retries up to maxAttempts times on transient errors (429, 5xx, network).
// Permanent errors (400, 401, 403, 404) are returned immediately.
func insertWithRetry(srv *calendar.Service, calID string, ev *calendar.Event) (*calendar.Event, error) {
	const maxAttempts = 5
	backoff := time.Second

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		created, err := srv.Events.Insert(calID, ev).Do()
		if err == nil {
			return created, nil
		}

		if !isRetriable(err) || attempt == maxAttempts {
			return nil, err
		}

		// Add up to 500ms of random jitter so parallel callers don't all
		// retry at the exact same moment (thundering herd).
		jitter := time.Duration(rand.N(500)) * time.Millisecond
		sleep := backoff + jitter
		logger.Warnf("attempt %d/%d failed (%v) — retrying in %s...", attempt, maxAttempts, err, sleep.Round(time.Millisecond))
		time.Sleep(sleep)
		backoff *= 2 // 1s → 2s → 4s → 8s
	}

	// unreachable, but satisfies the compiler
	return nil, errors.New("exceeded max retry attempts")
}

// isRetriable reports whether err is worth retrying.
// Google API errors with 429 or 5xx are transient; anything else (400, 401,
// 403, 404) is a permanent failure that a retry won't fix.
func isRetriable(err error) bool {
	var apiErr *googleapi.Error
	if errors.As(err, &apiErr) {
		return apiErr.Code == 429 || apiErr.Code >= 500
	}
	// Non-API errors (network timeouts, DNS, etc.) are worth retrying.
	return true
}

func toGCalEvent(ev Event) *calendar.Event {
	gcalEv := &calendar.Event{Summary: ev.Summary}

	if ev.AllDay {
		gcalEv.Start = &calendar.EventDateTime{Date: ev.Date}
		gcalEv.End = &calendar.EventDateTime{Date: ev.EndDate}
	} else {
		gcalEv.Start = &calendar.EventDateTime{
			DateTime: ev.Date + "T" + ev.StartTime + ":00",
			TimeZone: timeZone,
		}
		gcalEv.End = &calendar.EventDateTime{
			DateTime: ev.EndDate + "T" + ev.EndTime + ":00",
			TimeZone: timeZone,
		}
	}

	if ev.WithAlarm {
		gcalEv.Reminders = &calendar.EventReminders{
			UseDefault:      false,
			Overrides:       []*calendar.EventReminder{{Method: "popup", Minutes: 720}}, // 12h before
			ForceSendFields: []string{"UseDefault"},
		}
	} else {
		gcalEv.Reminders = &calendar.EventReminders{
			UseDefault:      false,
			ForceSendFields: []string{"UseDefault"},
		}
	}

	return gcalEv
}
