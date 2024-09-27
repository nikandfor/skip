package skip

import (
	"testing"
)

func TestStr(tb *testing.T) {
	type TC struct {
		Flags Str
		Want  Str
		In    string
		Res   string
		St, I int

		NoWrap bool
	}

	var buf []byte

	for j, tc := range []TC{
		{Flags: Raw | Quo | Sqt, Want: Raw, I: -1, In: "`abc`", Res: `abc`},
		{Flags: Raw | Quo | Sqt, Want: Quo, I: -1, In: `"abc"`, Res: `abc`},
		{Flags: Raw | Quo | Sqt, Want: Sqt, I: -1, In: `'abc'`, Res: `abc`},

		{Flags: Quo, Want: ErrQuote, I: 0, In: "`abc`"},
		{Flags: Raw, Want: ErrQuote, I: 0, In: `"abc"`},
		{Flags: Sqt, Want: ErrQuote, I: 0, In: `"abc"`},

		{Flags: Raw | Quo | Sqt, Want: ErrQuote, I: 8, In: "`abc\"", Res: "abc\"\n\n\t"},
		{Flags: Raw | Quo | Sqt, Want: ErrChar, I: 5, In: "\"abc`"},
		{Flags: Raw | Quo | Sqt, Want: ErrChar, I: 5, In: "\"abc'"},

		{Flags: Raw | Quo | Sqt, Want: Raw, I: -1, In: "`ab\nc`", Res: "ab\nc"},
		{Flags: Raw | Quo | Sqt, Want: Raw, I: -1, In: "`a\"b\n\"c`", Res: "a\"b\n\"c"},
		{Flags: Raw | Quo | Sqt, Want: Quo, I: -1, In: `"a\nb\tc"`, Res: "a\nb\tc"},
		{Flags: Raw | Quo | Sqt, Want: Sqt, I: -1, In: `'a\nb\tc'`, Res: "a\nb\tc"},

		{Flags: Raw | Quo | Sqt, Want: Raw, I: -1, In: "`абвгд`", Res: `абвгд`},
		{Flags: Raw | Quo | Sqt, Want: Quo, I: -1, In: `"абвгд"`, Res: `абвгд`},
		{Flags: Raw | Quo | Sqt, Want: Sqt, I: -1, In: `'абвгд'`, Res: `абвгд`},

		{Flags: Raw | Quo | Sqt, Want: Quo, I: -1, In: `".\n.\t.\x20.\u0030.\U00000035."`, Res: ".\n.\t. .0.5."},
		{Flags: Raw | Quo | Sqt, Want: Sqt, I: -1, In: `'.\n.\t.\x20.\u0030.\U00000035.'`, Res: ".\n.\t. .0.5."},

		{Flags: CSV | Quo | Sqt | Raw | ',', Want: CSV | Raw, I: 1, In: `1`, Res: `1`},
		{Flags: CSV | Quo | Sqt | Raw | ',', Want: CSV | Raw, I: 1, In: `1`, Res: `1`, NoWrap: true},

		{Flags: CSV | Quo | Sqt | Raw | ',', Want: CSV | Quo, I: 12, In: `"abc""d""ef"`, Res: `abc"d"ef`},
		{Flags: CSV | Quo | Sqt | Raw | ',', Want: CSV | Raw, I: 6, In: `abc ww`, Res: `abc ww`},

		{Flags: CSV | Quo | Sqt | Raw | ',', Want: CSV | Quo, St: 6, I: 12, In: `"abc","def","qwe"`, Res: `def`},
		{Flags: CSV | Quo | Sqt | Raw | ',', Want: CSV | Raw, St: 6, I: 14, In: `a b c, d e f ,q w e`, Res: ` d e f `},

		{Flags: CSV | Quo | Sqt | Raw | ',', Want: CSV | Quo, St: 12, I: 17, In: `"abc","def","qwe"`, Res: `qwe`},
		{Flags: CSV | Quo | Sqt | Raw | ',', Want: CSV | Raw, St: 14, I: 19, In: `a b c, d e f ,q w e`, Res: `q w e`},
	} {
		var pref []byte
		var in []byte

		if !tc.NoWrap {
			pref = []byte("\n\n\t")
		}

		in = append(pref, tc.In...)
		in = append(in, pref...)

		st := tc.St
		tci := tc.I

		if tci == -1 {
			tci = len(tc.In)
		}

		st += len(pref)
		tci += len(pref)

		s, l, i := String(in, st, tc.Flags)
		assert(tb, s == tc.Want, "s %v, wanted %v", s, tc.Want)
		assert(tb, i == tci, "index %v, wanted %v  of %v", i-len(pref), tc.I, len(tc.In))
		assert(tb, l == len(tc.Res), "len %v, wanted %v", l, len(tc.Res))

		s, buf, i = DecodeString(in, st, tc.Flags, buf[:0])
		assert(tb, s == tc.Want, "s %v, wanted %v", s, tc.Want)
		assert(tb, i == tci, "index %v, wanted %v  of %v", i-len(pref), tc.I, len(tc.In))
		assert(tb, Equal(buf, []byte(tc.Res)), "res %q", buf)

		if tb.Failed() {
			tb.Logf("failed at %d, %#v", j, tc)
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
