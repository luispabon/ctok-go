// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package norm contains the NFC-only x/text normalization implementation,
// adapted to preserve unbounded Python NFC segments without inserting CGJ.
package norm

import "unicode/utf8"

// Form identifies NFC, the only implemented form. No other values are supported.
type Form uint8

// NFC is canonical composition.
const NFC Form = 0

// String returns canonical composition. The caller must supply valid UTF-8;
// behavior on invalid byte sequences is unspecified.
func (Form) String(s string) string {
	n, normal := quickSpan(s, 0)
	if normal {
		return s
	}
	ext := measure(s, n)
	rb := reorderBuffer{
		rune:    make([]segRune, 0, ext.maxEntries),
		scratch: make([]uint32, 0, ext.maxUnordered),
	}
	rb.out.Grow(ext.outBytes)
	rb.out.WriteString(s[:n])
	for n < len(s) {
		n = decomposeSegment(&rb, s, n)
		end, _ := quickSpan(s, n)
		rb.out.WriteString(s[n:end])
		n = end
	}
	return rb.out.String()
}

// extent bounds storage before the non-normal path writes any output. It
// mirrors insert's internal flushes and decomposeSegment's segment boundaries.
type extent struct {
	maxEntries, maxUnordered, outBytes int
}

type extentBuffer struct {
	extent
	entries, nonstarters int
	lastCCC              uint8
	unordered            bool
}

func (e *extentBuffer) finishRun() {
	if e.unordered {
		e.maxUnordered = max(e.maxUnordered, e.nonstarters)
	}
	e.nonstarters, e.lastCCC, e.unordered = 0, 0, false
}

func (e *extentBuffer) flush() {
	e.finishRun()
	e.maxEntries = max(e.maxEntries, e.entries)
	e.entries = 0
}

func (e *extentBuffer) add(ccc uint8, size int) {
	e.entries++
	// Saturate rather than overflow for an output larger than an int can hold.
	if e.outBytes > int(^uint(0)>>1)-size {
		e.outBytes = int(^uint(0) >> 1)
	} else {
		e.outBytes += size
	}
	if ccc == 0 {
		e.finishRun()
		return
	}
	if e.lastCCC > ccc {
		e.unordered = true
	}
	e.lastCCC = ccc
	e.nonstarters++
}

func measure(s string, start int) extent {
	e := extentBuffer{extent: extent{outBytes: start}}
	for sp := start; sp < len(s); {
		info := lookupInfo(s[sp:])
		for {
			r, _ := utf8.DecodeRuneInString(s[sp:])
			switch {
			case hangulBase <= r && r < hangulEnd:
				count := 2
				if (r-hangulBase)%28 != 0 {
					count++
				}
				for range count {
					e.add(0, 3)
				}
			case info.hasDecomposition():
				for dcomp := info.Decomposition(); len(dcomp) > 0; {
					part := compInfo(nfcData.lookup(dcomp))
					if part.BoundaryBefore() && e.entries > 0 {
						e.flush()
					}
					_, size := utf8.DecodeRune(dcomp)
					e.add(part.ccc, size)
					dcomp = dcomp[size:]
				}
			default:
				e.add(info.ccc, utf8.RuneLen(r))
			}
			sp += int(info.size)
			if sp >= len(s) {
				break
			}
			info = lookupInfo(s[sp:])
			if info.nLeadingNonStarters() == 0 {
				break
			}
		}
		e.flush()
		end, _ := quickSpan(s, sp)
		if e.outBytes > int(^uint(0)>>1)-(end-sp) {
			e.outBytes = int(^uint(0) >> 1)
		} else {
			e.outBytes += end - sp
		}
		sp = end
	}
	return e.extent
}

func lookupInfo(s string) Properties { return compInfo(nfcData.lookupString(s)) }

func quickSpan(src string, i int) (int, bool) {
	var lastCC uint8
	lastSegStart := i
	for i < len(src) {
		if src[i] < utf8.RuneSelf {
			i++
			lastSegStart = i - 1
			lastCC = 0
			continue
		}
		info := lookupInfo(src[i:])
		if info.isInvalid() {
			return len(src), true
		}
		if info.nLeadingNonStarters() == 0 {
			lastSegStart = i
		} else if lastCC > info.ccc {
			return lastSegStart, false
		}
		if !info.isYesC() {
			return lastSegStart, false
		}
		lastCC = info.ccc
		i += int(info.size)
	}
	return i, true
}

func decomposeSegment(rb *reorderBuffer, src string, sp int) int {
	info := lookupInfo(src[sp:])
	for {
		r, _ := utf8.DecodeRuneInString(src[sp:])
		rb.insert(r, info)
		sp += int(info.size)
		if sp >= len(src) {
			break
		}
		info = lookupInfo(src[sp:])
		if info.nLeadingNonStarters() == 0 {
			break
		}
	}
	rb.doFlush()
	return sp
}
