package skip

type (
	Num int
)

const (
	Neg Num = 1 << iota // number is negative
	NaN                 // it's NaN
	Inf                 // it's Inf or -Inf
	Int                 // It's integer
	Flt                 // It's float

	Bin // integer is in binary format: 0bxxxx_xxxx
	Oct // integer is in octal format: 0oxxx or 0xxx
	Hex // integer or hex is in hex format. float hex format implies 'p' exponent form

	ExpNeg // negative exponent
	Exp    // there is exponent part

	NumOk = Int | Flt | NaN | Inf
	Base  = Bin | Oct | Hex
)

// Number validates and finds the end of a rumber.
// Number do not currently accept any flags.
func Number(b []byte, st int, flags Num) (n Num, i int) {
	if st >= len(b) {
		return
	}

	if n, i = Float(b, st, flags); n != 0 {
		return
	}

	if n, i = Integer(b, st, flags); n != 0 {
		return
	}

	return 0, st
}

func Float(b []byte, st int, flags Num) (n Num, i int) {
	//	defer func() { log.Printf("Float    %d %q -> %d  => %v  from %v", st, b[st:], i, n, loc.Caller(1)) }()
	if n, i = InfNaN(b, st, flags); n != 0 {
		return
	}

	i = st
	i, n = skipSign(b, i, n, Neg)

	set := Decimals.Wide()
	exp := byte('e')
	dot := false

	switch {
	case i == len(b):
		return 0, st
	case b[i] == '0' && i+1 < len(b) && (b[i+1] == 'x' || b[i+1] == 'X'):
		n |= Hex
		set = Hexes
		exp = 'p'
		i += 2
	case b[i] >= '0' && b[i] <= '9':
		//	n |= Int
	case b[i] == '.':
		dot = true
	default:
		return 0, st
	}

	var ok, ok2 bool

	i, ok = skipDigits(b, i, set, n.Is(Hex))

	if i < len(b) && b[i] == '.' {
		i++
		i, ok2 = skipDigits(b, i, set, false)
		dot = true
	}

	if !(ok || ok2) {
		return 0, st
	}

	if i == len(b) || b[i] != exp && b[i] != exp-0x20 {
		if dot {
			n |= Flt
		} else {
			n |= Int
		}

		return n, i
	}
	i++

	i, n = skipSign(b, i, n, ExpNeg)

	i, ok = skipDigits(b, i, Decimals.Wide(), false)
	if !ok {
		return 0, st
	}

	return n | Flt | Exp, i
}

func Integer(b []byte, st int, flags Num) (n Num, i int) {
	//	defer func() { log.Printf("Int      %d %q -> %d  => %v  from %v", st, b[st:], i, n, loc.Caller(1)) }()
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
	default:
		return 0, st
	}

	i, ok := skipDigits(b, i, set, n != 0)
	if !ok {
		return n, st
	}

	return n | Int, i
}

func InfNaN(b []byte, st int, flags Num) (n Num, i int) {
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

func (n Num) Ok() bool {
	return n&NumOk != 0
}

func (n Num) Is(f Num) bool {
	return n&f == f
}

func (n Num) Any(f Num) bool {
	return n&f != 0
}
