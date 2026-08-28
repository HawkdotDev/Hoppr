package ui

import (
	"fmt"
	"os"
	"strings"
)

// ANSI Escape Codes
const (
	Reset        = "\033[0m"
	BoldCode     = "\033[1m"
	DimCode      = "\033[2m"
	ItalicCode   = "\033[3m"
	UnderlineCode = "\033[4m"

	// 16-color standard
	FgBlack   = "\033[30m"
	FgRed     = "\033[31m"
	FgGreen   = "\033[32m"
	FgYellow  = "\033[33m"
	FgBlue    = "\033[34m"
	FgMagenta = "\033[35m"
	FgCyan    = "\033[36m"
	FgWhite   = "\033[37m"

	// High-intensity foreground
	FgHiBlack   = "\033[90m" // Gray / Slate
	FgHiRed     = "\033[91m"
	FgHiGreen   = "\033[92m"
	FgHiYellow  = "\033[93m"
	FgHiBlue    = "\033[94m"
	FgHiMagenta = "\033[95m"
	FgHiCyan    = "\033[96m"
	FgHiWhite   = "\033[97m"

	// Backgrounds
	BgBlack   = "\033[40m"
	BgRed     = "\033[41m"
	BgGreen   = "\033[42m"
	BgYellow  = "\033[43m"
	BgBlue    = "\033[44m"
	BgMagenta = "\033[45m"
	BgCyan    = "\033[46m"
	BgWhite   = "\033[47m"

	// 256-color palette presets
	FgViolet  = "\033[38;5;141m" // Electric Violet
	FgAmber   = "\033[38;5;214m" // Warm Amber
	FgEmerald = "\033[38;5;42m"  // Emerald Green
	FgRose    = "\033[38;5;204m" // Rose Pink
	FgSky     = "\033[38;5;75m"  // Sky Blue

	BgVioletPill = "\033[48;5;54;38;5;255m" // Deep Violet Pill with White Text
	BgCyanPill   = "\033[48;5;24;38;5;255m" // Deep Cyan Pill with White Text
	BgGrayPill   = "\033[48;5;238;38;5;250m"
)

// Symbols & Icons
const (
	IconCheck   = "✔"
	IconCross   = "✖"
	IconWarn    = "⚠"
	IconInfo    = "ℹ"
	IconArrow   = "→"
	IconZap     = "⚡"
	IconFolder  = "📁"
	IconBranch  = "├──"
	IconLast    = "└──"
	IconPipe    = "│  "
)

// ColorsEnabled checks whether terminal formatting should be applied.
func ColorsEnabled() bool {
	if _, exists := os.LookupEnv("NO_COLOR"); exists {
		return false
	}
	if os.Getenv("TERM") == "dumb" {
		return false
	}
	return true
}

// Colorize wraps text in ANSI code and resets.
func Colorize(code, text string) string {
	if !ColorsEnabled() {
		return text
	}
	return code + text + Reset
}

// Fluent Style Builder (Chalk-like API in Go)
type Style struct {
	codes []string
}

func NewStyle() *Style {
	return &Style{codes: make([]string, 0, 4)}
}

func (s *Style) Bold() *Style          { s.codes = append(s.codes, BoldCode); return s }
func (s *Style) Dim() *Style           { s.codes = append(s.codes, DimCode); return s }
func (s *Style) Italic() *Style        { s.codes = append(s.codes, ItalicCode); return s }
func (s *Style) Underline() *Style     { s.codes = append(s.codes, UnderlineCode); return s }
func (s *Style) Red() *Style           { s.codes = append(s.codes, FgRed); return s }
func (s *Style) Green() *Style         { s.codes = append(s.codes, FgGreen); return s }
func (s *Style) Yellow() *Style        { s.codes = append(s.codes, FgYellow); return s }
func (s *Style) Blue() *Style          { s.codes = append(s.codes, FgBlue); return s }
func (s *Style) Magenta() *Style       { s.codes = append(s.codes, FgMagenta); return s }
func (s *Style) Cyan() *Style          { s.codes = append(s.codes, FgCyan); return s }
func (s *Style) Gray() *Style          { s.codes = append(s.codes, FgHiBlack); return s }
func (s *Style) Violet() *Style        { s.codes = append(s.codes, FgViolet); return s }
func (s *Style) Amber() *Style         { s.codes = append(s.codes, FgAmber); return s }
func (s *Style) Emerald() *Style       { s.codes = append(s.codes, FgEmerald); return s }
func (s *Style) Rose() *Style          { s.codes = append(s.codes, FgRose); return s }
func (s *Style) Sky() *Style           { s.codes = append(s.codes, FgSky); return s }

func (s *Style) Render(format string, a ...any) string {
	text := fmt.Sprintf(format, a...)
	if !ColorsEnabled() || len(s.codes) == 0 {
		return text
	}
	return strings.Join(s.codes, "") + text + Reset
}

// Quick Style Helpers
func Bold(text string) string      { return Colorize(BoldCode, text) }
func Dim(text string) string       { return Colorize(DimCode, text) }
func Gray(text string) string      { return Colorize(FgHiBlack, text) }
func Red(text string) string       { return Colorize(FgRed, text) }
func Green(text string) string     { return Colorize(FgGreen, text) }
func Yellow(text string) string    { return Colorize(FgYellow, text) }
func Blue(text string) string      { return Colorize(FgBlue, text) }
func Magenta(text string) string   { return Colorize(FgMagenta, text) }
func Cyan(text string) string      { return Colorize(FgCyan, text) }
func Violet(text string) string    { return Colorize(FgViolet, text) }
func Amber(text string) string     { return Colorize(FgAmber, text) }
func Emerald(text string) string   { return Colorize(FgEmerald, text) }
func Rose(text string) string      { return Colorize(FgRose, text) }
func Sky(text string) string       { return Colorize(FgSky, text) }

// Badges & Pills
func Badge(label, bgCode string) string {
	if !ColorsEnabled() {
		return "[" + label + "]"
	}
	return bgCode + " " + label + " " + Reset
}

func ListBadge(listName string, isDefault bool) string {
	if isDefault {
		return Colorize(FgViolet+BoldCode, listName) + " " + Colorize(FgHiBlack, "(default)")
	}
	return Colorize(FgCyan+BoldCode, listName)
}

func PathStyle(path string) string {
	return Colorize(FgHiBlack, path)
}
