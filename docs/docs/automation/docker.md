# Docker

No published image yet — build it yourself (any OCI runtime: docker, podman, nerdctl…):

```sh
docker build -t gitagger .
docker run --rm -v "$PWD:/repo" -w /repo gitagger [args]
```

Multi-arch: `docker buildx build --platform linux/amd64,linux/arm64 -t gitagger .`

## Notes

- The image is non-root (`gitagger` user) with `git` and CA certificates installed — everything tagging needs, nothing else.
- It reads the mounted repo's `.gitagger.yml` like the binary does; all flags work the same.
- Match host file ownership with `--user "$(id -u):$(id -g)"`, and make sure the mounted directory is readable by the container user.

## Examples

```sh
# peek at the plan
docker run --rm -v "$PWD:/repo" -w /repo gitagger --dry-run

# tag without pushing
docker run --rm -v "$PWD:/repo" -w /repo gitagger --no-push

# inspect
docker run --rm -v "$PWD:/repo" -w /repo gitagger list
```
