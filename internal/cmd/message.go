// Message command runner: config cmds generate the tag message.
package cmd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
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
		bin, prefix, err := shellRunner(c.Shell)
		if err != nil {
			return "", GenericErr(err)
		}
		if _, err := exec.LookPath(bin); err != nil {
			return "", Generic("message command needs %q (%s not on PATH)", c.Shell, bin)
		}
		if o.Verbose {
			fmt.Println(style.Dim(fmt.Sprintf("message [%s]: %s", c.Shell, c.Cmd)))
		}
		args := append(append([]string{}, prefix...), c.Cmd)
		// #nosec G204 — binary is a fixed shell from the allowlist
		// above, never user input; the command string goes to the
		// shell as data, exactly like running it by hand.
		cmd := exec.Command(bin, args...)
		cmd.Dir = o.Dir
		cmd.Env = append(os.Environ(), mapEnv(env)...)
		var out bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
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
