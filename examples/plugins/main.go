package main

import (
	"fmt"

	"github.com/soulteary/html-to-markdown/v2/converter"
	"github.com/soulteary/html-to-markdown/v2/plugin/base"
	"github.com/soulteary/html-to-markdown/v2/plugin/commonmark"
	"github.com/soulteary/html-to-markdown/v2/plugin/strikethrough"
	"github.com/soulteary/html-to-markdown/v2/plugin/table"
)

const exampleHTML = `<table><tr><th>Name</th><th>Status</th></tr><tr><td>Legacy</td><td><s>active</s></td></tr></table>`

func convert(input string) (string, error) {
	conv := converter.NewConverter(converter.WithPlugins(
		base.NewBasePlugin(),
		commonmark.NewCommonmarkPlugin(),
		strikethrough.NewStrikethroughPlugin(),
		table.NewTablePlugin(table.WithCellPaddingBehavior(table.CellPaddingBehaviorNone)),
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
