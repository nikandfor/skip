package skip

import (
	"fmt"
	"unicode/utf16"
	"unicode/utf8"
)

type (
	Str int
)

const (
	// first byte is arg

	Quo Str = 1 << (iota + 8)
	Sqt
	Raw
	CSV

	_
	_
	_
	_

	ErrChar
	ErrRune
	ErrEscape
	ErrQuote
	ErrIndex
	ErrBuffer

	StrErr = ErrChar | ErrRune | ErrEscape | ErrQuote | ErrIndex | ErrBuffer
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

func DecodeString(b []byte, st int, flags Str, buf []byte) (s Str, _ []byte, i int) {
	s, buf, _, i = skipString(b, st, flags, buf, true)
	return s, buf, i
}

func String(b []byte, st int, flags Str) (s Str, l, i int) {
	s, _, l, i = skipString(b, st, flags, nil, false)
	return s, l, i
}

func skipString(b []byte, st int, flags Str, buf []byte, dec bool) (s Str, _ []byte, l, i int) {
	s, brk, skip, fin, i := openStr(b, st, flags)
	if s.Err() {
		return s, buf, 0, st
	}

	if s.Is(CSV) {
		return csvSkip(b, st, flags, buf, dec)
	}

	var r rune

	for i < len(b) {
		done := i

		s, l, i = skipStrPart(b, i, l, s, flags, brk, skip)
		if s.Err() {
			return s, buf, l, i
		}

		if dec {
			buf = append(buf, b[done:i]...)
		}

		if i == len(b) || fin.Is(b[i]) {
			break
		}

		s, r, i = decodeStrChar(b, i, s, flags)
		if s.Err() {
			return s, buf, l, i
		}

		l++
		buf = utf8.AppendRune(buf, r)
	}

	if i == len(b) || !fin.Is(b[i]) {
		return ErrQuote, buf, l, i
	}

	return s, buf, l, i + 1
}

func openStr(b []byte, st int, flags Str) (s Str, brk, skip, fin Wideset, i int) {
	//	defer func() { log.Printf("openStr  %d %q -> %d  => %v  from %v", st, b[st], i, s, loc.Caller(1)) }()
	i = st

	switch {
	case i >= len(b):
		s |= ErrIndex
	case flags.Is(CSV):
		s |= CSV
	case flags.Is(Raw) && b[i] == '`':
		s |= Raw
		fin.Merge("`")
		brk.Merge("`")
		skip.Merge("\t\n")
		i++
	case flags.Is(Quo) && b[i] == '"':
		s |= Quo
		fin.Merge(`"`)
		brk.Merge(`\"`)
		i++
	case flags.Is(Sqt) && b[i] == '\'':
		s |= Sqt
		fin.Merge(`'`)
		brk.Merge(`\'`)
		i++
	default:
		s |= ErrQuote
	}

	return s, brk, skip, fin, i
}

func skipStrPart(b []byte, st, l int, s, flags Str, brk, skip Wideset) (ss Str, ll, i int) {
	//	defer func() { log.Printf("skipStrP %d %q -> %d  => %v  from %v", st, b[st], i, ss, loc.Caller(1)) }()
	i = st

	for i < len(b) {
		if brk.Is(b[i]) {
			break
		}
		if b[i] < 0x20 && !skip.Is(b[i]) {
			return ErrChar, l, i
		}
		if b[i] < 0x80 {
			i++
			l++
			continue
		}

		r, size := utf8.DecodeRune(b[i:])
		if r == utf8.RuneError && flags.Is(ErrRune) {
			return s, l, i
		}
		if r == utf8.RuneError {
			return ErrRune, l, i
		}

		i += size
		l++
	}

	return s, l, i
}

func decodeStrChar(b []byte, st int, s, flags Str) (ss Str, r rune, i int) {
	//	defer func() { log.Printf("decStrCh %d %q -> %d  => %v  from %v", st, b[st], i, ss, loc.Caller(1)) }()
	i = st
	if i >= len(b) {
		return ErrIndex, 0, st
	}

	if b[i] != '\\' {
		if !utf8.FullRune(b[i:]) {
			return ErrBuffer, 0, i
		}

		r, size := utf8.DecodeRune(b[i:])
		if r == utf8.RuneError && !flags.Is(ErrRune) {
			return ErrRune, 0, i
		}

		i += size

		return s, r, i
	}

	i++

	if i == len(b) {
		return ErrEscape, 0, i
	}

	var size int

	switch b[i] {
	case 'x':
		i++
		size = 2
	case 'u':
		i++
		size = 4
	case 'U':
		i++
		size = 8
	case '0', '1', '2', '3', '4', '5', '6', '7':
		size = 3
		if i+size >= len(b) {
			return ErrEscape, 0, i - 1
		}

		for j := 0; j < size; j++ {
			if b[i+j] < '0' || b[i+j] > '7' {
				return ErrEscape, 0, i - 1
			}

			r = r*8 + rune(b[i+j]-'0')
		}

		return s, r, i + size
	case '\\', '\'', '"', 'a', 'b', 'f', 'n', 'r', 't', 'v':
		return s, esc2rune[b[i]], i + 1
	default:
		return ErrEscape, 0, i
	}

	decode := func(i, size int) (r rune) {
		for j := 0; j < size; j++ {
			c := b[i+j] | 0x20 // make lower
			r = r << 4

			if c >= '0' && c <= '9' {
				r += rune(c - '0')
			} else if c >= 'a' && c <= 'f' {
				r += 10 + rune(c-'a')
			} else {
				return -rune(ErrEscape)
			}
		}

		return r
	}

	if i+size > len(b) {
		return ErrBuffer, 0, i - 1
	}

	r = decode(i, size)
	if r < 0 {
		return -Str(r), 0, i - 1
	}

	if utf16.IsSurrogate(r) && b[i-1] == 'u' && i+10 <= len(b) && b[i+4] == '\\' && b[i+5] == 'u' {
		r2 := decode(i+6, size)
		if r2 < 0 {
			return -Str(r2), 0, i - 1
		}

		rr := utf16.DecodeRune(r, r2)
		if rr != utf8.RuneError {
			r = rr
			i += 6
		}
	}

	return s, r, i + size
}

func (s Str) Err() bool {
	return s&StrErr != 0
}

func (s Str) Is(f Str) bool {
	return s&f == f
}

func (s Str) Any(f Str) bool {
	return s&f != 0
}

func (s Str) Format(state fmt.State, v rune) {
	if s.Err() {
		fmt.Fprintf(state, s.Error())
		return
	}

	fmt.Fprintf(state, "%#x", int(s))
}

func (s Str) Error() string {
	if !s.Err() {
		return "ok"
	}

	r := ""
	comma := false

	add := func(e Str, t string) {
		if !s.Is(e) {
			return
		}

		r += csel(comma, ", ", "")
		r += t
		comma = true
	}

	add(ErrChar, "bad char")
	add(ErrRune, "bad rune")
	add(ErrEscape, "bad escape")
	add(ErrQuote, "bad quote")
	add(ErrIndex, "bad index")
	add(ErrBuffer, "short buffer")

	if r == "" {
		r = fmt.Sprintf("%#x", int(s))
	}

	return r
}

func csel[T any](c bool, x, y T) T {
	if c {
		return x
	}

	return y
}
