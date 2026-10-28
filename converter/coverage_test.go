package converter

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"golang.org/x/net/html"
)

type failingReader struct{ err error }

func (r failingReader) Read([]byte) (int, error) { return 0, r.err }

type testPlugin struct {
	name string
	err  error
}

func (p testPlugin) Name() string          { return p.name }
func (p testPlugin) Init(*Converter) error { return p.err }

func TestConversionErrors(t *testing.T) {
	sentinel := errors.New("configuration rejected")
	c := NewConverter(func(*Converter) error { return sentinel })
	if _, err := c.ConvertNode(nil); !errors.Is(err, sentinel) {
		t.Fatalf("option error: %v", err)
	}
	for _, p := range []testPlugin{{name: ""}, {name: "broken", err: sentinel}} {
		c = NewConverter(WithPlugins(p))
		if _, err := c.ConvertString("<p>x</p>"); err == nil {
			t.Fatal("invalid plugin accepted")
		}
	}
	if _, err := NewConverter().ConvertReader(failingReader{sentinel}); !errors.Is(err, sentinel) {
		t.Fatalf("reader error: %v", err)
	}
}

func TestContextAndRenderingAPIs(t *testing.T) {
	type key struct{}
	c := NewConverter()
	c.Register.Renderer(func(ctx Context, w Writer, n *html.Node) RenderStatus {
		if ctx.Value(key{}) != "present" {
			t.Error("caller context lost")
		}
		w.WriteString(ctx.AssembleAbsoluteURL(ctx, "a", "page"))
		ctx.RenderNodes(ctx, w, &html.Node{Type: html.TextNode, Data: "!"})
		return RenderSuccess
	}, PriorityStandard)
	out, err := c.ConvertString("", WithDomain("https://example.com/docs/"), WithContext(context.WithValue(context.Background(), key{}, "present")))
	if err != nil || out != "https://example.com/docs/page!" {
		t.Fatalf("got %q, %v", out, err)
	}
	if GetDomain(context.Background()) != "" {
		t.Fatal("missing domain should be empty")
	}
	if assembleAbsoluteURL(context.Background(), "a", "page") != "" {
		t.Fatal("missing URL assembler should be empty")
	}
	state := newGlobalState()
	ctx := state.provideGlobalState(context.Background())
	SetState(ctx, "key", "wrong type")
	UpdateState(ctx, "key", func(v int) int { return v + 2 })
	if GetState[int](ctx, "key") != 2 {
		t.Fatal("mismatched state should update from zero")
	}
	if parseBaseDomain("%") != nil {
		t.Fatal("invalid base should be rejected")
	}
}

func TestEscapeBoundaryCases(t *testing.T) {
	c := NewConverter()
	if got := string(c.escapeContent([]byte("中\x00文"))); got != "中\ufffd文" {
		t.Fatalf("NUL replacement: %q", got)
	}
	if got := c.unEscapeContent([]byte{'x', placeholderByte}); !bytes.Equal(got, []byte("x")) {
		t.Fatalf("trailing marker: %q", got)
	}
	c = NewConverter(WithEscapeMode(EscapeModeDisabled))
	input := []byte{'x', 0, placeholderByte}
	if !bytes.Equal(c.escapeContent(input), input) || !bytes.Equal(c.unEscapeContent(input), input) {
		t.Fatal("disabled escaping changed bytes")
	}
}

func TestInputMarkerCleanupAndExhaustion(t *testing.T) {
	t.Run("nil and moved nodes", func(t *testing.T) {
		doc := &html.Node{Type: html.DocumentNode}
		child := &html.Node{Type: html.TextNode, Data: "\a\uf002"}
		elem := &html.Node{Type: html.ElementNode, Data: "span", Attr: []html.Attribute{{Key: "title", Val: "\a\uf002"}}}
		elem.AppendChild(child)
		doc.AppendChild(elem)
		restore, cleanup := protectInputMarkers(doc, "")
		if restore(nil) != nil {
			t.Fatal("nil output lost")
		}
		if child.Data == "\a\uf002" {
			t.Fatal("markers were not protected")
		}
		doc.RemoveChild(elem)
		replacement := &html.Node{Type: html.ElementNode, Data: "p", Attr: elem.Attr}
		replacement.AppendChild(&html.Node{Type: html.TextNode, Data: child.Data})
		doc.AppendChild(replacement)
		cleanup()
		if child.Data != "\a\uf002" || elem.Attr[0].Val != "\a\uf002" || replacement.FirstChild.Data != "\a\uf002" || replacement.Attr[0].Val != "\a\uf002" {
			t.Fatal("detached or newly created nodes were not restored")
		}
	})
	for name, missing := range map[string]rune{"all private use runes": 0, "one available rune": 0xe000} {
		t.Run(name, func(t *testing.T) {
			var content strings.Builder
			for _, bounds := range [][2]rune{{0xe000, 0xf8ff}, {0xf0000, 0xffffd}, {0x100000, 0x10fffd}} {
				for r := bounds[0]; r <= bounds[1]; r++ {
					if r != missing {
						content.WriteRune(r)
					}
				}
			}
			content.WriteString("\a\uf002")
			original := content.String()
			doc := &html.Node{Type: html.TextNode, Data: original}
			restore, cleanup := protectInputMarkers(doc, "")
			if got := string(restore([]byte(doc.Data))); got != original {
				t.Fatal("exhausted-token restoration changed literal input")
			}
			cleanup()
			if doc.Data != original {
				t.Fatal("cleanup changed literal input")
			}
		})
	}
}

func TestInputMarkersRespectEncodedTokenCandidates(t *testing.T) {
	doc := &html.Node{Type: html.TextNode, Data: "%EE%80%80\a"}
	restore, cleanup := protectInputMarkers(doc, "")
	if got := string(restore([]byte(doc.Data))); got != "%EE%80%80\a" {
		t.Fatalf("encoded candidate changed: %q", got)
	}
	cleanup()
	if doc.Data != "%EE%80%80\a" {
		t.Fatal("cleanup changed encoded candidate")
	}
}
