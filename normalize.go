package ctok

import (
	"strings"

	"github.com/luispabon/ctok-go/internal/norm"
)

func normalizeV48(text string) string {
	text = norm.NFC.String(strings.ToValidUTF8(text, "\uFFFD"))
	text = strings.Map(func(r rune) rune {
		switch {
		case r == 0:
			return ' '
		case r >= 1 && r <= 8, r >= 0x0B && r <= 0x1F, r >= 0x7F && r <= 0x9F:
			return -1
		}
		return r
	}, text)
	text = strings.ReplaceAll(text, "\u0E4D\u0E32", "\u0E33")
	return strings.Map(func(r rune) rune {
		switch {
		case r >= 0xE000 && r <= 0xF8FF:
			return -1
		case r == 0x00A0, r == 0x1680, r >= 0x2000 && r <= 0x200A,
			r == 0x2028, r == 0x2029, r == 0x202F, r == 0x205F:
			return ' '
		}
		return r
	}, text)
}
