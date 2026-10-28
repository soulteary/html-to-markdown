package main

import (
	"errors"
	"testing"
)

func Example() {
	main()
	// Output:
	// # Library first
	//
	// A **small** example with [a link](https://example.com).
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
