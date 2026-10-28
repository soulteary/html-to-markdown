package tester

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

var updateGoldenFiles = flag.Bool("update", false, "update golden test files")

// AssertGolden compares actual bytes with a fixture, or updates it when the
// test binary receives -update. Normal test runs never write fixture files.
func AssertGolden(t reporter, path string, actual []byte) {
	t.Helper()
	if err := checkGolden(path, actual, *updateGoldenFiles); err != nil {
		t.Error(err)
	}
}

func checkGolden(path string, actual []byte, update bool) error {
	if update {
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return fmt.Errorf("create golden directory for %q: %w", path, err)
		}
		if err := os.WriteFile(path, actual, 0644); err != nil {
			return fmt.Errorf("update golden file %q: %w", path, err)
		}
		return nil
	}
	expected, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read golden file %q (use -update to create fixtures): %w", path, err)
	}
	if !bytes.Equal(expected, actual) {
		return fmt.Errorf("golden file %q differs\nExpected: %q\nGot: %q", path, expected, actual)
	}
	return nil
}
