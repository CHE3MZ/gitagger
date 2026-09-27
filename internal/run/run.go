// Package run ties config + detect + next + git together.
// main.go only parses flags and calls these helpers.
package run

import (
	"fmt"
	"strings"
	"time"

	"github.com/CHE3MZ/gitagger/internal/config"
	"github.com/CHE3MZ/gitagger/internal/detect"
	"github.com/CHE3MZ/gitagger/internal/git"
	"github.com/CHE3MZ/gitagger/internal/next"
)

// Options are effective settings after flags > config > defaults.
type Options struct {
	Dir          string
	Scale        string
	Pre          string
	Format       detect.Format
	Custom       string
	Remote       string
	Push         bool
	Confirm      bool
	Force        bool
	DryRun       bool
	Message      string
	RequireClean bool
	Verbose      bool
}

// FromConfig applies flag overrides onto loaded config.
func FromConfig(dir string, cfg config.Resolved, o Options) Options {
	o.Dir = dir
	if o.Scale == "" {
		o.Scale = cfg.Scale
	}
	if o.Pre == "" {
		o.Pre = cfg.Pre
	}
	if o.Format == "" {
		f, _ := detect.ParseFormat(cfg.Format)
		o.Format = f
	}
	if o.Custom == "" {
		o.Custom = cfg.Custom
	}
	if o.Remote == "" {
		o.Remote = cfg.Remote
	}
	// Push/Confirm/RequireClean/Message come from config only —
	// the CLI has no flags for them, BuildOptions copies the rest.
	return o
}

// Plan is the computed next tag.
type Plan struct {
	Prev     string
	Next     string
	Format   detect.Format
	VPrefix  bool
	Detected bool // true when format came from auto-detect
	Existing map[string]bool
}

// ComputePlan inspects tags and returns the next tag (no side effects).
func ComputePlan(o Options) (Plan, error) {
	if err := git.EnsureAvailable(); err != nil {
		return Plan{}, err
	}
	if !git.HeadExists(o.Dir) {
		return Plan{}, fmt.Errorf("no commits yet — make a commit first")
	}
	if o.RequireClean {
		st, err := git.StatusPorcelain(o.Dir)
		if err != nil {
			return Plan{}, err
		}
		if strings.TrimSpace(st) != "" {
			return Plan{}, fmt.Errorf("working tree is dirty (use --force flow or stash first)")
		}
	}
	tags, err := git.ListTags(o.Dir)
	if err != nil {
		return Plan{}, err
	}
	existing := map[string]bool{}
	for _, t := range tags {
		existing[t] = true
	}
	// Detect over the last 20 tags (most recent 20 by creatordate).
	detectInput := tags
	if len(detectInput) > 20 {
		detectInput = detectInput[len(detectInput)-20:]
	}
	format := o.Format
	detected := false
	var vPrefix bool
	if format == "" || format == detect.Auto {
		format, vPrefix = detect.Detect(detectInput)
		detected = true
	} else {
		// v-prefix still follows history when forcing a format
		_, vPrefix = detect.Detect(detectInput)
		if len(tags) == 0 {
			vPrefix = true
		}
	}
	var prev string
	if len(tags) > 0 {
		prev = tags[len(tags)-1] // oldest -> newest
	}
	// Gather sha info lazily for sha formats only.
	var sha, count string
	if format == detect.SHA || format == detect.SHANum {
		sha, _ = git.ShortSHA(o.Dir)
		count, _ = git.Count(o.Dir)
	}
	nextTag, err := next.Compute(next.Request{
		Prev: prev, Scale: o.Scale, Pre: o.Pre, Format: format,
		VPrefix: vPrefix, Now: time.Now(), Existing: existing,
		SHA: sha, Count: count,
	})
	if err != nil {
		return Plan{}, err
	}
	return Plan{Prev: prev, Next: nextTag, Format: format, VPrefix: vPrefix, Detected: detected, Existing: existing}, nil
}

// PushOutcome is the remote safety sequence result.
type PushOutcome struct {
	Pushed  bool
	Skipped string // human reason when not pushed
}

// EnsurePush runs the remote checks: configured? reachable? collision?
// It never deletes the local tag. Pushes exactly one ref.
func EnsurePush(o Options, tag string) (PushOutcome, error) {
	if !o.Push {
		return PushOutcome{Pushed: false, Skipped: "push disabled (--no-push)"}, nil
	}
	if _, err := git.RemoteURL(o.Dir, o.Remote); err != nil {
		return PushOutcome{Pushed: false, Skipped: fmt.Sprintf("no remote %q — tag kept locally only", o.Remote)}, nil
	}
	ls, err := git.LsRemoteTags(o.Dir, o.Remote)
	if err != nil {
		return PushOutcome{Pushed: false, Skipped: fmt.Sprintf("remote %q unreachable (offline?) — tag kept locally; run `git push %s %s` later", o.Remote, o.Remote, tag)}, nil
	}
	if git.RemoteHasTag(ls, tag) && !o.Force {
		return PushOutcome{}, fmt.Errorf("tag %s already exists on remote — use --force to overwrite or run `gitagger doctor`", tag)
	}
	if err := git.PushTag(o.Dir, o.Remote, tag, o.Force); err != nil {
		return PushOutcome{Pushed: false}, fmt.Errorf("push failed — tag %s kept locally (%v)", tag, err)
	}
	return PushOutcome{Pushed: true}, nil
}

// CheckHeadTagged aborts when HEAD already has any tag (unless --force).
func CheckHeadTagged(o Options, tag string) error {
	if o.Force {
		return nil
	}
	onHead, err := git.TagsPointingAtHEAD(o.Dir)
	if err != nil {
		return nil // don't block on introspection failure
	}
	if len(onHead) > 0 {
		return fmt.Errorf("already on %s, use --force to retag", strings.Join(onHead, ", "))
	}
	if existing, _ := git.ListTags(o.Dir); existing != nil {
		for _, t := range existing {
			if t == tag {
				return fmt.Errorf("tag %s already exists — use --force to overwrite", tag)
			}
		}
	}
	return nil
}
