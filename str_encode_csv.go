package skip

import "unicode/utf8"

func encodeCSV(b, str []byte, flags Str) (s Str, _ []byte) {
	s |= CSV

	if byte(flags) == 0 {
		flags |= ','
	}

	s = csvCheck(str, flags)
	if s.Err() {
		return s, b
	}

	var q byte
	needQuote := (byte(s) == byte(flags) || s.Any(strNewLine|Quo|Sqt))

	switch {
	case needQuote && !flags.Any(Quo|Sqt):
		return s | ErrQuote, b
	case !needQuote && flags.Is(Raw):
		q = 0
	case flags.Is(Quo|Sqt) && !s.Is(Quo) || flags.Is(Quo):
		q = '"'
	case flags.Is(Sqt):
		q = '\''
	}

	if !flags.Is(Continue) && q != 0 {
		b = append(b, q)
	}

	if q == 0 {
		b = append(b, str...)
	} else {
		brk := NewCharset(string(q))

		for i := 0; ; {
			done := i
			i = brk.SkipUntilUTF8(str, i)

			b = append(b, str[done:i]...)
			if i == len(b) {
				break
			}

			b = append(b, q, q)
			i++
		}
	}

	if !flags.Is(Continue) && q != 0 {
		b = append(b, q)
	}

	return
}

func csvCheck(str []byte, flags Str) (s Str) {
	halt := NewWidesetRange(0, 31)
	halt.Except("\n\t")

	for i := 0; i < len(str); {
		if str[i] < utf8.RuneSelf {
			switch {
			case str[i] == byte(flags):
				s |= flags & 0xff
			case str[i] == '"':
				s |= Quo
			case str[i] == '\'':
				s |= Sqt
			case str[i] == '\n':
				s |= strNewLine
			case halt.Is(str[i]):
				return s | ErrSymbol
			}

			i++

			continue
		}

		r, size := utf8.DecodeRune(str[i:])
		if r == utf8.RuneError && size == 1 {
			return s | ErrRune
		}

		i += size
	}

	return s
}
