package ctok

import (
	"encoding/json"
	"os"
	"testing"
)

func TestCaseReference(t *testing.T) {
	data, err := os.ReadFile("testdata/unicode/case.json")
	if err != nil {
		t.Fatal(err)
	}
	var pairs [][2]string
	if err := json.Unmarshal(data, &pairs); err != nil {
		t.Fatal(err)
	}
	if len(pairs) < 1000 {
		t.Fatalf("only %d case fixtures", len(pairs))
	}
	for _, pair := range pairs {
		if got := markCaseV48(pair[0]); got != pair[1] {
			t.Fatalf("markCaseV48(%q) = %q, want %q", pair[0], got, pair[1])
		}
	}
}

func TestCaseV48(t *testing.T) {
	for _, tt := range []struct{ input, want string }{
		{"", ""}, {"Hello", string(shift) + "hello"}, {"HELLO", "HELLO"},
		{"Σ", string(shift) + "σ"}, {"ΟΣ", "ΟΣ"}, {"Ος", string(shift) + "ος"},
		{"İ", "İ"}, {"ẞ", "ẞ"}, {"Abİ", string(shift) + "abİ"},
		{"ϴ", "ϴ"}, {"Θϴ", string(shift) + "θϴ"},
		{"\u1c89abc", string(shift) + "\u1c8aabc"},
		{"\ua7cbabc", string(shift) + "ɤabc"}, {"WiFi", "WiFi"},
	} {
		t.Run(tt.input, func(t *testing.T) {
			if got := markCaseV48(tt.input); got != tt.want {
				t.Errorf("markCaseV48(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
