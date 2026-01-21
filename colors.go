package main

// Utility to print all 256 terminal colors. This is useful for debugging themes
// and ensuring the terminal supports the expected color range.

import (
	"fmt"
	"os"

	"github.com/nsf/termbox-go"
)

// PrintColors initializes termbox and draws a grid of all 256 available colors.
func PrintColors() {
	err := termbox.Init()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to init termbox: %v\n", err)
		return
	}
	defer termbox.Close()

	// Enable 256-color mode for the output.
	termbox.SetOutputMode(termbox.Output256)
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)

	w, _ := termbox.Size()

	// Adjust grid columns based on terminal width.
	cols := 16
	if w < 64 {
		cols = 8
	}

	// Loop through all 256 colors and draw them in a grid.
	for i := 0; i < 256; i++ {
		row := (i / cols) * 2
		col := (i % cols) * 5

		bg := termbox.Attribute(i)
		fg := termbox.ColorWhite
		// Ensure text is readable against light/dark backgrounds.
		if i == 7 || i > 240 {
			fg = termbox.ColorBlack
		}

		// Draw the color index and a colored block.
		str := fmt.Sprintf("%5d", i)
		for j, r := range str {
			termbox.SetCell(col+j, row, r, fg, bg)
			termbox.SetCell(col+j, row+1, ' ', fg, bg)
		}
	}

	msg := "Press any key to exit..."
	for i, r := range msg {
		termbox.SetCell(i, (256/cols)*2, r, termbox.ColorWhite, termbox.ColorDefault)
	}

	termbox.Flush()
	// Wait for any key press before closing.
	termbox.PollEvent()
}
