package main

// Colon command handler (e.g., :q, :w, :help). It processes strings entered in
// ModeCommand and executes the corresponding actions.

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/nsf/termbox-go"
)

// Command provides a context for executing editor commands.
type Command struct {
	e *Editor
}

// IsValidCommand returns true if the command should be saved to history.
// Line numbers (pure integers) are not saved to history.
func (ch *Command) IsValidCommand(cmd string) bool {
	cmd = strings.TrimSpace(cmd)
	if cmd == "" {
		return false
	}

	// Check if it's a pure number (line jump command) - don't save to history
	if _, err := strconv.Atoi(cmd); err == nil {
		return false
	}

	// Valid if it's a known command
	switch cmd {
	case "q", "Q", "q!", "Q!", "w", "W", "wa", "WA", "wq", "WQ", "waq", "WAQ", "reload", "bd", "bd!", "debug", "help", "mouse", "e", "edit", "n":
		return true
	}

	// Valid if it starts with ! (shell command) or r! (read shell)
	if strings.HasPrefix(cmd, "!") || strings.HasPrefix(cmd, "r!") {
		return true
	}

	// Valid if it starts with w (write with filename)
	if strings.HasPrefix(cmd, "w ") {
		return true
	}

	// Everything else is considered invalid (will show "Command not found" message)
	return false
}

// HandleAndSaveToHistory executes a command and saves it to history if valid.
func (ch *Command) HandleAndSaveToHistory(cmd string) {
	ch.Handle(cmd)
	// Save to history only if it's a valid command (not a line number, not unrecognized)
	if ch.IsValidCommand(cmd) {
		if len(ch.e.commandHistory) == 0 || ch.e.commandHistory[len(ch.e.commandHistory)-1] != cmd {
			ch.e.commandHistory = append(ch.e.commandHistory, cmd)
		}
	}
}

// NavigateHistoryUp moves backward through command history.
func (ch *Command) NavigateHistoryUp() {
	if len(ch.e.commandHistory) == 0 {
		return
	}

	if ch.e.commandHistoryIdx == -1 {
		// Starting navigation from the end
		ch.e.commandHistoryIdx = len(ch.e.commandHistory) - 1
	} else if ch.e.commandHistoryIdx > 0 {
		ch.e.commandHistoryIdx--
	}

	if ch.e.commandHistoryIdx >= 0 && ch.e.commandHistoryIdx < len(ch.e.commandHistory) {
		ch.e.commandBuffer = []rune(ch.e.commandHistory[ch.e.commandHistoryIdx])
		ch.e.commandCursorX = len(ch.e.commandBuffer)
	}
}

// NavigateHistoryDown moves forward through command history.
func (ch *Command) NavigateHistoryDown() {
	if ch.e.commandHistoryIdx == -1 {
		return
	}

	ch.e.commandHistoryIdx++
	if ch.e.commandHistoryIdx >= len(ch.e.commandHistory) {
		// Reached the end, clear the buffer
		ch.e.commandHistoryIdx = -1
		ch.e.commandBuffer = []rune{}
		ch.e.commandCursorX = 0
	} else {
		ch.e.commandBuffer = []rune(ch.e.commandHistory[ch.e.commandHistoryIdx])
		ch.e.commandCursorX = len(ch.e.commandBuffer)
	}
}

// Handle parses and executes a command string.
func (ch *Command) Handle(cmd string) {
	cmd = strings.TrimSpace(cmd)
	switch {
	case cmd == "q" || cmd == "Q":
		ch.quit(false)
	case cmd == "q!" || cmd == "Q!":
		ch.quit(true)
	case cmd == "w" || cmd == "W":
		ch.write("")
	case strings.HasPrefix(cmd, "w "):
		filename := strings.TrimSpace(strings.TrimPrefix(cmd, "w "))
		ch.write(filename)
	case cmd == "wa" || cmd == "WA":
		ch.writeAll()
	case cmd == "wq" || cmd == "WQ":
		ch.writeQuit()
	case cmd == "waq" || cmd == "WAQ":
		ch.writeAll()
		// Check if any buffers are still modified (meaning save failed)
		hasModified := false
		for _, b := range ch.e.buffers {
			if b.modified && b.filename != "" && !b.readOnly {
				hasModified = true
				break
			}
		}
		// Only quit if all files were saved successfully
		if !hasModified {
			ch.quit(false)
		}
	case cmd == "reload":
		ch.reload()
	case cmd == "bd":
		ch.bufferDelete(false)
	case cmd == "bd!":
		ch.bufferDelete(true)
	case cmd == "n":
		ch.e.NewBuffer()
	case cmd == "debug":
		ch.e.toggleDebugWindow()
	case cmd == "help":
		// Load help content from the embedded filesystem.
		f, err := ContentFS.Open("content/help.txt")
		if err != nil {
			ch.e.message = fmt.Sprintf("Error opening help: %v", err)
		} else {
			defer f.Close()
			err = ch.e.LoadFromReader("help.txt", f)
			if err != nil {
				ch.e.message = fmt.Sprintf("Error loading help: %v", err)
			} else {
				// Help is read-only to prevent accidental edits.
				b := ch.e.activeBuffer()
				if b != nil {
					b.readOnly = true
					ch.e.message = "Help opened (Read-Only)"
				}
			}
		}
	case cmd == "mouse":
		ch.toggleMouse()
	case strings.HasPrefix(cmd, "e ") || strings.HasPrefix(cmd, "edit "):
		filename := ""
		if strings.HasPrefix(cmd, "e ") {
			filename = strings.TrimSpace(strings.TrimPrefix(cmd, "e "))
		} else {
			filename = strings.TrimSpace(strings.TrimPrefix(cmd, "edit "))
		}
		if filename != "" {
			err := ch.e.LoadFile(filename)
			if err != nil {
				ch.e.message = fmt.Sprintf("Error opening file: %v", err)
			} else {
				ch.e.message = fmt.Sprintf("Opened: %s", filename)
			}
		} else {
			ch.e.message = "No filename specified"
		}
	case cmd == "e" || cmd == "edit":
		ch.e.message = "No filename specified"
	default:
		if cmd == "" {
			break
		}
		// If the command starts with r!, execute it and insert output into buffer.
		if strings.HasPrefix(cmd, "r!") {
			shellCmd := strings.TrimPrefix(cmd, "r!")
			ch.readShell(shellCmd)
			break
		}
		// If the command starts with !, execute it as a shell command.
		if strings.HasPrefix(cmd, "!") {
			shellCmd := strings.TrimPrefix(cmd, "!")
			ch.executeShell(shellCmd)
			break
		}
		// If the command is a number, jump to that line.
		if lineNum, err := strconv.Atoi(cmd); err == nil {
			ch.goToLine(lineNum)
		} else {
			ch.e.message = fmt.Sprintf("Command not found: %s", cmd)
		}
	}
	// After executing a command, return to Normal mode and clear the command buffer.
	if ch.e.mode == ModeCommand {
		ch.e.mode = ModeNormal
	}
	ch.e.commandBuffer = []rune{}
}

// quit exits the editor, checking for unsaved changes unless 'force' is true.
func (ch *Command) quit(force bool) {
	if !force {
		// Check if any buffer has unsaved changes
		for _, b := range ch.e.buffers {
			if b.modified {
				ch.e.message = "No write since last change (use :q! to override)"
				return
			}
		}
	}
	termbox.Close()
	os.Exit(0)
}

// write saves the current active buffer to disk.
func (ch *Command) write(filename string) {
	if filename != "" {
		b := ch.e.activeBuffer()
		if b != nil {
			b.filename = filename
			b.fileType = getFileType(filename)
		}
	}
	err := ch.e.SaveFile(false)
	if err != nil {
		// Handle conflict if the file was changed externally.
		if err.Error() == "file changed on disk" {
			ch.e.message = "File changed on disk. Overwrite? (y/n) "
			ch.e.mode = ModeConfirm
			ch.e.pendingConfirm = func() {
				err := ch.e.SaveFile(true) // Force overwrite.
				if err != nil {
					ch.e.message = err.Error()
				} else {
					name := ch.e.activeBuffer().filename
					if name == "" {
						name = "[No Name]"
					}
					ch.e.message = fmt.Sprintf("\"%s\" written", name)
				}
			}
		} else {
			ch.e.message = err.Error()
		}
	} else {
		name := ch.e.activeBuffer().filename
		if name == "" {
			name = "[No Name]"
		}
		ch.e.message = fmt.Sprintf("\"%s\" written", name)
	}
}

// writeQuit saves the current buffer and exits.
func (ch *Command) writeQuit() {
	err := ch.e.SaveFile(false)
	if err != nil {
		if err.Error() == "file changed on disk" {
			ch.e.message = "File changed on disk. Overwrite? (y/n) "
			ch.e.mode = ModeConfirm
			ch.e.pendingConfirm = func() {
				err := ch.e.SaveFile(true)
				if err == nil {
					termbox.Close()
					os.Exit(0)
				} else {
					ch.e.message = err.Error()
				}
			}
		} else {
			ch.e.message = err.Error()
		}
	} else {
		termbox.Close()
		os.Exit(0)
	}
}

// writeAll saves all open buffers to disk.
func (ch *Command) writeAll() {
	savedCount := 0
	var lastErr error

	// Iterate through all buffers and save each one.
	for _, b := range ch.e.buffers {
		// Skip buffers without filenames (e.g., [No Name] buffers).
		if b.filename == "" {
			continue
		}

		// Skip read-only buffers.
		if b.readOnly {
			continue
		}

		// Save the buffer using the same logic as SaveFile but for each buffer.
		file, err := os.Create(b.filename)
		if err != nil {
			lastErr = err
			continue
		}

		writer := bufio.NewWriter(file)
		for i, line := range b.buffer {
			_, err := writer.WriteString(string(line))
			if err != nil {
				file.Close()
				lastErr = err
				continue
			}
			// Write newline if not the last line (or if buffer should end with newline).
			if i < len(b.buffer)-1 || (len(b.buffer) > 0 && (len(b.buffer) > 1 || len(b.buffer[0]) > 0)) {
				_, err = writer.WriteString("\n")
				if err != nil {
					file.Close()
					lastErr = err
					continue
				}
			}
		}

		err = writer.Flush()
		file.Close()

		if err == nil {
			b.modified = false
			info, err := os.Stat(b.filename)
			if err == nil {
				b.lastModTime = info.ModTime()
			}
			savedCount++
		} else {
			lastErr = err
		}
	}

	// Display appropriate message.
	if lastErr != nil {
		ch.e.message = fmt.Sprintf("Error saving some files: %v", lastErr)
	} else if savedCount == 0 {
		ch.e.message = "No files to save"
	} else if savedCount == 1 {
		ch.e.message = "1 file written"
	} else {
		ch.e.message = fmt.Sprintf("%d files written", savedCount)
	}
}

// bufferDelete closes the currently active buffer.
func (ch *Command) bufferDelete(force bool) {
	b := ch.e.activeBuffer()
	if !force && b != nil && b.modified {
		ch.e.message = "No write since last change (use :bd! to override)"
		return
	}
	ch.e.deleteCurrentBuffer()
}

// toggleMouse enables/disables mouse interaction in the terminal.
func (ch *Command) toggleMouse() {
	ch.e.mouseEnabled = !ch.e.mouseEnabled
	if ch.e.mouseEnabled {
		termbox.SetInputMode(termbox.InputEsc | termbox.InputMouse)
	} else {
		termbox.SetInputMode(termbox.InputEsc)
	}
}

// goToLine moves the cursor to the beginning of the specified line number.
func (ch *Command) goToLine(lineNum int) {
	b := ch.e.activeBuffer()
	if b != nil {
		targetY := lineNum - 1 // Convert 1-based UI line number to 0-based index.
		if targetY < 0 {
			targetY = 0
		}
		if targetY >= len(b.buffer) {
			targetY = len(b.buffer) - 1
		}
		b.PrimaryCursor().Y = targetY
		b.PrimaryCursor().X = 0
		ch.e.centerCursor()
	}
}

// reload re-reads the active buffer from disk.
func (ch *Command) reload() {
	b := ch.e.activeBuffer()
	if b != nil {
		err := ch.e.ReloadBuffer(b)
		if err != nil {
			ch.e.message = fmt.Sprintf("Reload failed: %v", err)
		} else {
			ch.e.message = fmt.Sprintf("\"%s\" reloaded", b.filename)
		}
	}
}

// executeShell runs a shell command and displays the output.
func (ch *Command) executeShell(shellCmd string) {
	shellCmd = strings.TrimSpace(shellCmd)
	if shellCmd == "" {
		ch.e.message = "No shell command specified"
		return
	}

	// Execute the command using sh -c for proper shell interpretation.
	cmd := exec.Command("/bin/sh", "-c", shellCmd)
	output, err := cmd.CombinedOutput()

	if err != nil {
		// Display error along with any output that was produced.
		if len(output) > 0 {
			ch.e.message = fmt.Sprintf("Error: %v | Output: %s", err, strings.TrimSpace(string(output)))
		} else {
			ch.e.message = fmt.Sprintf("Error executing command: %v", err)
		}
		return
	}

	// Display the command output, truncating if too long.
	outputStr := strings.TrimSpace(string(output))
	if outputStr == "" {
		ch.e.message = "Command executed successfully (no output)"
	} else {
		// Truncate output if it's too long for the message bar.
		const maxLen = 200
		if len(outputStr) > maxLen {
			ch.e.message = outputStr[:maxLen] + "..."
		} else {
			ch.e.message = outputStr
		}
	}
}

// readShell runs a shell command and inserts the output into the buffer at cursor position.
func (ch *Command) readShell(shellCmd string) {
	shellCmd = strings.TrimSpace(shellCmd)
	if shellCmd == "" {
		ch.e.message = "No shell command specified"
		return
	}

	b := ch.e.activeBuffer()
	if b == nil {
		return
	}

	if b.readOnly {
		ch.e.message = "File is read-only"
		return
	}

	// Execute the command using sh -c for proper shell interpretation.
	cmd := exec.Command("/bin/sh", "-c", shellCmd)
	output, err := cmd.CombinedOutput()

	if err != nil {
		// Display error along with any output that was produced.
		if len(output) > 0 {
			ch.e.message = fmt.Sprintf("Error: %v | Output: %s", err, strings.TrimSpace(string(output)))
		} else {
			ch.e.message = fmt.Sprintf("Error executing command: %v", err)
		}
		return
	}

	outputStr := string(output)
	if outputStr == "" {
		ch.e.message = "Command executed (no output to insert)"
		return
	}

	// Save state for undo.
	ch.e.saveState()

	// Split output into lines and insert them into the buffer.
	lines := strings.Split(outputStr, "\n")
	// Remove trailing empty line if present (common with command output).
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	if len(lines) == 0 {
		ch.e.message = "Command executed (no output to insert)"
		return
	}

	c := b.PrimaryCursor()
	currentY := c.Y

	// Insert output starting from the line after the cursor.
	for i, line := range lines {
		insertY := currentY + i + 1
		// Create new line in buffer.
		newLine := []rune(line)
		// Insert the line into the buffer.
		if insertY <= len(b.buffer) {
			b.buffer = append(b.buffer[:insertY], append([][]rune{newLine}, b.buffer[insertY:]...)...)
		} else {
			b.buffer = append(b.buffer, newLine)
		}
	}

	// Mark buffer as modified.
	ch.e.markModified()

	// Reparse syntax if needed.
	if b.syntax != nil {
		b.syntax.Reparse([]byte(b.toString()))
	}

	// Notify LSP of the change.
	if b.lspClient != nil {
		b.lspClient.SendDidChange(b.toString())
	}

	lineCount := len(lines)
	if lineCount == 1 {
		ch.e.message = "1 line inserted"
	} else {
		ch.e.message = fmt.Sprintf("%d lines inserted", lineCount)
	}
}
