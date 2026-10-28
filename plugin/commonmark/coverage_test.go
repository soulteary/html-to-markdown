package commonmark

import (
	"bytes"
	"github.com/soulteary/html-to-markdown/v2/converter"
	"github.com/soulteary/html-to-markdown/v2/plugin/base"
	"golang.org/x/net/html"
	"strings"
	"testing"
)

func TestRenderingHelperBoundaries(t *testing.T) {
	c := NewCommonmarkPlugin().(*commonmark)
	if c.getDelimiter(&html.Node{Type: html.ElementNode, Data: "span"}) != nil {
		t.Fatal("ordinary span has no delimiter")
	}
	if getHeadingLevel("custom") != 6 {
		t.Fatal("unknown heading should use level six")
	}
	if runeCount([]rune{'a', '\a', '中'}) != 2 {
		t.Fatal("internal markers should not affect heading width")
	}
	if getUnderlineWidth([]byte("x"), 3) != 3 {
		t.Fatal("short setext underline should use minimum width")
	}
	var out bytes.Buffer
	if c.renderComment(nil, &out, &html.Node{Type: html.CommentNode, Data: "ordinary"}) != converter.RenderTryNext || out.Len() != 0 {
		t.Fatal("ordinary comments should fall through")
	}
	pre := &html.Node{Type: html.ElementNode, Data: "pre"}
	for _, tag := range []string{"script", "style", "textarea"} {
		n := &html.Node{Type: html.ElementNode, Data: tag}
		n.AppendChild(&html.Node{Type: html.TextNode, Data: "ignored"})
		pre.AppendChild(n)
	}
	pre.AppendChild(&html.Node{Type: html.TextNode, Data: "kept"})
	if text, _ := getCodeWithoutTags(pre); string(text) != "kept" {
		t.Fatalf("non-code content leaked: %q", text)
	}
	cfg := fillInDefaultConfig(&config{})
	cfg.LinkStyle = "invalid"
	if err := validateConfig(&cfg); err == nil || !strings.Contains(err.Error(), "LinkStyle") {
		t.Fatalf("invalid link style accepted: %v", err)
	}
}

func TestUnsupportedLinkStyleFallsBack(t *testing.T) {
	c := NewCommonmarkPlugin().(*commonmark)
	c.LinkStyle = LinkStyleReferencedIndex
	conv := converter.NewConverter(converter.WithPlugins(base.NewBasePlugin(), c))
	got, err := conv.ConvertString(`<a href="/page">text</a>`)
	if err != nil || got != "text" {
		t.Fatalf("unsupported link style fallback: %q %v", got, err)
	}
}

func TestCodeTableRejectsExtraStructure(t *testing.T) {
	// Build the tree directly: HTML parsing would repair these malformed shapes.
	for _, kind := range []string{"extra row sibling", "extra cell element", "missing pre"} {
		t.Run(kind, func(t *testing.T) {
			root := &html.Node{Type: html.ElementNode, Data: "div"}
			table := &html.Node{Type: html.ElementNode, Data: "table"}
			body := &html.Node{Type: html.ElementNode, Data: "tbody"}
			row := &html.Node{Type: html.ElementNode, Data: "tr"}
			root.AppendChild(table)
			table.AppendChild(body)
			body.AppendChild(row)
			for _, class := range []string{"gutter", "code"} {
				cell := &html.Node{Type: html.ElementNode, Data: "td", Attr: []html.Attribute{{Key: "class", Val: class}}}
				row.AppendChild(cell)
				if kind != "missing pre" {
					p := &html.Node{Type: html.ElementNode, Data: "pre"}
					p.AppendChild(&html.Node{Type: html.TextNode, Data: "1"})
					cell.AppendChild(p)
				}
			}
			if kind == "extra row sibling" {
				body.AppendChild(&html.Node{Type: html.ElementNode, Data: "div"})
			}
			if kind == "extra cell element" {
				row.FirstChild.AppendChild(&html.Node{Type: html.ElementNode, Data: "span"})
			}
			unwrapCodeTables(root)
			if root.FirstChild != table {
				t.Fatal("non-code table was discarded")
			}
		})
	}
}

func TestHighlightedCodeTablesAndCodeLines(t *testing.T) {
	conv := converter.NewConverter(converter.WithPlugins(base.NewBasePlugin(), NewCommonmarkPlugin()))
	for _, tc := range []struct{ input, want string }{
		{`<table><tr><td class="other gutter"><pre>5<br>6<br></pre></td><td class="code other"><pre data-language="go">a<br>b<br></pre></td></tr></table>`, "```go\na\nb\n```"},
		{`<pre><code>a</code><code></code><code>b</code></pre>`, "```\na\n\nb\n```"},
	} {
		got, err := conv.ConvertString(tc.input)
		if err != nil || got != tc.want {
			t.Fatalf("code conversion: %q, %v; want %q", got, err, tc.want)
		}
	}
}

func TestCodeTableRejectionConditions(t *testing.T) {
	for _, kind := range []string{"not gutter", "many rows", "header cell", "text", "caption", "bad numbers", "mismatched lines"} {
		t.Run(kind, func(t *testing.T) {
			root := &html.Node{Type: html.ElementNode, Data: "div"}
			table := &html.Node{Type: html.ElementNode, Data: "table"}
			body := &html.Node{Type: html.ElementNode, Data: "tbody"}
			row := &html.Node{Type: html.ElementNode, Data: "tr"}
			root.AppendChild(table)
			table.AppendChild(body)
			body.AppendChild(row)
			for i, class := range []string{"gutter", "code"} {
				if kind == "not gutter" && i == 0 {
					class = "ordinary"
				}
				tag := "td"
				if kind == "header cell" {
					tag = "th"
				}
				cell := &html.Node{Type: html.ElementNode, Data: tag, Attr: []html.Attribute{{Key: "class", Val: class}}}
				row.AppendChild(cell)
				pre := &html.Node{Type: html.ElementNode, Data: "pre"}
				text := "1"
				if kind == "bad numbers" && i == 0 {
					text = "not a number"
				}
				if kind == "mismatched lines" && i == 1 {
					text = "one\ntwo"
				}
				pre.AppendChild(&html.Node{Type: html.TextNode, Data: text})
				cell.AppendChild(pre)
			}
			switch kind {
			case "many rows":
				body.AppendChild(&html.Node{Type: html.ElementNode, Data: "tr"})
			case "text":
				row.FirstChild.AppendChild(&html.Node{Type: html.TextNode, Data: "extra"})
			case "caption":
				table.AppendChild(&html.Node{Type: html.ElementNode, Data: "caption"})
			}
			unwrapCodeTables(root)
			if root.FirstChild != table {
				t.Fatal("table with unrelated content should be retained")
			}
		})
	}
}
