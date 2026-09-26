// Git integration tests: always use temp dirs + the user's git from PATH.
// They NEVER touch the repo's own .git — every helper takes an explicit dir.
// If git is missing from PATH, they fail with a friendly error (by design).
package tests

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/CHE3MZ/gitagger/internal/config"
	"github.com/CHE3MZ/gitagger/internal/git"
	"github.com/CHE3MZ/gitagger/internal/run"
)

func mustGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	full := append([]string{"-C", dir}, args...)
	out, err := exec.Command("git", full...).CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v (%s)", strings.Join(args, " "), err, strings.TrimSpace(string(out)))
	}
	return strings.TrimSpace(string(out))
}

// initRepo makes a throwaway repo with one commit.
// Identity is set repo-local only — never touches global/system config.
func initRepo(t *testing.T) string {
	t.Helper()
	if err := git.EnsureAvailable(); err != nil {
		t.Skipf("git not on PATH: %v", err)
	}
	dir := t.TempDir()
	mustGit(t, dir, "init")
	mustGit(t, dir, "config", "user.email", "gitagger-test@example.com")
	mustGit(t, dir, "config", "user.name", "gitagger-test")
	mustGit(t, dir, "config", "commit.gpgsign", "false")
	mustGit(t, dir, "config", "init.defaultBranch", "main")
	if err := os.WriteFile(filepath.Join(dir, "file.txt"), []byte("hi"), 0o644); err != nil {
		t.Fatal(err)
	}
	mustGit(t, dir, "add", ".")
	mustGit(t, dir, "commit", "-m", "first")
	return dir
}

func TestGitIsOnPATH(t *testing.T) {
	if err := git.EnsureAvailable(); err != nil {
		t.Fatalf("gitagger needs the user's git: %v", err)
	}
}

func TestEmptyRepoHasNoTags(t *testing.T) {
	dir := initRepo(t)
	tags, err := git.ListTags(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(tags) != 0 {
		t.Fatalf("fresh repo tags = %v, want none", tags)
	}
}

func TestCreateAndListTag(t *testing.T) {
	dir := initRepo(t)
	if err := git.CreateTag(dir, "v0.0.1", "", false); err != nil {
		t.Fatal(err)
	}
	tags, err := git.ListTags(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(tags) != 1 || tags[0] != "v0.0.1" {
		t.Fatalf("tags = %v, want [v0.0.1]", tags)
	}
}

func TestNoRemoteKeepsTagLocally(t *testing.T) {
	dir := initRepo(t)
	o := run.Options{Dir: dir, Scale: "patch", Pre: "stable", Remote: "origin", Push: true}
	plan, err := run.ComputePlan(o)
	if err != nil {
		t.Fatal(err)
	}
	if plan.Next != "v0.0.1" {
		t.Fatalf("next = %q, want v0.0.1", plan.Next)
	}
	if err := git.CreateTag(dir, plan.Next, "", false); err != nil {
		t.Fatal(err)
	}
	out, err := run.EnsurePush(o, plan.Next)
	if err != nil {
		t.Fatal(err)
	}
	if out.Pushed {
		t.Fatalf("pushed with no remote — must skip gracefully")
	}
	if !strings.Contains(out.Skipped, "no remote") {
		t.Fatalf("skip reason = %q, want mention of no remote", out.Skipped)
	}
}

func TestHeadAlreadyTaggedAborts(t *testing.T) {
	dir := initRepo(t)
	if err := git.CreateTag(dir, "v0.0.1", "", false); err != nil {
		t.Fatal(err)
	}
	o := run.Options{Dir: dir}
	if err := run.CheckHeadTagged(o, "v0.0.1"); err == nil {
		t.Fatalf("expected abort when HEAD already has v0.0.1")
	}
	if err := run.CheckHeadTagged(o, "v0.0.2"); err != nil {
		t.Fatalf("different tag should be fine: %v", err)
	}
}

func TestTempReposNeverTouchRootGit(t *testing.T) {
	dir := initRepo(t)
	if err := git.CreateTag(dir, "v9.9.9-temp-only", "", false); err != nil {
		t.Fatal(err)
	}
	// Root repo must not see the temp tag.
	rootTags, err := git.ListTags(".")
	if err != nil {
		t.Skipf("can't list root tags here (fine in CI subdir runs): %v", err)
	}
	for _, tg := range rootTags {
		if tg == "v9.9.9-temp-only" {
			t.Fatalf("temp tag leaked into root .git!")
		}
	}
}

func TestConfigFileLoads(t *testing.T) {
	dir := t.TempDir()
	content := "scale: minor\npre: rc\nformat: date\nremote: upstream\npush: false\nconfirm: true\n"
	if err := os.WriteFile(filepath.Join(dir, ".gitagger.yml"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, path, err := config.Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if path == "" {
		t.Fatalf("expected to find .gitagger.yml")
	}
	if cfg.Scale != "minor" || cfg.Pre != "rc" || cfg.Format != "date" || cfg.Remote != "upstream" {
		t.Fatalf("cfg = %+v, wrong values", cfg)
	}
	if cfg.Push || !cfg.Confirm {
		t.Fatalf("cfg push/confirm = %v/%v, want false/true", cfg.Push, cfg.Confirm)
	}
}
