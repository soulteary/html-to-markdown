// Adapted from github.com/JohannesKaufmann/dom v0.3.1.
// Copyright (c) 2024 Johannes Kaufmann.
// Licensed under the MIT license; see LICENSE.dom.

package domutils

import (
	"strings"
	"testing"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func TestRenderRepresentation_NoAttributes(t *testing.T) {
	input := "<a>\nText\n</a>"
	expected := "├─body\n│ ├─a\n│ │ ├─#text \"\\nText\\n\""

	doc, err := html.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	body := FindFirstNode(doc, func(node *html.Node) bool {
		return node.DataAtom == atom.Body
	})

	output := RenderRepresentation(body)

	if output != expected {
		t.Errorf("expected %q but got %q", expected, output)
	}
}

func TestRenderRepresentation_MultipleAttributes(t *testing.T) {
	input := `<a href="/page.html" target="_blank" class="button primary">Text</a>`
	expected := `├─body
│ ├─a (class="button primary" href="/page.html" target="_blank")
│ │ ├─#text "Text"`

	doc, err := html.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	body := FindFirstNode(doc, func(node *html.Node) bool {
		return node.DataAtom == atom.Body
	})

	output := RenderRepresentation(body)

	if output != expected {
		t.Errorf("expected %q but got %q", expected, output)
	}
}

func TestRenderRepresentation_Root(t *testing.T) {
	input := `<img src="/img.png" />`
	expected := strings.TrimSpace(`

#document
├─html
│ ├─head
│ ├─body
│ │ ├─img (src="/img.png")

`)

	doc, err := html.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}

	output := RenderRepresentation(doc)

	t.Log(output)
	if output != expected {
		t.Errorf("expected %q but got %q", expected, output)
	}
}
