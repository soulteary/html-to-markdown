package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/soulteary/html-to-markdown/v2/converter"
	"github.com/soulteary/html-to-markdown/v2/plugin/base"
	"github.com/soulteary/html-to-markdown/v2/plugin/commonmark"
	"golang.org/x/net/html"
)

const exampleHTML = `<h2>Go <star-rating count="4">four stars</star-rating></h2>`

func convert(input string) (string, error) {
	conv := converter.NewConverter(converter.WithPlugins(
		base.NewBasePlugin(), commonmark.NewCommonmarkPlugin(),
	))
	conv.Register.RendererFor("star-rating", converter.TagTypeInline, renderStarRating, converter.PriorityEarly)
	return conv.ConvertString(input)
}

func renderStarRating(_ converter.Context, w converter.Writer, node *html.Node) converter.RenderStatus {
	// Read attributes directly so the example can be copied into an external plugin.
	var rawCount string
	for _, attr := range node.Attr {
		if attr.Key == "count" {
			rawCount = attr.Val
			break
		}
	}
	count, err := strconv.Atoi(rawCount)
	if err != nil || count < 0 || count > 5 {
		// Let normal rendering preserve the element's fallback text.
		return converter.RenderTryNext
	}
	w.WriteString(strings.Repeat("★", count))
	return converter.RenderSuccess
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
