// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package norm

import "encoding/binary"

const (
	headerLenMask   = 0x1F
	headerFlagsMask = 0xE0
)

// Properties holds normalization data decoded from x/text's generated trie.
type Properties struct {
	size  uint8
	ccc   uint8
	tccc  uint8
	nLead uint8
	flags qcInfo
	index uint16
}

type qcInfo uint8

func (p Properties) isInvalid() bool            { return p.flags == 0x80 }
func (p Properties) isYesC() bool               { return p.flags&0x10 == 0 }
func (p Properties) combinesBackward() bool     { return p.flags&0x8 != 0 }
func (p Properties) hasDecomposition() bool     { return p.flags&0x4 != 0 }
func (p Properties) nLeadingNonStarters() uint8 { return p.nLead }

// BoundaryBefore reports whether this rune cannot interact with the left.
func (p Properties) BoundaryBefore() bool {
	return p.ccc == 0 && !p.combinesBackward()
}

// Decomposition returns the generated canonical decomposition, if any.
func (p Properties) Decomposition() []byte {
	if p.index == 0 {
		return nil
	}
	i := p.index
	n := decomps[i] & headerLenMask
	if n == 31 {
		n = 33
	}
	i++
	return decomps[i : i+uint16(n)]
}

const recompShift = 21

func buildRecompMap() {
	recompMap = make(map[uint64]rune, len(recompMapPacked)/8)
	var buf [8]byte
	for i := 0; i < len(recompMapPacked); i += 8 {
		copy(buf[:], recompMapPacked[i:i+8])
		e := binary.BigEndian.Uint64(buf[:])
		recompMap[e>>recompShift] = rune(e & (1<<recompShift - 1))
	}
}

func combine(a, b rune) rune {
	key := uint64(a)<<recompShift | uint64(b)
	return recompMap[key]
}

func compInfo(v uint16, sz int) Properties {
	if sz == 0 {
		return Properties{flags: 0x80, size: 1}
	}
	if v == 0 {
		return Properties{size: uint8(sz)}
	} else if v >= 0x8000 {
		p := Properties{
			size: uint8(sz), ccc: uint8(v), tccc: uint8(v),
			flags: qcInfo(v>>8) & 0x3f,
		}
		if p.ccc > 0 || p.combinesBackward() {
			p.nLead = uint8(p.flags & 0x3)
		}
		return p
	}
	h := decomps[v]
	f := (qcInfo(h&headerFlagsMask) >> 2) | 0x4
	p := Properties{size: uint8(sz), flags: f, index: v}
	if v >= firstCCC {
		n := uint16(h & headerLenMask)
		if n == 31 {
			n = 33
		}
		v += n + 1
		c := decomps[v]
		p.tccc = c >> 2
		p.flags |= qcInfo(c & 0x3)
		if v >= firstLeadingCCC {
			p.nLead = c & 0x3
			if v >= firstStarterWithNLead {
				p.flags &= 0x03
				p.index = 0
				return p
			}
			p.ccc = decomps[v+1]
		}
	}
	return p
}
