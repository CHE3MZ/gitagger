// Config schema tests: force/verbose wiring, legacy-key tolerance,
// custom rejection, and init template coverage.
package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/CHE3MZ/gitagger/internal/cmd"
	"github.com/CHE3MZ/gitagger/internal/config"
)

func writeConfig(t *testing.T, dir, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, ".gitagger.yml"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestForceVerboseFromConfig(t *testing.T) {
	dir := t.TempDir()
	writeConfig(t, dir, "force: true\nverbose: true\n")
	cfg, _, err := config.Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.Force || !cfg.Verbose {
		t.Fatalf("cfg = %+v, want force+verbose true", cfg)
	}
	o, err := cmd.BuildOptions(dir, cfg, "", cmd.TagFlags{})
	if err != nil {
		t.Fatal(err)
	}
	if !o.Force || !o.Verbose {
		t.Fatalf("options = %+v, want force+verbose true", o)
	}
}

func TestFlagsBeatConfigForceVerbose(t *testing.T) {
	cfg := config.Defaults()
	o, err := cmd.BuildOptions(t.TempDir(), cfg, "", cmd.TagFlags{Force: true, Verbose: true})
	if err != nil {
		t.Fatal(err)
	}
	if !o.Force || !o.Verbose {
		t.Fatalf("flags should enable force+verbose: %+v", o)
	}
}

func TestLegacyKeysIgnored(t *testing.T) {
	dir := t.TempDir()
	writeConfig(t, dir, "version: 1\nconfirm: true\nformat: auto\n")
	cfg, _, err := config.Load(dir)
	if err != nil {
		t.Fatalf("legacy keys should be ignored, got: %v", err)
	}
	if cfg.Format != "auto" {
		t.Fatalf("cfg = %+v", cfg)
	}
}

func TestCustomFormatLoads(t *testing.T) {
	dir := t.TempDir()
	writeConfig(t, dir, "format: custom\ncustom: \"build-<SHA>-v<MAJOR>.<MINOR>.<PATCH>\"\n")
	cfg, _, err := config.Load(dir)
	if err != nil {
		t.Fatalf("valid custom config should load: %v", err)
	}
	if cfg.Custom != "build-<SHA>-v<MAJOR>.<MINOR>.<PATCH>" {
		t.Fatalf("cfg.Custom = %q", cfg.Custom)
	}
	if err := cmd.RunCheck(dir, false); err != nil {
		t.Fatalf("check should accept valid custom config: %v", err)
	}
}

func TestCustomEmptyRejected(t *testing.T) {
	dir := t.TempDir()
	writeConfig(t, dir, "format: custom\n")
	if _, _, err := config.Load(dir); err == nil {
		t.Fatalf("format: custom with empty template should fail")
	}
}

func TestCustomSetButUnusedRejected(t *testing.T) {
	dir := t.TempDir()
	writeConfig(t, dir, "format: triple\ncustom: \"v<MAJOR>.<MINOR>.<PATCH>\"\n")
	if _, _, err := config.Load(dir); err == nil {
		t.Fatalf("dormant custom template should fail validation")
	}
}

func TestBadForceValue(t *testing.T) {
	dir := t.TempDir()
	writeConfig(t, dir, "force: maybe\n")
	if _, _, err := config.Load(dir); err == nil {
		t.Fatalf("force: maybe should fail")
	}
}

func TestInitTemplateCoversSchema(t *testing.T) {
	tmpl := config.DefaultFileContent()
	for _, key := range []string{"scale:", "pre:", "format:", "custom:", "remote:", "push:", "force:", "require_clean:", "doctor:", "verbose:", "message:"} {
		if !strings.Contains(tmpl, key) {
			t.Errorf("init template missing %q", key)
		}
	}
	for _, dead := range []string{"version:", "confirm:"} {
		if strings.Contains(tmpl, dead) {
			t.Errorf("init template still has dead %q", dead)
		}
	}
}
