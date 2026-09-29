<img src="assets/framed-icon-labelled.png" alt="Gitagger Logo" width="360" />

# Gitagger

### Gitagger makes tag creation for git projects *ridiculously* <u> simple and easy! </u>

Just [Install](#installation) it and run `gitagger` and it will integrate with your existing tag format or initialize one for you.

## Installation

```sh
go install github.com/CHE3MZ/gitagger/cmd/gitagger@latest
```

Or grab a ready-made binary from the [releases page](https://github.com/CHE3MZ/gitagger/releases) (Linux, macOS, Windows).

For CI, there is a GitHub Action:

```yaml
- uses: CHE3MZ/gitagger@v1
```

## Usage

```sh
gitagger                 # next patch tag, pushed if it can be
gitagger minor           # v1.2.3 -> v1.3.0
gitagger major --pre rc  # v1.2.3 -> v2.0.0-rc
gitagger --dry-run       # peek first, change nothing
gitagger --help          # everything else
```

Want the same settings every time? Run `gitagger init` once, 
tweak `.gitagger.yml`, then forget about flags.
running `gitagger` without any arguments will use your new config file.

Full docs are coming later — for now, `gitagger help <command>` tells you what each command does.

## License

Gitagger is Licensed under the **[Apache 2.0 License](LICENSE).**

<img src="assets/round-icon.png" alt="Gitagger End Section Logo" width="192" />