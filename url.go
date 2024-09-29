package skip

import (
	"unicode/utf8"
)

func urlSkip(b []byte, st int, flags Str, buf []byte, dec bool) (s Str, _ []byte, l, i int) {
	//	defer func() { log.Printf("urlSkip  from %v", loc.Caller(1)) }()
	s |= URL
	i = st

	brk := NewWideset(":@/?&=#%+[]!$()*,;")
	halt := NewWidesetRange(0, 31)

	var r rune

	for {
		done := i

		s, l, i = skipStrPart(b, i, l, s, flags, brk.OrCopy(halt))
		if dec {
			buf = append(buf, b[done:i]...)
		}
		if s.Err() {
			return s, buf, l, i
		}

		if i == len(b) || brk.Is(b[i]) && b[i] != '%' && b[i] != '+' {
			break
		}
		if halt.Is(b[i]) {
			return s | ErrChar, buf, l, i
		}

		s, r, i = decodeURLChar(b, i, s, flags)
		if s.Err() {
			return s, buf, l, i
		}

		l++
		if dec {
			buf = utf8.AppendRune(buf, r)
		}
	}

	return s, buf, l, i
}

func decodeURLChar(b []byte, st int, s, flags Str) (ss Str, r rune, i int) {
	i = st
	if i >= len(b) {
		return s | ErrIndex, 0, st
	}

	if b[i] == '+' {
		return s, ' ', i + 1
	}
	if b[i] != '%' {
		return s | ErrEscape, 0, st
	}

	i++

	if i+1 >= len(b) {
		return s | ErrBuffer, 0, st
	}

	r = decodeEscape(b, i, 2)
	if r < 0 {
		return s | Str(-r), 0, st
	}

	return s, r, i + 2
}
