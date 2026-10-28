package main

import (
	"fmt"
	"log"

	"github.com/soulteary/html-to-markdown/v2/converter"
	"github.com/soulteary/html-to-markdown/v2/plugin/base"
	"github.com/soulteary/html-to-markdown/v2/plugin/commonmark"
)

const exampleHTML = `<p><strong>Custom formatting</strong></p><ul><li>First</li><li>Second</li></ul>`

func convert(input string) (string, error) {
	conv := converter.NewConverter(converter.WithPlugins(
		base.NewBasePlugin(),
		commonmark.NewCommonmarkPlugin(
			commonmark.WithStrongDelimiter("__"),
			commonmark.WithBulletListMarker("*"),
		),
	))
	return conv.ConvertString(input)
}

func main() {
	markdown, err := convert(exampleHTML)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(markdown)
}
