package ctok

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"testing"

	"github.com/luispabon/ctok-go/internal/testcorpus"
)

func TestCountParity(t *testing.T) {
	data, err := os.ReadFile("testdata/parity/manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		Schema     int            `json:"schema"`
		Version    int            `json:"generator_version"`
		Seed       uint64         `json:"seed"`
		Count      int            `json:"count"`
		CorpusHash string         `json:"corpus_sha256"`
		CountsHash string         `json:"counts_sha256"`
		InputBytes int            `json:"input_bytes"`
		Categories map[string]int `json:"categories"`
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	if manifest.Schema != 1 || manifest.Version != testcorpus.Version || manifest.Seed != testcorpus.Seed || manifest.Count != testcorpus.Count {
		t.Fatal("corpus configuration differs")
	}
	digest := sha256.New()
	total := 0
	categories := make(map[string]int)
	var prefix [4]byte
	for i := range testcorpus.Count {
		text := testcorpus.Case(i)
		binary.LittleEndian.PutUint32(prefix[:], uint32(len(text)))
		digest.Write(prefix[:])
		digest.Write([]byte(text))
		total += len(text)
		categories[testcorpus.Category(i)]++
	}
	if hex.EncodeToString(digest.Sum(nil)) != manifest.CorpusHash || total != manifest.InputBytes {
		t.Fatal("corpus byte stream differs")
	}
	if len(categories) != len(manifest.Categories) {
		t.Fatal("category count differs")
	}
	for category, count := range categories {
		if manifest.Categories[category] != count {
			t.Fatalf("category%s count differs", category)
		}
	}
	data, err = os.ReadFile("testdata/parity/counts.bin.gz")
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(data)
	if hex.EncodeToString(sum[:]) != manifest.CountsHash {
		t.Fatal("expected counts hash differs")
	}
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	counter, err := New(V48)
	if err != nil {
		t.Fatal(err)
	}
	var record [8]byte
	for index := range testcorpus.Count {
		if _, err := io.ReadFull(reader, record[:]); err != nil {
			t.Fatal(err)
		}
		framed, content := int(binary.LittleEndian.Uint32(record[:4])), int(binary.LittleEndian.Uint32(record[4:]))
		text := testcorpus.Case(index)
		if got := counter.CountContent(text); got != content {
			t.Fatalf("case%d category%s input%q: content%d want%d", index, testcorpus.Category(index), text, got, content)
		}
		if got := counter.Count(text); got != framed || framed-content != 6 {
			t.Fatalf("case%d: framed%d want%d content%d", index, got, framed, content)
		}
	}
	if _, err := reader.Read(record[:1]); err != io.EOF {
		t.Fatalf("expected counts trailing data: %v", err)
	}
}
