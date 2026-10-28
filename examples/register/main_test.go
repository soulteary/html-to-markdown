package main

import (
	"errors"
	"testing"
)

func Example() {
	main()
	// Output:
	// ## Go ★★★★
}

func TestInvalidRatingsPreserveFallback(t *testing.T) {
	for _, count := range []string{"", "invalid", "-1", "6"} {
		markdown, err := convert(`<star-rating count="` + count + `">unrated</star-rating>`)
		if err != nil {
			t.Fatal(err)
		}
		if markdown != "unrated" {
			t.Fatalf("count %q: %q", count, markdown)
		}
	}
}

func TestMustPreservesConversionError(t *testing.T) {
	expected := errors.New("conversion failed")
	defer func() {
		if got := recover(); got != expected {
			t.Fatalf("panic: %v; want original conversion error", got)
		}
	}()
	must("", expected)
}
