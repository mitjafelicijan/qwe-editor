package main

// The entry point of the qwe editor. It handles command-line flags, initializes
// configuration, file types, terminal interface (termbox), and starts the main
// editor loop.

import (
	"flag"
	"fmt"
	"os"

	"github.com/nsf/termbox-go"
)

// Version of the editor, injected at build time.
var Version = "dev"

func main() {
	// Initialize configuration from flags and environment.
	InitConfig()

	// If -version flag is provided, print version and exit.
	if Config.ShowVersion {
		fmt.Println(Version)
		return
	}

	// Load supported file types for syntax highlighting.
	InitFileTypes()

	// Print available colors if -colors flag is provided.
	if Config.ShowColors {
		PrintColors()
		return
	}

	// Print system information if -info flag is provided.
	if Config.ShowInfo {
		PrintInfo()
		return
	}

	// Initialize termbox for TUI handling.
	err := termbox.Init()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to init termbox: %v\n", err)
		os.Exit(1)
	}
	defer termbox.Close()

	// Enable mouse support and escape key handling.
	termbox.SetInputMode(termbox.InputEsc | termbox.InputMouse)
	// Use 256 color mode for better aesthetics.
	termbox.SetOutputMode(termbox.Output256)

	// Create a new editor instance.
	editor := NewEditor(Config.DevMode)
	// Start background checks for Ollama AI and file changes on disk.
	editor.ollamaClient.PeriodicStatusCheck()
	editor.PeriodicFileChangesCheck()

	// Check if filenames were provided as arguments and load them into buffers.
	if flag.NArg() > 0 {
		for _, filename := range flag.Args() {
			if err := editor.LoadFile(filename); err != nil {
				termbox.Close()
				fmt.Fprintf(os.Stderr, "failed to open file %s: %v\n", filename, err)
				os.Exit(1)
			}
		}
		// Start with the first file active
		editor.activeBufferIndex = 0
	}

	// Enter the main event loop (keyboard and mouse input).
	editor.HandleEvents()
}
