package main

import (
	"fmt"

	htmltomarkdown "github.com/soulteary/html-to-markdown/v2"
)

const exampleHTML = `<h1>Library first</h1><p>A <strong>small</strong> example with <a href="https://example.com">a link</a>.</p>`

func convert(input string) (string, error) {
	return htmltomarkdown.ConvertString(input)
}

func main() {
	fmt.Println(must(convert(exampleHTML)))
}

func must(markdown string, err error) string {
	if err != nil {
		panic(err)
	}
	return markdown
}
