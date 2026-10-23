package ctok

import (
	"errors"
	"fmt"
)

// Family selects a pinned count-reconstruction algorithm and vocabulary.
type Family string

// V48 selects the ctok 4.8 count-reconstruction family.
const V48 Family = "4.8"

// ErrUnsupportedFamily indicates that New was given an unknown family.
var ErrUnsupportedFamily = errors.New("ctok: unsupported tokenizer family")

// Counter counts text using immutable shared data. Construct it with New;
// its zero value is not usable. A Counter supports concurrent calls.
type Counter struct {
	vocab *vocabulary
}

// New constructs a counter for family. Only V48 is supported.
func New(family Family) (*Counter, error) {
	if family != V48 {
		return nil, ErrUnsupportedFamily
	}
	vocab, err := loadVocabulary()
	if err != nil {
		return nil, fmt.Errorf("ctok: load vocabulary: %w", err)
	}
	return &Counter{vocab: vocab}, nil
}

// Count estimates one text-only user message, including its six-token frame.
// It is not a whole-request or tool-schema token count.
func (c *Counter) Count(text string) int {
	return c.CountContent(text) + 6
}

// CountContent counts normalized text without message framing. Counts are not
// additive across concatenation. Invalid UTF-8 runs become one U+FFFD each.
func (c *Counter) CountContent(text string) int {
	marked, trailing := streamV48(normalizeV48(text))
	return c.vocab.countMarked(marked) + c.vocab.countTail(trailing)
}
