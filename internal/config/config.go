// Package config loads .gitagger.yml (or .yaml / .gitagger).
// Precedence is flags > config > built-in defaults.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Resolved is config with defaults applied.
type Resolved struct {
	Scale        string
	Pre          string
	Format       string
	Remote       string
	Push         bool
	Force        bool
	RequireClean bool
	Verbose      bool
	Message      string
}

// Defaults: patch, stable, auto, origin, push on. Everything else off/empty.
func Defaults() Resolved {
	return Resolved{
		Scale: "patch", Pre: "stable", Format: "auto",
		Remote: "origin", Push: true,
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
	if v, ok := m["force"]; ok && v != "" {
		b, err := parseBool(v)
		if err != nil {
			return cfg, path, fmt.Errorf("bad force value %q in %s (want true/false)", v, path)
		}
		cfg.Force = b
	}
	if v, ok := m["verbose"]; ok && v != "" {
		b, err := parseBool(v)
		if err != nil {
			return cfg, path, fmt.Errorf("bad verbose value %q in %s (want true/false)", v, path)
		}
		cfg.Verbose = b
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
	case "auto", "triple", "double", "single", "date", "sha", "sha-num":
	default:
		return fmt.Errorf("bad format %q (want auto|triple|double|single|date|sha|sha-num)", c.Format)
	}
	if strings.TrimSpace(c.Remote) == "" {
		return fmt.Errorf("remote must not be empty")
	}
	return nil
}

// DefaultFileContent is what `gitagger init` writes.
// Every key is documented — this file is the whole schema.
func DefaultFileContent() string {
	return `# gitagger config — edit me, then just run ` + "`gitagger`" + `.
# Everything is optional. Flags beat config, config beats defaults.

# Tag size for triple/double/single styles.
# major | minor | patch
scale: patch

# Flavor appended to the tag. stable = no suffix.
# stable | rc | beta | build | nightly
pre: stable

# Tag style. auto follows your tag history.
# auto | triple | double | single | date | sha | sha-num
format: auto

# Push target. Must exist for auto-push to happen.
remote: origin

# Auto-push the new tag when the remote is reachable.
# true = push, false = keep the tag local (same as -n).
push: true

# Always overwrite clashing tags and force-push them.
# true = same as -f on every run. Keep false unless you mean it.
force: false

# Abort instead of tagging when the working tree is dirty.
require_clean: false

# Show detection details and the plan while working.
# true = same as -v on every run.
verbose: false

# Tag message. Empty = lightweight tag, set = annotated tag.
message: ""
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
