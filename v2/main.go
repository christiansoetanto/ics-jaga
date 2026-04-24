package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	var debug bool

	root := &cobra.Command{
		Use:   "jaga",
		Short: "Manage shift schedule on Google Calendar",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			logger.SetDebug(debug)
		},
	}
	root.PersistentFlags().BoolVar(&debug, "debug", true, "Enable debug logging")

	root.AddCommand(cmdDryRun(), cmdExecute(), cmdReset())

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func cmdDryRun() *cobra.Command {
	return &cobra.Command{
		Use:     "dry-run",
		Aliases: []string{"d"},
		Short:   "Print the schedule without touching Google Calendar",
		Run: func(cmd *cobra.Command, args []string) {
			logger.SetSilent(true)
			events := buildEvents()
			printDryRun(calendarName(), events)
		},
	}
}

func cmdExecute() *cobra.Command {
	return &cobra.Command{
		Use:     "execute",
		Aliases: []string{"e"},
		Short:   "Create the calendar and import events (fails if events already exist)",
		Run: func(cmd *cobra.Command, args []string) {
			logger.Infof("mode: execute")
			ctx := context.Background()
			srv := newCalendarService(ctx)
			events := buildEvents()
			runExecute(ctx, srv, calendarName(), events, false)
		},
	}
}

func cmdReset() *cobra.Command {
	return &cobra.Command{
		Use:     "reset",
		Aliases: []string{"r"},
		Short:   "Delete the calendar and recreate it from scratch (re-shares automatically)",
		Run: func(cmd *cobra.Command, args []string) {
			logger.Infof("mode: reset")
			ctx := context.Background()
			srv := newCalendarService(ctx)
			events := buildEvents()
			runExecute(ctx, srv, calendarName(), events, true)
		},
	}
}
