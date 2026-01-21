package main

// Data structures and methods for managing a file buffer, its content (lines of
// runes), and multiple cursors.

import (
	"strings"
	"time"
)

// Cursor represents a position in the buffer.
type Cursor struct {
	X            int // Column index (0-based).
	Y            int // Row index (0-based).
	PreferredCol int // Remembers the intended column when moving up/down.
}

// HistoryState stores a snapshot of the buffer and cursors for undo/redo.
type HistoryState struct {
	buffer  [][]rune
	cursors []Cursor
}

// Buffer represents an open file and its associated editor state.
type Buffer struct {
	buffer      [][]rune           // Slice of lines, where each line is a slice of runes.
	cursors     []Cursor           // Support for multiple cursors.
	scrollX     int                // Horizontal scroll offset.
	scrollY     int                // Vertical scroll offset.
	filename    string             // Path to the file on disk.
	modified    bool               // True if changes haven't been saved.
	readOnly    bool               // True if the buffer cannot be edited.
	undoStack   []HistoryState     // For undo functionality.
	redoStack   []HistoryState     // For redo functionality.
	fileType    *FileType          // Language-specific settings.
	lspClient   *LSPClient         // Associated LSP client for this buffer.
	diagnostics []Diagnostic       // Errors/warnings for this buffer.
	syntax      *SyntaxHighlighter // Syntax highlighting engine.
	lastModTime time.Time          // Last modified time of the file on disk.
}

// PrimaryCursor returns the first cursor in the list.
func (b *Buffer) PrimaryCursor() *Cursor {
	if len(b.cursors) == 0 {
		b.cursors = append(b.cursors, Cursor{X: 0, Y: 0})
	}
	// The first cursor is usually the main one used for most operations.
	return &b.cursors[0]
}

// AddCursor adds a new cursor at the specified position.
func (b *Buffer) AddCursor(x, y int) {
	b.cursors = append(b.cursors, Cursor{X: x, Y: y, PreferredCol: x})
}

// ClearCursors removes all cursors except the primary one.
func (b *Buffer) ClearCursors() {
	if len(b.cursors) > 1 {
		primary := b.cursors[0]
		b.cursors = []Cursor{primary}
	}
}

// getLineByteOffset calculates the byte index for a given column in a line of
// runes.
func (b *Buffer) getLineByteOffset(line []rune, col int) uint32 {
	// Rune indices != Byte indices for multi-byte characters (UTF-8).
	if col > len(line) {
		col = len(line)
	}
	return uint32(len(string(line[:col])))
}

// getByteOffset calculates the total byte offset from the start of the buffer
// to (row, col).
func (b *Buffer) getByteOffset(row, col int) uint32 {
	var offset uint32
	for i := 0; i < row && i < len(b.buffer); i++ {
		offset += uint32(len(string(b.buffer[i]))) + 1 // +1 for the newline character.
	}

	if row < len(b.buffer) {
		offset += b.getLineByteOffset(b.buffer[row], col)
	}
	return offset
}

// toString converts the entire buffer (slice of lines) into a single string.
func (b *Buffer) toString() string {
	var result strings.Builder
	for i, line := range b.buffer {
		result.WriteString(string(line))
		if i < len(b.buffer)-1 {
			result.WriteString("\n")
		}
	}
	return result.String()
}

// handleEdit is a placeholder for incremental syntax highlighting updates.
func (b *Buffer) handleEdit(startRow, startCol int, bytesRemoved, bytesAdded uint32, oldEndRow int, oldEndColBytes uint32, newEndRow int, newEndColBytes uint32) {
	if b.syntax == nil {
		return
	}

	// We batch syntax updates in editor.go via Reparse, so we don't need
	// incremental updates here.
}
