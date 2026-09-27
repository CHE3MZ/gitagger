// Flag parsing for the thin main.go dispatcher.
// No git, no detection here — just arg shapes.
package cmd

import (
	"fmt"
	"strings"
)

// TagFlags mirrors `gitagger [major|minor|patch] [flags]`.
type TagFlags struct {
	Pre          string
	Format       string
	Custom       string
	Remote       string
	Message      string
	NoPush       bool
	Confirm      bool
	YesCompat    bool
	Force        bool
	DryRun       bool
	RequireClean bool
	JSON         bool
	Verbose      bool
	Help         bool
}

// ParseTagFlags parses tag flags, returning positionals separately.
func ParseTagFlags(args []string) (TagFlags, []string, error) {
	var f TagFlags
	var pos []string
	i := 0
	for i < len(args) {
		a := args[i]
		switch {
		case a == "--":
			pos = append(pos, args[i+1:]...)
			i = len(args)
		case a == "-h" || a == "--help":
			f.Help = true
		case a == "-y" || a == "--yes":
			f.YesCompat = true
		case a == "--no-push":
			f.NoPush = true
		case a == "--confirm":
			f.Confirm = true
		case a == "--force":
			f.Force = true
		case a == "--dry-run":
			f.DryRun = true
		case a == "--require-clean":
			f.RequireClean = true
		case a == "--json":
			f.JSON = true
		case a == "--verbose":
			f.Verbose = true
		case strings.HasPrefix(a, "--pre="):
			f.Pre = strings.TrimPrefix(a, "--pre=")
		case a == "--pre" && i+1 < len(args):
			i++
			f.Pre = args[i]
		case strings.HasPrefix(a, "--format="):
			f.Format = strings.TrimPrefix(a, "--format=")
		case a == "--format" && i+1 < len(args):
			i++
			f.Format = args[i]
		case strings.HasPrefix(a, "--custom="):
			f.Custom = strings.TrimPrefix(a, "--custom=")
		case a == "--custom" && i+1 < len(args):
			i++
			f.Custom = args[i]
		case strings.HasPrefix(a, "--remote="):
			f.Remote = strings.TrimPrefix(a, "--remote=")
		case a == "--remote" && i+1 < len(args):
			i++
			f.Remote = args[i]
		case a == "-m" && i+1 < len(args):
			i++
			f.Message = args[i]
		case strings.HasPrefix(a, "--message="):
			f.Message = strings.TrimPrefix(a, "--message=")
		case a == "--message" && i+1 < len(args):
			i++
			f.Message = args[i]
		case strings.HasPrefix(a, "-"):
			return f, pos, fmt.Errorf("unknown flag %q — try `gitagger --help`", a)
		default:
			pos = append(pos, a)
		}
		i++
	}
	return f, pos, nil
}

// ParseScale validates the single optional positional.
func ParseScale(pos []string) (string, error) {
	scale := ""
	for _, p := range pos {
		switch strings.ToLower(p) {
		case "major", "minor", "patch":
			if scale != "" {
				return "", fmt.Errorf("pick just one of major|minor|patch")
			}
			scale = strings.ToLower(p)
		default:
			return "", fmt.Errorf("don't know what %q means — try `gitagger --help`", p)
		}
	}
	return scale, nil
}

// ListOpts for `gitagger list`.
type ListOpts struct {
	Limit int
	JSON  bool
}

// ParseListArgs parses `--limit N` and `--json`.
func ParseListArgs(args []string) ListOpts {
	o := ListOpts{Limit: 20}
	for i := 0; i < len(args); i++ {
		a := args[i]
		if a == "--json" {
			o.JSON = true
		} else if a == "--limit" && i+1 < len(args) {
			i++
			var n int
			if _, err := fmt.Sscanf(args[i], "%d", &n); err == nil && n > 0 {
				o.Limit = n
			}
		} else if strings.HasPrefix(a, "--limit=") {
			var n int
			if _, err := fmt.Sscanf(strings.TrimPrefix(a, "--limit="), "%d", &n); err == nil && n > 0 {
				o.Limit = n
			}
		}
	}
	return o
}
