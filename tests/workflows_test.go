// Workflow generation tests: list, init, refuse/force, unknown provider.
package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/CHE3MZ/gitagger/internal/cmd"
	"gopkg.in/yaml.v3"
)

func TestWorkflowsListed(t *testing.T) {
	ws := cmd.Workflows()
	if len(ws) != 2 || ws[0].ID != "gh" || ws[1].ID != "jenkins" {
		t.Fatalf("workflows = %+v", ws)
	}
	if _, ok := cmd.FindWorkflow("GH"); !ok {
		t.Fatalf("provider lookup should ignore case")
	}
	if _, ok := cmd.FindWorkflow("bogus"); ok {
		t.Fatalf("unknown provider should fail lookup")
	}
	if err := cmd.RunWorkflowsList(); err != nil {
		t.Fatalf("list should pass: %v", err)
	}
	if err := cmd.RunWorkflowsOverview(); err != nil {
		t.Fatalf("overview should pass: %v", err)
	}
}

func TestWorkflowsInitGH(t *testing.T) {
	dir := t.TempDir()
	if err := cmd.RunWorkflowsInit(dir, "gh", false, false); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(dir, ".github", "workflows", "gitagger.yml"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"CHE3MZ/gitagger@v1", "fetch-depth: 0", "contents: write"} {
		if !strings.Contains(string(raw), want) {
			t.Errorf("gh workflow missing %q", want)
		}
	}
	// Structural check: real YAML with the expected job/step shape.
	var doc map[string]any
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("generated workflow is not valid YAML: %v", err)
	}
	jobs, _ := doc["jobs"].(map[string]any)
	tag, _ := jobs["tag"].(map[string]any)
	steps, _ := tag["steps"].([]any)
	if len(steps) < 2 {
		t.Fatalf("generated workflow has no tag steps")
	}
	uses, _ := steps[1].(map[string]any)["uses"].(string)
	if uses != "CHE3MZ/gitagger@v1" {
		t.Errorf("tag step uses %q, want CHE3MZ/gitagger@v1", uses)
	}
	if err := cmd.RunWorkflowsInit(dir, "gh", false, false); err == nil {
		t.Fatalf("second init should refuse")
	} else if cmd.CodeOf(err) != 2 {
		t.Fatalf("refusal should exit 2, got %d", cmd.CodeOf(err))
	}
	if err := cmd.RunWorkflowsInit(dir, "gh", true, false); err != nil {
		t.Fatalf("forced init should overwrite: %v", err)
	}
}

func TestWorkflowsInitJenkins(t *testing.T) {
	dir := t.TempDir()
	if err := cmd.RunWorkflowsInit(dir, "jenkins", false, false); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(dir, ".jenkins", "gitagger.jenkinsfile"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"pipeline", "gitagger"} {
		if !strings.Contains(string(raw), want) {
			t.Errorf("jenkins file missing %q", want)
		}
	}
	// No Jenkins validator exists in CI; at least prove balanced structure.
	for _, pair := range [][2]string{{"{", "}"}, {"(", ")"}, {"[", "]"}} {
		if strings.Count(string(raw), pair[0]) != strings.Count(string(raw), pair[1]) {
			t.Errorf("jenkins file has unbalanced %s%s", pair[0], pair[1])
		}
	}
}

func TestWorkflowsInitUnknown(t *testing.T) {
	if err := cmd.RunWorkflowsInit(t.TempDir(), "bogus", false, false); err == nil {
		t.Fatalf("unknown provider should fail")
	}
}
