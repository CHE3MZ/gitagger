// Hooks tests: config loading/validation plus runner and RunTag wiring.
// Temp dirs only — never the repo's own .git.
package tests

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/CHE3MZ/gitagger/internal/cmd"
	"github.com/CHE3MZ/gitagger/internal/config"
	"github.com/CHE3MZ/gitagger/internal/git"
	"github.com/CHE3MZ/gitagger/internal/run"
)

func needSh(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("sh"); err != nil {
		t.Skip("sh not on PATH")
	}
}

func TestHooksLoadAllForms(t *testing.T) {
	dir := t.TempDir()
	writeConfig(t, dir, "scale: patch\n"+
		"on:\n"+
		"  start: echo starting\n"+
		"  success:\n"+
		"    - run: echo done\n"+
		"    - shell: bash\n"+
		"      os: linux\n"+
		"      run:\n"+
		"        - echo one\n"+
		"        - echo two\n"+
		"  failure:\n"+
		"    shell: pwsh\n"+
		"    run: echo oops\n"+
		"  finish:\n"+
		"    - shell: batch\n"+
		"      os: windows\n"+
		"      run: echo bye\n")
	cfg, _, err := config.Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Hooks.Start) != 1 || cfg.Hooks.Start[0].Shell != "sh" {
		t.Fatalf("start = %+v, want 1 sh block (defaulted)", cfg.Hooks.Start)
	}
	if len(cfg.Hooks.Success) != 2 || cfg.Hooks.Success[1].Shell != "bash" || cfg.Hooks.Success[1].OS != "linux" {
		t.Fatalf("success = %+v", cfg.Hooks.Success)
	}
	if len(cfg.Hooks.Failure) != 1 || cfg.Hooks.Failure[0].Shell != "pwsh" {
		t.Fatalf("failure = %+v", cfg.Hooks.Failure)
	}
	if len(cfg.Hooks.Finish) != 1 || cfg.Hooks.Finish[0].OS != "windows" {
		t.Fatalf("finish = %+v", cfg.Hooks.Finish)
	}
}

func TestHooksRejectBadShapes(t *testing.T) {
	cases := []struct{ name, content string }{
		{"unknown event", "on:\n  frobnicate: echo hi\n"},
		{"unknown shell", "on:\n  start:\n    shell: fish\n    run: echo hi\n"},
		{"empty run", "on:\n  start:\n    run: []\n"},
		{"blank command", "on:\n  start:\n    run: [\"\"]\n"},
		{"bad os", "on:\n  start:\n    os: plan9\n    run: echo hi\n"},
		{"bare string in list", "on:\n  start:\n    - echo hi\n"},
		{"top-level typo", "remtoe: origin\n"},
	}
	for _, c := range cases {
		dir := t.TempDir()
		writeConfig(t, dir, c.content)
		if _, _, err := config.Load(dir); err == nil {
			t.Errorf("%s should fail validation", c.name)
		}
	}
}

func TestHooksCaseInsensitiveKeys(t *testing.T) {
	dir := t.TempDir()
	writeConfig(t, dir, "Scale: minor\nOn:\n  Start: echo hi\n")
	cfg, _, err := config.Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Scale != "minor" || len(cfg.Hooks.Start) != 1 {
		t.Fatalf("cfg = %+v", cfg)
	}
}

func TestHooksBoolSpellings(t *testing.T) {
	dir := t.TempDir()
	writeConfig(t, dir, "push: yes\nforce: on\nverbose: 1\nrequire_clean: no\n")
	cfg, _, err := config.Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.Push || !cfg.Force || !cfg.Verbose || cfg.RequireClean {
		t.Fatalf("cfg = %+v", cfg)
	}
}

func TestInitTemplateLoads(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".gitagger.yml"), []byte(config.DefaultFileContent()), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, _, err := config.Load(dir)
	if err != nil {
		t.Fatalf("init template should load: %v", err)
	}
	if len(cfg.Hooks.Start)+len(cfg.Hooks.Success)+len(cfg.Hooks.Failure)+len(cfg.Hooks.Finish) != 0 {
		t.Fatalf("init template should have no active hooks: %+v", cfg.Hooks)
	}
}

func TestHooksSuccessMarker(t *testing.T) {
	needSh(t)
	dir := initRepo(t)
	o := run.Options{Dir: dir, Scale: "patch", Pre: "stable", Remote: "origin", Push: false,
		Hooks: config.Hooks{Success: []config.HookBlock{{Shell: "sh", Run: []string{"echo $GITAGGER_TAG > hook-tag"}}}}}
	if err := cmd.RunTag(o); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(dir, "hook-tag"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(raw)) != "v1.0.0" {
		t.Fatalf("hook saw tag %q, want v1.0.0", strings.TrimSpace(string(raw)))
	}
}

func TestHooksStartFailureAborts(t *testing.T) {
	needSh(t)
	dir := initRepo(t)
	o := run.Options{Dir: dir, Scale: "patch", Pre: "stable", Remote: "origin", Push: false,
		Hooks: config.Hooks{Start: []config.HookBlock{{Shell: "sh", Run: []string{"false"}}}}}
	if err := cmd.RunTag(o); err == nil {
		t.Fatalf("failing start hook should abort")
	}
	tags, err := git.ListTags(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(tags) != 0 {
		t.Fatalf("aborted run must create nothing, got %v", tags)
	}
}

func TestHooksFailureAndFinishRun(t *testing.T) {
	needSh(t)
	dir := initRepo(t)
	o := run.Options{Dir: dir, Scale: "patch", Pre: "stable", Remote: "origin", Push: false,
		Hooks: config.Hooks{
			Start:   []config.HookBlock{{Shell: "sh", Run: []string{"false"}}},
			Failure: []config.HookBlock{{Shell: "sh", Run: []string{"echo x > fail-ran"}}},
			Finish:  []config.HookBlock{{Shell: "sh", Run: []string{"echo x > finish-ran"}}},
		}}
	if err := cmd.RunTag(o); err == nil {
		t.Fatalf("expected abort")
	}
	for _, f := range []string{"fail-ran", "finish-ran"} {
		if _, err := os.Stat(filepath.Join(dir, f)); err != nil {
			t.Fatalf("%s marker missing: %v", f, err)
		}
	}
}

func TestHooksSkipOtherOS(t *testing.T) {
	needSh(t)
	other := "windows"
	if runtime.GOOS == "windows" {
		other = "linux"
	}
	var out strings.Builder
	err := cmd.RunHooks(t.TempDir(), "start",
		[]config.HookBlock{{Shell: "sh", OS: other, Run: []string{"echo nope"}}},
		map[string]string{}, true, &out, &out)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "skipped") {
		t.Fatalf("expected skip note, got %q", out.String())
	}
}
