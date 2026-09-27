// RunList shows tags oldest -> newest.
package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/CHE3MZ/gitagger/internal/git"
	"github.com/CHE3MZ/gitagger/internal/style"
)

// RunList prints up to limit tags (newest last), optionally as JSON.
func RunList(dir string, limit int, jsonOut bool) error {
	if jsonOut {
		style.SetEnabled(false)
	}
	if err := git.EnsureAvailable(); err != nil {
		return GenericErr(err)
	}
	tags, err := git.ListTags(dir)
	if err != nil {
		return GenericErr(err)
	}
	if len(tags) > limit {
		tags = tags[len(tags)-limit:]
	}
	if jsonOut {
		return json.NewEncoder(os.Stdout).Encode(tags)
	}
	if len(tags) == 0 {
		fmt.Println(style.Dim("no tags yet — run `gitagger` to make your first one"))
		return nil
	}
	for _, t := range tags {
		fmt.Println(style.White(t))
	}
	fmt.Println(style.Dim(fmt.Sprintf("%d tag(s), newest last", len(tags))))
	return nil
}
