package utils

import (
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
	"unicode"
)

func isNonSpacing(r rune) bool {
	return unicode.Is(unicode.Mn, r)
}

func RemoveDiacritics(text string) (string, error) {
	t := transform.Chain(norm.NFD, transform.RemoveFunc(isNonSpacing), norm.NFC)
	result, _, err := transform.String(t, text)
	if err != nil {
		return "", err
	}
	return result, nil
}
