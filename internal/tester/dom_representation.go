package tester

import (
	"strings"

	"github.com/soulteary/html-to-markdown/v2/internal/domutils"
	"golang.org/x/net/html"
)

func ExpectRepresentation(t reporter, doc *html.Node, name string, expectedHtml string) {
	actualHtml := domutils.RenderRepresentation(doc)

	actualHtml = strings.TrimSpace(actualHtml)
	expectedHtml = strings.TrimSpace(expectedHtml)

	if actualHtml != expectedHtml {
		t.Errorf("%s: expected \n%s\nbut got\n%s", name, expectedHtml, actualHtml)
	}
}
