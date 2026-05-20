# QWE Editor Documentation

QWE is a modal text editor written in Go.

## Modes

- **Normal**: Navigation and text manipulation (default).
- **Insert**: Text entry.
- **Visual**: Character-wise selection.
- **Visual Line**: Line-wise selection.
- **Visual Block**: Rectangular block selection.
- **Command**: Colon commands (`:`).
- **Find**: Incremental search (`/`).
- **Fuzzy**: File/buffer switching.
- **Replace**: Pattern replacement.
- **Confirm**: Action confirmation.

## Keybindings

### Navigation (Normal Mode)

| Key | Action |
|-----|--------|
| `Arrows` | Move cursor |
| `w` / `q` | Next / Previous word |
| `Q` / `W` | Line start / End |
| `{` / `}` | Top / Bottom of file |
| `[` / `]` | Previous / Next empty line |
| `Ctrl+O` / `Ctrl+I` | Jump history back / forward |
| `Ctrl+N` / `Ctrl+P` | Next / Previous buffer |
| `zz` | Center screen |

### Editing (Normal Mode)

| Key | Action |
|-----|--------|
| `i` / `I` | Insert (before cursor / start of line) |
| `a` / `A` | Append (after cursor / end of line) |
| `o` / `O` | Open new line (below / above) |
| `x` | Delete character |
| `dd` | Delete line |
| `D` | Delete to end of line |
| `cw` | Change word |
| `C` | Change to end of line |
| `cc` / `s` | Change character |
| `y` | Yank current line |
| `p` / `P` | Paste (below / above) |
| `u` / `U` | Undo / Redo |
| `zx` | Toggle comment |
| `zq` | Format text (80 char wrap) |
| `J` | Join lines |
| `Alt+Up/Down` | Move lines up / down |
| `Alt+Left/Right`| Unindent / Indent |
| `Ctrl+A` / `Ctrl+Z`| Increment / Decrement number |
| `c(`, `d[`, etc. | Change / Delete inside delimiters |

### Visual Mode

| Key | Action |
|-----|--------|
| `v` / `V` / `Ctrl+V` | Enter Visual / Visual Line / Visual Block |
| `y` / `d` / `x` | Yank / Delete selection |
| `c` / `p` | Change / Paste selection |
| `zx` / `zq` | Comment / Format selection |
| `~` | Toggle case |
| `o` | Swap cursor and selection anchor |
| `R` | Enter Range Replace mode |
| `Leader+o` | Ollama AI Code Completion |

### Multi-cursor

| Key | Action |
|-----|--------|
| `Ctrl+X` | Add cursor at next word match |
| `Ctrl+Up/Down` | Add cursor above / below |
| `Esc` | Clear secondary cursors |

### Search & Find

| Key | Action |
|-----|--------|
| `/` | Start search |
| `n` / `N` | Next / Previous result |
| `Leader+p` | Find Files |
| `Leader+b` | Find Buffers |
| `Leader+w` | Find Warnings (Diagnostics) |
| `Leader+m` | Find Bookmarks |
| `Leader+q` | Clear Highlighting |

### LSP & Features

| Key | Action |
|-----|--------|
| `gd` | Go to Definition |
| `gf` | Go to File |
| `Ctrl+K` | Show Hover Information |
| `Ctrl+N` (Insert) | Trigger Autocomplete |
| `F1` - `F5` | Goto Bookmark 1-5 |
| `Leader+F1-F5` | Set Bookmark 1-5 |
| `Leader+d` | Delete current buffer |
| `Leader+l` | Toggle debug window |

## Commands (:)

| Command | Action |
|---------|--------|
| `:w [file]` | Save buffer |
| `:wa` | Save all buffers |
| `:q` / `:q!` | Quit (force) |
| `:wq` | Save and quit |
| `:waq` | Save all and quit |
| `:bd` / `:bd!` | Close buffer (force) |
| `:n` | New buffer |
| `:e <file>` | Open file |
| `:reload` | Reload from disk |
| `:help` | Show help |
| `:mouse` | Toggle mouse support |
| `:debug` | Toggle debug window |
| `:!cmd` | Run shell command |
| `:r!cmd` | Run shell command and insert output |
| `:<number>` | Jump to line |

## Technical Features

### Language Server Protocol (LSP)
Supports Hover, Autocomplete, Diagnostics, and Go to Definition via JSON-RPC. Default servers: `gopls` (Go), `clangd` (C/C++), `typescript-language-server` (JS/TS), `pyright-langserver` (Python).

### AI Integration
Local completion via Ollama API. Triggered with `Leader+o` on visual selection.

### Tree-sitter
Syntax highlighting via Tree-sitter. Supported: Go, C, C++, JS/TS, Python, Bash, CSS, PHP, SQL, Markdown, Dockerfile.

### Multi-cursor
Simultaneous editing with multiple cursors. Supports vertical and pattern-based cursor addition.

### Range Replacement
Triggered with `R` in visual mode. Syntax: `/pattern/replacement/flags`. Supports regex and `g`/`i` flags.
