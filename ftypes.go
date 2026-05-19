package main

// Supported file types, their extensions, and language-specific settings like
// indentation and LSP commands.

import "path/filepath"

// FileType represents the configuration for a specific programming language.
type FileType struct {
	Name             string   // Display name of the file type.
	LanguageID       string   // LSP language identifier (e.g., "cpp", "typescriptreact").
	Extensions       []string // File extensions (e.g., .go, .py) or filenames (e.g., Makefile).
	UseTabs          bool     // Whether to use tabs for indentation.
	Comment          string   // Single-line comment prefix (e.g., // or #).
	TabWidth         int      // Number of spaces for a tab.
	EnableLSP        bool     // Whether to enable Language Server Protocol support.
	LSPCommand       string   // Executable name of the LSP server.
	LSPCommandArgs   []string // Arguments to pass to the LSP server.
	FormatterCommand string   // External command for formatting the file.
}

// fileTypes is a global list of all supported languages in the editor.
var fileTypes = []*FileType{
	{
		Name:           "Go",
		LanguageID:     "go",
		Extensions:     []string{".go"},
		UseTabs:        true,
		Comment:        "//",
		TabWidth:       Config.DefaultTabWidth,
		EnableLSP:      true,
		LSPCommand:     "gopls",
		LSPCommandArgs: []string{"serve"},
	},
	{
		Name:       "C",
		LanguageID: "c",
		Extensions: []string{".c", ".h"},
		UseTabs:    true,
		Comment:    "//",
		TabWidth:   Config.DefaultTabWidth,
		EnableLSP:  true,
		LSPCommand: "clangd",
	},
	{
		Name:       "C++",
		LanguageID: "cpp",
		Extensions: []string{".cpp", ".hpp", ".cc", ".hh", ".cxx", ".hxx"},
		UseTabs:    true,
		Comment:    "//",
		TabWidth:   Config.DefaultTabWidth,
		EnableLSP:  true,
		LSPCommand: "clangd",
	},
	{
		Name:           "JavaScript",
		LanguageID:     "javascript",
		Extensions:     []string{".js"},
		UseTabs:        true,
		Comment:        "//",
		TabWidth:       Config.DefaultTabWidth,
		EnableLSP:      true,
		LSPCommand:     "typescript-language-server",
		LSPCommandArgs: []string{"--stdio"},
	},
	{
		Name:           "TypeScript",
		LanguageID:     "typescript",
		Extensions:     []string{".ts"},
		UseTabs:        true,
		Comment:        "//",
		TabWidth:       Config.DefaultTabWidth,
		EnableLSP:      true,
		LSPCommand:     "typescript-language-server",
		LSPCommandArgs: []string{"--stdio"},
	},
	{
		Name:           "TSX",
		LanguageID:     "typescriptreact",
		Extensions:     []string{".tsx"},
		UseTabs:        true,
		Comment:        "//",
		TabWidth:       Config.DefaultTabWidth,
		EnableLSP:      true,
		LSPCommand:     "typescript-language-server",
		LSPCommandArgs: []string{"--stdio"},
	},
	{
		Name:           "Python",
		LanguageID:     "python",
		Extensions:     []string{".py"},
		UseTabs:        false,
		Comment:        "#",
		TabWidth:       Config.DefaultTabWidth,
		EnableLSP:      true,
		LSPCommand:     "pyright-langserver",
		LSPCommandArgs: []string{"--stdio"},
	},
	{
		Name:       "Bash",
		LanguageID: "shellscript",
		Extensions: []string{".sh"},
		UseTabs:    true,
		Comment:    "#",
		TabWidth:   Config.DefaultTabWidth,
	},
	{
		Name:       "CSS",
		LanguageID: "css",
		Extensions: []string{".css"},
		UseTabs:    false,
		Comment:    "//",
		TabWidth:   Config.DefaultTabWidth,
	},
	{
		Name:       "Dockerfile",
		LanguageID: "dockerfile",
		Extensions: []string{".dockerfile", "Dockerfile"},
		UseTabs:    false,
		Comment:    "#",
		TabWidth:   Config.DefaultTabWidth,
	},
	{
		Name:       "HTML",
		LanguageID: "html",
		Extensions: []string{".html", ".htm"},
		UseTabs:    false,
		Comment:    "",
		TabWidth:   Config.DefaultTabWidth,
	},
	{
		Name:       "Lua",
		LanguageID: "lua",
		Extensions: []string{".lua"},
		UseTabs:    true,
		Comment:    "--",
		TabWidth:   Config.DefaultTabWidth,
	},
	{
		Name:       "Markdown",
		LanguageID: "markdown",
		Extensions: []string{".md", ".markdown"},
		UseTabs:    false,
		Comment:    "",
		TabWidth:   Config.DefaultTabWidth,
	},
	{
		Name:       "PHP",
		LanguageID: "php",
		Extensions: []string{".php"},
		UseTabs:    true,
		Comment:    "//",
		TabWidth:   Config.DefaultTabWidth,
	},
	{
		Name:       "SQL",
		LanguageID: "sql",
		Extensions: []string{".sql"},
		UseTabs:    true,
		Comment:    "--",
		TabWidth:   Config.DefaultTabWidth,
	},
	{
		Name:       "Makefile",
		LanguageID: "makefile",
		Extensions: []string{".make", "Makefile", "makefile"},
		UseTabs:    true,
		Comment:    "#",
		TabWidth:   Config.DefaultTabWidth,
	},
	{
		Name:       "Text",
		LanguageID: "plaintext",
		Extensions: []string{},
		UseTabs:    false,
		Comment:    "",
		TabWidth:   Config.DefaultTabWidth,
	},
}

// getFileType detects the file type based on the filename or extension.
func getFileType(filename string) *FileType {
	ext := filepath.Ext(filename)
	base := filepath.Base(filename)
	for _, ft := range fileTypes {
		for _, e := range ft.Extensions {
			// Check if the extension matches or if the base filename (like 'Makefile') matches.
			if e == ext || e == base {
				return ft
			}
		}
	}
	// Return Default file type (Text) if no match is found.
	return fileTypes[len(fileTypes)-1]
}

// InitFileTypes resets language settings to the current global configuration.
func InitFileTypes() {
	for _, ft := range fileTypes {
		ft.TabWidth = Config.DefaultTabWidth
	}
}
