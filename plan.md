# gitagger — Plan v2

> Goal: Be the `goreleaser` of git tags. Just run `gitagger` and it does the right thing.
> Convention over configuration. Config and flags only when you need more control.

## 1. Philosophy

1. **Zero-config by default:** `gitagger` with no args inspects history and does the obvious next tag.
2. **Explicit when needed:** flags override config, config overrides auto-detect.
3. **Seamless but safe:** create locally always, auto-push when safe, never lose work, always `--dry-run`able.
4. **CI-friendly:** no prompts by default, `--no-push` / `--confirm` / `--json` for scripts.

Precedence: `CLI flags > .gitagger.yml > built-in defaults`.

Decision: **yes, auto-push by default** — if a `remote` is configured AND reachable, `gitagger` pushes the single new tag with no prompt. If no remote or offline, it keeps the local tag and warns instead of failing. No prompt unless user opts in with `--confirm`.

## 2. Default behavior (the seamless case)

```sh
$ gitagger
detected: triple + v-prefix + stable (from last 20 tags)
next: v1.2.4 (was v1.2.3)
created tag v1.2.4
pushed v1.2.4 to origin.
```

No prompt. Easiest possible flow. Other outcomes:

```sh
$ gitagger
created tag v1.2.4
warning: no remote 'origin' — tag kept locally only.

$ gitagger
created tag v1.2.4
warning: remote 'origin' unreachable (offline?) — tag kept locally.
run `git push origin v1.2.4` later.
```

Rules for bare `gitagger`:

* `scale = patch`, `pre = stable`, `format = auto`, `push = true`, `confirm = false`
* Always creates the local tag first, then tries to push (see §8). Push is skipped gracefully — never deletes the local tag on push failure.
* If repo has no tags: start at `v1.0.0` (triple + `v`, stable).
* If HEAD already tagged: abort with `already on <tag>, use --force to retag` instead of duplicate.
* If working tree dirty: warn but allow (tags don't depend on tree). Only abort tag creation if HEAD has no commits. Do NOT block on dirty — unlike goreleaser build. Add `--require-clean` opt-in.
* Prompt appears ONLY with `--confirm` flag or `confirm: true` in config AND interactive TTY.

## 3. Supported tag archetypes

All archetypes preserve `v` prefix if majority of existing tags use it. E.g. `1.2.3` stays prefix-less, `v1.2.3` stays `v`.

### 3.1 Semver-like

* `triple`: `v1.2.3` + optional `-<pre>` → `v1.2.3-rc`, `v1.2.3-beta`, `v1.2.3-build`, `v1.2.3-nightly`
* `double`: `v1.2` + optional `-<pre>`
* `single`: `v1` + optional `-<pre>`

Increment maps to `major|minor|patch` arg. `double`: patch bumps minor component (`v1.2` → `v1.3`), major bumps major. `single`: any bump bumps `N` (`v1` → `v2`).

### 3.2 Date

* `date`: `vYYYY.MM.DD` + optional `-<pre>` and optional collision counter `.N`
* ex: `v2026.09.26`, `v2026.09.26.1`, `v2026.09.26-rc`

`major|minor|patch` arg is ignored. Always uses today (local). If today's tag exists, append/increment `.N`. Pre suffix appended after date/counter.

### 3.3 SHA

* `sha`: `master-<short-sha>` + optional `-<pre>` → `master-2f88688`, `master-2f88688-rc`
* `sha-num`: `master-<count>-<short-sha>` → `master-1234-2f88688`

Branch name is literal `master` for compat (do not use current branch name in v1). `major|minor|patch` ignored. If tag for current SHA already exists: abort, suggest `--force`. `count` = `git rev-list --count HEAD`.

### 3.4 Custom (v2, not MVP)

Template string, see §7.

MVP implements: `triple`, `date`, `auto`. `double`, `single`, `sha`, `sha-num`, `custom` are parsed/detected but creation can return `not yet implemented` — or implement if trivial after MVP.

## 4. Auto-detect (`--format auto`, the default)

Input: `git tag --list`, sorted by `creatordate` descending, take last 20.

1. Normalize: strip trailing `-rc|-beta|-build|-nightly` and `.<counter>` for date, strip `v` for classification.
2. Classify each to `triple|double|single|date|sha|sha-num|unknown`.
3. Majority wins. Ties → most recent tag's format wins.
4. `unknown` tags ignored. If all unknown or no tags → `triple`.
5. `v-prefix`: true if >=50% of classified tags start with `v`.
6. `pre`: NOT auto-detected for next tag — next tag defaults to `stable` unless `--pre` given, except promotion case (§5.3).

Also detect `base branch word` for sha (`master` only in v1).

Log detected result in verbose / dry-run: `detected: triple, v-prefix=true`.

## 5. CLI spec

Use Cobra-style. No order-independent bare words.

```
gitagger [major|minor|patch] [flags]
gitagger <subcommand> [flags]
```

`major|minor|patch` is optional positional, default `patch`. Only one allowed.

### 5.1 Flags

| Flag | Default | Meaning |
|---|---|---|
| `--pre <stable\|rc\|beta\|build\|nightly>` | `stable` | prerelease channel. `stable` = no suffix |
| `--format <auto\|triple\|double\|single\|date\|sha\|sha-num\|custom>` | `auto` | force format, skip detect |
| `--custom <template>` | — | required if `--format custom` |
| `--no-push` | `false` | create locally only, skip auto-push |
| `--confirm` | `false` | opt-in to `push X? [Y/n]` prompt (TTY only). Default is no prompt |
| `-y, --yes` | accepted, no-op | compat: gitagger never prompts by default, so `--yes` just means "I confirm" |
| `--remote <name>` | `origin` | push remote |
| `--force` | `false` | allow retag / overwrite existing (`push --force` for that single tag only) |
| `--dry-run` | `false` | print next tag + push plan, create/push nothing |
| `-m, --message <msg>` | — | annotated tag message. If set, `git tag -a`. Else lightweight |
| `--require-clean` | `false` | abort if `git status --porcelain` non-empty |
| `--json` | `false` | machine output `{"prev":..,"next":..,"pushed":..,"remote":..}` |
| `--verbose` | `false` | show detection + git cmds |
| `-v, --version` | — | print version |
| `-h, --help` | — | help |

No order-free bare words like `gitagger patch release yes`. Scale is the only positional. Everything else is a flag or subcommand.

### 5.2 Subcommands

* `gitagger init [--force]` → writes `.gitagger.yml` with commented defaults. Refuse overwrite unless `--force`.
* `gitagger list` (alias `ls`, `view`) → list tags sorted by `creatordate`. Flags: `--limit N` (default 20), `--json`.
* `gitagger doctor [--fix]` → `git fetch --tags`, report: local-only tags, remote-only tags, malformed tags, HEAD-untagged distance. `--fix` prunes local-only malformed on confirm. Never delete without confirm.
* `gitagger version` → version.

Legacy compat: if first arg is `view` → alias to `list` with deprecation warning. If first arg is `yes` → treat as `--yes` with warning. Remove in v2.

### 5.3 Increment + promotion rules

Semver (`triple`):

* `patch`: `v1.2.3` → `v1.2.4`
* `minor`: → `v1.3.0`
* `major`: → `v2.0.0`
* With `--pre rc`: bump first, then suffix: `v1.2.3` + `patch --pre rc` → `v1.2.4-rc`
* Promotion: if latest is `v1.2.4-rc` and user runs `--pre stable` (or bare `gitagger` with explicit `--format triple`): strip suffix, no bump → `v1.2.4`. If latest is already stable and user runs stable: normal bump.
* Pre-to-pre: `v1.2.4-rc` + `--pre beta` → `v1.2.4-beta` (no bump).

Date/sha: no promotion concept in v1, suffix just appended.

## 6. Config file

Lookup in order: `./.gitagger.yml` → `./.gitagger.yaml` → `./.gitagger` (plain yaml). Stop at first found. No merging.

```yaml
# generated by `gitagger init`
version: 1
scale: patch        # major | minor | patch
pre: stable         # stable | rc | beta | build | nightly
format: auto        # auto | triple | double | single | date | sha | sha-num | custom
custom: ""          # template when format=custom
remote: origin
push: true          # true = auto-push when safe (default). false = same as --no-push
confirm: false      # true = ask `push X? [Y/n]` before pushing (TTY only). default no prompt
require_clean: false
message: ""         # annotated tag message
```

Prompt appears ONLY when `confirm=true` (or `--confirm`) AND TTY AND `push=true` AND not `--no-push/--dry-run`. In all other cases gitagger just does it.

## 7. Custom format (v2)

Template tokens (case-sensitive, uppercase):

* `YEAR`=`YYYY`, `MONTH`=`MM`, `DAY`=`DD`
* `MAJ`, `MIN`, `PCH` = bumped components per `scale`
* `SHA` = short SHA, `FULLSHA` = full SHA, `COUNT` = rev-list count
* `PRE` = `` or `-rc` etc.
* `N` = auto-increment counter to make unique (starts 0/empty, then 1,2…)
* `TAG` = previous tag (raw)

Ex: `format: custom`, `custom: "YEAR.MONTH.DAY.N"` → `2026.09.26`, `2026.09.26.1`
Ex: `custom: "myapp-YEAR.DAY-SHA"` etc. Literal text allowed `[a-zA-Z0-9._-]`.

Uniqueness: if rendered tag exists, increment `N` if template contains `N`, else error and suggest adding `N` or using `--force`.

## 8. Git behavior / push policy + remote safety

Golden rule: **local tag is never lost because push failed.** Create first, push second.

* Create: `git tag [<tag>]` or `git tag -a <tag> -m <msg>` if message set.
* Push (only if `push=true` and not `--no-push` / `--dry-run`):
  1. **Remote configured?** Run `git remote get-url <remote>`. If it fails → `warning: no remote '<remote>' — tag <tag> kept locally only.` Stop, exit 0, `pushed=false`.
  2. **Remote reachable?** Run `git ls-remote --tags <remote>` with short timeout (~10s). If it fails (offline / no internet / auth) → `warning: remote '<remote>' unreachable — tag <tag> kept locally. run 'git push <remote> <tag>' later.` Stop, exit 0, `pushed=false`. Never delete local tag.
  3. **Collision?** If remote already has `<tag>` pointing to a different SHA → do NOT push. Abort with `tag <tag> already exists on remote, use --force to overwrite or run doctor`. Keep local tag. Exit 2.
  4. **Confirm?** Only if `--confirm` / `confirm:true` + TTY → `push <tag> to <remote>? [Y/n]`. `n` = keep local, exit 0, `pushed=false`.
  5. **Push single ref only:** `git push <remote> <tag>`. Never `push --tags`. Only with `--force` → `git push --force <remote> <tag>`.
* On push command failure: keep local tag, print `push failed (...) — tag <tag> kept locally`, exit 1, `pushed=false`.
* Pre-flight: HEAD exists, tag doesn't exist locally (unless `--force`).
* `doctor` runs `git fetch --tags <remote>` read-only by default.
* Exit codes: `0` ok (including "created locally, push skipped" with warning), `1` push attempted but failed / generic, `2` already-exists / no-op, `3` bad config/args.

## 9. Examples

```sh
gitagger                    # auto patch stable + auto-push if remote reachable
gitagger minor              # v1.2.3 -> v1.3.0 + push
gitagger major --pre rc     # v1.2.3 -> v2.0.0-rc + push
gitagger --dry-run --verbose # preview only
gitagger --no-push          # local only
gitagger --confirm          # old prompt behavior, opt-in
gitagger --format date      # v2026.09.26 + push
gitagger --format sha       # master-2f88688 + push
gitagger patch --pre beta -m "beta release"
gitagger list --limit 10
gitagger doctor --fix
gitagger init
```

Config-only flow:

```sh
gitagger init   # edit .gitagger.yml once
gitagger        # forever after
```

## 10. MVP scope

v1 must have:

* [ ] bare `gitagger` + `major|minor|patch` + `--pre` + `--format auto|triple|date`
* [ ] auto-detect + `v` preservation + no-tags fallback
* [ ] auto-push w/ remote checks (§8): no-remote skip, offline skip, single-ref push only
* [ ] lightweight + annotated (`-m`), `--dry-run`, `--no-push`, `--confirm`, `--remote`, `--force`
* [ ] `init`, `list`, `doctor` (read-only), `version`
* [ ] `.gitagger.yml` load + precedence
* [ ] `--json` for CI
* [ ] repo layout + workflow per §12 (module path, `internal/` logic, ops scripts green)

v2: `double/single/sha/sha-num/custom`, `doctor --fix` delete, changelog notes, `--require-clean` default?, signing `-s`.

## 11. Decisions (locked for simplicity)

1. SHA prefix: keep literal `master-` in v1 for compat. Do not use current branch name.
2. Date collision suffix: `.N` → `v2026.09.26`, `v2026.09.26.1`. Not `-N`.
3. Double `v1.2` patch semantics: bump minor component (`v1.2` → `v1.3`).

## 12. Repo layout + dev workflow (mandatory)

This is not optional — the tool stays simple only if the code stays separated.

### 12.1 Module

* `go.mod` module path MUST be `github.com/CHE3MZ/gitagger`.
* Example: `module github.com/CHE3MZ/gitagger` + `go 1.23` (or current stable).
* All imports use that prefix, e.g. `github.com/CHE3MZ/gitagger/internal/detect`.

### 12.2 Code placement

* `cmd/gitagger/main.go` — CLI ONLY. Parse args/flags, load config, call `internal/`, print result, exit code. NO tag math, NO git exec, NO detection logic here. Keep under ~150 lines.
* `internal/` — ALL logic + src code goes here. Suggested packages:
  * `internal/config/` — find/load/validate `.gitagger.yml`, precedence `flags > config > defaults`
  * `internal/detect/` — classify tags, majority vote, `v`-prefix detect
  * `internal/next/` (or `version/`) — increment `triple/double/single/date`, pre promotion, custom render
  * `internal/git/` — thin wrappers around `git` CLI: `tag list`, `tag create`, `remote get-url`, `ls-remote`, `push`, `status`, `rev-list`. All exec isolated here for mocking.
  * `internal/push/` (or inside `git/`) — §8 remote safety sequence, never called from `main.go` directly
  * `internal/cmd/` — `runInit/runList/runDoctor/runTag` orchestration (CLI calls these)
* Rule: if you are writing `os/exec` or regex outside `internal/` → move it.

### 12.3 Tests

* All tests go into `tests/` (repo convention), NOT scattered.
* `tests/` uses real temp git repos (`git init`, commits, tags) for integration: detect matrix, next-tag matrix, no-remote skip, offline skip, collision abort, `--dry-run`, `--no-push`, `--json`.
* Unit tables for `triple/date` bumps + promotion rules.
* `go test ./...` must pass before every commit.

### 12.4 Ops scripts — run very frequently

* `scripts/ops/go-bugcheck.sh` (golangci-lint, gopls, govulncheck, staticcheck, errcheck, vet, test) and `scripts/ops/go-seccheck.sh` (gosec) MUST be run:
  * after each package is added/changed,
  * before every commit/push,
  * when debugging anything weird — run first, guess second.
* They write to `logs/` — check `logs/go-bugcheck-*.txt` on failure, fix, re-run until green.
* Do not merge/commit while either script is red. `test` failing inside bugcheck counts as bug — fix code, not script.
* CI should run the same two scripts.

### 12.5 Simplicity ladder (how easy vs. powerful stays separated)

* Level 0 — zero thinking: `gitagger` (auto everything + auto-push if safe).
* Level 1 — one word: `gitagger minor`, `gitagger major --pre rc`, `gitagger --format date`, `gitagger --no-push`, `gitagger --dry-run`.
* Level 2 — repeatable: `gitagger init` once, edit `.gitagger.yml`, then back to `gitagger`.
* Level 3 — power: `--custom`, `--message`, `--remote`, `--force`, `--json`, `doctor`, `list --limit`. Normal users never need this.
