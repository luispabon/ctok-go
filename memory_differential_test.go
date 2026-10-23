package ctok

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/luispabon/ctok-go/internal/testcorpus"
)

// fullCountMarked preserves the reviewed step-2 whole-input DP as a test oracle.
func fullCountMarked(v *vocabulary, stream []rune) int {
	best := make([]int, len(stream)+1)
	for end := 1; end <= len(stream); end++ {
		r := stream[end-1]
		node, ok := v.child(0, r)
		floor := 1
		if !structuralMarker(r) && (!ok || !v.nodes[node].terminal) {
			floor = v.byteFloor(r)
		}
		best[end] = best[end-1] + floor
		if !ok {
			continue
		}
		for start := end - 2; start >= 0; start-- {
			node, ok = v.child(node, stream[start])
			if !ok {
				break
			}
			if v.nodes[node].terminal && best[start]+1 < best[end] {
				best[end] = best[start] + 1
			}
		}
	}
	return best[len(stream)]
}

func TestTileRingDifferential(t *testing.T) {
	v := testVocabulary(t)
	if v.maxDepth < 32 {
		t.Fatalf("unexpected trie depth%d", v.maxDepth)
	}
	for i := range 5000 {
		stream, _ := streamV48(normalizeV48(testcorpus.Case(i)))
		if got, want := v.countMarked(stream), fullCountMarked(v, stream); got != want {
			t.Fatalf("ring case%d got%d want%d", i, got, want)
		}
	}
	for _, pieces := range [][]string{nil, {"a"}, {"a", "ab", "bc", "abcabc"}, {strings.Repeat("a", 97), "aa", "b"}} {
		toy := buildVocabulary(pieces, nil)
		for _, text := range []string{"", "abcabcab", strings.Repeat("a", 1000), "a😀b\ufdd0"} {
			if got, want := toy.countMarked([]rune(text)), fullCountMarked(toy, []rune(text)); got != want {
				t.Fatalf("toy ring got%d want%d", got, want)
			}
		}
	}
}

func TestStreamCaseDifferential(t *testing.T) {
	check := func(text string) {
		t.Helper()
		shifted := wordShift(text)
		var got []rune
		if shifted {
			got = append(got, shift)
		}
		got = appendWord(got, text, shifted)
		if want := markCaseV48(text); string(got) != want {
			t.Fatalf("case%q got%q want%q", text, string(got), want)
		}
	}
	for r := rune(0); r <= utf8.MaxRune; r++ {
		if r >= 0xd800 && r <= 0xdfff {
			continue
		}
		if properties(r).flags&12 != 0 {
			for _, text := range []string{string(r), string(r) + "abc", "A" + string(r), "Aİ" + string(r), string(r) + "ẞ"} {
				check(text)
			}
		}
	}
	for _, text := range []string{"", "İ", "Aİbc", "AİB", "ẞabc", "ϴabc", "\u1c89abc", "\ua7cbabc", "aİ", "Aa", "AA"} {
		check(text)
	}
	for r, mapped := range lowerMappings {
		if utf8.RuneCountInString(mapped) > 1 && r != 'İ' {
			t.Fatalf("capacity bound needs expanding lowercase U+%04X", r)
		}
	}
}

func TestStreamCapacity(t *testing.T) {
	for i := range 5000 {
		normalized := normalizeV48(testcorpus.Case(i))
		stream, _ := streamV48(normalized)
		if len(stream) > streamCapacity(normalized) {
			t.Fatalf("capacity bound case%d", i)
		}
	}
}
