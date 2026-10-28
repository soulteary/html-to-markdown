package converter

import (
	"net/url"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/soulteary/html-to-markdown/v2/marker"
	"golang.org/x/net/html"
)

var encodedBytes = regexp.MustCompile(`(?:%[0-9a-fA-F]{2})+`)

// protectInputMarkers separates literal input from internal render markers.
// Tokens are chosen per document, so input containing private-use characters
// cannot collide with them. Restoration runs only after every post-renderer.
func protectInputMarkers(doc *html.Node, domain string) (func([]byte) []byte, func()) {
	var nodes []*html.Node
	bell := string(marker.MarkerEscaping)
	newline := string(marker.MarkerCodeBlockNewline)
	containsMarker := func(text string) bool {
		return strings.Contains(text, bell) || strings.Contains(text, newline)
	}
	found := false
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		nodes = append(nodes, n)
		found = found || containsMarker(n.Data)
		for _, a := range n.Attr {
			found = found || containsMarker(a.Val)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
	if !found {
		return func(result []byte) []byte { return result }, func() {}
	}

	var content strings.Builder
	// Relative URLs may introduce characters from the base during rendering.
	// They must be excluded just like characters already present in the DOM.
	content.WriteString(domain)
	for _, n := range nodes {
		content.WriteString(n.Data)
		for _, a := range n.Attr {
			content.WriteString(a.Val)
		}
	}
	// Pick unused private-use runes, including decoded URL bytes. Single-rune
	// tokens cannot be fabricated by merging neighboring input text nodes.
	input := content.String()
	decoded := encodedBytes.ReplaceAllStringFunc(input, func(value string) string {
		text, _ := url.PathUnescape(value)
		return text
	})
	used := map[rune]bool{
		marker.MarkerEscaping:         true,
		marker.MarkerCodeBlockNewline: true,
	}
	for _, r := range input + decoded {
		used[r] = true
	}
	candidate := rune(0xe000)
	nextToken := func() string {
		for candidate <= 0x10fffd {
			if candidate == 0xf900 {
				candidate = 0xf0000
			}
			if candidate == 0xffffe {
				candidate = 0x100000
			}
			r := candidate
			candidate++
			if !used[r] {
				used[r] = true
				return string(r)
			}
		}
		// If every private-use rune occurs, a sequence longer than all input
		// cannot be constructed by concatenating its text nodes either.
		candidate++
		return strings.Repeat("\ue000", utf8.RuneCountInString(input)+int(candidate-0x10fffd))
	}
	bellToken, newlineToken := nextToken(), nextToken()
	protect := strings.NewReplacer(bell, bellToken, newline, newlineToken)
	restore := strings.NewReplacer(bellToken, bell, newlineToken, newline)
	for _, n := range nodes {
		if n.Type == html.TextNode {
			n.Data = protect.Replace(n.Data)
		}
		for i := range n.Attr {
			n.Attr[i].Val = protect.Replace(n.Attr[i].Val)
		}
	}
	cleanup := func() {
		// Pre-renderers may move text or create new nodes. Restore both retained
		// original nodes and the current tree, without undoing normal DOM changes.
		for _, n := range nodes {
			if n.Type == html.TextNode {
				n.Data = restore.Replace(n.Data)
			}
			for i := range n.Attr {
				n.Attr[i].Val = restore.Replace(n.Attr[i].Val)
			}
		}
		var restoreTree func(*html.Node)
		restoreTree = func(n *html.Node) {
			if n.Type == html.TextNode {
				n.Data = restore.Replace(n.Data)
			}
			for i := range n.Attr {
				n.Attr[i].Val = restore.Replace(n.Attr[i].Val)
			}
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				restoreTree(c)
			}
		}
		restoreTree(doc)
	}
	return func(result []byte) []byte {
		if result == nil {
			return nil
		}
		return []byte(restore.Replace(string(result)))
	}, cleanup
}
