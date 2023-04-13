package utils

import (
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
