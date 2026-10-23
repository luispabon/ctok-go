package ctok

import "unicode/utf8"

func isSelector(r rune) bool { return r >= 0xfe00 && r <= 0xfe0f }

func isSeparator(r rune) bool {
	p := properties(r)
	return r < 0x10000 && p.flags&2 != 0 && !isSelector(r) && p.flags&1 == 0
}

func strayMark(r rune) bool {
	if r == 0x0711 || r >= 0x0730 && r <= 0x073f {
		return false
	}
	return properties(r).combining != 0 && !isSeparator(r)
}

func marksLikePunct(r rune) bool {
	if r >= 0xe000 && r <= 0xe004 {
		r = bow + r - 0xe000
	}
	if r >= 0x10000 {
		return false
	}
	if isSeparator(r) {
		return true
	}
	category := properties(r).category
	if r >= 0x3001 && r <= 0x303f {
		if category[0] == 'P' {
			return false
		}
		switch r {
		case 0x3004, 0x3012, 0x3013, 0x3020, 0x3036, 0x3037, 0x303e, 0x303f:
			return false
		}
	}
	return category[0] == 'P' || category[0] == 'S' || category == "Cf" || category == "Cn"
}

func digitBorder(r rune) bool {
	category := properties(r).category
	return r < 0x10000 && (category == "No" || category == "Nd" && r >= 0x80)
}

func hardKind(r rune) byte {
	if marksLikePunct(r) {
		return 'p'
	}
	if digitBorder(r) {
		return 'n'
	}
	return 'l'
}

func digitRun(body string) bool {
	first, _ := utf8.DecodeRuneInString(body)
	category := properties(first).category
	if category != "Nd" && category != "No" {
		return false
	}
	for _, r := range body {
		if properties(r).category != category {
			return false
		}
	}
	return true
}

func hardBow(r rune) bool { return isSelector(r) || marksLikePunct(r) }

func takesRightBorder(run streamRun) bool {
	if run.body == "" {
		return false
	}
	last, _ := utf8.DecodeLastRuneInString(run.body)
	return run.class == classPunct || hardBow(last) ||
		(run.class == classDigit || run.class == classHard) && digitRun(run.body) && digitBorder(last)
}

func contractionSeam(current, previous, beforePrevious streamRun) bool {
	if previous.class != classPunct || previous.body != "'" {
		return false
	}
	switch current.body {
	case "s", "t", "d", "m", "ll", "re", "ve":
		return !takesRightBorder(beforePrevious)
	}
	return false
}
