// RunCheck validates the .gitagger.yml config file.
package cmd

import (
	"fmt"

	"github.com/CHE3MZ/gitagger/internal/config"
	"github.com/CHE3MZ/gitagger/internal/style"
)

// RunCheck reports whether dir's config is valid. Missing file is fine,
// defaults apply. With verbose it shows the resolved values.
func RunCheck(dir string, verbose bool) error {
	cfg, path, err := config.Load(dir)
	if err != nil {
		return BadArgs("invalid config: %s", err.Error())
	}
	if path == "" {
		fmt.Println(style.Dim("no .gitagger.yml — using defaults, config is valid."))
		return nil
	}
	fmt.Printf("%s %s\n", style.Green("config is valid:"), style.White(path))
	if verbose {
		fmt.Println(style.Dim(fmt.Sprintf("scale=%s pre=%s format=%s remote=%s push=%v force=%v require_clean=%v doctor=%v verbose=%v custom=%q",
			cfg.Scale, cfg.Pre, cfg.Format, cfg.Remote, cfg.Push, cfg.Force, cfg.RequireClean, cfg.Doctor, cfg.Verbose, cfg.Custom)))
	}
	return nil
}
