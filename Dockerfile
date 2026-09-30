# gitagger as a container image (any OCI runtime: docker, podman, nerdctl...).
# Build: docker build -t gitagger .
# Multi-arch: docker buildx build --platform linux/amd64,linux/arm64 -t gitagger .
# Run against a repo: docker run --rm -v "$PWD:/repo" -w /repo gitagger [args]
# Match host file ownership: docker run --user "$(id -u):$(id -g)" ...
# The host dir must be readable by the container user (chmod -R a+rwX).
FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG GITAGGER_VERSION=dev
RUN CGO_ENABLED=0 go build -trimpath \
    -ldflags "-s -w -X github.com/CHE3MZ/gitagger/internal/cmd.version=${GITAGGER_VERSION}" \
    -o /out/gitagger ./cmd/gitagger

# Runtime needs git (tagging shells out to it) and CA certs (https pushes).
FROM alpine:3
# Rolling alpine on purpose: always latest git/certs.
# hadolint ignore=DL3018
RUN apk add --no-cache git ca-certificates
COPY --from=build /out/gitagger /usr/local/bin/gitagger
RUN adduser -D -u 1000 gitagger && mkdir -p /repo
USER gitagger
# Bind-mounted repos are owned by another UID (CI checks out as one):
# allow git to operate on them. Single-purpose image, same as CI images do.
RUN git config --global --add safe.directory '*'
WORKDIR /repo
ENTRYPOINT ["gitagger"]
CMD ["--help"]
LABEL org.opencontainers.image.title="gitagger" \
    org.opencontainers.image.description="Automate git tags without the headache." \
    org.opencontainers.image.source="https://github.com/CHE3MZ/gitagger"
