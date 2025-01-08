package skip_test

import (
	"net/url"
	"testing"
	"unicode/utf8"

	"nikand.dev/go/skip"
)

func TestURL(tb *testing.T) {
	type TC struct {
		Want skip.Str
		In   string
	}

	fail := -1
	tcs := []TC{
		{In: "abc=def"},
		{In: "https://example.com/some/path?abc=def&123"},
		{In: "абс=деф", Want: skip.Unicode},
		{In: "a%62c=d%65f", Want: skip.Escapes},
	}

	for j, tc := range tcs {
		wrap := []byte("\n\t")
		in := []byte(tc.In)
		in = append(wrap, in...)
		in = append(in, wrap...)
		st := len(wrap)

		dec, err := url.QueryUnescape(tc.In)
		if !assert(tb, err == nil, "unquote: %v", err) {
			break
		}

		runes := utf8.RuneCountInString(dec)

		s, bs, rs, i := skip.URL(in, st, 0)
		assert(tb, s == tc.Want, "s %v (%#[1]v), wanted %v (%#[2]v)", s, tc.Want)
		assert(tb, i == len(in)-len(wrap), "i %d, wanted %d", i-len(wrap), len(tc.In))
		assert(tb, bs == len(dec), "bytes %v, wanted %v", bs, len(dec))
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

func TestDecodeURLQuery(tb *testing.T) {
	type TC struct {
		Flags   skip.Str
		Want    skip.Str
		In      string
		Out     string
		St, End int
		Value   bool
	}

	fail := -1
	tcs := []TC{
		{In: "abc=def", Out: "abc", End: 3},
		{In: "abc=def", Out: "", St: 3, End: 3},
		{In: "abc=def", Out: "def", St: 4, Flags: skip.URLValue},
		{In: "abc=def&", Out: "def", St: 4, End: -1, Flags: skip.URLValue},
		{In: "abc=def#", Out: "def", St: 4, End: -1, Flags: skip.URLValue},
		{In: "abc=def[", Out: "def[", St: 4, Flags: skip.URLValue},
		{In: "a%62c=d%65f#frag", Out: "def", St: 6, End: -5, Want: skip.Escapes},
		{In: "abc=def=123&", Out: "def=123", St: 4, End: -1, Flags: skip.URLValue},
	}

	for j, tc := range tcs {
		wrap := []byte("\n\t")
		in := []byte(tc.In)
		in = append(wrap, in...)
		in = append(in, wrap...)
		st := len(wrap) + tc.St
		end := csel(tc.End > 0, tc.End, len(tc.In)+tc.End)

		runes := utf8.RuneCountInString(tc.Out)

		s, res, rs, i := skip.DecodeURLQuery(in, st, tc.Flags, nil)
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

func assert(tb testing.TB, ok bool, msg string, args ...any) bool {
	tb.Helper()

	if ok {
		return true
	}

	tb.Errorf(msg, args...)

	return false
}

func csel[T any](c bool, x, y T) T {
	if c {
		return x
	}

	return y
}
