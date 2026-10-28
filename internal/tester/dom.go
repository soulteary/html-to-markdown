package tester

import (
	"io"
	"strings"

	"github.com/soulteary/html-to-markdown/v2/internal/domutils"
	"golang.org/x/net/html"
)

func Parse(t reporter, rawHTML string, startFrom string) *html.Node {
	t.Helper()
	return parseReader(t, strings.NewReader(strings.TrimSpace(rawHTML)), startFrom)
}

func parseReader(t reporter, input io.Reader, startFrom string) *html.Node {
	if startFrom == "" {
		startFrom = "body"
	}

	doc, err := html.Parse(input)
	if err != nil {
		t.Fatal(err)
	}

	var foundNode *html.Node
	var finder func(node *html.Node)
	finder = func(node *html.Node) {
		if foundNode != nil {
			return
		}
		if domutils.NodeName(node) == startFrom {
			foundNode = node
			return
		}

		for child := node.FirstChild; child != nil; child = child.NextSibling {
			finder(child)
		}
	}
	finder(doc)

	if foundNode == nil {
		t.Error("could not find node for 'startFrom'")
	}
	return foundNode
}
