// Path tests: -p/--path flag, path: config key, and directory resolution.
// Temp dirs only — never the repo's own .git.
package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/CHE3MZ/gitagger/internal/cmd"
	"github.com/CHE3MZ/gitagger/internal/config"
	"github.com/CHE3MZ/gitagger/internal/git"
	"github.com/CHE3MZ/gitagger/internal/run"
)

func sameDir(t *testing.T, a, b string) {
	t.Helper()
	sa, err := os.Stat(a)
	if err != nil {
		t.Fatal(err)
	}
	sb, err := os.Stat(b)
	if err != nil {
		t.Fatal(err)
	}
	if !os.SameFile(sa, sb) {
		t.Fatalf("%q is not %q", a, b)
	}
}

func TestPathFlagForms(t *testing.T) {
	f, _, err := cmd.ParseTagFlags([]string{"-p", "somewhere"})
	if err != nil || f.Path != "somewhere" {
		t.Fatalf("-p = %q,%v", f.Path, err)
	}
	f, _, err = cmd.ParseTagFlags([]string{"--path=/tmp"})
	if err != nil || f.Path != "/tmp" {
		t.Fatalf("--path= = %q,%v", f.Path, err)
	}
	f, _, err = cmd.ParseTagFlags([]string{"--path", "elsewhere"})
	if err != nil || f.Path != "elsewhere" {
		t.Fatalf("--path = %q,%v", f.Path, err)
	}
	rf, name, err := cmd.ParseRemoteArgs([]string{"upstream", "--path", "elsewhere"})
	if err != nil || rf.Path != "elsewhere" || name != "upstream" {
		t.Fatalf("remote -p = %+v,%q,%v", rf, name, err)
	}
}

func TestResolveDirFlag(t *testing.T) {
	base := t.TempDir()
	sub := filepath.Join(base, "sub")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	got, err := cmd.ResolveDir(base, sub)
	if err != nil {
		t.Fatal(err)
	}
	sameDir(t, got, sub)
	got, err = cmd.ResolveDir(base, "sub")
	if err != nil {
		t.Fatal(err)
	}
	sameDir(t, got, sub)
	if _, err := cmd.ResolveDir(base, "nope"); err == nil {
		t.Fatalf("missing dir should fail")
	} else if cmd.CodeOf(err) != 3 {
		t.Fatalf("missing dir should exit 3, got %d", cmd.CodeOf(err))
	}
}

func TestResolveDirDefaultsCwd(t *testing.T) {
	dir := t.TempDir()
	got, err := cmd.ResolveDir(dir, "")
	if err != nil {
		t.Fatal(err)
	}
	sameDir(t, got, dir)
}

func TestResolveDirConfigPath(t *testing.T) {
	home := t.TempDir()
	work := t.TempDir()
	writeConfig(t, home, "path: "+work+"\n")
	got, err := cmd.ResolveDir(home, "")
	if err != nil {
		t.Fatal(err)
	}
	sameDir(t, got, work)
	// Flag beats config.
	got, err = cmd.ResolveDir(home, home)
	if err != nil {
		t.Fatal(err)
	}
	sameDir(t, got, home)
}

func TestConfigPathLoads(t *testing.T) {
	dir := t.TempDir()
	writeConfig(t, dir, "path: elsewhere\n")
	cfg, _, err := config.Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Path != "elsewhere" {
		t.Fatalf("cfg.Path = %q", cfg.Path)
	}
}

func TestInitTemplateHasNoPath(t *testing.T) {
	for _, line := range strings.Split(config.DefaultFileContent(), "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "path:") {
			t.Fatalf("init template should not generate path:, got %q", line)
		}
	}
}

func TestTagInOtherDir(t *testing.T) {
	home := t.TempDir()
	work := t.TempDir()
	mustGitTempInit(t, work)
	resolved, err := cmd.ResolveDir(home, work)
	if err != nil {
		t.Fatal(err)
	}
	o := run.Options{Dir: resolved, Scale: "patch", Pre: "stable", Remote: "origin", Push: false}
	plan, err := run.ComputePlan(o)
	if err != nil {
		t.Fatal(err)
	}
	if plan.Next != "v1.0.0" {
		t.Fatalf("next = %q, want v1.0.0", plan.Next)
	}
	if err := git.CreateTag(resolved, plan.Next, "", false); err != nil {
		t.Fatal(err)
	}
	tags, err := git.ListTags(work)
	if err != nil {
		t.Fatal(err)
	}
	if len(tags) != 1 || tags[0] != "v1.0.0" {
		t.Fatalf("tags = %v", tags)
	}
}
