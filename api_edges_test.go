package ctok

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"testing"
)

func TestCountUpstreamAPIEdges(t *testing.T) {
	data, err := os.ReadFile("testdata/parity/api_edges.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Source string               `json:"source_sha256"`
		Groups map[string]int       `json:"groups"`
		Rows   [][4]json.RawMessage `json:"rows"`
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.Source != "5818442ec9855191280323b36f672ca39c285083cd9d5fd69dd5710c0aff9723" || fixture.Groups["MARK_ROWS"] != 48 || len(fixture.Rows) < 150 {
		t.Fatal("missing upstream API edges")
	}
	sum := sha256.Sum256(data)
	if hex.EncodeToString(sum[:]) != "78bb31da0fb67b987a7111cda106dc0d224c0ea95835c81cfc1e836d47eb8734" {
		t.Fatal("API edges artifact hash differs")
	}
	counter, err := New(V48)
	if err != nil {
		t.Fatal(err)
	}
	for index, row := range fixture.Rows {
		var text, want string
		var tail, count int
		for _, field := range []struct {
			raw json.RawMessage
			dst any
		}{{row[0], &text}, {row[1], &want}, {row[2], &tail}, {row[3], &count}} {
			if err := json.Unmarshal(field.raw, field.dst); err != nil {
				t.Fatal(err)
			}
		}
		marked, trailing := streamV48(normalizeV48(text))
		if string(marked) != want || trailing != tail || counter.Count(text) != count || counter.CountContent(text) != count-6 {
			t.Fatalf("API edge%d %q differs", index, text)
		}
	}
}
