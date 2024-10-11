package skip

func csvSkip(b []byte, st int, flags Str, buf []byte, dec bool) (s Str, _ []byte, bs, rs, i int) {
	//	defer func() { log.Printf("csvSkip  from %v", loc.Caller(1)) }()
	s |= CSV

	var brk Wideset
	halt := NewWidesetRange(0, 31)
	var q byte
	i = st

	switch {
	case flags.Is(Quo) && (b[i] == '"' || flags.Is(Continue)):
		s |= Quo
		q = '"'
		brk.Set(q)
		halt.AndNot(Whitespaces.Wide())

		i += csel(flags.Is(Continue), 0, 1)
	case flags.Is(Sqt) && (b[i] == '\'' || flags.Is(Continue)):
		s |= Sqt
		q = '\''
		brk.Set(q)
		halt.AndNot(Whitespaces.Wide())

		i += csel(flags.Is(Continue), 0, 1)
	case flags.Is(Raw):
		if byte(flags) == 0 {
			flags |= ','
		}

		s |= Raw
		brk.Set(byte(flags))
		brk.Merge("\n\r")
		halt.Except("\t")

		s, rs, i = skipStrPart(b, i, rs, s, flags, brk.OrCopy(halt))
		bs = i - st
		if dec {
			buf = append(buf, b[st:i]...)
		}
		if s.Err() {
			return s, buf, bs, rs, i
		}
		if i < len(b) && !brk.Is(b[i]) {
			return s | ErrSymbol, buf, bs, rs, i
		}

		i = csvSkipComma(b, i, byte(flags))

		return s | CSV, buf, bs, rs, i
	default:
		return s | ErrQuote, buf, 0, 0, st
	}

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

		if i == len(b) || b[i] != q {
			break
		}

		if i+1 < len(b) && b[i+1] == q {
			bs++
			rs++
			if dec {
				buf = append(buf, q)
			}

			i += 2
			continue
		}

		break
	}
	if i == len(b) || b[i] != q {
		return s | ErrSymbol, buf, bs, rs, i
	}

	i++

	i = csvSkipComma(b, i, byte(flags))

	return s, buf, bs, rs, i
}

func csvSkipComma(b []byte, i int, comma byte) int {
	if i < len(b) && b[i] == comma {
		return i + 1
	}

	return i
}
