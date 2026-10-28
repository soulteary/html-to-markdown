package main

import (
	"errors"
	"testing"
)

func Example() {
	main()
	// Output:
	// __Custom formatting__
	//
	// * First
	// * Second
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
