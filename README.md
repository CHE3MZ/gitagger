<img src="assets/framed-icon-labelled.png" alt="Gitagger Logo" width="360" />

# Gitagger

[![CI](https://github.com/CHE3MZ/gitagger/actions/workflows/ci.yml/badge.svg)](https://github.com/CHE3MZ/gitagger/actions/workflows/ci.yml)
[![Tests](https://github.com/CHE3MZ/gitagger/actions/workflows/action-test.yml/badge.svg)](https://github.com/CHE3MZ/gitagger/actions/workflows/action-test.yml)
[![Releases](https://github.com/CHE3MZ/gitagger/actions/workflows/gitagger.yml/badge.svg)](https://github.com/CHE3MZ/gitagger/actions/workflows/gitagger.yml)
[![GitHub release](https://img.shields.io/github/v/release/CHE3MZ/gitagger?display_name=tag&sort=semver&color=32CA55&style=flat&logo=rocket&logoColor=99A0A8&labelColor=373F46&label=Release)](https://github.com/CHE3MZ/gitagger/releases/latest)

### Gitagger makes tag creation for git projects *ridiculously* <u> simple and easy! </u>

Just [Install](#installation) it and run `gitagger` and it will integrate with your existing tag format or initialize one for you.

![demo](assets/demo.gif)

### Full documentation [**here!**](https://che3mz.github.io/gitagger/)

## Features

### ➙ Cross Platform
<img src="assets/features/platforms.png" alt="features-platforms" width="176" />

Gitagger is available on Windows, MacOS and Linux!

### ➙ Hooks
<img src="assets/features/hooks.png" alt="features-hooks" width="176" />

You can make gitagger run basic hooks via the .gitagger.yml file! you can check out this projects own config and usage of hooks [here!](.gitagger.yml)

### ➙ Integrations
<img src="assets/features/integrations.png" alt="features-integrations" width="176" />

Gitagger has native integrations for GitHub Actions, Docker and Jenkins!

## Installation

### ➙ Install with Go 

<img src="assets/installation-methods/go.png" alt="install-go" width="176" />

Requires [Go 1.26](https://go.dev/dl/) or newer:

```sh
go install github.com/CHE3MZ/gitagger/cmd/gitagger@latest
```

### ➙ Prebuilt binaries

<img src="assets/installation-methods/prebuilt.png" alt="install-prebuilt" width="176" />

You can also grab a ready-made binary from the [releases page](https://github.com/CHE3MZ/gitagger/releases) (Linux, macOS, Windows).

### ➙ GitHub Actions

<img src="assets/installation-methods/actions.png" alt="install-actions" width="176" />

```yaml
- uses: CHE3MZ/gitagger@v1
```

[![Marketplace](https://img.shields.io/badge/Marketplace-181717?style=social&logo=github)](https://github.com/marketplace/actions/gitagger)

### ➙ Docker

<img src="assets/installation-methods/docker.png" alt="install-docker" width="176" />

```sh
docker build -t gitagger .
docker run --rm -v "$PWD:/repo" -w /repo gitagger [args]
```

### ➙ Compile from source

<img src="assets/installation-methods/source.png" alt="install-source" width="176" />

Requires [Go 1.26](https://go.dev/dl/) or newer:

```sh
git clone https://github.com/CHE3MZ/gitagger.git
cd gitagger
go build -o gitagger ./cmd/gitagger
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

## Development

#### Contributing
If you'd like to contribute to the project you can check out the contribution guide [**here!**](CONTRIBUTING.md)

#### TODO
The TODO tasks for this project are defined in the TODO.md file [**here!**](TODO.md)

#### State
Gitagger is still pretty early in development and not super polished, but it'll get better with time.

## License

Gitagger is Licensed under the **[Apache 2.0 License](LICENSE).**

<img src="assets/round-icon.png" alt="Gitagger End Section Logo" width="192" />