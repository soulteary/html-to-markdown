package htmltomarkdown_test

import (
	"strings"
	"testing"

	htmltomarkdown "github.com/soulteary/html-to-markdown/v2"
	"github.com/soulteary/html-to-markdown/v2/converter"
	"github.com/soulteary/html-to-markdown/v2/plugin/base"
	"github.com/soulteary/html-to-markdown/v2/plugin/commonmark"
	"github.com/soulteary/html-to-markdown/v2/plugin/table"
)

var benchmarkOutput string

func benchmarkConverter(tables bool) *converter.Converter {
	plugins := []converter.Plugin{base.NewBasePlugin(), commonmark.NewCommonmarkPlugin()}
	if tables {
		plugins = append(plugins, table.NewTablePlugin())
	}
	return converter.NewConverter(converter.WithPlugins(plugins...))
}

func BenchmarkConvert(b *testing.B) {
	cases := []struct {
		name, input string
		tables      bool
	}{
		{"Small", `<h1>Hello</h1><p>A <strong>small</strong> example.</p>`, false},
		{"Article", strings.Repeat(`<section><h2>Go library</h2><p>HTML conversion with <strong>formatting</strong> and useful plain text.</p></section>`, 200), false},
		{"MarkdownText", `<p>` + strings.Repeat("中文 *word* _x_ [label](path) ! # + - 1. ~ `code` &lt;tag&gt; ", 500) + `</p>`, false},
		{"Links", strings.Repeat(`<p><a href="https://example.com/a?x=1&amp;y=2">A link</a> <img src="/photo.png" alt="photo"></p>`, 200), false},
		{"ManyNodes", `<div>` + strings.Repeat(`<span>text</span>`, 2000) + `</div>`, false},
		{"Table", `<table><tr><th>Name</th><th>Value</th></tr>` + strings.Repeat(`<tr><td><strong>Item</strong></td><td>42</td></tr>`, 200) + `</table>`, true},
	}
	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			for _, reuse := range []bool{false, true} {
				name := "Fresh"
				if reuse {
					name = "Reuse"
				}
				b.Run(name, func(b *testing.B) {
					conv := benchmarkConverter(tc.tables)
					b.ReportAllocs()
					b.SetBytes(int64(len(tc.input)))
					b.ResetTimer()
					for i := 0; i < b.N; i++ {
						var out string
						var err error
						if reuse {
							out, err = conv.ConvertString(tc.input)
						} else if tc.tables {
							out, err = benchmarkConverter(true).ConvertString(tc.input)
						} else {
							out, err = htmltomarkdown.ConvertString(tc.input)
						}
						if err != nil {
							b.Fatal(err)
						}
						benchmarkOutput = out
					}
				})
			}
		})
	}
}
