package skip

import (
	"unicode"
	"unicode/utf8"
)

type (
	ID int
)

const (
	_         ID = 1 << iota
	IDPrivate    // lowercase first letter
	IDPublic     // uppercase first letter
	IDUnicode    // unicode

	IDErrSymbol // improper symbol
	IDErrRune   // malformed rune

	IDErr = IDErrRune | IDErrSymbol
)

// Identifier validates and finds the end of an identifier.
// Identifier do not currently accept any flags.
func Identifier(b []byte, st int) (x ID, i int) {
	i = st

	if i == len(b) || (b[i] >= '0' && b[i] <= '9') {
		return x | IDErrSymbol, i
	}

	if b[i] < 0x80 {
		if Upper.Is(b[i]) {
			x |= IDPublic
		} else {
			x |= IDPrivate
		}

		i++
	} else {
		r, s := utf8.DecodeRune(b[i:])
		if s == 1 && r == utf8.RuneError {
			return x | IDErrRune, i
		}
		if unicode.IsUpper(r) {
			x |= IDPublic
		} else {
			x |= IDPrivate
		}

		x |= IDUnicode

		i += s
	}

	for i < len(b) {
		if b[i] < utf8.RuneSelf {
			if !IDRest.Is(b[i]) {
				return x | IDErrSymbol, i
			}

			i++
			continue
		}

		r, size := utf8.DecodeRune(b[i:])
		if r == utf8.RuneError {
			return x | IDErrRune, i
		}
		if !unicode.IsSymbol(r) {
			return x | IDErrSymbol, i
		}

		x |= IDUnicode
		i += size
	}

	return x, i
}

func (id ID) Err() bool {
	return id&IDErr != 0
}

func (id ID) Is(f ID) bool {
	return id&f == f
}

func (id ID) Any(f ID) bool {
	return id&f != 0
}
