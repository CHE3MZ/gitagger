// Next-command tests: prediction only, never creates or pushes tags.
package tests

import (
	"testing"

	"github.com/CHE3MZ/gitagger/internal/cmd"
	"github.com/CHE3MZ/gitagger/internal/git"
	"github.com/CHE3MZ/gitagger/internal/run"
)

func nextOptions(dir, scale string) run.Options {
	return run.Options{Dir: dir, Scale: scale, Pre: "stable"}
}

func TestNextTagFreshRepo(t *testing.T) {
	dir := initRepo(t)
	tag, err := cmd.NextTag(nextOptions(dir, "patch"))
	if err != nil {
		t.Fatalf("next on fresh repo should pass: %v", err)
	}
	if tag != "v1.0.0" {
		t.Fatalf("fresh next = %q, want v1.0.0", tag)
	}
}

func TestNextTagBumpsAndOverrides(t *testing.T) {
	dir := initRepo(t)
	if err := git.CreateTag(dir, "v1.2.3", "", false); err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct{ scale, want string }{
		{"patch", "v1.2.4"},
		{"minor", "v1.3.0"},
		{"major", "v2.0.0"},
	} {
		tag, err := cmd.NextTag(nextOptions(dir, c.scale))
		if err != nil {
			t.Fatalf("next %s: %v", c.scale, err)
		}
		if tag != c.want {
			t.Errorf("next %s = %q, want %q", c.scale, tag, c.want)
		}
	}
}

func TestNextTagNeverTags(t *testing.T) {
	dir := initRepo(t)
	if err := git.CreateTag(dir, "v1.0.0", "", false); err != nil {
		t.Fatal(err)
	}
	tag, err := cmd.NextTag(nextOptions(dir, "patch"))
	if err != nil {
		t.Fatal(err)
	}
	if tag != "v1.0.1" {
		t.Fatalf("next = %q, want v1.0.1", tag)
	}
	if tags := tagsOf(t, dir); len(tags) != 1 || tags[0] != "v1.0.0" {
		t.Fatalf("next must not create tags, got %v", tags)
	}
}

func TestNextTagNoCommits(t *testing.T) {
	dir := t.TempDir()
	mustGit(t, dir, "init")
	if _, err := cmd.NextTag(nextOptions(dir, "patch")); err == nil {
		t.Fatalf("next without commits should fail")
	}
}

func TestRunNextPrints(t *testing.T) {
	dir := initRepo(t)
	if err := cmd.RunNext(nextOptions(dir, "patch")); err != nil {
		t.Fatalf("plain next should pass: %v", err)
	}
	o := nextOptions(dir, "patch")
	o.Verbose = true
	if err := cmd.RunNext(o); err != nil {
		t.Fatalf("verbose next should pass: %v", err)
	}
}

func TestNextHelp(t *testing.T) {
	if !cmd.IsCommand("next") {
		t.Fatalf("IsCommand(next) = false")
	}
	cmd.CommandHelp("next") // must not crash
	if n := cmd.ExampleCount("next"); n == 0 || n > 3 {
		t.Fatalf("next has %d examples, want 1-3", n)
	}
}
