package ctok

import "strings"

func lowerRune(r rune) string {
	if mapped, ok := lowerMappings[r]; ok {
		return mapped
	}
	return string(r)
}

func lowerV48(text string) string {
	var out strings.Builder
	out.Grow(len(text))
	for _, r := range text {
		out.WriteString(lowerRune(r))
	}
	return out.String()
}

func markCaseV48(text string) string {
	if strings.ContainsRune(text, 'ẞ') || text == "" {
		return text
	}
	head, tail := rune(0), ""
	for i, r := range text {
		head = r
		tail = text[i+len(string(r)):]
		break
	}
	if strings.ContainsRune(text, 'İ') {
		if !isUpper(head) || head == 'İ' {
			return text
		}
		for _, r := range tail {
			if r != 'İ' && isUpper(r) {
				return text
			}
		}
		var out strings.Builder
		out.WriteRune(shift)
		out.WriteString(lowerRune(head))
		for _, r := range tail {
			if r == 'İ' {
				out.WriteRune(r)
			} else {
				out.WriteString(lowerRune(r))
			}
		}
		return out.String()
	}
	if !isUpper(head) {
		return text
	}
	for _, r := range tail {
		if isUpper(r) {
			return text
		}
	}
	return string(shift) + lowerV48(text)
}
