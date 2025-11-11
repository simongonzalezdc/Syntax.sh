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
	b.recordState()

	line := b.lines[b.cursorLine]
	before := line[:b.cursorCol]
	after := line[b.cursorCol:]
	b.lines[b.cursorLine] = before + string(r) + after
	b.cursorCol++
	b.modified = true
}

// InsertNewline inserts a newline at cursor position
func (b *Buffer) InsertNewline() {
	b.recordState()

	line := b.lines[b.cursorLine]
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
		b.recordState()
		line := b.lines[b.cursorLine]
		before := line[:b.cursorCol-1]
		after := line[b.cursorCol:]
		b.lines[b.cursorLine] = before + after
		b.cursorCol--
		b.modified = true
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
