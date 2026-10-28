package tester

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type ConvertFunc func(html []byte) (markdown []byte, err error)

const suffixInputFile = ".in.html"
const suffixOutputFile = ".out.md"

func getInputFiles(pathOfFolder string) ([]string, error) {
	files, err := os.ReadDir(pathOfFolder)
	if err != nil {
		return nil, fmt.Errorf("error while reading %q folder: %w", pathOfFolder, err)
	}

	var names []string
	for _, file := range files {
		if file.IsDir() {
			return nil, fmt.Errorf("did not expected a folder %q", file.Name())
		}
		if strings.HasSuffix(file.Name(), suffixOutputFile) {
			continue
		}
		if !strings.HasSuffix(file.Name(), suffixInputFile) {
			return nil, fmt.Errorf("only expect in or out files but got %q", file.Name())
		}

		name := strings.TrimSuffix(file.Name(), suffixInputFile)
		names = append(names, name)
	}

	return names, nil
}

type goldenRunner interface {
	reporter
	Name() string
	Run(string, func(*testing.T)) bool
}

func GoldenFiles(t goldenRunner, convert ConvertFunc) {
	pathOfFolder := filepath.Join("./testdata", strings.TrimPrefix(t.Name(), "Test"))
	runs, err := getInputFiles(pathOfFolder)
	if err != nil {
		t.Fatal(err)
	}
	if len(runs) == 0 {
		t.Fatalf("there were no golden files found in %q", pathOfFolder)
	}

	for _, run := range runs {
		t.Run(run, func(t *testing.T) {
			pathOfFile := filepath.Join(pathOfFolder, run+suffixInputFile)
			runGoldenFile(t, pathOfFile, convert)
		})
	}
}

func runGoldenFile(t reporter, pathOfFile string, convert ConvertFunc) {
	t.Helper()
	input, err := os.ReadFile(pathOfFile)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("running golden file test for %q", pathOfFile)
	output, err := convert(input)
	if err != nil {
		t.Fatal(err)
	}
	outputPath := strings.TrimSuffix(pathOfFile, suffixInputFile) + suffixOutputFile
	AssertGolden(t, outputPath, output)
}
