// BuildOptions merges scale + tag flags over config.
// Precedence: flags > config > defaults.
package cmd

import (
	"strings"

	"github.com/CHE3MZ/gitagger/internal/config"
	"github.com/CHE3MZ/gitagger/internal/detect"
	"github.com/CHE3MZ/gitagger/internal/run"
)

// BuildOptions resolves the effective run options for tag commands.
func BuildOptions(dir string, cfg config.Resolved, scale string, f TagFlags) (run.Options, error) {
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
			return o, BadArgs("bad --format %q (want auto|triple|double|single|date|sha|sha-num)", f.Format)
		}
		o.Format = ff
	}
	o.Push = cfg.Push
	if f.NoPush {
		o.Push = false
	}
	o.Force = cfg.Force || f.Force
	o.DryRun = f.DryRun
	o.Message = cfg.Message
	o.RequireClean = cfg.RequireClean
	o.Verbose = cfg.Verbose || f.Verbose
	return o, nil
}
