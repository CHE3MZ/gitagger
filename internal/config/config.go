// Package config loads .gitagger.yml (or .yaml / .gitagger).
// Precedence is flags > config > built-in defaults.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Config mirrors the YAML file. Keep field names friendly.
type Config struct {
	Version      int    `yaml:"version"`
	Scale        string `yaml:"scale"`
	Pre          string `yaml:"pre"`
	Format       string `yaml:"format"`
	Custom       string `yaml:"custom"`
	Remote       string `yaml:"remote"`
	Push         *bool  `yaml:"push"`
	Confirm      *bool  `yaml:"confirm"`
	RequireClean *bool  `yaml:"require_clean"`
	Message      string `yaml:"message"`
}

// Resolved is config with defaults applied (no pointers).
type Resolved struct {
	Scale        string
	Pre          string
	Format       string
	Custom       string
	Remote       string
	Push         bool
	Confirm      bool
	RequireClean bool
	Message      string
}

// Defaults per plan: patch, stable, auto, push on, no prompt.
func Defaults() Resolved {
	return Resolved{
		Scale: "patch", Pre: "stable", Format: "auto",
		Remote: "origin", Push: true, Confirm: false,
	}
}

// Candidates in lookup order.
func candidates(dir string) []string {
	return []string{
		filepath.Join(dir, ".gitagger.yml"),
		filepath.Join(dir, ".gitagger.yaml"),
		filepath.Join(dir, ".gitagger"),
	}
}

// Find returns the first config file that exists, or "".
func Find(dir string) string {
	for _, c := range candidates(dir) {
		if st, err := os.Stat(c); err == nil && !st.IsDir() {
			return c
		}
	}
	return ""
}

// Load reads dir's config file (or defaults when missing).
// Minimal YAML parse on purpose: no external dep, small schema.
func Load(dir string) (Resolved, string, error) {
	cfg := Defaults()
	path := Find(dir)
	if path == "" {
		return cfg, "", nil
	}
	// #nosec G304 — path always comes from Find(), which only
	// returns our three known filenames (.gitagger.yml/.yaml/.gitagger).
	raw, err := os.ReadFile(path)
	if err != nil {
		return cfg, path, err
	}
	m := parseSimple(string(raw))
	if v, ok := m["scale"]; ok && v != "" {
		cfg.Scale = strings.ToLower(v)
	}
	if v, ok := m["pre"]; ok && v != "" {
		cfg.Pre = strings.ToLower(v)
	}
	if v, ok := m["format"]; ok && v != "" {
		cfg.Format = strings.ToLower(v)
	}
	if v, ok := m["custom"]; ok {
		cfg.Custom = v
	}
	if v, ok := m["remote"]; ok && v != "" {
		cfg.Remote = v
	}
	if v, ok := m["push"]; ok && v != "" {
		b, err := parseBool(v)
		if err != nil {
			return cfg, path, fmt.Errorf("bad push value %q in %s (want true/false)", v, path)
		}
		cfg.Push = b
	}
	if v, ok := m["confirm"]; ok && v != "" {
		b, err := parseBool(v)
		if err != nil {
			return cfg, path, fmt.Errorf("bad confirm value %q in %s (want true/false)", v, path)
		}
		cfg.Confirm = b
	}
	if v, ok := m["require_clean"]; ok && v != "" {
		b, err := parseBool(v)
		if err != nil {
			return cfg, path, fmt.Errorf("bad require_clean value %q in %s", v, path)
		}
		cfg.RequireClean = b
	}
	if v, ok := m["message"]; ok {
		cfg.Message = v
	}
	if err := Validate(cfg); err != nil {
		return cfg, path, err
	}
	return cfg, path, nil
}

// Validate checks enum values with a helpful message.
func Validate(c Resolved) error {
	switch c.Scale {
	case "major", "minor", "patch":
	default:
		return fmt.Errorf("bad scale %q (want major|minor|patch)", c.Scale)
	}
	switch c.Pre {
	case "stable", "rc", "beta", "build", "nightly":
	default:
		return fmt.Errorf("bad pre %q (want stable|rc|beta|build|nightly)", c.Pre)
	}
	switch c.Format {
	case "auto", "triple", "double", "single", "date", "sha", "sha-num", "custom":
	default:
		return fmt.Errorf("bad format %q (want auto|triple|double|single|date|sha|sha-num|custom)", c.Format)
	}
	if c.Format == "custom" && strings.TrimSpace(c.Custom) == "" {
		return fmt.Errorf("format is custom but custom template is empty")
	}
	if strings.TrimSpace(c.Remote) == "" {
		return fmt.Errorf("remote must not be empty")
	}
	return nil
}

// DefaultFileContent is what `gitagger init` writes.
func DefaultFileContent() string {
	return `# gitagger config — edit me, then just run ` + "`gitagger`" + `
version: 1
scale: patch        # major | minor | patch
pre: stable         # stable | rc | beta | build | nightly
format: auto        # auto | triple | double | single | date | sha | sha-num | custom
custom: ""          # template when format=custom (v2)
remote: origin
push: true          # auto-push when safe; false = local tags only
confirm: false      # true = ask before pushing
require_clean: false
message: ""         # annotated tag message; empty = lightweight tag
`
}

func parseBool(s string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "true", "yes", "y", "1", "on":
		return true, nil
	case "false", "no", "n", "0", "off":
		return false, nil
	default:
		return false, fmt.Errorf("not a bool")
	}
}

// parseSimple handles flat "key: value" YAML (comments + quotes ok).
func parseSimple(src string) map[string]string {
	m := map[string]string{}
	for _, line := range strings.Split(src, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		idx := strings.Index(line, ":")
		if idx < 0 {
			continue
		}
		k := strings.TrimSpace(line[:idx])
		v := strings.TrimSpace(line[idx+1:])
		v = strings.TrimSpace(strings.Trim(v, `"'`))
		// drop trailing comments
		if i := strings.Index(v, " #"); i >= 0 {
			v = strings.TrimSpace(v[:i])
		}
		m[strings.ToLower(k)] = v
	}
	return m
}
