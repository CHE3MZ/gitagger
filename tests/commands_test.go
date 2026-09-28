// Command-surface tests: check, remote, scale parsing, short flags.
// All use temp dirs only — never the repo's own .git.
package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/CHE3MZ/gitagger/internal/cmd"
	"github.com/CHE3MZ/gitagger/internal/config"
	"github.com/CHE3MZ/gitagger/internal/git"
)

func TestParseScale(t *testing.T) {
	cases := []struct {
		in   []string
		want string
	}{
		{nil, ""},
		{[]string{"patch"}, "patch"},
		{[]string{"MINOR"}, "minor"},
		{[]string{"major"}, "major"},
	}
	for _, c := range cases {
		got, err := cmd.ParseScale(c.in)
		if err != nil || got != c.want {
			t.Errorf("ParseScale(%v) = %q,%v want %q", c.in, got, err, c.want)
		}
	}
	if _, err := cmd.ParseScale([]string{"bogus"}); err == nil {
		t.Errorf("ParseScale(bogus) should fail")
	}
	if _, err := cmd.ParseScale([]string{"major", "minor"}); err == nil {
		t.Errorf("ParseScale(major minor) should fail")
	}
}

func TestTagFlagShorts(t *testing.T) {
	f, _, err := cmd.ParseTagFlags([]string{"-n", "-d", "-f", "-v", "-r"})
	if err != nil {
		t.Fatal(err)
	}
	if !f.NoPush || !f.DryRun || !f.Force || !f.Verbose || !f.RequireClean {
		t.Fatalf("short flags not set: %+v", f)
	}
	f, _, err = cmd.ParseTagFlags([]string{"--no-push", "--dry-run", "--force", "--verbose"})
	if err != nil {
		t.Fatal(err)
	}
	if !f.NoPush || !f.DryRun || !f.Force || !f.Verbose {
		t.Fatalf("long flags not set: %+v", f)
	}
	f, _, err = cmd.ParseTagFlags([]string{"-m", "hello", "--require-clean"})
	if err != nil {
		t.Fatal(err)
	}
	if f.Message != "hello" || !f.RequireClean {
		t.Fatalf("message/require-clean not set: %+v", f)
	}
	// Removed flags must error.
	for _, a := range [][]string{{"--json"}, {"--confirm"}, {"--remote", "x"}, {"--custom", "x"}, {"-y"}} {
		if _, _, err := cmd.ParseTagFlags(a); err == nil {
			t.Errorf("ParseTagFlags(%v) should fail (flag removed)", a)
		}
	}
}

func TestParseCommon(t *testing.T) {
	h, v, p, err := cmd.ParseCommon([]string{"-h", "-v", "-p", "elsewhere"}, "list")
	if err != nil || !h || !v || p != "elsewhere" {
		t.Fatalf("ParseCommon = %v,%v,%v,%v", h, v, p, err)
	}
	if _, _, _, err := cmd.ParseCommon([]string{"--limit", "5"}, "list"); err == nil {
		t.Errorf("ParseCommon should reject --limit")
	}
	if _, _, p, err := cmd.ParseCommon([]string{"--path=/tmp"}, "list"); err != nil || p != "/tmp" {
		t.Errorf("ParseCommon --path= = %q,%v", p, err)
	}
}

func TestCheckValid(t *testing.T) {
	dir := t.TempDir()
	content := "scale: minor\npre: rc\nformat: triple\nremote: upstream\npush: false\n"
	if err := os.WriteFile(filepath.Join(dir, ".gitagger.yml"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := cmd.RunCheck(dir, true); err != nil {
		t.Fatalf("valid config should pass: %v", err)
	}
}

func TestCheckInvalid(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".gitagger.yml"), []byte("scale: bogus\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := cmd.RunCheck(dir, false); err == nil {
		t.Fatalf("invalid config should fail")
	}
}

func TestCheckMissing(t *testing.T) {
	if err := cmd.RunCheck(t.TempDir(), false); err != nil {
		t.Fatalf("missing config should pass with defaults: %v", err)
	}
}

func TestRemoteShowsURL(t *testing.T) {
	dir := initRepo(t)
	mustGit(t, dir, "remote", "add", "origin", "https://example.com/repo.git")
	if err := cmd.RunRemote(dir, "", true, false); err != nil {
		t.Fatalf("remote --show should show URL: %v", err)
	}
	out, err := git.Remotes(dir)
	if err != nil || !strings.Contains(out, "https://example.com/repo.git") {
		t.Fatalf("Remotes = %q,%v", out, err)
	}
}

func TestRemoteMissing(t *testing.T) {
	dir := initRepo(t)
	if err := cmd.RunRemote(dir, "", true, false); err == nil {
		t.Fatalf("remote --show with no remote should fail")
	}
}

func TestRemoteSetPersists(t *testing.T) {
	dir := initRepo(t)
	mustGit(t, dir, "remote", "add", "origin", "https://example.com/a.git")
	mustGit(t, dir, "remote", "add", "upstream", "https://example.com/b.git")
	if err := cmd.RunRemote(dir, "upstream", false, false); err != nil {
		t.Fatalf("remote set should pass: %v", err)
	}
	cfg, _, err := config.Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Remote != "upstream" {
		t.Fatalf("cfg.Remote = %q, want upstream", cfg.Remote)
	}
	// Existing keys survive the rewrite.
	dir2 := t.TempDir()
	mustGitTempInit(t, dir2)
	writeConfig(t, dir2, "scale: minor\n")
	if err := cmd.RunRemote(dir2, "origin", false, false); err != nil {
		mustGit(t, dir2, "remote", "add", "origin", "https://example.com/a.git")
		if err := cmd.RunRemote(dir2, "origin", false, false); err != nil {
			t.Fatalf("remote set should pass: %v", err)
		}
	}
	cfg, _, err = config.Load(dir2)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Remote != "origin" || cfg.Scale != "minor" {
		t.Fatalf("cfg = %+v, want origin/minor preserved", cfg)
	}
}

func TestRemoteSetMissing(t *testing.T) {
	dir := initRepo(t)
	if err := cmd.RunRemote(dir, "nope", false, false); err == nil {
		t.Fatalf("setting a missing remote should fail")
	}
}

func TestRemoteBareNeedsNameOrShow(t *testing.T) {
	dir := initRepo(t)
	if err := cmd.RunRemote(dir, "", false, false); err == nil {
		t.Fatalf("bare remote should fail")
	}
	if _, _, err := cmd.ParseRemoteArgs([]string{"a", "b"}); err == nil {
		t.Errorf("two names should fail")
	}
	f, name, err := cmd.ParseRemoteArgs([]string{"upstream", "-s"})
	if err != nil || !f.Show || name != "upstream" {
		t.Fatalf("parse = %+v,%q,%v", f, name, err)
	}
}

func TestBuildOptionsPrecedence(t *testing.T) {
	cfg := config.Defaults()
	cfg.Scale = "minor"
	o, err := cmd.BuildOptions(t.TempDir(), cfg, "", cmd.TagFlags{})
	if err != nil {
		t.Fatal(err)
	}
	if o.Scale != "minor" {
		t.Fatalf("scale = %q, want minor from config", o.Scale)
	}
	o, err = cmd.BuildOptions(t.TempDir(), cfg, "major", cmd.TagFlags{NoPush: true, Pre: "rc"})
	if err != nil {
		t.Fatal(err)
	}
	if o.Scale != "major" || o.Push || o.Pre != "rc" {
		t.Fatalf("flags should win: %+v", o)
	}
	if _, err := cmd.BuildOptions(t.TempDir(), cfg, "", cmd.TagFlags{Format: "bogus"}); err == nil {
		t.Errorf("bad format should fail")
	}
}

func TestVersionString(t *testing.T) {
	if got := cmd.VersionString(); got != "gitagger is on the dev build version." {
		t.Fatalf("version = %q", got)
	}
}

func TestKnownCommands(t *testing.T) {
	for _, c := range []string{"init", "check", "list", "ls", "help", "doctor", "remote", "version", "patch", "minor", "major"} {
		if !cmd.IsCommand(c) {
			t.Errorf("IsCommand(%q) = false", c)
		}
		cmd.CommandHelp(c) // must not crash
	}
	if cmd.IsCommand("bogus") {
		t.Errorf("IsCommand(bogus) = true")
	}
}

func TestRunList(t *testing.T) {
	dir := initRepo(t)
	if err := cmd.RunList(dir, false); err != nil {
		t.Fatalf("empty list should pass: %v", err)
	}
	if err := git.CreateTag(dir, "v1.0.0", "", false); err != nil {
		t.Fatal(err)
	}
	if err := cmd.RunList(dir, true); err != nil {
		t.Fatalf("list should pass: %v", err)
	}
}

func TestRunInitWritesTemplate(t *testing.T) {
	dir := t.TempDir()
	if err := cmd.RunInit(dir, false, false, false); err != nil {
		t.Fatal(err)
	}
	cfg, path, err := config.Load(dir)
	if err != nil {
		t.Fatalf("fresh init should load: %v", err)
	}
	if path == "" {
		t.Fatalf("expected to find .gitagger.yml")
	}
	if cfg.Scale != "patch" || cfg.Remote != "origin" || !cfg.Push {
		t.Fatalf("fresh init should hold defaults: %+v", cfg)
	}
	if err := cmd.RunCheck(dir, false); err != nil {
		t.Fatalf("fresh init should validate: %v", err)
	}
}

func TestRunInitRefusesOverwrite(t *testing.T) {
	dir := t.TempDir()
	if err := cmd.RunInit(dir, false, false, false); err != nil {
		t.Fatal(err)
	}
	err := cmd.RunInit(dir, false, false, false)
	if err == nil {
		t.Fatalf("second init should refuse")
	}
	if cmd.CodeOf(err) != 2 {
		t.Fatalf("refusal should exit 2, got %d", cmd.CodeOf(err))
	}
	if err := cmd.RunInit(dir, true, false, false); err != nil {
		t.Fatalf("forced init should overwrite: %v", err)
	}
}

func TestRunInitClean(t *testing.T) {
	dir := t.TempDir()
	if err := cmd.RunInit(dir, false, false, true); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(dir, ".gitagger.yml"))
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range strings.Split(string(raw), "\n") {
		if trimmed := strings.TrimSpace(line); trimmed != "" && strings.HasPrefix(trimmed, "#") {
			t.Fatalf("clean config should have no comments, got %q", line)
		}
	}
	cfg, _, err := config.Load(dir)
	if err != nil {
		t.Fatalf("clean config should load: %v", err)
	}
	if cfg.Scale != "patch" || len(cfg.Hooks.Failure) != 2 {
		t.Fatalf("clean config should hold defaults + failure hooks: %+v", cfg)
	}
}

func TestRunDoctorNoRemote(t *testing.T) {
	dir := initRepo(t)
	if err := cmd.RunDoctor(dir, "origin", false); err != nil {
		t.Fatalf("doctor without remote should stay graceful: %v", err)
	}
	if err := git.CreateTag(dir, "hello", "", false); err != nil {
		t.Fatal(err)
	}
	if err := cmd.RunDoctor(dir, "origin", true); err != nil {
		t.Fatalf("doctor with odd tags should still pass: %v", err)
	}
}

func TestRunDoctorWithRemote(t *testing.T) {
	dir := initRepo(t)
	remoteDir := t.TempDir()
	mustGit(t, remoteDir, "init", "--bare", "-q")
	mustGit(t, dir, "remote", "add", "origin", remoteDir)
	if err := git.CreateTag(dir, "v1.0.0", "", false); err != nil {
		t.Fatal(err)
	}
	// Local-only tag, reachable remote: full compare path, still no error.
	if err := cmd.RunDoctor(dir, "origin", true); err != nil {
		t.Fatalf("doctor should pass: %v", err)
	}
}
