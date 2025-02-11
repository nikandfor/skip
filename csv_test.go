package skip_test

import (
	"testing"
	"unicode/utf8"

	"nikand.dev/go/skip"
)

func TestCSV(tb *testing.T) {
	type TC struct {
		Flags   skip.Str
		Want    skip.Str
		In      string
		Out     string
		St, End int
		Comma   byte
	}

	fail := -1
	tcs := []TC{
		{In: `1`, Out: "1", Want: skip.Raw},
		{In: `abc`, Out: "abc", Want: skip.Raw},
		{In: `"abc"`, Out: "abc", Want: skip.Dqt},
		{In: `"a""bc"`, Out: `a"bc`, Want: skip.Dqt | skip.Escapes},
		{In: `a"bc`, Out: `a"bc`, Want: skip.Raw},
		{In: "abc\x11", Out: `abc`, End: -1, Want: skip.Raw | skip.ErrSymbol},
		{In: `1,2`, Out: "1", End: 1, Want: skip.Raw},
		{In: `abc,cde`, Out: "abc", End: 3, Want: skip.Raw},
		{In: `1,2`, Out: "", St: 1, End: 1, Want: 0},
		{In: `1`, Out: "", St: 1, End: 1, Want: 0},
		{In: `""`, Out: "", Want: skip.Dqt},
		{In: `1|2`, Out: "1", End: 1, Want: skip.Raw, Comma: '|'},
	}

	for j, tc := range tcs {
		wrap := []byte("\n\t")
		in := []byte(tc.In)
		in = append(wrap, in...)
		in = append(in, wrap...)
		st := len(wrap) + tc.St
		end := csel(tc.End > 0, tc.End, len(tc.In)+tc.End)
		comma := csel(tc.Comma != 0, tc.Comma, ',')

		runes := utf8.RuneCountInString(tc.Out)

		s, bs, rs, i := skip.CSV(in, st, tc.Flags, comma)
		assert(tb, s == tc.Want, "s %v (%#[1]v), wanted %v (%#[2]v)", s, tc.Want)
		assert(tb, i == len(wrap)+end, "i %d, wanted %d", i-len(wrap), end)
		assert(tb, bs == len(tc.Out), "bytes %d, wanted %d", bs, len(tc.Out))
		assert(tb, rs == runes, "runes %v, wanted %v", rs, runes)

		s, res, rs, i := skip.DecodeCSV(in, st, tc.Flags, comma, nil)
		assert(tb, s == tc.Want, "s %v (%#[1]v), wanted %v (%#[2]v)", s, tc.Want)
		assert(tb, i == len(wrap)+end, "i %d, wanted %d", i-len(wrap), end)
		assert(tb, string(res) == tc.Out, "bytes %s, wanted %s", res, tc.Out)
		assert(tb, rs == runes, "runes %v, wanted %v", rs, runes)

		if tb.Failed() {
			fail = j
			break
		}
	}

	if tb.Failed() {
		tb.Logf("failed at #%d, %#v", fail, tcs[fail])
	}
}
