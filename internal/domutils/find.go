// Adapted from github.com/JohannesKaufmann/dom v0.3.1.
// Copyright (c) 2024 Johannes Kaufmann.
// Licensed under the MIT license; see LICENSE.dom.

package domutils

import "golang.org/x/net/html"

func FindFirstNode(startNode *html.Node, matchFn func(node *html.Node) bool) *html.Node {
	nextFunc := initGetNeighbor(
		FirstChildNode,
		NextSiblingNode,
		func(node *html.Node) bool {
			// We should not get higher up than the startNode...
			return node == startNode
		},
	)

	child := startNode.FirstChild
	for child != nil {
		if matchFn(child) {
			return child
		}

		child = nextFunc(child)
	}
	return nil
}

func FindAllNodes(startNode *html.Node, matchFn func(node *html.Node) bool) (foundNodes []*html.Node) {
	nextFunc := initGetNeighbor(
		FirstChildNode,
		NextSiblingNode,
		func(node *html.Node) bool {
			// We should not get higher up than the startNode...
			return node == startNode
		},
	)

	child := startNode.FirstChild
	for child != nil {
		if matchFn(child) {
			foundNodes = append(foundNodes, child)
		}

		child = nextFunc(child)
	}
	return foundNodes
}
