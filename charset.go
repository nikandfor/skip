package skip

import "strings"

type (
	Charset uint64
	Wideset [2]Charset
)

var (
	Whitespaces = NewCharset(" \t\n\r")
	Decimals    = NewCharset("0123456789")
	Octals      = NewCharset("01234567")
	Nibbles     = NewCharset("0123")
	Binaries    = NewCharset("01")
	LowHexes    = Decimals.Wide().MergeCopy("abcdef")
	HighHexes   = Decimals.Wide().MergeCopy("ABCDEF")
	Hexes       = LowHexes.OrCopy(HighHexes)

	Lower   = NewWideset("").MergeRangeCopy('a', 'z')
	Upper   = NewWideset("").MergeRangeCopy('A', 'Z')
	Letters = Lower.OrCopy(Upper)

	Underscore   = NewWideset("_")
	Punctuations = NewWideset(".,-+")

	WordSymbols = Letters.OrCopy(Decimals.Wide()).MergeCopy("_")

	IDFirst = Letters.SetCopy('_')
	IDRest  = IDFirst.OrCopy(Decimals.Wide())
)

func Spaces(b []byte, i int) int {
	return Whitespaces.Skip(b, i)
}

func NewWideset(s string) (x Wideset) {
	for _, c := range s {
		if c >= 64 {
			x[1] |= 1 << (c - 64)
		}

		x[0] |= 1 << c
	}

	return x
}

func NewCharset(s string) (x Charset) {
	for _, c := range s {
		if c >= 64 {
			panic(c)
		}

		x |= 1 << c
	}

	return x
}

func (x Wideset) Skip(b []byte, i int) int {
	for i < len(b) && x.Is(b[i]) {
		i++
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

func (x *Wideset) Or(y Wideset) {
	*x = x.OrCopy(y)
}

func (x Wideset) OrCopy(y Wideset) Wideset {
	x[0] |= y[0]
	x[1] |= y[1]

	return x
}

func (x *Wideset) Not(y Wideset) {
	*x = x.NotCopy(y)
}

func (x Wideset) NotCopy(y Wideset) Wideset {
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

func (x *Charset) Set(b byte) {
	*x = x.SetCopy(b)
}

func (x Charset) SetCopy(b byte) Charset {
	if b >= 64 {
		panic(b)
	}

	return x | 1<<b
}

func (x *Charset) Or(y Charset) {
	*x |= y
}

func (x Charset) OrCopy(y Charset) Charset {
	return x | y
}

func (x *Charset) Not(y Charset) {
	*x &^= y
}

func (x Charset) NotCopy(y Charset) Charset {
	return x &^ y
}

func (x Charset) Wide() Wideset { return Wideset{x, 0} }

func (x Wideset) String() string {
	var b strings.Builder

	for i := byte(0); i < 128; i++ {
		if x.Is(i) {
			b.WriteByte(i)
		}
	}

	return b.String()
}

func (x Charset) String() string {
	var b strings.Builder

	for i := byte(0); i < 64; i++ {
		if x.Is(i) {
			b.WriteByte(i)
		}
	}

	return b.String()
}
