# Generated workflows

Don't hand-write CI — generate it:

```sh
gitagger workflow             # overview
gitagger workflow --list      # available presets
gitagger workflow init gh     # .github/workflows/gitagger.yml
gitagger workflow init jenkins  # .jenkins/gitagger.jenkinsfile
```

Add `-f` to overwrite, `-p <dir>` to generate elsewhere. Generation also ensures a `.gitagger.yml` exists (writing the default one if missing) without touching an existing config.

## GitHub (`gh`)

Tags on `workflow_dispatch` using the gitagger action. Needs `contents: write`, a full checkout (`fetch-depth: 0`), and a git identity — all included in the template. Extend it with dispatch inputs as shown in [GitHub Action](github-action.md).

## Jenkins (`jenkins`)

Tags via `go install` plus `gitagger`, with the git identity configured. Runs wherever your agent runs.

!!! note "Gitea / Forgejo"
    Gitea Actions (and Forgejo) speak GitHub-compatible workflow YAML — the `gh` template works there as-is.
