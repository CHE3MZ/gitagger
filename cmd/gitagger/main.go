// Command gitagger is a friendly tag maker: just run it.
// All real work lives in internal/ — this file only parses
// flags, loads config, calls internal/cmd, and sets exit codes.
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/CHE3MZ/gitagger/internal/config"
	"github.com/CHE3MZ/gitagger/internal/detect"
	"github.com/CHE3MZ/gitagger/internal/run"
	icmd "github.com/CHE3MZ/gitagger/internal/cmd"
	"github.com/CHE3MZ/gitagger/internal/style"
)

const version = "0.1.0"

func main() {
	if err := runCLI(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, style.Error(err.Error()))
		os.Exit(icmd.CodeOf(err))
	}
}

func runCLI(args []string) error {
	if len(args) > 0 && args[0] == "view" {
		fmt.Println(style.Dim("heads up: `view` is now `list` — still works, just renamed"))
		args = append([]string{"list"}, args[1:]...)
	}
	for i, a := range args {
		if a == "yes" && !strings.HasPrefix(a, "-") {
			fmt.Println(style.Dim("heads up: bare `yes` is old style — gitagger never prompts by default now"))
			args = append(args[:i], args[i+1:]...)
			break
		}
	}
	if len(args) > 0 {
		switch args[0] {
		case "init":
			return cmdInit(args[1:])
		case "list", "ls":
			return cmdList(args[1:])
		case "doctor":
			return cmdDoctor(args[1:])
		case "version", "-v", "--version":
			fmt.Printf("gitagger %s\n", version)
			return nil
		case "-h", "--help", "help":
			icmd.PrintHelp()
			return nil
		}
	}
	return cmdTag(args)
}

func cmdTag(args []string) error {
	for _, a := range args {
		if a == "-v" || a == "--version" {
			fmt.Printf("gitagger %s\n", version)
			return nil
		}
	}
	f, pos, err := icmd.ParseTagFlags(args)
	if err != nil {
		return icmd.BadArgs("%s", err.Error())
	}
	if f.Help {
		icmd.PrintHelp()
		return nil
	}
	scale, err := icmd.ParseScale(pos)
	if err != nil {
		return icmd.BadArgs("%s", err.Error())
	}
	dir, _ := os.Getwd()
	cfg, _, err := config.Load(dir)
	if err != nil {
		return icmd.BadArgs("%s", err.Error())
	}
	o := run.Options{Dir: dir}
	o = run.FromConfig(dir, cfg, o)
	if scale != "" {
		o.Scale = scale
	}
	if f.Pre != "" {
		o.Pre = strings.ToLower(f.Pre)
	}
	if f.Format != "" {
		ff, ok := detect.ParseFormat(f.Format)
		if !ok {
			return icmd.BadArgs("bad --format %q (want auto|triple|double|single|date|sha|sha-num|custom)", f.Format)
		}
		o.Format = ff
	}
	if f.Custom != "" {
		o.Custom = f.Custom
	}
	if f.Remote != "" {
		o.Remote = f.Remote
	}
	o.Push = cfg.Push
	if f.NoPush {
		o.Push = false
	}
	o.Confirm = cfg.Confirm || f.Confirm
	o.Force = f.Force
	o.DryRun = f.DryRun
	o.Message = cfg.Message
	if f.Message != "" {
		o.Message = f.Message
	}
	if f.RequireClean {
		o.RequireClean = true
	} else {
		o.RequireClean = cfg.RequireClean
	}
	o.JSON = f.JSON
	o.Verbose = f.Verbose
	if o.JSON {
		style.SetEnabled(false)
	}
	return icmd.RunTag(o)
}

func cmdInit(args []string) error {
	force := false
	for _, a := range args {
		if a == "--force" {
			force = true
		}
	}
	dir, _ := os.Getwd()
	return icmd.RunInit(dir, force)
}

func cmdList(args []string) error {
	lo := icmd.ParseListArgs(args)
	dir, _ := os.Getwd()
	return icmd.RunList(dir, lo.Limit, lo.JSON)
}

func cmdDoctor(args []string) error {
	fix := false
	for _, a := range args {
		if a == "--fix" {
			fix = true
		}
	}
	dir, _ := os.Getwd()
	cfg, _, _ := config.Load(dir)
	remote := cfg.Remote
	if remote == "" {
		remote = "origin"
	}
	return icmd.RunDoctor(dir, remote, fix)
}
