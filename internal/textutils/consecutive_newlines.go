package textutils

import "bytes"

func TrimUnnecessaryHardLineBreaks(content []byte) []byte {
	content = bytes.ReplaceAll(content, []byte("  \n\n"), []byte("\n\n"))
	content = bytes.ReplaceAll(content, []byte("  \n  \n"), []byte("\n\n"))
	content = bytes.ReplaceAll(content, []byte("  \n \n"), []byte("\n\n"))
	// out = bytes.ReplaceAll(out, []byte("\n  \n"), []byte("\n\n"))

	return content
}

// TrimConsecutiveNewlines retains at most two consecutive newlines, ignoring
// ASCII spaces between them. It returns independent storage without changing
// input bytes, including malformed UTF-8 and non-ASCII whitespace.
func TrimConsecutiveNewlines(input []byte) []byte {
	if len(input) == 0 {
		return nil
	}
	if bytes.IndexByte(input, '\n') == -1 {
		return bytes.Clone(input)
	}
	// Size the result after accounting for dropped newline/space runs. A long
	// run may collapse to only two bytes, so reserving len(input) wastes memory.
	removed, newlineCount, pending := 0, 0, 0
	for i, char := range input {
		if char == ' ' {
			continue
		}
		if char == '\n' {
			newlineCount++
			if newlineCount > 2 {
				removed += i + 1 - pending
			}
		} else {
			newlineCount = 0
		}
		pending = i + 1
	}
	if removed == 0 {
		return bytes.Clone(input)
	}
	result := make([]byte, 0, len(input)-removed)
	newlineCount, pending = 0, 0
	for i, char := range input {
		if char == ' ' {
			continue
		}
		if char == '\n' {
			newlineCount++
			if newlineCount <= 2 {
				result = append(result, input[pending:i+1]...)
			}
		} else {
			newlineCount = 0
			result = append(result, input[pending:i+1]...)
		}
		pending = i + 1
	}
	return append(result, input[pending:]...)
}
