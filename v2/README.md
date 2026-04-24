# jaga

Automates shift schedule management on Google Calendar.

## What it does

- Creates a calendar named `Jaga <Month> <Year>` (e.g. `Jaga May 2026`)
- Imports your shift schedule as events with correct times
- Shares the calendar with `nadiae.nurtanto@gmail.com` as owner
- Prevents duplicate imports — fails if events already exist unless you use `reset`

## One-time setup

1. Go to [Google Cloud Console](https://console.cloud.google.com)
2. Create a project → enable **Google Calendar API**
3. Go to **APIs & Services → Credentials → Create Credentials → OAuth 2.0 Client ID**
4. Choose **Desktop App** → download as `credentials.json`
5. Place `credentials.json` in `ics_jaga/v2/`

The first time you run `execute` or `reset`, a browser auth URL will be printed. Open it, authorize, paste the code back. A `token.json` is cached for future runs.

## Configuring the schedule

Edit the config block at the top of `schedule.go`:

```go
var config = struct { ... }{
    Year:      2026,
    Month:     time.May,
    FirstDay:  1,
    Shifts:    "Mg P Mg M LM Mg P P M LM ...",
    CredsFile: "credentials.json",
}
```

### Shift codes

| Code | Title        | Time              | Alarm |
|------|--------------|-------------------|-------|
| `P`  | Pagi         | 08:00 – 16:00     | 12h   |
| `M`  | Malam        | 16:00 – 08:00+1   | 12h   |
| `LM` | Lepas Malam  | 08:00 – 23:59     | —     |
| `Mg` | OFF          | all-day           | —     |

## Usage

```bash
# Build
go build -o jaga .

# Preview schedule without touching Google Calendar
./jaga dry-run

# Create calendar, share it, and import all events
# Fails if events already exist (overlap protection)
./jaga execute

# Delete the calendar and recreate from scratch
# Use this when you need to fix a wrong schedule
./jaga reset

# Enable verbose debug logging for any command
./jaga --debug dry-run
./jaga --debug execute
```

## Workflow for a new month

1. Update `schedule.go` with the new month, year, and shifts
2. Run `./jaga dry-run` to verify the schedule looks correct
3. Run `./jaga execute` to create and populate the calendar

## Workflow for fixing a wrong schedule

1. Fix the shifts in `schedule.go`
2. Run `./jaga dry-run` to verify
3. Run `./jaga reset` — deletes the old calendar and recreates it (sharing is re-applied automatically)
