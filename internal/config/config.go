// Package config loads .gitagger.yml (or .yaml / .gitagger).
// Precedence is flags > config > built-in defaults.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/CHE3MZ/gitagger/internal/next"
	"gopkg.in/yaml.v3"
)

// Resolved is config with defaults applied.
type Resolved struct {
	Scale        string
	Pre          string
	Format       string
	Custom       string
	Remote       string
	Push         bool
	Force        bool
	RequireClean bool
	Doctor       bool
	Verbose      bool
	Unsafe       bool
	Message      string
	Hooks        Hooks
	Path         string
}

// Defaults: patch, stable, auto, origin, push on. Everything else off/empty.
func Defaults() Resolved {
	return Resolved{
		Scale: "patch", Pre: "stable", Format: "auto",
		Remote: "origin", Push: true,
	}
}

// HookBlock is one shell block: optional argument/shell/os plus commands.
// Shell defaults to sh, os defaults to all. A block with an argument runs
// only when that argument is given (-a/--argument).
type HookBlock struct {
	Argument string     `yaml:"argument"`
	Shell    string     `yaml:"shell"`
	OS       string     `yaml:"os"`
	Run      StringList `yaml:"run"`
}

// HookList is one block or a list of blocks.
type HookList []HookBlock

// UnmarshalYAML accepts `event: echo hi`, `event: {shell, run}` or a list
// of blocks. Bare strings inside a list are rejected on purpose: write
// `- run: echo hi` so intent is never ambiguous.
func (h *HookList) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind == yaml.ScalarNode {
		*h = []HookBlock{{Run: StringList{value.Value}}}
		return nil
	}
	if value.Kind == yaml.MappingNode {
		var b HookBlock
		if err := value.Decode(&b); err != nil {
			return err
		}
		*h = []HookBlock{b}
		return nil
	}
	if value.Kind != yaml.SequenceNode {
		return fmt.Errorf("want a block or list of blocks")
	}
	bs := make([]HookBlock, 0, len(value.Content))
	for _, item := range value.Content {
		if item.Kind != yaml.MappingNode {
			return fmt.Errorf("want a block with run: (got %q)", item.Value)
		}
		var b HookBlock
		if err := item.Decode(&b); err != nil {
			return err
		}
		bs = append(bs, b)
	}
	*h = bs
	return nil
}

// StringList is one string or a list of strings.
type StringList []string

// UnmarshalYAML accepts `run: echo hi` or a list.
func (s *StringList) UnmarshalYAML(value *yaml.Node) error {
	var single string
	if err := value.Decode(&single); err == nil {
		*s = []string{single}
		return nil
	}
	var multi []string
	if err := value.Decode(&multi); err == nil {
		*s = multi
		return nil
	}
	return fmt.Errorf("want a string or list of strings")
}

// Hooks holds lifecycle hooks. finish runs always (success and failure).
type Hooks struct {
	Start   HookList       `yaml:"start"`
	Success HookList       `yaml:"success"`
	Failure HookList       `yaml:"failure"`
	Finish  HookList       `yaml:"finish"`
	Extra   map[string]any `yaml:",inline"`
}

// YBool accepts true/false plus common spellings (compat with old loader).
type YBool bool

// UnmarshalYAML accepts true/false/yes/no/on/off/1/0 in any case.
func (b *YBool) UnmarshalYAML(value *yaml.Node) error {
	switch strings.ToLower(value.Value) {
	case "true", "yes", "y", "1", "on":
		*b = true
		return nil
	case "false", "no", "n", "0", "off":
		*b = false
		return nil
	default:
		return fmt.Errorf("bad bool value %q (want true/false)", value.Value)
	}
}

// fileConfig mirrors the YAML file. Unknown top-level keys land in Extra.
type fileConfig struct {
	Scale        string         `yaml:"scale"`
	Pre          string         `yaml:"pre"`
	Format       string         `yaml:"format"`
	Custom       string         `yaml:"custom"`
	Remote       string         `yaml:"remote"`
	Push         *YBool         `yaml:"push"`
	Force        *YBool         `yaml:"force"`
	RequireClean *YBool         `yaml:"require_clean"`
	Doctor       *YBool         `yaml:"doctor"`
	Verbose      *YBool         `yaml:"verbose"`
	Unsafe       *YBool         `yaml:"unsafe"`
	Message      string         `yaml:"message"`
	On           Hooks          `yaml:"on"`
	Path         string         `yaml:"path"`
	Extra        map[string]any `yaml:",inline"`
}

// legacyKeys are tolerated but ignored (removed features).
var legacyKeys = map[string]bool{"version": true, "confirm": true}

// validShells are the only shell names hooks accept.
var validShells = map[string]bool{
	"sh": true, "bash": true, "pwsh": true, "batch": true,
}

// validHookOS limits blocks to platforms.
var validHookOS = map[string]bool{"linux": true, "macos": true, "windows": true}

// splitOS splits "macos, linux" into selectors. Empty means all platforms.
func splitOS(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		if p = strings.ToLower(strings.TrimSpace(p)); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// validArgumentName checks argument identifiers: letters, digits, _ and -.
func validArgumentName(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '_' || r == '-' {
			continue
		}
		return false
	}
	return true
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

// lowercaseKeys lowercases every mapping key in place so files are
// case-insensitive like the old loader. Values are never touched.
func lowercaseKeys(n *yaml.Node) {
	switch n.Kind {
	case yaml.MappingNode:
		for i := 0; i+1 < len(n.Content); i += 2 {
			if n.Content[i].Kind == yaml.ScalarNode {
				n.Content[i].Value = strings.ToLower(n.Content[i].Value)
			}
			lowercaseKeys(n.Content[i+1])
		}
	case yaml.DocumentNode, yaml.SequenceNode:
		for _, c := range n.Content {
			lowercaseKeys(c)
		}
	}
}

// Load reads dir's config file (or defaults when missing).
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
	var doc yaml.Node
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		return cfg, path, fmt.Errorf("bad yaml in %s (%v)", path, err)
	}
	if len(doc.Content) == 0 {
		return cfg, path, nil
	}
	lowercaseKeys(doc.Content[0])
	var fc fileConfig
	if err := doc.Content[0].Decode(&fc); err != nil {
		return cfg, path, fmt.Errorf("bad config in %s (%v)", path, err)
	}
	for k := range fc.Extra {
		if !legacyKeys[k] {
			return cfg, path, fmt.Errorf("unknown config key %q in %s", k, path)
		}
	}
	if fc.Scale != "" {
		cfg.Scale = strings.ToLower(fc.Scale)
	}
	if fc.Pre != "" {
		cfg.Pre = strings.ToLower(fc.Pre)
	}
	if fc.Format != "" {
		cfg.Format = strings.ToLower(fc.Format)
	}
	cfg.Custom = fc.Custom
	if fc.Remote != "" {
		cfg.Remote = fc.Remote
	}
	if fc.Push != nil {
		cfg.Push = bool(*fc.Push)
	}
	if fc.Force != nil {
		cfg.Force = bool(*fc.Force)
	}
	if fc.RequireClean != nil {
		cfg.RequireClean = bool(*fc.RequireClean)
	}
	if fc.Doctor != nil {
		cfg.Doctor = bool(*fc.Doctor)
	}
	if fc.Verbose != nil {
		cfg.Verbose = bool(*fc.Verbose)
	}
	if fc.Unsafe != nil {
		cfg.Unsafe = bool(*fc.Unsafe)
	}
	cfg.Message = fc.Message
	cfg.Path = fc.Path
	cfg.Hooks = fc.On
	normalizeHooks(&cfg.Hooks)
	if err := Validate(cfg); err != nil {
		return cfg, path, err
	}
	return cfg, path, nil
}

// normalizeHooks lowercases shell/os and defaults empty shells to sh.
func normalizeHooks(h *Hooks) {
	for _, list := range []*HookList{&h.Start, &h.Success, &h.Failure, &h.Finish} {
		for i := range *list {
			(*list)[i].Shell = strings.ToLower(strings.TrimSpace((*list)[i].Shell))
			if (*list)[i].Shell == "" {
				(*list)[i].Shell = "sh"
			}
			(*list)[i].OS = strings.ToLower(strings.TrimSpace((*list)[i].OS))
		}
	}
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
	if c.Format == "custom" {
		if err := next.ValidateTemplate(c.Custom); err != nil {
			return err
		}
	} else if strings.TrimSpace(c.Custom) != "" {
		return fmt.Errorf("custom template is set but format is not custom")
	}
	if strings.TrimSpace(c.Remote) == "" {
		return fmt.Errorf("remote must not be empty")
	}
	return validateHooks(c.Hooks)
}

// validateHooks checks events, shells, platforms, and commands.
func validateHooks(h Hooks) error {
	if len(h.Extra) > 0 {
		keys := make([]string, 0, len(h.Extra))
		for k := range h.Extra {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		return fmt.Errorf("unknown hook event %q (want start|success|failure|finish)", strings.Join(keys, ", "))
	}
	events := []struct {
		name   string
		blocks HookList
	}{
		{"start", h.Start}, {"success", h.Success},
		{"failure", h.Failure}, {"finish", h.Finish},
	}
	for _, ev := range events {
		for _, b := range ev.blocks {
			if b.Argument != "" && !validArgumentName(b.Argument) {
				return fmt.Errorf("bad argument %q in %s hook (want [a-zA-Z0-9_-]+)", b.Argument, ev.name)
			}
			if !validShells[b.Shell] {
				return fmt.Errorf("bad shell %q in %s hook (want sh|bash|pwsh|batch)", b.Shell, ev.name)
			}
			for _, o := range splitOS(b.OS) {
				if !validHookOS[o] {
					return fmt.Errorf("bad os %q in %s hook (want linux|macos|windows)", b.OS, ev.name)
				}
			}
			if len(b.Run) == 0 {
				return fmt.Errorf("%s hook block has no run commands", ev.name)
			}
			for _, c := range b.Run {
				if strings.TrimSpace(c) == "" {
					return fmt.Errorf("%s hook has an empty run command", ev.name)
				}
			}
		}
	}
	return nil
}

// DefaultFileContent is what `gitagger init` writes.
// Every key is documented — this file is the whole schema.
func DefaultFileContent() string {
	return `# gitagger config — edit me, then just run ` + "`gitagger`" + `.
# Everything is optional. Flags beat config, config beats defaults.
# Run gitagger handbook to view the full handbook for the gitagger config!

# Tag size for triple/double/single styles.
# major | minor | patch
scale: patch

# Flavor appended to the tag. stable = no suffix.
# stable | rc | beta | build | nightly
pre: stable

# Tag style. auto follows your tag history.
# auto | triple | double | single | date | sha | sha-num
format: auto

# Custom tag template. Only used when format is custom above.
# API: MAJOR MINOR PATCH YEAR MONTH DAY DATE SHA FULLSHA COUNT PRE NUMBER TAG
# EXAMPLE: "build-<SHA>-v<MAJOR>.<MINOR>.<PATCH>-<DATE><PRE>"
custom: ""

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

# Run a doctor check before creating the tag.
# Aborts early if the remote already has the next tag.
doctor: false

# Show detection details and the plan while working.
# true = same as -v on every run.
verbose: false

# Tag message. Empty = lightweight tag, set = annotated tag.
message: ""

# Hooks run shell commands at lifecycle events: start, success etc.
# Each event takes one block or a list. shell defaults to sh, os defaults to all.
# API: on:, start:, success:, failure:, finish:, run:, shell:, os:, argument:
# EVENTS: start:, success:, failure:, finish: - determine when to run the hook.
# SHELLS: sh, bash, pwsh, batch - determine what shell type to run the command on.
# OS: windows, macos, linux - determine what OS type to run the command on.
# ARGS: you can make Blocks take argument: <name> to run only with gitagger -a <name>.
# ARGS: use argument: none to prevent a block from running when no arguments are used.
# ARGS: use argument: any to make a block run, only when an argument is used.
on:
  failure:
    - shell: sh
      os: macos, linux
      run:
        - echo "Oops! Something went wrong..."
    - shell: batch
      os: windows
      run:
        - echo "Oops! Something went wrong..."
`
}

// CleanFileContent is DefaultFileContent without comments or blank lines.
func CleanFileContent() string {
	var out []string
	for _, line := range strings.Split(DefaultFileContent(), "\n") {
		if t := strings.TrimSpace(line); t != "" && !strings.HasPrefix(t, "#") {
			out = append(out, line)
		}
	}
	return strings.Join(out, "\n") + "\n"
}

// SetRemote writes remote: <name> into the config file, creating a minimal
// file when none exists. Other keys are left alone. Returns the file path.
func SetRemote(dir, name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", fmt.Errorf("remote must not be empty")
	}
	path := Find(dir)
	if path == "" {
		path = filepath.Join(dir, ".gitagger.yml")
		content := "# gitagger config — edit me, then just run `gitagger`.\nremote: " + name + "\n"
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			return "", err
		}
		return path, nil
	}
	// #nosec G304 — path always comes from Find(), which only
	// returns our three known filenames (.gitagger.yml/.yaml/.gitagger).
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	mode := os.FileMode(0o644)
	if st, err := os.Stat(path); err == nil {
		mode = st.Mode().Perm()
	}
	lines := strings.Split(string(raw), "\n")
	replaced := false
	for i, ln := range lines {
		t := strings.TrimSpace(ln)
		if t == "" || strings.HasPrefix(t, "#") {
			continue
		}
		if idx := strings.Index(t, ":"); idx >= 0 && strings.TrimSpace(t[:idx]) == "remote" {
			lines[i] = "remote: " + name
			replaced = true
			break
		}
	}
	if !replaced {
		lines = append(lines, "remote: "+name)
	}
	// #nosec G306 — keeps the config file's existing permissions.
	// #nosec G703 — path is the Find() result above (fixed basenames only).
	if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")), mode); err != nil {
		return "", err
	}
	if _, _, err := Load(dir); err != nil {
		return "", err
	}
	return path, nil
}
