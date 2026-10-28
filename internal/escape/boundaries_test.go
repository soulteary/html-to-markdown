package escape

import "testing"

func TestBackslashAndPublicRuneLookup(t *testing.T) {
	if IsBackslash([]byte(`\`), 0) != 1 || IsBackslash([]byte("x"), 0) != -1 {
		t.Fatal("incorrect backslash classification")
	}
	if got := GetNextAsRune([]byte("a中"), 0); got != '中' {
		t.Fatalf("next Unicode rune: %q", got)
	}
}
