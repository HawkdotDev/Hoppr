package ui

import (
	"os"
)

// ANSI color escape codes
const (
	ColorReset  = "\033[0m"
	ColorBold   = "\033[1m"
	ColorDim    = "\033[2m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorCyan   = "\033[36m"
)

// ColorsEnabled checks whether ANSI colors should be used.
func ColorsEnabled() bool {
	// Respect NO_COLOR standard (https://no-color.org)
	if _, exists := os.LookupEnv("NO_COLOR"); exists {
		return false
	}
	if os.Getenv("TERM") == "dumb" {
		return false
	}
	return true
}

func Colorize(color, text string) string {
	if !ColorsEnabled() {
		return text
	}
	return color + text + ColorReset
}

func Bold(text string) string   { return Colorize(ColorBold, text) }
func Green(text string) string  { return Colorize(ColorGreen, text) }
func Yellow(text string) string { return Colorize(ColorYellow, text) }
func Red(text string) string    { return Colorize(ColorRed, text) }
func Cyan(text string) string   { return Colorize(ColorCyan, text) }
func Dim(text string) string    { return Colorize(ColorDim, text) }
