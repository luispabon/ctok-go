package ctok

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"math"
	"os"
	"strings"
	"sync"
	"testing"
)

type vocabularyFixture struct {
	Schema       int                  `json:"schema"`
	Groups       map[string]int       `json:"groups"`
	PayloadHash  string               `json:"payload_sha256"`
	PiecesHash   string               `json:"pieces_sha256"`
	Expanded     int                  `json:"expanded_count"`
	Units        int                  `json:"unit_count"`
	BlobHash     string               `json:"blob_sha256"`
	Floors       [][2]json.RawMessage `json:"floors"`
	Marked       [][2]json.RawMessage `json:"marked"`
	Tails        [][2]int             `json:"tails"`
	Contractions []string             `json:"contractions"`
}

func readVocabularyFixture(t *testing.T) vocabularyFixture {
	t.Helper()
	data, err := os.ReadFile("testdata/vocabulary.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture vocabularyFixture
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	return fixture
}

func testVocabulary(t *testing.T) *vocabulary {
	t.Helper()
	v, err := loadVocabulary()
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func TestVocabularyReference(t *testing.T) {
	fixture := readVocabularyFixture(t)
	v := testVocabulary(t)
	sum := sha256.Sum256(vocabularyBlob)
	if hex.EncodeToString(sum[:]) != fixture.BlobHash {
		t.Fatal("blob hash differs")
	}
	payloadSum := sha256.Sum256(vocabularyBlob[40:])
	if fixture.Schema != 1 || fixture.PayloadHash != hex.EncodeToString(payloadSum[:]) ||
		fixture.PiecesHash != "f49fb02d8600ebeb28a38f99f31d652d998919b1734280c325599fa185524c4a" {
		t.Fatal("schema or canonical vocabulary hash differs")
	}
	groups := 0
	for _, count := range fixture.Groups {
		groups += count
	}
	if groups != 15283 || fixture.Groups["bytes_fallback"] != 468 || fixture.Groups["markers"] != 3 {
		t.Fatal("source group counts differ")
	}
	if binary.LittleEndian.Uint32(vocabularyBlob[44:48]) != 468 || len(v.bytes) != 468+2755 {
		t.Fatal("byte-prefix or combined floor membership count differs")
	}
	terminals := 0
	for _, node := range v.nodes {
		if node.terminal {
			terminals++
		}
	}
	if terminals != fixture.Expanded || terminals != 14819 {
		t.Fatalf("expanded count %d", terminals)
	}
	if len(v.units) != fixture.Units || len(v.units) != 2755 {
		t.Fatalf("unit count %d", len(v.units))
	}
	for _, contraction := range fixture.Contractions {
		node := 0
		for _, r := range reverseRunes(contraction) {
			var ok bool
			node, ok = v.child(node, r)
			if !ok {
				t.Fatalf("missing contraction %q", contraction)
			}
		}
		if !v.nodes[node].terminal {
			t.Fatalf("nonterminal contraction %q", contraction)
		}
	}
	for _, marker := range []rune{bow, eow, shift} {
		if got := v.countMarked([]rune{marker}); got != 1 {
			t.Fatalf("marker U+%04X cost %d", marker, got)
		}
	}
}

func reverseRunes(s string) []rune {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return runes
}

func encodeTestVocabulary(pieces, prefixes []string) []byte {
	payload := binary.LittleEndian.AppendUint32(nil, uint32(len(pieces)))
	payload = binary.LittleEndian.AppendUint32(payload, uint32(len(prefixes)))
	for _, section := range [][]string{pieces, prefixes} {
		for _, record := range section {
			payload = binary.LittleEndian.AppendUint32(payload, uint32(len(record)))
			payload = append(payload, record...)
		}
	}
	sum := sha256.Sum256(payload)
	blob := append([]byte("CTOK48\x00\x01"), sum[:]...)
	return append(blob, payload...)
}

func TestVocabularyMalformed(t *testing.T) {
	valid := encodeTestVocabulary([]string{"a", "abc"}, []string{"\xff"})
	if _, err := decodeVocabulary(valid); err != nil {
		t.Fatal(err)
	}
	rehash := func(blob []byte) []byte { sum := sha256.Sum256(blob[40:]); copy(blob[8:40], sum[:]); return blob }
	badCount := append([]byte(nil), valid...)
	binary.LittleEndian.PutUint32(badCount[40:], math.MaxUint32)
	badLength := append([]byte(nil), valid...)
	binary.LittleEndian.PutUint32(badLength[48:], math.MaxUint32)
	cases := map[string][]byte{
		"empty": nil, "header": append([]byte("invalid!"), valid[8:]...),
		"schema":   append([]byte("CTOK48\x00\x02"), valid[8:]...),
		"checksum": append(append([]byte(nil), valid...), 0),
		"count":    rehash(badCount), "length": rehash(badLength),
		"utf8":             encodeTestVocabulary([]string{"\xff"}, nil),
		"empty piece":      encodeTestVocabulary([]string{""}, nil),
		"duplicate":        encodeTestVocabulary([]string{"a", "a"}, nil),
		"unordered":        encodeTestVocabulary([]string{"b", "a"}, nil),
		"long prefix":      encodeTestVocabulary([]string{"a"}, []string{"12345"}),
		"duplicate prefix": encodeTestVocabulary([]string{"a"}, []string{"x", "x"}),
		"trailing":         rehash(append(append([]byte(nil), valid...), 0)),
	}
	for name, blob := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := decodeVocabulary(blob); err == nil {
				t.Fatal("accepted malformed blob")
			}
		})
	}
	for n := 0; n < len(valid); n++ {
		blob := append([]byte(nil), valid[:n]...)
		if n >= 40 {
			blob = rehash(blob)
		}
		if _, err := decodeVocabulary(blob); err == nil {
			t.Fatalf("accepted truncation %d", n)
		}
	}
}

func TestVocabularyConcurrent(t *testing.T) {
	v := testVocabulary(t)
	var wg sync.WaitGroup
	for range 16 {
		wg.Go(func() {
			for range 100 {
				got, err := loadVocabulary()
				if err != nil || got != v {
					t.Error("shared load differs")
				}
				if v.countMarked([]rune(strings.Repeat("\n", 32))) != 1 {
					t.Error("shared count differs")
				}
			}
		})
	}
	wg.Wait()
}
