package main

import (
	"context"
	"strings"

	"google.golang.org/api/calendar/v3"
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
		calID = createCalendar(srv, calName)
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

func createCalendar(srv *calendar.Service, name string) string {
	logger.Infof("creating calendar %q...", name)
	cal, err := srv.Calendars.Insert(&calendar.Calendar{
		Summary:  name,
		TimeZone: timeZone,
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
	logger.Infof("setting calendar color for %s season (%s)...", lc.Season, lc.BG)
	_, err := srv.CalendarList.Patch(calID, &calendar.CalendarListEntry{
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
		created, err := srv.Events.Insert(calID, gcalEv).Do()
		if err != nil {
			logger.Fatalf("insert event [%s] %s: %v", ev.Date, ev.Summary, err)
		}
		logger.Infof("(%d/%d) inserted: %s — event id: %s", i+1, len(events), ev, created.Id)
	}
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
