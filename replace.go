package main

// Vim-style range replacement feature (s/pattern/replace/g). Works within a
// visual selection and supports regex patterns and flags.

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/nsf/termbox-go"
)

// startReplaceMode captures the current visual selection and enters Replace mode.
func (e *Editor) startReplaceMode() {
	b := e.activeBuffer()
	if b == nil {
		return
	}

	// Calculate and store the bounds of the selection to be operated on.
	if e.visualStartY < b.PrimaryCursor().Y || (e.visualStartY == b.PrimaryCursor().Y && e.visualStartX < b.PrimaryCursor().X) {
		e.replaceSelStartX = e.visualStartX
		e.replaceSelStartY = e.visualStartY
		e.replaceSelEndX = b.PrimaryCursor().X
		e.replaceSelEndY = b.PrimaryCursor().Y
	} else {
		e.replaceSelStartX = b.PrimaryCursor().X
		e.replaceSelStartY = b.PrimaryCursor().Y
		e.replaceSelEndX = e.visualStartX
		e.replaceSelEndY = e.visualStartY
	}

	// Logging for debugging purposes.
	e.addLog("Replace", fmt.Sprintf("Selection: (%d,%d) to (%d,%d)", e.replaceSelStartY, e.replaceSelStartX, e.replaceSelEndY, e.replaceSelEndX))
	e.addLog("Replace", fmt.Sprintf("Mode: %v", e.mode))

	// Handle Visual Line mode by selecting entire lines.
	if e.mode == ModeVisualLine {
		e.replaceSelStartX = 0
		if e.replaceSelEndY < len(b.buffer) {
			e.replaceSelEndX = len(b.buffer[e.replaceSelEndY])
		}
	} else {
		// In character-wise visual mode, include the character at the end position.
		if e.replaceSelEndY < len(b.buffer) && e.replaceSelEndX < len(b.buffer[e.replaceSelEndY]) {
			e.replaceSelEndX++
		}
	}

	// Initialize the replace input prompt with a starting slash.
	e.replaceInput = []rune{'/'}
	e.replaceMatches = []MatchRange{}
	e.mode = ModeReplace
}

// handleReplaceMode processes input for the replace prompt (/pattern/replacement/flags).
func (e *Editor) handleReplaceMode(ev termbox.Event) {
	switch ev.Key {
	case termbox.KeyEsc:
		e.mode = ModeNormal
		e.replaceInput = []rune{}
		e.replaceMatches = []MatchRange{}
	case termbox.KeyEnter:
		// User finished typing; execute the replacement.
		e.executeReplace()
	case termbox.KeyBackspace, termbox.KeyBackspace2:
		if len(e.replaceInput) > 0 {
			e.replaceInput = e.replaceInput[:len(e.replaceInput)-1]
			e.updateReplacePreview()
		} else {
			e.mode = ModeNormal
		}
	case termbox.KeySpace:
		e.replaceInput = append(e.replaceInput, ' ')
		e.updateReplacePreview()
	default:
		if ev.Ch != 0 {
			e.replaceInput = append(e.replaceInput, ev.Ch)
			e.updateReplacePreview() // Live preview of matches as user types.
		}
	}
}

// parseReplaceCommand splits the raw input string into pattern, replacement, and flags.
func parseReplaceCommand(input string) (pattern, replacement string, globalFlag, ignoreCaseFlag bool, err error) {
	// Expected syntax: /pattern/replacement/[flags]
	if !strings.HasPrefix(input, "/") {
		return "", "", false, false, nil
	}

	parts := []string{}
	current := ""
	escaped := false
	slashCount := 0

	// Custom parser to handle escaped slashes within patterns.
	for i, ch := range input {
		if i == 0 {
			slashCount++
			continue
		}

		if escaped {
			current += string(ch)
			escaped = false
			continue
		}

		if ch == '\\' {
			escaped = true
			current += string(ch)
			continue
		}

		if ch == '/' {
			slashCount++
			parts = append(parts, current)
			current = ""
			continue
		}

		current += string(ch)
	}

	if current != "" || slashCount >= 2 {
		parts = append(parts, current)
	}

	if len(parts) < 2 {
		return "", "", false, false, nil
	}

	pattern = parts[0]
	replacement = parts[1]

	// Check optional flags (e.g., 'g' for global, 'i' for case-insensitive).
	if len(parts) >= 3 {
		flags := parts[2]
		globalFlag = strings.Contains(flags, "g")
		ignoreCaseFlag = strings.Contains(flags, "i")
	}

	return pattern, replacement, globalFlag, ignoreCaseFlag, nil
}

// updateReplacePreview finds and highlights matches in the buffer based on the current prompt.
func (e *Editor) updateReplacePreview() {
	e.replaceMatches = []MatchRange{}

	input := string(e.replaceInput)
	pattern, _, globalFlag, _, err := parseReplaceCommand(input)
	if err != nil || pattern == "" {
		return
	}

	// Always use case-insensitive matching by default (?i).
	regexPattern := "(?i)" + pattern

	re, err := regexp.Compile(regexPattern)
	if err != nil {
		return
	}

	b := e.activeBuffer()
	if b == nil {
		return
	}

	// Scan each line within the selected range for matches.
	for lineIdx := e.replaceSelStartY; lineIdx <= e.replaceSelEndY && lineIdx < len(b.buffer); lineIdx++ {
		line := b.buffer[lineIdx]
		lineStr := string(line)

		startCol := 0
		endCol := len(line)

		if lineIdx == e.replaceSelStartY {
			startCol = e.replaceSelStartX
		}
		if lineIdx == e.replaceSelEndY {
			endCol = e.replaceSelEndX
		}

		if startCol >= len(line) {
			continue
		}

		searchStr := lineStr[startCol:endCol]

		if globalFlag {
			matches := re.FindAllStringIndex(searchStr, -1)
			for _, match := range matches {
				e.replaceMatches = append(e.replaceMatches, MatchRange{
					startLine: lineIdx,
					startCol:  startCol + match[0],
					endLine:   lineIdx,
					endCol:    startCol + match[1],
				})
			}
		} else {
			match := re.FindStringIndex(searchStr)
			if match != nil {
				e.replaceMatches = append(e.replaceMatches, MatchRange{
					startLine: lineIdx,
					startCol:  startCol + match[0],
					endLine:   lineIdx,
					endCol:    startCol + match[1],
				})
			}
		}
	}
}

// executeReplace performs the actual string transformation in the active buffer.
func (e *Editor) executeReplace() {
	input := string(e.replaceInput)
	pattern, replacement, globalFlag, ignoreCaseFlag, err := parseReplaceCommand(input)

	// Logging for debugging purposes.
	e.addLog("Replace", fmt.Sprintf("Input: '%s'", input))
	e.addLog("Replace", fmt.Sprintf("Pattern: '%s', Replacement: '%s', g=%v, i=%v", pattern, replacement, globalFlag, ignoreCaseFlag))

	if err != nil {
		e.message = "Invalid regex pattern"
		e.mode = ModeNormal
		e.replaceInput = []rune{}
		e.replaceMatches = []MatchRange{}
		return
	}

	if pattern == "" {
		e.message = "No pattern specified"
		e.mode = ModeNormal
		e.replaceInput = []rune{}
		e.replaceMatches = []MatchRange{}
		return
	}

	// Always use case-insensitive matching by default (?i).
	regexPattern := "(?i)" + pattern

	re, err := regexp.Compile(regexPattern)
	if err != nil {
		e.message = "Invalid regex pattern"
		e.mode = ModeNormal
		e.replaceInput = []rune{}
		e.replaceMatches = []MatchRange{}
		return
	}

	b := e.activeBuffer()
	if b == nil {
		return
	}

	// Save state for Undo/Redo support before modifying text.
	e.saveState()

	replacementCount := 0

	e.addLog("Replace", fmt.Sprintf("Starting replacement: lines %d-%d", e.replaceSelStartY, e.replaceSelEndY))

	// Important: Iterate backwards from top to bottom through lines,
	// but this loop actually goes from replaceSelEndY down to replaceSelStartY.
	// This helps maintain line index stability during multi-line operations.
	for lineIdx := e.replaceSelEndY; lineIdx >= e.replaceSelStartY && lineIdx < len(b.buffer); lineIdx-- {
		line := b.buffer[lineIdx]
		lineStr := string(line)

		startCol := 0
		endCol := len(line)

		if lineIdx == e.replaceSelStartY {
			startCol = e.replaceSelStartX
		}
		if lineIdx == e.replaceSelEndY {
			endCol = e.replaceSelEndX
		}

		if startCol >= len(line) {
			e.addLog("Replace", fmt.Sprintf("Line %d: skipped (startCol >= len)", lineIdx))
			continue
		}

		prefix := lineStr[:startCol]
		searchPart := lineStr[startCol:endCol]
		suffix := ""
		if endCol < len(lineStr) {
			suffix = lineStr[endCol:]
		}

		e.addLog("Replace", fmt.Sprintf("Line %d: searching '%s' in range [%d:%d]", lineIdx, searchPart, startCol, endCol))

		var newSearchPart string
		if globalFlag {
			// Replace all occurrences in the slice.
			newSearchPart = re.ReplaceAllString(searchPart, replacement)
			matches := re.FindAllStringIndex(searchPart, -1)
			matchCount := len(matches)
			replacementCount += matchCount
			e.addLog("Replace", fmt.Sprintf("Line %d: found %d matches (global)", lineIdx, matchCount))
		} else {
			// Replace first match only.
			if re.MatchString(searchPart) {
				newSearchPart = re.ReplaceAllStringFunc(searchPart, func(match string) string {
					if replacementCount == 0 {
						replacementCount++
						return re.ReplaceAllString(match, replacement)
					}
					return match
				})
				e.addLog("Replace", fmt.Sprintf("Line %d: found 1 match (first only)", lineIdx))
			} else {
				newSearchPart = searchPart
				e.addLog("Replace", fmt.Sprintf("Line %d: no matches", lineIdx))
			}
		}

		// Update the line content and notify syntax highlighter of the edit.
		oldLine := b.buffer[lineIdx]
		newLineStr := prefix + newSearchPart + suffix
		b.buffer[lineIdx] = []rune(newLineStr)
		e.addLog("Replace", fmt.Sprintf("Line %d: '%s' -> '%s'", lineIdx, lineStr, newLineStr))

		if b.syntax != nil {
			oldLineBytes := uint32(len(string(oldLine)))
			newLineBytes := uint32(len(newLineStr))
			oldEndColBytes := b.getLineByteOffset(oldLine, endCol)
			newEndColBytes := b.getLineByteOffset(b.buffer[lineIdx], len(prefix)+len(newSearchPart))

			b.handleEdit(
				lineIdx, startCol,
				oldLineBytes, newLineBytes,
				lineIdx, oldEndColBytes,
				lineIdx, newEndColBytes,
			)
		}
	}

	if replacementCount > 0 {
		e.message = fmt.Sprintf("%d replacements made", replacementCount)
		e.markModified()
	} else {
		e.message = "Pattern not found"
	}

	// Force a full reparse of syntax to ensure all highlights are correct after mass edits.
	if b.syntax != nil {
		b.syntax.Reparse([]byte(b.toString()))
	}

	e.mode = ModeNormal
	e.replaceInput = []rune{}
	e.replaceMatches = []MatchRange{}
}
