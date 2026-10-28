package main

import (
	"fmt"
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
	fmt.Println(must(convert(exampleHTML)))
}

func must(markdown string, err error) string {
	if err != nil {
		panic(err)
	}
	return markdown
}
