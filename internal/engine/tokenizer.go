package engine

import (
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

var nonAlphaNum = regexp.MustCompile(`[^\p{L}\p{N}\s]`)

func tokenize(text string) []string {
	text = strings.ToLower(text)
	text = removeDiacritics(text)
	text = nonAlphaNum.ReplaceAllString(text, " ")
	tokens := strings.Fields(text)
	return tokens
}

func removeDiacritics(text string) string {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	result, _, _ := transform.String(t, text)
	return result
}
