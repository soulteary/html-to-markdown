package pathglob

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func fixture(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	for _, name := range []string{"root.html", "a.html", "a/one.html", "a/b/two.html", "b/one.html", "b/data.txt", ".hidden.html", ".hidden/inside.html", "empty/.keep", "a-/dash.html", "z.html", "中.html", "braces{a}/one.html", "[literal]/one.html"} {
		full := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, nil, 0644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func relative(base string, paths []string) []string {
	var result []string
	for _, p := range paths {
		rel, _ := filepath.Rel(base, p)
		result = append(result, filepath.ToSlash(rel))
	}
	return result
}

// Recorded against doublestar v4.10.0 with FilesOnly + NoFollow. Repeated **
// matches are intentionally deduplicated to avoid converting a file twice.
func TestCompatibilityCorpus(t *testing.T) {
	data, err := os.ReadFile("testdata/compatibility.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		Pattern    string
		Matches    []string
		BadPattern bool
	}
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	dir := fixture(t)
	t.Chdir(dir)
	for _, tc := range cases {
		for _, absolute := range []bool{false, true} {
			name := "relative/" + tc.Pattern
			if absolute {
				name = "absolute/" + tc.Pattern
			}
			t.Run(name, func(t *testing.T) {
				pattern := filepath.FromSlash(tc.Pattern)
				if absolute && pattern != "" {
					pattern = filepath.Join(dir, pattern)
				}
				got, err := Files(pattern)
				if tc.BadPattern {
					if !errors.Is(err, filepath.ErrBadPattern) {
						t.Fatalf("want malformed-pattern error, got %v", err)
					}
					return
				}
				if err != nil {
					t.Fatal(err)
				}
				if absolute {
					got = relative(dir, got)
				} else {
					for i := range got {
						got[i] = filepath.ToSlash(got[i])
					}
				}
				if !reflect.DeepEqual(got, tc.Matches) {
					t.Fatalf("got %q; want %q", got, tc.Matches)
				}
			})
		}
	}
}

func TestBase(t *testing.T) {
	for _, tc := range []struct{ pattern, want string }{
		{"input/**/*.html", "input"}, {"input/*.html", "input"}, {"*.html", "."},
		{"input/page.html", "input"}, {"input/{a,b}/**", "input"}, {"/input/**/*.html", "/input"},
		{"../input/**/*.html", "../input"}, {"input/./**/*.html", "input"},
	} {
		if got := Base(filepath.FromSlash(tc.pattern)); filepath.ToSlash(got) != tc.want {
			t.Errorf("%q: got %q; want %q", tc.pattern, got, tc.want)
		}
	}
	if runtime.GOOS == "windows" {
		for _, pattern := range []string{`C:\**\*.html`, `C:/**/*.html`} {
			if got := Base(pattern); got != `C:\` {
				t.Errorf("drive root: %q", got)
			}
		}
	} else if got := Base(`input/\[literal\]/*.html`); got != "input/[literal]" {
		t.Fatal(got)
	}
}

func TestEscapesAndNewlines(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("backslash is a Windows path separator; newline filenames are unsupported")
	}
	dir := fixture(t)
	for _, pattern := range []string{`\[literal\]/*.html`, `braces\{a\}/*.html`} {
		got, err := Files(filepath.Join(dir, pattern))
		if err != nil || len(got) != 1 {
			t.Fatalf("%q: %q %v", pattern, got, err)
		}
	}
	name := filepath.Join(dir, "line\nbreak.html")
	if err := os.WriteFile(name, nil, 0644); err != nil {
		t.Fatal(err)
	}
	got, err := Files(filepath.Join(dir, "line?break.html"))
	if err != nil || !reflect.DeepEqual(got, []string{name}) {
		t.Fatalf("newline match: %q %v", got, err)
	}
	if _, err := Files(filepath.Join(dir, `missing\`)); !errors.Is(err, filepath.ErrBadPattern) {
		t.Fatalf("invalid escape: %v", err)
	}
}

func TestNoFollowSymlinks(t *testing.T) {
	dir := fixture(t)
	for link, target := range map[string]string{"link": "a", "file.html": "root.html", "broken.html": "missing", "a/back": ".."} {
		if err := os.Symlink(target, filepath.Join(dir, filepath.FromSlash(link))); err != nil {
			t.Skipf("symlinks unavailable: %v", err)
		}
	}
	got, err := Files(filepath.Join(dir, "**", "*.html"))
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range relative(dir, got) {
		if strings.HasPrefix(name, "link/") || strings.Contains(name, "/back/") {
			t.Fatalf("followed symlink: %s", name)
		}
	}
	for _, pattern := range []string{"link/*.html", "{link,a}/*.html"} {
		got, err := Files(filepath.Join(dir, filepath.FromSlash(pattern)))
		if err != nil {
			t.Fatal(err)
		}
		want := []string{"link/one.html"}
		if strings.HasPrefix(pattern, "{") {
			want = []string{"a/one.html", "link/one.html"}
		}
		if rel := relative(dir, got); !reflect.DeepEqual(rel, want) {
			t.Fatalf("explicit prefix %q: %q", pattern, rel)
		}
	}
	got, err = Files(filepath.Join(dir, "*"))
	if err != nil {
		t.Fatal(err)
	}
	found := map[string]bool{}
	for _, p := range got {
		found[filepath.Base(p)] = true
	}
	for _, name := range []string{"link", "file.html", "broken.html"} {
		if !found[name] {
			t.Errorf("missing non-directory symlink %q", name)
		}
	}
	got, err = Files(filepath.Join(dir, "*", "one.html"))
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range relative(dir, got) {
		if name == "link/one.html" {
			t.Fatal("followed wildcard directory symlink")
		}
	}
}
