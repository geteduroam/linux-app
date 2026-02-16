// Package utilsx implements various utility functions
// the suffix x is needed to make revive linter quiet for a useless package name
// the obvious TODO is then:
// TODO: make this package indeed obsolete
package utilsx

import (
	"fmt"
	"math"
	"time"
	"unicode"

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

// IsVerbose returns true if messages should be logged verbosed
var IsVerbose bool

// Verbosef conditionally (format) print verbose messages
func Verbosef(msg string, args ...any) {
	if IsVerbose {
		fmt.Printf(msg+"\n", args...)
	}
}

// ValidityDays returns the amount of days left for the validity timestamp
func ValidityDays(v time.Time) int {
	now := time.Now()
	if now.After(v) {
		return 0
	}
	days := v.Sub(now).Hours() / 24
	return int(math.Ceil(days))
}

// DeltaTime gives a human readable output for a time difference
// markb and marke mark the beginning and end markers, e.g. bold text
func DeltaTime(d time.Duration, markb string, marke string) string {
	n := int(d.Seconds())
	mins := n / 60
	secs := n % 60

	minText := "minutes"
	secText := "seconds"
	if mins == 1 {
		minText = "minute"
	}
	if secs == 1 {
		secText = "second"
	}

	switch {
	case mins > 0 && secs > 0:
		return fmt.Sprintf("%s%d%s %s and %s%d%s %s", markb, mins, marke, minText, markb, secs, marke, secText)

	case mins > 0:
		return fmt.Sprintf("%s%d%s %s", markb, mins, marke, minText)
	default:
		return fmt.Sprintf("%s%d%s %s", markb, secs, marke, secText)
	}
}
