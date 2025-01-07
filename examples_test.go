//go:build ignore

package skip_test

import (
	"fmt"

	"nikand.dev/go/skip"
)

func ExampleString() {
	data := []byte(`abc "some\nstring\x3a\u0020val" // something`)

	i := 0

	// skip until the string beginning
	i = skip.WordSymbols.OrCopy(skip.Whitespaces.Wide()).Skip(data, i)

	flags := skip.Quo | skip.Sqt // which strings we accept: " or ' quoted. Python style

	var buf []byte // buffer for decoded string

	s, buf, rc, end := skip.DecodeString(data, i, flags, buf[:0])
	if s.Err() {
		fmt.Printf("error at %d: %v\n", end, s)
		return
	}

	// raw data including quotes
	_ = data[i:end]

	fmt.Printf("string of %d runes and %d bytes (double quoted: %v, single quoted: %v):\n%s", rc, len(buf), s.Is(skip.Quo), s.Is(skip.Sqt), buf)

	// Output:
	// string of 16 runes and 16 bytes (double quoted: true, single quoted: false):
	// some
	// string: val
}

func ExampleNumber() {
	data := []byte(`  -1.456 `)

	i := 0
	i = skip.Spaces(data, i) // skip to the number beginning

	n, end := skip.Number(data, i, 0)
	if !n.Ok() {
		fmt.Printf("not a number or incorrectly formatted\n")
		return
	}

	fmt.Printf("number (is negative: %v, is int: %v, is float: %v, with exponent: %v): %s\n", n.Is(skip.Neg), n.Is(skip.Int), n.Is(skip.Flt), n.Is(skip.Exp), data[i:end])

	// Output:
	// number (is negative: true, is int: false, is float: true, with exponent: false): -1.456
}

func ExampleIdentifier() {
	data := []byte(`var UnicodeVar_世界 = 3`)

	i := 0
	i = skip.WordSymbols.Skip(data, i) // skip var
	i = skip.Spaces(data, i)           // skip spaces

	id, end := skip.Identifier(data, i, 0)
	if id.Err() {
		fmt.Printf("error at %d: %v\n", end, id)
		return
	}

	fmt.Printf("id (is public: %v, is private: %v, with unicode: %v): %s\n", id.Is(skip.IDPublic), id.Is(skip.IDPrivate), id.Is(skip.IDUnicode), data[i:end])

	// Output:
	// id (is public: true, is private: false, with unicode: true): UnicodeVar_世界
}
