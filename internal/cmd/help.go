// Help text. The main text is the owner's words, verbatim.
// Per-command help follows the same direct style.
package cmd

import "fmt"

// PrintHelp prints the main help text.
func PrintHelp() {
	fmt.Print(mainHelp)
}

const mainHelp = `automate git tags without the headache.

USAGE

  gitagger [command] [--flags]

EXAMPLES
  gitagger                   # create a new tag and push
  gitagger patch             # create a new patch tag
  gitagger minor             # create a new minor tag
  gitagger major             # create a new major tag
  gitagger major --pre beta  # create a new major beta tag

COMMANDS
  init       [--flags]       Generates a .gitagger.yml file
  check      [--flags]       Checks if your .gitagger.yml config is valid
  list       [--flags]       Shows all tags
  help       [--flags]       Print this help text
  doctor     [--flags]       Audit local vs remote tag health status
  remote     [--flags]       View the remote URL for this git project
  patch      [--flags]       New tag addition by    0.0.X
  minor      [--flags]       New tag rounding to    0.X.0
  major      [--flags]       New tag rounding to    X.0.0

Flags
  --pre          stable|rc|beta|build|nightly                    prerelease flavor (default stable)
  --format       auto|triple|double|single|date|sha|sha-num      force a style (default auto)
  -h --help      Print the help text for a command.
  -v --verbose   Enable verbose mode.
  -n --no-push   Create tag without pushing.
  -d --dry-run   Do A dry-run for testing.
  -f --force     Force push to remote.
`

// CommandHelp prints help for one command in the same direct style.
func CommandHelp(cmd string) {
	if h, ok := commandHelp[cmd]; ok {
		fmt.Print(h)
		return
	}
	PrintHelp()
}

var commandHelp = map[string]string{
	"init": `Generates a .gitagger.yml file.

USAGE
  gitagger init [-f]

FLAGS
  -f --force   Overwrite the config file if it exists.
  -h --help    Print the help text for a command.
`,
	"check": `Checks if your .gitagger.yml config is valid.

USAGE
  gitagger check [-v]

FLAGS
  -v --verbose   Show the resolved config values.
  -h --help      Print the help text for a command.
`,
	"list": `Shows all tags.

USAGE
  gitagger list [-v]

FLAGS
  -v --verbose   Also show the remote URL.
  -h --help      Print the help text for a command.
`,
	"help": `Print this help text.

USAGE
  gitagger help [command]
`,
	"doctor": `Audit local vs remote tag health status.

USAGE
  gitagger doctor [-v]

FLAGS
  -v --verbose   Also show the remote URL.
  -h --help      Print the help text for a command.
`,
	"remote": `View the remote URL for this git project.

USAGE
  gitagger remote [-v]

FLAGS
  -v --verbose   Show all remotes.
  -h --help      Print the help text for a command.
`,
	"patch": `New tag addition by    0.0.X.

USAGE
  gitagger patch [--flags]

FLAGS
  --pre          stable|rc|beta|build|nightly                    prerelease flavor (default stable)
  --format       auto|triple|double|single|date|sha|sha-num      force a style (default auto)
  -n --no-push   Create tag without pushing.
  -d --dry-run   Do A dry-run for testing.
  -f --force     Force push to remote.
  -v --verbose   Enable verbose mode.
  -h --help      Print the help text for a command.
`,
	"minor": `New tag rounding to    0.X.0.

USAGE
  gitagger minor [--flags]

FLAGS
  --pre          stable|rc|beta|build|nightly                    prerelease flavor (default stable)
  --format       auto|triple|double|single|date|sha|sha-num      force a style (default auto)
  -n --no-push   Create tag without pushing.
  -d --dry-run   Do A dry-run for testing.
  -f --force     Force push to remote.
  -v --verbose   Enable verbose mode.
  -h --help      Print the help text for a command.
`,
	"major": `New tag rounding to    X.0.0.

USAGE
  gitagger major [--flags]

FLAGS
  --pre          stable|rc|beta|build|nightly                    prerelease flavor (default stable)
  --format       auto|triple|double|single|date|sha|sha-num      force a style (default auto)
  -n --no-push   Create tag without pushing.
  -d --dry-run   Do A dry-run for testing.
  -f --force     Force push to remote.
  -v --verbose   Enable verbose mode.
  -h --help      Print the help text for a command.
`,
}

// KnownCommands lists every command for `gitagger help [command]`.
var KnownCommands = []string{"init", "check", "list", "help", "doctor", "remote", "patch", "minor", "major"}

// IsCommand reports whether name is a known command.
func IsCommand(name string) bool {
	for _, c := range KnownCommands {
		if name == c {
			return true
		}
	}
	return false
}
