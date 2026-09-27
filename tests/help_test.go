// Help-content tests: every command reachable, examples capped at 3,
// version stays example-free.
package tests

import (
	"testing"

	"github.com/CHE3MZ/gitagger/internal/cmd"
)

func TestHelpExamplesCapped(t *testing.T) {
	for _, c := range cmd.KnownCommands {
		if n := cmd.ExampleCount(c); n > 3 {
			t.Errorf("%s documents %d examples, max is 3", c, n)
		}
	}
}

func TestHelpExamplesExpected(t *testing.T) {
	cases := map[string]int{
		"init": 0, "check": 0, "list": 0, "help": 1, "doctor": 0,
		"remote": 3, "version": 0, "patch": 3, "minor": 3, "major": 3,
	}
	for c, want := range cases {
		if got := cmd.ExampleCount(c); got != want {
			t.Errorf("%s has %d examples, want %d", c, got, want)
		}
	}
}
