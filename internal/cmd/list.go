// RunList shows all tags, oldest -> newest.
package cmd

import (
	"fmt"

	"github.com/CHE3MZ/gitagger/internal/git"
	"github.com/CHE3MZ/gitagger/internal/style"
)

// RunList prints every tag. With verbose it also shows the remote URL.
func RunList(dir string, verbose bool) error {
	if err := git.EnsureAvailable(); err != nil {
		return GenericErr(err)
	}
	tags, err := git.ListTags(dir)
	if err != nil {
		return GenericErr(err)
	}
	if len(tags) == 0 {
		fmt.Println(style.Dim("no tags yet — run `gitagger` to make your first one"))
		return nil
	}
	for _, t := range tags {
		fmt.Println(style.White(t))
	}
	fmt.Println(style.Dim(fmt.Sprintf("%d tag(s), newest last", len(tags))))
	if verbose {
		if url, err := git.RemoteURL(dir, "origin"); err == nil {
			fmt.Println(style.Dim("remote: " + url))
		}
	}
	return nil
}
