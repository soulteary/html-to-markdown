package main

import (
	"fmt"

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
	fmt.Println(must(convert(exampleHTML)))
}

func must(markdown string, err error) string {
	if err != nil {
		panic(err)
	}
	return markdown
}
