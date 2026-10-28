package domutils

import (
	"golang.org/x/net/html"
)

func RemoveRedundant(doc *html.Node, matchFn func(*html.Node, *html.Node) bool) {
	for _, node := range AllNodes(doc) {
		if hasSameTypeAncestor(node, matchFn) {
			UnwrapNode(node)
		}
	}

}

func hasSameTypeAncestor(n *html.Node, matchFn func(*html.Node, *html.Node) bool) bool {
	if !matchFn(n, n) {
		return false
	}

	for p := n.Parent; p != nil; p = p.Parent {
		if matchFn(n, p) {
			return true
		}
	}

	return false
}
