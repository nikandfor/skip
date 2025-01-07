//go:build ignore

package skip

import (
	"unicode/utf16"
	"unicode/utf8"
)

var char2esc = []byte{
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

func EncodeString(buf, str []byte, flags Str) (s Str, _ []byte) {
	if flags.Is(CSV) {
		return encodeCSV(buf, str, flags)
	}
	if flags.Is(URL) {
		return encodeURL(buf, str, flags)
	}

	var q byte
	var brk Wideset

	switch {
	case flags.Is(Bqt):
		q = '`'
		brk.Set(q)
	case flags.Is(Quo):
		q = '"'
		brk.Merge(`"\`)
		brk.MergeRange(0, 31)
	case flags.Is(Sqt):
		q = '\''
		brk.Merge(`'\`)
		brk.MergeRange(0, 31)
	default:
		return s | ErrQuote, buf
	}

	if !flags.Is(Continue) {
		buf = append(buf, q)
	}

	bs := NewCharset("\a\b\t\n\v\f\r")
	hex := csel(flags.Is(EscUpper), hexu, hex)

	for i := 0; ; {
		done := i
		brk.SkipUntilUTF8(str, i)

		buf = append(buf, str[done:i]...)
		if i == len(buf) {
			break
		}

		r, size := utf8.DecodeRune(str[i:])
		if r == utf8.RuneError && size == 1 {
			return s | ErrRune, buf
		}

		c := byte(r)

		switch {
		case flags.Is(EscDouble) && r == rune(q):
			buf = append(buf, q, q)
		case flags.Is(EscZero) && r == 0,
			flags.Is(EscBackslash) && r != 0 && r < 32 && bs.Is(c):
			buf = append(buf, '\\', char2esc[c])
		case flags.Is(EscPercent) && r < 0x100:
			buf = append(buf, '%', hex[r>>4], hex[r&0xf])
		case flags.Is(EscXX) && r < 0x100:
			buf = append(buf, '\\', 'x', hex[r>>4], hex[r&0xf])
		case flags.Is(EscU4) && r < 0x10000:
			buf = append(buf, '\\', 'u', hex[r>>12&0xf], hex[r>>8&0xf], hex[r>>4&0xf], hex[r&0xf])
		case flags.Is(EscU8):
			buf = append(buf, '\\', 'U', '0', '0', hex[r>>20&0xf], hex[r>>16&0xf], hex[r>>12&0xf], hex[r>>8&0xf], hex[r>>4&0xf], hex[r&0xf])
		case flags.Is(EscU4):
			r1, r2 := utf16.EncodeRune(r)

			buf = append(buf, '\\', 'u', hex[r1>>12&0xf], hex[r1>>8&0xf], hex[r1>>4&0xf], hex[r1&0xf])
			buf = append(buf, '\\', 'u', hex[r2>>12&0xf], hex[r2>>8&0xf], hex[r2>>4&0xf], hex[r2&0xf])
		default:
			return s | ErrEscape, buf
		}

		i += size
	}

	if !flags.Is(Continue) {
		buf = append(buf, q)
	}

	return
}
