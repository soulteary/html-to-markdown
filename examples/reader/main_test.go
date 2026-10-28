package main

import "fmt"

func Example() {
	markdown, err := convert(exampleHTML)
	if err != nil {
		panic(err)
	}
	fmt.Println(markdown)
	// Output:
	// [Guide](https://example.com/docs/guide) ![Logo](https://example.com/logo.png)
}
