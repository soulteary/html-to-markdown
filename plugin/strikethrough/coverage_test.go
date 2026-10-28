package strikethrough

import "testing"

func TestUnEscaperIgnoresOtherPunctuation(t *testing.T) {
	p := NewStrikethroughPlugin().(*strikethroughPlugin)
	if p.handleUnEscapers([]byte("*word"), 0) != -1 {
		t.Fatal("non-tilde punctuation should fall through")
	}
}
