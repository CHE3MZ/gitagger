// Help text. Words are the owner's; colors follow the house scheme:
// gray for info, blue for flags, green for commands, red stays for errors.
package cmd

import (
	"fmt"
	"strings"

	"github.com/CHE3MZ/gitagger/internal/style"
)

// version is stamped at release time via ldflags (-X .../internal/cmd.version=...).
// Unstamped dev builds report the dev string below.
var version = "dev"

// VersionString reports the build version.
func VersionString() string {
	if version == "dev" {
		return "gitagger is on the dev build version."
	}
	return "gitagger " + version
}

// PrintHelp prints the main help text.
func PrintHelp() {
	fmt.Println(style.Bold("automate git tags without the headache."))
	fmt.Println()
	fmt.Println(style.Header("USAGE"))
	fmt.Println()
	fmt.Printf("  %s %s %s\n", style.White("gitagger"), style.Green("[command]"), style.Blue("[--flags]"))
	fmt.Println()
	fmt.Println(style.Header("EXAMPLES"))
	fmt.Println()
	for _, e := range examples {
		inv, comment, _ := strings.Cut(e, "#")
		head, rest, _ := strings.Cut(inv, " ")
		fmt.Printf("  %s%s%s\n", style.White(head), style.Green(" "+rest), style.Gray("#"+comment))
	}
	fmt.Println()
	fmt.Println(style.Header("COMMANDS"))
	fmt.Println()
	for _, c := range commands {
		fmt.Printf("  %s%s\n", style.Green(c.name), style.Gray(c.rest))
	}
	fmt.Println()
	fmt.Println(style.Header("Flags"))
	for _, l := range renderFlagRows(2, 24) {
		fmt.Println(l)
	}
}

type commandRow struct {
	name string
	rest string
}

var examples = []string{
	"gitagger                   # create a new tag and push",
	"gitagger patch             # create a new patch tag",
	"gitagger minor             # create a new minor tag",
	"gitagger major             # create a new major tag",
	"gitagger major --pre beta  # create a new major beta tag",
}

var commands = []commandRow{
	{"init", "       [--flags]       Generates a .gitagger.yml file"},
	{"check", "      [--flags]       Checks if your .gitagger.yml config is valid"},
	{"list", "       [--flags]       Shows all tags"},
	{"help", "       [--flags]       Print this help text"},
	{"doctor", "     [--flags]       Audit local vs remote tag health status"},
	{"remote", "     [--flags]       Change the remote URL the tag will get pushed to"},
	{"version", "    [--flags]       Show the build version"},
	{"patch", "      [--flags]       New tag addition by    0.0.X"},
	{"minor", "      [--flags]       New tag rounding to    0.X.0"},
	{"major", "      [--flags]       New tag rounding to    X.0.0"},
}

type flagRow struct {
	spec   string // blue: the flag itself
	desc   string // gray: what it does
	values string // blue: allowed values, empty when none
}

var flagRows = []flagRow{
	{"--pre", "prerelease flavor:", "stable|rc|beta|build|nightly (default stable)"},
	{"--format", "force a style:", "auto|triple|double|single|date|sha|sha-num (default auto)"},
	{"-h --help", "Print the help text for a command.", ""},
	{"-v --verbose", "Enable verbose mode.", ""},
	{"-n --no-push", "Create tag without pushing.", ""},
	{"-d --dry-run", "Do A dry-run for testing.", ""},
	{"-f --force", "Force push to remote or force an action.", ""},
	{"-m --message <msg>", "Tag message. (empty by default)", ""},
	{"-r --require-clean", "Abort if the working tree is dirty.", ""},
}

// renderFlagRows paints one flag block: blue specs and values, gray descs,
// all descs starting at descCol. Colors bake in after style init.
func renderFlagRows(indent, descCol int) []string {
	out := make([]string, 0, len(flagRows))
	for _, f := range flagRows {
		line := strings.Repeat(" ", indent) + style.Blue(f.spec) +
			strings.Repeat(" ", descCol-indent-len(f.spec)) + style.Gray(f.desc)
		if f.values != "" {
			line += " " + style.Blue(f.values)
		}
		out = append(out, line)
	}
	return out
}

// CommandHelp prints help for one command in the same direct style.
func CommandHelp(cmd string) {
	if cmd == "ls" {
		cmd = "list"
	}
	if h, ok := commandHelp[cmd]; ok {
		fmt.Println(style.Gray(h.desc))
		fmt.Println()
		fmt.Println(style.Header("USAGE"))
		fmt.Println()
		fmt.Printf("  %s %s\n", style.White("gitagger"), style.Green(h.usage))
		if len(h.examples) > 0 {
			fmt.Println()
			fmt.Println(style.Header("EXAMPLE"))
			fmt.Println()
			for _, e := range h.examples {
				head, rest, _ := strings.Cut(e, " ")
				fmt.Printf("  %s%s\n", style.White(head), style.Green(" "+rest))
			}
		}
		if len(h.flags) > 0 {
			fmt.Println()
			fmt.Println(style.Header("FLAGS"))
			for _, f := range h.flags {
				fmt.Println(style.Blue(f))
			}
		}
		return
	}
	PrintHelp()
}

type cmdHelp struct {
	desc     string
	usage    string
	examples []string
	flags    []string
}

var tagFlags = renderFlagRows(0, 20)

var commandHelp = map[string]cmdHelp{
	"init": {
		"Generates a .gitagger.yml file.",
		"init [-f]",
		nil,
		[]string{
			"-f --force     Overwrite the config file if it exists.",
			"-v --verbose   Also show the full path.",
			"-h --help      Print the help text for a command.",
		},
	},
	"check": {
		"Checks if your .gitagger.yml config is valid.",
		"check [-v]",
		nil,
		[]string{
			"-v --verbose   Show the resolved config values.",
			"-h --help      Print the help text for a command.",
		},
	},
	"list": {
		"Shows all tags.",
		"list [-v]",
		nil,
		[]string{
			"-v --verbose   Also show the remote URL.",
			"-h --help      Print the help text for a command.",
		},
	},
	"help": {
		"Print this help text.",
		"help [command]",
		[]string{"gitagger help init"},
		nil,
	},
	"doctor": {
		"Audit local vs remote tag health status.",
		"doctor [-v]",
		nil,
		[]string{
			"-v --verbose   Also show the remote URL.",
			"-h --help      Print the help text for a command.",
		},
	},
	"remote": {
		"Change the remote URL the tag will get pushed to.",
		"remote <name>",
		[]string{"gitagger remote origin", "gitagger remote upstream", "gitagger remote --show"},
		[]string{
			"-s --show      Show the current remote instead.",
			"-v --verbose   Show all remotes with --show.",
			"-h --help      Print the help text for a command.",
		},
	},
	"version": {
		"Show the build version.",
		"version",
		nil,
		[]string{
			"-h --help    Print the help text for a command.",
		},
	},
	"patch": {"New tag addition by    0.0.X.", "patch [--flags]", []string{"gitagger patch", "gitagger patch --pre rc", "gitagger patch -m \"fix login\" -n"}, tagFlags},
	"minor": {"New tag rounding to    0.X.0.", "minor [--flags]", []string{"gitagger minor", "gitagger minor --pre beta", "gitagger minor -d"}, tagFlags},
	"major": {"New tag rounding to    X.0.0.", "major [--flags]", []string{"gitagger major", "gitagger major --pre rc", "gitagger major --format date"}, tagFlags},
}

// KnownCommands lists every command for `gitagger help [command]`.
var KnownCommands = []string{"init", "check", "list", "ls", "help", "doctor", "remote", "version", "patch", "minor", "major"}

// IsCommand reports whether name is a known command.
func IsCommand(name string) bool {
	for _, c := range KnownCommands {
		if name == c {
			return true
		}
	}
	return false
}

// ExampleCount returns how many example lines a command documents.
func ExampleCount(cmd string) int {
	if cmd == "ls" {
		cmd = "list"
	}
	if h, ok := commandHelp[cmd]; ok {
		return len(h.examples)
	}
	return 0
}
