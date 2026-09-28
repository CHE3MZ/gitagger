// Argument-gated hooks: named blocks run only with -a/--argument.
package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/CHE3MZ/gitagger/internal/cmd"
	"github.com/CHE3MZ/gitagger/internal/config"
	"github.com/CHE3MZ/gitagger/internal/run"
)

func TestArgumentFlagCollects(t *testing.T) {
	f, _, err := cmd.ParseTagFlags([]string{"-a", "release", "--argument=deploy"})
	if err != nil {
		t.Fatal(err)
	}
	if len(f.Arguments) != 2 || f.Arguments[0] != "release" || f.Arguments[1] != "deploy" {
		t.Fatalf("arguments = %v", f.Arguments)
	}
	o, err := cmd.BuildOptions(t.TempDir(), config.Defaults(), "", cmd.TagFlags{Arguments: []string{"release"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(o.Arguments) != 1 || o.Arguments[0] != "release" {
		t.Fatalf("options = %+v", o)
	}
}

func TestArgumentBadName(t *testing.T) {
	dir := t.TempDir()
	writeConfig(t, dir, "on:\n  success:\n    argument: \"a b\"\n    run: echo hi\n")
	if _, _, err := config.Load(dir); err == nil {
		t.Fatalf("spaced argument name should fail")
	}
}

func TestArgumentGating(t *testing.T) {
	needSh(t)
	mkHooks := func() config.Hooks {
		return config.Hooks{Success: []config.HookBlock{
			{Argument: "release", Shell: "sh", Run: []string{"echo rel > rel-ran"}},
			{Shell: "sh", Run: []string{"echo always > always-ran"}},
			{Argument: "other", Shell: "sh", Run: []string{"echo other > other-ran"}},
		}}
	}
	marker := func(dir, name string) bool {
		_, err := os.Stat(filepath.Join(dir, name))
		return err == nil
	}

	// With -a release: release + untagged run, other stays dormant.
	dir := initRepo(t)
	o := run.Options{Dir: dir, Scale: "patch", Pre: "stable", Remote: "origin", Push: false,
		Arguments: []string{"release"}, Hooks: mkHooks()}
	if err := cmd.RunTag(o); err != nil {
		t.Fatal(err)
	}
	if !marker(dir, "rel-ran") || !marker(dir, "always-ran") || marker(dir, "other-ran") {
		t.Fatalf("wrong blocks ran")
	}

	// Bare run: only untagged blocks run.
	dir = initRepo(t)
	o = run.Options{Dir: dir, Scale: "patch", Pre: "stable", Remote: "origin", Push: false,
		Hooks: mkHooks()}
	if err := cmd.RunTag(o); err != nil {
		t.Fatal(err)
	}
	if marker(dir, "rel-ran") || !marker(dir, "always-ran") || marker(dir, "other-ran") {
		t.Fatalf("named blocks should stay dormant without -a")
	}
}

func TestArgumentCaseInsensitive(t *testing.T) {
	needSh(t)
	dir := initRepo(t)
	o := run.Options{Dir: dir, Scale: "patch", Pre: "stable", Remote: "origin", Push: false,
		Arguments: []string{"release"},
		Hooks:     config.Hooks{Success: []config.HookBlock{{Argument: "Release", Shell: "sh", Run: []string{"echo x > ran"}}}}}
	if err := cmd.RunTag(o); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "ran")); err != nil {
		t.Fatalf("mixed-case argument should match: %v", err)
	}
}

func TestArgumentEnv(t *testing.T) {
	needSh(t)
	dir := initRepo(t)
	o := run.Options{Dir: dir, Scale: "patch", Pre: "stable", Remote: "origin", Push: false,
		Arguments: []string{"release"},
		Hooks:     config.Hooks{Success: []config.HookBlock{{Shell: "sh", Run: []string{"echo $GITAGGER_ARGUMENT > arg-was"}}}}}
	if err := cmd.RunTag(o); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(dir, "arg-was"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(raw)) != "release" {
		t.Fatalf("GITAGGER_ARGUMENT = %q", strings.TrimSpace(string(raw)))
	}
}
