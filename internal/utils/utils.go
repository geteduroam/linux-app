package utils

import (
	"fmt"
	"unicode"

	"golang.org/x/exp/slog"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// RemoveDiacritics removes "diacritics" :^)
// Okay, diacritics are special characters, e.g. GÉANT, becomes GEANT
// This is useful when using it for substring matching
func RemoveDiacritics(text string) (string, error) {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	result, _, err := transform.String(t, text)
	if err != nil {
		return "", err
	}
	return result, nil
}

// ErrorString returns an error message for an error
// If the error is nil it returns the empty string
func ErrorString(e error) string {
	if e == nil {
		return ""
	}
	return e.Error()
}

var IsVerbose bool

// Conditionally (format) print verbose messages
func Verbosef(msg string, args ...any) {
	if IsVerbose {
		fmt.Printf(msg+"\n", args...)
	}
}

// Testfunction to test logLevel setting
func PrintLevels() {
	msg := "Test"
	Verbosef("Verbose: %s", msg)
	slog.Debug("Debug", "debug", msg)
	slog.Info("Info", "info", msg)
	slog.Warn("Warn", "warn", msg)
	slog.Error("Error", "error", msg)
}
