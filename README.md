`qwe` is a small, opinionated modal text editor built with Go with some
batteries included like Tree-sitter integration for better syntax highlighting,
basic LSP support, Ollama integration, and more.

> [!NOTE]
> This is a work in progress and built specifically for personal use. Do not
> expect miracles. I will add features as I need them.

I made this editor to learn about modal text editors and have a place to
experiment with different ideas. It is not intended to emulate any existing
editor, even though it shares some similarities with some of them like Vim in
particular.

Most of the keybindings are based on Vim, but there are some differences.

https://github.com/user-attachments/assets/2e0ebe2d-98a6-447d-8e20-c00aa88fd7ee

## Features

- Modal Design (Insert/Normal/Visual/Command)
- Tree-sitter Syntax Highlighting
- LSP Support (Hover, Autocomplete, Definition, Diagnostics)
- Fuzzy Finder (Files, Buffers, Warning Quickfix)
- Jumplists (Normal/Visual)
- Multi-Cursor (Normal/Visual)
- Text Formatting (Normal/Visual)
- Ollama Support (using local models)

## Configuration

Instead of having a configuration file, the editor uses command-line flags:

- `-colors`: Show available colors
- `-dev`: Enable development mode
- `-file-check-interval`: File check interval (default 2s)
- `-fuzzy-height`: Height of fuzzy finder (default 8)
- `-gutter-width`: Width of the gutter (default 7)
- `-info`: Show file associations and LSP info
- `-leader`: Leader key (default "\\")
- `-log`: Enable logging to file
- `-log-path`: Path to log file (default "/tmp/qwe-editor-debug.log")
- `-num-logs`: Number of logs in debug window (default 10)
- `-ollama-interval`: Ollama check interval (default 5s)
- `-ollama-model`: Ollama model (default "qwen2.5-coder:latest")
- `-ollama-url`: Ollama URL (default "http://localhost:11434")
- `-tab-width`: Default tab width (default 4)
- `-version`: Show version

## Download pre-built binary

Download a pre-built binary from [GitHub
releases](https://github.com/mitjafelicijan/qwe-editor/releases).

> [!IMPORTANT]
> macOS users will need to remove the quarantine bit from the binary before
> running it with `xattr -d com.apple.quarantine qwe-macos`

## Build from source

- Clone the repository
- Run `make -B`
- Run `./qwe` to start the editor
- Or run `make install` to install the editor

## Release process

- Tag a new version with `git tag vX.Y.Z`
- Push the tag with `git push origin --tags`
- Wait for the GitHub Actions workflow to finish

## Special thanks

- https://github.com/tree-sitter/tree-sitter
- https://github.com/orgs/tree-sitter/repositories
- https://github.com/sourcegraph/go-lsp
- https://github.com/ollama/ollama
