// Numeration tests: pure math, no git, no filesystem except temp dirs.
// Safe to run anywhere — never touches the repo's own .git.
package tests

import (
	"testing"
	"time"

	"github.com/CHE3MZ/gitagger/internal/detect"
	"github.com/CHE3MZ/gitagger/internal/next"
)

func TestClassify(t *testing.T) {
	cases := map[string]detect.Format{
		"v1.2.3":             detect.Triple,
		"1.2.3":              detect.Triple,
		"v1.2.3-rc":          detect.Triple,
		"v1.2":               detect.Double,
		"v1":                 detect.Single,
		"v2026.09.26":        detect.Date,
		"v2026.09.26.1":      detect.Date,
		"master-2f88688":     detect.SHA,
		"master-1234-2f88688": detect.SHANum,
		"hello":              detect.Unknown,
		"":                   detect.Unknown,
	}
	for tag, want := range cases {
		if got := detect.Classify(tag); got != want {
			t.Errorf("Classify(%q) = %q, want %q", tag, got, want)
		}
	}
}

func TestDetectMajorityAndVPrefix(t *testing.T) {
	tags := []string{"v1.2.3", "v1.2.4", "v2026.09.26"} // oldest -> newest
	f, v := detect.Detect(tags)
	if f != detect.Triple {
		t.Errorf("majority = %q, want triple", f)
	}
	if !v {
		t.Errorf("vPrefix = false, want true")
	}
}

func TestDetectTieGoesToNewest(t *testing.T) {
	tags := []string{"v1.2.3", "v2026.09.26"}
	f, _ := detect.Detect(tags)
	if f != detect.Date {
		t.Errorf("tie = %q, want date (newest)", f)
	}
}

func TestDetectEmptyFallsBackToTriple(t *testing.T) {
	f, v := detect.Detect(nil)
	if f != detect.Triple || !v {
		t.Errorf("empty = %q,%v want triple,true", f, v)
	}
}

func TestDetectKeepsPrefixless(t *testing.T) {
	tags := []string{"1.2.3", "1.2.4"}
	_, v := detect.Detect(tags)
	if v {
		t.Errorf("vPrefix = true, want false for prefixless history")
	}
}

func TestTripleBumps(t *testing.T) {
	cases := []struct {
		prev, scale, pre, want string
		vPrefix                bool
	}{
		{"v1.2.3", "patch", "stable", "v1.2.4", true},
		{"v1.2.3", "minor", "stable", "v1.3.0", true},
		{"v1.2.3", "major", "stable", "v2.0.0", true},
		{"1.2.3", "patch", "stable", "1.2.4", false},
		{"v1.2.3", "patch", "rc", "v1.2.4-rc", true},
		{"v1.2.4-rc", "patch", "stable", "v1.2.4", true}, // promote
		{"v1.2.4-rc", "patch", "beta", "v1.2.4-beta", true}, // pre swap, no bump
	}
	for _, c := range cases {
		got, err := next.Compute(next.Request{
			Prev: c.prev, Scale: c.scale, Pre: c.pre,
			Format: detect.Triple, VPrefix: c.vPrefix,
		})
		if err != nil {
			t.Errorf("Compute(%v) error: %v", c, err)
			continue
		}
		if got != c.want {
			t.Errorf("Compute(%s %s %s) = %q, want %q", c.prev, c.scale, c.pre, got, c.want)
		}
	}
}

func TestDoubleAndSingle(t *testing.T) {
	got, _ := next.Compute(next.Request{Prev: "v1.2", Scale: "patch", Pre: "stable", Format: detect.Double, VPrefix: true})
	if got != "v1.3" {
		t.Errorf("double patch = %q, want v1.3", got)
	}
	got, _ = next.Compute(next.Request{Prev: "v1.2", Scale: "major", Pre: "stable", Format: detect.Double, VPrefix: true})
	if got != "v2.0" {
		t.Errorf("double major = %q, want v2.0", got)
	}
	got, _ = next.Compute(next.Request{Prev: "v1", Scale: "patch", Pre: "stable", Format: detect.Single, VPrefix: true})
	if got != "v2" {
		t.Errorf("single patch = %q, want v2", got)
	}
}

func TestDateCollisionCounter(t *testing.T) {
	now := time.Date(2026, 9, 26, 12, 0, 0, 0, time.UTC)
	req := next.Request{Format: detect.Date, VPrefix: true, Pre: "stable", Now: now}
	got, _ := next.Compute(req)
	if got != "v2026.09.26" {
		t.Fatalf("date fresh = %q, want v2026.09.26", got)
	}
	req.Existing = map[string]bool{"v2026.09.26": true}
	got, _ = next.Compute(req)
	if got != "v2026.09.26.1" {
		t.Fatalf("date collision = %q, want v2026.09.26.1", got)
	}
}

func TestFreshRepoStartsAtTriple(t *testing.T) {
	got, err := next.Compute(next.Request{Format: detect.Triple, VPrefix: true, Scale: "patch", Pre: "stable"})
	if err != nil || got != "v0.0.1" {
		t.Errorf("fresh = %q,%v want v0.0.1", got, err)
	}
}
