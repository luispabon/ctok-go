// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package norm

import (
	"strings"
	"unicode/utf8"
)

// segRune replaces the fixed byte/position storage in x/text. Its class is
// the monotonically compressed CCC index from the generated tables.
type segRune struct {
	r        rune
	ccc      uint8
	backward bool
}

func (s segRune) combinesBackward() bool { return s.backward }

// A Unicode scalar needs 21 bits; CCC and backward occupy the next nine bits.
func packEntry(entry segRune) uint32 {
	packed := uint32(entry.r) | uint32(entry.ccc)<<21
	if entry.backward {
		packed |= 1 << 29
	}
	return packed
}

func unpackEntry(packed uint32) segRune {
	return segRune{r: rune(packed & (1<<21 - 1)), ccc: uint8(packed >> 21), backward: packed&(1<<29) != 0}
}

type reorderBuffer struct {
	rune    []segRune
	scratch []uint32
	nrune   int
	out     strings.Builder
}

func growSegment(s []segRune, needed int) []segRune {
	if needed <= cap(s) {
		return s[:needed]
	}
	capacity := max(32, cap(s)*2, needed)
	out := make([]segRune, needed, capacity)
	copy(out, s)
	return out
}

func (rb *reorderBuffer) appendEntry(r rune, info Properties) {
	rb.rune = growSegment(rb.rune, rb.nrune+1)
	rb.rune[rb.nrune] = segRune{r: r, ccc: info.ccc, backward: info.combinesBackward()}
	rb.nrune++
}

func (rb *reorderBuffer) appendRune(r rune)          { rb.appendEntry(r, Properties{}) }
func (rb *reorderBuffer) assignRune(pos int, r rune) { rb.rune[pos] = segRune{r: r} }
func (rb *reorderBuffer) runeAt(n int) rune          { return rb.rune[n].r }

func (rb *reorderBuffer) insert(r rune, info Properties) {
	if hangulBase <= r && r < hangulEnd {
		rb.decomposeHangul(r)
		return
	}
	if !info.hasDecomposition() {
		rb.appendEntry(r, info)
		return
	}
	dcomp := info.Decomposition()
	for len(dcomp) > 0 {
		part := compInfo(nfcData.lookup(dcomp))
		if part.BoundaryBefore() && rb.nrune > 0 {
			rb.doFlush()
		}
		r, size := utf8.DecodeRune(dcomp)
		rb.appendEntry(r, part)
		dcomp = dcomp[size:]
	}
}

// order performs stable counting sort within nonzero-CCC runs. Zero classes
// remain fixed barriers, matching x/text's canonical insertion ordering.
func (rb *reorderBuffer) order() {
	for start := 0; start < rb.nrune; {
		if rb.rune[start].ccc == 0 {
			start++
			continue
		}
		end := start + 1
		ordered := true
		for end < rb.nrune && rb.rune[end].ccc != 0 {
			if rb.rune[end-1].ccc > rb.rune[end].ccc {
				ordered = false
			}
			end++
		}
		if !ordered {
			var counts [256]int
			for _, entry := range rb.rune[start:end] {
				counts[entry.ccc]++
			}
			position := 0
			for i, count := range counts {
				counts[i] = position
				position += count
			}
			if cap(rb.scratch) < end-start {
				rb.scratch = make([]uint32, end-start)
			} else {
				rb.scratch = rb.scratch[:end-start]
			}
			for _, entry := range rb.rune[start:end] {
				i := counts[entry.ccc]
				rb.scratch[i] = packEntry(entry)
				counts[entry.ccc]++
			}
			for i, entry := range rb.scratch {
				rb.rune[start+i] = unpackEntry(entry)
			}
		}
		start = end
	}
}

func (rb *reorderBuffer) doFlush() {
	rb.order()
	rb.compose()
	for _, entry := range rb.rune[:rb.nrune] {
		rb.out.WriteRune(entry.r)
	}
	rb.nrune = 0
	rb.rune = rb.rune[:0]
}
