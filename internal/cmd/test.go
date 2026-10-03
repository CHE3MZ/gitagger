// RunTest exercises the full tag flow in a disposable clone.
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/CHE3MZ/gitagger/internal/config"
	"github.com/CHE3MZ/gitagger/internal/git"
	"github.com/CHE3MZ/gitagger/internal/style"
)

// RunTest clones dir into the OS temp dir, strips all remotes (and refuses
// to continue if any survive), adds an empty probe commit, and runs the
// real tag flow with pushing disabled. Hooks execute for real with
// GITAGGER_DRY_RUN=true so they can guard destructive halves.
// The sandbox is deleted afterwards unless keep is set, in which case its
// path is printed. Returns the sandbox path (exists only when kept).
func RunTest(dir string, keep, verbose bool) (string, error) {
	if err := git.EnsureAvailable(); err != nil {
		return "", GenericErr(err)
	}
	tmp, err := os.MkdirTemp("", "gitagger-test-")
	if err != nil {
		return "", Generic("couldn't make a sandbox (%v)", err)
	}
	// Best-effort cleanup on every path except an explicit keep.
	defer func() {
		if keep {
			return
		}
		if err := os.RemoveAll(tmp); err != nil {
			fmt.Println(style.Warn(fmt.Sprintf("couldn't delete sandbox %s (%v)", tmp, err)))
		}
	}()
	fmt.Println(style.Dim("sandbox: " + tmp))
	fmt.Println(style.Warn("hooks run for real in the sandbox (GITAGGER_DRY_RUN=true)"))
	// Clones carry committed files only: uncommitted changes (including
	// the config) stay behind. Say so instead of testing stale state.
	if st, err := git.StatusPorcelain(dir); err == nil && strings.TrimSpace(st) != "" {
		fmt.Println(style.Dim("uncommitted changes won't reach the sandbox — commit first for full fidelity"))
	}

	if err := git.Clone(dir, tmp); err != nil {
		return "", Generic("couldn't clone into sandbox (%v)", err)
	}
	names, err := git.RemoteNames(tmp)
	if err != nil {
		return "", GenericErr(err)
	}
	for _, n := range names {
		if err := git.RemoveRemote(tmp, n); err != nil {
			return "", Generic("couldn't strip remote %q from sandbox (%v)", n, err)
		}
	}
	if names, err = git.RemoteNames(tmp); err != nil || len(names) != 0 {
		return "", Generic("sandbox still has remotes — refusing to test")
	}
	if err := git.SetIdentity(tmp, "gitagger-test@example.com", "gitagger-test"); err != nil {
		return "", GenericErr(err)
	}
	if err := git.CommitEmpty(tmp, "gitagger test probe"); err != nil {
		return "", Generic("couldn't make a probe commit (%v)", err)
	}
	cfg, _, err := config.Load(tmp)
	if err != nil {
		return "", BadArgs("invalid config: %s", err.Error())
	}
	o, err := BuildOptions(tmp, cfg, "", TagFlags{Verbose: verbose})
	if err != nil {
		return "", err
	}
	o.Push = false
	o.TestMode = true
	if runErr := RunTag(o); runErr != nil {
		return "", runErr
	}
	if keep {
		fmt.Println(style.Dim("kept sandbox at " + tmp))
		return tmp, nil
	}
	return "", nil
}
