package skip

import (
	"fmt"
	"path/filepath"
	"runtime"
	"unicode/utf16"
	"unicode/utf8"
)

type (
	Str int
)

const (
	Raw Str = 1 << iota
	Dqt
	Sqt
	Bqt

	Continue
	Decode
	Escapes
	Unicode

	//

	ErrBuffer
	ErrQuote
	ErrRune
	ErrEscape

	ErrSymbol
	_
	EscEmpty
	EscPlus

	//

	EscPercent
	EscXX
	EscU4
	EscU8

	EscControl
	EscZero
	EscOctal
	DecNewline

	//

	Quo = Dqt

	StrErr     = 0x3f00
	StrEscapes = 0xff_c000
)

func String(b []byte, st int, flags Str) (s Str, bs, rs, i int) {
	s, _, bs, rs, i = defaultString(b, st, flags&^Decode, nil)
	return
}

func DecodeString(b []byte, st int, flags Str, buf []byte) (s Str, res []byte, rs, i int) {
	s, res, _, rs, i = defaultString(b, st, flags|Decode, buf)
	return
}

func defaultString(b []byte, st int, flags Str, buf []byte) (s Str, res []byte, bs, rs, i int) {
	var brk Wideset
	fin := ASCIIControls.Wide()

	s, q, i := StringOpen(b, st, flags)
	if s.Err() {
		return s, buf, bs, rs, i
	}

	if s.Any(Dqt | Sqt) {
		brk.Set('\\')
		flags |= EscControl | EscXX | EscU4 | EscU8
	}
	if s.Is(Bqt) {
		fin.Except("\n\t")
	}

	s, buf, bs, rs, i = StringBody(b, i, flags, s, buf, brk, fin.OrCopy(q))
	if s.Err() {
		return s, buf, bs, rs, i
	}
	if i < len(b) && fin.Is(b[i]) {
		return s | ErrSymbol, buf, bs, rs, i
	}

	s, i = StringClose(b, i, flags, s)
	if s.Err() {
		return s, buf, bs, rs, i
	}

	return s, buf, bs, rs, i
}

func StringOpen(b []byte, st int, flags Str) (s Str, q Wideset, i int) {
	i = st

	switch {
	case i >= len(b):
		return ErrBuffer, q, st
	case flags.Is(Dqt) && (flags.Is(Continue) || b[i] == '"'):
		s |= Dqt
		q.Set('"')
	case flags.Is(Sqt) && (flags.Is(Continue) || b[i] == '\''):
		s |= Sqt
		q.Set('\'')
	case flags.Is(Bqt) && (flags.Is(Continue) || b[i] == '`'):
		s |= Bqt
		q.Set('`')
	case flags.Is(Raw):
		s |= Raw
	default:
		return ErrQuote, q, st
	}

	if s.Any(Dqt|Sqt|Bqt) && !flags.Is(Continue) {
		i++
	}

	return s, q, i
}

func StringClose(b []byte, st int, flags, s Str) (ss Str, i int) {
	i = st

	switch {
	case s.Is(Raw):
		return s, st
	case i >= len(b):
		return s | ErrBuffer, st
	case s.Is(Dqt) && b[i] == '"':
	case s.Is(Sqt) && b[i] == '\'':
	case s.Is(Bqt) && b[i] == '`':
	default:
		return s | ErrQuote, st
	}

	return s, i + 1
}

func StringBody(b []byte, st int, flags, s Str, buf []byte, brk, fin Wideset) (_ Str, _ []byte, bs, rs, i int) {
	i = st
	stop := brk.OrCopy(fin)

	for i < len(b) {
		done := i

		s, rs, i = StringUntil(b, i, flags, s, rs, stop)
		bs += i - done
		if flags.Is(Decode) {
			buf = append(buf, b[done:i]...)
		}
		if s.Err() {
			return s, buf, bs, rs, i
		}

		if i == len(b) || fin.Is(b[i]) {
			break
		}

		var r rune
		s, r, i = DecodeRune(b, i, flags, s)
		if s.Err() {
			return s, buf, bs, rs, i
		}

		bs += runelen(r)
		rs++
		if flags.Is(Decode) {
			buf = utf8.AppendRune(buf, r)
		}
	}

	return s, buf, bs, rs, i
}

func StringUntil(b []byte, st int, flags, s Str, rs int, brk Wideset) (_ Str, rs1, i int) {
	i = st

	for i < len(b) {
		if brk.Is(b[i]) {
			break
		}
		if b[i] < utf8.RuneSelf {
			i++
			rs++
			continue
		}

		r, size := utf8.DecodeRune(b[i:])
		if r == utf8.RuneError && !flags.Is(ErrRune) {
			return s | ErrRune, rs, i
		}
		if r == utf8.RuneError {
			return s, rs, i
		}

		s |= Unicode
		i += size
		rs++
	}

	return s, rs, i
}

func DecodeRune(b []byte, st int, flags, s Str) (ss Str, r rune, i int) {
	//	defer func() {
	//		log.Printf("DecodeRune: f %#8v  s %#4v => ss %#4v  %3v -> %3v  from %v  b %q", flags, s, ss, st, i, caller(1), b[st:])
	//	}()
	i = st
	if i >= len(b) {
		return s | ErrBuffer, 0, st
	}

	if flags.Is(DecNewline) && (b[i] == '\n' || b[i] == '\r') {
		i++

		for i < len(b) && (b[i] == '\n' || b[i] == '\r') {
			i++
		}

		return s | DecNewline, '\n', i
	}

	if b[i] == '+' && flags.Is(EscPlus) {
		return s | Escapes, ' ', i + 1
	}

	if b[i] == '%' && flags.Is(EscPercent) {
		if i+3 > len(b) {
			return s | ErrBuffer, 0, st
		}

		r = DecodeHex(b, i+1, 2)
		if r < 0 {
			return s | Str(-r), 0, st
		}

		return s | Escapes, r, i + 3
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

	s |= Escapes

	var size int

	switch {
	case b[i] == 'x' && flags.Is(EscXX):
		i++
		size = 2
	case b[i] == 'u' && flags.Is(EscU4):
		i++
		size = 4
	case b[i] == 'U' && flags.Is(EscU8):
		i++
		size = 8
	case b[i] == '0' && flags.Is(EscZero):
		return s, '\x00', i + 1
	case b[i] >= '0' && b[i] <= '7' && flags.Is(EscOctal):
		size = 3
		if i+size > len(b) {
			return s | ErrBuffer, 0, st
		}

		for j := 0; j < size; j++ {
			if b[i+j] < '0' || b[i+j] > '7' {
				return s | ErrEscape, 0, st
			}

			r = r<<3 + rune(b[i+j]-'0')
		}

		return s, r, i + size
	default:
		if !flags.Is(EscControl) || int(b[i]) >= len(esc2sym) {
			return s | ErrEscape, 0, st
		}

		x := esc2sym[b[i]]
		if x == 0 {
			return s | ErrEscape, 0, st
		}

		return s, rune(x), i + 1
	}

	if i+size > len(b) {
		return s | ErrBuffer, 0, st
	}

	r = DecodeHex(b, i, size)
	if r < 0 && flags.Is(ErrRune) {
		return s, utf8.RuneError, i + size
	}
	if r < 0 {
		return s | Str(-r), 0, st
	}

	if utf16.IsSurrogate(r) && b[i-1] != 'u' { // expect surrogate only from \uXXXX encoding
		return s | ErrEscape, 0, st
	}

	if !utf16.IsSurrogate(r) {
		return s, r, i + size
	}

	i += 4

	if i == len(b) {
		return s | ErrBuffer, r, st
	}

	if b[i] != '\\' && flags.Any(ErrRune|ErrEscape) {
		return s, utf8.RuneError, i
	}
	if b[i] != '\\' {
		return s | ErrEscape, r, st
	}

	if i+1 == len(b) {
		return s | ErrBuffer, r, st
	}
	if b[i+1] != 'u' && flags.Is(ErrEscape) {
		return s, utf8.RuneError, i
	}
	if b[i+1] != 'u' {
		return s | ErrEscape, r, st
	}

	if i+6 > len(b) {
		return s | ErrBuffer, r, st
	}

	r2 := DecodeHex(b, i+2, size)
	if r2 < 0 {
		return s | Str(-r2), 0, st
	}

	r = utf16.DecodeRune(r, r2)
	if r == utf8.RuneError && flags.Is(ErrRune) {
		return s, r, i
	}
	if r == utf8.RuneError {
		return s | ErrRune, r, st
	}

	return s, r, i + 6
}

func DecodeHex(b []byte, i, size int) (r rune) {
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

	if r < 0 {
		return -rune(ErrRune)
	}

	return r
}

func runelen(r rune) int {
	b := utf8.RuneLen(r)
	return csel(b >= 0, b, 3)
}

func (s Str) Is(f Str) bool {
	return s&f == f
}

func (s Str) Any(f Str) bool {
	return s&f != 0
}

func (s Str) Err() bool {
	return s&StrErr != 0
}

func (s Str) Suppress(x Str) Str {
	return s &^ x
}

func (s Str) GoString() string {
	return fmt.Sprintf("0x%x", int64(s))
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

	add(ErrBuffer, "short buffer")
	add(ErrQuote, "bad quote")
	add(ErrRune, "bad rune")
	add(ErrEscape, "bad escape")
	add(ErrSymbol, "bad symbol")

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

func caller(d int) string {
	_, file, line, _ := runtime.Caller(1 + d)

	return fmt.Sprintf("%v:%v", filepath.Base(file), line)
}

var esc2sym = []byte{
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

var sym2esc = []byte{
	0:    '0',
	'\\': '\\',
	'/':  '/',
	'"':  '"',
	'\'': '\'',
	'\a': 'a',
	'\b': 'b',
	'\t': 't',
	'\n': 'n',
	'\v': 'v',
	'\f': 'f',
	'\r': 'r',
}
