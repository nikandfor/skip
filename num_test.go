package skip

import (
	"testing"
)

func TestNum(tb *testing.T) {
	for _, tc := range []string{
		"1",
		"10",
		"1.1",
		"1.",
		".1",
		"1e10",
		"0x1e10",
		"1e+10",
		"0x1ffp-10",
		"nan",
		"inf",
		"-infinity",
	} {
		n, i := Number([]byte(tc), 0)
		if !n.Ok() || i != len(tc) {
			tb.Errorf("%s => %v %v", tc, n, i)
		}
	}
}
