// RunRemote shows the remote URL for a git project.
package cmd

import (
	"fmt"
	"strings"

	"github.com/CHE3MZ/gitagger/internal/git"
	"github.com/CHE3MZ/gitagger/internal/style"
)

// RunRemote prints the URL of remote. With verbose it shows all remotes.
func RunRemote(dir, remote string, verbose bool) error {
	if err := git.EnsureAvailable(); err != nil {
		return GenericErr(err)
	}
	if strings.TrimSpace(remote) == "" {
		remote = "origin"
	}
	url, err := git.RemoteURL(dir, remote)
	if err != nil {
		return Generic("no remote %q for this project.", remote)
	}
	fmt.Printf("%s: %s\n", style.White(remote), style.White(url))
	if verbose {
		if out, err := git.Remotes(dir); err == nil && strings.TrimSpace(out) != "" {
			fmt.Println(style.Dim(out))
		}
	}
	return nil
}
