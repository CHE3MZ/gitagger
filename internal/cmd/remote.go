// RunRemote changes or shows the push remote for a git project.
package cmd

import (
	"fmt"
	"strings"

	"github.com/CHE3MZ/gitagger/internal/config"
	"github.com/CHE3MZ/gitagger/internal/git"
	"github.com/CHE3MZ/gitagger/internal/style"
)

// EffectiveRemote resolves the push remote: config or origin.
func EffectiveRemote(dir string) string {
	cfg, _, _ := config.Load(dir)
	if strings.TrimSpace(cfg.Remote) != "" {
		return cfg.Remote
	}
	return "origin"
}

// RunRemote sets the push remote to name, or with show prints the current
// remote URL instead. With verbose it shows all remotes.
func RunRemote(dir, name string, show, verbose bool) error {
	if err := git.EnsureAvailable(); err != nil {
		return GenericErr(err)
	}
	if show {
		if name != "" {
			return BadArgs("use either a name or --show, not both")
		}
		return showRemote(dir, EffectiveRemote(dir), verbose)
	}
	if strings.TrimSpace(name) == "" {
		return BadArgs("give a remote name or use --show — try `gitagger remote --help`")
	}
	url, err := git.RemoteURL(dir, name)
	if err != nil {
		return Generic("no remote %q for this project.", name)
	}
	path, err := config.SetRemote(dir, name)
	if err != nil {
		return Generic("couldn't save remote %q (%v)", name, err)
	}
	if verbose {
		fmt.Printf("%s %s %s\n", style.Green("remote is now"), style.White(name), style.Dim(url+" ("+path+")"))
	} else {
		fmt.Printf("%s %s\n", style.Green("remote is now"), style.White(name))
	}
	return nil
}

func showRemote(dir, remote string, verbose bool) error {
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
