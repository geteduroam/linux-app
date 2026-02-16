// Package logwrap implements wrapper around loggers
package logwrap

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/geteduroam/linux-app/internal/config"
	"github.com/geteduroam/linux-app/internal/utilsx"
	"golang.org/x/exp/slog"
)

// Location returns the location of the log file
// and creates it if not exists
func Location(program string) (string, error) {
	logfile := fmt.Sprintf("%s.log", program)
	dir, err := config.Directory()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	return filepath.Join(dir, logfile), nil
}

func newLogFile(program string) (*os.File, string, error) {
	fpath, err := Location(program)
	if err != nil {
		return nil, "", err
	}
	fp, err := os.OpenFile(fpath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, "", err
	}
	return fp, fpath, nil
}

// Initialize creates the logger from the program name and whether or not to enable debug logging
// Logging is done to a file if possible, otherwise the console
func Initialize(program string, debug bool) {
	logLevel := &slog.LevelVar{}
	opts := &slog.HandlerOptions{
		Level: logLevel,
	}
	logfile, fpath, err := newLogFile(program)
	if err == nil {
		slog.SetDefault(slog.New(slog.NewTextHandler(logfile, opts)))
		if debug {
			fmt.Printf("Writing debug logs to %s\n", fpath)
		} else {
			utilsx.Verbosef("Writing logs to %s", fpath)
		}
	} else {
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, opts)))
		if debug {
			fmt.Println("Writing debug logs to console, due to error: ", err)
		} else {
			utilsx.Verbosef("Writing logs to console, due to error: ", err)
		}
	}
	if debug {
		logLevel.Set(slog.LevelDebug)
	}
}
