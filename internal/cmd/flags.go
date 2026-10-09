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
	Arguments    []string
	NoPush       bool
	Force        bool
	DryRun       bool
	RequireClean bool
	Verbose      bool
	Unsafe       bool
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
		case a == "-a" && i+1 < len(args):
			i++
			f.Arguments = append(f.Arguments, args[i])
		case strings.HasPrefix(a, "--argument="):
			f.Arguments = append(f.Arguments, strings.TrimPrefix(a, "--argument="))
		case a == "--argument" && i+1 < len(args):
			i++
			f.Arguments = append(f.Arguments, args[i])
		case a == "-r" || a == "--require-clean":
			f.RequireClean = true
		case a == "-u" || a == "--unsafe":
			f.Unsafe = true
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
			return "", fmt.Errorf("don't know what %q means — try `gitagger --help`%s", p, SuggestionText(Suggest(p, SuggestCandidates())))
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

// ViewFlags for `gitagger view <tag>`.
type ViewFlags struct {
	Verbose bool
	Help    bool
	Path    string
}

// ParseViewArgs parses one tag plus -h/--help, -v/--verbose and -p/--path.
func ParseViewArgs(args []string) (ViewFlags, string, error) {
	var f ViewFlags
	tag := ""
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "-h" || a == "--help":
			f.Help = true
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
			return f, "", fmt.Errorf("unknown flag %q — try `gitagger view --help`", a)
		default:
			if tag != "" {
				return f, "", fmt.Errorf("pick just one tag — try `gitagger view --help`")
			}
			tag = a
		}
	}
	return f, tag, nil
}

// RemoveFlags for `gitagger remove <tag>`.
type RemoveFlags struct {
	Confirm  bool
	NoRemote bool
	Verbose  bool
	Help     bool
	Path     string
}

// ParseRemoveArgs parses one tag plus -c/--confirm, -n/--no-remote,
// -v/--verbose and -p/--path.
func ParseRemoveArgs(args []string) (RemoveFlags, string, error) {
	var f RemoveFlags
	tag := ""
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "-h" || a == "--help":
			f.Help = true
		case a == "-c" || a == "--confirm":
			f.Confirm = true
		case a == "-n" || a == "--no-remote":
			f.NoRemote = true
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
			return f, "", fmt.Errorf("unknown flag %q — try `gitagger remove --help`", a)
		default:
			if tag != "" {
				return f, "", fmt.Errorf("pick just one tag — try `gitagger remove --help`")
			}
			tag = a
		}
	}
	return f, tag, nil
}

// TestFlags for `gitagger test`.
type TestFlags struct {
	Keep    bool
	Verbose bool
	Help    bool
	Path    string
}

// ParseTestArgs parses -k/--keep, -v/--verbose and -p/--path.
// Test takes no positionals: anything else is an error.
func ParseTestArgs(args []string) (TestFlags, error) {
	var f TestFlags
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "-h" || a == "--help":
			f.Help = true
		case a == "-k" || a == "--keep":
			f.Keep = true
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
		default:
			return f, fmt.Errorf("unknown flag %q — try `gitagger test --help`", a)
		}
	}
	return f, nil
}
