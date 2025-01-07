//go:build ignore

package skip

import (
	"unicode/utf16"
	"unicode/utf8"
)

type (
	StrFlags int64

	StrDecoder struct {
		Flags StrFlags

		Final Wideset
		Halt  Wideset
	}
)

const (
	Raw StrFlags = 1 << iota
	Dqt
	Sqt
	Bqt

	_
	_
	Unicode
	Continue

	//

	ErrBuffer
	ErrSymbol
	ErrRune
	ErrEscape

	ErrQuote
	_
	_
	EscUpper

	//

	EscPercent
	EscXX
	EscU4
	EscU8

	EscZero
	EscBackslash
	EscDoubleQuote
	_

	//

	AllowNewline
	AllowControl

	//

	Quo = Dqt

	StrErr = ErrSymbol | ErrRune | ErrEscape | ErrQuote | ErrBuffer
)

func CSVDecoder(comma byte) *StrDecoder {
	return &StrDecoder{
		Flags: Dqt | Raw | EscDoubleQuote,
	}
}

func (d *StrDecoder) Skip(b []byte, st int) (s StrFlags, bs, rs, i int) {
	s, _, bs, rs, i = d.str(b, st, nil, false)
	return
}

func (d *StrDecoder) Decode(b []byte, st int, buf []byte) (s StrFlags, _ []byte, rs, i int) {
	s, buf, _, rs, i = d.str(b, st, buf, true)
	return
}

func (d *StrDecoder) str(b []byte, st int, buf []byte, dec bool) (s StrFlags, _ []byte, bs, rs, i int) {
	i = st

	var q byte

	switch {
	case i >= len(b):
		return s | ErrBuffer, buf, bs, rs, i
	case d.Flags.Is(Dqt) && (b[i] == '"' || d.Flags.Is(Continue)):
		q = '"'
		i += csel(d.Flags.Is(Continue), 0, 1)
	case d.Flags.Is(Sqt) && (b[i] == '\'' || d.Flags.Is(Continue)):
		q = '\''
		i += csel(d.Flags.Is(Continue), 0, 1)
	case d.Flags.Is(Bqt) && (b[i] == '`' || d.Flags.Is(Continue)):
		q = '`'
		i += csel(d.Flags.Is(Continue), 0, 1)
	case d.Flags.Is(Raw):
	default:
		return s | ErrQuote, buf, bs, rs, i
	}

	var r rune

	brk := d.Final
	brk.Or(d.Halt)

	for i < len(b) {
		done := i

		s, rs, i = skipStrPart(b, i, rs, s, d.Flags, brk)
		bs += i - done
		if dec {
			buf = append(buf, b[done:i]...)
		}
		if s.Err() {
			return s, buf, bs, rs, i
		}

		if i == len(b) || q != 0 && b[i] == q || d.Final.Is(b[i]) {
			break
		}
		if d.Halt.Is(b[i]) {
			return s | ErrSymbol, buf, bs, rs, i
		}

		s, r, i = decodeStrChar(b, i, s, d.Flags)
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
	if q != 0 && b[i] == q {
		i++
	}

	return s, buf, bs, rs, i
}

func skipStrPart(b []byte, st, l int, s, flags StrFlags, brk Wideset) (ss StrFlags, ll, i int) {
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

		s |= Unicode
		i += size
		l++
	}

	return s, l, i
}

func decodeStrChar(b []byte, st int, s, flags StrFlags) (ss StrFlags, r rune, i int) {
	//	defer func() { log.Printf("decStrCh %d %q -> %d  => %v  from %v", st, b[st], i, ss, loc.Caller(1)) }()
	i = st
	if i >= len(b) {
		return s | ErrBuffer, 0, st
	}

	if b[i] != '\\' {
		if !utf8.FullRune(b[i:]) {
			return s | ErrBuffer, 0, st
		}

		r, size := utf8.DecodeRune(b[i:])
		if r == utf8.RuneError && !flags.Is(ErrRune) {
			return s | ErrRune, 0, st
		}

		return s | Unicode, r, i + size
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
	default:
		if x := esc2char[b[i]]; x != 0 {
			return s, rune(x), i + 1
		}

		return s | ErrEscape, 0, st
	}

	if i+size > len(b) {
		return s | ErrBuffer, 0, st
	}

	r = decodeEscape(b, i, size)
	if r < 0 {
		return s | StrFlags(-r), 0, st
	}

	if utf16.IsSurrogate(r) && b[i-1] == 'u' {
		if i+10 > len(b) {
			return s | ErrBuffer, r, st
		}

		if b[i+4] == '\\' && b[i+5] == 'u' {
			r2 := decodeEscape(b, i+6, size)
			if r2 < 0 {
				return s | StrFlags(-r2), 0, st
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

func (s StrFlags) Is(f StrFlags) bool {
	return s&f == f
}

func (s StrFlags) Any(f StrFlags) bool {
	return s&f != 0
}

func (s StrFlags) Err() bool {
	return s&StrErr != 0
}

func csel[T any](c bool, x, y T) T {
	if c {
		return x
	}

	return y
}

var esc2char = []byte{
	'\\': '\\',
	'/':  '/',
	'"':  '"',
	'\'': '\'',
	'a':  '\a',
	'b':  '\b',
	't':  '\t',
	'n':  '\n',
	'v':  '\v',
	'f':  '\f',
	'r':  '\r',
}
