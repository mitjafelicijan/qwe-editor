package main

// Input processing engine. It contains the main event loop and dispatches
// keyboard/mouse events to mode-specific handlers (Normal, Insert, Visual,
// etc.).

import (
	"github.com/nsf/termbox-go"
)

// HandleEvents is the central loop that waits for and processes all user input.
func (e *Editor) HandleEvents() {
	for {
		// Redraw the screen before waiting for the next event.
		e.draw()
		ev := termbox.PollEvent()

		// Handle interrupt events (triggered by diagnostic updates).
		// Fetch latest diagnostics from LSP client.
		if ev.Type == termbox.EventInterrupt {
			b := e.activeBuffer()
			if b != nil && b.lspClient != nil {
				b.diagnostics = b.lspClient.GetDiagnostics()
			}
			e.CheckFilesOnDisk()
			continue
		}

		if ev.Type == termbox.EventKey {
			// Clear message on any key press unless specifically set.
			e.message = ""
			// Hide hover popup if any key other than Ctrl+K is pressed.
			if e.showHover && ev.Key != termbox.KeyCtrlK {
				e.showHover = false
			}

			// If dev mode, exit the editor with Ctrl+C.
			if ev.Key == termbox.KeyCtrlC && e.devMode {
				return
			}

			// Dispatch the key event to the handler for the current editor mode.
			switch e.mode {
			case ModeNormal:
				e.handleNormalMode(ev)
			case ModeInsert:
				e.handleInsertMode(ev)
			case ModeCommand:
				e.handleCommandMode(ev)
			case ModeFuzzy:
				e.handleFuzzyMode(ev)
			case ModeFind:
				e.handleFindMode(ev)
			case ModeVisual:
				e.handleVisualMode(ev)
			case ModeVisualLine:
				e.handleVisualLineMode(ev)
			case ModeVisualBlock:
				e.handleVisualBlockMode(ev)
			case ModeReplace:
				e.handleReplaceMode(ev)
			case ModeConfirm:
				e.handleConfirmMode(ev)
			}
		} else if ev.Type == termbox.EventMouse {
			e.handleMouseEvent(ev)
		}
	}
}

// handleNormalMode processes keyboard input when the editor is in Normal mode.
func (e *Editor) handleNormalMode(ev termbox.Event) {
	// Escape clears any pending multi-key commands or secondary cursors.
	if ev.Key == termbox.KeyEsc {
		b := e.activeBuffer()
		if b != nil && len(b.cursors) > 1 {
			e.clearSecondaryCursors()
			e.pendingKey = 0
			e.message = "Cleared secondary cursors"
			return
		}
		e.pendingKey = 0
		return
	}

	switch ev.Key {
	case termbox.KeyArrowLeft:
		e.moveCursor(-1, 0)
	case termbox.KeyArrowRight:
		e.moveCursor(1, 0)
	case termbox.KeyArrowUp:
		if ev.Mod != 0 {
			e.addCursorAbove()
		} else {
			e.moveCursor(0, -1)
		}
	case termbox.KeyArrowDown:
		if ev.Mod != 0 {
			e.addCursorBelow()
		} else {
			e.moveCursor(0, 1)
		}
	case termbox.KeyCtrlX:
		e.addCursorBelow()
	case termbox.KeyCtrlP:
		e.prevBuffer()
	case termbox.KeyCtrlN:
		e.nextBuffer()
	case termbox.KeyCtrlO:
		e.jumpBack()
	case termbox.KeyCtrlI:
		e.jumpForward()
	case termbox.KeyCtrlV:
		b := e.activeBuffer()
		if b != nil {
			e.visualStartX = b.PrimaryCursor().X
			e.visualStartY = b.PrimaryCursor().Y
		}
		e.mode = ModeVisualBlock
	case termbox.KeyCtrlK:
		e.triggerHover()
	}

	// Prevent key event fallthrough.
	if ev.Key != 0 {
		return
	}

	switch ev.Ch {
	case 'i':
		e.saveState()
		e.mode = ModeInsert
		e.introDismissed = true
	case 'a':
		e.saveState()
		e.moveCursor(1, 0)
		e.mode = ModeInsert
		e.introDismissed = true
	case 'A':
		e.saveState()
		e.jumpToLineEnd()
		e.mode = ModeInsert
		e.introDismissed = true
	case 'I':
		e.saveState()
		e.jumpToFirstNonBlank()
		e.mode = ModeInsert
		e.introDismissed = true
	case 'o':
		e.saveState()
		e.insertLineBelow()
		e.mode = ModeInsert
		e.introDismissed = true
	case 'O':
		e.saveState()
		e.insertLineAbove()
		e.mode = ModeInsert
		e.introDismissed = true
	case ']':
		e.pushJump()
		e.jumpToNextEmptyLine()
	case '}':
		e.pushJump()
		e.jumpToBottom()
	case 'v':
		b := e.activeBuffer()
		if b != nil {
			e.visualStartX = b.PrimaryCursor().X
			e.visualStartY = b.PrimaryCursor().Y
		}
		e.mode = ModeVisual
	case 'V':
		b := e.activeBuffer()
		if b != nil {
			e.visualStartX = b.PrimaryCursor().X
			e.visualStartY = b.PrimaryCursor().Y
		}
		e.mode = ModeVisualLine
	case ':':
		e.mode = ModeCommand
		e.commandBuffer = []rune{}
		e.commandCursorX = 0
	case '/':
		e.findSavedSearch = e.lastSearch
		e.mode = ModeFind
		e.findBuffer = []rune{}
	case Config.LeaderKey:
		e.pendingKey = Config.LeaderKey
	case 'l':
		if e.pendingKey == Config.LeaderKey {
			e.toggleDebugWindow()
			e.pendingKey = 0
		}
	case 'w':
		if e.pendingKey == 'd' {
			e.saveState()
			e.deleteWord(true)
			e.checkDiagnostics()
			e.pendingKey = 0
		} else if e.pendingKey == 'c' {
			e.saveState()
			e.changeWord()
			e.checkDiagnostics()
			e.pendingKey = 0
		} else if e.pendingKey == Config.LeaderKey {
			e.startWarningsFuzzyFinder()
			e.pendingKey = 0
		} else {
			e.moveWordForward()
		}
	case 'q':
		if e.pendingKey == 'z' {
			e.formatText()
			e.checkDiagnostics()
			e.pendingKey = 0
		} else if e.pendingKey == Config.LeaderKey {
			e.lastSearch = ""
			e.pendingKey = 0
		} else {
			e.moveWordBackward()
		}
	case 'Q':
		e.jumpToFirstNonBlank()
	case 'W':
		e.jumpToLineEnd()
	case 'g':
		e.pendingKey = 'g'
	case 'j':
		e.saveState()
		e.JoinLines()
		e.checkDiagnostics()
	case 'f':
		if e.pendingKey == 'g' {
			e.gotoFile()
			e.pendingKey = 0
		}
	case 'd':
		if e.pendingKey == Config.LeaderKey {
			e.deleteCurrentBuffer()
			e.pendingKey = 0
		} else if e.pendingKey == 'd' {
			e.saveState()
			e.deleteLine()
			e.checkDiagnostics()
			e.pendingKey = 0
		} else if e.pendingKey == 'g' {
			e.gotoDefinition()
			e.pendingKey = 0
		} else {
			e.pendingKey = 'd'
		}
	case 'y':
		e.yankLine()
		e.message = "Line yanked"
	case 'x':
		if e.pendingKey == 'z' {
			e.saveState()
			e.toggleCommentLine()
			e.checkDiagnostics()
			e.pendingKey = 0
		} else {
			e.saveState()
			e.DeleteChar()
			e.checkDiagnostics()
			e.pendingKey = 0
		}
	case 'z':
		if e.pendingKey == 'z' {
			e.centerScreen()
			e.pendingKey = 0
		} else {
			e.pendingKey = 'z'
		}
	case 'c':
		if e.pendingKey == 'd' {
			e.saveState()
			e.DeleteChar()
			e.checkDiagnostics()
			e.pendingKey = 0
		} else if e.pendingKey == 'c' {
			e.saveState()
			e.changeCharacter()
			e.checkDiagnostics()
			e.pendingKey = 0
		} else {
			e.pendingKey = 'c'
		}
	case 'C':
		e.saveState()
		e.changeToEndOfLine()
		e.checkDiagnostics()
		e.pendingKey = 0
	case 'D':
		e.saveState()
		e.deleteToEndOfLine()
		e.checkDiagnostics()
		e.pendingKey = 0
	case '(':
		if e.pendingKey == 'c' {
			e.saveState()
			e.changeInside('(', ')')
			e.checkDiagnostics()
			e.pendingKey = 0
		} else if e.pendingKey == 'd' {
			e.saveState()
			e.deleteInside('(', ')')
			e.checkDiagnostics()
			e.pendingKey = 0
		}
	case '[':
		if e.pendingKey == 'c' {
			e.saveState()
			e.changeInside('[', ']')
			e.checkDiagnostics()
			e.pendingKey = 0
		} else if e.pendingKey == 'd' {
			e.saveState()
			e.deleteInside('[', ']')
			e.checkDiagnostics()
			e.pendingKey = 0
		} else {
			e.pushJump()
			e.jumpToPrevEmptyLine()
		}
	case '{':
		if e.pendingKey == 'c' {
			e.saveState()
			e.changeInside('{', '}')
			e.checkDiagnostics()
			e.pendingKey = 0
		} else if e.pendingKey == 'd' {
			e.saveState()
			e.deleteInside('{', '}')
			e.checkDiagnostics()
			e.pendingKey = 0
		} else {
			e.pushJump()
			e.jumpToTop()
		}
	case '\'':
		if e.pendingKey == 'c' {
			e.saveState()
			e.changeInside('\'', '\'')
			e.checkDiagnostics()
			e.pendingKey = 0
		} else if e.pendingKey == 'd' {
			e.saveState()
			e.deleteInside('\'', '\'')
			e.checkDiagnostics()
			e.pendingKey = 0
		}
	case '"':
		if e.pendingKey == 'c' {
			e.saveState()
			e.changeInside('"', '"')
			e.checkDiagnostics()
			e.pendingKey = 0
		} else if e.pendingKey == 'd' {
			e.saveState()
			e.deleteInside('"', '"')
			e.checkDiagnostics()
			e.pendingKey = 0
		}
	case 's':
		e.saveState()
		e.changeCharacter()
		e.checkDiagnostics()
		e.pendingKey = 0
	case 'n':
		e.findNext()
		e.centerCursor()
	case 'N':
		e.findPrev()
		e.centerCursor()
	case 'u':
		e.undo()
		e.checkDiagnostics()
		e.pendingKey = 0
	case 'U':
		e.redo()
		e.checkDiagnostics()
		e.pendingKey = 0
	case 'p':
		if e.pendingKey == Config.LeaderKey {
			e.startFileFuzzyFinder()
			e.pendingKey = 0
		} else {
			e.saveState()
			e.pasteLine()
			e.checkDiagnostics()
			e.pendingKey = 0
		}
	case 'b':
		if e.pendingKey == Config.LeaderKey {
			e.startBufferFuzzyFinder()
			e.pendingKey = 0
		}
	case 'P':
		if e.pendingKey == Config.LeaderKey {
			e.pendingKey = 0
		} else {
			e.saveState()
			e.pasteLineAbove()
			e.checkDiagnostics()
			e.pendingKey = 0
		}
	default:
		e.pendingKey = 0
	}
}

// handleInsertMode processes keyboard input when the editor is in Insert mode.
func (e *Editor) handleInsertMode(ev termbox.Event) {
	if e.showAutocomplete {
		switch ev.Key {
		case termbox.KeyArrowUp:
			e.autocompleteIndex--
			if e.autocompleteIndex < 0 {
				e.autocompleteIndex = len(e.autocompleteItems) - 1
			}
			// Adjust scroll to keep selection visible
			if e.autocompleteIndex < e.autocompleteScroll {
				e.autocompleteScroll = e.autocompleteIndex
			}
			if e.autocompleteIndex >= e.autocompleteScroll+10 {
				e.autocompleteScroll = e.autocompleteIndex - 9
			}
			return
		case termbox.KeyArrowDown:
			e.autocompleteIndex++
			if e.autocompleteIndex >= len(e.autocompleteItems) {
				e.autocompleteIndex = 0
			}
			// Adjust scroll to keep selection visible
			if e.autocompleteIndex < e.autocompleteScroll {
				e.autocompleteScroll = e.autocompleteIndex
			}
			if e.autocompleteIndex >= e.autocompleteScroll+10 {
				e.autocompleteScroll = e.autocompleteIndex - 9
			}
			return
		case termbox.KeyEnter:
			e.insertCompletion(e.autocompleteItems[e.autocompleteIndex])
			return
		case termbox.KeyEsc:
			e.showAutocomplete = false
			return
		}
	}

	switch ev.Key {
	case termbox.KeyEsc:
		// Return to Normal mode and trigger a diagnostic check.
		e.mode = ModeNormal
		e.checkDiagnostics()
	case termbox.KeyEnter:
		e.insertNewline()
	case termbox.KeySpace:
		e.insertRune(' ')
	case termbox.KeyBackspace, termbox.KeyBackspace2:
		e.backspace()
		if e.showAutocomplete {
			e.showAutocomplete = false
		}
	case termbox.KeyTab:
		e.insertTab()
		if e.showAutocomplete {
			e.showAutocomplete = false
		}
	case termbox.KeyArrowLeft:
		e.moveCursor(-1, 0)
		if e.showAutocomplete {
			e.showAutocomplete = false
		}
	case termbox.KeyArrowRight:
		e.moveCursor(1, 0)
		if e.showAutocomplete {
			e.showAutocomplete = false
		}
	case termbox.KeyArrowUp:
		e.moveCursor(0, -1)
		if e.showAutocomplete {
			e.showAutocomplete = false
		}
	case termbox.KeyArrowDown:
		e.moveCursor(0, 1)
		if e.showAutocomplete {
			e.showAutocomplete = false
		}
	case termbox.KeyCtrlW:
		e.deleteWordBackward()
	case termbox.KeyCtrlN:
		e.triggerAutocomplete()
	default:
		// If a character key was pressed, insert the character.
		if ev.Ch != 0 {
			e.insertRune(ev.Ch)
			// Close autocomplete if user keeps typing.
			if e.showAutocomplete {
				e.showAutocomplete = false
			}
		}
	}
}

// handleCommandMode processes keyboard input for the colon command line.
func (e *Editor) handleCommandMode(ev termbox.Event) {
	switch ev.Key {
	case termbox.KeyEsc:
		// Cancel command entry.
		e.mode = ModeNormal
		e.commandBuffer = []rune{}
		e.commandCursorX = 0
		e.commandHistoryIdx = -1
		e.checkDiagnostics()
	case termbox.KeyEnter:
		// Execute the entered command and save to history if valid.
		cmd := string(e.commandBuffer)
		e.commands.HandleAndSaveToHistory(cmd)
		e.commandHistoryIdx = -1
	case termbox.KeyBackspace, termbox.KeyBackspace2:
		if e.commandCursorX > 0 {
			// Delete character before cursor
			e.commandBuffer = append(e.commandBuffer[:e.commandCursorX-1], e.commandBuffer[e.commandCursorX:]...)
			e.commandCursorX--
		} else if len(e.commandBuffer) == 0 {
			// If buffer is empty, backspace returns to Normal mode.
			e.mode = ModeNormal
		}
		e.commandHistoryIdx = -1
	case termbox.KeySpace:
		// Insert space at cursor position
		e.commandBuffer = append(e.commandBuffer[:e.commandCursorX], append([]rune{' '}, e.commandBuffer[e.commandCursorX:]...)...)
		e.commandCursorX++
		e.commandHistoryIdx = -1
	case termbox.KeyCtrlW:
		e.deleteWordBackwardFromBuffer()
		e.commandHistoryIdx = -1
	case termbox.KeyArrowLeft:
		// Move cursor left
		if e.commandCursorX > 0 {
			e.commandCursorX--
		}
	case termbox.KeyArrowRight:
		// Move cursor right
		if e.commandCursorX < len(e.commandBuffer) {
			e.commandCursorX++
		}
	case termbox.KeyArrowUp:
		// Navigate to previous command in history
		e.commands.NavigateHistoryUp()
	case termbox.KeyArrowDown:
		// Navigate to next command in history
		e.commands.NavigateHistoryDown()
	default:
		if ev.Ch != 0 {
			// Insert character at cursor position
			e.commandBuffer = append(e.commandBuffer[:e.commandCursorX], append([]rune{ev.Ch}, e.commandBuffer[e.commandCursorX:]...)...)
			e.commandCursorX++
			e.commandHistoryIdx = -1
		}
	}
}

// handleFuzzyMode processes input for the fuzzy finder (files or buffers).
func (e *Editor) handleFuzzyMode(ev termbox.Event) {
	switch ev.Key {
	case termbox.KeyEsc:
		e.mode = ModeNormal
	case termbox.KeyEnter:
		// Open the currently selected item in the list.
		e.openSelectedFile()
	case termbox.KeyArrowUp:
		e.fuzzyMove(1)
	case termbox.KeyArrowDown:
		e.fuzzyMove(-1)
	case termbox.KeyBackspace, termbox.KeyBackspace2:
		if len(e.fuzzyBuffer) > 0 {
			e.fuzzyBuffer = e.fuzzyBuffer[:len(e.fuzzyBuffer)-1]
			e.updateFuzzyResults()
		}
	case termbox.KeySpace:
		e.fuzzyBuffer = append(e.fuzzyBuffer, ' ')
		e.updateFuzzyResults()
	default:
		// Update filter as user types.
		if ev.Ch != 0 {
			e.fuzzyBuffer = append(e.fuzzyBuffer, ev.Ch)
			e.updateFuzzyResults()
		}
	}
}

// handleFindMode processes input for the in-file search (/).
func (e *Editor) handleFindMode(ev termbox.Event) {
	switch ev.Key {
	case termbox.KeyEsc:
		e.mode = ModeNormal
		e.findBuffer = []rune{}
		// Revert to the last successful search term.
		e.lastSearch = e.findSavedSearch
		e.checkDiagnostics()
	case termbox.KeyEnter:
		if len(e.findBuffer) > 0 {
			e.lastSearch = string(e.findBuffer)
			e.findNext()
			e.centerCursor()
		}
		e.mode = ModeNormal
	case termbox.KeyBackspace, termbox.KeyBackspace2:
		if len(e.findBuffer) > 0 {
			e.findBuffer = e.findBuffer[:len(e.findBuffer)-1]
			e.lastSearch = string(e.findBuffer)
		} else {
			e.lastSearch = e.findSavedSearch
		}
	case termbox.KeySpace:
		e.findBuffer = append(e.findBuffer, ' ')
		e.lastSearch = string(e.findBuffer)
	default:
		// Incremental search: update e.lastSearch as the user types.
		if ev.Ch != 0 {
			e.findBuffer = append(e.findBuffer, ev.Ch)
			e.lastSearch = string(e.findBuffer)
		}
	}
}

// handleVisualMode processes input for character-wise visual selection.
func (e *Editor) handleVisualMode(ev termbox.Event) {
	if ev.Key == termbox.KeyEsc {
		// Exit visual mode and return to Normal.
		e.mode = ModeNormal
		return
	}

	switch ev.Key {
	case termbox.KeyArrowLeft:
		e.moveCursor(-1, 0)
	case termbox.KeyArrowRight:
		e.moveCursor(1, 0)
	case termbox.KeyArrowUp:
		e.moveCursor(0, -1)
	case termbox.KeyArrowDown:
		e.moveCursor(0, 1)
	}

	// Prevent key event fallthrough.
	if ev.Key != 0 {
		return
	}

	switch ev.Ch {
	case Config.LeaderKey:
		e.pendingKey = Config.LeaderKey
	case 'w':
		e.moveWordForward()
	case 'q':
		if e.pendingKey == 'z' {
			e.formatText()
			e.checkDiagnostics()
			e.pendingKey = 0
		} else {
			e.moveWordBackward()
		}
	case 'y':
		e.yankVisualSelection()
		e.message = "Selection yanked"
	case 'd':
		e.saveState()
		e.deleteVisualSelection()
		e.checkDiagnostics()
		e.message = "Selection deleted"
	case 'x':
		if e.pendingKey == 'z' {
			e.saveState()
			e.commentVisualSelection()
			e.checkDiagnostics()
			e.pendingKey = 0
		} else {
			e.saveState()
			e.deleteVisualSelection()
			e.checkDiagnostics()
			e.message = "Selection deleted"
		}
	case 'p':
		e.saveState()
		e.pasteVisualSelection()
		e.checkDiagnostics()
	case 'c':
		e.saveState()
		e.changeVisualSelection()
		e.checkDiagnostics()
	case 'Q':
		e.jumpToFirstNonBlank()
	case 'W':
		e.jumpToLineEnd()
	case '~':
		e.saveState()
		e.ToggleCaseVisualSelection()
		e.checkDiagnostics()
	case 'o':
		if e.pendingKey == Config.LeaderKey {
			e.ollamaComplete()
			e.pendingKey = 0
		} else {
			// Swap cursor and visual anchor
			b := e.activeBuffer()
			if b != nil {
				tmpX, tmpY := b.PrimaryCursor().X, b.PrimaryCursor().Y
				b.PrimaryCursor().X, b.PrimaryCursor().Y = e.visualStartX, e.visualStartY
				e.visualStartX, e.visualStartY = tmpX, tmpY
			}
		}
	case '{':
		e.jumpToTop()
	case '}':
		e.jumpToBottom()
	case '[':
		e.jumpToPrevEmptyLine()
	case ']':
		e.jumpToNextEmptyLine()
	case ':':
		e.mode = ModeCommand
		e.commandBuffer = []rune{}
		e.commandCursorX = 0
	case 'V':
		e.mode = ModeVisualLine
	case 'z':
		e.pendingKey = 'z'
	case 'R':
		e.startReplaceMode()
	}
}

func (e *Editor) handleVisualLineMode(ev termbox.Event) {
	if ev.Key == termbox.KeyEsc {
		e.mode = ModeNormal
		return
	}

	switch ev.Key {
	case termbox.KeyArrowLeft:
		e.moveCursor(-1, 0)
	case termbox.KeyArrowRight:
		e.moveCursor(1, 0)
	case termbox.KeyArrowUp:
		e.moveCursor(0, -1)
	case termbox.KeyArrowDown:
		e.moveCursor(0, 1)
	}

	// Prevent key event fallthrough.
	if ev.Key != 0 {
		return
	}

	switch ev.Ch {
	case Config.LeaderKey:
		e.pendingKey = Config.LeaderKey
	case 'w':
		e.moveWordForward()
	case 'q':
		if e.pendingKey == 'z' {
			e.formatText()
			e.checkDiagnostics()
			e.pendingKey = 0
		} else {
			e.moveWordBackward()
		}
	case 'y':
		e.yankVisualSelection()
		e.message = "Selection yanked"
	case 'd':
		e.saveState()
		e.deleteVisualSelection()
		e.checkDiagnostics()
		e.message = "Selection deleted"
	case 'x':
		if e.pendingKey == 'z' {
			e.saveState()
			e.commentVisualSelection()
			e.checkDiagnostics()
			e.pendingKey = 0
		} else {
			e.saveState()
			e.deleteVisualSelection()
			e.checkDiagnostics()
			e.message = "Selection deleted"
		}
	case 'p':
		e.saveState()
		e.pasteVisualSelection()
		e.checkDiagnostics()
	case 'c':
		e.saveState()
		e.changeVisualSelection()
		e.checkDiagnostics()
	case 'Q':
		e.jumpToFirstNonBlank()
	case 'W':
		e.jumpToLineEnd()
	case '~':
		e.saveState()
		e.ToggleCaseVisualSelection()
		e.checkDiagnostics()
	case 'o':
		if e.pendingKey == Config.LeaderKey {
			e.ollamaComplete()
			e.pendingKey = 0
		} else {
			// Swap cursor and visual anchor
			b := e.activeBuffer()
			if b != nil {
				tmpX, tmpY := b.PrimaryCursor().X, b.PrimaryCursor().Y
				b.PrimaryCursor().X, b.PrimaryCursor().Y = e.visualStartX, e.visualStartY
				e.visualStartX, e.visualStartY = tmpX, tmpY
			}
		}
	case '{':
		e.jumpToTop()
	case '}':
		e.jumpToBottom()
	case '[':
		e.jumpToPrevEmptyLine()
	case ']':
		e.jumpToNextEmptyLine()
	case 'z':
		e.pendingKey = 'z'
	case 'v':
		e.mode = ModeVisual
	case 'V':
		e.mode = ModeNormal
	case 'R':
		e.startReplaceMode()
	}
}

// handleVisualBlockMode processes input for column-wise (rectangular) selection.
func (e *Editor) handleVisualBlockMode(ev termbox.Event) {
	if ev.Key == termbox.KeyEsc {
		e.mode = ModeNormal
		return
	}

	switch ev.Key {
	case termbox.KeyArrowLeft:
		e.moveCursor(-1, 0)
	case termbox.KeyArrowRight:
		e.moveCursor(1, 0)
	case termbox.KeyArrowUp:
		e.moveCursor(0, -1)
	case termbox.KeyArrowDown:
		e.moveCursor(0, 1)
	}

	// Prevent key event fallthrough.
	if ev.Key != 0 {
		return
	}

	switch ev.Ch {
	case Config.LeaderKey:
		e.pendingKey = Config.LeaderKey
	case 'w':
		e.moveWordForward()
	case 'q':
		if e.pendingKey == 'z' {
			e.formatText()
			e.checkDiagnostics()
			e.pendingKey = 0
		} else {
			e.moveWordBackward()
		}
	case 'y':
		e.yankVisualSelection()
		e.message = "Selection yanked"
	case 'd':
		e.saveState()
		e.deleteVisualSelection()
		e.checkDiagnostics()
		e.message = "Selection deleted"
	case 'x':
		if e.pendingKey == 'z' {
			e.saveState()
			e.commentVisualSelection()
			e.checkDiagnostics()
			e.pendingKey = 0
		} else {
			e.saveState()
			e.deleteVisualSelection()
			e.checkDiagnostics()
			e.message = "Selection deleted"
		}
	case 'p':
		e.saveState()
		e.pasteVisualSelection()
		e.checkDiagnostics()
	case 'c':
		e.saveState()
		e.changeVisualSelection()
		e.checkDiagnostics()
	case 'Q':
		e.jumpToFirstNonBlank()
	case 'W':
		e.jumpToLineEnd()
	case '~':
		e.saveState()
		e.ToggleCaseVisualSelection()
		e.checkDiagnostics()
	case 'o':
		if e.pendingKey == Config.LeaderKey {
			e.ollamaComplete()
			e.pendingKey = 0
		} else {
			// Swap cursor and visual anchor
			b := e.activeBuffer()
			if b != nil {
				tmpX, tmpY := b.PrimaryCursor().X, b.PrimaryCursor().Y
				b.PrimaryCursor().X, b.PrimaryCursor().Y = e.visualStartX, e.visualStartY
				e.visualStartX, e.visualStartY = tmpX, tmpY
			}
		}
	case '{':
		e.jumpToTop()
	case '}':
		e.jumpToBottom()
	case '[':
		e.jumpToPrevEmptyLine()
	case ']':
		e.jumpToNextEmptyLine()
	case 'z':
		e.pendingKey = 'z'
	case 'v':
		e.mode = ModeVisual
	case 'V':
		e.mode = ModeVisualLine
	case 'R':
		e.startReplaceMode()
	}
}

// handleMouseEvent handles simple mouse wheel scrolling.
func (e *Editor) handleMouseEvent(ev termbox.Event) {
	switch ev.Key {
	case termbox.MouseWheelUp:
		e.moveCursor(0, -1) // Scroll up by moving the cursor.
	case termbox.MouseWheelDown:
		e.moveCursor(0, 1) // Scroll down by moving the cursor.
	}
}

// handleConfirmMode processes yes/no confirmations for dangerous actions (like overwriting files).
func (e *Editor) handleConfirmMode(ev termbox.Event) {
	if ev.Key == termbox.KeyEsc {
		e.mode = ModeNormal
		e.pendingConfirm = nil
		e.message = "Cancelled"
		return
	}

	if ev.Key == termbox.KeyEnter {
		// Default Enter to "no/cancel" to avoid accidental execution.
		e.mode = ModeNormal
		e.pendingConfirm = nil
		e.message = "Cancelled"
		return
	}

	// Prevent key event fallthrough.
	if ev.Key != 0 {
		return
	}

	switch ev.Ch {
	case 'y', 'Y':
		if e.pendingConfirm != nil {
			action := e.pendingConfirm
			e.pendingConfirm = nil
			e.mode = ModeNormal
			action()
		} else {
			e.mode = ModeNormal
		}
	case 'n', 'N':
		e.mode = ModeNormal
		e.pendingConfirm = nil
		e.message = "Cancelled"
	}
}
