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
		case "help":
			return cmdHelp(args[1:])
		case "version":
			return cmdVersion(args[1:])
		case "doctor":
			return cmdWithRemote(args[1:], "doctor", icmd.RunDoctor)
		case "remote":
			return cmdRemote(args[1:])
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
	dir, _ := os.Getwd()
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
	for _, a := range args {
		switch a {
		case "-h", "--help":
			icmd.CommandHelp("init")
			return nil
		case "-v", "--verbose":
			verbose = true
		case "-f", "--force":
			force = true
		case "--clean":
			clean = true
		default:
			return icmd.BadArgs("unknown flag %q — try `gitagger init --help`", a)
		}
	}
	dir, _ := os.Getwd()
	return icmd.RunInit(dir, force, verbose, clean)
}

func cmdSimple(args []string, name string, run func(string, bool) error) error {
	h, v, err := icmd.ParseCommon(args, name)
	if err != nil {
		return icmd.BadArgs("%s", err.Error())
	}
	if h {
		icmd.CommandHelp(name)
		return nil
	}
	dir, _ := os.Getwd()
	return run(dir, v)
}

func cmdWithRemote(args []string, name string, run func(string, string, bool) error) error {
	h, v, err := icmd.ParseCommon(args, name)
	if err != nil {
		return icmd.BadArgs("%s", err.Error())
	}
	if h {
		icmd.CommandHelp(name)
		return nil
	}
	dir, _ := os.Getwd()
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
	dir, _ := os.Getwd()
	return icmd.RunRemote(dir, name, f.Show, f.Verbose)
}

func cmdVersion(args []string) error {
	if len(args) > 0 {
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
		return icmd.BadArgs("don't know what %q means — try `gitagger help`", strings.Join(args, " "))
	}
	icmd.CommandHelp(args[0])
	return nil
}
