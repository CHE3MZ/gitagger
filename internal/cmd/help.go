// PrintHelp shows friendly usage.
package cmd

import (
	"fmt"

	"github.com/CHE3MZ/gitagger/internal/style"
)

// PrintHelp prints the short friendly help.
func PrintHelp() {
	fmt.Println(style.Bold(style.Blue("gitagger — tags without the headache")))
	fmt.Println(style.Blue("Just run `gitagger` and it picks the next tag from your history, then pushes it if it can."))
	fmt.Println()
	fmt.Println(style.Header("Use it"))
	fmt.Printf("  %s  %s\n", style.BoldGreen("gitagger"), style.White("make the next patch + push"))
	fmt.Printf("  %s   %s\n", style.White("gitagger minor"), style.Dim("v1.2.3 -> v1.3.0"))
	fmt.Printf("  %s  %s\n", style.White("gitagger major --pre rc"), style.Dim("v1.2.3 -> v2.0.0-rc"))
	fmt.Printf("  %s  %s\n", style.White("gitagger --dry-run"), style.Dim("peek first, change nothing"))
	fmt.Printf("  %s   %s\n", style.White("gitagger --no-push"), style.Dim("tag locally only"))
	fmt.Println()
	fmt.Println(style.Header("Flags"))
	fmt.Println(style.Blue("  --pre stable|rc|beta|build|nightly   prerelease flavor (default stable)"))
	fmt.Println(style.Blue("  --format auto|triple|double|single|date|sha|sha-num   force a style (default auto)"))
	fmt.Println(style.Blue("  --no-push, --confirm, --remote, --force, --dry-run, -m, --json, --verbose"))
	fmt.Println()
	fmt.Println(style.Header("More"))
	fmt.Printf("  %s  %s\n", style.White("gitagger init"), style.Dim("write a .gitagger.yml once, then forget flags"))
	fmt.Printf("  %s  %s\n", style.White("gitagger list"), style.Dim("show tags, newest last"))
	fmt.Printf("  %s  %s\n", style.White("gitagger doctor"), style.Dim("sanity check local vs remote"))
	fmt.Println()
	fmt.Println(style.Dim("No prompts by default. No remote or offline? Your tag stays local with a hint."))
}
