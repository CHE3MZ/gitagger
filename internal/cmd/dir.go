// ResolveDir picks the operating directory: -p flag > config path: > cwd.
// Relative paths start from cwd. The config file itself always comes from
// the -p/cwd directory: path: redirects operations, not settings.
// Prints where it goes when redirected.
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/CHE3MZ/gitagger/internal/config"
	"github.com/CHE3MZ/gitagger/internal/style"
)

// ResolveDir resolves the operating directory for a command.
func ResolveDir(cwd, flagPath string) (string, error) {
	target := cwd
	if strings.TrimSpace(flagPath) != "" {
		target = flagPath
	} else if cfg, _, err := config.Load(cwd); err != nil {
		return "", BadArgs("%s", err.Error())
	} else if strings.TrimSpace(cfg.Path) != "" {
		target = cfg.Path
	}
	if !filepath.IsAbs(target) {
		target = filepath.Join(cwd, target)
	}
	abs, err := filepath.Abs(target)
	if err != nil {
		return "", BadArgs("no such directory %q", target)
	}
	st, err := os.Stat(abs)
	if err != nil || !st.IsDir() {
		return "", BadArgs("no such directory %q", target)
	}
	if abs != cwd {
		fmt.Println(style.Dim("in " + abs))
	}
	return abs, nil
}
