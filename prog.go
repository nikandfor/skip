package skip

import (
	"unicode"
	"unicode/utf8"
)

type (
	ID int
)

const (
	IDOk ID = 1 << iota

	Private
	Public

	Unicode
)

func Identifier(b []byte, st int) (x ID, i int) {
	i = st

	if i == len(b) || (b[i] >= '0' && b[i] <= '9') {
		return 0, i
	}

	if Upper.Is(b[i]) {
		x |= Public
	} else {
		x |= Private
	}

	for i = st; i < len(b); {
		if IDRest.Is(b[i]) {
			i++
			continue
		}

		r, size := utf8.DecodeRune(b[i:])
		if r == utf8.RuneError {
			return 0, st
		}
		if !unicode.IsSymbol(r) {
			return 0, st
		}

		x |= Unicode
		i += size
	}

	return x, i
}

func (id ID) Ok() bool {
	return id&IDOk != 0
}

func (id ID) Err() bool {
	return id == 0
}

func (id ID) Is(f ID) bool {
	return id&f == f
}

func (id ID) Any(f ID) bool {
	return id&f != 0
}
