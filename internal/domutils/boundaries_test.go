package domutils

import (
	"reflect"
	"testing"

	"golang.org/x/net/html"
)

func TestFindNodesStayWithinSubtree(t *testing.T) {
	root := &html.Node{Type: html.ElementNode, Data: "root"}
	subtree := &html.Node{Type: html.ElementNode, Data: "article"}
	text := &html.Node{Type: html.TextNode, Data: "text"}
	leaf := &html.Node{Type: html.ElementNode, Data: "span"}
	comment := &html.Node{Type: html.CommentNode, Data: "comment"}
	outside := &html.Node{Type: html.ElementNode, Data: "footer"}
	root.AppendChild(subtree)
	root.AppendChild(outside)
	subtree.AppendChild(text)
	subtree.AppendChild(leaf)
	leaf.AppendChild(comment)
	matchAll := func(*html.Node) bool { return true }
	if got := FindAllNodes(subtree, matchAll); !reflect.DeepEqual(got, []*html.Node{text, leaf, comment}) {
		t.Fatalf("unexpected descendants: %v", got)
	}
	if got := AllNodes(subtree); !reflect.DeepEqual(got, []*html.Node{subtree, text, leaf, comment}) {
		t.Fatalf("unexpected subtree: %v", got)
	}
	if got := FindFirstNode(subtree, func(n *html.Node) bool { return n == subtree || n == outside }); got != nil {
		t.Fatal("search included its root or escaped its subtree")
	}
}

func TestDetachedNodeMutations(t *testing.T) {
	RemoveNode(nil)
	UnwrapNode(nil)
	root := &html.Node{Type: html.ElementNode, Data: "div"}
	child := &html.Node{Type: html.TextNode, Data: "keep"}
	root.AppendChild(child)
	RemoveNode(root)
	UnwrapNode(root)
	wrapper := &html.Node{Type: html.ElementNode, Data: "section"}
	if got := WrapNode(root, wrapper); got != root || root.FirstChild != child || root.Parent != nil || wrapper.FirstChild != nil {
		t.Fatal("detached root changed")
	}
	if got := WrapNode(nil, wrapper); got != nil {
		t.Fatal("nil node should remain nil")
	}
}

func TestEmptyAttributeDiffersFromMissing(t *testing.T) {
	node := &html.Node{Attr: []html.Attribute{{Key: "title", Val: ""}}}
	if value, ok := GetAttribute(node, "title"); !ok || value != "" {
		t.Fatal("empty attribute not found")
	}
	if got := GetAttributeOr(node, "title", "fallback"); got != "" {
		t.Fatal("empty attribute replaced by fallback")
	}
	if got := GetAttributeOr(node, "missing", "fallback"); got != "fallback" {
		t.Fatal("missing attribute did not use fallback")
	}
}
