package skip

const (
	hex  = "0123456789abcdef"
	hexu = "0123456789ABCDEF"
)

func encodeURL(b, str []byte, flags Str) (s Str, _ []byte) {
	s |= URL

	brk := NewWideset(":@/?&=#%+[]!$()*,;")
	brk.MergeRange(0, 31)

	hex := csel(flags.Is(EscUpper), hexu, hex)

	for i := 0; ; {
		done := i
		i = brk.SkipUntilUTF8(str, i)

		b = append(b, str[done:i]...)
		if i == len(b) {
			break
		}

		b = append(b, '%', hex[str[i]>>4], hex[str[i]&0xf])
		i++
	}

	return s, b
}
