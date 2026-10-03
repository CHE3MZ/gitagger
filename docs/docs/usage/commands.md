# Commands

Every command supports `-h` / `--help`. Unknown flags are always an error naming the right help text. Typo a command and gitagger suggests the closest match:

```text
error: don't know what "hlp" means — try `gitagger --help`

The most similar command is
        help
```

Exit codes: `0` ok, `1` runtime error, `2` already-exists / no-op, `3` bad arguments.

## Tagging

| Command | What it does |
|---|---|
| `gitagger` | Next tag from config (default: patch bump), pushed when possible |
| `gitagger patch` | Bump `0.0.X` |
| `gitagger minor` | Bump `0.X.0` |
| `gitagger major` | Bump `X.0.0` |

Common flags: `--pre rc|beta|build|nightly`, `--format auto|triple|double|single|date|sha|sha-num`, `-m "message"` (annotated tag), `-n` keep local, `-d` dry-run, `-f` force, `-r` abort on dirty tree, `-v` verbose, `-p <dir>` operate elsewhere, `-a <name>` hook arguments. See [Tagging](tagging.md).

## Inspecting

| Command | What it does |
|---|---|
| `list` (`ls`) | All tags, oldest first |
| `view <tag>` | Name, type (annotated/lightweight), message, commit, date |
| `check [-v]` | Is `.gitagger.yml` valid? (`-v` shows resolved values) |
| `doctor [-v]` | Local vs remote tag health: odd styles, local-only/remote-only tags, HEAD distance |

## Removing

| Command | What it does |
|---|---|
| `remove <tag>` (`rm`) | Asks `are you sure…? [y/n]`, then deletes locally and on the remote when present |
| `remove <tag> -c` | Remove without asking |
| `remove <tag> -n` | Remove only the local tag, keep the remote one |

## Testing

| Command | What it does |
|---|---|
| `test [--keep]` | Clone to a temp sandbox (remotes stripped), run the full tag flow with hooks, print results, delete the sandbox |

## Setup

| Command | What it does |
|---|---|
| `init [-f] [--clean] [-p]` | Write `.gitagger.yml` (refuses to overwrite without `-f`) |
| `remote <name>` | Save another push remote into the config |
| `remote --show` | Show the current remote URL |
| `workflow` | Generate CI workflow files (see [Workflows](../automation/workflows.md)) |
| `handbook` | Print the `.gitagger.yml` handbook |
| `help [command]` | This reference, per command |
| `version` | Build version |
