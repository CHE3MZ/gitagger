// RunTag orchestration: compute plan, create tag, push safely.
// All human printing lives here so main.go stays thin.
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/CHE3MZ/gitagger/internal/detect"
	"github.com/CHE3MZ/gitagger/internal/git"
	"github.com/CHE3MZ/gitagger/internal/run"
	"github.com/CHE3MZ/gitagger/internal/style"
)

// RunTag creates the next tag and pushes when safe.
func RunTag(o run.Options) error {
	o.Hooks = filterHooks(o.Hooks, o.Arguments)
	if !detect.ValidPreCheck(o.Pre) {
		return BadArgs("bad --pre %q (want stable|rc|beta|build|nightly)", o.Pre)
	}
	plan, err := run.ComputePlan(o)
	if err != nil {
		return mapPlanError(err)
	}
	if err := run.CheckHeadTagged(o, plan.Next); err != nil {
		return runFailureHooks(o, plan, false, Exists("%s", err.Error()))
	}

	// Doctor gate: with doctor: true, fail early on remote collision
	// instead of creating a tag that can't push.
	if o.Doctor && o.Push && !o.DryRun {
		if err := run.DoctorGate(o, plan.Next); err != nil {
			return runFailureHooks(o, plan, false, Exists("%s", err.Error()))
		}
		if o.Verbose {
			fmt.Println(style.Dim("doctor: pre-push check passed"))
		}
	}

	// Dirty-tree warning: warn but allow unless require_clean is set.
	// ComputePlan already aborted when RequireClean is set.
	if st, serr := git.StatusPorcelain(o.Dir); serr == nil && strings.TrimSpace(st) != "" {
		fmt.Println(style.Warn("working tree is dirty — tagging anyway"))
	}

	if o.DryRun {
		printDryRun(o, plan)
		return nil
	}

	greetPlan(o, plan)

	if err := RunHooks(o.Dir, "start", o.Hooks.Start, hookEnv(o, plan, false), o.Verbose, os.Stdout, os.Stderr); err != nil {
		return runFailureHooks(o, plan, false, err)
	}

	if err := git.CreateTag(o.Dir, plan.Next, o.Message, o.Force); err != nil {
		// Local-exists collision without -f is a no-op (exit 2).
		msg := strings.ToLower(err.Error())
		if strings.Contains(msg, "already exists") {
			return runFailureHooks(o, plan, false, Exists("tag %s already exists — use -f to overwrite", plan.Next))
		}
		return runFailureHooks(o, plan, false, Generic("couldn't create tag %s (%v)", plan.Next, err))
	}
	fmt.Printf("%s %s\n", style.Green("created tag"), style.BoldGreen(plan.Next))

	// Success hooks gate the push: a tag whose hooks fail is incomplete,
	// so it is rolled back and never pushed. (Push failures below still
	// keep the local tag — offline tagging must survive a dead remote.)
	if err := RunHooks(o.Dir, "success", o.Hooks.Success, hookEnv(o, plan, false), o.Verbose, os.Stdout, os.Stderr); err != nil {
		serr := Generic("tag %s created but success hook failed (%v)", plan.Next, err)
		rollbackTag(o, plan)
		return runFailureHooks(o, plan, false, serr)
	}

	if !o.Push {
		fmt.Println(style.Dim("kept locally (-n). push later with `git push " + o.Remote + " " + plan.Next + "`"))
		return runFinishHooks(o, plan, false)
	}

	outcome, err := run.EnsurePush(o, plan.Next)
	if err != nil {
		msg := err.Error()
		if strings.Contains(msg, "already exists on remote") {
			return runFailureHooks(o, plan, false, Exists("%s", msg))
		}
		return runFailureHooks(o, plan, false, GenericErr(err))
	}
	if outcome.Pushed {
		fmt.Printf("%s %s to %s\n", style.Green("pushed"), style.BoldGreen(plan.Next), style.White(o.Remote))
	} else {
		fmt.Println(style.Warn(outcome.Skipped))
	}
	return runFinishHooks(o, plan, outcome.Pushed)
}

// rollbackTag undoes a tag this run created after its success hooks
// failed, so broken tags are never pushed or kept. When -f overwrote
// an existing tag, the previous target is restored instead.
func rollbackTag(o run.Options, plan run.Plan) {
	if plan.PrevRef != "" {
		if err := git.UpdateRef(o.Dir, "refs/tags/"+plan.Next, plan.PrevRef); err != nil {
			fmt.Println(style.Warn(fmt.Sprintf("couldn't restore tag %s (%v)", plan.Next, err)))
			return
		}
		fmt.Println(style.Dim(fmt.Sprintf("restored tag %s (rolled back)", plan.Next)))
		return
	}
	if err := git.DeleteTag(o.Dir, plan.Next); err != nil {
		fmt.Println(style.Warn(fmt.Sprintf("couldn't remove tag %s (%v)", plan.Next, err)))
		return
	}
	fmt.Println(style.Dim(fmt.Sprintf("removed tag %s (rolled back)", plan.Next)))
}

// hookEnv builds hook environment for the current plan.
func hookEnv(o run.Options, plan run.Plan, pushed bool) map[string]string {
	pushedStr := "false"
	if pushed {
		pushedStr = "true"
	}
	dryStr := "false"
	if o.TestMode {
		dryStr = "true"
	}
	return map[string]string{
		"GITAGGER_TAG":      plan.Next,
		"GITAGGER_PREV":     plan.Prev,
		"GITAGGER_REMOTE":   o.Remote,
		"GITAGGER_PUSHED":   pushedStr,
		"GITAGGER_DRY_RUN":  dryStr,
		"GITAGGER_ARGUMENT": strings.Join(o.Arguments, ","),
	}
}

// runFailureHooks runs failure then finish hooks, keeping the original error.
// Failure hooks only cover post-plan failures (pre-plan errors have no tag
// context). A failing failure/finish hook is reported, never re-triggered.
func runFailureHooks(o run.Options, plan run.Plan, pushed bool, err error) error {
	env := hookEnv(o, plan, pushed)
	if ferr := RunHooks(o.Dir, "failure", o.Hooks.Failure, env, o.Verbose, os.Stdout, os.Stderr); ferr != nil {
		fmt.Fprintln(os.Stderr, style.Error(fmt.Sprintf("failure hook failed: %v", ferr)))
	}
	if ferr := RunHooks(o.Dir, "finish", o.Hooks.Finish, env, o.Verbose, os.Stdout, os.Stderr); ferr != nil {
		fmt.Fprintln(os.Stderr, style.Error(fmt.Sprintf("finish hook failed: %v", ferr)))
	}
	return err
}

// runFinishHooks runs finish hooks only (success already ran).
// A failing finish hook is reported, never re-triggered.
func runFinishHooks(o run.Options, plan run.Plan, pushed bool) error {
	env := hookEnv(o, plan, pushed)
	if err := RunHooks(o.Dir, "finish", o.Hooks.Finish, env, o.Verbose, os.Stdout, os.Stderr); err != nil {
		return Generic("tag %s created but finish hook failed (%v)", plan.Next, err)
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
		fmt.Println(style.Dim(fmt.Sprintf("hooks: start:%d success:%d failure:%d finish:%d",
			len(o.Hooks.Start), len(o.Hooks.Success), len(o.Hooks.Failure), len(o.Hooks.Finish))))
	}
	if plan.Prev != "" {
		fmt.Printf("was:  %s\n", style.White(plan.Prev))
	}
	fmt.Printf("next: %s  %s\n", style.BoldGreen(plan.Next),
		style.Dim(fmt.Sprintf("(%s, %s)", plan.Format, o.Pre)))
	if o.Push {
		fmt.Println(style.Dim("would push to " + o.Remote))
	} else {
		fmt.Println(style.Dim("would keep locally (-n)"))
	}
}

// mapPlanError assigns exit codes: exists->2, bad args->3, else 1.
func mapPlanError(err error) error {
	msg := err.Error()
	lower := strings.ToLower(msg)
	switch {
	case strings.Contains(lower, "already exists"),
		strings.Contains(lower, "already tagged"),
		strings.Contains(lower, "already on"),
		strings.Contains(lower, "nothing to do"):
		return Exists("%s", msg)
	case strings.Contains(lower, "unknown pre"),
		strings.Contains(lower, "still auto"),
		strings.Contains(lower, "unsupported format"),
		strings.Contains(lower, "bad scale"),
		strings.Contains(lower, "bad pre"),
		strings.Contains(lower, "bad format"),
		strings.Contains(lower, "remote must not"),
		strings.Contains(lower, "custom template"):
		return BadArgs("%s", msg)
	default:
		return GenericErr(err)
	}
}
