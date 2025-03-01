package skip

import (
	"fmt"
	"unicode"
	"unicode/utf8"
)

type (
	ID int
)

const (
	// first byte is arg

	IDPrefix  ID = 1 << (8 + iota)
	IDPrivate    // lowercase first letter
	IDPublic     // uppercase first letter
	IDUnicode    // unicode

	_
	_
	_
	_

	IDErrPrefix // prefix doesn't match
	IDErrSymbol // improper symbol
	IDErrRune   // malformed rune
	IDErrIndex  // index out of bounds

	IDErrBuffer // short buffer

	IDErr = IDPrefix | IDErrSymbol | IDErrRune | IDErrIndex | IDErrBuffer
)

// Identifier validates and finds the end of an identifier.
// Identifier do not currently accept any flags.
func Identifier(b []byte, st int, flags ID) (x ID, i int) {
	i = st

	if i == len(b) {
		return x | IDErrIndex, st
	}

	if flags.Is(IDPrefix) {
		if b[i] != byte(flags) {
			return x | IDErrPrefix, st
		}

		i++
	}

	if i == len(b) {
		return x | IDErrBuffer, st
	}

	if b[i] < utf8.RuneSelf {
		if !IDFirst.Is(b[i]) {
			return x | IDErrSymbol, i
		}

		if Upper.Is(b[i]) {
			x |= IDPublic
		} else {
			x |= IDPrivate
		}

		i++
	} else {
		r, size := utf8.DecodeRune(b[i:])
		if size == 1 && r == utf8.RuneError {
			return x | IDErrRune, i
		}
		if !unicode.IsLetter(r) {
			return x | IDErrSymbol, i
		}

		if unicode.IsUpper(r) {
			x |= IDPublic
		} else {
			x |= IDPrivate
		}

		x |= IDUnicode
		i += size
	}

	for i < len(b) {
		if b[i] < utf8.RuneSelf {
			if !IDRest.Is(b[i]) {
				return x, i
			}

			i++

			continue
		}

		r, size := utf8.DecodeRune(b[i:])
		if size == 1 && r == utf8.RuneError {
			return x | IDErrRune, i
		}
		if !unicode.IsLetter(r) && !unicode.IsNumber(r) { // _ is < utf8.RuneSelf
			return x, i
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

func (id ID) Error() string {
	if !id.Err() {
		return "ok"
	}

	r := ""
	comma := false

	add := func(e ID, t string) {
		if !id.Is(e) {
			return
		}

		r += csel(comma, ", ", "")
		r += t
		comma = true
	}

	add(IDErrSymbol, "bad symbol")
	add(IDErrRune, "bad rune")

	if r == "" {
		r = fmt.Sprintf("%#x", int(id))
	}

	return r
}
