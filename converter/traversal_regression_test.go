package converter

import (
	"bytes"
	"context"
	"fmt"
	"golang.org/x/net/html"
	"strings"
	"testing"
)

func TestRenderChildNodesRetainsOriginalSnapshot(t *testing.T) {
	for _, size := range []int{0, 1, 8, 9, 1024} {
		t.Run(fmt.Sprint(size), func(t *testing.T) {
			parent := &html.Node{Type: html.ElementNode, Data: "div"}
			var originals []*html.Node
			var want strings.Builder
			for i := 0; i < size; i++ {
				n := &html.Node{Type: html.ElementNode, Data: "item", Attr: []html.Attribute{{Key: "id", Val: fmt.Sprint(i)}}}
				parent.AppendChild(n)
				originals = append(originals, n)
				fmt.Fprintf(&want, "%d,", i)
			}
			c := NewConverter()
			c.Register.Renderer(func(_ Context, w Writer, n *html.Node) RenderStatus {
				if n == originals[0] {
					parent.RemoveChild(n)
					last := originals[len(originals)-1]
					if last.Parent != nil {
						parent.RemoveChild(last)
					}
					if len(originals) > 2 {
						second := originals[1]
						parent.RemoveChild(second)
						parent.AppendChild(second)
					}
					parent.AppendChild(&html.Node{Type: html.ElementNode, Data: "item", Attr: []html.Attribute{{Key: "id", Val: "inserted"}}})
				}
				w.WriteString(n.Attr[0].Val + ",")
				return RenderSuccess
			}, PriorityStandard)
			ctx := newConverterContext(context.Background(), c)
			var output bytes.Buffer
			ctx.RenderChildNodes(ctx, &output, parent)
			if output.String() != want.String() {
				t.Fatalf("mutating renderer changed snapshot order: got %q; want %q", output.String(), want.String())
			}
		})
	}
}

func TestRenderChildSnapshotAllocations(t *testing.T) {
	for _, size := range []int{0, 1, 8, 9, 1024} {
		t.Run(fmt.Sprint(size), func(t *testing.T) {
			c := NewConverter()
			ctx := newConverterContext(context.Background(), c)
			parent := &html.Node{Type: html.ElementNode, Data: "div"}
			for i := 0; i < size; i++ {
				parent.AppendChild(&html.Node{Type: html.TextNode, Data: "text"})
			}
			want := float64(0)
			if size > 8 {
				want = 1
			}
			got := testing.AllocsPerRun(100, func() { ctx.RenderChildNodes(ctx, discardWriter{}, parent) })
			if got != want {
				t.Fatalf("snapshot allocations: got %v; want %v", got, want)
			}
		})
	}
}
