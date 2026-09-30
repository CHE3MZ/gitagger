// View/remove tests: inspect, show, confirm-delete, remote cleanup.
// All use temp dirs only — never the repo's own .git.
package tests

import (
	"strings"
	"testing"

	"github.com/CHE3MZ/gitagger/internal/cmd"
	"github.com/CHE3MZ/gitagger/internal/git"
)

func TestParseViewArgs(t *testing.T) {
	f, tag, err := cmd.ParseViewArgs([]string{"v1.2.3", "-v"})
	if err != nil || tag != "v1.2.3" || !f.Verbose {
		t.Fatalf("ParseViewArgs = %+v,%q,%v", f, tag, err)
	}
	f, tag, err = cmd.ParseViewArgs([]string{"--path=/tmp", "v1.0.0"})
	if err != nil || tag != "v1.0.0" || f.Path != "/tmp" {
		t.Fatalf("ParseViewArgs path = %+v,%q,%v", f, tag, err)
	}
	if _, _, err := cmd.ParseViewArgs([]string{"v1", "v2"}); err == nil {
		t.Errorf("two tags should fail")
	}
	if _, _, err := cmd.ParseViewArgs([]string{"v1", "--limit"}); err == nil {
		t.Errorf("unknown flag should fail")
	}
	if _, _, err := cmd.ParseViewArgs([]string{"-c"}); err == nil {
		t.Errorf("-c is not a view flag and should fail")
	}
}

func TestParseRemoveArgs(t *testing.T) {
	f, tag, err := cmd.ParseRemoveArgs([]string{"v1.2.3", "-c"})
	if err != nil || tag != "v1.2.3" || !f.Confirm {
		t.Fatalf("ParseRemoveArgs -c = %+v,%q,%v", f, tag, err)
	}
	f, _, err = cmd.ParseRemoveArgs([]string{"--confirm", "v1.2.3"})
	if err != nil || !f.Confirm {
		t.Fatalf("ParseRemoveArgs --confirm = %+v,%v", f, err)
	}
	if _, _, err := cmd.ParseRemoveArgs([]string{"v1", "v2"}); err == nil {
		t.Errorf("two tags should fail")
	}
	if _, _, err := cmd.ParseRemoveArgs([]string{"v1", "--force"}); err == nil {
		t.Errorf("--force is not a remove flag and should fail")
	}
}

func TestAskConfirm(t *testing.T) {
	for _, yes := range []string{"y\n", "Y\n", "yes\n", "YES\n", "  y  \n"} {
		if !cmd.AskConfirm(strings.NewReader(yes), "prompt?") {
			t.Errorf("AskConfirm(%q) = false, want true", yes)
		}
	}
	for _, no := range []string{"n\n", "N\n", "no\n", "\n", "maybe\n", "ye\n"} {
		if cmd.AskConfirm(strings.NewReader(no), "prompt?") {
			t.Errorf("AskConfirm(%q) = true, want false", no)
		}
	}
}

func TestInspectTagLightweight(t *testing.T) {
	dir := initRepo(t)
	if err := git.CreateTag(dir, "v1.0.0", "", false); err != nil {
		t.Fatal(err)
	}
	info, err := git.InspectTag(dir, "v1.0.0")
	if err != nil {
		t.Fatal(err)
	}
	if info.Name != "v1.0.0" || info.Kind != "lightweight" {
		t.Fatalf("info = %+v, want lightweight v1.0.0", info)
	}
	if info.Message != "" {
		t.Fatalf("lightweight message = %q, want empty", info.Message)
	}
	if info.Commit == "" || info.Date == "" {
		t.Fatalf("info = %+v, want commit and date set", info)
	}
}

func TestInspectTagAnnotated(t *testing.T) {
	dir := initRepo(t)
	if err := git.CreateTag(dir, "v2.0.0", "hello release", false); err != nil {
		t.Fatal(err)
	}
	info, err := git.InspectTag(dir, "v2.0.0")
	if err != nil {
		t.Fatal(err)
	}
	if info.Kind != "annotated" {
		t.Fatalf("kind = %q, want annotated", info.Kind)
	}
	if !strings.Contains(info.Message, "hello release") {
		t.Fatalf("message = %q, want hello release", info.Message)
	}
}

func TestInspectTagMissing(t *testing.T) {
	dir := initRepo(t)
	if _, err := git.InspectTag(dir, "v9.9.9"); err == nil {
		t.Fatalf("missing tag should fail")
	}
	if _, err := git.InspectTag(dir, ""); err == nil {
		t.Fatalf("empty tag should fail")
	}
}

func TestRunView(t *testing.T) {
	dir := initRepo(t)
	if err := git.CreateTag(dir, "v1.0.0", "", false); err != nil {
		t.Fatal(err)
	}
	if err := cmd.RunView(dir, "v1.0.0", false); err != nil {
		t.Fatalf("view lightweight should pass: %v", err)
	}
	if err := git.CreateTag(dir, "v1.0.1", "nice release", false); err != nil {
		t.Fatal(err)
	}
	if err := cmd.RunView(dir, "v1.0.1", false); err != nil {
		t.Fatalf("view annotated should pass: %v", err)
	}
	mustGit(t, dir, "remote", "add", "origin", "https://example.com/repo.git")
	if err := cmd.RunView(dir, "v1.0.1", true); err != nil {
		t.Fatalf("verbose view should pass: %v", err)
	}
}

func TestRunViewMissing(t *testing.T) {
	dir := initRepo(t)
	err := cmd.RunView(dir, "v9.9.9", false)
	if err == nil {
		t.Fatalf("view of missing tag should fail")
	}
	if cmd.CodeOf(err) != 1 {
		t.Fatalf("missing tag should exit 1, got %d", cmd.CodeOf(err))
	}
	if err := cmd.RunView(dir, "", false); err == nil {
		t.Fatalf("view without a tag should fail")
	} else if cmd.CodeOf(err) != 3 {
		t.Fatalf("missing arg should exit 3, got %d", cmd.CodeOf(err))
	}
}

func TestRunRemoveConfirmed(t *testing.T) {
	dir := initRepo(t)
	if err := git.CreateTag(dir, "v1.0.0", "", false); err != nil {
		t.Fatal(err)
	}
	if err := cmd.RunRemove(dir, "v1.0.0", true, false); err != nil {
		t.Fatalf("confirmed remove should pass: %v", err)
	}
	tags, err := git.ListTags(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, tg := range tags {
		if tg == "v1.0.0" {
			t.Fatalf("tag still present after remove: %v", tags)
		}
	}
}

func TestRunRemoveMissing(t *testing.T) {
	dir := initRepo(t)
	err := cmd.RunRemove(dir, "v9.9.9", true, false)
	if err == nil {
		t.Fatalf("remove of missing tag should fail")
	}
	if cmd.CodeOf(err) != 1 {
		t.Fatalf("missing tag should exit 1, got %d", cmd.CodeOf(err))
	}
	if err := cmd.RunRemove(dir, "", true, false); err == nil {
		t.Fatalf("remove without a tag should fail")
	} else if cmd.CodeOf(err) != 3 {
		t.Fatalf("missing arg should exit 3, got %d", cmd.CodeOf(err))
	}
}

func TestRunRemovePrompt(t *testing.T) {
	old := cmd.Stdin
	defer func() { cmd.Stdin = old }()

	// "n" aborts and keeps the tag.
	dir := initRepo(t)
	if err := git.CreateTag(dir, "v1.0.0", "", false); err != nil {
		t.Fatal(err)
	}
	cmd.Stdin = strings.NewReader("n\n")
	if err := cmd.RunRemove(dir, "v1.0.0", false, false); err != nil {
		t.Fatalf("declined remove should pass as a no-op: %v", err)
	}
	tags, err := git.ListTags(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(tags) != 1 || tags[0] != "v1.0.0" {
		t.Fatalf("declined remove dropped the tag: %v", tags)
	}

	// Empty answer also aborts.
	cmd.Stdin = strings.NewReader("\n")
	if err := cmd.RunRemove(dir, "v1.0.0", false, false); err != nil {
		t.Fatalf("empty answer should abort cleanly: %v", err)
	}
	if tags, _ := git.ListTags(dir); len(tags) != 1 {
		t.Fatalf("empty answer dropped the tag: %v", tags)
	}

	// "y" deletes.
	cmd.Stdin = strings.NewReader("y\n")
	if err := cmd.RunRemove(dir, "v1.0.0", false, false); err != nil {
		t.Fatalf("confirmed remove should pass: %v", err)
	}
	if tags, _ := git.ListTags(dir); len(tags) != 0 {
		t.Fatalf("tag still present after y: %v", tags)
	}
}

func TestRunRemoveRemote(t *testing.T) {
	dir := initRepo(t)
	remoteDir := t.TempDir()
	mustGit(t, remoteDir, "init", "--bare", "-q")
	mustGit(t, dir, "remote", "add", "origin", remoteDir)
	if err := git.CreateTag(dir, "v1.0.0", "", false); err != nil {
		t.Fatal(err)
	}
	mustGit(t, dir, "push", "origin", "v1.0.0")
	if tags, err := git.RemoteTags(dir, "origin"); err != nil || len(tags) != 1 {
		t.Fatalf("remote tags = %v,%v, want [v1.0.0]", tags, err)
	}
	if err := cmd.RunRemove(dir, "v1.0.0", true, false); err != nil {
		t.Fatalf("remove should pass: %v", err)
	}
	if tags, _ := git.ListTags(dir); len(tags) != 0 {
		t.Fatalf("local tag still present: %v", tags)
	}
	tags, err := git.RemoteTags(dir, "origin")
	if err != nil {
		t.Fatal(err)
	}
	for _, tg := range tags {
		if tg == "v1.0.0" {
			t.Fatalf("remote tag still present: %v", tags)
		}
	}
}

func TestViewRemoveHelp(t *testing.T) {
	for _, c := range []string{"view", "remove"} {
		if !cmd.IsCommand(c) {
			t.Errorf("IsCommand(%q) = false", c)
		}
		cmd.CommandHelp(c) // must not crash
		if n := cmd.ExampleCount(c); n == 0 || n > 3 {
			t.Errorf("%s has %d examples, want 1-3", c, n)
		}
	}
}
