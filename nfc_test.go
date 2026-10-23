package ctok

import (
	"compress/gzip"
	"encoding/json"
	"os"
	"testing"

	"github.com/luispabon/ctok-go/internal/norm"
)

func TestNFCConformance(t *testing.T) {
	file, err := os.Open("testdata/unicode/conformance.json.gz")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	reader, err := gzip.NewReader(file)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	var pairs [][2]string
	if err := json.NewDecoder(reader).Decode(&pairs); err != nil {
		t.Fatal(err)
	}
	if len(pairs) < 90000 {
		t.Fatalf("only %d conformance pairs", len(pairs))
	}
	for i, pair := range pairs {
		if got := norm.NFC.String(pair[0]); got != pair[1] {
			t.Fatalf("conformance pair %d: NFC(%q) = %q, want %q", i, pair[0], got, pair[1])
		}
	}
}

func TestNFCReference(t *testing.T) {
	data, err := os.ReadFile("testdata/unicode/nfc.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Scalars      []json.RawMessage    `json:"scalar_changes"`
		Sequences    [][2]string          `json:"sequences"`
		Interactions [][3]json.RawMessage `json:"interactions"`
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	changes := make(map[rune]string)
	for _, raw := range fixture.Scalars {
		var pair [2]json.RawMessage
		if err := json.Unmarshal(raw, &pair); err != nil {
			t.Fatal(err)
		}
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
	mismatches := 0
	for r := rune(0); r <= 0x10FFFF; r++ {
		if r >= 0xD800 && r <= 0xDFFF {
			continue
		}
		input := string(r)
		want, ok := changes[r]
		if !ok {
			want = input
		}
		if got := norm.NFC.String(input); got != want {
			if mismatches < 10 {
				t.Errorf("U+%04X: got %q want %q (tables %s)", r, got, want, norm.Version)
			}
			mismatches++
		}
	}
	for i, pair := range fixture.Sequences {
		if got := norm.NFC.String(pair[0]); got != pair[1] {
			t.Errorf("sequence %d (%d bytes) differs (tables %s)", i, len(pair[0]), norm.Version)
			mismatches++
		}
	}
	interactions := make(map[string]string, len(fixture.Interactions))
	for _, row := range fixture.Interactions {
		var text, want string
		if err := json.Unmarshal(row[1], &text); err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(row[2], &want); err != nil {
			t.Fatal(err)
		}
		interactions[text] = want
	}
	for r := rune(0); r <= 0x10FFFF; r++ {
		if r >= 0xD800 && r <= 0xDFFF {
			continue
		}
		for _, text := range []string{"a" + string(r) + "\u0300", "a\u0315" + string(r)} {
			want, ok := interactions[text]
			if !ok {
				want = text
			}
			if got := norm.NFC.String(text); got != want {
				if mismatches < 10 {
					t.Errorf("interaction U+%04X: got %q want %q", r, got, want)
				}
				mismatches++
			}
		}
	}
	if mismatches != 0 {
		t.Fatalf("%d NFC differences from pinned oracle", mismatches)
	}
}
