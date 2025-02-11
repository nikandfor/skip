package skip

const (
	URLValue = Bqt
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

func URLQuery(b []byte, st int, flags Str) (s Str, bs, rs, i int) {
	flags &^= Decode

	s, _, bs, rs, i = urlQuery(b, st, flags, nil)
	return
}

func DecodeURLQuery(b []byte, st int, flags Str, buf []byte) (s Str, res []byte, rs, i int) {
	flags |= Decode

	s, res, _, rs, i = urlQuery(b, st, flags, buf)
	return
}

func urlQuery(b []byte, st int, flags Str, buf []byte) (s Str, res []byte, bs, rs, i int) {
	fin := ASCIIControls.Wide()
	fin.Merge("&#")
	if !flags.Is(URLValue) {
		fin.Set('=')
	}

	flags |= EscPlus | EscPercent

	s, buf, bs, rs, i = StringBody(b, st, flags, s, buf, percent, fin)
	if s.Suppress(flags & StrErr).Err() {
		return s, buf, bs, rs, i
	}

	return s, buf, bs, rs, i
}
