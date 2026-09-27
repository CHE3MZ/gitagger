// Package style gives gitagger its look: mostly plain text,
// bold for headings, red for errors, green for wins,
// blue for help, gray for side notes, white for everything else.
package style

import "os"

const (
	reset = "\x1b[0m"
	bold  = "\x1b[1m"
	red   = "\x1b[31m"
	green = "\x1b[92m"
	blue  = "\x1b[34m"
	gray  = "\x1b[90m"
	white = "\x1b[37m"
)

// enabled controls whether color codes are emitted.
// Disabled with NO_COLOR, or on dumb terminals.
var enabled = true

func init() {
	if os.Getenv("NO_COLOR") != "" {
		enabled = false
		return
	}
	if os.Getenv("TERM") == "dumb" {
		enabled = false
	}
}

func paint(code, s string) string {
	if !enabled {
		return s
	}
	return code + s + reset
}

// Red is for errors and negative stuff (No, false, failed).
func Red(s string) string { return paint(red, s) }

// Green is for good stuff (yes, true, pushed, created).
func Green(s string) string { return paint(green, s) }

// Blue is for help text and hints.
func Blue(s string) string { return paint(blue, s) }

// Gray is for extra / side-note stuff.
func Gray(s string) string { return paint(gray, s) }

// White is the default for everything else.
func White(s string) string { return paint(white, s) }

// Bold highlights headings and tag names.
func Bold(s string) string { return paint(bold, s) }

// BoldGreen is a tag name that just got created.
func BoldGreen(s string) string {
	if !enabled {
		return s
	}
	return "\x1b[1;92m" + s + reset
}

// Error formats an error line.
func Error(msg string) string { return Red("error: ") + White(msg) }

// Warn formats a warning line (keeps local tag, skips push, etc).
func Warn(msg string) string { return paint("\x1b[33m", "warning: ") + White(msg) }

// Dim formats a side note.
func Dim(msg string) string { return Gray(msg) }

// Header bolds a section title.
func Header(s string) string { return Bold(White(s)) }
