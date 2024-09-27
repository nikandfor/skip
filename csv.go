package skip

func csvSkip(b []byte, st int, flags Str, buf []byte, dec bool) (s Str, _ []byte, l, i int) {
	//	defer func() { log.Printf("csvSkip  from %v", loc.Caller(1)) }()
	s |= CSV

	var brk, skip Wideset
	var q byte
	i = st

	switch {
	case flags.Is(Quo) && b[i] == '"':
		s |= Quo
		q = '"'
		brk.Set(q)
		skip.Merge("\t\n\r")
		i++
	case flags.Is(Sqt) && b[i] == '\'':
		s |= Sqt
		q = '\''
		brk.Set(q)
		skip.Merge("\t\n\r")
		i++
	case flags.Is(Raw):
		if byte(flags) == 0 {
			flags |= ','
		}

		s |= Raw
		brk.Set(byte(flags))
		brk.Merge("\n\r")
		skip.Merge("\t")

		s, l, i = skipStrPart(b, i, l, s, flags, brk, skip)
		if s.Err() {
			return s, buf, l, i
		}
		if i < len(b) && !brk.Is(b[i]) {
			return s | ErrChar, buf, l, i
		}

		if dec {
			buf = append(buf, b[st:i]...)
		}

		i = csvSkipComma(b, i)

		return s | CSV, buf, l, i
	default:
		return s | ErrQuote, buf, 0, st
	}

	for i < len(b) {
		done := i

		s, l, i = skipStrPart(b, i, l, s, flags, brk, skip)
		if s.Err() {
			return s, buf, i - st, i
		}

		if dec {
			buf = append(buf, b[done:i]...)
		}

		if i == len(b) || b[i] != q {
			break
		}

		if i+1 < len(b) && b[i+1] == q {
			l++
			if dec {
				buf = append(buf, q)
			}

			i += 2
			continue
		}

		break
	}
	if i == len(b) || b[i] != q {
		return s | ErrQuote, buf, l, i
	}

	i++

	i = csvSkipComma(b, i)

	return s, buf, l, i
}

func csvSkipComma(b []byte, i int) int {
	if i < len(b) && b[i] == ',' {
		return i + 1
	}

	return i
}
