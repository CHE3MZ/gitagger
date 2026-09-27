// Custom template tests: validation, render matrix, and end to end runs.
// Temp repos only — never the repo's own .git.
package tests

import (
	"testing"
	"time"

	"github.com/CHE3MZ/gitagger/internal/cmd"
	"github.com/CHE3MZ/gitagger/internal/config"
	"github.com/CHE3MZ/gitagger/internal/detect"
	"github.com/CHE3MZ/gitagger/internal/git"
	"github.com/CHE3MZ/gitagger/internal/next"
	"github.com/CHE3MZ/gitagger/internal/run"
)

func customReq(prev, scale, pre, tmpl string) next.Request {
	return next.Request{
		Prev: prev, Scale: scale, Pre: pre, Format: detect.Custom,
		VPrefix: true, Now: time.Date(2026, 9, 28, 12, 0, 0, 0, time.UTC),
		SHA: "a1b2c3d", FullSHA: "a1b2c3d4e5f6", Count: "42",
		Custom: tmpl, Existing: map[string]bool{},
	}
}

func TestCustomValidation(t *testing.T) {
	good := []string{
		"build-<SHA>-v<MAJOR>.<MINOR>.<PATCH>-<DATE>-<PRE>",
		"v<MAJOR>.<MINOR>.<PATCH>",
		"r-<TAG>-<NUMBER>",
		"<YEAR><MONTH><DAY>-<COUNT>-<FULLSHA>",
	}
	for _, tmpl := range good {
		if err := next.ValidateTemplate(tmpl); err != nil {
			t.Errorf("ValidateTemplate(%q) = %v, want nil", tmpl, err)
		}
	}
	bad := []string{
		"",
		"   ",
		"hello",
		"build-<SHA",
		"build-SHA>",
		"a-<FOO>",
		"a-<sha>",
		"a<>",
		"has space",
		"a/b",
	}
	for _, tmpl := range bad {
		if err := next.ValidateTemplate(tmpl); err == nil {
			t.Errorf("ValidateTemplate(%q) should fail", tmpl)
		}
	}
}

func TestCustomRenderMatrix(t *testing.T) {
	cases := []struct {
		name, prev, scale, pre, tmpl, want string
	}{
		{"triple patch", "v1.2.3", "patch", "stable", "v<MAJOR>.<MINOR>.<PATCH>", "v1.2.4"},
		{"triple minor", "v1.2.3", "minor", "stable", "v<MAJOR>.<MINOR>.<PATCH>", "v1.3.0"},
		{"triple major", "v1.2.3", "major", "stable", "v<MAJOR>.<MINOR>.<PATCH>", "v2.0.0"},
		{"user standard", "build-2f88688-v1.2.3-2026.09.27", "patch", "stable",
			"build-<SHA>-v<MAJOR>.<MINOR>.<PATCH>-<DATE><PRE>", "build-a1b2c3d-v1.2.4-2026.09.28"},
		{"user standard rc", "build-2f88688-v1.2.3-2026.09.27", "minor", "rc",
			"build-<SHA>-v<MAJOR>.<MINOR>.<PATCH>-<DATE><PRE>", "build-a1b2c3d-v1.3.0-2026.09.28-rc"},
		{"fresh starts at 1.0.0", "", "major", "stable", "v<MAJOR>.<MINOR>.<PATCH>", "v1.0.0"},
		{"no version tokens ignores scale", "whatever", "major", "stable",
			"build-<SHA>-<DATE>", "build-a1b2c3d-2026.09.28"},
		{"all tokens", "v1.2.3", "minor", "beta",
			"x-<YEAR>.<MONTH>.<DAY>-<COUNT>-<FULLSHA>-<TAG><PRE>", "x-2026.09.28-42-a1b2c3d4e5f6-v1.2.3-beta"},
		{"tag raw kept", "release-9", "patch", "stable", "r-<TAG>-f", "r-release-9-f"},
	}
	for _, c := range cases {
		got, err := next.Compute(customReq(c.prev, c.scale, c.pre, c.tmpl))
		if err != nil {
			t.Errorf("%s: error %v", c.name, err)
			continue
		}
		if got != c.want {
			t.Errorf("%s = %q, want %q", c.name, got, c.want)
		}
	}
}

func TestCustomNeedsVersionInPrev(t *testing.T) {
	_, err := next.Compute(customReq("master-abc1234", "patch", "stable", "v<MAJOR>.<MINOR>.<PATCH>"))
	if err == nil {
		t.Fatalf("version tokens with versionless prev should fail")
	}
}

func TestCustomNumberCollision(t *testing.T) {
	req := customReq("", "patch", "stable", "t-<SHA>-<NUMBER>")
	req.Existing = map[string]bool{"t-a1b2c3d-": true}
	got, err := next.Compute(req)
	if err != nil || got != "t-a1b2c3d-1" {
		t.Fatalf("got %q,%v want t-a1b2c3d-1", got, err)
	}
	// No NUMBER token and a clash: hard error, not a guess.
	req2 := customReq("", "patch", "stable", "t-<SHA>")
	req2.Existing = map[string]bool{"t-a1b2c3d": true}
	if _, err := next.Compute(req2); err == nil {
		t.Fatalf("clash without NUMBER should fail")
	}
}

func TestCustomEndToEnd(t *testing.T) {
	dir := seedRepo(t, []string{"v1.2.2", "v1.2.3"})
	short, err := git.ShortSHA(dir)
	if err != nil {
		t.Fatal(err)
	}
	today := time.Now().Format("2006.01.02")
	tmpl := "build-<SHA>-v<MAJOR>.<MINOR>.<PATCH>-<DATE>"
	plan, err := run.ComputePlan(run.Options{
		Dir: dir, Scale: "patch", Pre: "stable",
		Format: detect.Custom, Custom: tmpl,
	})
	if err != nil {
		t.Fatal(err)
	}
	if want := "build-" + short + "-v1.2.4-" + today; plan.Next != want {
		t.Fatalf("next = %q, want %q", plan.Next, want)
	}
	if err := git.CreateTag(dir, plan.Next, "", false); err != nil {
		t.Fatal(err)
	}
}

func TestCustomNumberEndToEnd(t *testing.T) {
	dir := seedRepo(t, []string{"v1.2.3"})
	short, err := git.ShortSHA(dir)
	if err != nil {
		t.Fatal(err)
	}
	// Occupy the bare render so NUMBER has to kick in.
	if err := git.CreateTag(dir, "t-"+short+"-", "", false); err != nil {
		t.Fatal(err)
	}
	plan, err := run.ComputePlan(run.Options{
		Dir: dir, Scale: "patch", Pre: "stable",
		Format: detect.Custom, Custom: "t-<SHA>-<NUMBER>",
	})
	if err != nil {
		t.Fatal(err)
	}
	if plan.Next != "t-"+short+"-1" {
		t.Fatalf("next = %q, want t-%s-1", plan.Next, short)
	}
}

func TestCustomViaConfigFile(t *testing.T) {
	dir := t.TempDir()
	mustGitTempInit(t, dir)
	writeConfig(t, dir, "format: custom\ncustom: \"v<MAJOR>.<MINOR>.<PATCH>\"\n")
	cfg, _, err := config.Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	o, err := cmd.BuildOptions(dir, cfg, "", cmd.TagFlags{})
	if err != nil {
		t.Fatal(err)
	}
	if o.Format != detect.Custom || o.Custom != "v<MAJOR>.<MINOR>.<PATCH>" {
		t.Fatalf("options = %+v", o)
	}
	plan, err := run.ComputePlan(o)
	if err != nil {
		t.Fatal(err)
	}
	if plan.Next != "v1.0.0" {
		t.Fatalf("next = %q, want v1.0.0", plan.Next)
	}
}
