# GitHub Actions

Tag from CI with the marketplace action:

```yaml
- uses: CHE3MZ/gitagger@v1
```

## Inputs

| Input | Default | Meaning |
|---|---|---|
| `args` | `""` | Extra args passed to gitagger, e.g. `"minor --pre rc"` |
| `version` | `latest` | gitagger release to use (`v1.0.2`, or `latest`) |
| `working-directory` | `.` | Directory to run gitagger in |
| `token` | `github.token` | Token for release API calls (higher rate limits) |

## Outputs

| Output | Meaning |
|---|---|
| `tag` | Tag now pointing at HEAD (empty when none) |

## Example: manual bump choice

```yaml
on:
  workflow_dispatch:
    inputs:
      bump:
        description: 'Tag bump type'
        required: false
        default: 'patch'
        type: choice
        options:
          - patch
          - minor
          - major
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
        with:
          args: ${{ inputs.bump }}
```

!!! note "History matters"
    Auto-detect needs history, so check out with `fetch-depth: 0`. The action fetches tags itself as best effort, but a full clone is the reliable setup. The action runs on Linux, macOS, and Windows runners.
