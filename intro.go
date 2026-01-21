package main

// Handles drawing the splash screen (introduction) that appears when the editor
// starts with no files.

import (
	"github.com/nsf/termbox-go"
)

// drawIntro clears the screen and draws an informational box with version and basic commands.
func (e *Editor) drawIntro() {
	w, h := termbox.Size()

	// Define specific attributes for the intro screen elements.
	const (
		cTitle   = termbox.Attribute(254) | termbox.AttrBold
		cText    = termbox.Attribute(248)
		cVersion = termbox.Attribute(239)
		cSubtext = termbox.Attribute(248)
		cKey     = termbox.Attribute(254)
		cLink    = termbox.Attribute(248) | termbox.AttrUnderline
	)

	// List of lines to display in the intro box.
	lines := []struct {
		text string
		fg   termbox.Attribute
	}{
		{"qwe editor", cTitle},
		{Version, cVersion},
		{"", cText},
		{"", cText},
		{"By Mitja Felicijan et al.", cText},
		{"Small, opinionated modal text editor", cSubtext},
		{"", cText},
		{" type  :q<Enter>             to exit", cKey},
		{" type  :help<Enter>         for help", cKey},
		{"", cText},
		{"https://github.com/mitjafelicijan/qwe-editor", cLink},
	}

	// Calculate the maximum line length to center the box.
	maxLen := 0
	for _, line := range lines {
		if len(line.text) > maxLen {
			maxLen = len(line.text)
		}
	}

	boxWidth := maxLen + 2
	boxHeight := len(lines)

	// Center point for the box.
	startX := (w - boxWidth) / 2
	startY := (h - boxHeight) / 2

	_, bg := GetThemeColor(ColorDefault)
	for i, line := range lines {
		// Center each line individually within the box.
		lineX := startX + (maxLen-len(line.text))/2
		lineY := startY + i
		for j, char := range line.text {
			termbox.SetCell(lineX+j, lineY, char, line.fg, bg)
		}
	}
}
