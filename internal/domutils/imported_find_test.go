// Adapted from github.com/JohannesKaufmann/dom v0.3.1.
// Copyright (c) 2024 Johannes Kaufmann.
// Licensed under the MIT license; see LICENSE.dom.

package domutils

import (
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func TestFindFirstNode(t *testing.T) {
	input := `
<html>
	<head></head>
	<body>
		<article>
			<div>
				<h3>Heading</h3>
				<p>short description</p>
			</div>
		</article>

		<section>
			<h4>Heading</h4>
			<p>another description</p>
		</section>
	</body>
</html>
	`

	doc, err := html.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}

	article := FindFirstNode(doc, func(node *html.Node) bool {
		return NodeName(node) == "article"
	})
	if article == nil || article.Data != "article" {
		t.Error("got different node")
	}

	h3 := FindFirstNode(article, func(node *html.Node) bool {
		return NodeName(node) == "h3"
	})
	if h3 == nil || h3.Data != "h3" {
		t.Error("got different node")
	}

	h4 := FindFirstNode(article, func(node *html.Node) bool {
		return NodeName(node) == "h4"
	})
	if h4 != nil {
		t.Error("expected nil node")
	}
}

func TestFindAllNodes(t *testing.T) {
	input := `
<html>
	<head></head>
	<body>
		<article>
			<div>
				<h3>Heading</h3>
				<p>short description</p>
			</div>
		</article>

		<section>
			<h4>Heading</h4>
			<p>another description</p>
		</section>
	</body>
</html>
	`

	doc, err := html.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}

	paragraphs := FindAllNodes(doc, func(node *html.Node) bool {
		return NodeName(node) == "p"
	})
	if len(paragraphs) != 2 {
		t.Error("expected 2 nodes")
	}
	if paragraphs[0].Data != "p" || paragraphs[1].Data != "p" {
		t.Error("expected paragraph nodes")
	}
}
