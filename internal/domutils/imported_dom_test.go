// Adapted from github.com/JohannesKaufmann/dom v0.3.1.
// Copyright (c) 2024 Johannes Kaufmann.
// Licensed under the MIT license; see LICENSE.dom.

package domutils

import (
	"testing"

	"golang.org/x/net/html"
)

func TestAllNodes(t *testing.T) {
	child0 := &html.Node{
		Type: html.ElementNode,
	}
	child0.AppendChild(&html.Node{
		Type: html.TextNode,
	})

	child1 := &html.Node{
		Type: html.ElementNode,
	}

	doc := &html.Node{}
	doc.AppendChild(child0)
	doc.AppendChild(child1)

	nodes := AllNodes(doc)
	if len(nodes) != 4 {
		t.Errorf("expected different length, but got %d", len(nodes))
	}
}

// - - - - - - - - - - - - - - - //

func TestAllChildNodes(t *testing.T) {
	child0 := &html.Node{
		Type: html.ElementNode,
	}
	child1 := &html.Node{
		Type: html.TextNode,
	}
	child2 := &html.Node{
		Type: html.ElementNode,
	}

	doc := &html.Node{}
	doc.AppendChild(child0)
	doc.AppendChild(child1)
	doc.AppendChild(child2)

	nodes := AllChildNodes(doc)
	if len(nodes) != 3 {
		t.Error("expected different length")
	}
	if nodes[0] != child0 || nodes[1] != child1 || nodes[2] != child2 {
		t.Error("expected different nodes")
	}
}

func TestFirstChildNode(t *testing.T) {
	node := &html.Node{
		Type: html.ElementNode,
	}
	child := &html.Node{
		Type: html.TextNode,
	}
	node.AppendChild(child)

	res := FirstChildNode(node)
	if res != child {
		t.Error("expected the first child node")
	}

	res = FirstChildNode(child)
	if res != nil {
		t.Error("expected the first child to be nil")
	}

}

func TestFirstChildElement(t *testing.T) {
	node := &html.Node{
		Type: html.ElementNode,
	}
	child1 := &html.Node{
		Type: html.TextNode,
	}
	child2 := &html.Node{
		Type: html.ElementNode,
	}
	node.AppendChild(child1)
	node.AppendChild(child2)

	res := FirstChildElement(node)
	if res != child2 {
		t.Error("expected the first child node to be child2")
	}
}

// - - - - - - - - - - - - - - - //

func TestNextElementSibling(t *testing.T) {
	end := &html.Node{
		Type: html.ElementNode,
		Data: "end",
	}
	text := &html.Node{
		Type: html.TextNode,
		Data: "between",

		NextSibling: end,
	}

	start := &html.Node{
		Type: html.ElementNode,
		Data: "start",

		NextSibling: text,
	}

	output := NextSiblingElement(start)
	if output != end {
		t.Error("expected 'end' node")
	}

	output = NextSiblingElement(text)
	if output != end {
		t.Error("expected 'end' node")
	}

	output = NextSiblingElement(end)
	if output != nil {
		t.Error("expected nil node")
	}
}
