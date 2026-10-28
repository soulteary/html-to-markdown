package table

import (
	"github.com/JohannesKaufmann/html-to-markdown/v2/internal/domutils"
	"golang.org/x/net/html"
)

func selectHeaderRowNode(node *html.Node) *html.Node {
	thead := domutils.FindFirstNode(node, func(n *html.Node) bool {
		return domutils.NodeName(n) == "thead"
	})
	if thead != nil {
		firstTr := domutils.FindFirstNode(thead, func(n *html.Node) bool {
			return domutils.NodeName(n) == "tr"
		})
		if firstTr != nil {
			// YEAH we found the "tr" inside the "thead"
			return firstTr
		}
	}

	firstTh := domutils.FindFirstNode(node, func(n *html.Node) bool {
		return domutils.NodeName(n) == "th"
	})
	if firstTh != nil {
		// YEAH we found the "th"
		return firstTh.Parent
	}

	return nil
}
func selectNormalRowNodes(tableNode *html.Node, selectedHeaderRowNode *html.Node) []*html.Node {
	var collected []*html.Node

	var finder func(node *html.Node)
	finder = func(node *html.Node) {
		name := domutils.NodeName(node)
		if name == "tr" && node != selectedHeaderRowNode {
			// We want to make sure to not select the header row a *second* time.
			collected = append(collected, node)
		}

		for child := node.FirstChild; child != nil; child = child.NextSibling {
			finder(child)
		}
	}
	finder(tableNode)

	return collected
}
