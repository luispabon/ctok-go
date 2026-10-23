package ctok

import (
	"strings"
	"testing"
)

func TestTailReference(t *testing.T) {
	v := testVocabulary(t)
	fixture := readVocabularyFixture(t)
	if len(fixture.Tails) < 250 {
		t.Fatal("missing tail fixtures")
	}
	for _, pair := range fixture.Tails {
		if got := v.countTail(pair[0]); got != pair[1] {
			t.Fatalf("tail%d=%d want%d", pair[0], got, pair[1])
		}
	}
	if v.countTail(29) != 1 || v.countTail(30) != 0 || v.countTail(31) != 1 {
		t.Fatal("nonmonotonic frame-tail witness differs")
	}
}

func TestTailLong(t *testing.T) {
	v := testVocabulary(t)
	for _, n := range []int{1, 31, 100, 10000, 100000} {
		want := v.countMarked([]rune(strings.Repeat("\n", n+2))) - 1
		if got := v.countTail(n); got != want {
			t.Fatalf("tail%d=%d want%d", n, got, want)
		}
	}
	bytesPerCall := func(n int) int64 {
		return testing.Benchmark(func(b *testing.B) {
			for range b.N {
				v.countTail(n)
			}
		}).AllocedBytesPerOp()
	}
	short, long := bytesPerCall(100), bytesPerCall(100000)
	if long > short || long > 1024 {
		t.Fatalf("tail allocated bytes grew or exceeded window budget: %d to %d", short, long)
	}
}

func TestTailNonpositive(t *testing.T) {
	v := testVocabulary(t)
	for _, n := range []int{-100, -3, -2, -1, 0} {
		if got := v.countTail(n); got != 0 {
			t.Fatalf("tail%d=%d, want0", n, got)
		}
	}
}
