package main

// Provides a way to view all detected file types and their associated LSP
// (Language Server Protocol) configurations.

import (
	"fmt"
	"strings"
)

// PrintInfo prints a summary table of all supported languages and their LSP setup.
func PrintInfo() {
	// Table header.
	fmt.Printf("%-15s %-10s %-20s\n", "Name", "LSP", "Command")
	fmt.Println(strings.Repeat("-", 80))

	for _, ft := range fileTypes {
		lspEnabled := "no"
		if ft.EnableLSP {
			lspEnabled = "yes"
		}

		lspCmd := ft.LSPCommand
		// Append arguments if they exist (e.g., --stdio).
		if len(ft.LSPCommandArgs) > 0 {
			lspCmd += " " + strings.Join(ft.LSPCommandArgs, " ")
		}

		fmt.Printf("%-15s %-10s %-20s\n", ft.Name, lspEnabled, lspCmd)
	}
}
