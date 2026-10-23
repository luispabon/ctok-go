package ctok

// countTail includes the frame's two newlines, then removes its final token.
// Only the newline ladder is relevant, so a rolling DP avoids input-sized scratch.
func (v *vocabulary) countTail(trailingNewlines int) int {
	if trailingNewlines <= 0 {
		return 0
	}
	window := 2
	if len(v.newlines) > 0 {
		window = v.newlines[len(v.newlines)-1] + 1
	}
	best := make([]int, window)
	for end := 1; end <= trailingNewlines+2; end++ {
		cost := best[(end-1)%window] + 1
		for _, length := range v.newlines {
			if length > end {
				break
			}
			if candidate := best[(end-length)%window] + 1; candidate < cost {
				cost = candidate
			}
		}
		best[end%window] = cost
	}
	return best[(trailingNewlines+2)%window] - 1
}
