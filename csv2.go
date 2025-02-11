package skip

import (
	"fmt"
	"runtime"
)

func CSV(b []byte, st int, flags Str, comma byte) (s Str, bs, rs, i int) {
	flags &^= Decode

	s, _, bs, rs, i = csv(b, st, flags, comma, nil)
	return
}

func DecodeCSV(b []byte, st int, flags Str, comma byte, buf []byte) (s Str, res []byte, rs, i int) {
	flags |= Decode

	s, res, _, rs, i = csv(b, st, flags, comma, buf)
	return
}

func csv(b []byte, st int, flags Str, comma byte, buf []byte) (s Str, res []byte, bs, rs, i int) {
	if !flags.Any(Dqt | Sqt | Bqt | Raw) {
		flags |= Dqt | Raw
	}

	//	defer func() { log.Printf("csv %v: %d -> %d  %d, %d   %q  from %v", s, st, i, bs, rs, b, caller(1)) }()

	s, q, i := StringOpen(b, st, flags)
	if s.Err() {
		return s, buf, bs, rs, i
	}

	var brk Wideset
	fin := ASCIIControls.Wide().ExceptCopy("\t\n\r")

	if s.Is(Raw) {
		brk.Set(comma)
		brk.Set('\n')
		brk.Set('\r')
	} else {
		brk.Or(q)
	}

	stop := fin.OrCopy(brk)

	for i < len(b) {
		done := i

		s, rs, i = StringUntil(b, i, flags, s, rs, stop)
		bs += i - done
		if flags.Is(Decode) {
			buf = append(buf, b[done:i]...)
		}
		if s.Suppress(flags & StrErr).Err() {
			return s, buf, bs, rs, i
		}

		if i == len(b) || s.Is(Raw) {
			break
		}
		if !q.Is(b[i]) {
			return s | ErrSymbol, buf, bs, rs, i
		}

		if i+1 == len(b) || b[i] != b[i+1] {
			break
		}

		bs++
		rs++
		if flags.Is(Decode) {
			buf = append(buf, b[i])
		}

		i += 2
		s |= Escapes
	}

	if s.Suppress(flags & StrErr).Err() {
		return s, buf, bs, rs, i
	}
	if i < len(b) && fin.Is(b[i]) {
		return s | ErrSymbol, buf, bs, rs, i
	}

	s, i = StringClose(b, i, flags, s)
	if s.Err() {
		return s, buf, bs, rs, i
	}

	if bs == 0 && s.Is(Raw) {
		s &^= Raw
	}

	return s, buf, bs, rs, i
}

func caller(d int) string {
	_, file, line, _ := runtime.Caller(1 + d)

	return fmt.Sprintf("%v:%d", file, line)
}
