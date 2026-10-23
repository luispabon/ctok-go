package norm

import (
	"math/rand/v2"
	"slices"
	"sort"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestPackedEntry(t *testing.T) {
	for r := rune(0); r <= utf8.MaxRune; r++ {
		entry := segRune{r: r, ccc: uint8(r), backward: r&1 != 0}
		if got := unpackEntry(packEntry(entry)); got != entry {
			t.Fatalf("packed U+%04X differs", r)
		}
	}
	for ccc := range 256 {
		for _, backward := range []bool{false, true} {
			entry := segRune{r: utf8.MaxRune, ccc: uint8(ccc), backward: backward}
			if got := unpackEntry(packEntry(entry)); got != entry {
				t.Fatal("packed flags differ")
			}
		}
	}
}

func TestOrderStable(t *testing.T) {
	rng := rand.New(rand.NewPCG(931, 3))
	for range 100 {
		entries := make([]segRune, 200)
		for i := range entries {
			entries[i] = segRune{r: rune(0x10000 + i*4000), ccc: uint8(rng.IntN(16)), backward: rng.IntN(2) == 0}
		}
		want := slices.Clone(entries)
		for start := 0; start < len(want); {
			if want[start].ccc == 0 {
				start++
				continue
			}
			end := start + 1
			for end < len(want) && want[end].ccc != 0 {
				end++
			}
			sort.SliceStable(want[start:end], func(i, j int) bool { return want[start+i].ccc < want[start+j].ccc })
			start = end
		}
		rb := reorderBuffer{rune: entries, nrune: len(entries)}
		rb.order()
		if !slices.Equal(rb.rune, want) {
			t.Fatal("packed stable counting sort differs")
		}
	}
}

func TestMeasureStorage(t *testing.T) {
	for _, tt := range []struct {
		input              string
		entries, unordered int
	}{
		{strings.Repeat("\u0344", 1000), 2000, 0},
		{strings.Repeat("\u0f73", 1000), 2000, 2000},
		{strings.Repeat("\u0344", 1000) + "\u0f73", 2002, 2002},
		{"\u1100\u1161\u11a8", 3, 0},
		{"\uac00\u0301", 3, 0},
		{"\uac01\u0301", 4, 0},
		{"a\u0f43", 1, 0},
	} {
		assertMeasuredCapacity(t, tt.input)
		if allocations := testing.AllocsPerRun(10, func() { measure(tt.input, 0) }); allocations != 0 {
			t.Fatalf("measurement allocated %g times", allocations)
		}
		e := measure(tt.input, 0)
		if e.maxEntries != tt.entries || e.maxUnordered != tt.unordered {
			t.Fatalf("extent%+v want entries%d unordered%d", e, tt.entries, tt.unordered)
		}
		if e.outBytes < len(NFC.String(tt.input)) {
			t.Fatal("output bound too small")
		}
	}
	for r := rune(0); r <= utf8.MaxRune; r++ {
		if r >= 0xd800 && r <= 0xdfff {
			continue
		}
		input := string(r)
		assertMeasuredCapacity(t, input)
		assertMeasuredCapacity(t, "a"+input+"\u0300")
		assertMeasuredCapacity(t, "a\u0315"+input)
		e := measure(input, 0)
		if e.outBytes < len(NFC.String(input)) {
			t.Fatalf("scalar output bound U+%04X", r)
		}
	}
}

// assertMeasuredCapacity follows String while checking every storage capacity.
// Any fallback growth is a failure even when output and resource budgets pass.
func assertMeasuredCapacity(t *testing.T, input string) {
	t.Helper()
	start, normal := quickSpan(input, 0)
	if normal {
		return
	}
	e := measure(input, start)
	rb := reorderBuffer{rune: make([]segRune, 0, e.maxEntries), scratch: make([]uint32, 0, e.maxUnordered)}
	rb.out.Grow(e.outBytes)
	outputCapacity := rb.out.Cap()
	rb.out.WriteString(input[:start])
	for start < len(input) {
		start = decomposeSegment(&rb, input, start)
		end, _ := quickSpan(input, start)
		rb.out.WriteString(input[start:end])
		start = end
		if cap(rb.rune) != e.maxEntries || cap(rb.scratch) != e.maxUnordered || rb.out.Cap() != outputCapacity {
			t.Fatalf("measured storage grew for %q: extent %+v", input, e)
		}
	}
	if rb.out.Len() > e.outBytes {
		t.Fatal("measured output bound exceeded")
	}
}

func TestExpandingSegments(t *testing.T) {
	for _, size := range []int{31, 100, 10000, 500000} {
		for _, tt := range []struct{ atom, first, second string }{
			{"\u0344", "\u0308\u0301", ""},
			{"\u0f73", "\u0f71", "\u0f72"},
			{"\u0f75", "\u0f71", "\u0f74"},
			{"\u0f81", "\u0f71", "\u0f80"},
		} {
			want := strings.Repeat(tt.first, size) + strings.Repeat(tt.second, size)
			if got := NFC.String(strings.Repeat(tt.atom, size)); got != want {
				t.Fatalf("expanding%U size%d differs", []rune(tt.atom)[0], size)
			}
		}
	}
}
