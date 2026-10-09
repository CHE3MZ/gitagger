// Rollback tests: failing success hooks gate the push — the tag is
// removed locally (or restored when -f overwrote one) and never pushed.
package tests

import (
	"os"
	"path/filepath"
	"strings"
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

func TestTagFlagUnsafe(t *testing.T) {
	f, _, err := cmd.ParseTagFlags([]string{"-u"})
	if err != nil || !f.Unsafe {
		t.Fatalf("ParseTagFlags(-u) = %+v,%v", f, err)
	}
	f, _, err = cmd.ParseTagFlags([]string{"--unsafe", "-n"})
	if err != nil || !f.Unsafe || !f.NoPush {
		t.Fatalf("ParseTagFlags(--unsafe -n) = %+v,%v", f, err)
	}
}

func TestUnsafeConfigKey(t *testing.T) {
	dir := t.TempDir()
	writeConfig(t, dir, "unsafe: true\n")
	cfg, _, err := config.Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.Unsafe {
		t.Fatalf("cfg.Unsafe = false, want true")
	}
	o, err := cmd.BuildOptions(dir, cfg, "", cmd.TagFlags{})
	if err != nil {
		t.Fatal(err)
	}
	if !o.Unsafe {
		t.Fatalf("config unsafe should flow into options")
	}
	// Unsafe is opt-in: init must not generate it.
	if strings.Contains(config.DefaultFileContent(), "unsafe") {
		t.Errorf("init template should not contain unsafe")
	}
}

func TestUnsafePushesDespiteHookFailure(t *testing.T) {
	dir := initRepo(t)
	remoteDir := t.TempDir()
	mustGit(t, remoteDir, "init", "--bare", "-q")
	mustGit(t, dir, "remote", "add", "origin", remoteDir)
	o := tagOptions(dir, false)
	o.Unsafe = true
	if err := cmd.RunTag(o); err != nil {
		t.Fatalf("unsafe run should push anyway: %v", err)
	}
	if tags, _ := git.ListTags(dir); len(tags) != 1 || tags[0] != "v1.0.0" {
		t.Fatalf("local tags = %v, want [v1.0.0]", tags)
	}
	remoteTags, err := git.RemoteTags(dir, "origin")
	if err != nil {
		t.Fatal(err)
	}
	if len(remoteTags) != 1 || remoteTags[0] != "v1.0.0" {
		t.Fatalf("remote tags = %v, want pushed [v1.0.0]", remoteTags)
	}
}

func TestUnsafeBypassesHeadTagged(t *testing.T) {
	dir := initRepo(t)
	if err := git.CreateTag(dir, "v1.0.0", "", false); err != nil {
		t.Fatal(err)
	}
	o := tagOptions(dir, false)
	o.Push = false
	o.Unsafe = true
	o.Hooks = config.Hooks{}
	if err := cmd.RunTag(o); err != nil {
		t.Fatalf("unsafe should skip the HEAD-tagged abort: %v", err)
	}
	if tags, _ := git.ListTags(dir); len(tags) != 2 {
		t.Fatalf("tags = %v, want v1.0.0 plus the next tag", tags)
	}
}

func TestUnsafeBypassesRequireClean(t *testing.T) {
	dir := initRepo(t)
	if err := os.WriteFile(filepath.Join(dir, "f.txt"), []byte("dirty"), 0o644); err != nil {
		t.Fatal(err)
	}
	o := tagOptions(dir, false)
	o.Push = false
	o.RequireClean = true
	o.Hooks = config.Hooks{}
	if err := cmd.RunTag(o); err == nil {
		t.Fatalf("require_clean should abort on a dirty tree")
	}
	o.Unsafe = true
	if err := cmd.RunTag(o); err != nil {
		t.Fatalf("unsafe should skip the require_clean abort: %v", err)
	}
	if tags, _ := git.ListTags(dir); len(tags) != 1 {
		t.Fatalf("tags = %v, want one tag despite dirty tree", tags)
	}
}

func TestRollbackRestoresRemoteOverwrite(t *testing.T) {
	dir := initRepo(t)
	remoteDir := t.TempDir()
	mustGit(t, remoteDir, "init", "--bare", "-q")
	mustGit(t, dir, "remote", "add", "origin", remoteDir)
	if err := git.CreateTag(dir, "v1.0.0", "", false); err != nil {
		t.Fatal(err)
	}
	mustGit(t, dir, "push", "origin", "v1.0.0")
	oldSHA, err := git.FullSHA(dir)
	if err != nil {
		t.Fatal(err)
	}
	mustGit(t, dir, "commit", "--allow-empty", "-qm", "second")
	// Local tag gone (never fetched), remote kept: the forced run recreates
	// v1.0.0 on the new commit and force-pushes over the remote one, so a
	// hook failure must restore the remote side to the old commit.
	mustGit(t, dir, "tag", "-d", "v1.0.0")
	o := tagOptions(dir, true)
	if err := cmd.RunTag(o); err == nil {
		t.Fatalf("failing hook should fail the run")
	}
	if tags, _ := git.ListTags(dir); len(tags) != 0 {
		t.Fatalf("local tags = %v, want rolled back to none", tags)
	}
	out := mustGit(t, dir, "ls-remote", "origin", "refs/tags/v1.0.0")
	fields := strings.Fields(out)
	if len(fields) < 1 || fields[0] != oldSHA {
		t.Fatalf("remote tag = %q, want restored to %s", out, oldSHA)
	}
}
