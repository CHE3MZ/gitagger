// Message command tests: config forms, output joining, env vars,
// fail-before-create, flag override, dry run, and os skipping.
package tests

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/CHE3MZ/gitagger/internal/cmd"
	"github.com/CHE3MZ/gitagger/internal/config"
	"github.com/CHE3MZ/gitagger/internal/git"
	"github.com/CHE3MZ/gitagger/internal/run"
)

func messageOptions(t *testing.T, dir string) run.Options {
	t.Helper()
	cfg, _, err := config.Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	o, err := cmd.BuildOptions(dir, cfg, "", cmd.TagFlags{})
	if err != nil {
		t.Fatal(err)
	}
	o.Push = false
	return o
}

func firstTag(t *testing.T, dir string) string {
	t.Helper()
	tags, err := git.ListTags(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(tags) != 1 {
		t.Fatalf("tags = %v, want exactly one", tags)
	}
	return tags[0]
}

func tagMessage(t *testing.T, dir, tag string) string {
	t.Helper()
	info, err := git.InspectTag(dir, tag)
	if err != nil {
		t.Fatal(err)
	}
	return info.Message
}

func TestMessageConfigForms(t *testing.T) {
	dir := t.TempDir()
	writeConfig(t, dir, "message:\n"+
		"  cmds:\n"+
		"    - echo bare\n"+
		"    - cmd: echo block\n"+
		"      shell: bash\n"+
		"      os: linux, macos\n")
	cfg, _, err := config.Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Message != "" {
		t.Fatalf("cfg.Message = %q, want empty", cfg.Message)
	}
	if len(cfg.MessageCmds) != 2 {
		t.Fatalf("MessageCmds = %+v, want 2", cfg.MessageCmds)
	}
	if cfg.MessageCmds[0].Cmd != "echo bare" || cfg.MessageCmds[0].Shell != "" {
		t.Fatalf("bare cmd = %+v, want empty shell (platform default)", cfg.MessageCmds[0])
	}
	if cfg.MessageCmds[1].Shell != "bash" || cfg.MessageCmds[1].OS != "linux, macos" {
		t.Fatalf("block cmd = %+v", cfg.MessageCmds[1])
	}
}

func TestMessageScalarStaysLiteral(t *testing.T) {
	dir := t.TempDir()
	writeConfig(t, dir, "message: \"$(echo hi)\"\n")
	cfg, _, err := config.Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Message != "$(echo hi)" || len(cfg.MessageCmds) != 0 {
		t.Fatalf("cfg.Message = %q cmds = %+v, want literal text only", cfg.Message, cfg.MessageCmds)
	}
}

func TestMessageConfigRejects(t *testing.T) {
	cases := []struct{ name, content string }{
		{"unknown message key", "message:\n  cmd: echo hi\n"},
		{"unknown block key", "message:\n  cmds:\n    - command: echo hi\n"},
		{"empty cmds", "message:\n  cmds: []\n"},
		{"empty cmd", "message:\n  cmds:\n    - cmd: \"\"\n"},
		{"bad shell", "message:\n  cmds:\n    - cmd: echo hi\n      shell: fish\n"},
		{"bad os", "message:\n  cmds:\n    - cmd: echo hi\n      os: plan9\n"},
		{"nested list", "message:\n  cmds:\n    - - echo hi\n"},
		{"bare sequence", "message:\n  - echo hi\n"},
	}
	for _, c := range cases {
		dir := t.TempDir()
		writeConfig(t, dir, c.content)
		if _, _, err := config.Load(dir); err == nil {
			t.Errorf("%s should fail validation", c.name)
		}
	}
}

func TestMessageCmdsJoin(t *testing.T) {
	needSh(t)
	dir := initRepo(t)
	writeConfig(t, dir, "message:\n  cmds:\n    - echo hello\n    - echo 123\n")
	if err := cmd.RunTag(messageOptions(t, dir)); err != nil {
		t.Fatal(err)
	}
	if msg := tagMessage(t, dir, firstTag(t, dir)); msg != "hello\n123" {
		t.Fatalf("message = %q, want %q", msg, "hello\n123")
	}
}

func TestMessageCmdsEnv(t *testing.T) {
	needSh(t)
	dir := initRepo(t)
	writeConfig(t, dir, "message:\n  cmds:\n    - cmd: echo \"$GITAGGER_EVENT:$GITAGGER_TAG\"\n      shell: sh\n")
	if err := cmd.RunTag(messageOptions(t, dir)); err != nil {
		t.Fatal(err)
	}
	tag := firstTag(t, dir)
	if msg := tagMessage(t, dir, tag); msg != "message:"+tag {
		t.Fatalf("message = %q, want %q", msg, "message:"+tag)
	}
}

func TestMessageCmdFailureAborts(t *testing.T) {
	needSh(t)
	dir := initRepo(t)
	writeConfig(t, dir, "message:\n  cmds:\n    - echo hi\n    - exit 3\n")
	if err := cmd.RunTag(messageOptions(t, dir)); err == nil {
		t.Fatalf("failing message command should abort the run")
	}
	if tags, _ := git.ListTags(dir); len(tags) != 0 {
		t.Fatalf("tags = %v, want none (abort before create)", tags)
	}
}

func TestMessageFlagBeatsCmds(t *testing.T) {
	needSh(t)
	dir := initRepo(t)
	writeConfig(t, dir, "message:\n  cmds:\n    - echo from-config > applied.txt\n")
	cfg, _, err := config.Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	o, err := cmd.BuildOptions(dir, cfg, "", cmd.TagFlags{Message: "from-flag"})
	if err != nil {
		t.Fatal(err)
	}
	o.Push = false
	if err := cmd.RunTag(o); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "applied.txt")); err == nil {
		t.Fatalf("message commands ran despite -m override")
	}
	if msg := tagMessage(t, dir, firstTag(t, dir)); msg != "from-flag" {
		t.Fatalf("message = %q, want from-flag", msg)
	}
}

func TestMessageDryRunDoesNotRun(t *testing.T) {
	needSh(t)
	dir := initRepo(t)
	writeConfig(t, dir, "message:\n  cmds:\n    - echo oops > applied.txt\n")
	o := messageOptions(t, dir)
	o.DryRun = true
	if err := cmd.RunTag(o); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "applied.txt")); err == nil {
		t.Fatalf("message commands ran during dry run")
	}
	if tags, _ := git.ListTags(dir); len(tags) != 0 {
		t.Fatalf("dry run created tags: %v", tags)
	}
}

func TestMessageOSSkip(t *testing.T) {
	needSh(t)
	dir := initRepo(t)
	other := "windows"
	if runtime.GOOS == "windows" {
		other = "linux"
	}
	writeConfig(t, dir, "message:\n  cmds:\n"+
		"    - cmd: echo skipped\n"+
		"      os: "+other+"\n"+
		"    - echo kept\n")
	if err := cmd.RunTag(messageOptions(t, dir)); err != nil {
		t.Fatal(err)
	}
	if msg := tagMessage(t, dir, firstTag(t, dir)); msg != "kept" {
		t.Fatalf("message = %q, want kept", msg)
	}
}

func TestMessageEmptyOutputLightweight(t *testing.T) {
	needSh(t)
	dir := initRepo(t)
	writeConfig(t, dir, "message:\n  cmds:\n    - true\n")
	if err := cmd.RunTag(messageOptions(t, dir)); err != nil {
		t.Fatal(err)
	}
	info, err := git.InspectTag(dir, firstTag(t, dir))
	if err != nil {
		t.Fatal(err)
	}
	if info.Kind != "lightweight" {
		t.Fatalf("kind = %q, want lightweight for empty output", info.Kind)
	}
}
