package skip

import (
	"unicode"
	"unicode/utf8"
)

func Equal(x, y []byte) bool {
	return len(x) == len(y) && len(x) == Common(x, y)
}

func EqualFold(x, y []byte) bool {
	return len(x) == len(y) && len(x) == CommonFold(x, y)
}

// Common finds common prefix length.
func Common(x, y []byte) int {
	i := 0

	for i < len(x) && i < len(y) && x[i] == y[i] {
		i++
	}

	return i
}

// CommonFold finds common prefix length case insensitive.
func CommonFold(x, y []byte) int {
	i := 0

	for i < len(x) && i < len(y) {
		if x[i] < 0x80 {
			if x[i] == y[i] {
				i++
				continue
			}

			xx := x[i] | 0x20
			yy := y[i] | 0x20

			if xx == yy && xx >= 'a' && xx <= 'z' {
				i++
				continue
			}
		} else {
			xx, xs := utf8.DecodeRune(x[i:])
			yy, ys := utf8.DecodeRune(y[i:])

			if xs != ys {
				return i
			}

			if unicode.ToLower(xx) != unicode.ToLower(yy) {
				return i
			}

			i += xs
			continue
		}

		break
	}

	return i
}

func PrefixAt(b, prefix []byte, i int) bool {
	return i+len(prefix) <= len(b) && Common(b[i:], prefix) == len(prefix)
}

func PrefixFoldAt(b, prefix []byte, i int) bool {
	return i+len(prefix) <= len(b) && CommonFold(b[i:], prefix) == len(prefix)
}

func Prefix(b, prefix []byte) bool {
	return Common(b, prefix) == len(prefix)
}

func PrefixFold(b, prefix []byte) bool {
	return CommonFold(b, prefix) == len(prefix)
}
