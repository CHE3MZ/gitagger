// Safety tests: push moves tags only, require_clean blocks dirty trees,
// doctor:true fails early on collision, messages make annotated tags.
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

func tagType(t *testing.T, dir, tag string) string {
	t.Helper()
	out := mustGit(t, dir, "cat-file", "-t", tag)
	return out
}

func TestPushOnlyPushesTag(t *testing.T) {
	dir := initRepo(t)
	remoteDir := t.TempDir()
	mustGit(t, remoteDir, "init", "--bare", "-q")
	mustGit(t, dir, "remote", "add", "origin", remoteDir)
	mustGit(t, dir, "commit", "--allow-empty", "-m", "second")
	if err := git.CreateTag(dir, "v1.0.0", "", false); err != nil {
		t.Fatal(err)
	}
	o := run.Options{Dir: dir, Remote: "origin", Push: true}
	out, err := run.EnsurePush(o, "v1.0.0")
	if err != nil {
		t.Fatal(err)
	}
	if !out.Pushed {
		t.Fatalf("expected push to succeed")
	}
	refs := mustGit(t, dir, "ls-remote", remoteDir)
	if !strings.Contains(refs, "refs/tags/v1.0.0") {
		t.Fatalf("remote missing tag:\n%s", refs)
	}
	if strings.Contains(refs, "refs/heads/") {
		t.Fatalf("branches must never push, got:\n%s", refs)
	}
}

func TestRequireCleanBlocks(t *testing.T) {
	dir := initRepo(t)
	if err := os.WriteFile(filepath.Join(dir, "dirty.txt"), []byte("dirty"), 0o644); err != nil {
		t.Fatal(err)
	}
	o := run.Options{Dir: dir, Scale: "patch", Pre: "stable", RequireClean: true}
	if _, err := run.ComputePlan(o); err == nil {
		t.Fatalf("dirty tree with require_clean should fail")
	}
}

func TestRequireCleanFlagWires(t *testing.T) {
	cfg := config.Defaults()
	o, err := cmd.BuildOptions(t.TempDir(), cfg, "", cmd.TagFlags{RequireClean: true})
	if err != nil {
		t.Fatal(err)
	}
	if !o.RequireClean {
		t.Fatalf("flag should enable require_clean: %+v", o)
	}
}

func TestMessageMakesAnnotatedTag(t *testing.T) {
	dir := initRepo(t)
	o := run.Options{Dir: dir, Scale: "patch", Pre: "stable", Format: detect.Triple, Remote: "origin", Push: false, Message: "first release"}
	if err := cmd.RunTag(o); err != nil {
		t.Fatal(err)
	}
	if got := tagType(t, dir, "v1.0.0"); got != "tag" {
		t.Fatalf("tag type = %q, want annotated tag", got)
	}
}

func TestNoMessageMakesLightweightTag(t *testing.T) {
	dir := initRepo(t)
	o := run.Options{Dir: dir, Scale: "patch", Pre: "stable", Format: detect.Triple, Remote: "origin", Push: false}
	if err := cmd.RunTag(o); err != nil {
		t.Fatal(err)
	}
	if got := tagType(t, dir, "v1.0.0"); got != "commit" {
		t.Fatalf("tag type = %q, want lightweight commit tag", got)
	}
}

func TestDoctorGateBlocksBeforeCreating(t *testing.T) {
	dir := initRepo(t)
	remoteDir := t.TempDir()
	mustGit(t, remoteDir, "init", "--bare", "-q")
	mustGit(t, dir, "remote", "add", "origin", remoteDir)
	mustGit(t, dir, "tag", "v1.0.0")
	// Remote owns the NEXT tag: create it here, push, drop locally.
	mustGit(t, dir, "tag", "v1.0.1")
	mustGit(t, dir, "push", "origin", "v1.0.1")
	mustGit(t, dir, "tag", "-d", "v1.0.1")
	mustGit(t, dir, "commit", "--allow-empty", "-m", "second")
	o := run.Options{Dir: dir, Scale: "patch", Pre: "stable", Remote: "origin", Push: true, Doctor: true}
	if err := cmd.RunTag(o); err == nil {
		t.Fatalf("doctor gate should abort on remote collision")
	}
	tags, err := git.ListTags(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, tg := range tags {
		if tg == "v1.0.1" {
			t.Fatalf("gate must abort before creating, got %v", tags)
		}
	}
}

func TestDoctorGateSkipsOffline(t *testing.T) {
	dir := initRepo(t)
	mustGit(t, dir, "remote", "add", "origin", "https://invalid.invalid/nope.git")
	o := run.Options{Dir: dir, Scale: "patch", Pre: "stable", Remote: "origin", Push: false, Doctor: true, DryRun: true}
	if err := cmd.RunTag(o); err != nil {
		t.Fatalf("offline + doctor must stay graceful: %v", err)
	}
}
