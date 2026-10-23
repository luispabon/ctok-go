package ctok

import (
	"fmt"
	"strings"
	"testing"
)

func benchmarkInput(atom string, size int) string {
	return strings.Repeat(atom, size/len(atom)) + strings.Repeat("a", size%len(atom))
}

func BenchmarkCount(b *testing.B) {
	counter, err := New(V48)
	if err != nil {
		b.Fatal(err)
	}
	for _, input := range []struct{ name, atom string }{
		{"ascii", "func Count(s string) int { return len(s) }\n"},
		{"mixed", "Hello 世界 Καλημέρα مرحبا नमस्ते 👩\u200d💻\n"},
		{"short_words", "A a "},
		{"combining", "\u0315\u0300"},
		{"newlines", "\n"},
		{"dense_upper", "A,"},
		{"dense_lower", "a."},
		{"punct_space", " , "},
		{"upper_space", "A, "},
		{"stray_fusion", "\u0345A,"},
		{"digit_border", " १ ! "},
		{"nfc_exclusion", "\u0344"},
		{"nfc_mixed", "\u0344"},
		{"nfc_devanagari", "\u0958"},
		{"nfc_tibetan", "\u0f73"},
		{"nfc_tibetan75", "\u0f75"},
		{"nfc_tibetan81", "\u0f81"},
		{"nfc_singleton40", "\u0340"},
		{"nfc_singleton41", "\u0341"},
		{"nfc_singleton43", "\u0343"},
		{"nfc_hebrew2c", "\ufb2c"},
		{"nfc_hebrew2d", "\ufb2d"},
		{"nfc_musical160", "\U0001d160"},
		{"nfc_musical", "\U0001d15e"},
		{"hangul", "\u1100\u1161\u11a8"},
	} {
		for _, size := range []int{1024, 10240, 102400, 1024 * 1024, 4 * 1024 * 1024} {
			text := benchmarkInput(input.atom, size)
			if input.name == "nfc_mixed" {
				text = strings.Repeat("\u0344", (size-3)/2) + "\u0f73" + strings.Repeat("a", (size-3)%2)
			}
			b.Run(fmt.Sprintf("%s/%d", input.name, size), func(b *testing.B) {
				b.ReportAllocs()
				b.SetBytes(int64(len(text)))
				b.ResetTimer()
				for range b.N {
					counter.CountContent(text)
				}
			})
		}
	}
}

func BenchmarkVocabularyDecode(b *testing.B) {
	b.ReportAllocs()
	for range b.N {
		if _, err := decodeVocabulary(vocabularyBlob); err != nil {
			b.Fatal(err)
		}
	}
}
