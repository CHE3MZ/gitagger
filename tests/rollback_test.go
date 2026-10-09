// Rollback tests: failing success hooks gate the push — the tag is
// removed locally (or restored when -f overwrote one) and never pushed.
package tests

import (
	"testing"

	"github.com/CHE3MZ/gitagger/internal/cmd"
	"github.com/CHE3MZ/gitagger/internal/config"
	"github.com/CHE3MZ/gitagger/internal/detect"
	"github.com/CHE3MZ/gitagger/internal/git"
	"github.com/CHE3MZ/gitagger/internal/run"
)

func failingHooks() config.Hooks {
	return config.Hooks{
		Success: []config.HookBlock{{Shell: "sh", Run: []string{"false"}}},
	}
}

func tagOptions(dir string, force bool) run.Options {
	return run.Options{
		Dir: dir, Scale: "patch", Pre: "stable",
		Remote: "origin", Push: true, Force: force,
		Hooks: failingHooks(),
	}
}

func TestSuccessHookFailureRollsBack(t *testing.T) {
	dir := initRepo(t)
	remoteDir := t.TempDir()
	mustGit(t, remoteDir, "init", "--bare", "-q")
	mustGit(t, dir, "remote", "add", "origin", remoteDir)
	err := cmd.RunTag(tagOptions(dir, false))
	if err == nil {
		t.Fatalf("failing hook should fail the run")
	}
	if cmd.CodeOf(err) != 1 {
		t.Fatalf("hook failure should exit 1, got %d", cmd.CodeOf(err))
	}
	// Rolled back locally and never pushed.
	if tags, _ := git.ListTags(dir); len(tags) != 0 {
		t.Fatalf("local tags = %v, want rolled back to none", tags)
	}
	remoteTags, err := git.RemoteTags(dir, "origin")
	if err != nil {
		t.Fatal(err)
	}
	if len(remoteTags) != 0 {
		t.Fatalf("remote tags = %v, want nothing pushed", remoteTags)
	}
}

func TestRollbackRestoresOverwrite(t *testing.T) {
	dir := initRepo(t)
	remoteDir := t.TempDir()
	mustGit(t, remoteDir, "init", "--bare", "-q")
	mustGit(t, dir, "remote", "add", "origin", remoteDir)
	// Same HEAD + sha format always renders the same tag: with -f the
	// run overwrites it, so a hook failure must restore, not delete.
	short, err := git.ShortSHA(dir)
	if err != nil {
		t.Fatal(err)
	}
	name := "master-" + short
	if err := git.CreateTag(dir, name, "", false); err != nil {
		t.Fatal(err)
	}
	before, err := git.InspectTag(dir, name)
	if err != nil {
		t.Fatal(err)
	}
	o := tagOptions(dir, true)
	o.Format = detect.SHA
	if err := cmd.RunTag(o); err == nil {
		t.Fatalf("failing hook should fail the run")
	}
	after, err := git.InspectTag(dir, name)
	if err != nil {
		t.Fatalf("overwritten tag should be restored, got %v", err)
	}
	if after.Commit != before.Commit || after.Kind != before.Kind {
		t.Fatalf("tag restored to %+v, want %+v", after, before)
	}
	remoteTags, err := git.RemoteTags(dir, "origin")
	if err != nil {
		t.Fatal(err)
	}
	if len(remoteTags) != 0 {
		t.Fatalf("remote tags = %v, want nothing pushed", remoteTags)
	}
}
