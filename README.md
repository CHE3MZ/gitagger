<img src="assets/framed-icon-labelled.png" alt="Gitagger Logo" width="360" />

# Gitagger

[![CI](https://github.com/CHE3MZ/gitagger/actions/workflows/ci.yml/badge.svg)](https://github.com/CHE3MZ/gitagger/actions/workflows/ci.yml)
[![GitHub release](https://img.shields.io/github/v/release/CHE3MZ/gitagger?display_name=tag&sort=semver&color=32CA55&style=flat&logo=rocket&logoColor=99A0A8&labelColor=373F46&label=Release)](https://github.com/CHE3MZ/gitagger/releases/latest)
[![Container](https://img.shields.io/github/actions/workflow/status/CHE3MZ/gitagger/container.yml?style=flat&logo=githubactions&logoColor=99A0A8&labelColor=373F46&label=Container&color=32CA55)](https://github.com/CHE3MZ/gitagger/actions/workflows/container.yml)

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

Gitagger has integrations for Docker and containers in general!

## Installation

### ➙ Install with Go 

<img src="assets/installation-methods/go.png" alt="install-go" width="176" />

Requires [Go 1.27](https://go.dev/dl/) or newer:

```sh
go install github.com/CHE3MZ/gitagger/cmd/gitagger@latest
```

### ➙ Prebuilt binaries

<img src="assets/installation-methods/prebuilt.png" alt="install-prebuilt" width="176" />

You can also download a binary from the [releases page](https://github.com/CHE3MZ/gitagger/releases) for Linux, macOS, or Windows.

### ➙ Docker

<img src="assets/installation-methods/docker.png" alt="install-docker" width="176" />

```sh
docker pull ghcr.io/che3mz/gitagger:latest
```

[![Github Packages](https://img.shields.io/badge/Packages_Registry-181717?style=social&logo=github)](https://github.com/che3mz/gitagger/pkgs/container/gitagger/)

### ➙ Compile from source

<img src="assets/installation-methods/source.png" alt="install-source" width="176" />

#### Requires [Go 1.27](https://go.dev/dl/) or newer:

```sh
git clone https://github.com/CHE3MZ/gitagger.git
cd gitagger
go build -o gitagger ./cmd/gitagger
```

#### Building the Container Image:

```sh
docker build -t gitagger .
docker run --rm -v "$PWD:/repo" -w /repo gitagger [args]
```

## Usage

<img src="assets/demo-help.svg" alt="demo-help" width="700" />

```sh
gitagger
gitagger minor
gitagger --help
```

Run `gitagger init` once to write `.gitagger.yml`, edit it, then run plain `gitagger` without flags.

### Full documentation [**here!**](https://che3mz.github.io/gitagger/)

## Contributing

See the **[contribution guide](.github/CONTRIBUTING.md).**

## License

Gitagger is licensed under the **[Apache 2.0 License](LICENSE).**

<img src="assets/round-icon.png" alt="Gitagger End Section Logo" width="192" />
