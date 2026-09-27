// Prompt helpers: confirm prompts only on interactive TTY.
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/CHE3MZ/gitagger/internal/style"
)

// IsTTY reports whether stdin is an interactive terminal.
func IsTTY() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

// AskYes prompts "[Y/n]" and defaults to yes on empty input.
// Non-TTY callers should skip prompting entirely (see ShouldPrompt).
func AskYes(prompt string) bool {
	fmt.Printf("%s %s ", style.White(prompt), style.Gray("[Y/n]"))
	line, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		return true
	}
	line = strings.ToLower(strings.TrimSpace(line))
	return line == "" || line == "y" || line == "yes"
}

// ShouldPrompt: prompt only when confirm is set
// AND pushing AND not dry-run AND on a TTY.
func ShouldPrompt(confirm, push, dryRun bool) bool {
	return confirm && push && !dryRun && IsTTY()
}
