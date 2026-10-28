package main

import "fmt"

func Example() {
	markdown, err := convert(exampleHTML)
	if err != nil {
		panic(err)
	}
	fmt.Println(markdown)
	// Output:
	// __Custom formatting__
	//
	// * First
	// * Second
}
