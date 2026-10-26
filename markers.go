package ctok

const (
	bow   = '\uFDD0'
	eow   = '\uFDD1'
	pad   = '\uFDD2'
	shift = '\uFDD3'
	caps  = '\uFDD4'
)

// escapeMarker maps the marker range to literal U+E000..U+E004.
func escapeMarker(r rune) rune {
	if r >= bow && r <= caps {
		return 0xe000 + r - bow
	}
	return r
}

// unescapeMarker reverses escapeMarker for literal U+E000..U+E004.
func unescapeMarker(r rune) rune {
	if r >= 0xe000 && r <= 0xe000+caps-bow {
		return bow + r - 0xe000
	}
	return r
}
