// PrintHandbook prints the .gitagger.yml reference.
package cmd

import (
	"fmt"

	"github.com/CHE3MZ/gitagger/internal/style"
)

// RunHandbook prints the handbook. dir/verbose only exist so the
// dispatcher can treat handbook like every other simple command.
func RunHandbook(_ string, _ bool) error {
	PrintHandbook()
	return nil
}

func hbHeader(s string) {
	fmt.Println(style.Header(s))
	fmt.Println()
}

func hbKey(s string) {
	fmt.Println("     " + style.Green(s))
}

func hbDesc(s string) {
	fmt.Println("     " + style.Gray(s))
}

func hbExample(lines ...string) {
	fmt.Println("     " + style.Bold("EXAMPLE:"))
	for _, l := range lines {
		fmt.Println("         " + style.White(l))
	}
	fmt.Println()
}

// PrintHandbook prints every config key with its values and an example.
func PrintHandbook() {
	fmt.Println(style.Header("The Handbook for .gitagger.yml"))
	fmt.Println()
	fmt.Println(style.Gray("Everything is optional. Flags beat config, config beats defaults."))
	fmt.Println()

	hbHeader("SCALE")
	hbKey("scale: major | minor | patch")
	hbDesc("Tag size for triple/double/single styles. Default: patch.")
	hbExample("scale: minor", "scale: patch")

	hbHeader("PRE")
	hbKey("pre: stable | rc | beta | build | nightly")
	hbDesc("Flavor appended to the tag. stable = no suffix. Default: stable.")
	hbExample("pre: rc")

	hbHeader("FORMAT")
	hbKey("format: auto | triple | double | single | date | sha | sha-num | custom")
	hbDesc("Tag style. auto follows your tag history. Default: auto.")
	hbExample("format: date")

	hbHeader("CUSTOM")
	hbKey(`custom: "<template>"`)
	hbDesc("Custom template. Only used when format is custom. Default: empty (unused).")
	hbDesc("Tokens: MAJOR MINOR PATCH YEAR MONTH DAY DATE SHA FULLSHA COUNT PRE NUMBER TAG")
	hbExample(`custom: "build-<SHA>-v<MAJOR>.<MINOR>.<PATCH>-<DATE><PRE>"`)

	hbHeader("REMOTE")
	hbKey("remote: <name>")
	hbDesc("Push target. Must exist for auto-push. Default: origin.")
	hbExample("remote: upstream")

	hbHeader("PUSH")
	hbKey("push: true | false")
	hbDesc("Auto-push when reachable. false keeps tags local (same as -n). Default: true.")
	hbExample("push: false")

	hbHeader("FORCE")
	hbKey("force: true | false")
	hbDesc("Overwrite clashes and force-push. Same as -f. Default: false.")
	hbExample("force: true")

	hbHeader("REQUIRE_CLEAN")
	hbKey("require_clean: true | false")
	hbDesc("Abort instead of tagging a dirty tree. Default: false.")
	hbExample("require_clean: true")

	hbHeader("DOCTOR")
	hbKey("doctor: true | false")
	hbDesc("Pre-tag remote check. Aborts early on collision. Default: false.")
	hbExample("doctor: true")

	hbHeader("VERBOSE")
	hbKey("verbose: true | false")
	hbDesc("Show detection details. Same as -v. Default: false.")
	hbExample("verbose: true")

	hbHeader("MESSAGE")
	hbKey(`message: "<text>"`)
	hbDesc("Tag message. Empty = lightweight tag. Same as -m. Default: empty.")
	hbExample(`message: "rc1"`)

	hbHeader("PATH")
	hbKey(`path: "<dir>"`)
	hbDesc("Operate in another directory. Relative paths start where you run. Default: . (here).")
	hbExample(`path: "submodule/tool/"`)

	hbHeader("HOOKS")
	hbKey("on:")
	hbDesc("Shell commands at lifecycle events: start runs first, success after tagging, failure on errors, finish always.")
	hbExample(
		"on:",
		"  failure:",
		"    - shell: sh",
		"      run:",
		`        - echo "Oops! Something went wrong..."`,
	)
	hbDesc("shell: sh | bash | pwsh | batch. Default: sh.")
	hbDesc("os: linux | macos | windows. Default: all. Comma lists work.")
	hbDesc("run: one command or a list. Each item runs separately.")
	hbDesc("argument: <name>. Block runs only with gitagger -a <name>.")
	hbDesc("argument: none runs only when no arguments are given.")
	hbDesc("argument: any runs only when any argument is given.")
	hbDesc("Hooks get GITAGGER_TAG, GITAGGER_PREV, GITAGGER_REMOTE, GITAGGER_PUSHED, GITAGGER_EVENT, GITAGGER_ARGUMENT.")

	hbHeader("NOTES")
	fmt.Println(style.Gray("     Config files: ./.gitagger.yml, ./.gitagger.yaml, ./.gitagger (first found wins)."))
	fmt.Println(style.Gray("     Check yours: gitagger check. Validate verbosely: gitagger check -v."))
}
