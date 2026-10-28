package table

import (
	"golang.org/x/net/html"
	"testing"
)

func TestOptionDefaultsAndInvalidValues(t *testing.T) {
	p := &tablePlugin{cellPaddingBehavior: CellPaddingBehaviorNone}
	for _, opt := range []option{WithSpanCellBehavior(""), WithNewlineBehavior(""), WithCellPaddingBehavior("")} {
		if err := opt(p); err != nil {
			t.Fatal(err)
		}
	}
	if p.cellPaddingBehavior != CellPaddingBehaviorAligned {
		t.Fatal("empty padding should restore default")
	}
	for _, opt := range []option{WithNewlineBehavior("invalid"), WithCellPaddingBehavior("invalid")} {
		if err := opt(p); err == nil {
			t.Fatal("invalid option accepted")
		}
	}
	strike := &html.Node{Type: html.ElementNode, Data: "del"}
	table := &html.Node{Type: html.ElementNode, Data: "table"}
	strike.AppendChild(table)
	if !hasProblematicParentNode(table) {
		t.Fatal("strikethrough table parent should be rejected")
	}
}
