package skip

import (
	"math/bits"
	"strings"
	"unicode/utf8"
)

type (
	Charset uint64
	Wideset [2]Charset
)

var (
	Whitespaces = NewCharset(" \t\n\v\r")
	Decimals    = NewCharset("0123456789")
	Octals      = NewCharset("01234567")
	Nibbles     = NewCharset("0123")
	Binaries    = NewCharset("01")
	LowHexes    = Decimals.Wide().MergeCopy("abcdef")
	HighHexes   = Decimals.Wide().MergeCopy("ABCDEF")
	Hexes       = LowHexes.OrCopy(HighHexes)

	Lower   = NewWidesetRange('a', 'z')
	Upper   = NewWidesetRange('A', 'Z')
	Letters = Lower.OrCopy(Upper)

	Underscore = NewWideset("_")
	//	Punctuations = NewWideset(".,-+")

	WordSymbols = Letters.OrCopy(Decimals.Wide()).MergeCopy("_")

	IDFirst = Letters.SetCopy('_')
	IDRest  = IDFirst.OrCopy(Decimals.Wide())

	ASCIILow = NewCharsetRange(0, 31)
)

func Spaces(b []byte, i int) int {
	return Whitespaces.Skip(b, i)
}

func NewWideset(s string) (x Wideset) {
	return Wideset{}.MergeCopy(s)
}

func NewWidesetRange(a, b byte) Wideset {
	return Wideset{}.MergeRangeCopy(a, b)
}

func NewCharset(s string) (x Charset) {
	return Charset(0).MergeCopy(s)
}

func NewCharsetRange(a, b byte) Charset {
	return Charset(0).MergeRangeCopy(a, b)
}

func (x Wideset) Skip(b []byte, i int) int {
	for i < len(b) && x.Is(b[i]) {
		i++
	}

	return i
}

func (x Wideset) SkipUntil(b []byte, i int) int {
	for i < len(b) && !x.Is(b[i]) {
		i++
	}

	return i
}

func (x Wideset) SkipUntilUTF8(b []byte, i int) int {
	for i < len(b) {
		if b[i] < utf8.RuneSelf {
			if x.Is(b[i]) {
				return i
			}

			i++
		} else {
			r, size := utf8.DecodeRune(b[i:])
			if r == utf8.RuneError || x.Is('\n') && (r == '\u2028' || r == '\u2029') {
				return i
			}

			i += size
		}
	}

	return i
}

func (x Wideset) Is(b byte) bool {
	if b < 64 {
		return x[0].Is(b)
	}

	return x[1].Is(b - 64)
}

func (x *Wideset) Merge(s string) {
	*x = x.MergeCopy(s)
}

func (x Wideset) MergeCopy(s string) Wideset {
	for _, c := range []byte(s) {
		if c < 64 {
			x[0].Set(c)
		} else {
			x[1].Set(c - 64)
		}
	}

	return x
}

func (x *Wideset) MergeRange(a, b byte) {
	*x = x.MergeRangeCopy(a, b)
}

func (x Wideset) MergeRangeCopy(a, b byte) Wideset {
	for c := a; c <= b; c++ {
		if c < 64 {
			x[0].Set(c)
		} else {
			x[1].Set(c - 64)
		}
	}

	return x
}

func (x *Wideset) Except(s string) {
	*x = x.ExceptCopy(s)
}

func (x Wideset) ExceptCopy(s string) Wideset {
	for _, c := range []byte(s) {
		if c < 64 {
			x[0].Unset(c)
		} else {
			x[1].Unset(c - 64)
		}
	}

	return x
}

func (x *Wideset) ExceptRange(a, b byte) {
	*x = x.ExceptRangeCopy(a, b)
}

func (x Wideset) ExceptRangeCopy(a, b byte) Wideset {
	for c := a; c <= b; c++ {
		if c < 64 {
			x[0].Unset(c)
		} else {
			x[1].Unset(c - 64)
		}
	}

	return x
}

func (x *Wideset) Set(b byte) {
	*x = x.SetCopy(b)
}

func (x Wideset) SetCopy(b byte) Wideset {
	if b < 64 {
		x[0].Set(b)
	} else {
		x[1].Set(b - 64)
	}

	return x
}

func (x *Wideset) Unset(b byte) {
	*x = x.UnsetCopy(b)
}

func (x Wideset) UnsetCopy(b byte) Wideset {
	if b < 64 {
		x[0].Unset(b)
	} else {
		x[1].Unset(b - 64)
	}

	return x
}

func (x *Wideset) Or(y Wideset) {
	*x = x.OrCopy(y)
}

func (x Wideset) OrCopy(y Wideset) Wideset {
	x[0] |= y[0]
	x[1] |= y[1]

	return x
}

func (x *Wideset) AndNot(y Wideset) {
	*x = x.AndNotCopy(y)
}

func (x Wideset) AndNotCopy(y Wideset) Wideset {
	x[0] &^= y[0]
	x[1] &^= y[1]

	return x
}

func (x Charset) Skip(b []byte, i int) int {
	for i < len(b) && x.Is(b[i]) {
		i++
	}

	return i
}

func (x Charset) SkipUntil(b []byte, i int) int {
	for i < len(b) && !x.Is(b[i]) {
		i++
	}

	return i
}

func (x Charset) SkipUntilUTF8(b []byte, i int) int {
	for i < len(b) {
		if b[i] < utf8.RuneSelf {
			if x.Is(b[i]) {
				return i
			}

			i++
		} else {
			r, size := utf8.DecodeRune(b[i:])
			if r == utf8.RuneError || x.Is('\n') && (r == '\u2028' || r == '\u2029') {
				return i
			}

			i += size
		}
	}

	return i
}

func (x Charset) Is(b byte) bool {
	return b < 64 && x&(1<<b) == (1<<b)
}

func (x *Charset) Merge(s string) {
	*x = x.MergeCopy(s)
}

func (x Charset) MergeCopy(s string) Charset {
	for _, c := range s {
		x.Set(byte(c))
	}

	return x
}

func (x *Charset) MergeRange(a, b byte) {
	*x = x.MergeRangeCopy(a, b)
}

func (x Charset) MergeRangeCopy(a, b byte) Charset {
	for c := a; c <= b; c++ {
		x.Set(byte(c))
	}

	return x
}

func (x *Charset) Except(s string) {
	*x = x.ExceptCopy(s)
}

func (x Charset) ExceptCopy(s string) Charset {
	for _, c := range s {
		x.Unset(byte(c))
	}

	return x
}

func (x *Charset) ExceptRange(a, b byte) {
	*x = x.ExceptRangeCopy(a, b)
}

func (x Charset) ExceptRangeCopy(a, b byte) Charset {
	for c := a; c <= b; c++ {
		x.Unset(byte(c))
	}

	return x
}

func (x *Charset) Set(b byte) {
	*x = x.SetCopy(b)
}

func (x Charset) SetCopy(b byte) Charset {
	if b >= 64 {
		panic(b)
	}

	return x | 1<<b
}

func (x *Charset) Unset(b byte) {
	*x = x.UnsetCopy(b)
}

func (x Charset) UnsetCopy(b byte) Charset {
	if b >= 64 {
		panic(b)
	}

	return x &^ (1 << b)
}

func (x *Charset) Or(y Charset) {
	*x |= y
}

func (x Charset) OrCopy(y Charset) Charset {
	return x | y
}

func (x *Charset) AndNot(y Charset) {
	*x &^= y
}

func (x Charset) AndNotCopy(y Charset) Charset {
	return x &^ y
}

func (x Charset) Wide() Wideset { return Wideset{x, 0} }

func (x Wideset) String() string {
	var b strings.Builder

	_, _ = b.WriteString(x[0].String())

	for i := byte(64); i < 128; i++ {
		if x.Is(i) {
			b.WriteByte(i)
		}
	}

	return b.String()
}

func (x Charset) String() string {
	const hex = "0123456789abcdef"

	var b strings.Builder

	switch {
	case x&ASCIILow == ASCIILow:
		_, _ = b.WriteString("_low_")
	case x&ASCIILow == 0:
	case bits.OnesCount64(uint64(x&ASCIILow)) < 16:
		for i := byte(0); i < 32; i++ {
			if !x.Is(i) {
				continue
			}
			if e := sym2esc[i]; e != 0 {
				_, _ = b.Write([]byte{'\\', e})
				continue
			}

			_, _ = b.Write([]byte{'\\', 'x', hex[i>>4], hex[i&0xf]})
		}
	default:
		_, _ = b.WriteString("_exc_")

		for i := byte(0); i < 32; i++ {
			if x.Is(i) {
				continue
			}
			if e := sym2esc[i]; e != 0 {
				_, _ = b.Write([]byte{'\\', e})
				continue
			}

			_, _ = b.Write([]byte{'\\', 'x', hex[i>>4], hex[i&0xf]})
		}
	}

	for i := byte(32); i < 64; i++ {
		if x.Is(i) {
			_ = b.WriteByte(i)
		}
	}

	return b.String()
}
