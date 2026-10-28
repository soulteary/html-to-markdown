package textutils

import (
	"bytes"
	"testing"
	"unicode/utf8"
)

// Retain the v2.0.0 algorithm as a compatibility oracle, independently using
// decoded runes and a separate whitespace buffer instead of byte offsets.
func referenceTrimNewlines(input []byte) []byte {
	var result, spaces []byte
	count := 0
	for i := 0; i < len(input); {
		r, size := utf8.DecodeRune(input[i:])
		switch r {
		case '\n':
			count++
			if count <= 2 {
				result = append(result, spaces...)
				result = append(result, '\n')
			}
			spaces = spaces[:0]
		case ' ':
			spaces = append(spaces, input[i:i+size]...)
		default:
			count = 0
			result = append(result, spaces...)
			result = append(result, input[i:i+size]...)
			spaces = spaces[:0]
		}
		i += size
	}
	return append(result, spaces...)
}

func checkNewlineBuffer(t *testing.T, input []byte) {
	t.Helper()
	original := bytes.Clone(input)
	got := TrimConsecutiveNewlines(input)
	want := referenceTrimNewlines(input)
	if !bytes.Equal(got, want) {
		t.Fatalf("input %q: got %q; want %q", input, got, want)
	}
	if !bytes.Equal(input, original) {
		t.Fatal("input changed during normalization")
	}
	if len(got) > 0 {
		got[0] ^= 0xff
		if !bytes.Equal(input, original) {
			t.Fatal("output aliases input")
		}
	}
	if len(input) == 0 && got != nil {
		t.Fatal("empty input should preserve nil result")
	}
}

func TestNewlineBufferCompatibility(t *testing.T) {
	alphabet := []byte{'a', ' ', '\n', '\r', '\t', 0xff, 0x80}
	var enumerate func([]byte, int)
	enumerate = func(prefix []byte, depth int) {
		checkNewlineBuffer(t, prefix)
		if depth == 0 {
			return
		}
		for _, b := range alphabet {
			enumerate(append(bytes.Clone(prefix), b), depth-1)
		}
	}
	enumerate(nil, 5)
	for _, input := range [][]byte{[]byte("中文 🌟\n  \n   \n  tail"), []byte("\n\u00a0\n\n\n"), bytes.Repeat([]byte(" "), 4096), []byte{0xe4, 0xb8, ' ', '\n', 0xad, 0xff}} {
		checkNewlineBuffer(t, input)
	}
}

func FuzzNewlineBufferCompatibility(f *testing.F) {
	for _, input := range [][]byte{nil, []byte("one\n \n   \n two "), []byte("中文\n\n🌟"), {0xff, '\n', 0x80}} {
		f.Add(input)
	}
	f.Fuzz(func(t *testing.T, input []byte) { checkNewlineBuffer(t, input) })
}

func TestCollapsedNewlineRunDoesNotKeepInputSizedCapacity(t *testing.T) {
	input := append(append([]byte("start"), bytes.Repeat([]byte("\n "), 32768)...), []byte("end")...)
	output := TrimConsecutiveNewlines(input)
	if string(output) != "start\n \n end" {
		t.Fatalf("long run normalized incorrectly: %q", output)
	}
	if cap(output) > 64 {
		t.Fatalf("short output retains excessive capacity: %d", cap(output))
	}
	checkNewlineBuffer(t, input)
}
