package main

// Global configuration of the editor. Settings are populated from command-line
// flags during initialization.

import (
	"flag"
	"time"
)

// Configuration holds all adjustable settings for the editor.
type Configuration struct {
	GutterWidth          int           // Width of the left column (line numbers, LSP signs).
	DefaultTabWidth      int           // Number of spaces a tab character represents.
	FuzzyFinderHeight    int           // Number of rows the fuzzy finder takes up.
	LeaderKey            rune          // The prefix key for many custom commands (default: \).
	UseLogFile           bool          // Whether to write debug logs to a file.
	LogFilePath          string        // Where to store the debug logs.
	NumLogsInDebugWindow int           // How many recent logs to show in the UI debug window.
	OllamaCheckInterval  time.Duration // How often to check if Ollama is running.
	FileCheckInterval    time.Duration // How often to check for external file changes.
	OllamaURL            string        // Endpoint for the Ollama AI service.
	OllamaModel          string        // The specific AI model to use for completion.
	DevMode              bool          // Enables verbose logging and developer tools.
	ShowColors           bool          // Command-line flag to show available colors and exit.
	ShowInfo             bool          // Command-line flag to show file types and exit.
	ShowVersion          bool          // Command-line flag to show version and exit.
	FormatterMarkers     []string      // List of comment prefixes for text formatting (no CLI flag).
}

// Config is the global configuration instance.
var Config Configuration

// InitConfig sets up command-line flags and parses them into the global Config.
func InitConfig() {
	var leaderKey string

	flag.IntVar(&Config.GutterWidth, "gutter-width", 7, "Width of the gutter")
	flag.IntVar(&Config.DefaultTabWidth, "tab-width", 4, "Default tab width")
	flag.IntVar(&Config.FuzzyFinderHeight, "fuzzy-height", 8, "Height of fuzzy finder")
	flag.StringVar(&leaderKey, "leader", "\\", "Leader key")
	flag.BoolVar(&Config.UseLogFile, "log", false, "Enable logging to file")
	flag.StringVar(&Config.LogFilePath, "log-path", "/tmp/qwe-editor-debug.log", "Path to log file")
	flag.IntVar(&Config.NumLogsInDebugWindow, "num-logs", 10, "Number of logs in debug window")
	flag.DurationVar(&Config.OllamaCheckInterval, "ollama-interval", 5*time.Second, "Ollama check interval")
	flag.DurationVar(&Config.FileCheckInterval, "file-check-interval", 2*time.Second, "File check interval")
	flag.StringVar(&Config.OllamaURL, "ollama-url", "http://localhost:11434", "Ollama URL")
	flag.StringVar(&Config.OllamaModel, "ollama-model", "qwen2.5-coder:latest", "Ollama model")
	flag.BoolVar(&Config.DevMode, "dev", false, "Enable development mode")
	flag.BoolVar(&Config.ShowColors, "colors", false, "Show available colors")
	flag.BoolVar(&Config.ShowInfo, "info", false, "Show file associations and LSP info")
	flag.BoolVar(&Config.ShowVersion, "version", false, "Show version")

	flag.Parse()

	// Convert the first character of the leader flag into a rune.
	if len(leaderKey) > 0 {
		Config.LeaderKey = rune(leaderKey[0])
	}

	// Initialize formatter markers for text formatting.
	Config.FormatterMarkers = []string{
		"//", // C/C++/Go/JavaScript/Rust
		"--", // SQL/Lua/Haskell
		"#",  // Python/Shell/Ruby/YAML
		";;", // Lisp/Scheme
		"%",  // LaTeX/MATLAB
		">",  // Markdown quote
	}
}
