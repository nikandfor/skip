package skip

const (
	URLValue = Continue
)

var (
	percent = NewWideset("%")
	urlSet  = NewWideset(":@/?&=#%+[]!$()*,;" + ".")
)

func URL(b []byte, st int, flags Str) (s Str, bs, rs, i int) {
	fine := Letters.OrCopy(Decimals.Wide())
	fine.Or(urlSet)
	fin := NewWidesetRange(0, 127).AndNotCopy(fine)

	flags &^= Decode
	flags |= EscPlus | EscPercent

	s, _, bs, rs, i = StringBody(b, st, flags, s, nil, percent, fin)
	if s.Suppress(flags & StrErr).Err() {
		return s, bs, rs, i
	}

	return s, bs, rs, i
}

func DecodeURLQuery(b []byte, st int, flags Str, buf []byte) (s Str, res []byte, rs, i int) {
	fin := ASCIILow.Wide()
	fin.Merge("&#")
	if !flags.Is(URLValue) {
		fin.Set('=')
	}

	flags |= Decode | EscPlus | EscPercent

	s, buf, _, rs, i = StringBody(b, st, flags, s, nil, percent, fin)
	if s.Suppress(flags & StrErr).Err() {
		return s, buf, rs, i
	}

	return s, buf, rs, i
}
