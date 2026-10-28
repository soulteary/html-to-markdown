package comparison_test

import (
	"fmt"
	upstream "github.com/JohannesKaufmann/html-to-markdown/v2"
	uc "github.com/JohannesKaufmann/html-to-markdown/v2/converter"
	ub "github.com/JohannesKaufmann/html-to-markdown/v2/plugin/base"
	um "github.com/JohannesKaufmann/html-to-markdown/v2/plugin/commonmark"
	ut "github.com/JohannesKaufmann/html-to-markdown/v2/plugin/table"
	fork "github.com/soulteary/html-to-markdown/v2"
	fc "github.com/soulteary/html-to-markdown/v2/converter"
	fb "github.com/soulteary/html-to-markdown/v2/plugin/base"
	fm "github.com/soulteary/html-to-markdown/v2/plugin/commonmark"
	ft "github.com/soulteary/html-to-markdown/v2/plugin/table"
	"strings"
	"testing"
)

var sink string

func upstreamConverter(tables bool) *uc.Converter {
	p := []uc.Plugin{ub.NewBasePlugin(), um.NewCommonmarkPlugin()}
	if tables {
		p = append(p, ut.NewTablePlugin())
	}
	return uc.NewConverter(uc.WithPlugins(p...))
}
func forkConverter(tables bool) *fc.Converter {
	p := []fc.Plugin{fb.NewBasePlugin(), fm.NewCommonmarkPlugin()}
	if tables {
		p = append(p, ft.NewTablePlugin())
	}
	return fc.NewConverter(fc.WithPlugins(p...))
}
func BenchmarkV201VsUpstream(b *testing.B) {
	cases := []struct {
		name, input string
		tables      bool
	}{
		{"Small", `<h1>Hello</h1><p>A <strong>small</strong> example.</p>`, false},
		{"Article", strings.Repeat(`<section><h2>Go library</h2><p>HTML conversion with <strong>formatting</strong> and useful plain text.</p></section>`, 200), false},
		{"MarkdownText", `<p>` + strings.Repeat("中文 *word* _x_ [label](path) ! # + - 1. ~ `code` &lt;tag&gt; ", 500) + `</p>`, false},
		{"Links", strings.Repeat(`<p><a href="https://example.com/a?x=1&amp;y=2">A link</a> <img src="/photo.png" alt="photo"></p>`, 200), false},
		{"ManyNodes", `<div>` + strings.Repeat(`<span>text</span>`, 2000) + `</div>`, false},
		{"WideRemovals", `<div>` + strings.Repeat(`<span>kept</span><script>ignored</script><style>ignored</style>`, 1000) + `</div>`, false},
		{"LongUnicodeText", `<p>` + strings.Repeat("中文 text with   spaces and punctuation. ", 2000) + `</p>`, false},
		{"DeepTree", strings.Repeat(`<div>`, 128) + `<p>deep text</p>` + strings.Repeat(`</div>`, 128), false},
		{"Table", `<table><tr><th>Name</th><th>Value</th></tr>` + strings.Repeat(`<tr><td><strong>Item</strong></td><td>42</td></tr>`, 200) + `</table>`, true},
	}
	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			for _, reuse := range []bool{false, true} {
				mode := "Fresh"
				if reuse {
					mode = "Reuse"
				}
				b.Run(mode, func(b *testing.B) {
					// Check identical output before timing and alternate measurement order.
					u, ue := upstreamConverter(tc.tables).ConvertString(tc.input)
					f, fe := forkConverter(tc.tables).ConvertString(tc.input)
					if ue != nil || fe != nil || u != f {
						b.Fatal("outputs differ", ue, fe)
					}
					for round := 0; round < 8; round++ {
						b.Run(fmt.Sprintf("Round%d", round), func(b *testing.B) {
							names := []string{"Upstream", "Fork"}
							if round%2 == 1 {
								names[0], names[1] = names[1], names[0]
							}
							for _, name := range names {
								b.Run(name, func(b *testing.B) {
									var run func(string) (string, error)
									if name == "Upstream" {
										c := upstreamConverter(tc.tables)
										run = func(s string) (string, error) {
											if reuse {
												return c.ConvertString(s)
											}
											if tc.tables {
												return upstreamConverter(true).ConvertString(s)
											}
											return upstream.ConvertString(s)
										}
									} else {
										c := forkConverter(tc.tables)
										run = func(s string) (string, error) {
											if reuse {
												return c.ConvertString(s)
											}
											if tc.tables {
												return forkConverter(true).ConvertString(s)
											}
											return fork.ConvertString(s)
										}
									}
									b.ReportAllocs()
									b.SetBytes(int64(len(tc.input)))
									b.ResetTimer()
									for i := 0; i < b.N; i++ {
										out, err := run(tc.input)
										if err != nil {
											b.Fatal(err)
										}
										sink = out
									}
								})
							}
						})
					}
				})
			}
		})
	}
}
