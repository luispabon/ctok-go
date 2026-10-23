package ctok

import (
	"encoding/json"
	"os"
	"testing"
)

func TestClassifyReference(t *testing.T) {
	data, err := os.ReadFile("testdata/unicode/properties.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Ranges [][]json.RawMessage `json:"ranges"`
		Lower  [][]json.RawMessage `json:"lower"`
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	checked := 0
	for _, row := range fixture.Ranges {
		if len(row) != 6 {
			t.Fatalf("range has %d fields, want 6", len(row))
		}
		var numbers [5]int
		for i := range numbers {
			if err := json.Unmarshal(row[i], &numbers[i]); err != nil {
				t.Fatal(err)
			}
		}
		var category string
		if err := json.Unmarshal(row[5], &category); err != nil {
			t.Fatal(err)
		}
		for r := rune(numbers[0]); r <= rune(numbers[1]); r++ {
			got := properties(r)
			if got.class != runeClass(numbers[2]) || got.flags != uint8(numbers[3]) ||
				got.combining != uint8(numbers[4]) || got.category != category {
				t.Fatalf("U+%04X: properties %+v disagree with oracle %v %s", r, got, numbers, category)
			}
			if classifyRune(r) != got.class {
				t.Fatalf("U+%04X: classification disagrees with properties", r)
			}
			checked++
		}
	}
	if checked != 0x110000-0x800 {
		t.Fatalf("checked %d scalars, want %d", checked, 0x110000-0x800)
	}
	mappings := make(map[rune]string, len(fixture.Lower))
	for _, row := range fixture.Lower {
		if len(row) != 2 {
			t.Fatalf("lower mapping has %d fields, want 2", len(row))
		}
		var cp rune
		var want string
		if err := json.Unmarshal(row[0], &cp); err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(row[1], &want); err != nil {
			t.Fatal(err)
		}
		mappings[cp] = want
	}
	for r := rune(0); r <= 0x10FFFF; r++ {
		if r >= 0xD800 && r <= 0xDFFF {
			continue
		}
		want, ok := mappings[r]
		if !ok {
			want = string(r)
		}
		if got := lowerRune(r); got != want {
			t.Fatalf("U+%04X: lower %q, want %q", r, got, want)
		}
	}
}
