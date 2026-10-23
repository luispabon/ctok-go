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
	rb := reorderBuffer{out: make([]byte, n, len(s))}
	copy(rb.out, s[:n])
	for n < len(s) {
		n = decomposeSegment(&rb, s, n)
		end, _ := quickSpan(s, n)
		rb.out = append(rb.out, s[n:end]...)
		n = end
	}
	return string(rb.out)
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
