package domutils

import (
	"context"

	"github.com/JohannesKaufmann/dom"
	"golang.org/x/net/html"
)

func hasTextChildNodes(startNode *html.Node) bool {
	var found bool

	var finder func(*html.Node)
	finder = func(node *html.Node) {
		if node.Type == html.TextNode && node.Data != "" {
			found = true
			return
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			finder(child)
		}
	}
	finder(startNode)

	return found
}

// IsInsidePre reports whether a node belongs to a preformatted block.
func IsInsidePre(node *html.Node) bool {
	for parent := node.Parent; parent != nil; parent = parent.Parent {
		if dom.NodeName(parent) == "pre" {
			return true
		}
	}
	return false
}

func RemoveEmptyCode(ctx context.Context, doc *html.Node) {
	node := doc
	for node != nil {
		if dom.NodeName(node) == "code" && !IsInsidePre(node) && !hasTextChildNodes(node) {
			next := dom.GetNextNeighborNodeExcludingOwnChild(node)

			dom.RemoveNode(node)

			node = next
			continue
		}

		node = dom.GetNextNeighborNode(node)
	}
}
