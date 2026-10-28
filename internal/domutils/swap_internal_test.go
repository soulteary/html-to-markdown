package domutils

import (
	"testing"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func generateANodes() *html.Node {
	div := &html.Node{
		Namespace: "",
		Type:      html.ElementNode,
		DataAtom:  atom.Div,

		Attr: []html.Attribute{},
		Data: "div",
	}
	link := &html.Node{
		Namespace: "",
		Type:      html.ElementNode,
		DataAtom:  atom.A,

		Attr: []html.Attribute{
			{
				Namespace: "",
				Key:       "KeyA",
				Val:       "ValA",
			},
		},

		Data: "a",
	}
	text := &html.Node{
		Namespace: "",
		Type:      html.TextNode,
		DataAtom:  0,

		Data: "ContentA",
	}
	link.AppendChild(text)
	div.AppendChild(link)

	return div
}
func generateBNodes() *html.Node {
	div := &html.Node{
		Namespace: "",
		Type:      html.ElementNode,
		DataAtom:  atom.Main,

		Attr: []html.Attribute{},
		Data: "main",
	}
	link := &html.Node{
		Namespace: "",
		Type:      html.ElementNode,
		DataAtom:  atom.B,

		Attr: []html.Attribute{
			{
				Namespace: "",
				Key:       "KeyB",
				Val:       "ValB",
			},
		},

		Data: "b",
	}
	text := &html.Node{
		Namespace: "",
		Type:      html.TextNode,
		DataAtom:  0,

		Data: "ContentB",
	}
	link.AppendChild(text)
	div.AppendChild(link)

	return div
}
func TestSwapTagsOfNodes_Basics(t *testing.T) {
	a := generateANodes()
	b := generateBNodes()

	swapTagsOfNodes(a.FirstChild, b.FirstChild)

	// These fields should have changed:
	if a.FirstChild.DataAtom != atom.B {
		t.Error("expected different a atom")
	}
	if a.FirstChild.Data != "b" {
		t.Error("expected different a data")
	}
	if len(a.FirstChild.Attr) != 1 {
		t.Error("expected different a attributes length")
	}
	if a.FirstChild.Attr[0].Key != "KeyB" {
		t.Error("expected different a attribute key")
	}
	if a.FirstChild.Attr[0].Val != "ValB" {
		t.Error("expected different a attribute key")
	}

	// The pointers should NOT have changed:
	if a.FirstChild.Parent.Data != "div" {
		t.Error("expected the same parent for a")
	}
	if a.FirstChild.FirstChild.Data != "ContentA" {
		t.Error("expected the same text for a")
	}

	// - - - //

	if b.FirstChild.DataAtom != atom.A {
		t.Error("expected different b atom")
	}
	if b.FirstChild.Data != "a" {
		t.Error("expected different b data")
	}
	if len(b.FirstChild.Attr) != 1 {
		t.Error("expected different b attributes length")
	}
	if b.FirstChild.Attr[0].Key != "KeyA" {
		t.Error("expected different b attribute key")
	}
	if b.FirstChild.Attr[0].Val != "ValA" {
		t.Error("expected different b attribute key")
	}

	// The pointers should NOT have changed:
	if b.FirstChild.Parent.Data != "main" {
		t.Error("expected the same parent for b")
	}
	if b.FirstChild.FirstChild.Data != "ContentB" {
		t.Error("expected the same text for b")
	}
}
