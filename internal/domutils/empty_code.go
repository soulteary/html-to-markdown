package domutils

import (
	"context"

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
		if NodeName(parent) == "pre" {
			return true
		}
	}
	return false
}

func RemoveEmptyCode(ctx context.Context, doc *html.Node) {
	node := doc
	for node != nil {
		if NodeName(node) == "code" && !IsInsidePre(node) && !hasTextChildNodes(node) {
			next := GetNextNeighborNodeExcludingOwnChild(node)

			RemoveNode(node)

			node = next
			continue
		}

		node = GetNextNeighborNode(node)
	}
}
