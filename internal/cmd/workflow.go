// Workflow generation: CI files that run gitagger for you.
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/CHE3MZ/gitagger/internal/config"
	"github.com/CHE3MZ/gitagger/internal/style"
)

// Workflow is a generatable CI file.
type Workflow struct {
	ID     string // gh, jenkins...
	Target string // path written, relative to the project dir
	Desc   string // one line for listings
	Blurb  string // very short description for the overview
}

// Workflows lists every generatable file, in display order.
func Workflows() []Workflow {
	return []Workflow{
		{ID: "gh", Target: filepath.Join(".github", "workflows", "gitagger.yml"), Desc: "GitHub Actions: tags on push using the gitagger action", Blurb: "GitHub Actions workflow file"},
		{ID: "jenkins", Target: filepath.Join(".jenkins", "gitagger.jenkinsfile"), Desc: "Jenkins pipeline: tags using go install", Blurb: "Jenkins workflow file"},
	}
}

// FindWorkflow returns the provider by id (case-insensitive).
func FindWorkflow(id string) (Workflow, bool) {
	for _, w := range Workflows() {
		if strings.ToLower(id) == w.ID {
			return w, true
		}
	}
	return Workflow{}, false
}

// RunWorkflowOverview prints the compact COMMANDS + WORKFLOWS view.
func RunWorkflowOverview() error {
	fmt.Println(style.Header("COMMANDS"))
	fmt.Println()
	fmt.Printf("  %s %s\n", style.White("gitagger"), style.Green("workflow init"))
	fmt.Println()
	fmt.Println(style.Header("WORKFLOWS"))
	fmt.Println()
	for _, w := range Workflows() {
		fmt.Printf("  %s%s\n", style.Green(w.ID), style.Gray(strings.Repeat(" ", 10-len(w.ID))+w.Blurb))
	}
	return nil
}

// RunWorkflowList prints every generatable workflow.
func RunWorkflowList() error {
	for _, w := range Workflows() {
		fmt.Printf("  %s\n", style.BoldGreen(w.ID))
		fmt.Printf("    %s\n", style.White(w.Target))
		fmt.Printf("    %s\n", style.Dim(w.Desc))
	}
	fmt.Println(style.Dim("generate one with `gitagger workflow init <id>` (`-f` overwrites)"))
	return nil
}

// RunWorkflowInit writes the provider file under dir, making parents.
// Refuses to overwrite unless force.
func RunWorkflowInit(dir, id string, force, verbose bool) error {
	w, ok := FindWorkflow(id)
	if !ok {
		return BadArgs("don't know workflow %q — try `gitagger workflow --help`", id)
	}
	path := filepath.Join(dir, w.Target)
	if st, err := os.Stat(path); err == nil && !st.IsDir() && !force {
		return Exists("workflow already exists at %s (use -f to overwrite)", path)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return Generic("couldn't create %s (%v)", filepath.Dir(path), err)
	}
	content := ghWorkflow
	if w.ID == "jenkins" {
		content = jenkinsWorkflow
	}
	// #nosec G306 — CI files are committed, group-readable is standard.
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return Generic("couldn't write %s (%v)", path, err)
	}
	if verbose {
		fmt.Printf("%s %s\n", style.Green("wrote"), style.White(path))
	} else {
		fmt.Printf("%s %s\n", style.Green("wrote"), style.White(w.Target))
	}
	if config.Find(dir) == "" {
		cfgPath := filepath.Join(dir, ".gitagger.yml")
		// #nosec G306 — 0600, same as `gitagger init`.
		if err := os.WriteFile(cfgPath, []byte(config.DefaultFileContent()), 0o600); err != nil {
			return Generic("couldn't write .gitagger.yml (%v)", err)
		}
		fmt.Printf("%s %s\n", style.Green("wrote"), style.White(".gitagger.yml"))
	}
	return nil
}

const ghWorkflow = `name: Gitagger

on:
  workflow_dispatch:

jobs:
  tag:
    runs-on: ubuntu-latest
    permissions:
      contents: write
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
      - uses: CHE3MZ/gitagger@v1
        # with:
        #   args: minor
        #   version: vN.N.N
`

const jenkinsWorkflow = `pipeline {
  agent any
  stages {
    stage('Tag') {
      steps {
        sh 'go install github.com/CHE3MZ/gitagger/cmd/gitagger@latest'
        sh 'gitagger'
      }
    }
  }
}
`
