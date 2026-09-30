// RunView shows details for one local tag.
package cmd

import (
	"fmt"
	"strings"

	"github.com/CHE3MZ/gitagger/internal/git"
	"github.com/CHE3MZ/gitagger/internal/style"
)

// RunView prints name, type, message, commit and date for tag.
// Missing tags are an error (exit 1). Read-only, never touches the remote.
func RunView(dir, tag string, verbose bool) error {
	if err := git.EnsureAvailable(); err != nil {
		return GenericErr(err)
	}
	if strings.TrimSpace(tag) == "" {
		return BadArgs("give a tag name — try `gitagger view --help`")
	}
	info, err := git.InspectTag(dir, tag)
	if err != nil {
		return Generic("no tag %q in this project.", strings.TrimSpace(tag))
	}
	fmt.Printf("tag: %s\n", style.Bold(info.Name))
	fmt.Printf("type: %s\n", style.White(info.Kind))
	if info.Message != "" {
		fmt.Printf("message: %s\n", style.White(info.Message))
	} else {
		fmt.Println(style.Dim("message: (none — lightweight tag)"))
	}
	fmt.Printf("commit: %s\n", style.White(info.Commit))
	fmt.Printf("date: %s\n", style.White(info.Date))
	if verbose {
		if url, err := git.RemoteURL(dir, EffectiveRemote(dir)); err == nil {
			fmt.Println(style.Dim("remote: " + url))
		}
	}
	return nil
}
