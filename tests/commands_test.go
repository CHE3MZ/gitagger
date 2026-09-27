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
	f, _, err := cmd.ParseTagFlags([]string{"-n", "-d", "-f", "-v"})
	if err != nil {
		t.Fatal(err)
	}
	if !f.NoPush || !f.DryRun || !f.Force || !f.Verbose {
		t.Fatalf("short flags not set: %+v", f)
	}
	f, _, err = cmd.ParseTagFlags([]string{"--no-push", "--dry-run", "--force", "--verbose"})
	if err != nil {
		t.Fatal(err)
	}
	if !f.NoPush || !f.DryRun || !f.Force || !f.Verbose {
		t.Fatalf("long flags not set: %+v", f)
	}
	// Removed flags must error.
	for _, a := range [][]string{{"--json"}, {"--confirm"}, {"--remote", "x"}, {"-m", "x"}, {"--require-clean"}, {"--custom", "x"}, {"-y"}} {
		if _, _, err := cmd.ParseTagFlags(a); err == nil {
			t.Errorf("ParseTagFlags(%v) should fail (flag removed)", a)
		}
	}
}

func TestParseCommon(t *testing.T) {
	h, v, err := cmd.ParseCommon([]string{"-h", "-v"}, "list")
	if err != nil || !h || !v {
		t.Fatalf("ParseCommon = %v,%v,%v", h, v, err)
	}
	if _, _, err := cmd.ParseCommon([]string{"--limit", "5"}, "list"); err == nil {
		t.Errorf("ParseCommon should reject --limit")
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
	if err := cmd.RunRemote(dir, "origin", false); err != nil {
		t.Fatalf("remote should show URL: %v", err)
	}
	out, err := git.Remotes(dir)
	if err != nil || !strings.Contains(out, "https://example.com/repo.git") {
		t.Fatalf("Remotes = %q,%v", out, err)
	}
}

func TestRemoteMissing(t *testing.T) {
	dir := initRepo(t)
	if err := cmd.RunRemote(dir, "origin", false); err == nil {
		t.Fatalf("remote with no remote should fail")
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
