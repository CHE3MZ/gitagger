// Sandbox tests: `gitagger test` clones to temp, strips remotes, runs
// the real tag flow, and deletes the clone. The real repo must never
// gain tags, files, or markers. Temp dirs only, as always.
package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/CHE3MZ/gitagger/internal/cmd"
	"github.com/CHE3MZ/gitagger/internal/git"
)

func TestParseTestArgs(t *testing.T) {
	f, err := cmd.ParseTestArgs([]string{"-k", "-v", "-p", "elsewhere"})
	if err != nil || !f.Keep || !f.Verbose || f.Path != "elsewhere" {
		t.Fatalf("ParseTestArgs = %+v,%v", f, err)
	}
	f, err = cmd.ParseTestArgs([]string{"--keep", "--verbose", "--path=/tmp"})
	if err != nil || !f.Keep || !f.Verbose || f.Path != "/tmp" {
		t.Fatalf("ParseTestArgs long = %+v,%v", f, err)
	}
	if _, err := cmd.ParseTestArgs([]string{"v1.0.0"}); err == nil {
		t.Errorf("positional tag should fail (test takes none)")
	}
	if _, err := cmd.ParseTestArgs([]string{"--force"}); err == nil {
		t.Errorf("--force is not a test flag and should fail")
	}
	if _, err := cmd.ParseTestArgs([]string{"-c"}); err == nil {
		t.Errorf("-c is not a test flag and should fail")
	}
}

func tagsOf(t *testing.T, dir string) []string {
	t.Helper()
	tags, err := git.ListTags(dir)
	if err != nil {
		t.Fatal(err)
	}
	return tags
}

func TestRunTestBasic(t *testing.T) {
	dir := initRepo(t)
	// HEAD already tagged: exercises the probe-commit path.
	if err := git.CreateTag(dir, "v1.0.0", "", false); err != nil {
		t.Fatal(err)
	}
	sandbox, err := cmd.RunTest(dir, false, false)
	if err != nil {
		t.Fatalf("test run should pass: %v", err)
	}
	if sandbox != "" {
		t.Fatalf("sandbox should be deleted, got path %q", sandbox)
	}
	if tags := tagsOf(t, dir); len(tags) != 1 || tags[0] != "v1.0.0" {
		t.Fatalf("real repo tags = %v, want untouched [v1.0.0]", tags)
	}
}

func TestRunTestHooks(t *testing.T) {
	dir := initRepo(t)
	writeConfig(t, dir, "push: false\non:\n  success:\n    - run: echo hi > hook-ran\n    - run: echo $GITAGGER_DRY_RUN > dry-marker\n    - run: echo $GITAGGER_TAG > tag-marker\n")
	// Configs travel with the clone: commit it like a real user would.
	mustGit(t, dir, "add", ".gitagger.yml")
	mustGit(t, dir, "commit", "-qm", "config")
	sandbox, err := cmd.RunTest(dir, true, false)
	if err != nil {
		t.Fatalf("test run should pass: %v", err)
	}
	defer func() { _ = os.RemoveAll(sandbox) }()
	if sandbox == "" {
		t.Fatalf("keep should return the sandbox path")
	}
	// Markers land in the sandbox, never in the real repo.
	if _, err := os.Stat(filepath.Join(dir, "hook-ran")); !os.IsNotExist(err) {
		t.Fatalf("hook marker leaked into the real repo")
	}
	if tags := tagsOf(t, dir); len(tags) != 0 {
		t.Fatalf("real repo tags = %v, want none", tags)
	}
	dry, err := os.ReadFile(filepath.Join(sandbox, "dry-marker"))
	if err != nil {
		t.Fatalf("dry marker missing from sandbox: %v", err)
	}
	if strings.TrimSpace(string(dry)) != "true" {
		t.Fatalf("GITAGGER_DRY_RUN = %q, want true", string(dry))
	}
	tag, err := os.ReadFile(filepath.Join(sandbox, "tag-marker"))
	if err != nil {
		t.Fatalf("tag marker missing from sandbox: %v", err)
	}
	sandboxTags := tagsOf(t, sandbox)
	if len(sandboxTags) != 1 || strings.TrimSpace(string(tag)) != sandboxTags[0] {
		t.Fatalf("tag marker = %q, sandbox tags = %v", string(tag), sandboxTags)
	}
}

func TestRunTestNoRemote(t *testing.T) {
	dir := initRepo(t)
	remoteDir := t.TempDir()
	mustGit(t, remoteDir, "init", "--bare", "-q")
	mustGit(t, dir, "remote", "add", "origin", remoteDir)
	if err := git.CreateTag(dir, "v1.0.0", "", false); err != nil {
		t.Fatal(err)
	}
	mustGit(t, dir, "push", "origin", "v1.0.0")
	if _, err := cmd.RunTest(dir, false, false); err != nil {
		t.Fatalf("test run should pass: %v", err)
	}
	// Neither local nor remote tags may change.
	if tags := tagsOf(t, dir); len(tags) != 1 {
		t.Fatalf("local tags = %v, want untouched", tags)
	}
	remoteTags, err := git.RemoteTags(dir, "origin")
	if err != nil {
		t.Fatal(err)
	}
	if len(remoteTags) != 1 || remoteTags[0] != "v1.0.0" {
		t.Fatalf("remote tags = %v, want untouched [v1.0.0]", remoteTags)
	}
}

func TestRunTestHookFailure(t *testing.T) {
	dir := initRepo(t)
	writeConfig(t, dir, "push: false\non:\n  success:\n    - run: exit 3\n")
	mustGit(t, dir, "add", ".gitagger.yml")
	mustGit(t, dir, "commit", "-qm", "config")
	_, err := cmd.RunTest(dir, false, false)
	if err == nil {
		t.Fatalf("failing hook should fail the test run")
	}
	if cmd.CodeOf(err) != 1 {
		t.Fatalf("hook failure should exit 1, got %d", cmd.CodeOf(err))
	}
	if tags := tagsOf(t, dir); len(tags) != 0 {
		t.Fatalf("real repo tags = %v, want none", tags)
	}
}

func TestRunTestKeepDeletes(t *testing.T) {
	dir := initRepo(t)
	sandbox, err := cmd.RunTest(dir, false, false)
	if err != nil {
		t.Fatal(err)
	}
	if sandbox != "" {
		t.Fatalf("without keep, path should be empty, got %q", sandbox)
	}
}

func TestTestHelp(t *testing.T) {
	if !cmd.IsCommand("test") {
		t.Errorf("IsCommand(test) = false")
	}
	cmd.CommandHelp("test") // must not crash
	if n := cmd.ExampleCount("test"); n == 0 || n > 3 {
		t.Errorf("test has %d examples, want 1-3", n)
	}
}
