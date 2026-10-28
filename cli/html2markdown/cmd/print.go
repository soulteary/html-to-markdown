package cmd

import (
	"fmt"
	"io"
)

type Printer interface {
	Print(w io.Writer)
}

// - - - - - - - //

type coloredBox struct {
	prefix string
	text   string
}

// ColoredBox retains the existing printer API but now emits plain text.
func ColoredBox(prefix string, text string) Printer {
	return &coloredBox{prefix, text}
}

func (p coloredBox) Print(w io.Writer) {
	fmt.Fprintf(w, "%s: %s\n", p.prefix, p.text)
}

// - - - - - - - //

type paragraph struct {
	text string
}

func Paragraph(text string) Printer {
	return &paragraph{text}
}
func (p paragraph) Print(w io.Writer) {
	fmt.Fprintln(w, p.text)
}

// - - - - - - - //

type codeBlock struct {
	code string
}

func CodeBlock(code string) Printer {
	return &codeBlock{code}
}
func (cb codeBlock) Print(w io.Writer) {
	// TODO: what about indenting multiline?
	fmt.Fprintf(w, "    %s\n", cb.code)
}
