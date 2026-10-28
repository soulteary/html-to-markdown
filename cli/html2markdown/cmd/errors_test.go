package cmd

import (
	"errors"
	"strings"
	"testing"
)

func TestDiagnosticsRemainPlainText(t *testing.T) {
	// Diagnostics must stay plain even when the environment requests color.
	t.Setenv("CLICOLOR_FORCE", "1")
	t.Setenv("FORCE_COLOR", "1")
	t.Setenv("TERM", "xterm-256color")
	for _, mode := range []struct {
		name     string
		terminal bool
	}{{"pipe", false}, {"terminal", true}} {
		t.Run(mode.name, func(t *testing.T) {
			stderr := &FakeFile{mode: modePipe}
			if mode.terminal {
				stderr.mode = modeTerminal
			}
			cli := CLI{Stderr: stderr}
			cli.PrintErr(nil)
			cli.PrintWarn(nil)
			if stderr.Len() != 0 {
				t.Fatalf("nil diagnostics produced %q", stderr.String())
			}
			cli.PrintErr(NewCLIError(errors.New("conversion failed"), Paragraph("Check the input."), CodeBlock("html2markdown --help")))
			cli.PrintWarn(errors.New("existing file skipped"))
			want := "\nerror: conversion failed\n\nCheck the input.\n\n    html2markdown --help\n\n\nwarning: existing file skipped\n\n"
			if stderr.String() != want {
				t.Fatalf("got %q; want %q", stderr.String(), want)
			}
		})
	}
}

func TestUnknownFlagsNeverSuggestAlternatives(t *testing.T) {
	for _, flag := range []string{"--versio", "--opt_strong_delimiter", "-x"} {
		t.Run(flag, func(t *testing.T) {
			stdin := &FakeFile{mode: modePipe}
			stdin.WriteString("<b>text</b>")
			stdout, stderr := &FakeFile{mode: modePipe}, &FakeFile{mode: modePipe}
			Run(stdin, stdout, stderr, []string{"html2markdown", flag}, Release{})
			want := "\nerror: unknown flag: " + flag + "\n\n"
			if stderr.String() != want {
				t.Fatalf("got %q; want %q", stderr.String(), want)
			}
			if stdout.Len() != 0 || strings.Contains(stderr.String(), "Did you mean") {
				t.Fatal("unexpected output for unknown flag")
			}
		})
	}
}
