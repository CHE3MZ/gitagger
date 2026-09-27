// RunDoctor audits local vs remote tag health. Read-only, never deletes.
package cmd

import (
	"fmt"
	"strings"

	"github.com/CHE3MZ/gitagger/internal/detect"
	"github.com/CHE3MZ/gitagger/internal/git"
	"github.com/CHE3MZ/gitagger/internal/style"
)

// RunDoctor fetches remote tags and reports local-only, remote-only,
// malformed, and HEAD distance. With verbose it also shows the remote URL.
func RunDoctor(dir, remote string, verbose bool) error {
	if err := git.EnsureAvailable(); err != nil {
		return GenericErr(err)
	}
	if strings.TrimSpace(remote) == "" {
		remote = "origin"
	}
	tags, err := git.ListTags(dir)
	if err != nil {
		return GenericErr(err)
	}
	fmt.Println(style.Header("gitagger doctor"))
	fmt.Printf("local tags: %s\n", style.White(fmt.Sprint(len(tags))))

	var bad []string
	for _, t := range tags {
		if detect.Classify(t) == detect.Unknown {
			bad = append(bad, t)
		}
	}
	if len(bad) > 0 {
		fmt.Printf("%s %s\n", style.Red("odd tags:"), style.White(strings.Join(bad, ", ")))
		fmt.Println(style.Dim("these don't look like triple/double/date/sha — harmless, just ignored"))
	} else {
		fmt.Println(style.Green("tag styles look clean"))
	}

	// Remote compare (graceful when offline).
	remoteURL, err := git.RemoteURL(dir, remote)
	if err != nil {
		fmt.Println(style.Warn(fmt.Sprintf("no remote %q — nothing to compare", remote)))
		printHeadState(dir)
		return nil
	}
	if verbose {
		fmt.Println(style.Dim("remote: " + remoteURL))
	}
	if err := git.FetchTags(dir, remote); err != nil {
		fmt.Println(style.Warn(fmt.Sprintf("couldn't reach %q — skipping remote check", remote)))
		printHeadState(dir)
		return nil
	}
	fmt.Println(style.Green("remote reachable, tags fetched"))

	remoteTags, err := git.RemoteTags(dir, remote)
	if err != nil {
		fmt.Println(style.Warn("couldn't list remote tags — skipping compare"))
		printHeadState(dir)
		return nil
	}

	localSet := map[string]bool{}
	for _, t := range tags {
		localSet[t] = true
	}
	remoteSet := map[string]bool{}
	for _, t := range remoteTags {
		remoteSet[t] = true
	}
	var localOnly, remoteOnly []string
	for _, t := range tags {
		if !remoteSet[t] {
			localOnly = append(localOnly, t)
		}
	}
	for _, t := range remoteTags {
		if !localSet[t] {
			remoteOnly = append(remoteOnly, t)
		}
	}

	if len(localOnly) > 0 {
		fmt.Printf("local-only (%d): %s\n", len(localOnly), style.White(strings.Join(localOnly, ", ")))
	} else {
		fmt.Println(style.Dim("no local-only tags"))
	}
	if len(remoteOnly) > 0 {
		fmt.Printf("remote-only (%d): %s\n", len(remoteOnly), style.White(strings.Join(remoteOnly, ", ")))
	} else {
		fmt.Println(style.Dim("no remote-only tags"))
	}

	printHeadState(dir)
	printDistance(dir, tags)
	return nil
}

func printHeadState(dir string) {
	headTags, _ := git.TagsPointingAtHEAD(dir)
	if len(headTags) == 0 {
		fmt.Println(style.Dim("HEAD has no tag yet — `gitagger` will fix that"))
	} else {
		fmt.Printf("HEAD tagged: %s\n", style.White(strings.Join(headTags, ", ")))
	}
}

func printDistance(dir string, tags []string) {
	if len(tags) == 0 {
		if n, err := git.HeadAheadCount(dir, ""); err == nil {
			fmt.Println(style.Dim(fmt.Sprintf("HEAD distance: %d commit(s) since beginning (no tags yet)", n)))
		}
		return
	}
	prev := tags[len(tags)-1]
	n, err := git.HeadAheadCount(dir, prev)
	if err != nil {
		return
	}
	if n == 0 {
		fmt.Println(style.Dim("HEAD distance: 0 — HEAD is tagged, make a new commit or use -f"))
	} else {
		fmt.Println(style.Dim(fmt.Sprintf("HEAD distance: %d commit(s) since %s", n, prev)))
	}
}
