package skip

func csvSkip(b []byte, st int, flags Str, buf []byte, dec bool) (s Str, _ []byte, l, i int) {
	//	defer func() { log.Printf("csvSkip  from %v", loc.Caller(1)) }()

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

		s, i = skipStrPart(b, i, s, brk, skip)
		l = i - st
		if s.Err() {
			return s, buf, l, i
		}
		if i < len(b) && !brk.Is(b[i]) {
			return ErrChar, buf, l, i
		}

		if dec {
			buf = append(buf, b[st:i]...)
		}

		i = csvSkipComma(b, i)

		return s | CSV, buf, l, i
	default:
		return ErrQuote, buf, 0, st
	}

	for i < len(b) {
		done := i

		s, i = skipStrPart(b, i, s, brk, skip)
		if s.Err() {
			return s, buf, i - st, i
		}

		l += i - done
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
		return ErrQuote, buf, l, i
	}

	i++

	i = csvSkipComma(b, i)

	return s | CSV, buf, l, i
}

func csvSkipComma(b []byte, i int) int {
	if i < len(b) && b[i] == ',' {
		return i + 1
	}

	if i < len(b) && b[i] == '\r' {
		i++
	}
	if i < len(b) && b[i] == '\n' {
		i++
	}

	return i
}
