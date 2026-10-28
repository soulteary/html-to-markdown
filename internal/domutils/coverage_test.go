package domutils

import (
	"context"
	"golang.org/x/net/html"
	"slices"
	"testing"
)

func TestTextNeighborsAcrossSpans(t *testing.T) {
	parent := &html.Node{Type: html.ElementNode, Data: "div"}
	before := &html.Node{Type: html.TextNode, Data: "before"}
	left := &html.Node{Type: html.ElementNode, Data: "span"}
	middle := &html.Node{Type: html.ElementNode, Data: "b"}
	right := &html.Node{Type: html.ElementNode, Data: "span"}
	after := &html.Node{Type: html.TextNode, Data: "after"}
	for _, n := range []*html.Node{before, left, middle, right, after} {
		parent.AppendChild(n)
	}
	if getNextTextNode(middle) != after || getPrevTextNode(middle) != before {
		t.Fatal("transparent spans interrupted text neighbors")
	}
	if getNextTextNode(after) != nil || getPrevTextNode(before) != nil {
		t.Fatal("unexpected edge neighbor")
	}
	blocker := &html.Node{Type: html.ElementNode, Data: "hr"}
	parent.InsertBefore(blocker, right)
	if getNextTextNode(middle) != nil {
		t.Fatal("block element should stop text lookup")
	}
	MergeAdjacentTextNodes(nil)
}

func TestPreAlternativeInsideInline(t *testing.T) {
	root := &html.Node{Type: html.ElementNode, Data: "strong"}
	pre := &html.Node{Type: html.ElementNode, Data: "pre"}
	pre.AppendChild(&html.Node{Type: html.TextNode, Data: "code"})
	root.AppendChild(pre)
	LeafBlockAlternatives(context.Background(), root)
	if pre.Data != "code" || pre.FirstChild.Data != "code" {
		t.Fatal("pre inside inline should become inline code")
	}
}
func TestSwapRejectsTextNodes(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("text-node swap did not panic")
		}
	}()
	swapTagsOfNodes(&html.Node{Type: html.TextNode}, &html.Node{Type: html.ElementNode, Data: "em"})
}

func TestSpaceChildLookupBoundaries(t *testing.T) {
	match := func(n *html.Node) bool { return n.Data == "strong" }
	root := &html.Node{Type: html.ElementNode, Data: "div"}
	if getFirstChildNode(root, match) != nil || getLastChildNode(root, match) != nil {
		t.Fatal("empty node has matching children")
	}
	child := &html.Node{Type: html.ElementNode, Data: "p"}
	root.AppendChild(child)
	if getFirstChildNode(root, match) != nil || getLastChildNode(root, match) != nil {
		t.Fatal("non-matching child accepted")
	}
	root.RemoveChild(child)
	span := &html.Node{Type: html.ElementNode, Data: "span"}
	strong := &html.Node{Type: html.ElementNode, Data: "strong"}
	span.AppendChild(strong)
	root.AppendChild(span)
	if getFirstChildNode(root, match) != strong || getLastChildNode(root, match) != strong {
		t.Fatal("transparent span should expose its child")
	}
}

func TestBlockAlternativesAndTextContent(t *testing.T) {
	for _, name := range []string{"div", "unknown"} {
		if getMarkdownStructure(name) != "" {
			t.Fatalf("unexpected structure for %q", name)
		}
	}
	root := &html.Node{Type: html.ElementNode, Data: "h1"}
	quote := &html.Node{Type: html.ElementNode, Data: "blockquote"}
	root.AppendChild(quote)
	list := &html.Node{Type: html.ElementNode, Data: "ul"}
	root.AppendChild(list)
	LeafBlockAlternatives(context.Background(), root)
	if quote.Data != "span" || quote.PrevSibling.Data != ` "` || quote.NextSibling.Data != `" ` || list.Data != "span" {
		t.Fatal("nested block alternatives lost structure")
	}
	code := &html.Node{Type: html.ElementNode, Data: "code"}
	span := &html.Node{Type: html.ElementNode, Data: "span"}
	span.AppendChild(&html.Node{Type: html.TextNode, Data: "value"})
	code.AppendChild(span)
	if !hasTextChildNodes(code) {
		t.Fatal("nested code text should count as content")
	}
}

func TestListEndBoundaries(t *testing.T) {
	for _, kind := range []string{"list", "li", "comment", "text", "hr", "none", "transparent"} {
		t.Run(kind, func(t *testing.T) {
			root := &html.Node{Type: html.ElementNode, Data: "div"}
			list := &html.Node{Type: html.ElementNode, Data: "ul"}
			root.AppendChild(list)
			switch kind {
			case "list":
				root.AppendChild(&html.Node{Type: html.ElementNode, Data: "ol"})
			case "li":
				root.AppendChild(&html.Node{Type: html.ElementNode, Data: "li"})
			case "comment":
				root.AppendChild(&html.Node{Type: html.CommentNode, Data: ListEndCommentData})
			case "text":
				root.AppendChild(&html.Node{Type: html.TextNode, Data: "separate"})
			case "hr":
				root.AppendChild(&html.Node{Type: html.ElementNode, Data: "hr"})
			case "transparent":
				root.AppendChild(&html.Node{Type: html.CommentNode, Data: "ordinary"})
				root.AppendChild(&html.Node{Type: html.ElementNode, Data: "span"})
				root.AppendChild(&html.Node{Type: html.ElementNode, Data: "ol"})
			}
			want := kind == "list" || kind == "transparent"
			if nextNameIsList(list) != want {
				t.Fatal("wrong list adjacency")
			}
		})
	}
	root := &html.Node{Type: html.ElementNode, Data: "div"}
	a := &html.Node{Type: html.ElementNode, Data: "ul"}
	b := &html.Node{Type: html.ElementNode, Data: "ol"}
	root.AppendChild(a)
	root.AppendChild(b)
	AddListEndComments(context.Background(), root)
	AddListEndComments(context.Background(), root)
	if a.NextSibling.Type != html.CommentNode || a.NextSibling.Data != ListEndCommentData || a.NextSibling.NextSibling != b {
		t.Fatal("adjacent lists need exactly one separator")
	}
}

func TestLeafAlternativesPreserveSiblingOrderDuringMutation(t *testing.T) {
	root := &html.Node{Type: html.ElementNode, Data: "strong"}
	for _, tag := range []string{"h1", "hr", "blockquote", "pre", "span", "hr"} {
		n := &html.Node{Type: html.ElementNode, Data: tag}
		if tag != "hr" {
			n.AppendChild(&html.Node{Type: html.TextNode, Data: tag})
		}
		root.AppendChild(n)
	}
	LeafBlockAlternatives(context.Background(), root)
	want := []string{"strong", "br", "#text", "span", "#text", "code", "span"}
	var names []string
	for n := root.FirstChild; n != nil; n = n.NextSibling {
		names = append(names, NodeName(n))
		if n.Parent != root {
			t.Fatal("replacement lost parent link")
		}
	}
	if !slices.Equal(names, want) {
		t.Fatalf("alternative sibling order: %v; want %v", names, want)
	}
	if root.FirstChild.FirstChild.Data != "h1" || root.LastChild.FirstChild.Data != "span" {
		t.Fatal("alternative traversal changed retained text")
	}
}
