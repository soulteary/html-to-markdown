package main

import (
	"fmt"
	"testing"
)

func Example() {
	markdown, err := convert(exampleHTML)
	if err != nil {
		panic(err)
	}
	fmt.Println(markdown)
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
