package skip

import (
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
		{Flags: Raw | Quo | Sqt, Want: Raw, End: -1, In: "`abc`", Res: `abc`},
		{Flags: Raw | Quo | Sqt, Want: Quo, End: -1, In: `"abc"`, Res: `abc`},
		{Flags: Raw | Quo | Sqt, Want: Sqt, End: -1, In: `'abc'`, Res: `abc`},

		{Flags: Raw | Quo | Sqt, Want: Quo, End: -1, In: `"a\x16c"`, Res: "a\x16c"},

		{Flags: Quo, Want: ErrQuote, End: 0, In: "`abc`"},
		{Flags: Raw, Want: ErrQuote, End: 0, In: `"abc"`},
		{Flags: Sqt, Want: ErrQuote, End: 0, In: `"abc"`},

		{Flags: Raw | Quo | Sqt, Want: Raw | ErrBuffer, End: 8, In: "`abc\"", Res: "abc\"\n\n\t"},
		{Flags: Raw | Quo | Sqt, Want: Quo | ErrChar, End: 5, In: "\"abc`", Res: "abc`"},
		{Flags: Raw | Quo | Sqt, Want: Quo | ErrChar, End: 5, In: "\"abc'", Res: "abc'"},

		{Flags: Raw | Quo | Sqt, Want: Raw, End: -1, In: "`ab\nc`", Res: "ab\nc"},
		{Flags: Raw | Quo | Sqt, Want: Raw, End: -1, In: "`a\"b\n\"c`", Res: "a\"b\n\"c"},
		{Flags: Raw | Quo | Sqt, Want: Quo, End: -1, In: `"a\nb\tc"`, Res: "a\nb\tc"},
		{Flags: Raw | Quo | Sqt, Want: Sqt, End: -1, In: `'a\nb\tc'`, Res: "a\nb\tc"},

		{Flags: Raw | Quo | Sqt, Want: Raw, End: -1, In: "`абвгд`", Res: `абвгд`},
		{Flags: Raw | Quo | Sqt, Want: Quo, End: -1, In: `"абвгд"`, Res: `абвгд`},
		{Flags: Raw | Quo | Sqt, Want: Sqt, End: -1, In: `'абвгд'`, Res: `абвгд`},

		{Flags: Raw | Quo | Sqt, Want: Quo, End: -1, In: `".\n.\t.\x20.\u0030.\U00000035."`, Res: ".\n.\t. .0.5."},
		{Flags: Raw | Quo | Sqt, Want: Sqt, End: -1, In: `'.\n.\t.\x20.\u0030.\U00000035.'`, Res: ".\n.\t. .0.5."},

		{Flags: CSV | Quo | Sqt | Raw | ',', Want: CSV | Raw, End: 1, In: `1`, Res: `1`},
		{Flags: CSV | Quo | Sqt | Raw | ',', Want: CSV | Raw, End: 1, In: `1`, Res: `1`, NoWrap: true},

		{Flags: CSV | Quo | Sqt | Raw | ',', Want: CSV | Quo, End: 12, In: `"abc""d""ef"`, Res: `abc"d"ef`},
		{Flags: CSV | Quo | Sqt | Raw | ',', Want: CSV | Raw, End: 6, In: `abc ww`, Res: `abc ww`},

		{Flags: CSV | Quo | Sqt | Raw | ',', Want: CSV | Quo, St: 6, End: 12, In: `"abc","def","qwe"`, Res: `def`},
		{Flags: CSV | Quo | Sqt | Raw | ',', Want: CSV | Raw, St: 6, End: 14, In: `a b c, d e f ,q w e`, Res: ` d e f `},

		{Flags: CSV | Quo | Sqt | Raw | ',', Want: CSV | Quo, St: 12, End: 17, In: `"abc","def","qwe"`, Res: `qwe`},
		{Flags: CSV | Quo | Sqt | Raw | ',', Want: CSV | Raw, St: 14, End: 19, In: `a b c, d e f ,q w e`, Res: `q w e`},

		{Flags: Quo | Sqt | Raw, Want: Quo, End: -1, In: `"%20%33"`, Res: `%20%33`},
		{Flags: URL, Want: URL | ErrChar, End: -1, In: `"%20%33"`, Res: `" 3"`},
		{Flags: URL, Want: URL, St: 8, End: 12, In: `abc=def&qw+e=asd&zxc`, Res: `qw e`, NoWrap: true},

		{Flags: Quo | ErrRune, Want: Quo, End: -1, In: `"\uD800\uDC00"`, Res: `𐀀`},
		{Flags: Quo | ErrRune, Want: Quo, End: -1, In: `"\uD80000DC00"`, Res: `�00DC00`},
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

		s, l, i := String(in, st, tc.Flags)
		assert(tb, s == tc.Want, "s %#v, wanted %#v", s, tc.Want)
		assert(tb, i == end, "index %v, wanted %v  of %v", i, end, 2*len(pref)+len(tc.In))
		assert(tb, l == ll, "len %v, wanted %v", l, ll)

		if tb.Failed() {
			fail = tc
			failj = j
			break
		}

		s, buf, i = DecodeString(in, st, tc.Flags, buf[:0])
		assert(tb, s == tc.Want, "s %v, wanted %v", s, tc.Want)
		assert(tb, i == end, "index %v, wanted %v  of %v", i-len(pref), end, len(tc.In))
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

	var w []byte

	for st := 1; st < len(b); st++ {
		w = w[:0]

		s, w, i := DecodeString(b[:st], 0, Quo, w)
		tb.Logf("%#x: %[1]v", s)
		assert(tb, s.Is(ErrBuffer), "wanted error: %v", s)

		s, w, i = DecodeString(b, i, s|Continue, w)
		tb.Logf("%#x: %[1]v", s)
		assert(tb, !s.Err(), "didn't want error: %v", s)

		if tb.Failed() {
			break
		}
	}
}

func assert(tb testing.TB, ok bool, msg string, args ...any) bool {
	tb.Helper()

	if ok {
		return true
	}

	tb.Errorf(msg, args...)

	return false
}
