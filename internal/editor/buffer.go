package editor

import (
	"strings"
	"time"
)

// Buffer represents an in-memory text buffer
type Buffer struct {
	lines      []string
	cursorLine int
	cursorCol  int
	modified   bool
	history    *EditHistory

	// Search state
	searchTerm    string
	searchResults []SearchResult
	searchIndex   int
}

// SearchResult represents a search match
type SearchResult struct {
	Line int
	Col  int
	Text string
}

// BufferState represents a snapshot for undo/redo
type BufferState struct {
	Lines      []string
	CursorLine int
	CursorCol  int
	Timestamp  time.Time
}

// EditHistory manages undo/redo
type EditHistory struct {
	past    []BufferState
	future  []BufferState
	maxSize int
}

// NewBuffer creates a new text buffer
func NewBuffer(content string) *Buffer {
	lines := strings.Split(content, "\n")
	if len(lines) == 0 {
		lines = []string{""}
	}

	return &Buffer{
		lines:      lines,
		cursorLine: 0,
		cursorCol:  0,
		modified:   false,
		history: &EditHistory{
			past:    make([]BufferState, 0, 100),
			future:  make([]BufferState, 0, 100),
			maxSize: 100,
		},
	}
}

// GetContent returns the full buffer content
func (b *Buffer) GetContent() string {
	return strings.Join(b.lines, "\n")
}

// GetLines returns all lines
func (b *Buffer) GetLines() []string {
	return b.lines
}

// GetLine returns a specific line
func (b *Buffer) GetLine(lineNum int) string {
	if lineNum < 0 || lineNum >= len(b.lines) {
		return ""
	}
	return b.lines[lineNum]
}

// LineCount returns the number of lines
func (b *Buffer) LineCount() int {
	return len(b.lines)
}

// CursorPosition returns current cursor position
func (b *Buffer) CursorPosition() (line, col int) {
	return b.cursorLine, b.cursorCol
}

// InsertRune inserts a character at cursor position
func (b *Buffer) InsertRune(r rune) {
	// Validate cursor position
	if b.cursorLine < 0 || b.cursorLine >= len(b.lines) {
		return
	}

	line := b.lines[b.cursorLine]

	// Clamp cursor column to valid range
	if b.cursorCol < 0 {
		b.cursorCol = 0
	}
	if b.cursorCol > len(line) {
		b.cursorCol = len(line)
	}

	b.recordState()

	before := line[:b.cursorCol]
	after := line[b.cursorCol:]
	b.lines[b.cursorLine] = before + string(r) + after
	b.cursorCol++
	b.modified = true
}

// InsertNewline inserts a newline at cursor position
func (b *Buffer) InsertNewline() {
	// Validate cursor position
	if b.cursorLine < 0 || b.cursorLine >= len(b.lines) {
		return
	}

	line := b.lines[b.cursorLine]

	// Clamp cursor column to valid range
	if b.cursorCol < 0 {
		b.cursorCol = 0
	}
	if b.cursorCol > len(line) {
		b.cursorCol = len(line)
	}

	b.recordState()

	before := line[:b.cursorCol]
	after := line[b.cursorCol:]

	b.lines[b.cursorLine] = before
	// Insert new line after current
	b.lines = append(b.lines[:b.cursorLine+1], append([]string{after}, b.lines[b.cursorLine+1:]...)...)

	b.cursorLine++
	b.cursorCol = 0
	b.modified = true
}

// DeleteChar deletes character before cursor (backspace)
func (b *Buffer) DeleteChar() {
	// Validate cursor position
	if b.cursorLine < 0 || b.cursorLine >= len(b.lines) {
		return
	}

	if b.cursorCol == 0 {
		// At start of line - merge with previous
		if b.cursorLine > 0 {
			b.recordState()
			prevLine := b.lines[b.cursorLine-1]
			currLine := b.lines[b.cursorLine]
			b.lines[b.cursorLine-1] = prevLine + currLine
			b.lines = append(b.lines[:b.cursorLine], b.lines[b.cursorLine+1:]...)
			b.cursorLine--
			b.cursorCol = len(prevLine)
			b.modified = true
		}
	} else {
		// Ensure cursorCol is within bounds
		line := b.lines[b.cursorLine]
		if b.cursorCol > len(line) {
			b.cursorCol = len(line)
		}
		if b.cursorCol > 0 {
			b.recordState()
			before := line[:b.cursorCol-1]
			after := line[b.cursorCol:]
			b.lines[b.cursorLine] = before + after
			b.cursorCol--
			b.modified = true
		}
	}
}

// MoveCursor moves the cursor
func (b *Buffer) MoveCursorUp() {
	if b.cursorLine > 0 {
		b.cursorLine--
		if b.cursorCol > len(b.lines[b.cursorLine]) {
			b.cursorCol = len(b.lines[b.cursorLine])
		}
	}
}

func (b *Buffer) MoveCursorDown() {
	if b.cursorLine < len(b.lines)-1 {
		b.cursorLine++
		if b.cursorCol > len(b.lines[b.cursorLine]) {
			b.cursorCol = len(b.lines[b.cursorLine])
		}
	}
}

func (b *Buffer) MoveCursorLeft() {
	if b.cursorCol > 0 {
		b.cursorCol--
	} else if b.cursorLine > 0 {
		// Move to end of previous line
		b.cursorLine--
		b.cursorCol = len(b.lines[b.cursorLine])
	}
}

func (b *Buffer) MoveCursorRight() {
	if b.cursorCol < len(b.lines[b.cursorLine]) {
		b.cursorCol++
	} else if b.cursorLine < len(b.lines)-1 {
		// Move to start of next line
		b.cursorLine++
		b.cursorCol = 0
	}
}

// MoveCursorTo moves cursor to specific position
func (b *Buffer) MoveCursorTo(line, col int) {
	if line >= 0 && line < len(b.lines) {
		b.cursorLine = line
		if col >= 0 && col <= len(b.lines[line]) {
			b.cursorCol = col
		}
	}
}

// IsModified returns whether buffer has been modified
func (b *Buffer) IsModified() bool {
	return b.modified
}

// SetModified sets the modified flag
func (b *Buffer) SetModified(modified bool) {
	b.modified = modified
}

// recordState saves current state for undo
func (b *Buffer) recordState() {
	// Only keep last 100 states
	if len(b.history.past) >= b.history.maxSize {
		b.history.past = b.history.past[1:]
	}

	state := BufferState{
		Lines:      make([]string, len(b.lines)),
		CursorLine: b.cursorLine,
		CursorCol:  b.cursorCol,
		Timestamp:  time.Now(),
	}
	copy(state.Lines, b.lines)

	b.history.past = append(b.history.past, state)
	b.history.future = nil // Clear redo stack
}

// Undo reverts to previous state
func (b *Buffer) Undo() bool {
	if len(b.history.past) == 0 {
		return false
	}

	// Save current state to future
	current := BufferState{
		Lines:      make([]string, len(b.lines)),
		CursorLine: b.cursorLine,
		CursorCol:  b.cursorCol,
		Timestamp:  time.Now(),
	}
	copy(current.Lines, b.lines)
	b.history.future = append(b.history.future, current)

	// Restore previous state
	prev := b.history.past[len(b.history.past)-1]
	b.history.past = b.history.past[:len(b.history.past)-1]

	b.lines = make([]string, len(prev.Lines))
	copy(b.lines, prev.Lines)
	b.cursorLine = prev.CursorLine
	b.cursorCol = prev.CursorCol
	b.modified = true

	return true
}

// Redo re-applies previously undone change
func (b *Buffer) Redo() bool {
	if len(b.history.future) == 0 {
		return false
	}

	// Save current to past
	current := BufferState{
		Lines:      make([]string, len(b.lines)),
		CursorLine: b.cursorLine,
		CursorCol:  b.cursorCol,
		Timestamp:  time.Now(),
	}
	copy(current.Lines, b.lines)
	b.history.past = append(b.history.past, current)

	// Restore future state
	next := b.history.future[len(b.history.future)-1]
	b.history.future = b.history.future[:len(b.history.future)-1]

	b.lines = make([]string, len(next.Lines))
	copy(b.lines, next.Lines)
	b.cursorLine = next.CursorLine
	b.cursorCol = next.CursorCol
	b.modified = true

	return true
}

// Find searches for the given term and stores results
func (b *Buffer) Find(term string, caseSensitive bool) int {
	b.searchTerm = term
	b.searchResults = nil
	b.searchIndex = -1

	if term == "" {
		return 0
	}

	searchTerm := term
	if !caseSensitive {
		searchTerm = strings.ToLower(term)
	}

	for lineNum, line := range b.lines {
		searchLine := line
		if !caseSensitive {
			searchLine = strings.ToLower(line)
		}

		col := 0
		for {
			idx := strings.Index(searchLine[col:], searchTerm)
			if idx == -1 {
				break
			}

			actualCol := col + idx
			b.searchResults = append(b.searchResults, SearchResult{
				Line: lineNum,
				Col:  actualCol,
				Text: line[actualCol : actualCol+len(term)],
			})
			col = actualCol + 1
		}
	}

	if len(b.searchResults) > 0 {
		b.searchIndex = 0
		// Jump to first result
		b.cursorLine = b.searchResults[0].Line
		b.cursorCol = b.searchResults[0].Col
	}

	return len(b.searchResults)
}

// FindNext jumps to the next search result
func (b *Buffer) FindNext() bool {
	if len(b.searchResults) == 0 {
		return false
	}

	b.searchIndex = (b.searchIndex + 1) % len(b.searchResults)
	result := b.searchResults[b.searchIndex]
	b.cursorLine = result.Line
	b.cursorCol = result.Col
	return true
}

// FindPrevious jumps to the previous search result
func (b *Buffer) FindPrevious() bool {
	if len(b.searchResults) == 0 {
		return false
	}

	b.searchIndex--
	if b.searchIndex < 0 {
		b.searchIndex = len(b.searchResults) - 1
	}

	result := b.searchResults[b.searchIndex]
	b.cursorLine = result.Line
	b.cursorCol = result.Col
	return true
}

// ClearSearch clears the search results
func (b *Buffer) ClearSearch() {
	b.searchTerm = ""
	b.searchResults = nil
	b.searchIndex = -1
}

// GetSearchInfo returns current search information
func (b *Buffer) GetSearchInfo() (term string, current int, total int) {
	if len(b.searchResults) == 0 {
		return b.searchTerm, 0, 0
	}
	return b.searchTerm, b.searchIndex + 1, len(b.searchResults)
}

// ReplaceAll replaces all occurrences of the search term
func (b *Buffer) ReplaceAll(searchTerm, replacement string, caseSensitive bool) int {
	count := b.Find(searchTerm, caseSensitive)
	if count == 0 {
		return 0
	}

	b.recordState()

	// Replace in reverse order to avoid index shifting
	for i := len(b.searchResults) - 1; i >= 0; i-- {
		result := b.searchResults[i]
		line := b.lines[result.Line]
		before := line[:result.Col]
		after := line[result.Col+len(searchTerm):]
		b.lines[result.Line] = before + replacement + after
	}

	b.modified = true
	b.ClearSearch()
	return count
}
