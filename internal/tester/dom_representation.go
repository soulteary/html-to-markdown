package tester

import (
	"strings"
	"testing"

	"github.com/JohannesKaufmann/html-to-markdown/v2/internal/domutils"
	"golang.org/x/net/html"
)

func ExpectRepresentation(t *testing.T, doc *html.Node, name string, expectedHtml string) {
	actualHtml := domutils.RenderRepresentation(doc)

	actualHtml = strings.TrimSpace(actualHtml)
	expectedHtml = strings.TrimSpace(expectedHtml)

	if actualHtml != expectedHtml {
		t.Errorf("%s: expected \n%s\nbut got\n%s", name, expectedHtml, actualHtml)
	}
}
