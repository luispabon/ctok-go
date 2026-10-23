package ctok

import "sort"

type runeClass uint8

const (
	classWordy runeClass = iota
	classHard
	classDigit
	classPunct
	classSpace
)

type unicodeRange struct {
	lo, hi    rune
	class     runeClass
	flags     uint8
	combining uint8
	category  string
}

func properties(r rune) unicodeRange {
	i := sort.Search(len(unicodeRanges), func(i int) bool { return unicodeRanges[i].hi >= r })
	if i < len(unicodeRanges) && unicodeRanges[i].lo <= r {
		return unicodeRanges[i]
	}
	return unicodeRange{class: classHard, category: "Cn"}
}

func classifyRune(r rune) runeClass { return properties(r).class }
