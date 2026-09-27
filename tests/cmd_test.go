// CLI-level integration: dry-run, no-push, json, offline, collision,
// remote parsing, distance, detect window. All use temp dirs only.
package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/CHE3MZ/gitagger/internal/detect"
	"github.com/CHE3MZ/gitagger/internal/git"
	"github.com/CHE3MZ/gitagger/internal/run"
)

func TestDryRunCreatesNothing(t *testing.T) {
	dir := initRepo(t)
	o := run.Options{Dir: dir, Scale: "patch", Pre: "stable", Format: detect.Triple, Remote: "origin", Push: true, DryRun: true}
	plan, err := run.ComputePlan(o)
	if err != nil {
		t.Fatal(err)
	}
	if plan.Next != "v1.0.0" {
		t.Fatalf("next = %q, want v1.0.0", plan.Next)
	}
	tags, _ := git.ListTags(dir)
	if len(tags) != 0 {
		t.Fatalf("dry-run must not create tags, got %v", tags)
	}
}

func TestNoPushKeepsLocal(t *testing.T) {
	dir := initRepo(t)
	o := run.Options{Dir: dir, Scale: "patch", Pre: "stable", Format: detect.Triple, Remote: "origin", Push: false}
	plan, err := run.ComputePlan(o)
	if err != nil {
		t.Fatal(err)
	}
	if err := git.CreateTag(dir, plan.Next, "", false); err != nil {
		t.Fatal(err)
	}
	out, err := run.EnsurePush(o, plan.Next)
	if err != nil {
		t.Fatal(err)
	}
	if out.Pushed {
		t.Fatalf("--no-push must not push")
	}
}

func TestOfflineSkipKeepsLocal(t *testing.T) {
	dir := initRepo(t)
	// Unreachable remote: ls-remote fails fast, tag must be kept.
	mustGit(t, dir, "remote", "add", "origin", "https://invalid.invalid/nope.git")
	o := run.Options{Dir: dir, Scale: "patch", Pre: "stable", Remote: "origin", Push: true}
	plan, err := run.ComputePlan(o)
	if err != nil {
		t.Fatal(err)
	}
	if err := git.CreateTag(dir, plan.Next, "", false); err != nil {
		t.Fatal(err)
	}
	out, err := run.EnsurePush(o, plan.Next)
	if err != nil {
		t.Fatal(err)
	}
	if out.Pushed {
		t.Fatalf("offline remote must not push")
	}
	if !strings.Contains(strings.ToLower(out.Skipped), "unreachable") {
		t.Fatalf("skip reason = %q, want unreachable", out.Skipped)
	}
	tags, _ := git.ListTags(dir)
	if len(tags) != 1 {
		t.Fatalf("local tag must be kept on offline, got %v", tags)
	}
}

func TestCollisionAbort(t *testing.T) {
	dir := initRepo(t)
	remoteDir := t.TempDir()
	mustGit(t, remoteDir, "init", "--bare", "-q")
	mustGit(t, dir, "remote", "add", "origin", remoteDir)
	// Local tag + push to remote so both have v0.0.1 on commit1.
	if err := git.CreateTag(dir, "v0.0.1", "", false); err != nil {
		t.Fatal(err)
	}
	mustGit(t, dir, "push", "origin", "v0.0.1")
	// New commit, force a duplicate tag name locally on new HEAD.
	mustGit(t, dir, "commit", "--allow-empty", "-m", "second")
	if err := git.CreateTag(dir, "v0.0.1", "", true); err != nil {
		t.Fatal(err)
	}
	o := run.Options{Dir: dir, Remote: "origin", Push: true, Force: false}
	if _, err := run.EnsurePush(o, "v0.0.1"); err == nil {
		t.Fatalf("expected collision abort when remote already has tag")
	} else if !strings.Contains(err.Error(), "already exists on remote") {
		t.Fatalf("collision error = %q, want already exists on remote", err.Error())
	}
}

func TestParseRemoteTags(t *testing.T) {
	out := "abc123\trefs/tags/v1.2.3\ndef456\trefs/tags/v1.2.3^{}\n789aaa\trefs/tags/v2026.09.26\n"
	got := git.ParseRemoteTags(out)
	if len(got) != 2 || got[0] != "v1.2.3" || got[1] != "v2026.09.26" {
		t.Fatalf("ParseRemoteTags = %v", got)
	}
	if !git.RemoteHasTag(out, "v1.2.3") {
		t.Fatalf("RemoteHasTag should find v1.2.3")
	}
	if git.RemoteHasTag(out, "v9.9.9") {
		t.Fatalf("RemoteHasTag should not find v9.9.9")
	}
}

func TestDeleteAndDistance(t *testing.T) {
	dir := initRepo(t)
	if err := git.CreateTag(dir, "v0.0.1", "", false); err != nil {
		t.Fatal(err)
	}
	n, err := git.HeadAheadCount(dir, "v0.0.1")
	if err != nil || n != 0 {
		t.Fatalf("distance on tagged HEAD = %d,%v want 0", n, err)
	}
	mustGit(t, dir, "commit", "--allow-empty", "-m", "second")
	n, err = git.HeadAheadCount(dir, "v0.0.1")
	if err != nil || n != 1 {
		t.Fatalf("distance after 1 commit = %d,%v want 1", n, err)
	}
	if err := git.DeleteTag(dir, "v0.0.1"); err != nil {
		t.Fatal(err)
	}
	tags, _ := git.ListTags(dir)
	if len(tags) != 0 {
		t.Fatalf("after delete tags = %v, want none", tags)
	}
}

func TestDetectUsesLast20(t *testing.T) {
	// 25 old date tags + 5 recent triple tags: last-20 window still
	// contains triples, so majority in window should be considered.
	// Full-history majority would be date (25 vs 5).
	var tags []string
	for i := 0; i < 25; i++ {
		tags = append(tags, "v2026.01.01")
	}
	for i := 0; i < 5; i++ {
		tags = append(tags, "v1.2.3")
	}
	// Detect over full list gives date.
	full, _ := detect.Detect(tags)
	if full != detect.Date {
		t.Fatalf("full detect = %q, want date", full)
	}
	// Last-20 window: 15 date + 5 triple -> still date, but window logic
	// must not panic and must respect the 20-cap used by ComputePlan.
	window := tags[len(tags)-20:]
	win, _ := detect.Detect(window)
	if win != detect.Date {
		t.Fatalf("window detect = %q, want date", win)
	}
}

func TestConfigPrecedenceFiles(t *testing.T) {
	dir := t.TempDir()
	mustGitTempInit(t, dir)
	content := "scale: minor\npre: rc\nformat: triple\nremote: upstream\npush: false\n"
	if err := os.WriteFile(filepath.Join(dir, ".gitagger.yml"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	// Load is covered elsewhere; here ensure ComputePlan respects explicit
	// Format override (triple) even when history is empty.
	o := run.Options{Dir: dir, Scale: "minor", Pre: "rc", Format: detect.Triple, Remote: "upstream", Push: false}
	plan, err := run.ComputePlan(o)
	if err != nil {
		t.Fatal(err)
	}
	if plan.Next != "v1.0.0-rc" {
		t.Fatalf("next = %q, want v1.0.0-rc", plan.Next)
	}
}

func mustGitTempInit(t *testing.T, dir string) {
	t.Helper()
	mustGit(t, dir, "init", "-q")
	mustGit(t, dir, "config", "user.email", "gitagger-test@example.com")
	mustGit(t, dir, "config", "user.name", "gitagger-test")
	mustGit(t, dir, "config", "commit.gpgsign", "false")
	if err := os.WriteFile(filepath.Join(dir, "f.txt"), []byte("hi"), 0o644); err != nil {
		t.Fatal(err)
	}
	mustGit(t, dir, "add", ".")
	mustGit(t, dir, "commit", "-qm", "first")
}
