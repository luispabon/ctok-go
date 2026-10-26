package ctok

import (
	"bytes"
	"crypto/sha256"
	_ "embed"
	"encoding/binary"
	"errors"
	"sort"
	"strings"
	"sync"
	"unicode/utf8"
)

//go:embed data/v48.bin
var vocabularyBlob []byte

var vocabularyOnce sync.Once
var sharedVocabulary *vocabulary
var vocabularyError error

type trieNode struct {
	first, count int
	terminal     bool
}

type trieEdge struct {
	r    rune
	next int
}

type vocabulary struct {
	nodes    []trieNode
	edges    []trieEdge
	bytes    map[string]struct{}
	units    map[rune]struct{}
	newlines []int
	maxDepth int
}

func loadVocabulary() (*vocabulary, error) {
	vocabularyOnce.Do(func() {
		sharedVocabulary, vocabularyError = decodeVocabulary(vocabularyBlob)
	})
	return sharedVocabulary, vocabularyError
}

func decodeVocabulary(blob []byte) (*vocabulary, error) {
	if len(blob) < 48 || !bytes.Equal(blob[:8], []byte("CTOK48\x00\x01")) {
		return nil, errors.New("invalid vocabulary header or schema")
	}
	sum := sha256.Sum256(blob[40:])
	if !bytes.Equal(blob[8:40], sum[:]) {
		return nil, errors.New("vocabulary checksum mismatch")
	}
	data := blob[40:]
	nPieces := uint64(binary.LittleEndian.Uint32(data))
	nPrefixes := uint64(binary.LittleEndian.Uint32(data[4:]))
	data = data[8:]
	if nPieces == 0 || nPieces+nPrefixes > uint64(len(data))/5 {
		return nil, errors.New("invalid vocabulary record counts")
	}
	pieces := make([]string, 0, int(nPieces))
	prefixes := make([]string, 0, int(nPrefixes))
	for i := uint64(0); i < nPieces+nPrefixes; i++ {
		if len(data) < 4 {
			return nil, errors.New("truncated vocabulary length")
		}
		n := uint64(binary.LittleEndian.Uint32(data))
		data = data[4:]
		if n == 0 || n > uint64(len(data)) {
			return nil, errors.New("invalid vocabulary record length")
		}
		value := string(data[:int(n)])
		data = data[int(n):]
		if i < nPieces {
			if !utf8.ValidString(value) {
				return nil, errors.New("invalid UTF-8 vocabulary piece")
			}
			if len(pieces) > 0 && pieces[len(pieces)-1] >= value {
				return nil, errors.New("unordered or duplicate vocabulary piece")
			}
			pieces = append(pieces, value)
		} else {
			if len(value) > utf8.UTFMax {
				return nil, errors.New("oversized byte prefix")
			}
			if len(prefixes) > 0 && prefixes[len(prefixes)-1] >= value {
				return nil, errors.New("unordered or duplicate byte prefix")
			}
			prefixes = append(prefixes, value)
		}
	}
	if len(data) != 0 {
		return nil, errors.New("trailing vocabulary data")
	}
	return buildVocabulary(pieces, prefixes), nil
}

func structuralMarker(r rune) bool {
	return r == bow || r == eow || r == shift || r == caps
}

func buildVocabulary(pieces, prefixes []string) *vocabulary {
	v := &vocabulary{nodes: []trieNode{{}}, bytes: make(map[string]struct{}), units: make(map[rune]struct{})}
	for _, prefix := range prefixes {
		v.bytes[prefix] = struct{}{}
	}
	reversed := make([]string, 0, len(pieces))
	for _, piece := range pieces {
		runes := []rune(piece)
		if len(runes) == 1 && !structuralMarker(runes[0]) {
			v.units[runes[0]] = struct{}{}
			v.bytes[piece] = struct{}{}
		}
		if strings.Trim(piece, "\n") == "" {
			v.newlines = append(v.newlines, len(runes))
		}
		for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
			runes[i], runes[j] = runes[j], runes[i]
		}
		reversed = append(reversed, string(runes))
	}
	sort.Strings(reversed)
	type pendingEdge struct {
		parent int
		edge   trieEdge
	}
	var edges []pendingEdge
	var previous []rune
	path := []int{0}
	for _, piece := range reversed {
		runes := []rune(piece)
		common := 0
		for common < len(previous) && common < len(runes) && previous[common] == runes[common] {
			common++
		}
		path = path[:common+1]
		for _, r := range runes[common:] {
			next := len(v.nodes)
			v.nodes = append(v.nodes, trieNode{})
			edges = append(edges, pendingEdge{path[len(path)-1], trieEdge{r, next}})
			path = append(path, next)
		}
		v.maxDepth = max(v.maxDepth, len(path)-1)
		v.nodes[path[len(path)-1]].terminal = true
		previous = runes
	}
	sort.Slice(edges, func(i, j int) bool {
		if edges[i].parent != edges[j].parent {
			return edges[i].parent < edges[j].parent
		}
		return edges[i].edge.r < edges[j].edge.r
	})
	v.edges = make([]trieEdge, len(edges))
	for i, edge := range edges {
		node := &v.nodes[edge.parent]
		if node.count == 0 {
			node.first = i
		}
		node.count++
		v.edges[i] = edge.edge
	}
	sort.Ints(v.newlines)
	return v
}

func (v *vocabulary) child(parent int, r rune) (int, bool) {
	node := v.nodes[parent]
	edges := v.edges[node.first : node.first+node.count]
	i := sort.Search(len(edges), func(i int) bool { return edges[i].r >= r })
	if i == len(edges) || edges[i].r != r {
		return 0, false
	}
	return edges[i].next, true
}
