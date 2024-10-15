package skip

import (
	"strconv"
	"testing"
	"unicode/utf8"
)

func TestStr(tb *testing.T) {
	type TC struct {
		Flags   Str
		Want    Str
		In      string
		Res     string
		St, End int
		Len     int

		NoWrap bool
	}

	var buf []byte
	var fail TC
	var failj int

	for j, tc := range []TC{
		{Flags: Bqt | Quo | Sqt, Want: Bqt, End: -1, In: "`abc`", Res: `abc`},
		{Flags: Bqt | Quo | Sqt, Want: Quo, End: -1, In: `"abc"`, Res: `abc`},
		{Flags: Bqt | Quo | Sqt, Want: Sqt, End: -1, In: `'abc'`, Res: `abc`},

		{Flags: Bqt | Quo | Sqt, Want: Quo, End: -1, In: `"a\x16c"`, Res: "a\x16c"},

		{Flags: Quo, Want: ErrQuote, End: 0, In: "`abc`"},
		{Flags: Bqt, Want: ErrQuote, End: 0, In: `"abc"`},
		{Flags: Sqt, Want: ErrQuote, End: 0, In: `"abc"`},

		{Flags: Bqt | Quo | Sqt, Want: Bqt | ErrBuffer, End: 8, In: "`abc\"", Res: "abc\"\n\n\t"},
		{Flags: Bqt | Quo | Sqt, Want: Quo | ErrSymbol, End: 5, In: "\"abc`", Res: "abc`"},
		{Flags: Bqt | Quo | Sqt, Want: Quo | ErrSymbol, End: 5, In: "\"abc'", Res: "abc'"},

		{Flags: Bqt | Quo | Sqt, Want: Bqt, End: -1, In: "`ab\nc`", Res: "ab\nc"},
		{Flags: Bqt | Quo | Sqt, Want: Bqt, End: -1, In: "`a\"b\n\"c`", Res: "a\"b\n\"c"},
		{Flags: Bqt | Quo | Sqt, Want: Quo, End: -1, In: `"a\nb\tc"`, Res: "a\nb\tc"},
		{Flags: Bqt | Quo | Sqt, Want: Sqt, End: -1, In: `'a\nb\tc'`, Res: "a\nb\tc"},

		{Flags: Bqt | Quo | Sqt, Want: Bqt | Unicode, End: -1, In: "`абвгд`", Res: `абвгд`},
		{Flags: Bqt | Quo | Sqt, Want: Quo | Unicode, End: -1, In: `"абвгд"`, Res: `абвгд`},
		{Flags: Bqt | Quo | Sqt, Want: Sqt | Unicode, End: -1, In: `'абвгд'`, Res: `абвгд`},

		{Flags: Bqt | Quo | Sqt, Want: Quo, End: -1, In: `".\n.\t.\x20.\u0030.\U00000035."`, Res: ".\n.\t. .0.5."},
		{Flags: Bqt | Quo | Sqt, Want: Sqt, End: -1, In: `'.\n.\t.\x20.\u0030.\U00000035.'`, Res: ".\n.\t. .0.5."},

		{Flags: CSV | Quo | Sqt | Bqt | Raw | ',', Want: CSV | Raw, End: 1, In: `1`, Res: `1`},
		{Flags: CSV | Quo | Sqt | Bqt | Raw | ',', Want: CSV | Raw, End: 1, In: `1`, Res: `1`, NoWrap: true},

		{Flags: CSV | Quo | Sqt | Bqt | Raw | ',', Want: CSV | Quo, End: 12, In: `"abc""d""ef"`, Res: `abc"d"ef`},
		{Flags: CSV | Quo | Sqt | Bqt | Raw | ',', Want: CSV | Raw, End: 6, In: `abc ww`, Res: `abc ww`},

		{Flags: CSV | Quo | Sqt | Bqt | Raw | ',', Want: CSV | Quo, St: 6, End: 12, In: `"abc","def","qwe"`, Res: `def`},
		{Flags: CSV | Quo | Sqt | Bqt | Raw | ',', Want: CSV | Raw, St: 6, End: 14, In: `a b c, d e f ,q w e`, Res: ` d e f `},

		{Flags: CSV | Quo | Sqt | Bqt | Raw | ',', Want: CSV | Quo, St: 12, End: 17, In: `"abc","def","qwe"`, Res: `qwe`},
		{Flags: CSV | Quo | Sqt | Bqt | Raw | ',', Want: CSV | Raw, St: 14, End: 19, In: `a b c, d e f ,q w e`, Res: `q w e`},

		{Flags: Quo | Sqt | Bqt, Want: Quo, End: -1, In: `"%20%33"`, Res: `%20%33`},
		{Flags: URL, Want: URL | ErrSymbol, End: -1, In: `"%20%33"`, Res: `" 3"`},
		{Flags: URL, Want: URL, St: 8, End: 12, In: `abc=def&qw+e=asd&zxc`, Res: `qw e`, NoWrap: true},

		{Flags: Quo | ErrRune, Want: Quo, End: -1, In: `"\uD800\uDC00"`, Res: `𐀀`},
		{Flags: Quo | ErrRune, Want: Quo, End: -1, In: `"\uD80000DC00"`, Res: `�00DC00`},

		{Flags: Quo | Sqt, Want: Quo, End: -1, In: `"a\/b"`, Res: `a/b`},
	} {
		var pref []byte
		var in []byte

		if !tc.NoWrap {
			pref = []byte("\n\n\t")
		}

		in = append(pref, tc.In...)
		in = append(in, pref...)

		st := tc.St
		end := tc.End

		if end == -1 {
			end = len(tc.In)
		}

		st += len(pref)
		end += len(pref)
		ll := utf8.RuneCountInString(tc.Res)

		s, bs, rs, i := String(in, st, tc.Flags)
		assert(tb, s == tc.Want, "s %v (%#[1]v), wanted %v (%#[2]v)", s, tc.Want)
		assert(tb, i == end, "index %v, wanted %v  of %v", i, end, 2*len(pref)+len(tc.In))
		assert(tb, bs == len(tc.Res), "bytes %v, wanted %v", bs, len(tc.Res))
		assert(tb, rs == ll, "runes %v, wanted %v", rs, ll)

		if tb.Failed() {
			fail = tc
			failj = j
			break
		}

		s, buf, rs, i = DecodeString(in, st, tc.Flags, buf[:0])
		assert(tb, s == tc.Want, "s %v (%#[1]v), wanted %v (%#[2]v)", s, tc.Want)
		assert(tb, i == end, "index %v, wanted %v  of %v", i-len(pref), end, len(tc.In))
		assert(tb, rs == ll, "runes %v, wanted %v", rs, ll)
		assert(tb, Equal(buf, []byte(tc.Res)), "res %q", buf)

		if tb.Failed() {
			fail = tc
			failj = j
			break
		}
	}

	if tb.Failed() {
		tb.Logf("failed at %d, %#v", failj, fail)
	}
}

func TestStrContinue(tb *testing.T) {
	b := []byte(`"ab\u0030cd"`)
	exprs := 5

	var w []byte

	for st := 1; st < len(b); st++ {
		w = w[:0]

		s, w, rs0, i := DecodeString(b[:st], 0, Quo, w)
		tb.Logf("%#x: %[1]v", s)
		assert(tb, s.Is(ErrBuffer), "wanted error: %v", s)

		s, w, rs1, i := DecodeString(b, i, s|Continue, w)
		tb.Logf("%#x: %[1]v", s)
		assert(tb, !s.Err(), "didn't want error: %v", s)
		assert(tb, rs0+rs1 == exprs, "rune count %d + %d != %d  (%d: %s|%s)", rs0, rs1, exprs, st, b[:st], b[st:])

		if tb.Failed() {
			break
		}
	}
}

func TestIssue1(tb *testing.T) {
	data := `"The supermarket baby food aisle in the United States is packed with non-nutritious foods containing far too much sugar and salt and misleading marketing claims, a new study found.\n\nSixty percent of 651 foods that are marketed for children ages 6 months to 36 months on 10 supermarkets' shelves in the US failed to meet recommended World Health Organization nutritional guidelines for infant and toddler foods, according to the study, which was published this month in the peer-reviewed journal Nutrients.\n\nAlmost none of the foods met all of the WHO standards for advertising, which focus on clear labeling of ingredients and accurate health claims.\n\nRead more about the research at the link in our bio.\n\n📸: Maria Argutinskaya/iStockphoto/Getty Images"`

	exp, err := strconv.Unquote(data)
	assert(tb, err == nil, "unquote: %v", err)

	expl := utf8.RuneCountInString(exp)
	b := []byte(data)

	s, bs, rs, i := String(b, 0, Quo)
	assert(tb, !s.Err(), "error: %v", s)
	assert(tb, i == len(b), "i / len(data): %d / %d", i, len(data))
	assert(tb, bs == len([]byte(exp)), "len %v, wanted %v", rs, len([]byte(exp)))
	assert(tb, rs == expl, "len %v, wanted %v", rs, expl)

	s, buf, rs, i := DecodeString(b, 0, Quo, nil)
	assert(tb, !s.Err(), "error: %v", s)
	assert(tb, i == len(b), "i / len(data): %d / %d", i, len(data))
	assert(tb, rs == expl, "decoded: %s", buf)
	assert(tb, string(buf) == exp, "decoded: %s", buf)
}

func assert(tb testing.TB, ok bool, msg string, args ...any) bool {
	tb.Helper()

	if ok {
		return true
	}

	tb.Errorf(msg, args...)

	return false
}
