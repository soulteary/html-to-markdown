// Adapted from github.com/JohannesKaufmann/dom v0.3.1.
// Copyright (c) 2024 Johannes Kaufmann.
// Licensed under the MIT license; see LICENSE.dom.

package domutils

import "golang.org/x/net/html"

func AllNodes(startNode *html.Node) (allNodes []*html.Node) {
	var finder func(*html.Node)
	finder = func(node *html.Node) {
		allNodes = append(allNodes, node)

		for child := node.FirstChild; child != nil; child = child.NextSibling {
			finder(child)
		}
	}

	finder(startNode)

	return allNodes
}

func AllChildNodes(node *html.Node) (children []*html.Node) {
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		children = append(children, child)
	}
	return children
}

func FirstChildNode(node *html.Node) *html.Node {
	return node.FirstChild
}

func FirstChildElement(node *html.Node) *html.Node {
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == html.ElementNode {
			return child
		}
	}
	return nil
}

func PrevSiblingNode(node *html.Node) *html.Node {
	return node.PrevSibling
}

func NextSiblingNode(node *html.Node) *html.Node {
	return node.NextSibling
}

func NextSiblingElement(node *html.Node) *html.Node {
	for sibling := node.NextSibling; sibling != nil; sibling = sibling.NextSibling {
		if sibling.Type == html.ElementNode {
			return sibling
		}
	}
	return nil
}
