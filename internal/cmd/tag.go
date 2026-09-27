// RunTag orchestration: compute plan, create tag, push safely.
// All human printing lives here so main.go stays thin.
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/CHE3MZ/gitagger/internal/detect"
	"github.com/CHE3MZ/gitagger/internal/git"
	"github.com/CHE3MZ/gitagger/internal/run"
	"github.com/CHE3MZ/gitagger/internal/style"
)

// RunTag creates the next tag and pushes when safe.
// JSON mode suppresses human lines and emits one JSON object.
func RunTag(o run.Options) error {
	if !detect.ValidPreCheck(o.Pre) {
		return BadArgs("bad --pre %q (want stable|rc|beta|build|nightly)", o.Pre)
	}
	if o.Format == detect.Custom && strings.TrimSpace(o.Custom) == "" {
		return BadArgs("format is custom but --custom is empty (v2 feature anyway)")
	}

	plan, err := run.ComputePlan(o)
	if err != nil {
		return mapPlanError(err)
	}
	if err := run.CheckHeadTagged(o, plan.Next); err != nil {
		return Exists("%s", err.Error())
	}

	// Dirty-tree warning (plan §2): warn but allow unless --require-clean.
	// ComputePlan already aborted when RequireClean is set.
	if st, serr := git.StatusPorcelain(o.Dir); serr == nil && strings.TrimSpace(st) != "" {
		if !o.JSON {
			fmt.Println(style.Warn("working tree is dirty — tagging anyway (use --require-clean to block)"))
		}
	}

	if o.DryRun {
		if o.JSON {
			return printPlanJSON(plan.Prev, plan.Next, false, o.Remote)
		}
		printDryRun(o, plan)
		return nil
	}

	if o.JSON {
		style.SetEnabled(false)
	}

	if !o.JSON {
		greetPlan(o, plan)
	}

	if err := git.CreateTag(o.Dir, plan.Next, o.Message, o.Force); err != nil {
		// Local-exists collision without --force is a no-op (exit 2).
		msg := strings.ToLower(err.Error())
		if strings.Contains(msg, "already exists") {
			return Exists("tag %s already exists — use --force to overwrite", plan.Next)
		}
		return Generic("couldn't create tag %s (%v)", plan.Next, err)
	}
	if !o.JSON {
		fmt.Printf("%s %s\n", style.Green("created tag"), style.BoldGreen(plan.Next))
	}

	if !o.Push {
		if o.JSON {
			return printPlanJSON(plan.Prev, plan.Next, false, o.Remote)
		}
		fmt.Println(style.Dim("kept locally (--no-push). push later with `git push " + o.Remote + " " + plan.Next + "`"))
		return nil
	}

	if o.Confirm {
		if ShouldPrompt(true, o.Push, o.DryRun) {
			if !AskYes(fmt.Sprintf("push %s to %s?", plan.Next, o.Remote)) {
				if o.JSON {
					return printPlanJSON(plan.Prev, plan.Next, false, o.Remote)
				}
				fmt.Println(style.Dim("kept locally. push later with `git push " + o.Remote + " " + plan.Next + "`"))
				return nil
			}
		}
		// Non-TTY + --confirm: proceed without prompting (CI-friendly).
	}

	outcome, err := run.EnsurePush(o, plan.Next)
	if err != nil {
		msg := err.Error()
		if strings.Contains(msg, "already exists on remote") {
			return Exists("%s", msg)
		}
		return GenericErr(err)
	}
	if o.JSON {
		return printPlanJSON(plan.Prev, plan.Next, outcome.Pushed, o.Remote)
	}
	if outcome.Pushed {
		fmt.Printf("%s %s to %s\n", style.Green("pushed"), style.BoldGreen(plan.Next), style.White(o.Remote))
	} else {
		fmt.Println(style.Warn(outcome.Skipped))
	}
	return nil
}

func greetPlan(o run.Options, plan run.Plan) {
	if o.Verbose {
		fmt.Println(style.Dim(fmt.Sprintf("detected: %s, v-prefix=%v (from history)", plan.Format, plan.VPrefix)))
		if plan.Prev != "" {
			fmt.Println(style.Dim("was: " + plan.Prev))
		} else {
			fmt.Println(style.Dim("no tags yet — starting fresh"))
		}
		fmt.Println(style.Dim("next: " + plan.Next))
	}
}

func printDryRun(o run.Options, plan run.Plan) {
	fmt.Println(style.Header("dry run — nothing created"))
	if o.Verbose {
		fmt.Println(style.Dim(fmt.Sprintf("detected: %s, v-prefix=%v (from history)", plan.Format, plan.VPrefix)))
	}
	if plan.Prev != "" {
		fmt.Printf("was:  %s\n", style.White(plan.Prev))
	}
	fmt.Printf("next: %s  %s\n", style.BoldGreen(plan.Next),
		style.Dim(fmt.Sprintf("(%s, %s)", plan.Format, o.Pre)))
	if o.Push {
		fmt.Println(style.Dim("would push to " + o.Remote))
	} else {
		fmt.Println(style.Dim("would keep locally (--no-push)"))
	}
}

func printPlanJSON(prev, next string, pushed bool, remote string) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(map[string]any{
		"prev": prev, "next": next, "pushed": pushed, "remote": remote,
	})
}

// mapPlanError assigns exit codes: exists->2, bad args->3, else 1.
func mapPlanError(err error) error {
	msg := err.Error()
	lower := strings.ToLower(msg)
	switch {
	case strings.Contains(lower, "already exists"),
		strings.Contains(lower, "already tagged"),
		strings.Contains(lower, "nothing to do"):
		return Exists("%s", msg)
	case strings.Contains(lower, "unknown pre"),
		strings.Contains(lower, "still auto"),
		strings.Contains(lower, "unsupported format"),
		strings.Contains(lower, "bad scale"),
		strings.Contains(lower, "bad pre"),
		strings.Contains(lower, "bad format"),
		strings.Contains(lower, "remote must not"),
		strings.Contains(lower, "custom template is empty"),
		strings.Contains(lower, "format is custom"):
		return BadArgs("%s", msg)
	default:
		return GenericErr(err)
	}
}
