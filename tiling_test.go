package ctok

import (
	"encoding/json"
	"math/rand/v2"
	"strings"
	"testing"
)

func TestByteReference(t *testing.T) {
	v := testVocabulary(t)
	fixture := readVocabularyFixture(t)
	if len(fixture.Floors) < 2000 {
		t.Fatal("missing floor fixtures")
	}
	for _, pair := range fixture.Floors {
		var text string
		var want int
		if err := json.Unmarshal(pair[0], &text); err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(pair[1], &want); err != nil {
			t.Fatal(err)
		}
		if got := v.byteFloor([]rune(text)[0]); got != want {
			t.Fatalf("floor %q=%d, want%d", text, got, want)
		}
	}
}

func TestTileReference(t *testing.T) {
	v := testVocabulary(t)
	fixture := readVocabularyFixture(t)
	if len(fixture.Marked) < 1000 {
		t.Fatal("missing marked fixtures")
	}
	for i, pair := range fixture.Marked {
		var text string
		var want int
		if err := json.Unmarshal(pair[0], &text); err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(pair[1], &want); err != nil {
			t.Fatal(err)
		}
		if got := v.countMarked([]rune(text)); got != want {
			t.Fatalf("marked%d %q=%d, want%d", i, text, got, want)
		}
	}
}

func TestTileBruteForce(t *testing.T) {
	rng := rand.New(rand.NewPCG(931, 2))
	for range 50 {
		pieces := []string{"a", "b"}
		for range 8 {
			var piece strings.Builder
			for range rng.IntN(4) + 1 {
				piece.WriteByte("abc"[rng.IntN(3)])
			}
			pieces = append(pieces, piece.String())
		}
		v := buildVocabulary(pieces, nil)
		var brute func(string) int
		brute = func(s string) int {
			if s == "" {
				return 0
			}
			best := 1 + brute(s[1:])
			for _, p := range pieces {
				if strings.HasPrefix(s, p) {
					best = min(best, 1+brute(s[len(p):]))
				}
			}
			return best
		}
		for range 20 {
			var input strings.Builder
			for range rng.IntN(9) {
				input.WriteByte("abc"[rng.IntN(3)])
			}
			text := input.String()
			if got, want := v.countMarked([]rune(text)), brute(text); got != want {
				t.Fatalf("%q got%d want%d", text, got, want)
			}
		}
	}
}

func TestByteMinimumAndEscapes(t *testing.T) {
	v := buildVocabulary([]string{"é"}, []string{"\xf0\x9f", "\x9f\x99\x82"})
	for _, tt := range []struct {
		r    rune
		want int
	}{{'é', 1}, {'🙂', 2}, {'𐀀', 4}, {'\uE000', 3}} {
		if got := v.byteFloor(tt.r); got != tt.want {
			t.Errorf("floor U+%04X=%d want%d", tt.r, got, tt.want)
		}
	}
	for _, marker := range []rune{bow, eow, shift, caps} {
		if got := v.countMarked([]rune{marker}); got != 1 {
			t.Errorf("structural marker cost%d", got)
		}
	}
	if got := v.countMarked([]rune{'\uE000'}); got != 3 {
		t.Fatalf("literal escape cost%d", got)
	}
}
