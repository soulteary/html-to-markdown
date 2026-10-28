package commonmark

import (
	"strconv"
	"strings"

	"github.com/JohannesKaufmann/html-to-markdown/v2/internal/domutils"
	"golang.org/x/net/html"
)

func hasClass(n *html.Node, class string) bool {
	for _, name := range strings.Fields(domutils.GetAttributeOr(n, "class", "")) {
		if name == class {
			return true
		}
	}
	return false
}

// unwrapCodeTables recognizes the two-column gutter/code layout used by
// syntax highlighters. Ordinary data tables, even those with code, are retained.
func unwrapCodeTables(doc *html.Node) {
	var visit func(*html.Node)
	visit = func(n *html.Node) {
		for child := n.FirstChild; child != nil; {
			next := child.NextSibling
			visit(child)
			child = next
		}
		if n.Type != html.ElementNode || n.Data != "table" || n.Parent == nil {
			return
		}
		rows := domutils.FindAllNodes(n, func(c *html.Node) bool { return domutils.NodeName(c) == "tr" })
		if len(rows) != 1 {
			return
		}
		var cells []*html.Node
		for c := rows[0].FirstChild; c != nil; c = c.NextSibling {
			if c.Type == html.ElementNode {
				if c.Data != "td" {
					return
				}
				cells = append(cells, c)
			}
		}
		if len(cells) != 2 || !hasClass(cells[0], "gutter") || !hasClass(cells[1], "code") {
			return
		}
		// Do not discard captions, extra text, or nested tables with a similar layout.
		for _, container := range []*html.Node{n, rows[0].Parent, rows[0], cells[0], cells[1]} {
			for c := container.FirstChild; c != nil; c = c.NextSibling {
				if c.Type == html.TextNode && strings.TrimSpace(c.Data) != "" {
					return
				}
				if c.Type == html.ElementNode {
					switch container {
					case n:
						if c.Data != "tbody" && c.Data != "tr" {
							return
						}
					case rows[0].Parent:
						if c != rows[0] {
							return
						}
					case cells[0], cells[1]:
						if c.Data != "pre" {
							return
						}
					}
				}
			}
		}
		gutter := domutils.FindAllNodes(cells[0], func(c *html.Node) bool { return domutils.NodeName(c) == "pre" })
		code := domutils.FindAllNodes(cells[1], func(c *html.Node) bool { return domutils.NodeName(c) == "pre" })
		if len(gutter) != 1 || len(code) != 1 {
			return
		}
		numbers, _ := getCodeWithoutTags(gutter[0])
		lines := strings.Split(strings.TrimSpace(string(numbers)), "\n")
		var previous int
		for i, line := range lines {
			value, err := strconv.Atoi(strings.TrimSpace(line))
			if err != nil || value < 0 || (i > 0 && value != previous+1) {
				return
			}
			previous = value
		}
		source, _ := getCodeWithoutTags(code[0])
		if len(lines) != len(strings.Split(strings.TrimSuffix(string(source), "\n"), "\n")) {
			return
		}
		pre := code[0]
		pre.Parent.RemoveChild(pre)
		n.Parent.InsertBefore(pre, n)
		n.Parent.RemoveChild(n)
	}
	visit(doc)
}
