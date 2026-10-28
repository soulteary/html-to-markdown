package tester

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/net/html"
)

type recordedReporter struct {
	name     string
	messages []string
}

func (r *recordedReporter) Helper()                           {}
func (r *recordedReporter) Logf(string, ...any)               {}
func (r *recordedReporter) Error(a ...any)                    { r.messages = append(r.messages, fmt.Sprint(a...)) }
func (r *recordedReporter) Errorf(f string, a ...any)         { r.Error(fmt.Sprintf(f, a...)) }
func (r *recordedReporter) Fatal(a ...any)                    { r.Error(a...); panic(reportedFatal{}) }
func (r *recordedReporter) Fatalf(f string, a ...any)         { r.Fatal(fmt.Sprintf(f, a...)) }
func (r *recordedReporter) Name() string                      { return r.name }
func (r *recordedReporter) Run(string, func(*testing.T)) bool { panic("unexpected subtest") }

type reportedFatal struct{}

func expectDiagnostic(t *testing.T, want string, fn func(*recordedReporter)) {
	t.Helper()
	r := &recordedReporter{}
	func() {
		defer func() {
			if p := recover(); p != nil {
				if _, ok := p.(reportedFatal); !ok {
					panic(p)
				}
			}
		}()
		fn(r)
	}()
	if got := strings.Join(r.messages, "\n"); !strings.Contains(got, want) {
		t.Fatalf("diagnostic %q missing %q", got, want)
	}
}

type brokenInput struct{}

func (brokenInput) Read([]byte) (int, error) { return 0, io.ErrUnexpectedEOF }

func TestParsingAndRepresentationDiagnostics(t *testing.T) {
	root := Parse(t, " <p>one</p><p>two</p> ", "")
	first := Parse(t, "<p>one</p><p>two</p>", "p")
	if root.Data != "body" || first.FirstChild.Data != "one" {
		t.Fatal("wrong parsed node")
	}
	expectDiagnostic(t, "could not find", func(r *recordedReporter) {
		if Parse(r, "text", "missing") != nil {
			t.Error("missing node should be nil")
		}
	})
	expectDiagnostic(t, "unexpected EOF", func(r *recordedReporter) { parseReader(r, brokenInput{}, "") })
	doc := &html.Node{Type: html.TextNode, Data: "actual"}
	expectDiagnostic(t, "expected", func(r *recordedReporter) { ExpectRepresentation(r, doc, "representation", "wrong") })
}

func TestGoldenDiagnostics(t *testing.T) {
	path := filepath.Join(t.TempDir(), "fixture.golden")
	previous := *updateGoldenFiles
	defer func() { *updateGoldenFiles = previous }()
	*updateGoldenFiles = true
	AssertGolden(t, path, []byte("expected"))
	*updateGoldenFiles = false
	AssertGolden(t, path, []byte("expected"))
	expectDiagnostic(t, "differs", func(r *recordedReporter) { AssertGolden(r, path, []byte("different")) })
	expectDiagnostic(t, "missing.in.html", func(r *recordedReporter) { runGoldenFile(r, filepath.Join(t.TempDir(), "missing.in.html"), nil) })
	input := filepath.Join(t.TempDir(), "input.in.html")
	if err := os.WriteFile(input, []byte("input"), 0600); err != nil {
		t.Fatal(err)
	}
	expectDiagnostic(t, "conversion rejected", func(r *recordedReporter) {
		runGoldenFile(r, input, func([]byte) ([]byte, error) { return nil, errors.New("conversion rejected") })
	})
}

func TestGetInputFilesValidation(t *testing.T) {
	if _, err := getInputFiles(filepath.Join(t.TempDir(), "missing")); err == nil {
		t.Fatal("missing directory accepted")
	}
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "nested"), 0700); err != nil {
		t.Fatal(err)
	}
	if _, err := getInputFiles(dir); err == nil || !strings.Contains(err.Error(), "folder") {
		t.Fatalf("nested directory accepted: %v", err)
	}
	dir = t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "other.txt"), nil, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := getInputFiles(dir); err == nil || !strings.Contains(err.Error(), "only expect") {
		t.Fatalf("unknown fixture accepted: %v", err)
	}
}

func TestGoldenFilesCoverage(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := os.MkdirAll("testdata/GoldenFilesCoverage", 0700); err != nil {
		t.Fatal(err)
	}
	for name, content := range map[string]string{"case.in.html": "input", "case.out.md": "output"} {
		if err := os.WriteFile(filepath.Join("testdata/GoldenFilesCoverage", name), []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
	}
	GoldenFiles(t, func(b []byte) ([]byte, error) {
		if string(b) != "input" {
			t.Errorf("wrong fixture input: %q", b)
		}
		return []byte("output"), nil
	})
	expectDiagnostic(t, "folder", func(r *recordedReporter) { r.name = "TestMissing"; GoldenFiles(r, nil) })
	if err := os.Mkdir("testdata/Empty", 0700); err != nil {
		t.Fatal(err)
	}
	expectDiagnostic(t, "no golden files", func(r *recordedReporter) { r.name = "TestEmpty"; GoldenFiles(r, nil) })
}
