// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package norm

const (
	hangulBase   = 0xAC00
	hangulEnd    = hangulBase + jamoLVTCount
	jamoLBase    = 0x1100
	jamoLEnd     = 0x1113
	jamoVBase    = 0x1161
	jamoVEnd     = 0x1176
	jamoTBase    = 0x11A7
	jamoTEnd     = 0x11C3
	jamoVCount   = 21
	jamoTCount   = 28
	jamoVTCount  = 21 * 28
	jamoLVTCount = 19 * 21 * 28
)

func (rb *reorderBuffer) decomposeHangul(r rune) {
	r -= hangulBase
	x := r % jamoTCount
	r /= jamoTCount
	rb.appendRune(jamoLBase + r/jamoVCount)
	rb.appendRune(jamoVBase + r%jamoVCount)
	if x != 0 {
		rb.appendRune(jamoTBase + x)
	}
}

func (rb *reorderBuffer) combineHangul(s, i, k int) {
	b := rb.rune[:]
	bn := rb.nrune
	for ; i < bn; i++ {
		cccB := b[k-1].ccc
		cccC := b[i].ccc
		if cccB == 0 {
			s = k - 1
		}
		if s != k-1 && cccB >= cccC {
			b[k] = b[i]
			k++
		} else {
			l := rb.runeAt(s)
			v := rb.runeAt(i)
			switch {
			case jamoLBase <= l && l < jamoLEnd && jamoVBase <= v && v < jamoVEnd:
				rb.assignRune(s, hangulBase+(l-jamoLBase)*jamoVTCount+(v-jamoVBase)*jamoTCount)
			case hangulBase <= l && l < hangulEnd && jamoTBase < v && v < jamoTEnd && ((l-hangulBase)%jamoTCount) == 0:
				rb.assignRune(s, l+v-jamoTBase)
			default:
				if b[i].combinesBackward() {
					if combined := combine(l, v); combined != 0 {
						rb.assignRune(s, combined)
						continue
					}
				}
				b[k] = b[i]
				k++
			}
		}
	}
	rb.nrune = k
}

func (rb *reorderBuffer) compose() {
	recompMapOnce.Do(buildRecompMap)
	bn := rb.nrune
	if bn == 0 {
		return
	}
	k := 1
	b := rb.rune[:]
	for s, i := 0, 1; i < bn; i++ {
		if rb.runeAt(i)>>8 == 0x11 {
			rb.combineHangul(s, i, k)
			return
		}
		ii := b[i]
		cccB := b[k-1].ccc
		cccC := ii.ccc
		blocked := false
		if cccB == 0 {
			s = k - 1
		} else {
			blocked = s != k-1 && cccB >= cccC
		}
		if ii.combinesBackward() && !blocked {
			combined := combine(rb.runeAt(s), rb.runeAt(i))
			if combined != 0 {
				rb.assignRune(s, combined)
				continue
			}
		}
		b[k] = b[i]
		k++
	}
	rb.nrune = k
}
