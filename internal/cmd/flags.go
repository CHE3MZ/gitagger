// Flag parsing for the thin main.go dispatcher.
// Only the flags from the help text exist. No git, no detection here.
package cmd

import (
	"fmt"
	"strings"
)

// TagFlags mirrors `gitagger [patch|minor|major] [--flags]`.
type TagFlags struct {
	Pre          string
	Format       string
	Message      string
	Path         string
	NoPush       bool
	Force        bool
	DryRun       bool
	RequireClean bool
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
		case a == "-n" || a == "--no-push":
			f.NoPush = true
		case a == "-f" || a == "--force":
			f.Force = true
		case a == "-d" || a == "--dry-run":
			f.DryRun = true
		case a == "-v" || a == "--verbose":
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
		case a == "-m" && i+1 < len(args):
			i++
			f.Message = args[i]
		case strings.HasPrefix(a, "--message="):
			f.Message = strings.TrimPrefix(a, "--message=")
		case a == "--message" && i+1 < len(args):
			i++
			f.Message = args[i]
		case a == "-p" && i+1 < len(args):
			i++
			f.Path = args[i]
		case strings.HasPrefix(a, "--path="):
			f.Path = strings.TrimPrefix(a, "--path=")
		case a == "--path" && i+1 < len(args):
			i++
			f.Path = args[i]
		case a == "-r" || a == "--require-clean":
			f.RequireClean = true
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

// ParseCommon parses -h/--help, -v/--verbose and -p/--path for subcommands.
// Anything else is an error naming the command's own help.
func ParseCommon(args []string, cmd string) (help, verbose bool, path string, err error) {
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "-h" || a == "--help":
			help = true
		case a == "-v" || a == "--verbose":
			verbose = true
		case a == "-p" && i+1 < len(args):
			i++
			path = args[i]
		case strings.HasPrefix(a, "--path="):
			path = strings.TrimPrefix(a, "--path=")
		case a == "--path" && i+1 < len(args):
			i++
			path = args[i]
		default:
			return false, false, "", fmt.Errorf("unknown flag %q — try `gitagger %s --help`", a, cmd)
		}
	}
	return help, verbose, path, nil
}

// RemoteFlags for `gitagger remote <name>` and `gitagger remote --show`.
type RemoteFlags struct {
	Show    bool
	Verbose bool
	Help    bool
	Path    string
}

// ParseRemoteArgs parses an optional remote name plus -s/--show and -p.
func ParseRemoteArgs(args []string) (RemoteFlags, string, error) {
	var f RemoteFlags
	name := ""
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "-h" || a == "--help":
			f.Help = true
		case a == "-s" || a == "--show":
			f.Show = true
		case a == "-v" || a == "--verbose":
			f.Verbose = true
		case a == "-p" && i+1 < len(args):
			i++
			f.Path = args[i]
		case strings.HasPrefix(a, "--path="):
			f.Path = strings.TrimPrefix(a, "--path=")
		case a == "--path" && i+1 < len(args):
			i++
			f.Path = args[i]
		case strings.HasPrefix(a, "-"):
			return f, "", fmt.Errorf("unknown flag %q — try `gitagger remote --help`", a)
		default:
			if name != "" {
				return f, "", fmt.Errorf("pick just one remote name")
			}
			name = a
		}
	}
	return f, name, nil
}
