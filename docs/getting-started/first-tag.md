# Your first tag

You need a git repo with at least one commit. Then:

```sh
cd your-project
gitagger --dry-run
```

Dry-run prints the plan and changes nothing:

```text
dry run: nothing created
next: v1.0.0  (triple, stable)
would push to origin
```

No tags yet, so gitagger starts at `v1.0.0`. With existing tags, it detects the format (`triple`, `double`, `single`, `date`, `sha`, and others) and the `v` prefix from history, then follows them. When the plan looks right, run it for real:

```sh
gitagger
```

That creates the tag and pushes it to `origin` when it can (no remote or offline? it keeps the tag locally and tells you how to push later).

## Make it yours

Run `gitagger init` once to write a `.gitagger.yml`, edit it, then run plain `gitagger`: it uses your config from then on. See [Configuration](../usage/configuration.md).

## Useful next commands

```sh
gitagger list        # all tags, oldest first
gitagger view v1.0.0  # details for one tag
gitagger check -v    # is my config valid?
gitagger doctor      # local vs remote tag health
gitagger --help      # everything else
```

Full reference: [Commands](../usage/commands.md).
