package base

import (
	"github.com/soulteary/html-to-markdown/v2/converter"
	"golang.org/x/net/html"
	"strings"
	"testing"
)

func TestRemovalWalkKeepsAdjacentTextAndSkipsRemovedSubtrees(t *testing.T) {
	doc := &html.Node{Type: html.ElementNode, Data: "div"}
	var removed []*html.Node
	for i := 0; i < 128; i++ {
		for _, tag := range []string{"script", "style"} {
			n := &html.Node{Type: html.ElementNode, Data: tag}
			n.AppendChild(&html.Node{Type: html.TextNode, Data: "ignored"})
			doc.AppendChild(n)
			removed = append(removed, n)
		}
		doc.AppendChild(&html.Node{Type: html.TextNode, Data: "a"})
	}
	conv := converter.NewConverter(converter.WithPlugins(NewBasePlugin()))
	conv.Register.Renderer(func(converter.Context, converter.Writer, *html.Node) converter.RenderStatus {
		return converter.RenderTryNext
	}, converter.PriorityStandard)
	got, err := conv.ConvertNode(doc)
	if err != nil || string(got) != strings.Repeat("a", 128) {
		t.Fatalf("adjacent removals lost text: %q, %v", got, err)
	}
	if doc.FirstChild != doc.LastChild || doc.FirstChild.Type != html.TextNode {
		t.Fatal("text separated by removed siblings was not merged")
	}
	for _, n := range removed {
		if n.Parent != nil {
			t.Fatal("removed sibling remains attached")
		}
		if n.FirstChild.Data != "ignored" {
			t.Fatal("removed subtree was unexpectedly rewritten")
		}
	}
}
