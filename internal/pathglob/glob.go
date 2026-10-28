// Package pathglob expands CLI input patterns using only the standard library.
package pathglob

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"
)

// Base returns the literal directory preceding the first wildcard. It is also
// the stable base used to calculate duplicate output-name hashes.
func Base(pattern string) string {
	base, _ := splitPattern(filepath.ToSlash(filepath.Clean(pattern)))
	return filepath.FromSlash(base)
}

func splitPattern(pattern string) (string, string) {
	slash := -1
	for i := 0; i < len(pattern); i++ {
		switch pattern[i] {
		case '\\':
			i++
		case '/':
			slash = i
		case '*', '?', '[', '{':
			return splitAt(pattern, slash)
		}
	}
	return splitAt(pattern, slash)
}

func splitAt(pattern string, slash int) (string, string) {
	if slash < 0 {
		return ".", pattern
	}
	if slash == 0 {
		return "/", pattern[1:]
	}
	base := pattern[:slash]
	// Keep a Windows drive/UNC root absolute rather than drive-relative.
	if volume := filepath.ToSlash(filepath.VolumeName(filepath.FromSlash(pattern))); volume != "" && slash == len(volume) {
		base += "/"
	}
	base = strings.NewReplacer(`\*`, "*", `\?`, "?", `\[`, "[", `\]`, "]", `\{`, "{", `\}`, "}").Replace(base)
	return base, pattern[slash+1:]
}

type segment struct {
	recursive bool
	literal   string
	match     *regexp.Regexp
}

// Files returns non-directory matches in glob traversal order, including hidden files.
// Overlapping alternatives and recursive segments are deduplicated.
// A complete ** segment matches zero or more directory levels. Directory
// symlinks are not traversed after the first wildcard, but an explicitly named
// literal base may itself contain symlinks. Filesystem errors are ignored, as
// with filepath.Glob; malformed patterns return filepath.ErrBadPattern.
func Files(pattern string) ([]string, error) {
	if pattern == "" {
		return nil, nil
	}
	pattern = filepath.ToSlash(filepath.Clean(pattern))
	if err := validatePattern(pattern); err != nil {
		return nil, err
	}
	_, originalTail := splitPattern(pattern)
	literalPrefix := len(pattern) - len(originalTail)
	// A brace group following ** gathers its matched prefixes in sorted order
	// before expanding the remaining path. Preserve that order for output hashes.
	if head, rest, ok := recursiveAlternative(pattern); ok {
		prefixes := make(map[string]bool)
		err := expandBraces(head, func(expanded string) error {
			if !validExpansion(expanded, literalPrefix) {
				return nil
			}
			base, tail := splitPattern(expanded)
			paths, err := walkPattern(filepath.FromSlash(base), tail, true)
			for _, name := range paths {
				prefixes[name] = true
			}
			return err
		})
		if err != nil {
			return nil, err
		}
		var ordered []string
		for name := range prefixes {
			ordered = append(ordered, name)
		}
		sort.Slice(ordered, func(i, j int) bool { return filepath.ToSlash(ordered[i]) < filepath.ToSlash(ordered[j]) })
		var matches []string
		seen := make(map[string]bool)
		add := func(name string) {
			if !seen[name] {
				seen[name] = true
				matches = append(matches, name)
			}
		}
		for _, prefix := range ordered {
			info, err := os.Lstat(prefix)
			if err != nil {
				continue
			}
			if rest == "" {
				if !info.IsDir() {
					add(prefix)
				}
				continue
			}
			if !info.IsDir() {
				continue
			}
			paths, err := walkPattern(prefix, rest, false)
			if err != nil {
				return nil, err
			}
			for _, name := range paths {
				add(name)
			}
		}
		return matches, nil
	}
	var matches []string
	err := expandBraces(pattern, func(expanded string) error {
		if !validExpansion(expanded, literalPrefix) {
			return nil
		}
		base, tail := splitPattern(expanded)
		current, err := walkPattern(filepath.FromSlash(base), tail, false)
		if err != nil {
			return err
		}
		matches = mergeMatches(matches, current)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return matches, nil
}

func walkPattern(base, tail string, includeDirs bool) ([]string, error) {
	var current []string
	found := make(map[string]bool)
	add := func(name string) {
		if !found[name] {
			found[name] = true
			current = append(current, name)
		}
	}
	var parts []segment
	for _, part := range strings.Split(tail, "/") {
		seg, err := compileSegment(part)
		if err != nil {
			return nil, err
		}
		parts = append(parts, seg)
	}
	type state struct {
		dir   string
		index int
	}
	visited := make(map[state]bool)
	var walk func(string, int)
	walk = func(dir string, index int) {
		key := state{dir, index}
		if visited[key] {
			return
		}
		visited[key] = true
		if index == len(parts) {
			return
		}
		part := parts[index]
		if part.recursive {
			if includeDirs && index == len(parts)-1 {
				add(dir)
			}
			walk(dir, index+1)
			entries, _ := os.ReadDir(dir)
			for _, entry := range entries {
				name := filepath.Join(dir, entry.Name())
				if entry.IsDir() {
					walk(name, index)
				} else if index == len(parts)-1 {
					add(name)
				}
			}
			return
		}
		visit := func(name string, isDir bool) {
			if index == len(parts)-1 {
				if includeDirs || !isDir {
					add(name)
				}
			} else if isDir {
				walk(name, index+1)
			}
		}
		if part.match == nil {
			if part.literal == "." && dir != base {
				return
			}
			name := filepath.Join(dir, part.literal)
			if info, err := os.Lstat(name); err == nil {
				visit(name, info.IsDir())
			}
			return
		}
		entries, _ := os.ReadDir(dir)
		for _, entry := range entries {
			if part.match.MatchString(entry.Name()) {
				visit(filepath.Join(dir, entry.Name()), entry.IsDir())
			}
		}
	}

	walk(base, 0)
	return current, nil
}

func recursiveAlternative(pattern string) (head, rest string, ok bool) {
	last, open := -1, -1
	for i := 0; i < len(pattern); i++ {
		switch pattern[i] {
		case '\\':
			i++
		case '[':
			_, end, err := characterClass(pattern, i)
			if err != nil {
				return "", "", false
			}
			i = end
		case '{':
			if open < 0 {
				open = i
			}
		case '}':
			last = i
		}
	}
	if last < 0 {
		return "", "", false
	}
	if open < 0 {
		return "", "", false
	}
	recursive := false
	for _, part := range strings.Split(pattern[:open], "/") {
		if part == "**" {
			recursive = true
		}
	}
	if !recursive {
		return "", "", false
	}
	end := strings.IndexByte(pattern[last+1:], '/')
	if end < 0 {
		return pattern, "", true
	}
	end += last + 1
	return pattern[:end], pattern[end+1:], true
}

// Expand alternatives lazily, including nested groups and alternatives that
// contain separators. Escaped braces and braces in character classes are literal.
func expandBraces(pattern string, yield func(string) error) error {
	open := -1
	depth := 0
	start := 0
	var choices []string
	for i := 0; i < len(pattern); i++ {
		switch pattern[i] {
		case '\\':
			i++
			if i >= len(pattern) {
				return filepath.ErrBadPattern
			}
		case '[':
			_, end, err := characterClass(pattern, i)
			if err != nil {
				return err
			}
			i = end
		case '{':
			if depth == 0 {
				open = i
				start = i + 1
			}
			depth++
		case ',':
			if depth == 1 {
				choices = append(choices, pattern[start:i])
				start = i + 1
			}
		case '}':
			if depth == 0 {
				return filepath.ErrBadPattern
			}
			depth--
			if depth == 0 {
				if start < i {
					choices = append(choices, pattern[start:i])
				}
				for _, choice := range choices {
					if err := expandBraces(pattern[:open]+choice+pattern[i+1:], yield); err != nil {
						return err
					}
				}
				return nil
			}
		}
	}
	if depth != 0 {
		return filepath.ErrBadPattern
	}
	return yield(pattern)
}

func compileSegment(pattern string) (segment, error) {
	if pattern == "**" {
		return segment{recursive: true}, nil
	}
	var expression, literal strings.Builder
	expression.WriteString("(?s)^")
	meta := false
	for i := 0; i < len(pattern); {
		r, size := utf8.DecodeRuneInString(pattern[i:])
		switch r {
		case '\\':
			i += size
			if i >= len(pattern) {
				return segment{}, filepath.ErrBadPattern
			}
			r, size = utf8.DecodeRuneInString(pattern[i:])
			expression.WriteString(regexp.QuoteMeta(string(r)))
			literal.WriteRune(r)
		case '*':
			meta = true
			expression.WriteString(".*")
		case '?':
			meta = true
			expression.WriteByte('.')
		case '[':
			class, end, err := characterClass(pattern, i)
			if err != nil {
				return segment{}, err
			}
			meta = true
			expression.WriteString(class)
			i = end
			size = 1
		default:
			expression.WriteString(regexp.QuoteMeta(string(r)))
			literal.WriteRune(r)
		}
		i += size
	}
	if !meta {
		return segment{literal: literal.String()}, nil
	}
	expression.WriteByte('$')
	compiled, err := regexp.Compile(expression.String())
	if err != nil {
		return segment{}, filepath.ErrBadPattern
	}
	return segment{match: compiled}, nil
}

// Translate glob classes to RE2 classes. Both ! and ^ negate. A leading or
// trailing hyphen is literal; ranges retain the preceding literal endpoint.
func characterClass(pattern string, start int) (string, int, error) {
	i := start + 1
	var out strings.Builder
	out.WriteByte('[')
	if i < len(pattern) && (pattern[i] == '!' || pattern[i] == '^') {
		out.WriteByte('^')
		i++
	}
	if i >= len(pattern) || pattern[i] == ']' {
		return "", 0, filepath.ErrBadPattern
	}
	quote := func(r rune) string {
		if strings.ContainsRune(`\-[]^`, r) {
			return `\` + string(r)
		}
		return string(r)
	}
	var previous rune
	hasPrevious := false
	for i < len(pattern) && pattern[i] != ']' {
		r, size := utf8.DecodeRuneInString(pattern[i:])
		i += size
		if r == '-' && hasPrevious && i < len(pattern) && pattern[i] != ']' {
			if pattern[i] == '\\' {
				i++
				if i >= len(pattern) {
					return "", 0, filepath.ErrBadPattern
				}
			}
			end, n := utf8.DecodeRuneInString(pattern[i:])
			i += n
			if previous <= end {
				out.WriteString(quote(previous))
				out.WriteByte('-')
				out.WriteString(quote(end))
			}
			hasPrevious = false
			continue
		}
		if r == '\\' {
			if i >= len(pattern) {
				return "", 0, filepath.ErrBadPattern
			}
			r, size = utf8.DecodeRuneInString(pattern[i:])
			i += size
		}
		out.WriteString(quote(r))
		previous = r
		hasPrevious = r != utf8.MaxRune
	}
	if i >= len(pattern) {
		return "", 0, filepath.ErrBadPattern
	}
	out.WriteByte(']')
	return out.String(), i, nil
}

// Merge alternatives without reordering the first traversal unnecessarily.
// Recursive glob walks are deterministic but not globally lexicographic.
func mergeMatches(left, right []string) []string {
	pending := append([]string(nil), right...)
	var merged []string
	seen := make(map[string]bool)
	add := func(name string) {
		if !seen[name] {
			seen[name] = true
			merged = append(merged, name)
		}
	}
	for _, candidate := range left {
		if len(pending) == 0 || filepath.ToSlash(candidate) < filepath.ToSlash(pending[0]) {
			add(candidate)
			continue
		}
		if candidate == pending[0] {
			add(candidate)
			pending = pending[1:]
			continue
		}
		add(pending[0])
		pending = pending[1:]
		position := 0
		for position < len(pending) && filepath.ToSlash(pending[position]) < filepath.ToSlash(candidate) {
			position++
		}
		pending = append(pending, "")
		copy(pending[position+1:], pending[position:])
		pending[position] = candidate
	}
	for _, name := range pending {
		add(name)
	}
	return merged
}

// Validate the whole pattern before visiting the filesystem, including suffixes
// of alternatives that may have no matches.
func validatePattern(pattern string) error {
	depth := 0
	for i := 0; i < len(pattern); i++ {
		switch pattern[i] {
		case '\\':
			i++
			if i >= len(pattern) {
				return filepath.ErrBadPattern
			}
		case '[':
			class, end, err := characterClass(pattern, i)
			if err != nil {
				return err
			}
			if _, err := regexp.Compile(class); err != nil {
				return filepath.ErrBadPattern
			}
			i = end
		case '{':
			depth++
		case '}':
			depth--
			if depth < 0 {
				return filepath.ErrBadPattern
			}
		}
	}
	if depth != 0 {
		return filepath.ErrBadPattern
	}
	return nil
}

func validExpansion(expanded string, literalPrefix int) bool {
	for _, part := range strings.Split(expanded[literalPrefix:], "/") {
		if part == "" || part == ".." {
			return false
		}
	}
	return true
}
