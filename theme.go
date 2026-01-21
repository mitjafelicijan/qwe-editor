package main

// Color palette and theme used by the editor. Maps semantic color names (like
// ColorNormalMode) to specific terminal attributes (foreground and background).

import "github.com/nsf/termbox-go"

// To see available colors execute `qwe -colors`.

// Color represents a pair of foreground and background terminal attributes.
type Color struct {
	Background termbox.Attribute
	Foreground termbox.Attribute
}

// ColorName is an enum-like type for semantic color identifiers.
type ColorName int

const (
	ColorDefault ColorName = iota // Default terminal colors.
	ColorSourceString
	ColorSourceKeyword
	ColorSourceNumber
	ColorSourceComment
	ColorSourceMacro
	ColorSourceOther

	ColorAnnotationTodo  // Highlighting for TODO comments.
	ColorAnnotationFixme // Highlighting for FIXME comments.

	ColorStatusBar           // Main status bar at the bottom.
	ColorDebugWindow         // Overlay window for logs/debug info.
	ColorNormalMode          // Status bar indicator for Normal mode.
	ColorInsertMode          // Status bar indicator for Insert mode.
	ColorHighlightedLine     // Background for the line where the cursor is.
	ColorVisualModeSelection // Selection color in all visual modes.
	ColorVisualMode          // Status bar indicator for Visual mode.
	ColorSearchMatch         // Highlighting for found search terms.
	ColorReplaceMatch        // Highlighting for replacement targets.
	ColorCursor              // The color of the cursor itself.

	ColorGutterLineNumber   // Line numbers in the left gutter.
	ColorGutterSignError    // LSP error icons in the gutter.
	ColorGutterSignWarning  // LSP warning icons in the gutter.
	ColorGutterSignInfo     // LSP info icons in the gutter.
	ColorGutterSignHint     // LSP hint icons in the gutter.
	ColorFuzzyResult        // Plain text in fuzzy finder results.
	ColorFuzzySelected      // Highlighted item in fuzzy finder.
	ColorEmptyLineMarker    // The '~' marker for lines beyond EOF.
	ColorDebugTitle         // Header for the debug window.
	ColorDiagSummaryError   // Error count in the status bar.
	ColorDiagSummaryWarning // Warning count in the status bar.
	ColorFuzzyModeBuffers   // Indicator that fuzzy finder is searching buffers.
	ColorFuzzyModeFiles     // Indicator that fuzzy finder is searching files.
	ColorFuzzyModeWarnings  // Indicator that fuzzy finder is searching diagnostics.

	// Colors for Tree-sitter syntax highlighting.
	ColorTSFunction
	ColorTSVariable
	ColorTSType
	ColorTSString
	ColorTSKeyword
	ColorTSComment
	ColorTSNumber
	ColorTSBoolean
	ColorTSNull
	ColorTSProperty
	ColorTSTag
	ColorTSAttribute
	ColorTSConstant

	// External service status indicators.
	ColorLSPStatusConnected
	ColorLSPStatusDisconnected
	ColorOllamaStatusConnected
	ColorOllamaStatusDisconnected

	ColorHoverWindow // LSP hover information popup.
	ColorAutocompleteWindow
	ColorAutocompleteSelected
)

// Theme maps each ColorName to its actual visual attributes.
var Theme = map[ColorName]Color{
	ColorDefault: {Background: termbox.ColorDefault, Foreground: termbox.Attribute(254)},

	// Annotations
	ColorAnnotationTodo:  {Background: termbox.Attribute(221), Foreground: termbox.Attribute(1)},
	ColorAnnotationFixme: {Background: termbox.Attribute(142), Foreground: termbox.Attribute(1)},

	// Status bar
	ColorStatusBar: {Background: termbox.Attribute(250), Foreground: termbox.Attribute(1)},

	// Debug window
	ColorDebugWindow: {Background: termbox.Attribute(19), Foreground: termbox.Attribute(16)},

	// UI colors
	ColorNormalMode:          {Background: termbox.Attribute(250), Foreground: termbox.Attribute(1)},
	ColorInsertMode:          {Background: termbox.Attribute(58), Foreground: termbox.Attribute(255)},
	ColorHighlightedLine:     {Background: termbox.Attribute(235), Foreground: termbox.ColorDefault},
	ColorVisualModeSelection: {Background: termbox.Attribute(46), Foreground: termbox.Attribute(1)},
	ColorVisualMode:          {Background: termbox.Attribute(30), Foreground: termbox.Attribute(16)},
	ColorSearchMatch:         {Background: termbox.Attribute(166), Foreground: termbox.Attribute(1)},
	ColorReplaceMatch:        {Background: termbox.Attribute(221), Foreground: termbox.Attribute(1)},
	ColorCursor:              {Background: termbox.Attribute(252), Foreground: termbox.ColorWhite},

	ColorGutterLineNumber:  {Background: termbox.ColorDefault, Foreground: termbox.Attribute(244)},
	ColorGutterSignError:   {Background: termbox.Attribute(125), Foreground: termbox.Attribute(16)},
	ColorGutterSignWarning: {Background: termbox.Attribute(221), Foreground: termbox.Attribute(1)},
	ColorGutterSignInfo:    {Background: termbox.Attribute(221), Foreground: termbox.Attribute(1)},
	ColorGutterSignHint:    {Background: termbox.Attribute(221), Foreground: termbox.Attribute(1)},

	ColorFuzzyResult:       {Background: termbox.ColorDefault, Foreground: termbox.Attribute(254)},
	ColorFuzzySelected:     {Background: termbox.Attribute(236), Foreground: termbox.Attribute(254)},
	ColorFuzzyModeBuffers:  {Background: termbox.Attribute(125), Foreground: termbox.Attribute(255)},
	ColorFuzzyModeFiles:    {Background: termbox.Attribute(125), Foreground: termbox.Attribute(255)},
	ColorFuzzyModeWarnings: {Background: termbox.Attribute(33), Foreground: termbox.Attribute(255)},

	ColorEmptyLineMarker: {Background: termbox.ColorDefault, Foreground: termbox.Attribute(244)},

	ColorDebugTitle:         {Background: termbox.Attribute(19), Foreground: termbox.Attribute(215)},
	ColorDiagSummaryError:   {Background: termbox.ColorDefault, Foreground: termbox.Attribute(166)},
	ColorDiagSummaryWarning: {Background: termbox.ColorDefault, Foreground: termbox.Attribute(221)},

	// Tree-sitter
	ColorTSFunction:  {Background: termbox.ColorDefault, Foreground: termbox.Attribute(3)},
	ColorTSVariable:  {Background: termbox.ColorDefault, Foreground: termbox.Attribute(255)},
	ColorTSType:      {Background: termbox.ColorDefault, Foreground: termbox.Attribute(112)},
	ColorTSString:    {Background: termbox.ColorDefault, Foreground: termbox.Attribute(37)},
	ColorTSKeyword:   {Background: termbox.ColorDefault, Foreground: termbox.Attribute(178)},
	ColorTSComment:   {Background: termbox.ColorDefault, Foreground: termbox.Attribute(244)},
	ColorTSNumber:    {Background: termbox.ColorDefault, Foreground: termbox.Attribute(135)},
	ColorTSBoolean:   {Background: termbox.ColorDefault, Foreground: termbox.Attribute(2)},
	ColorTSNull:      {Background: termbox.ColorDefault, Foreground: termbox.Attribute(135)},
	ColorTSProperty:  {Background: termbox.ColorDefault, Foreground: termbox.Attribute(230)},
	ColorTSTag:       {Background: termbox.ColorDefault, Foreground: termbox.Attribute(118)},
	ColorTSAttribute: {Background: termbox.ColorDefault, Foreground: termbox.Attribute(215)},
	ColorTSConstant:  {Background: termbox.ColorDefault, Foreground: termbox.Attribute(254)},

	// Status bar indicators
	ColorLSPStatusConnected:       {Background: termbox.Attribute(29), Foreground: termbox.Attribute(255)},
	ColorLSPStatusDisconnected:    {Background: termbox.Attribute(239), Foreground: termbox.Attribute(255)},
	ColorOllamaStatusConnected:    {Background: termbox.Attribute(131), Foreground: termbox.Attribute(255)},
	ColorOllamaStatusDisconnected: {Background: termbox.Attribute(239), Foreground: termbox.Attribute(255)},

	ColorHoverWindow: {Background: termbox.Attribute(253), Foreground: termbox.Attribute(1)},

	ColorAutocompleteWindow:   {Background: termbox.Attribute(253), Foreground: termbox.Attribute(1)},
	ColorAutocompleteSelected: {Background: termbox.Attribute(239), Foreground: termbox.Attribute(255)},
}

// GetThemeColor returns the foreground and background attributes for a given semantic name.
func GetThemeColor(name ColorName) (termbox.Attribute, termbox.Attribute) {
	if c, ok := Theme[name]; ok {
		return c.Foreground, c.Background
	}
	// Fallback to default if name is not found.
	return termbox.ColorDefault, termbox.ColorDefault
}
