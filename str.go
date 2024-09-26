package skip

import "unicode/utf8"

type (
	Str int
)

const (
	Quo Str = 1 << iota
	Raw

	BadChar
	BadRune
	BadEscape
	BadQuote
	BadIndex

	StrErr = BadChar | BadRune | BadEscape | BadQuote | BadIndex
	StrOk  = Quo | Raw
)

var esc2rune = []rune{
	'\\': '\\',
	'\'': '\'',
	'"':  '"',
	'a':  '\a',
	'b':  '\b',
	'f':  '\f',
	'n':  '\n',
	'r':  '\r',
	't':  '\t',
	'v':  '\v',
}

func DecodeString(b []byte, st int, buf []byte) (s Str, _ []byte, i int) {
	i = st

	switch {
	case i >= len(b):
		return BadIndex, buf, st
	case b[i] == '`':
		s |= Raw
	case b[i] == '"':
		s |= Quo
	default:
		return BadQuote, buf, st
	}

	if s.Is(Raw) {
		s, _, i = String(b, st)
		if s.Any(StrErr) {
			return s, buf, i
		}

		buf = append(buf, b[st+1:i-1]...)

		return s, buf, i
	}

	var r rune
	done := i

	for i < len(b) {
		s, i = skipQuoRaw(b, i)
		if s.Any(StrErr) {
			return s, buf, i
		}

		buf = append(buf, b[done:i]...)
		done = i

		if i == len(b) || b[i] == '"' {
			break
		}

		s, r, i = decodeQuoChar(b, i)
		if s.Any(StrErr) {
			return s, buf, i
		}

		buf = utf8.AppendRune(buf, r)
	}

	if i == len(b) || b[i] != '"' {
		return BadQuote, buf, i
	}

	return s | Quo, buf, i + 1
}

func String(b []byte, st int) (s Str, l, i int) {
	i = st

	switch {
	case i >= len(b):
		return BadIndex, 0, st
	case b[i] == '`':
		s |= Raw
	case b[i] == '"':
		s |= Quo
	default:
		return BadQuote, 0, st
	}

	i++

	if s.Is(Raw) {
		for i < len(b) {
			if b[i] == '`' {
				break
			}
			if b[i] >= 0x20 && b[i] < 0x80 || b[i] == '\n' {
				i++
				continue
			}

			r, size := utf8.DecodeRune(b[i:])
			if r == utf8.RuneError {
				return BadRune, 0, i
			}

			i += size
		}

		if i == len(b) || b[i] != '`' {
			return BadQuote, 0, i
		}

		return s | Raw, i - st - 2, i + 1
	}

	done := i

	for i < len(b) {
		s, i = skipQuoRaw(b, i)
		if s.Any(StrErr) {
			return s, 0, i
		}

		l += i - done
		done = i

		if i == len(b) || b[i] == '"' {
			break
		}

		s, _, i = decodeQuoChar(b, i)
		if s.Any(StrErr) {
			return s, 0, i
		}

		l++
	}

	if i == len(b) || b[i] != '"' {
		return BadQuote, 0, i
	}

	return s | Quo, l, i + 1
}

func skipQuoRaw(b []byte, st int) (s Str, i int) {
	i = st

	for i < len(b) {
		if b[i] == '"' {
			break
		}
		if b[i] == '\\' {
			break
		}
		if b[i] < 0x20 {
			return BadChar, i
		}
		if b[i] < 0x80 {
			i++
			continue
		}

		r, size := utf8.DecodeRune(b[i:])
		if r == utf8.RuneError {
			return BadRune, i
		}

		i += size
	}

	return s, i
}

func decodeQuoChar(b []byte, st int) (s Str, r rune, i int) {
	i = st
	if i >= len(b) {
		return BadIndex, 0, st
	}

	if b[i] != '\\' {
		r, size := utf8.DecodeRune(b[i:])
		if r == utf8.RuneError {
			return BadRune, 0, i
		}

		i += size

		return 0, r, i
	}

	i++

	if i == len(b) {
		return BadEscape, 0, i
	}

	var size int

	switch b[i] {
	case 'x':
		size = 2
	case 'u':
		size = 4
	case 'U':
		size = 8
	case '0', '1', '2', '3', '4', '5', '6', '7':
		size = 3
		if i+size >= len(b) {
			return BadEscape, 0, i - 1
		}

		for j := 0; j < size; j++ {
			if b[i+j] < '0' || b[i+j] > '7' {
				return BadEscape, 0, i - 1
			}

			r = r*8 + rune(b[i+j]-'0')
		}

		return 0, r, i + size
	case '\\', '\'', '"', 'a', 'b', 'f', 'n', 'r', 't', 'v':
		return 0, esc2rune[b[i]], i + 1
	}

	if i+size >= len(b) {
		return BadEscape, 0, i - 1
	}

	for j := 0; j < size; j++ {
		c := b[i+j] | 0x20 // make lower
		r = r << 4

		if c >= '0' && c <= '9' {
			r += rune(c - '0')
		} else if c >= 'a' && c <= 'f' {
			r += 10 + rune(c-'a')
		} else {
			return BadEscape, 0, i - 1
		}
	}

	return 0, r, i + size
}

func (s Str) Is(x Str) bool {
	return s&x == x
}

func (s Str) Any(x Str) bool {
	return s&x != 0
}
