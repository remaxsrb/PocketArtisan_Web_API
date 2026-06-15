package utils

import (
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// SerbianCyrillicToLatin maps Cyrillic characters to their Latin counterparts
func SerbianCyrillicToLatin(s string) string {
	mapping := map[rune]string{
		'а': "a", 'б': "b", 'в': "v", 'г': "g", 'д': "d", 'ђ': "dj",
		'е': "e", 'ж': "z", 'з': "z", 'и': "i", 'ј': "j", 'к': "k",
		'л': "l", 'љ': "lj", 'м': "m", 'н': "n", 'њ': "nj", 'о': "o",
		'п': "p", 'р': "r", 'с': "s", 'т': "t", 'ћ': "c", 'у': "u",
		'ф': "f", 'х': "h", 'ц': "c", 'ч': "c", 'џ': "dz", 'ш': "s",
	}

	var b strings.Builder
	for _, r := range s {
		lowerR := unicode.ToLower(r)
		if val, ok := mapping[lowerR]; ok {
			b.WriteString(val)
		} else {
			b.WriteRune(lowerR)
		}
	}
	return b.String()
}

// NormalizeForSearch converts "Ćirilica" or "Ћирилица" to "cirilica"
func NormalizeForSearch(s string) string {
	// 1. Handle Cyrillic first
	s = SerbianCyrillicToLatin(s)

	// 2. Remove diacritics (e.g., ć -> c)
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	result, _, _ := transform.String(t, s)

	return strings.ToLower(result)
}
