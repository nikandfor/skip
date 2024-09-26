package skip

func Equal(x, y []byte) int {
	i := 0

	for i < len(x) && i < len(y) && x[i] == y[i] {
		i++
	}

	return i
}

func EqualFold(x, y []byte) int {
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
		}
	}

	return i
}

func PrefixAt(b, prefix []byte, i int) bool {
	return i+len(prefix) <= len(b) && Equal(b[i:], prefix) == len(prefix)
}

func PrefixFoldAt(b, prefix []byte, i int) bool {
	return i+len(prefix) <= len(b) && EqualFold(b[i:], prefix) == len(prefix)
}

func Prefix(b, prefix []byte) bool {
	return Equal(b, prefix) == len(prefix)
}

func PrefixFold(b, prefix []byte) bool {
	return EqualFold(b, prefix) == len(prefix)
}
