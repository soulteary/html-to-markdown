package main

import "fmt"

func Example() {
	markdown, err := convert(exampleHTML)
	if err != nil {
		panic(err)
	}
	fmt.Println(markdown)
	// Output:
	// # Library first
	//
	// A **small** example with [a link](https://example.com).
}
