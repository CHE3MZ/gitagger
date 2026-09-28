// Hooks runner: executes config lifecycle blocks in order.
package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/CHE3MZ/gitagger/internal/config"
	"github.com/CHE3MZ/gitagger/internal/style"
)

// shellRunner maps shell names to binary + leading args.
func shellRunner(shell string) (string, []string, error) {
	switch shell {
	case "", "sh":
		return "sh", []string{"-c"}, nil
	case "bash":
		return "bash", []string{"-c"}, nil
	case "pwsh":
		return "pwsh", []string{"-Command"}, nil
	case "batch":
		if runtime.GOOS != "windows" {
			return "", nil, fmt.Errorf("shell %q needs Windows", shell)
		}
		return "cmd", []string{"/C"}, nil
	default:
		return "", nil, fmt.Errorf("unknown shell %q (want sh|bash|pwsh|batch)", shell)
	}
}

// OsMatches reports whether an os: selector targets a GOOS value.
// The selector is one platform or a comma list ("macos, linux").
// "macos" is accepted for Go's "darwin"; empty means all platforms.
func OsMatches(selector, goos string) bool {
	matched := false
	for _, p := range strings.Split(selector, ",") {
		p = strings.ToLower(strings.TrimSpace(p))
		if p == "" {
			continue
		}
		matched = true
		if p == "macos" && goos == "darwin" {
			return true
		}
		if p == goos {
			return true
		}
	}
	return !matched
}
// filterHooks keeps untagged blocks plus those matching given arguments
// (case-insensitive). Runs without -a see untagged blocks only.
func filterHooks(h config.Hooks, args []string) config.Hooks {
	want := map[string]bool{}
	for _, a := range args {
		want[strings.ToLower(a)] = true
	}
	keep := func(blocks []config.HookBlock) []config.HookBlock {
		var out []config.HookBlock
		for _, b := range blocks {
			if b.Argument == "" || want[strings.ToLower(b.Argument)] {
				out = append(out, b)
			}
		}
		return out
	}
	h.Start = keep(h.Start)
	h.Success = keep(h.Success)
	h.Failure = keep(h.Failure)
	h.Finish = keep(h.Finish)
	return h
}

// RunHooks runs every block for event in order, streaming output live.
// Blocks for other platforms are skipped. env values are exported to
// each command alongside GITAGGER_EVENT.
func RunHooks(dir, event string, blocks []config.HookBlock, env map[string]string, verbose bool, stdout, stderr io.Writer) error {
	full := map[string]string{"GITAGGER_EVENT": event}
	for k, v := range env {
		full[k] = v
	}
	for _, b := range blocks {
		shell := strings.ToLower(strings.TrimSpace(b.Shell))
		if shell == "" {
			shell = "sh"
		}
		osName := strings.ToLower(strings.TrimSpace(b.OS))
		if !OsMatches(osName, runtime.GOOS) {
			if verbose {
				first, rest := "", b.Run
				if len(rest) > 0 {
					first = rest[0]
				}
				_, _ = fmt.Fprintln(stdout, style.Dim(fmt.Sprintf("hook skipped for %s: %s", osName, first)))
			}
			continue
		}
		bin, prefix, err := shellRunner(shell)
		if err != nil {
			return GenericErr(err)
		}
		if _, err := exec.LookPath(bin); err != nil {
			return Generic("hook needs %q (%s not on PATH)", shell, bin)
		}
		for _, c := range b.Run {
			_, _ = fmt.Fprintln(stdout, style.Dim(fmt.Sprintf("run [%s]: %s", shell, c)))
			args := append(append([]string{}, prefix...), c)
			// #nosec G204 — binary is a fixed shell from the allowlist
			// above, never user input; the command string goes to the
			// shell as data, exactly like running it by hand.
			cmd := exec.Command(bin, args...)
			cmd.Dir = dir
			cmd.Env = append(os.Environ(), mapEnv(full)...)
			cmd.Stdout = stdout
			cmd.Stderr = stderr
			if err := cmd.Run(); err != nil {
				return Generic("hook %q failed (%v)", c, err)
			}
		}
	}
	return nil
}

func mapEnv(env map[string]string) []string {
	out := make([]string, 0, len(env))
	for k, v := range env {
		out = append(out, k+"="+v)
	}
	return out
}
