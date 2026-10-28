package tester

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestCheckGolden(t *testing.T) {
	for _, content := range [][]byte{nil, []byte("a\n\n"), []byte("中文\r\n"), {0, 7, 255}} {
		path := filepath.Join(t.TempDir(), "nested", "fixture.golden")
		if err := checkGolden(path, content, false); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("missing fixture: %v", err)
		}
		if _, err := os.Stat(filepath.Dir(path)); !errors.Is(err, os.ErrNotExist) {
			t.Fatal("comparison created a directory")
		}
		if err := checkGolden(path, content, true); err != nil {
			t.Fatal(err)
		}
		if err := checkGolden(path, content, false); err != nil {
			t.Fatal(err)
		}
		changed := append(append([]byte{}, content...), 'x')
		err := checkGolden(path, changed, false)
		if err == nil || !strings.Contains(err.Error(), strconv.Quote(path)) || !strings.Contains(err.Error(), "Expected:") || !strings.Contains(err.Error(), "Got:") {
			t.Fatalf("missing mismatch diagnostic: %v", err)
		}
		stored, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(stored, content) {
			t.Fatal("comparison modified fixture")
		}
		if err := checkGolden(path, changed, true); err != nil {
			t.Fatal(err)
		}
		if err := checkGolden(path, changed, false); err != nil {
			t.Fatal(err)
		}
	}
}

func TestCheckGoldenMissingPathDiagnostic(t *testing.T) {
	// A Windows-style path must remain recognizable after diagnostic escaping,
	// including when this test runs on Unix.
	path := filepath.Join(t.TempDir(), `C:\fixtures\missing.golden`)
	err := checkGolden(path, nil, false)
	if err == nil || !strings.Contains(err.Error(), strconv.Quote(path)) {
		t.Fatalf("missing quoted fixture path: %v", err)
	}
}

func TestCheckGoldenIOErrors(t *testing.T) {
	dir := t.TempDir()
	if err := checkGolden(dir, nil, false); err == nil {
		t.Fatal("reading directory should fail")
	}
	if err := checkGolden(dir, nil, true); err == nil {
		t.Fatal("writing directory should fail")
	}
	file := filepath.Join(dir, "file")
	if err := os.WriteFile(file, nil, 0644); err != nil {
		t.Fatal(err)
	}
	if err := checkGolden(filepath.Join(file, "child"), nil, true); err == nil {
		t.Fatal("file used as a parent directory should fail")
	}
}
