# Configuration

`gitagger init` writes a `.gitagger.yml` (recognized siblings: `.gitagger.yaml`, `.gitagger` — first found wins). Everything is optional. Precedence: **flags > config > defaults**. Run `gitagger check -v` to see the resolved values, or `gitagger handbook` for this same reference in your terminal.

```yaml
# Tag size for triple/double/single styles.
scale: patch            # major | minor | patch

# Flavor appended to the tag. stable = no suffix.
pre: stable              # stable | rc | beta | build | nightly

# Tag style. auto follows your tag history.
format: auto              # auto | triple | double | single | date | sha | sha-num | custom

# Custom template. Only used when format is custom.
custom: ""                # e.g. "build-<SHA>-v<MAJOR>.<MINOR>.<PATCH>-<DATE><PRE>"

# Push target. Must exist for auto-push to happen.
remote: origin

# Auto-push the new tag when the remote is reachable.
push: true                # false keeps tags local (same as -n)

# Always overwrite clashing tags and force-push them.
force: false              # same as -f on every run. Keep false unless you mean it.

# Abort instead of tagging when the working tree is dirty.
require_clean: false

# Pre-tag remote check. Aborts early on collision.
doctor: false

# Show detection details and the plan while working.
verbose: false            # same as -v on every run

# Tag message. Empty = lightweight tag, set = annotated tag.
message: ""

# Operate in another directory (same as -p).
path: ""

# Hooks: shell commands at lifecycle events. See Hooks.
on:
  failure:
    - shell: sh
      os: macos, linux
      run:
        - echo "Oops! Something went wrong..."
    - shell: batch
      os: windows
      run:
        - echo "Oops! Something went wrong..."
```

!!! note "Strict but forgiving"
    Unknown config keys are an error (typos under known sections too) — except retired keys (`version:`, `confirm:`), which still load fine. A broken config aborts before anything mutates; a missing file just means defaults.
