package skip

func Equal(x, y []byte) bool {
	return len(x) == len(y) && len(x) == Match(x, y)
}

func EqualFold(x, y []byte) bool {
	return len(x) == len(y) && len(x) == MatchFold(x, y)
}

func Match(x, y []byte) int {
	i := 0

	for i < len(x) && i < len(y) && x[i] == y[i] {
		i++
	}

	return i
}

func MatchFold(x, y []byte) int {
	i := 0

	for i < len(x) && i < len(y) {
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

		break
	}

	return i
}

func PrefixAt(b, prefix []byte, i int) bool {
	return i+len(prefix) <= len(b) && Match(b[i:], prefix) == len(prefix)
}

func PrefixFoldAt(b, prefix []byte, i int) bool {
	return i+len(prefix) <= len(b) && MatchFold(b[i:], prefix) == len(prefix)
}

func Prefix(b, prefix []byte) bool {
	return Match(b, prefix) == len(prefix)
}

func PrefixFold(b, prefix []byte) bool {
	return MatchFold(b, prefix) == len(prefix)
}
