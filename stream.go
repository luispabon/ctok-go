package ctok

import "unicode/utf8"

const classStray runeClass = 5

type streamRun struct {
	class runeClass
	body  string
}

// runReader keeps only adjacent runs, not input-sized run metadata.
type runReader struct {
	text string
	pos  int
}

func (r *runReader) next() streamRun {
	if r.pos == len(r.text) {
		return streamRun{}
	}
	start := r.pos
	first, size := utf8.DecodeRuneInString(r.text[start:])
	class := classifyRune(first)
	if class == classWordy && strayMark(first) {
		class = classStray
	}
	kind := byte(0)
	if class == classHard {
		kind = hardKind(first)
	}
	r.pos += size
	for r.pos < len(r.text) {
		ch, width := utf8.DecodeRuneInString(r.text[r.pos:])
		c := classifyRune(ch)
		if class == classStray {
			if c != classWordy || !strayMark(ch) {
				break
			}
		} else if c != class {
			break
		}
		if class == classHard && !isSelector(ch) && hardKind(ch) != kind {
			break
		}
		r.pos += width
	}
	return streamRun{class: class, body: r.text[start:r.pos]}
}

// streamV48 accepts normalized text and returns its marked content and LF tail.
func streamV48(normalized string) ([]rune, int) {
	end := len(normalized)
	for end > 0 && normalized[end-1] == '\n' {
		end--
	}
	trailing := len(normalized) - end
	text := normalized[:end]
	// Literal structural glyphs are escaped after normalization strips BMP PUA.
	needsEscape := false
	for _, r := range text {
		if r >= bow && r <= caps {
			needsEscape = true
			break
		}
	}
	if needsEscape {
		escaped := make([]rune, 0, utf8.RuneCountInString(text))
		for _, r := range text {
			if r >= bow && r <= caps {
				r = 0xe000 + r - bow
			}
			escaped = append(escaped, r)
		}
		text = string(escaped)
	}
	if text == "" {
		return nil, trailing
	}
	reader := runReader{text: text}
	current, next := reader.next(), reader.next()
	var previous, beforePrevious streamRun
	out := make([]rune, 0, streamCapacity(text))
	for current.body != "" {
		leftSpace := previous.class == classSpace && len(previous.body) > 0 && previous.body[len(previous.body)-1] == ' '
		rightSpace := next.class == classSpace && len(next.body) > 0 && next.body[0] == ' ' && (len(next.body) == 1 || next.body[1] != ' ')
		first, _ := utf8.DecodeRuneInString(current.body)
		last, _ := utf8.DecodeLastRuneInString(current.body)
		switch {
		case current.class == classWordy:
			body := current.body
			fused := previous.body != "" && previous.class == classStray
			lowered := !fused && wordShift(body)
			if lowered {
				out = append(out, shift)
			}
			if !fused && !contractionSeam(current, previous, beforePrevious) {
				out = append(out, bow)
			}
			out = appendWord(out, body, lowered)
			out = append(out, eow)
		case current.class == classStray:
			out = append(out, bow)
			out = append(out, []rune(current.body)...)
			if next.body == "" || next.class != classWordy {
				out = append(out, eow)
			}
		case current.class == classPunct || hardBow(first) || hardBow(last):
			opensWord := current.body == "'" && next.body != "" && (next.class == classWordy || next.class == classStray)
			if leftSpace && !opensWord && (current.class == classPunct || hardBow(first)) {
				out = append(out, bow)
			}
			out = append(out, []rune(current.body)...)
			if rightSpace && (current.class == classPunct || hardBow(last)) {
				out = append(out, eow)
			}
		case (current.class == classDigit || current.class == classHard) && digitRun(current.body):
			if leftSpace && digitBorder(first) {
				out = append(out, bow)
			}
			out = append(out, []rune(current.body)...)
			if rightSpace && digitBorder(last) {
				out = append(out, eow)
			}
		default:
			out = append(out, []rune(current.body)...)
		}
		beforePrevious, previous = previous, current
		current, next = next, reader.next()
	}
	return collapseSeams(out), trailing
}

// collapseSeams mirrors the non-overlapping (.) EOW space case* BOW regex.
// Its dot excludes LF, and it deletes only the single literal ASCII space.
func collapseSeams(stream []rune) []rune {
	write := 0
	for read := 0; read < len(stream); {
		if stream[read] != '\n' && read+3 < len(stream) && stream[read+1] == eow && stream[read+2] == ' ' {
			end := read + 3
			for end < len(stream) && (stream[end] == shift || stream[end] == caps) {
				end++
			}
			if end < len(stream) && stream[end] == bow {
				stream[write], stream[write+1] = stream[read], eow
				write += 2
				write += copy(stream[write:], stream[read+3:end+1])
				read = end + 1
				continue
			}
		}
		stream[write] = stream[read]
		write++
		read++
	}
	return stream[:write]
}
