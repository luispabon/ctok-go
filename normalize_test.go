package ctok

import (
	"encoding/json"
	"os"
	"testing"
)

func TestNormalizeReference(t *testing.T) {
	data, err := os.ReadFile("testdata/unicode/normalize.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Scalars   [][2]json.RawMessage `json:"scalar_changes"`
		Sequences [][2]string          `json:"sequences"`
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if len(fixture.Scalars) == 0 || len(fixture.Sequences) == 0 {
		t.Fatal("no normalization fixtures")
	}
	changes := make(map[rune]string, len(fixture.Scalars))
	for _, pair := range fixture.Scalars {
		var cp rune
		var text string
		if err := json.Unmarshal(pair[0], &cp); err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(pair[1], &text); err != nil {
			t.Fatal(err)
		}
		changes[cp] = text
	}
	for r := rune(0); r <= 0x10FFFF; r++ {
		if r >= 0xD800 && r <= 0xDFFF {
			continue
		}
		want, ok := changes[r]
		if !ok {
			want = string(r)
		}
		if got := normalizeV48(string(r)); got != want {
			t.Fatalf("normalize U+%04X = %q, want %q", r, got, want)
		}
	}
	for _, pair := range fixture.Sequences {
		if got := normalizeV48(pair[0]); got != pair[1] {
			t.Fatalf("normalizeV48(%q) = %q, want %q", pair[0], got, pair[1])
		}
	}
}

func TestNormalizeV48(t *testing.T) {
	for _, tt := range []struct{ name, input, want string }{
		{"empty", "", ""}, {"nfc", "A\u030a", "Å"},
		{"controls", "a\x00\x01\t\n\r\x7fb", "a \t\nb"},
		{"thai", "\u0e4d\u0e32", "\u0e33"}, {"lao", "\u0ecd\u0eb2", "\u0ecd\u0eb2"},
		{"private", "a\ue000b\uf8ff", "ab"}, {"astral private", "\U000f0000", "\U000f0000"},
		{"spaces", "\u00a0\u2000\u2028\u3000", "   \u3000"},
		{"quotes", "‘“x”’", "‘“x”’"}, {"markers", "\ufdd0\ufdd4", "\ufdd0\ufdd4"},
		{"invalid run", "a\xff\xfeb", "a�b"}, {"separate invalid", "\xffx\xff", "�x�"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeV48(tt.input); got != tt.want {
				t.Errorf("normalizeV48(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
