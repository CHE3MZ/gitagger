// RunNext prints the tag the next real run would create. Read-only:
// nothing is tagged, pushed, or written.
package cmd

import (
	"fmt"

	"github.com/CHE3MZ/gitagger/internal/run"
	"github.com/CHE3MZ/gitagger/internal/style"
)

// NextTag returns the tag the next real run would create (no side effects).
func NextTag(o run.Options) (string, error) {
	plan, err := run.ComputePlan(o)
	if err != nil {
		return "", mapPlanError(err)
	}
	return plan.Next, nil
}

// RunNext prints just that tag, so scripts can grab it. With verbose it
// shows how the prediction was computed instead.
func RunNext(o run.Options) error {
	plan, err := run.ComputePlan(o)
	if err != nil {
		return mapPlanError(err)
	}
	if o.Verbose {
		fmt.Println(style.Dim(fmt.Sprintf("detected: %s, v-prefix=%v (from history)", plan.Format, plan.VPrefix)))
		if plan.Prev != "" {
			fmt.Println(style.Dim("was: " + plan.Prev))
		} else {
			fmt.Println(style.Dim("no tags yet — starting fresh"))
		}
		fmt.Println(style.Dim("next: " + plan.Next))
		return nil
	}
	fmt.Println(plan.Next)
	return nil
}
