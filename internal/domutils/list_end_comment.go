package domutils

import (
	"context"

	"golang.org/x/net/html"
)

var ListEndCommentData = "THE END"

func AddListEndComments(ctx context.Context, doc *html.Node) {
	node := doc
	for node != nil {
		if nameIsList(node) && nextNameIsList(node) {
			insertComment(node)
		}

		node = GetNextNeighborElement(node)
	}
}

func nameIsList(node *html.Node) bool {
	name := NodeName(node)
	return name == "ul" || name == "ol"
}

func insertComment(listNode *html.Node) {
	comment := &html.Node{
		Type: html.CommentNode,
		Data: ListEndCommentData,
	}
	listNode.Parent.InsertBefore(comment, listNode.NextSibling)
}

func nextNameIsList(startNode *html.Node) bool {
	node := GetNextNeighborNodeExcludingOwnChild(startNode)

	for node != nil {
		name := NodeName(node)
		if name == "ul" || name == "ol" {
			return true
		}
		if name == "li" {
			return false
		}
		if name == "#comment" && node.Data == ListEndCommentData {
			return false
		}

		// If there is any text between two lists
		// they are automatically not connected anymore.
		if node.Type == html.TextNode {
			return false
		}

		// - - - - //

		if name == "hr" {
			// A divider already seperates two lists...
			return false
		}

		// TODO: RunContext.Render()
		// -> get acess to keepRemoveMap

		// TODO: look in the KeepRemoveMap?
		// e.g. ul then script then ul

		node = GetNextNeighborNode(node)
		continue
	}
	return false
}
