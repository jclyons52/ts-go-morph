package tsmorph

import (
	"fmt"
	"sort"
)

// textEdit replaces the byte range [start, end) with newText.
type textEdit struct {
	start, end int
	newText    string
}

// applyTextEdits applies edits to text. Edits must not overlap; they are
// applied in descending position order so offsets remain valid.
func applyTextEdits(text string, edits []textEdit) (string, error) {
	sorted := make([]textEdit, len(edits))
	copy(sorted, edits)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].start > sorted[j].start })

	for i, e := range sorted {
		if e.start < 0 || e.end > len(text) || e.start > e.end {
			return "", fmt.Errorf("tsmorph: edit out of bounds: [%d, %d) in text of length %d", e.start, e.end, len(text))
		}
		if i+1 < len(sorted) && sorted[i+1].end > e.start {
			return "", fmt.Errorf("tsmorph: overlapping edits: [%d, %d) and [%d, %d)",
				sorted[i+1].start, sorted[i+1].end, e.start, e.end)
		}
	}
	result := text
	for _, e := range sorted {
		result = result[:e.start] + e.newText + result[e.end:]
	}
	return result, nil
}

// replaceRange edits the file by replacing the byte range [start, end) with
// newText and re-parsing. All existing Node wrappers for this file are
// forgotten.
func (s *SourceFile) replaceRange(start, end int, newText string) {
	s.applyEdits([]textEdit{{start: start, end: end, newText: newText}})
}

// applyEdits applies text edits to the file's current text, stores the
// result in the in-memory overlay, and rebuilds the project's program.
// All existing Node wrappers for this file are forgotten.
func (s *SourceFile) applyEdits(edits []textEdit) {
	newText, err := applyTextEdits(s.Text(), edits)
	if err != nil {
		panic(err)
	}
	s.project.fsys.setOverlay(s.path, newText)
	s.project.invalidate()
	s.project.bumpGeneration(s.canonicalPath())
}

// lineStart returns the byte offset of the start of the line containing pos.
func lineStart(text string, pos int) int {
	for pos > 0 && text[pos-1] != '\n' {
		pos--
	}
	return pos
}

// lineEnd returns the byte offset just past the end of the line containing
// pos (the '\n' itself, or len(text)).
func lineEnd(text string, pos int) int {
	for pos < len(text) && text[pos] != '\n' {
		pos++
	}
	return pos
}

// indentOfLine returns the leading whitespace of the line containing pos.
func indentOfLine(text string, pos int) string {
	start := lineStart(text, pos)
	i := start
	for i < len(text) && (text[i] == ' ' || text[i] == '\t') {
		i++
	}
	return text[start:i]
}

// isBlank reports whether s consists only of whitespace.
func isBlank(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] != ' ' && s[i] != '\t' && s[i] != '\r' && s[i] != '\n' {
			return false
		}
	}
	return true
}
