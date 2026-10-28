package main

import (
	"fmt"
	"log"

	htmltomarkdown "github.com/soulteary/html-to-markdown/v2"
)

const exampleHTML = `<h1>Library first</h1><p>A <strong>small</strong> example with <a href="https://example.com">a link</a>.</p>`

func convert(input string) (string, error) {
	return htmltomarkdown.ConvertString(input)
}

func main() {
	markdown, err := convert(exampleHTML)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(markdown)
}
