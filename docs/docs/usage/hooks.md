# Hooks

Shell commands at lifecycle events: `start` runs first, `success` after tagging (and before pushing — a failing success hook rolls the tag back and nothing is pushed), `failure` on errors, `finish` always (both outcomes). Each event takes one block or a list of blocks.

```yaml
on:
  success:
    - shell: sh
      os: linux, macos
      run:
        - echo "tagged $GITAGGER_TAG"
    - shell: pwsh
      os: windows
      run: echo "tagged $GITAGGER_TAG"
```

## Block keys

| Key | Meaning |
|---|---|
| `run:` | One command or a list. Each item runs separately. Required. |
| `shell:` | `sh` (default), `bash`, `pwsh`, `batch`. No `system` shell — explicit only. |
| `os:` | `linux`, `macos`, `windows`. Default: all. Comma lists work (`os: macos, linux`). |
| `argument:` | Block runs only with `gitagger -a <name>`. Special values: `none` (only when no `-a` given), `any` (only when some `-a` given). |

## Environment

Every hook gets:

| Variable | Contents |
|---|---|
| `GITAGGER_TAG` | The new tag |
| `GITAGGER_PREV` | The previous tag (empty on first tag) |
| `GITAGGER_REMOTE` | The push remote |
| `GITAGGER_PUSHED` | `true`/`false` |
| `GITAGGER_DRY_RUN` | `true` inside `gitagger test`, `false` otherwise |
| `GITAGGER_EVENT` | The lifecycle event |
| `GITAGGER_ARGUMENT` | Comma-joined `-a` values |

## Failure semantics

A failing `success` hook still runs the `failure` hooks first (try/catch/finally), then reports the original error. A failing `failure`/`finish` hook is reported, never re-triggered — hooks can't loop forever.

!!! tip "Changelogs and releases"
    Hooks compose with outside tools: run your release script or changelog generator as a `success` hook (it gets `$GITAGGER_TAG` and friends) instead of asking gitagger to own changelogs.
