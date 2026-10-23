package ctok

import (
	"strings"
	"unicode/utf8"
)

// wordShift is the allocation-free V48 case decision. Lowering is per rune;
// dotted capital I stays literal, and capital sharp S blocks the entire span.
func wordShift(body string) bool {
	if body == "" || strings.ContainsRune(body, 'ẞ') {
		return false
	}
	head, size := utf8.DecodeRuneInString(body)
	if head == 'İ' || !isUpper(head) {
		return false
	}
	for _, r := range body[size:] {
		if r != 'İ' && isUpper(r) {
			return false
		}
	}
	return true
}

func appendWord(out []rune, body string, lowered bool) []rune {
	for _, r := range body {
		if lowered && r != 'İ' {
			if mapped, ok := lowerMappings[r]; ok {
				for _, lower := range mapped {
					out = append(out, lower)
				}
				continue
			}
		}
		out = append(out, r)
	}
	return out
}

// streamCapacity bounds output by input runes plus possible run markers. Full
// lowercase expands only dotted I in the pinned tables, which V48 keeps literal.
// Counting runs avoids byte-based over-allocation on long multibyte mark runs.
func streamCapacity(text string) int {
	capacity := utf8.RuneCountInString(text)
	reader := runReader{text: text}
	for run := reader.next(); run.body != ""; run = reader.next() {
		switch run.class {
		case classWordy:
			capacity += 3
		case classStray, classPunct, classHard, classDigit:
			capacity += 2
		}
	}
	return capacity
}
