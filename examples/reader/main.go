package main

import (
	"fmt"
	"log"
	"strings"

	htmltomarkdown "github.com/soulteary/html-to-markdown/v2"
	"github.com/soulteary/html-to-markdown/v2/converter"
)

const exampleHTML = `<p><a href="guide">Guide</a> <img src="../logo.png" alt="Logo"></p>`

func convert(input string) (string, error) {
	// ConvertReader also accepts files, HTTP response bodies, and other io.Readers.
	output, err := htmltomarkdown.ConvertReader(strings.NewReader(input),
		converter.WithDomain("https://example.com/docs/"))
	return string(output), err
}

func main() {
	markdown, err := convert(exampleHTML)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(markdown)
}
