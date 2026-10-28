// Adapted from github.com/JohannesKaufmann/dom v0.3.1.
// Copyright (c) 2024 Johannes Kaufmann.
// Licensed under the MIT license; see LICENSE.dom.

package domutils

import (
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func TestRemoveNode(t *testing.T) {
	child := &html.Node{
		Data: "child",
	}
	doc := &html.Node{
		Data: "parent",
	}

	if doc.FirstChild != nil {
		t.Error("expected FirstChild to be nil")
	}
	if child.Parent != nil {
		t.Error("expected Parent to be nil")
	}
	doc.AppendChild(child)
	if doc.FirstChild == nil {
		t.Error("expected FirstChild not to be nil anymore")
	}
	if child.Parent == nil {
		t.Error("expected Parent not to be nil anymore")
	}

	RemoveNode(child)
	if doc.FirstChild != nil {
		t.Error("expected FirstChild to be nil again")
	}
	if child.Parent != nil {
		t.Error("expected Parent to be nil again")
	}

	// Should not crash if run again...
	RemoveNode(child)
	if doc.Parent != nil {
		t.Error("expected Parent to still be nil")
	}
}

func TestUnwrapNode(t *testing.T) {
	child1 := &html.Node{
		Type: html.ElementNode,
		Data: "child1",
	}
	child2 := &html.Node{
		Type: html.ElementNode,
		Data: "child2",
	}
	child2.AppendChild(&html.Node{
		Type: html.ElementNode,
		Data: "child2.1",
	})

	parent := &html.Node{
		Type: html.ElementNode,
		Data: "parent",
	}
	parent.AppendChild(child1)
	parent.AppendChild(child2)

	root := &html.Node{
		Type: html.ElementNode,
		Data: "root",
	}
	root.AppendChild(parent)

	// - - - - - //

	expectedBefore := strings.TrimSpace(`
root
├─parent
│ ├─child1
│ ├─child2
│ │ ├─child2.1
	`)
	expectedAfter := strings.TrimSpace(`
root
├─child1
├─child2
│ ├─child2.1
	`)

	if RenderRepresentation(root) != expectedBefore {
		t.Error("expected a different initial render")
	}
	if root.FirstChild != parent {
		t.Error("expected the parent to be under root")
	}
	UnwrapNode(parent)
	if root.FirstChild == parent {
		t.Error("expected the parent to not be under root anymore")
	}
	if RenderRepresentation(root) != expectedAfter {
		t.Error("expected a different final render")
	}

	UnwrapNode(root)
	if root.Data != "root" {
		t.Error("expected the root to still be the root")
	}
}

func TestWrapNode(t *testing.T) {
	child1 := &html.Node{
		Type: html.ElementNode,
		Data: "child1",
	}
	child2 := &html.Node{
		Type: html.ElementNode,
		Data: "child2",
	}
	child2.AppendChild(&html.Node{
		Type: html.ElementNode,
		Data: "child2.1",
	})

	parent := &html.Node{
		Type: html.ElementNode,
		Data: "parent",
	}
	parent.AppendChild(child1)
	parent.AppendChild(child2)

	root := &html.Node{
		Type: html.ElementNode,
		Data: "root",
	}
	root.AppendChild(parent)

	// - - - - - //
	wrapper := &html.Node{
		Type: html.ElementNode,
		Data: "wrapper",
	}

	expectedBefore := strings.TrimSpace(`
root
├─parent
│ ├─child1
│ ├─child2
│ │ ├─child2.1
	`)
	expectedAfter := strings.TrimSpace(`
root
├─wrapper
│ ├─parent
│ │ ├─child1
│ │ ├─child2
│ │ │ ├─child2.1
	`)

	if RenderRepresentation(root) != expectedBefore {
		t.Error("expected a different initial render")
	}
	if root.FirstChild != parent {
		t.Error("expected the parent to be under root")
	}
	WrapNode(parent, wrapper)
	if RenderRepresentation(root) != expectedAfter {
		t.Error("expected a different final render")
	}
	if root.FirstChild != wrapper {
		t.Error("expected the wrapper to be under root")
	}

	// With no parent, there should be no change.
	WrapNode(root, &html.Node{Type: html.ElementNode, Data: "sky"})
	if RenderRepresentation(root) != expectedAfter {
		t.Error("there should be no changes")
	}
}
