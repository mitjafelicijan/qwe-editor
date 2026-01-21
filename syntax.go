package main

// Syntax highlighting using tree-sitter. It parses the buffer content, executes
// queries to find semantic tokens, and maps those tokens to theme colors.

import (
	"context"
	"fmt"

	sitter "github.com/mitjafelicijan/go-tree-sitter"
	"github.com/mitjafelicijan/go-tree-sitter/bash"
	"github.com/mitjafelicijan/go-tree-sitter/c"
	"github.com/mitjafelicijan/go-tree-sitter/cpp"
	"github.com/mitjafelicijan/go-tree-sitter/css"
	"github.com/mitjafelicijan/go-tree-sitter/dockerfile"
	"github.com/mitjafelicijan/go-tree-sitter/golang"
	"github.com/mitjafelicijan/go-tree-sitter/html"
	"github.com/mitjafelicijan/go-tree-sitter/javascript"
	"github.com/mitjafelicijan/go-tree-sitter/lua"
	markdown "github.com/mitjafelicijan/go-tree-sitter/markdown/tree-sitter-markdown"
	"github.com/mitjafelicijan/go-tree-sitter/php"
	"github.com/mitjafelicijan/go-tree-sitter/python"
	"github.com/mitjafelicijan/go-tree-sitter/sql"
	"github.com/mitjafelicijan/go-tree-sitter/typescript/tsx"
	"github.com/mitjafelicijan/go-tree-sitter/typescript/typescript"
	"github.com/nsf/termbox-go"
)

// SyntaxHighlighter manages the tree-sitter parser, tree, and calculated highlights for a buffer.
type SyntaxHighlighter struct {
	Parser     *sitter.Parser
	Tree       *sitter.Tree
	Lang       *sitter.Language
	Query      *sitter.Query
	Language   string
	Highlights map[int]map[int]termbox.Attribute // Cached colors: Line -> Col -> termbox.Attribute
	Log        func(string, string)              // Debug logging function.
}

// NewSyntaxHighlighter initializes a parser for the given file type.
func NewSyntaxHighlighter(fileType string, log func(string, string)) *SyntaxHighlighter {
	parser := sitter.NewParser()
	var lang *sitter.Language
	var langName string

	// Map internal FileType names to tree-sitter languages.
	switch fileType {
	case "C":
		lang = c.GetLanguage()
		langName = "c"
	case "C++":
		lang = cpp.GetLanguage()
		langName = "cpp"
	case "Go":
		lang = golang.GetLanguage()
		langName = "go"
	case "JavaScript":
		lang = javascript.GetLanguage()
		langName = "javascript"
	case "TypeScript":
		lang = typescript.GetLanguage()
		langName = "typescript"
	case "TSX":
		lang = tsx.GetLanguage()
		langName = "tsx"
	case "Python":
		lang = python.GetLanguage()
		langName = "python"
	case "Bash":
		lang = bash.GetLanguage()
		langName = "bash"
	case "CSS":
		lang = css.GetLanguage()
		langName = "css"
	case "Dockerfile":
		lang = dockerfile.GetLanguage()
		langName = "dockerfile"
	case "HTML":
		lang = html.GetLanguage()
		langName = "html"
	case "Lua":
		lang = lua.GetLanguage()
		langName = "lua"
	case "Markdown":
		lang = markdown.GetLanguage()
		langName = "markdown"
	case "PHP":
		lang = php.GetLanguage()
		langName = "php"
	case "SQL":
		lang = sql.GetLanguage()
		langName = "sql"
	default:
		return nil
	}

	parser.SetLanguage(lang)
	s := &SyntaxHighlighter{
		Parser:     parser,
		Lang:       lang,
		Language:   langName,
		Highlights: make(map[int]map[int]termbox.Attribute),
		Log:        log,
	}

	// Load the tree-sitter query file (.scm) for this language.
	queryPath := fmt.Sprintf("queries/%s.scm", langName)
	s.LoadQuery(queryPath)

	return s
}

// LoadQuery reads and compiles a tree-sitter query from the embedded filesystem.
func (s *SyntaxHighlighter) LoadQuery(path string) {
	if s.Log != nil {
		s.Log("TS", fmt.Sprintf("Loading query for %s", path))
	}

	content, err := QueriesFS.ReadFile(path)
	if err != nil {
		if s.Log != nil {
			s.Log("TS", fmt.Sprintf("LoadQuery failed to read %s: %v", path, err))
		}
		return
	}

	q, err := sitter.NewQuery(content, s.Lang)
	if err == nil {
		s.Query = q
	} else if s.Log != nil {
		s.Log("TS", fmt.Sprintf("LoadQuery failed to compile query for %s: %v", path, err))
	}
}

// Parse runs a full parse of the content and updates the highlight cache.
func (s *SyntaxHighlighter) Parse(content []byte) {
	if s.Parser == nil {
		return
	}
	tree, _ := s.Parser.ParseCtx(context.Background(), nil, content)
	s.Tree = tree
	s.updateHighlights(content)
}

// Reparse is a wrapper around Parse (used for batch updates).
func (s *SyntaxHighlighter) Reparse(content []byte) {
	s.Parse(content)
}

// Edit is a placeholder for incremental parsing (currently does a full reparse).
func (s *SyntaxHighlighter) Edit(edit sitter.EditInput, newContent []byte) {
	s.Reparse(newContent)
}

// updateHighlights executes the tree-sitter query on the syntax tree and populates the highlight cache.
func (s *SyntaxHighlighter) updateHighlights(source []byte) {
	// Always clear previous highlights to prevent ghosting.
	s.Highlights = make(map[int]map[int]termbox.Attribute)

	if s.Tree == nil || s.Query == nil {
		return
	}

	qc := sitter.NewQueryCursor()
	qc.Exec(s.Query, s.Tree.RootNode())

	for {
		m, ok := qc.NextMatch()
		if !ok {
			break
		}

		for _, c := range m.Captures {
			// Find the theme attribute for the capture name (e.g., "function", "keyword").
			captureName := s.Query.CaptureNameForId(c.Index)
			attr := getTermboxAttr(captureName)

			startRow := int(c.Node.StartPoint().Row)
			startCol := int(c.Node.StartPoint().Column)
			endRow := int(c.Node.EndPoint().Row)
			endCol := int(c.Node.EndPoint().Column)

			// Map the capture span to line/column color attributes.
			for r := startRow; r <= endRow; r++ {
				if _, ok := s.Highlights[r]; !ok {
					s.Highlights[r] = make(map[int]termbox.Attribute)
				}

				cStart := 0
				if r == startRow {
					cStart = startCol
				}

				cEnd := -1
				if r == endRow {
					cEnd = endCol
				}

				limit := cEnd
				if limit == -1 {
					limit = 1000 // A reasonable overflow limit for whole lines.
				}

				for col := cStart; col < limit; col++ {
					s.Highlights[r][col] = attr
				}
			}
		}
	}
}

// getTermboxAttr maps a tree-sitter capture name to a color name from our theme.
func getTermboxAttr(captureName string) termbox.Attribute {
	var cn ColorName
	switch captureName {
	case "function":
		cn = ColorTSFunction
	case "tag":
		cn = ColorTSTag
	case "attribute":
		cn = ColorTSAttribute
	case "constant":
		cn = ColorTSConstant
	case "variable":
		cn = ColorTSVariable
	case "type":
		cn = ColorTSType
	case "string":
		cn = ColorTSString
	case "keyword":
		cn = ColorTSKeyword
	case "comment":
		cn = ColorTSComment
	case "number":
		cn = ColorTSNumber
	case "boolean":
		cn = ColorTSBoolean
	case "null":
		cn = ColorTSNull
	case "property":
		cn = ColorTSProperty
	default:
		return termbox.ColorDefault
	}

	fg, _ := GetThemeColor(cn)
	return fg
}

// Highlight returns a slice of attributes for each character in a line.
func (s *SyntaxHighlighter) Highlight(lineIdx int, lineContent []rune) []termbox.Attribute {
	attrs := make([]termbox.Attribute, len(lineContent))
	// Fill with default foreground color first.
	defaultFg, _ := GetThemeColor(ColorDefault)
	for i := range attrs {
		attrs[i] = defaultFg
	}

	// Apply cached highlights if they exist for this line.
	if lineHighlights, ok := s.Highlights[lineIdx]; ok {
		for col, color := range lineHighlights {
			if col < len(attrs) {
				attrs[col] = color
			}
		}
	}

	return attrs
}
