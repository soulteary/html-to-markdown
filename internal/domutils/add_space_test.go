package domutils_test

import (
	"context"
	"testing"

	"github.com/JohannesKaufmann/html-to-markdown/v2/internal/domutils"
	"github.com/JohannesKaufmann/html-to-markdown/v2/internal/tester"
	"golang.org/x/net/html"
)

func TestAddSpace(t *testing.T) {
	runs := []struct {
		desc     string
		input    string
		expected string
	}{
		{
			desc:  "space needed before & after",
			input: `before<strong><code>inline code</code></strong>after`,
			expected: `
├─body
│ ├─#text "before "
│ ├─strong
│ │ ├─code
│ │ │ ├─#text "inline code"
│ ├─#text " after"
			`,
		},
		{
			desc:  "no surrounding text",
			input: `<strong><code>inline code</code></strong>`,
			expected: `
├─body
│ ├─strong
│ │ ├─code
│ │ │ ├─#text "inline code"
			`,
		},
	}
	for _, run := range runs {
		t.Run(run.desc, func(t *testing.T) {
			doc := tester.Parse(t, run.input, "")

			domutils.AddSpace(context.Background(), doc, func(n *html.Node) bool {
				name := domutils.NodeName(n)
				if name == "strong" || name == "b" {
					return true
				}
				if name == "em" || name == "i" {
					return true
				}
				return false
			}, func(n *html.Node) bool {
				return domutils.NodeName(n) == "code"
			})

			tester.ExpectRepresentation(t, doc, "output", run.expected)
		})
	}
}
