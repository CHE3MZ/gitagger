// Command gitagger automates git tags without the headache.
// All real work lives in internal/ — this file only parses
// flags, loads config, calls internal/cmd, and sets exit codes.
package main

import (
	"fmt"
	"os"
	"strings"

	icmd "github.com/CHE3MZ/gitagger/internal/cmd"
	"github.com/CHE3MZ/gitagger/internal/config"
	"github.com/CHE3MZ/gitagger/internal/style"
)

func main() {
	if err := runCLI(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, style.Error(err.Error()))
		os.Exit(icmd.CodeOf(err))
	}
}

func runCLI(args []string) error {
	if len(args) > 0 {
		switch args[0] {
		case "init":
			return cmdInit(args[1:])
		case "check":
			return cmdSimple(args[1:], "check", icmd.RunCheck)
		case "list", "ls":
			return cmdSimple(args[1:], "list", icmd.RunList)
		case "view":
			return cmdView(args[1:])
		case "remove", "rm":
			return cmdRemove(args[1:])
		case "test":
			return cmdTest(args[1:])
		case "help":
			return cmdHelp(args[1:])
		case "handbook":
			return cmdSimple(args[1:], "handbook", icmd.RunHandbook)
		case "version":
			return cmdVersion(args[1:])
		case "doctor":
			return cmdWithRemote(args[1:], "doctor", icmd.RunDoctor)
		case "remote":
			return cmdRemote(args[1:])
		case "workflow":
			return cmdWorkflow(args[1:])
		case "patch", "minor", "major":
			return cmdTag(args[0], args[1:])
		case "-h", "--help":
			icmd.PrintHelp()
			return nil
		}
	}
	return cmdTag("", args)
}

func cmdTag(scale string, args []string) error {
	f, pos, err := icmd.ParseTagFlags(args)
	if err != nil {
		return icmd.BadArgs("%s", err.Error())
	}
	if f.Help {
		if scale == "" {
			icmd.PrintHelp()
		} else {
			icmd.CommandHelp(scale)
		}
		return nil
	}
	posScale, err := icmd.ParseScale(pos)
	if err != nil {
		return icmd.BadArgs("%s", err.Error())
	}
	if scale == "" {
		scale = posScale
	} else if posScale != "" && posScale != scale {
		return icmd.BadArgs("pick just one of major|minor|patch")
	}
	cwd, _ := os.Getwd()
	dir, err := icmd.ResolveDir(cwd, f.Path)
	if err != nil {
		return err
	}
	cfg, _, err := config.Load(dir)
	if err != nil {
		return icmd.BadArgs("%s", err.Error())
	}
	o, err := icmd.BuildOptions(dir, cfg, scale, f)
	if err != nil {
		return err
	}
	return icmd.RunTag(o)
}

func cmdInit(args []string) error {
	force, verbose, clean := false, false, false
	flagPath := ""
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "-h" || a == "--help":
			icmd.CommandHelp("init")
			return nil
		case a == "-v" || a == "--verbose":
			verbose = true
		case a == "-f" || a == "--force":
			force = true
		case a == "-c" || a == "--clean":
			clean = true
		case a == "-p" && i+1 < len(args):
			i++
			flagPath = args[i]
		case strings.HasPrefix(a, "--path="):
			flagPath = strings.TrimPrefix(a, "--path=")
		case a == "--path" && i+1 < len(args):
			i++
			flagPath = args[i]
		default:
			return icmd.BadArgs("unknown flag %q — try `gitagger init --help`", a)
		}
	}
	cwd, _ := os.Getwd()
	dir, err := icmd.ResolveDir(cwd, flagPath)
	if err != nil {
		return err
	}
	return icmd.RunInit(dir, force, verbose, clean)
}

func cmdSimple(args []string, name string, run func(string, bool) error) error {
	h, v, p, err := icmd.ParseCommon(args, name)
	if err != nil {
		return icmd.BadArgs("%s", err.Error())
	}
	if h {
		icmd.CommandHelp(name)
		return nil
	}
	cwd, _ := os.Getwd()
	dir, err := icmd.ResolveDir(cwd, p)
	if err != nil {
		return err
	}
	return run(dir, v)
}

func cmdWithRemote(args []string, name string, run func(string, string, bool) error) error {
	h, v, p, err := icmd.ParseCommon(args, name)
	if err != nil {
		return icmd.BadArgs("%s", err.Error())
	}
	if h {
		icmd.CommandHelp(name)
		return nil
	}
	cwd, _ := os.Getwd()
	dir, err := icmd.ResolveDir(cwd, p)
	if err != nil {
		return err
	}
	return run(dir, icmd.EffectiveRemote(dir), v)
}

func cmdRemote(args []string) error {
	f, name, err := icmd.ParseRemoteArgs(args)
	if err != nil {
		return icmd.BadArgs("%s", err.Error())
	}
	if f.Help {
		icmd.CommandHelp("remote")
		return nil
	}
	cwd, _ := os.Getwd()
	dir, err := icmd.ResolveDir(cwd, f.Path)
	if err != nil {
		return err
	}
	return icmd.RunRemote(dir, name, f.Show, f.Verbose)
}

func cmdWorkflow(args []string) error {
	if len(args) == 0 {
		return icmd.RunWorkflowOverview()
	}
	switch args[0] {
	case "-l", "--list":
		if len(args) > 1 {
			return icmd.BadArgs("don't know what %q means — try `gitagger workflow --help`%s", args[1], icmd.SuggestionText(icmd.Suggest(strings.TrimLeft(args[1], "-"), icmd.WorkflowSuggestCandidates())))
		}
		return icmd.RunWorkflowList()
		case "-h", "--help":
			icmd.CommandHelp("workflow")
			return nil
	case "init":
		return cmdWorkflowInit(args[1:])
	default:
		return icmd.BadArgs("don't know what %q means — try `gitagger workflow --help`%s", args[0], icmd.SuggestionText(icmd.Suggest(strings.TrimLeft(args[0], "-"), icmd.WorkflowSuggestCandidates())))
	}
}

func cmdWorkflowInit(args []string) error {
	id := ""
	force, verbose := false, false
	flagPath := ""
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "-h" || a == "--help":
			icmd.CommandHelp("workflow")
			return nil
		case a == "-f" || a == "--force":
			force = true
		case a == "-v" || a == "--verbose":
			verbose = true
		case a == "-p" && i+1 < len(args):
			i++
			flagPath = args[i]
		case strings.HasPrefix(a, "--path="):
			flagPath = strings.TrimPrefix(a, "--path=")
		case a == "--path" && i+1 < len(args):
			i++
			flagPath = args[i]
		case strings.HasPrefix(a, "-"):
			return icmd.BadArgs("unknown flag %q — try `gitagger workflow --help`", a)
		default:
			if id != "" {
				return icmd.BadArgs("pick just one provider: `gitagger workflow init <gh|jenkins>`")
			}
			id = strings.ToLower(a)
		}
	}
	if id == "" {
		return icmd.BadArgs("give a provider: `gitagger workflow init <gh|jenkins>`")
	}
	cwd, _ := os.Getwd()
	dir, err := icmd.ResolveDir(cwd, flagPath)
	if err != nil {
		return err
	}
	return icmd.RunWorkflowInit(dir, id, force, verbose)
}

func cmdVersion(args []string) error {	if len(args) > 0 {
		if args[0] == "-h" || args[0] == "--help" {
			icmd.CommandHelp("version")
			return nil
		}
		return icmd.BadArgs("unknown flag %q — try `gitagger version --help`", args[0])
	}
	fmt.Println(icmd.VersionString())
	return nil
}

func cmdHelp(args []string) error {
	if len(args) == 0 {
		icmd.PrintHelp()
		return nil
	}
	if args[0] == "-h" || args[0] == "--help" {
		icmd.PrintHelp()
		return nil
	}
	if len(args) > 1 || !icmd.IsCommand(args[0]) {
		sug := ""
		if len(args) == 1 {
			sug = icmd.SuggestionText(icmd.Suggest(args[0], icmd.SuggestCandidates()))
		}
		return icmd.BadArgs("don't know what %q means — try `gitagger help`%s", strings.Join(args, " "), sug)
	}
	icmd.CommandHelp(args[0])
	return nil
}

func cmdView(args []string) error {
	f, tag, err := icmd.ParseViewArgs(args)
	if err != nil {
		return icmd.BadArgs("%s", err.Error())
	}
	if f.Help {
		icmd.CommandHelp("view")
		return nil
	}
	cwd, _ := os.Getwd()
	dir, err := icmd.ResolveDir(cwd, f.Path)
	if err != nil {
		return err
	}
	return icmd.RunView(dir, tag, f.Verbose)
}

func cmdRemove(args []string) error {
	f, tag, err := icmd.ParseRemoveArgs(args)
	if err != nil {
		return icmd.BadArgs("%s", err.Error())
	}
	if f.Help {
		icmd.CommandHelp("remove")
		return nil
	}
	cwd, _ := os.Getwd()
	dir, err := icmd.ResolveDir(cwd, f.Path)
	if err != nil {
		return err
	}
	return icmd.RunRemove(dir, tag, f.Confirm, f.NoRemote, f.Verbose)
}

func cmdTest(args []string) error {
	f, err := icmd.ParseTestArgs(args)
	if err != nil {
		return icmd.BadArgs("%s", err.Error())
	}
	if f.Help {
		icmd.CommandHelp("test")
		return nil
	}
	cwd, _ := os.Getwd()
	dir, err := icmd.ResolveDir(cwd, f.Path)
	if err != nil {
		return err
	}
	_, err = icmd.RunTest(dir, f.Keep, f.Verbose)
	return err
}
