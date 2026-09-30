// RunRemove deletes a tag after confirmation.
package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/CHE3MZ/gitagger/internal/git"
	"github.com/CHE3MZ/gitagger/internal/style"
)

// Stdin feeds the remove confirmation prompt (os.Stdin; tests override it).
var Stdin io.Reader = os.Stdin

// AskConfirm reports whether r answers prompt with y/yes.
// Anything else (including empty) means no.
func AskConfirm(r io.Reader, prompt string) bool {
	fmt.Print(prompt + " ")
	line, _ := bufio.NewReader(r).ReadString('\n')
	switch strings.ToLower(strings.TrimSpace(line)) {
	case "y", "yes":
		return true
	default:
		return false
	}
}

// RunRemove deletes tag locally after confirmation, and from the remote
// when it is there too. confirm skips the prompt (same as -c).
// Offline or remote-less repos stay graceful: the local delete still happens.
func RunRemove(dir, tag string, confirm, verbose bool) error {
	if err := git.EnsureAvailable(); err != nil {
		return GenericErr(err)
	}
	tag = strings.TrimSpace(tag)
	if tag == "" {
		return BadArgs("give a tag name — try `gitagger remove --help`")
	}
	if _, err := git.InspectTag(dir, tag); err != nil {
		return Generic("no tag %q in this project.", tag)
	}
	if !confirm {
		if !AskConfirm(Stdin, "are you sure you want to remove this tag? [y/n]") {
			fmt.Println(style.Dim(fmt.Sprintf("kept tag %s (aborted)", tag)))
			return nil
		}
	}
	if err := git.DeleteTag(dir, tag); err != nil {
		return Generic("couldn't remove tag %s (%v)", tag, err)
	}
	fmt.Printf("%s %s\n", style.Green("removed tag"), style.BoldGreen(tag))
	remote := EffectiveRemote(dir)
	if _, err := git.RemoteURL(dir, remote); err != nil {
		if verbose {
			fmt.Println(style.Dim(fmt.Sprintf("no remote %q — removed locally only", remote)))
		}
		return nil
	}
	ls, err := git.LsRemoteTags(dir, remote)
	if err != nil {
		fmt.Println(style.Warn(fmt.Sprintf("remote %q unreachable — removed locally only", remote)))
		return nil
	}
	if !git.RemoteHasTag(ls, tag) {
		if verbose {
			fmt.Println(style.Dim(fmt.Sprintf("not on remote %q — removed locally only", remote)))
		}
		return nil
	}
	if err := git.DeleteRemoteTag(dir, remote, tag); err != nil {
		fmt.Println(style.Warn(fmt.Sprintf("removed locally, but couldn't remove %s from %q (%v)", tag, remote, err)))
		return nil
	}
	fmt.Printf("%s %s from %s\n", style.Green("removed"), style.BoldGreen(tag), style.White(remote))
	return nil
}
