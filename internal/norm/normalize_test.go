package norm

import (
	"strings"
	"testing"
)

func TestLongSegments(t *testing.T) {
	for _, size := range []int{31, 100, 10000, 500000} {
		input := "a" + strings.Repeat("\u0301", size)
		want := "á" + strings.Repeat("\u0301", size-1)
		if got := NFC.String(input); got != want {
			t.Fatalf("acute run %d differs", size)
		}
		if got := NFC.String("a" + strings.Repeat("\u0315", size)); got != "a"+strings.Repeat("\u0315", size) {
			t.Fatalf("already normalized run %d changed", size)
		}
		input = "a" + strings.Repeat("\u0315\u0300", size/2)
		want = "à" + strings.Repeat("\u0300", size/2-1) + strings.Repeat("\u0315", size/2)
		if got := NFC.String(input); got != want {
			t.Fatalf("reordered run %d differs", size)
		}
	}
}

func TestComposition(t *testing.T) {
	for _, tt := range []struct{ input, want string }{
		{"\u1100\u1161\u11a8", "각"}, {"\u1100\u1161", "가"},
		{"\uac00\u11a8", "각"}, {"\u0344", "\u0308\u0301"},
		{"\u0f73", "\u0f71\u0f72"}, {"A\u030a", "Å"},
		{"a\u0315\u0300", "à\u0315"},
		{"\u1100\u1161\u11a8" + strings.Repeat("\u0301", 100), "각" + strings.Repeat("\u0301", 100)},
		{"a\u0344" + strings.Repeat("\u0301", 100), "ä" + strings.Repeat("\u0301", 101)},
		{"a\u0f73" + strings.Repeat("\u0301", 100), "á\u0f71\u0f72" + strings.Repeat("\u0301", 99)},
	} {
		if got := NFC.String(tt.input); got != tt.want {
			t.Errorf("NFC(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func BenchmarkNFC(b *testing.B) {
	for _, tt := range []struct{ name, text string }{
		{"ASCII_1MiB", strings.Repeat("a", 1<<20)},
		{"acute_1MiB", "a" + strings.Repeat("\u0301", (1<<19)-1)},
		{"reorder_1MiB", "a" + strings.Repeat("\u0315\u0300", (1<<18)-1)},
	} {
		b.Run(tt.name, func(b *testing.B) {
			b.SetBytes(int64(len(tt.text)))
			b.ReportAllocs()
			for b.Loop() {
				if got := NFC.String(tt.text); len(got) == 0 {
					b.Fatal("empty normalization output")
				}
			}
		})
	}
}
