# Installation

Pick whichever fits. They all give you the same `gitagger` binary.

## Go

<img src="../assets/installation-methods/go.png" alt="Go" width="64" />

```sh
go install github.com/CHE3MZ/gitagger/cmd/gitagger@latest
```

Requires Go 1.26 or newer.

## Prebuilt binaries

<img src="../assets/installation-methods/prebuilt.png" alt="Prebuilt binaries" width="64" />

Grab one from the [releases page](https://github.com/CHE3MZ/gitagger/releases) — Linux, macOS, and Windows, amd64 and arm64.

## From source

<img src="../assets/installation-methods/source.png" alt="Source" width="64" />

```sh
git clone https://github.com/CHE3MZ/gitagger.git
cd gitagger
go build -o gitagger ./cmd/gitagger
```

Requires Go 1.26 or newer.

## Docker

<img src="../assets/installation-methods/docker.png" alt="Docker" width="64" />

No published image yet — build it yourself (any OCI runtime: docker, podman, nerdctl…):

```sh
docker build -t gitagger .
docker run --rm -v "$PWD:/repo" -w /repo gitagger [args]
```

Match host file ownership with `--user "$(id -u):$(id -g)"`. The mounted directory must be readable by the container user.

## GitHub Action

<img src="../assets/installation-methods/actions.png" alt="GitHub Actions" width="64" />

```yaml
- uses: CHE3MZ/gitagger@v1
```

See the [GitHub Action guide](../automation/github-action.md) for inputs (`args`, `version`, `working-directory`) and details.

## Next step

[Cut your first tag](first-tag.md).
