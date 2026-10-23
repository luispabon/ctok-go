package ctok

import "unicode/utf8"

// byteFloor uses at most four UTF-8 bytes and no shared mutable rune cache.
func (v *vocabulary) byteFloor(r rune) int {
	if r >= 0xE000 && r <= 0xE004 {
		r = bow + (r - 0xE000)
	}
	if _, ok := v.units[r]; ok {
		return 1
	}
	var encoded [utf8.UTFMax]byte
	n := utf8.EncodeRune(encoded[:], r)
	var best [utf8.UTFMax + 1]int
	for end := 1; end <= n; end++ {
		best[end] = best[end-1] + 1
		for start := 0; start < end-1; start++ {
			if _, ok := v.bytes[string(encoded[start:end])]; ok && best[start]+1 < best[end] {
				best[end] = best[start] + 1
			}
		}
	}
	return best[n]
}

func (v *vocabulary) countMarked(stream []rune) int {
	best := make([]int, len(stream)+1)
	for end := 1; end <= len(stream); end++ {
		r := stream[end-1]
		node, ok := v.child(0, r)
		floor := 1
		if !structuralMarker(r) && (!ok || !v.nodes[node].terminal) {
			floor = v.byteFloor(r)
		}
		best[end] = best[end-1] + floor
		if !ok {
			continue
		}
		for start := end - 2; start >= 0; start-- {
			node, ok = v.child(node, stream[start])
			if !ok {
				break
			}
			if v.nodes[node].terminal && best[start]+1 < best[end] {
				best[end] = best[start] + 1
			}
		}
	}
	return best[len(stream)]
}
