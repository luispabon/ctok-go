package ctok

import "testing"

func TestEscapeMarker(t *testing.T) {
	for _, tt := range []struct {
		name string
		r    rune
		want rune
	}{
		{"bow", bow, 0xE000},
		{"eow", eow, 0xE001},
		{"pad", pad, 0xE002},
		{"shift", shift, 0xE003},
		{"caps", caps, 0xE004},
		{"below bow", bow - 1, bow - 1},
		{"above caps", caps + 1, caps + 1},
		{"ascii", 'a', 'a'},
		{"supplementary", '🙂', '🙂'},
	} {
		if got := escapeMarker(tt.r); got != tt.want {
			t.Errorf("%s: escapeMarker(U+%04X)=U+%04X want U+%04X", tt.name, tt.r, got, tt.want)
		}
	}
}

func TestUnescapeMarker(t *testing.T) {
	for _, tt := range []struct {
		name string
		r    rune
		want rune
	}{
		{"U+E000", 0xE000, bow},
		{"U+E001", 0xE001, eow},
		{"U+E002", 0xE002, pad},
		{"U+E003", 0xE003, shift},
		{"U+E004", 0xE004, caps},
		{"below U+E000", 0xE000 - 1, 0xE000 - 1},
		{"above U+E004", 0xE004 + 1, 0xE004 + 1},
		{"ascii", 'a', 'a'},
		{"supplementary", '🙂', '🙂'},
	} {
		if got := unescapeMarker(tt.r); got != tt.want {
			t.Errorf("%s: unescapeMarker(U+%04X)=U+%04X want U+%04X", tt.name, tt.r, got, tt.want)
		}
	}
}

func TestMarkerMappingRoundtrip(t *testing.T) {
	for r := bow; r <= caps; r++ {
		if got := unescapeMarker(escapeMarker(r)); got != r {
			t.Errorf("unescapeMarker(escapeMarker(U+%04X))=U+%04X", r, got)
		}
	}
	for r := rune(0xE000); r <= 0xE000+caps-bow; r++ {
		if got := escapeMarker(unescapeMarker(r)); got != r {
			t.Errorf("escapeMarker(unescapeMarker(U+%04X))=U+%04X", r, got)
		}
	}
}
