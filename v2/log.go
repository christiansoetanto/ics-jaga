package main

import (
	"fmt"
	"log"
	"os"
)

// Simple leveled logger wrapping the standard log package.

type level int

const (
	levelDebug level = iota
	levelInfo
	levelWarn
	levelError
)

type Logger struct {
	l        *log.Logger
	minLevel level
	silent   bool
}

var logger = &Logger{
	l:        log.New(os.Stderr, "", log.Ldate|log.Ltime|log.Lmsgprefix),
	minLevel: levelInfo,
}

func (lg *Logger) SetDebug(enabled bool) {
	if enabled {
		lg.minLevel = levelDebug
	} else {
		lg.minLevel = levelInfo
	}
}

func (lg *Logger) SetSilent(s bool) {
	lg.silent = s
}

func (lg *Logger) Debugf(format string, args ...any) {
	if !lg.silent && lg.minLevel <= levelDebug {
		lg.l.SetPrefix("[DEBUG] ")
		lg.l.Output(2, fmt.Sprintf(format, args...))
	}
}

func (lg *Logger) Infof(format string, args ...any) {
	if !lg.silent && lg.minLevel <= levelInfo {
		lg.l.SetPrefix("[INFO]  ")
		lg.l.Output(2, fmt.Sprintf(format, args...))
	}
}

func (lg *Logger) Warnf(format string, args ...any) {
	if !lg.silent && lg.minLevel <= levelWarn {
		lg.l.SetPrefix("[WARN]  ")
		lg.l.Output(2, fmt.Sprintf(format, args...))
	}
}

func (lg *Logger) Fatalf(format string, args ...any) {
	lg.l.SetPrefix("[FATAL] ")
	lg.l.Output(2, fmt.Sprintf(format, args...))
	os.Exit(1)
}
