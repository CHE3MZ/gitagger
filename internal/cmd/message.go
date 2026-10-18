// Message command runner: config cmds generate the tag message.
package cmd

import (
	"bytes"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/CHE3MZ/gitagger/internal/run"
	"github.com/CHE3MZ/gitagger/internal/style"
)

// resolveMessage runs message commands in order and joins their stdout
// with newlines. Commands for other platforms are skipped; the rest see
// the hook environment plus GITAGGER_EVENT=message. A failing command
// aborts before the tag exists, so there is nothing to roll back.
func resolveMessage(o run.Options, plan run.Plan) (string, error) {
	env := hookEnv(o, plan, false)
	env["GITAGGER_EVENT"] = "message"
	var parts []string
	for _, c := range o.MessageCmds {
		if !OsMatches(c.OS, runtime.GOOS) {
			if o.Verbose {
				fmt.Println(style.Dim(fmt.Sprintf("message command skipped for %s: %s", c.OS, c.Cmd)))
			}
			continue
		}
		bin, prefix, err := resolveShell(c.Shell)
		if err != nil {
			return "", Generic("message command %v", err)
		}
		if o.Verbose {
			fmt.Println(style.Dim(fmt.Sprintf("message [%s]: %s", bin, c.Cmd)))
		}
		var out bytes.Buffer
		if err := runShellCommand(o.Dir, bin, prefix, c.Cmd, env, &out, os.Stderr); err != nil {
			return "", Generic("message command %q failed (%v)", c.Cmd, err)
		}
		parts = append(parts, strings.TrimRight(out.String(), "\r\n"))
	}
	msg := strings.TrimRight(strings.Join(parts, "\n"), "\r\n")
	if msg == "" {
		fmt.Println(style.Warn("message commands produced no output — creating a lightweight tag"))
	}
	return msg, nil
}
