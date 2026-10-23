package ctok

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"testing"

	"github.com/luispabon/ctok-go/internal/testcorpus"
)

func TestStreamReference(t *testing.T) {
	data, err := os.ReadFile("testdata/parity/streams.json.gz")
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(data)
	if hex.EncodeToString(sum[:]) != "09c84815ff6d9750cc0a386c4e89f80fd07fdccd8ed28ff3bde4a181c001ead6" {
		t.Fatal("stream artifact hash differs")
	}
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	var rows [][4]json.RawMessage
	if err := json.NewDecoder(reader).Decode(&rows); err != nil {
		t.Fatal(err)
	}
	if len(rows) != 5000 {
		t.Fatalf("stream fixture count%d", len(rows))
	}
	counter, err := New(V48)
	if err != nil {
		t.Fatal(err)
	}
	for index, row := range rows {
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
		if text != testcorpus.Case(index) {
			t.Fatalf("stream fixture input%d differs from corpus", index)
		}
		marked, trailing := streamV48(normalizeV48(text))
		if string(marked) != want || trailing != tail {
			t.Fatalf("case%d %q: stream%q tail%d want%q tail%d", index, text, string(marked), trailing, want, tail)
		}
		if got := counter.Count(text); got != count {
			t.Fatalf("case%d: count%d want%d", index, got, count)
		}
	}
}
