// Command gitagger is a friendly tag maker: just run it.
// All the real work lives in internal/ — this file only
// parses flags, asks internal/ what to do, and prints nicely.
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/CHE3MZ/gitagger/internal/config"
	"github.com/CHE3MZ/gitagger/internal/detect"
	"github.com/CHE3MZ/gitagger/internal/git"
	"github.com/CHE3MZ/gitagger/internal/run"
	"github.com/CHE3MZ/gitagger/internal/style"
)

const version = "0.1.0"

func main() {
	if err := runCLI(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, style.Error(err.Error()))
		os.Exit(1)
	}
}

func runCLI(args []string) error {
	// Legacy bare words: `gitagger view` -> list, `gitagger yes` -> no-op.
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
			printHelp()
			return nil
		}
	}
	return cmdTag(args)
}

// optsFrom parses tag flags. Keeps it simple: no external CLI lib.
type flagSet struct {
	pre          string
	format       string
	custom       string
	remote       string
	message      string
	noPush       bool
	confirm      bool
	yesCompat    bool
	force        bool
	dryRun       bool
	requireClean bool
	json         bool
	verbose      bool
	help         bool
}

func parseFlags(args []string) (flagSet, []string, error) {
	var f flagSet
	var pos []string
	i := 0
	for i < len(args) {
		a := args[i]
		switch {
		case a == "--":
			pos = append(pos, args[i+1:]...)
			i = len(args)
		case a == "-h" || a == "--help":
			f.help = true
		case a == "-y" || a == "--yes":
			f.yesCompat = true // no-op: no prompts by default
		case a == "--no-push":
			f.noPush = true
		case a == "--confirm":
			f.confirm = true
		case a == "--force":
			f.force = true
		case a == "--dry-run":
			f.dryRun = true
		case a == "--require-clean":
			f.requireClean = true
		case a == "--json":
			f.json = true
		case a == "--verbose":
			f.verbose = true
		case a == "-v" || a == "--version":
			fmt.Printf("gitagger %s\n", version)
			os.Exit(0)
		case strings.HasPrefix(a, "--pre="):
			f.pre = strings.TrimPrefix(a, "--pre=")
		case a == "--pre" && i+1 < len(args):
			i++
			f.pre = args[i]
		case strings.HasPrefix(a, "--format="):
			f.format = strings.TrimPrefix(a, "--format=")
		case a == "--format" && i+1 < len(args):
			i++
			f.format = args[i]
		case strings.HasPrefix(a, "--custom="):
			f.custom = strings.TrimPrefix(a, "--custom=")
		case a == "--custom" && i+1 < len(args):
			i++
			f.custom = args[i]
		case strings.HasPrefix(a, "--remote="):
			f.remote = strings.TrimPrefix(a, "--remote=")
		case a == "--remote" && i+1 < len(args):
			i++
			f.remote = args[i]
		case a == "-m" && i+1 < len(args):
			i++
			f.message = args[i]
		case strings.HasPrefix(a, "--message="):
			f.message = strings.TrimPrefix(a, "--message=")
		case a == "--message" && i+1 < len(args):
			i++
			f.message = args[i]
		case strings.HasPrefix(a, "-"):
			return f, pos, fmt.Errorf("unknown flag %q — try `gitagger --help`", a)
		default:
			pos = append(pos, a)
		}
		i++
	}
	return f, pos, nil
}

func cmdTag(args []string) error {
	f, pos, err := parseFlags(args)
	if err != nil {
		return err
	}
	if f.help {
		printHelp()
		return nil
	}
	if f.json {
		style.SetEnabled(false)
	}
	// Scale is the only positional: major|minor|patch.
	scale := ""
	for _, p := range pos {
		switch strings.ToLower(p) {
		case "major", "minor", "patch":
			if scale != "" {
				return fmt.Errorf("pick just one of major|minor|patch")
			}
			scale = strings.ToLower(p)
		default:
			return fmt.Errorf("don't know what %q means — try `gitagger --help`", p)
		}
	}

	dir, _ := os.Getwd()
	cfg, _, err := config.Load(dir)
	if err != nil {
		return err
	}
	o := run.Options{Dir: dir}
	o = run.FromConfig(dir, cfg, o)
	if scale != "" {
		o.Scale = scale
	}
	if f.pre != "" {
		o.Pre = strings.ToLower(f.pre)
	}
	if f.format != "" {
		ff, ok := detect.ParseFormat(f.format)
		if !ok {
			return fmt.Errorf("bad --format %q (want auto|triple|double|single|date|sha|sha-num|custom)", f.format)
		}
		o.Format = ff
	}
	if f.custom != "" {
		o.Custom = f.custom
	}
	if f.remote != "" {
		o.Remote = f.remote
	}
	// Bools: config defaults, flags override.
	o.Push = cfg.Push
	if f.noPush {
		o.Push = false
	}
	o.Confirm = cfg.Confirm || f.confirm
	o.Force = f.force
	o.DryRun = f.dryRun
	o.Message = cfg.Message
	if f.message != "" {
		o.Message = f.message
	}
	if f.requireClean {
		o.RequireClean = true
	} else {
		o.RequireClean = cfg.RequireClean
	}
	o.JSON = f.json
	o.Verbose = f.verbose

	if !detect.ValidPreCheck(o.Pre) {
		return fmt.Errorf("bad --pre %q (want stable|rc|beta|build|nightly)", o.Pre)
	}
	if o.Format == detect.Custom && strings.TrimSpace(o.Custom) == "" {
		return fmt.Errorf("format is custom but --custom is empty (v2 feature anyway)")
	}

	plan, err := run.ComputePlan(o)
	if err != nil {
		return err
	}
	if err := run.CheckHeadTagged(o, plan.Next); err != nil {
		return err
	}

	if o.DryRun {
		printDryRun(o, plan)
		return nil
	}
	if o.JSON {
		return printJSON(plan.Prev, plan.Next, false, o.Remote)
	}

	greetPlan(o, plan)

	// Create the local tag first — push never deletes it.
	if err := git.CreateTag(dir, plan.Next, o.Message, o.Force); err != nil {
		return fmt.Errorf("couldn't create tag %s (%v)", plan.Next, err)
	}
	fmt.Printf("%s %s\n", style.Green("created tag"), style.BoldGreen(plan.Next))

	if !o.Push {
		fmt.Println(style.Dim("kept locally (--no-push). push later with `git push " + o.Remote + " " + plan.Next + "`"))
		return nil
	}
	if o.Confirm && !askYes(fmt.Sprintf("push %s to %s?", plan.Next, o.Remote)) {
		fmt.Println(style.Dim("kept locally. push later with `git push " + o.Remote + " " + plan.Next + "`"))
		return printJSONMaybe(o, plan.Prev, plan.Next, false, o.Remote)
	}
	outcome, err := run.EnsurePush(o, plan.Next)
	if err != nil {
		return err
	}
	if outcome.Pushed {
		fmt.Printf("%s %s to %s\n", style.Green("pushed"), style.BoldGreen(plan.Next), style.White(o.Remote))
	} else {
		fmt.Println(style.Warn(outcome.Skipped))
	}
	return printJSONMaybe(o, plan.Prev, plan.Next, outcome.Pushed, o.Remote)
}

func greetPlan(o run.Options, plan run.Plan) {
	if o.Verbose {
		fmt.Println(style.Dim(fmt.Sprintf("detected: %s, v-prefix=%v (from history)", plan.Format, plan.VPrefix)))
		if plan.Prev != "" {
			fmt.Println(style.Dim("was: " + plan.Prev))
		} else {
			fmt.Println(style.Dim("no tags yet — starting fresh"))
		}
		fmt.Println(style.Dim("next: " + plan.Next))
	}
}

func printDryRun(o run.Options, plan run.Plan) {
	fmt.Println(style.Header("dry run — nothing created"))
	if plan.Prev != "" {
		fmt.Printf("was:  %s\n", style.White(plan.Prev))
	}
	fmt.Printf("next: %s  %s\n", style.BoldGreen(plan.Next),
		style.Dim(fmt.Sprintf("(%s, %s)", plan.Format, o.Pre)))
	if o.Push {
		fmt.Println(style.Dim("would push to " + o.Remote))
	} else {
		fmt.Println(style.Dim("would keep locally (--no-push)"))
	}
}

func printJSONMaybe(o run.Options, prev, next string, pushed bool, remote string) error {
	if !o.JSON {
		return nil
	}
	return printJSON(prev, next, pushed, remote)
}

func printJSON(prev, next string, pushed bool, remote string) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(map[string]any{
		"prev": prev, "next": next, "pushed": pushed, "remote": remote,
	})
}

func askYes(prompt string) bool {
	fmt.Printf("%s %s ", style.White(prompt), style.Gray("[Y/n]"))
	line, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		return true // non-interactive + --confirm: proceed
	}
	line = strings.ToLower(strings.TrimSpace(line))
	return line == "" || line == "y" || line == "yes"
}

func cmdInit(args []string) error {
	force := false
	for _, a := range args {
		if a == "--force" {
			force = true
		} else if a == "-h" || a == "--help" {
			fmt.Println(style.Blue("Makes a .gitagger.yml with sensible defaults so `gitagger` alone does the right thing."))
			fmt.Println(style.Blue("Won't overwrite without --force."))
			return nil
		}
	}
	dir, _ := os.Getwd()
	existing := config.Find(dir)
	if existing != "" && !force {
		return fmt.Errorf("config already exists at %s (use --force to overwrite)", existing)
	}
	path := dir + string(os.PathSeparator) + ".gitagger.yml"
	// Owner-only perms: gosec-clean and no reason for group/others to read it.
	if err := os.WriteFile(path, []byte(config.DefaultFileContent()), 0o600); err != nil {
		return err
	}
	fmt.Printf("%s %s\n", style.Green("wrote"), style.White(".gitagger.yml"))
	fmt.Println(style.Dim("tweak it if you like, then just run `gitagger`"))
	return nil
}

func cmdList(args []string) error {
	limit := 20
	jsonOut := false
	for i := 0; i < len(args); i++ {
		a := args[i]
		if a == "--json" {
			jsonOut = true
		} else if a == "--limit" && i+1 < len(args) {
			i++
			var n int
			if _, err := fmt.Sscanf(args[i], "%d", &n); err == nil && n > 0 {
				limit = n
			}
		} else if strings.HasPrefix(a, "--limit=") {
			var n int
			if _, err := fmt.Sscanf(strings.TrimPrefix(a, "--limit="), "%d", &n); err == nil && n > 0 {
				limit = n
			}
		}
	}
	if jsonOut {
		style.SetEnabled(false)
	}
	if err := git.EnsureAvailable(); err != nil {
		return err
	}
	dir, _ := os.Getwd()
	tags, err := git.ListTags(dir)
	if err != nil {
		return err
	}
	if len(tags) > limit {
		tags = tags[len(tags)-limit:]
	}
	if jsonOut {
		return json.NewEncoder(os.Stdout).Encode(tags)
	}
	if len(tags) == 0 {
		fmt.Println(style.Dim("no tags yet — run `gitagger` to make your first one"))
		return nil
	}
	for _, t := range tags {
		fmt.Println(style.White(t))
	}
	fmt.Println(style.Dim(fmt.Sprintf("%d tag(s), newest last", len(tags))))
	return nil
}

func cmdDoctor(args []string) error {
	fix := false
	for _, a := range args {
		if a == "--fix" {
			fix = true
		}
	}
	if err := git.EnsureAvailable(); err != nil {
		return err
	}
	dir, _ := os.Getwd()
	cfg, _, _ := config.Load(dir)
	remote := cfg.Remote
	if remote == "" {
		remote = "origin"
	}
	tags, err := git.ListTags(dir)
	if err != nil {
		return err
	}
	fmt.Println(style.Header("gitagger doctor"))
	fmt.Printf("local tags: %s\n", style.White(fmt.Sprint(len(tags))))

	// Malformed tags (unknown style).
	var bad []string
	for _, t := range tags {
		if detect.Classify(t) == detect.Unknown {
			bad = append(bad, t)
		}
	}
	if len(bad) > 0 {
		fmt.Printf("%s %s\n", style.Red("odd tags:"), style.White(strings.Join(bad, ", ")))
		fmt.Println(style.Dim("these don't look like triple/double/date/sha — harmless, just ignored by auto-detect"))
		if fix {
			if !askYes(fmt.Sprintf("delete %d local odd tag(s)?", len(bad))) {
				fmt.Println(style.Dim("ok, left them alone"))
			} else {
				for _, t := range bad {
					fmt.Println(style.Dim("would delete " + t + " (use `git tag -d " + t + "`)"))
				}
				fmt.Println(style.Dim("dry preview only in v1 — delete by hand for now"))
			}
		}
	} else {
		fmt.Println(style.Green("tag styles look clean"))
	}

	// Remote compare (graceful when offline).
	if _, err := git.RemoteURL(dir, remote); err != nil {
		fmt.Println(style.Warn(fmt.Sprintf("no remote %q — nothing to compare", remote)))
		return nil
	}
	if err := git.FetchTags(dir, remote); err != nil {
		fmt.Println(style.Warn(fmt.Sprintf("couldn't reach %q — skipping remote check", remote)))
		return nil
	}
	fmt.Println(style.Green("remote reachable, tags fetched"))
	headTags, _ := git.TagsPointingAtHEAD(dir)
	if len(headTags) == 0 {
		fmt.Println(style.Dim("HEAD has no tag yet — `gitagger` will fix that"))
	} else {
		fmt.Printf("HEAD tagged: %s\n", style.White(strings.Join(headTags, ", ")))
	}
	return nil
}

func printHelp() {
	// Blue help, friendly tone, short lines.
	fmt.Println(style.Bold(style.Blue("gitagger — tags without the headache")))
	fmt.Println(style.Blue("Just run `gitagger` and it picks the next tag from your history, then pushes it if it can."))
	fmt.Println()
	fmt.Println(style.Header("Use it"))
	fmt.Printf("  %s  %s\n", style.BoldGreen("gitagger"), style.White("make the next patch + push"))
	fmt.Printf("  %s   %s\n", style.White("gitagger minor"), style.Dim("v1.2.3 -> v1.3.0"))
	fmt.Printf("  %s  %s\n", style.White("gitagger major --pre rc"), style.Dim("v1.2.3 -> v2.0.0-rc"))
	fmt.Printf("  %s  %s\n", style.White("gitagger --dry-run"), style.Dim("peek first, change nothing"))
	fmt.Printf("  %s   %s\n", style.White("gitagger --no-push"), style.Dim("tag locally only"))
	fmt.Println()
	fmt.Println(style.Header("Flags"))
	fmt.Println(style.Blue("  --pre stable|rc|beta|build|nightly   prerelease flavor (default stable)"))
	fmt.Println(style.Blue("  --format auto|triple|double|single|date|sha|sha-num   force a style (default auto)"))
	fmt.Println(style.Blue("  --no-push, --confirm, --remote, --force, --dry-run, -m, --json, --verbose"))
	fmt.Println()
	fmt.Println(style.Header("More"))
	fmt.Printf("  %s  %s\n", style.White("gitagger init"), style.Dim("write a .gitagger.yml once, then forget flags"))
	fmt.Printf("  %s  %s\n", style.White("gitagger list"), style.Dim("show tags, newest last"))
	fmt.Printf("  %s  %s\n", style.White("gitagger doctor"), style.Dim("sanity check local vs remote"))
	fmt.Println()
	fmt.Println(style.Dim("No prompts by default. No remote or offline? Your tag stays local with a hint."))
}
