package converter

import (
	"context"
	"fmt"
	"golang.org/x/net/html"
	"testing"
)

type discardWriter struct{}

func (discardWriter) Write(p []byte) (int, error)       { return len(p), nil }
func (discardWriter) WriteString(s string) (int, error) { return len(s), nil }
func (discardWriter) WriteByte(byte) error              { return nil }
func (discardWriter) WriteRune(r rune) (int, error)     { return len(string(r)), nil }

func BenchmarkRenderChildNodes(b *testing.B) {
	for _, size := range []int{0, 1, 8, 9, 1024} {
		b.Run(fmt.Sprint(size), func(b *testing.B) {
			c := NewConverter()
			ctx := newConverterContext(context.Background(), c)
			n := &html.Node{Type: html.ElementNode, Data: "div"}
			for i := 0; i < size; i++ {
				n.AppendChild(&html.Node{Type: html.TextNode, Data: "text"})
			}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				ctx.RenderChildNodes(ctx, discardWriter{}, n)
			}
		})
	}
}
