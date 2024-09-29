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
	_

	CSV
	URL
	_
	Continue

	ErrChar
	ErrRune
	ErrEscape
	ErrQuote

	ErrIndex
	ErrBuffer
	_
	_

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

func DecodeString(b []byte, st int, flags Str, buf []byte) (s Str, _ []byte, rs, i int) {
	s, buf, _, rs, i = skipString(b, st, flags, buf, true)
	return s, buf, rs, i
}

func String(b []byte, st int, flags Str) (s Str, bs, rs, i int) {
	s, _, bs, rs, i = skipString(b, st, flags, nil, false)
	return s, bs, rs, i
}

func skipString(b []byte, st int, flags Str, buf []byte, dec bool) (s Str, _ []byte, bs, rs, i int) {
	if flags.Is(CSV) {
		return csvSkip(b, st, flags, buf, dec)
	}
	if flags.Is(URL) {
		return urlSkip(b, st, flags, buf, dec)
	}

	//	defer func() { log.Printf("skipStr  %d (%s) -> %d  => %v  from %v", st, b, i, s, loc.Caller(1)) }()
	s, brk, halt, fin, i := openStr(b, st, flags)
	if s.Err() {
		return s, buf, 0, 0, st
	}

	var r rune

	for i < len(b) {
		done := i

		s, rs, i = skipStrPart(b, i, rs, s, flags, brk.OrCopy(halt))
		bs += i - done
		if dec {
			buf = append(buf, b[done:i]...)
		}
		if s.Err() {
			return s, buf, bs, rs, i
		}

		if i == len(b) || fin.Is(b[i]) {
			break
		}
		if halt.Is(b[i]) {
			return s | ErrChar, buf, bs, rs, i
		}

		s, r, i = decodeStrChar(b, i, s, flags)
		if s.Err() {
			return s, buf, bs, rs, i
		}

		bs += runelen(r)
		rs++
		if dec {
			buf = utf8.AppendRune(buf, r)
		}
	}

	if i == len(b) {
		return s | ErrBuffer, buf, bs, rs, i
	}
	if !fin.Is(b[i]) {
		return s | ErrQuote, buf, bs, rs, i
	}

	return s, buf, bs, rs, i + 1
}

func openStr(b []byte, st int, flags Str) (s Str, brk, halt, fin Wideset, i int) {
	//	defer func() { log.Printf("openStr  %d %q  %#v -> %d  => %v  from %v", st, b[st], flags, i, s, loc.Caller(1)) }()
	i = st
	halt = NewWidesetRange(0, 31)

	switch {
	case i >= len(b):
		s |= ErrIndex
	case flags.Is(Raw) && (b[i] == '`' || flags.Is(Continue)):
		s |= Raw
		fin.Merge("`")
		brk.Merge("`")
		halt.AndNot(Whitespaces.Wide())

		i += csel(flags.Is(Continue), 0, 1)
	case flags.Is(Quo) && (b[i] == '"' || flags.Is(Continue)):
		s |= Quo
		fin.Merge(`"`)
		brk.Merge(`\"`)

		i += csel(flags.Is(Continue), 0, 1)
	case flags.Is(Sqt) && (b[i] == '\'' || flags.Is(Continue)):
		s |= Sqt
		fin.Merge(`'`)
		brk.Merge(`\'`)

		i += csel(flags.Is(Continue), 0, 1)
	default:
		s |= ErrQuote
	}

	return s, brk, halt, fin, i
}

func skipStrPart(b []byte, st, l int, s, flags Str, brk Wideset) (ss Str, ll, i int) {
	//	defer func() { log.Printf("skipStrP %d %q -> %d  => %v  from %v", st, b[st], i, ss, loc.Caller(1)) }()
	i = st

	for i < len(b) {
		if brk.Is(b[i]) {
			break
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
			return s | ErrRune, l, i
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
		return s | ErrIndex, 0, st
	}

	if b[i] != '\\' {
		if !utf8.FullRune(b[i:]) {
			return s | ErrBuffer, 0, st
		}

		r, size := utf8.DecodeRune(b[i:])
		if r == utf8.RuneError && !flags.Is(ErrRune) {
			return s | ErrRune, 0, st
		}

		return s, r, i + size
	}

	i++

	if i == len(b) {
		return s | ErrBuffer, 0, st
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
			return s | ErrBuffer, 0, st
		}

		for j := 0; j < size; j++ {
			if b[i+j] < '0' || b[i+j] > '7' {
				return s | ErrEscape, 0, st
			}

			r = r*8 + rune(b[i+j]-'0')
		}

		return s, r, i + size
	case '\\', '\'', '"', 'a', 'b', 'f', 'n', 'r', 't', 'v':
		return s, esc2rune[b[i]], i + 1
	default:
		return s | ErrEscape, 0, st
	}

	if i+size > len(b) {
		return s | ErrBuffer, 0, st
	}

	r = decodeEscape(b, i, size)
	if r < 0 {
		return s | Str(-r), 0, st
	}

	if utf16.IsSurrogate(r) && b[i-1] == 'u' {
		if i+10 > len(b) {
			return s | ErrBuffer, r, st
		}

		if b[i+4] == '\\' && b[i+5] == 'u' {
			r2 := decodeEscape(b, i+6, size)
			if r2 < 0 {
				return s | Str(-r2), 0, st
			}

			rr := utf16.DecodeRune(r, r2)
			if rr != utf8.RuneError {
				r = rr
				i += 6
			}
		}
	}

	return s, r, i + size
}

func decodeEscape(b []byte, i, size int) (r rune) {
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

func runelen(r rune) int {
	b := utf8.RuneLen(r)
	return csel(b >= 0, b, 3)
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
	if state.Flag('#') {
		fmt.Fprintf(state, "%#x", int(s))
		return
	}

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
