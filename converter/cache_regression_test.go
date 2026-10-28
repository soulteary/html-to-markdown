package converter_test

import (
	"strings"
	"sync"
	"testing"

	"github.com/soulteary/html-to-markdown/v2/converter"
	"github.com/soulteary/html-to-markdown/v2/plugin/base"
	"github.com/soulteary/html-to-markdown/v2/plugin/commonmark"
	"golang.org/x/net/html"
)

func cachedConverter() *converter.Converter {
	return converter.NewConverter(converter.WithPlugins(base.NewBasePlugin(), commonmark.NewCommonmarkPlugin()))
}

func TestRegistrationAfterConversion(t *testing.T) {
	tests := []struct {
		name, input, before, after string
		register                   func(*converter.Converter)
	}{
		{"renderer", `<strong>x</strong>`, "**x**", "override", func(c *converter.Converter) {
			c.Register.RendererFor("strong", converter.TagTypeInline, func(_ converter.Context, w converter.Writer, _ *html.Node) converter.RenderStatus {
				w.WriteString("override")
				return converter.RenderSuccess
			}, converter.PriorityEarly)
		}},
		{"pre-renderer", "alpha", "alpha", "changed", func(c *converter.Converter) {
			c.Register.PreRenderer(func(_ converter.Context, n *html.Node) {
				var walk func(*html.Node)
				walk = func(n *html.Node) {
					if n.Type == html.TextNode {
						n.Data = "changed"
					}
					for child := n.FirstChild; child != nil; child = child.NextSibling {
						walk(child)
					}
				}
				walk(n)
			}, converter.PriorityEarly)
		}},
		{"post-renderer", "alpha", "alpha", "alpha!", func(c *converter.Converter) {
			c.Register.PostRenderer(func(_ converter.Context, content []byte) []byte { return append(content, '!') }, converter.PriorityLate+1)
		}},
		{"text-transformer", "alpha", "alpha", "ALPHA", func(c *converter.Converter) {
			c.Register.TextTransformer(func(_ converter.Context, s string) string { return strings.ToUpper(s) }, converter.PriorityEarly)
		}},
		{"unescaper", "^", "^", `\^`, func(c *converter.Converter) {
			c.Register.EscapedChar('^')
			c.Register.UnEscaper(func(chars []byte, index int) int {
				if chars[index] == '^' {
					return 1
				}
				return -1
			}, converter.PriorityEarly)
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := cachedConverter()
			// Warm every handler list, including the unescapers, before registration.
			if _, err := c.ConvertString("*literal*"); err != nil {
				t.Fatal(err)
			}
			if got, err := c.ConvertString(tc.input); err != nil || got != tc.before {
				t.Fatalf("before: %q, %v", got, err)
			}
			tc.register(c)
			for i := 0; i < 2; i++ {
				if got, err := c.ConvertString(tc.input); err != nil || got != tc.after {
					t.Fatalf("after: %q, %v; want %q", got, err, tc.after)
				}
			}
		})
	}
}

func TestConcurrentHandlerSnapshots(t *testing.T) {
	c := cachedConverter()
	// Multiple priorities used to cause readers to sort shared tag-type storage.
	c.Register.TagType("strong", converter.TagTypeBlock, converter.PriorityLate)
	c.Register.TagType("strong", converter.TagTypeInline, converter.PriorityEarly)
	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				got, err := c.ConvertString(`<p><strong>Hello</strong> and *literal*</p>`)
				if err != nil || got != `**Hello** and \*literal*` {
					t.Errorf("conversion: %q, %v", got, err)
					return
				}
			}
		}()
	}
	wg.Wait()
}
