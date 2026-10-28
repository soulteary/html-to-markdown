package cmd

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestExecuteRecursiveBatchMappings(t *testing.T) {
	for _, absolute := range []bool{false, true} {
		t.Run(map[bool]string{false: "relative", true: "absolute"}[absolute], func(t *testing.T) {
			dir := t.TempDir()
			chdirWithCleanup(t, dir)
			for name, content := range map[string]string{
				"input/page.html": "<b>root</b>", "input/a/page.html": "<b>first</b>", "input/b/page.html": "<b>second</b>", "input/中文/guide.html": "<p>中文</p>", "input/a/skip.txt": "not html",
			} {
				if err := os.MkdirAll(filepath.Dir(name), 0755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(name, []byte(content), 0644); err != nil {
					t.Fatal(err)
				}
			}
			if err := os.MkdirAll("input/directory.html", 0755); err != nil {
				t.Fatal(err)
			}
			pattern := filepath.FromSlash("input/**/*.html")
			if absolute {
				pattern = filepath.Join(dir, pattern)
			}
			stdin, stdout, stderr := &FakeFile{mode: modeTerminal}, &FakeFile{mode: modePipe}, &FakeFile{mode: modePipe}
			Run(stdin, stdout, stderr, []string{"html2markdown", "--input", pattern, "--output", "output/"}, testRelease)
			if stderr.Len() != 0 || stdout.Len() != 0 {
				t.Fatalf("stdout %q stderr %q", stdout.String(), stderr.String())
			}
			entries, err := os.ReadDir("output")
			if err != nil {
				t.Fatal(err)
			}
			got := map[string]string{}
			for _, entry := range entries {
				content, err := os.ReadFile(filepath.Join("output", entry.Name()))
				if err != nil {
					t.Fatal(err)
				}
				got[entry.Name()] = string(content)
			}
			want := map[string]string{"page.md": "**root**", "page.6a156652ad.md": "**first**", "page.3d95a058c2.md": "**second**", "guide.md": "中文"}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("got %#v; want %#v", got, want)
			}
		})
	}
}

func TestExecuteOverlappingRecursivePatterns(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"a.html", "nested/b.html"} {
		full := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte("<b>ok</b>"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	stdin, stdout, stderr := &FakeFile{mode: modeTerminal}, &FakeFile{mode: modePipe}, &FakeFile{mode: modePipe}
	output := filepath.Join(dir, "out") + string(os.PathSeparator)
	Run(stdin, stdout, stderr, []string{"html2markdown", "--input", filepath.Join(dir, "**", "**", "{a,b,a}.html"), "--output", output}, testRelease)
	if stderr.Len() != 0 || stdout.Len() != 0 {
		t.Fatalf("stdout %q stderr %q", stdout.String(), stderr.String())
	}
	entries, err := os.ReadDir(output)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 || entries[0].Name() != "a.md" || entries[1].Name() != "b.md" {
		t.Fatalf("duplicate conversion: %v", entries)
	}
}

func TestExecuteMalformedGlob(t *testing.T) {
	dir := t.TempDir()
	output := filepath.Join(dir, "out") + string(os.PathSeparator)
	stdin, stdout, stderr := &FakeFile{mode: modeTerminal}, &FakeFile{mode: modePipe}, &FakeFile{mode: modePipe}
	Run(stdin, stdout, stderr, []string{"html2markdown", "--input", filepath.Join(dir, "missing", "**", "{a,b}", "["), "--output", output}, testRelease)
	if !strings.Contains(stderr.String(), "syntax error in pattern") || stdout.Len() != 0 {
		t.Fatalf("stdout %q stderr %q", stdout.String(), stderr.String())
	}
	if _, err := os.Stat(output); !os.IsNotExist(err) {
		t.Fatalf("malformed pattern created output: %v", err)
	}
}
