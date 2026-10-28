package domutils

import (
	"golang.org/x/net/html"
)

func getNextTextNode(startNode *html.Node) *html.Node {
	node := GetNextNeighborNodeExcludingOwnChild(startNode)

	for node != nil {
		if node.Type == html.TextNode {
			return node
		}

		if NodeName(node) == "span" {
			// A span has no special meaning. So we just skip it...
			node = GetNextNeighborNode(node)
			continue
		}

		return nil
	}
	return nil
}
func getPrevTextNode(startNode *html.Node) *html.Node {
	node := GetPrevNeighborNodeExcludingOwnChild(startNode)

	for node != nil {
		if node.Type == html.TextNode {
			return node
		}

		if NodeName(node) == "span" {
			// A span has no special meaning. So we just skip it...
			node = GetPrevNeighborNode(node)
			continue
		}

		return nil
	}
	return nil
}
