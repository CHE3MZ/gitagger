# Tagging

## How the next tag is chosen

1. **Format.** `--format` wins; otherwise `auto` detects it from your last 20 tags by majority vote (ties go to the newest). Fresh repos start at triple with a `v` prefix.
2. **Previous tag.** The highest version in that format (not just the newest date — back-tagged histories are handled).
3. **Bump.** `major`/`minor`/`patch` from the command line, falling back to config, defaulting to `patch`.
4. **Flavor.** `--pre` (or config): `stable` means no suffix; otherwise `-rc`, `-beta`, `-build`, `-nightly`. Promoting `v1.2.4-rc` to stable yields `v1.2.4` (no bump); switching flavors swaps the suffix.

## Formats

| Format | Looks like | Bumps |
|---|---|---|
| `triple` | `v1.2.3` | Per scale |
| `double` | `v1.2` | `major` bumps major, else minor |
| `single` | `v7` | Any scale bumps the number |
| `date` | `v2026.10.04` | Today's date (`.N` on collision) |
| `sha` | `master-abc1234` | Current short SHA |
| `sha-num` | `master-99-abc1234` | Commit count + short SHA |
| `custom` | Your template | Config only (see below) |

The `v` prefix follows your history: keep it if most tags have it.

## Custom format (config only)

Set `format: custom` plus a `custom:` template. Tokens: `MAJOR MINOR PATCH YEAR MONTH DAY DATE SHA FULLSHA COUNT PRE NUMBER TAG`.

```yaml
format: custom
custom: "build-<SHA>-v<MAJOR>.<MINOR>.<PATCH>-<DATE><PRE>"
```

No auto-detect for custom — explicit is on purpose.

## Safety rails

- **HEAD already tagged** → aborts (exit 2) unless `-f`.
- **Tag exists locally or remotely** → aborts unless `-f` (and never silently overwrites a push).
- **Dirty tree** → warns and continues, unless `require_clean` is set (then it aborts).
- **No remote / offline** → keeps the tag locally and tells you the exact `git push` to run later.
- **`--dry-run`** → prints the plan, changes nothing. Use it liberally.
- **`doctor: true`** → checks the remote for collisions *before* creating the tag, so a doomed tag never gets created locally.
- **Failing success hooks** → the tag is rolled back locally and removed from the remote (an overwritten tag is restored instead); nothing broken survives.
- **`--unsafe` / `unsafe: true`** → escape hatch: skips the aborts above, skips rollback, pushes anyway. Loud warnings included. Not written by `init` — add it by hand.