package skip

type (
	Num int
)

const (
	Neg Num = 1 << iota
	NaN
	Inf

	Int
	Flt

	Bin
	Oct
	Hex

	ExpNeg
	Exp

	emb Num = iota

	Base         = Bin | Oct | Hex
	Embedded Num = 1 << emb
)

func Number(b []byte, st int) (n Num, i int) {
	if st >= len(b) {
		return
	}

	if n, i = InfNaN(b, st); n != 0 {
		return
	}

	if n, i = Integer(b, st); n != 0 {
		return
	}

	if n, i = Float(b, st); n != 0 {
		return
	}

	return 0, st
}

func Float(b []byte, st int) (n Num, i int) {
	i = st

	i, n = skipSign(b, i, n, Neg)

	set := Decimals.Wide()
	exp := byte('e')

	switch {
	case i == len(b):
		return 0, st
	case b[i] == '0' && i+1 < len(b) && (b[i] == 'x' || b[i] == 'X'):
		n |= Hex
		set = Hexes
		exp = 'p'
		i += 2
	case b[i] >= '0' && b[i] <= '9':
		n |= Int
	default:
		return 0, st
	}

	var ok, dot, ok2 bool

	i, ok = skipDigits(b, i, set, n.Is(Hex))

	if i < len(b) && b[i] == '.' {
		i++
		i, ok = skipDigits(b, i, set, false)
		dot = true
	} else {
		ok2 = true
	}

	if !ok && !ok2 {
		return 0, st
	}

	if i == len(b) || b[i] != exp && b[i] != exp-0x20 {
		if exp == 'e' && dot {
			return n | Flt, i
		}

		return 0, st
	}
	i++

	i, n = skipSign(b, i, n, ExpNeg)

	i, ok = skipDigits(b, i, Decimals.Wide(), false)
	if !ok {
		return 0, st
	}

	return n | Flt | Exp, i
}

func Integer(b []byte, st int) (n Num, i int) {
	i = st

	i, n = skipSign(b, i, n, Neg)

	set := Decimals.Wide()

	switch {
	case i == len(b):
		return 0, st
	case b[i] == '0':
		i++

		if i == len(b) {
			return Int, i
		}

		switch {
		case b[i] == 'x', b[i] == 'X':
			n |= Hex
			set = Hexes
			i++
		case b[i] == 'b', b[i] == 'B':
			n |= Bin
			set = Binaries.Wide()
			i++
		case b[i] == 'o', b[i] == 'O':
			i++
			fallthrough
		case b[i] == '_' || b[i] >= '0' && b[i] < '8':
			n |= Oct
			set = Octals.Wide()
		default:
			return 0, st
		}
	case b[i] >= '1' && b[i] <= '9':
		i++
	default:
		return 0, st
	}

	i, ok := skipDigits(b, i, set, true)
	if !ok {
		return n, st
	}

	return n | Int, i
}

func InfNaN(b []byte, st int) (n Num, i int) {
	i = st

	if PrefixFold(b[i:], []byte("nan")) {
		return NaN, i + 3
	}

	i, n = skipSign(b, i, n, Neg)

	if PrefixFoldAt(b, []byte("inf"), i) {
		i += 3

		if PrefixFoldAt(b, []byte("inity"), i) {
			i += 5

			if i < len(b) && IDRest.Is(b[i]) {
				return 0, st
			}

			return n | Inf, i
		}

		if i < len(b) && IDRest.Is(b[i]) {
			return 0, st
		}

		return n | Inf, i
	}

	return 0, i
}

func (n Num) Is(x Num) bool {
	return n&x == x
}

func (n Num) Any(x Num) bool {
	return n&x != 0
}

func skipDigits(b []byte, i int, set Wideset, pad bool) (_ int, ok bool) {
	for i < len(b) {
		if pad && b[i] == '_' && i+1 < len(b) && set.Is(b[i+1]) {
			i += 2
		} else if set.Is(b[i]) {
			i++
		} else {
			break
		}

		pad = true
		ok = true
	}

	return i, ok
}

func skipSign(b []byte, i int, n, neg Num) (int, Num) {
	if i < len(b) && (b[i] == '-' || b[i] == '+') {
		if b[i] == '-' {
			n ^= neg
		}

		i++
	}

	return i, n
}
