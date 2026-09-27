// Format-matrix stress tests: every archetype under auto mode, mixed and
// all-types histories, pre-release flavors, and the full create-next loop.
// All end to end in temp repos — never the repo's own .git.
package tests

import (
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/CHE3MZ/gitagger/internal/detect"
	"github.com/CHE3MZ/gitagger/internal/git"
	"github.com/CHE3MZ/gitagger/internal/run"
)

// commitDated makes an empty commit stamped with an explicit date so tag
// order (creatordate) is deterministic without sleeping.
func commitDated(t *testing.T, dir, datestr string) {
	t.Helper()
	cmd := exec.Command("git", "-C", dir, "commit", "--allow-empty", "-m", "seed")
	cmd.Env = append(os.Environ(), "GIT_AUTHOR_DATE="+datestr, "GIT_COMMITTER_DATE="+datestr)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git commit: %v (%s)", err, strings.TrimSpace(string(out)))
	}
}

// seedRepo builds a repo with one dated commit per tag (2026-01-01 + i days)
// and tags HEAD with tags[i] after each commit. Last tag is the newest.
func seedRepo(t *testing.T, tags []string) string {
	t.Helper()
	if err := git.EnsureAvailable(); err != nil {
		t.Skipf("git not on PATH: %v", err)
	}
	dir := t.TempDir()
	mustGit(t, dir, "init")
	mustGit(t, dir, "config", "user.email", "gitagger-test@example.com")
	mustGit(t, dir, "config", "user.name", "gitagger-test")
	mustGit(t, dir, "config", "commit.gpgsign", "false")
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	for i, tag := range tags {
		commitDated(t, dir, base.Add(time.Duration(i)*24*time.Hour).Format(time.RFC3339))
		mustGit(t, dir, "tag", tag)
	}
	return dir
}

// seedShaRepo tags each dated commit with its own short SHA (num=false) or
// master-<count>-<short> (num=true). Returns dir + latest short + count.
func seedShaRepo(t *testing.T, num bool) (string, string, string) {
	t.Helper()
	if err := git.EnsureAvailable(); err != nil {
		t.Skipf("git not on PATH: %v", err)
	}
	dir := t.TempDir()
	mustGit(t, dir, "init")
	mustGit(t, dir, "config", "user.email", "gitagger-test@example.com")
	mustGit(t, dir, "config", "user.name", "gitagger-test")
	mustGit(t, dir, "config", "commit.gpgsign", "false")
	base := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	var short, count string
	for i := 0; i < 3; i++ {
		commitDated(t, dir, base.Add(time.Duration(i)*24*time.Hour).Format(time.RFC3339))
		var err error
		short, err = git.ShortSHA(dir)
		if err != nil {
			t.Fatal(err)
		}
		count, err = git.Count(dir)
		if err != nil {
			t.Fatal(err)
		}
		count = strings.TrimSpace(count)
		if num {
			mustGit(t, dir, "tag", "master-"+count+"-"+short)
		} else {
			mustGit(t, dir, "tag", "master-"+short)
		}
	}
	return dir, short, count
}

// autoPlan runs ComputePlan with no explicit format (pure auto mode).
func autoPlan(t *testing.T, dir, scale, pre string) run.Plan {
	t.Helper()
	plan, err := run.ComputePlan(run.Options{Dir: dir, Scale: scale, Pre: pre})
	if err != nil {
		t.Fatal(err)
	}
	return plan
}

func TestAutoTripleHistory(t *testing.T) {
	dir := seedRepo(t, []string{"v1.2.1", "v1.2.2", "v1.2.3"})
	p := autoPlan(t, dir, "patch", "stable")
	if p.Format != detect.Triple || p.Next != "v1.2.4" || !p.VPrefix || !p.Detected {
		t.Fatalf("got %+v, want triple v1.2.4", p)
	}
}

func TestAutoDoubleHistory(t *testing.T) {
	dir := seedRepo(t, []string{"v1.1", "v1.2", "v1.3"})
	p := autoPlan(t, dir, "patch", "stable")
	if p.Format != detect.Double || p.Next != "v1.4" {
		t.Fatalf("got %+v, want double v1.4", p)
	}
}

func TestAutoSingleHistory(t *testing.T) {
	dir := seedRepo(t, []string{"v5", "v6", "v7"})
	p := autoPlan(t, dir, "patch", "stable")
	if p.Format != detect.Single || p.Next != "v8" {
		t.Fatalf("got %+v, want single v8", p)
	}
}

func TestAutoDateHistory(t *testing.T) {
	dir := seedRepo(t, []string{"v2020.01.01", "v2020.01.02", "v2020.01.03"})
	p := autoPlan(t, dir, "patch", "stable")
	want := "v" + time.Now().Format("2006.01.02")
	if p.Format != detect.Date || p.Next != want {
		t.Fatalf("got %+v, want date %s", p, want)
	}
}

func TestAutoShaHistory(t *testing.T) {
	dir, short, _ := seedShaRepo(t, false)
	p := autoPlan(t, dir, "patch", "stable")
	if p.Format != detect.SHA || p.Next != "master-"+short {
		t.Fatalf("got %+v, want sha master-%s", p, short)
	}
}

func TestAutoShaNumHistory(t *testing.T) {
	dir, short, count := seedShaRepo(t, true)
	p := autoPlan(t, dir, "patch", "stable")
	if p.Format != detect.SHANum || p.Next != "master-"+count+"-"+short {
		t.Fatalf("got %+v, want sha-num master-%s-%s", p, count, short)
	}
}

func TestAutoPrefixlessStaysPrefixless(t *testing.T) {
	dir := seedRepo(t, []string{"1.2.1", "1.2.2", "1.2.3"})
	p := autoPlan(t, dir, "patch", "stable")
	if p.Next != "1.2.4" || p.VPrefix {
		t.Fatalf("got %+v, want prefixless 1.2.4", p)
	}
}

func TestAutoMixedMajorityTriple(t *testing.T) {
	dir := seedRepo(t, []string{"v2020.01.01", "v2020.01.02", "v1.2.1", "v1.2.2", "v1.2.3"})
	p := autoPlan(t, dir, "patch", "stable")
	if p.Format != detect.Triple || p.Next != "v1.2.4" {
		t.Fatalf("got %+v, want triple v1.2.4", p)
	}
}

func TestAutoMixedMajorityDate(t *testing.T) {
	dir := seedRepo(t, []string{"v1.2.1", "v1.2.2", "v2020.01.01", "v2020.01.02", "v2020.01.03"})
	p := autoPlan(t, dir, "patch", "stable")
	want := "v" + time.Now().Format("2006.01.02")
	if p.Format != detect.Date || p.Next != want {
		t.Fatalf("got %+v, want date %s", p, want)
	}
}

func TestAutoTieNewestWinsDate(t *testing.T) {
	dir := seedRepo(t, []string{"v1.2.3", "v2020.01.01"})
	p := autoPlan(t, dir, "patch", "stable")
	want := "v" + time.Now().Format("2006.01.02")
	if p.Format != detect.Date || p.Next != want {
		t.Fatalf("got %+v, want date %s (newest wins tie)", p, want)
	}
}

func TestAutoTieNewestWinsTriple(t *testing.T) {
	dir := seedRepo(t, []string{"v2020.01.01", "v1.2.3"})
	p := autoPlan(t, dir, "patch", "stable")
	if p.Format != detect.Triple || p.Next != "v1.2.4" {
		t.Fatalf("got %+v, want triple v1.2.4 (newest wins tie)", p)
	}
}

func TestAutoAllTypesTieUnit(t *testing.T) {
	all := []string{"v1.2.3", "v1.2", "v7", "v2020.01.01", "master-abc1234", "master-99-def5678"}
	f, _ := detect.Detect(all)
	if f != detect.SHANum {
		t.Fatalf("6-way tie = %q, want sha-num (newest)", f)
	}
	rev := []string{"master-99-def5678", "master-abc1234", "v2020.01.01", "v7", "v1.2", "v1.2.3"}
	f, _ = detect.Detect(rev)
	if f != detect.Triple {
		t.Fatalf("reversed 6-way tie = %q, want triple (newest)", f)
	}
}

func TestAutoVPrefixMixedUnit(t *testing.T) {
	_, v := detect.Detect([]string{"1.2.3", "v1.2.4", "v2020.01.01"})
	if !v {
		t.Errorf("2/3 v-prefix should keep v")
	}
	_, v = detect.Detect([]string{"1.2.3", "1.2.4", "v2020.01.01"})
	if v {
		t.Errorf("1/3 v-prefix should drop v")
	}
}

func TestAutoPromoteRcToStable(t *testing.T) {
	dir := seedRepo(t, []string{"v1.2.2", "v1.2.3", "v1.2.4-rc"})
	p := autoPlan(t, dir, "patch", "stable")
	if p.Next != "v1.2.4" {
		t.Fatalf("got %q, want promotion v1.2.4", p.Next)
	}
}

func TestAutoPreSwapRcToBeta(t *testing.T) {
	dir := seedRepo(t, []string{"v1.2.2", "v1.2.3", "v1.2.4-rc"})
	p := autoPlan(t, dir, "patch", "beta")
	if p.Next != "v1.2.4-beta" {
		t.Fatalf("got %q, want v1.2.4-beta", p.Next)
	}
}

func TestAutoDoublePreCases(t *testing.T) {
	dir := seedRepo(t, []string{"v1.1", "v1.2-rc"})
	if p := autoPlan(t, dir, "patch", "stable"); p.Next != "v1.2" {
		t.Fatalf("got %q, want double promotion v1.2", p.Next)
	}
	dir = seedRepo(t, []string{"v1.1", "v1.2-rc"})
	if p := autoPlan(t, dir, "patch", "beta"); p.Next != "v1.2-beta" {
		t.Fatalf("got %q, want v1.2-beta", p.Next)
	}
}

func TestAutoSinglePrePromote(t *testing.T) {
	dir := seedRepo(t, []string{"v6", "v7-rc"})
	p := autoPlan(t, dir, "patch", "stable")
	if p.Next != "v7" {
		t.Fatalf("got %q, want single promotion v7", p.Next)
	}
}

func TestAutoDateWithPre(t *testing.T) {
	dir := seedRepo(t, []string{"v2020.01.01", "v2020.01.02"})
	p := autoPlan(t, dir, "patch", "rc")
	want := "v" + time.Now().Format("2006.01.02") + "-rc"
	if p.Format != detect.Date || p.Next != want {
		t.Fatalf("got %+v, want %s", p, want)
	}
}

func TestAutoShaWithPre(t *testing.T) {
	dir, short, _ := seedShaRepo(t, false)
	p := autoPlan(t, dir, "patch", "rc")
	if p.Format != detect.SHA || p.Next != "master-"+short+"-rc" {
		t.Fatalf("got %+v, want master-%s-rc", p, short)
	}
}

func TestAutoFullLoop(t *testing.T) {
	dir := seedRepo(t, []string{"v1.2.2", "v1.2.3"})
	p := autoPlan(t, dir, "patch", "stable")
	if p.Next != "v1.2.4" {
		t.Fatalf("plan = %q, want v1.2.4", p.Next)
	}
	if err := git.CreateTag(dir, p.Next, "", false); err != nil {
		t.Fatal(err)
	}
	tags, err := git.ListTags(dir)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, tg := range tags {
		if tg == "v1.2.4" {
			found = true
		}
	}
	if !found {
		t.Fatalf("created tag missing from %v", tags)
	}
	// HEAD is tagged now: planning the same next tag must abort.
	o := run.Options{Dir: dir, Scale: "patch", Pre: "stable"}
	if err := run.CheckHeadTagged(o, "v1.2.4"); err == nil {
		t.Fatalf("expected abort when HEAD already tagged")
	}
	// New commit continues the sequence on auto.
	commitDated(t, dir, time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC).Format(time.RFC3339))
	p = autoPlan(t, dir, "patch", "stable")
	if p.Format != detect.Triple || p.Next != "v1.2.5" {
		t.Fatalf("got %+v, want triple v1.2.5 after new commit", p)
	}
}

func TestMaxPrevUnit(t *testing.T) {
	cases := []struct {
		name   string
		tags   []string
		format detect.Format
		want   string
	}{
		{"backtagged triple", []string{"v1.0.5", "v1.0.4"}, detect.Triple, "v1.0.5"},
		{"normal triple", []string{"v1.0.3", "v1.0.4"}, detect.Triple, "v1.0.4"},
		{"tie keeps newest", []string{"v1.2.4", "v1.2.4-rc"}, detect.Triple, "v1.2.4-rc"},
		{"tie keeps newest reversed", []string{"v1.2.4-rc", "v1.2.4"}, detect.Triple, "v1.2.4"},
		{"ignores other formats", []string{"v9.9.9", "v1.2.3"}, detect.Double, ""},
		{"double max", []string{"v1.9", "v1.10"}, detect.Double, "v1.10"},
		{"single max", []string{"v9", "v10"}, detect.Single, "v10"},
		{"date max", []string{"v2026.09.27", "v2026.09.26"}, detect.Date, "v2026.09.27"},
		{"sha has no order", []string{"v1.2.3"}, detect.SHA, ""},
		{"empty", nil, detect.Triple, ""},
	}
	for _, c := range cases {
		if got := detect.MaxPrev(c.tags, c.format); got != c.want {
			t.Errorf("%s = %q, want %q", c.name, got, c.want)
		}
	}
}

func TestAutoBacktaggedHistory(t *testing.T) {
	// Mirror a real back-tag: v1.0.5 lives on the OLDER commit while
	// v1.0.4 is newest by date. Bumping must follow max version (v1.0.6),
	// not newest date (which would collide with existing v1.0.5).
	dir := initRepo(t)
	mustGit(t, dir, "tag", "v1.0.5")
	commitDated(t, dir, time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339))
	mustGit(t, dir, "tag", "v1.0.4")
	p := autoPlan(t, dir, "patch", "stable")
	if p.Format != detect.Triple || p.Next != "v1.0.6" {
		t.Fatalf("got %+v, want triple v1.0.6", p)
	}
}
