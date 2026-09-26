// Package git wraps the user's git from PATH.
// All functions take an explicit dir and run `git -C <dir> ...`,
// so tests can use temp dirs and never touch the repo's own .git.
// If git is not on PATH, every helper returns a friendly error.
package git

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

const lookName = "git"

// EnsureAvailable errors when git cannot be found on PATH.
func EnsureAvailable() error {
	if _, err := exec.LookPath(lookName); err != nil {
		return fmt.Errorf("git not found on PATH — install git and try again")
	}
	return nil
}

// run executes git in dir with a timeout.
func run(dir string, timeout time.Duration, args ...string) (string, error) {
	if err := EnsureAvailable(); err != nil {
		return "", err
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	full := []string{}
	if dir != "" {
		full = append(full, "-C", dir)
	}
	full = append(full, args...)
	cmd := exec.CommandContext(ctx, lookName, full...)
	out, err := cmd.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		return "", fmt.Errorf("git %s timed out", strings.Join(args, " "))
	}
	if err != nil {
		return strings.TrimSpace(string(out)), fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	return strings.TrimSpace(string(out)), nil
}

func runDefault(dir string, args ...string) (string, error) {
	return run(dir, 30*time.Second, args...)
}

// ListTags returns tags oldest -> newest (creatordate order).
func ListTags(dir string) ([]string, error) {
	out, err := runDefault(dir, "for-each-ref", "--sort=creatordate", "--format=%(refname:short)", "refs/tags")
	if err != nil {
		// Empty repo / no tags: for-each-ref prints nothing with exit 0,
		// but be tolerant and fall back.
		if out == "" {
			return []string{}, nil
		}
		return nil, err
	}
	if out == "" {
		return []string{}, nil
	}
	lines := strings.Split(out, "\n")
	tags := make([]string, 0, len(lines))
	for _, l := range lines {
		if t := strings.TrimSpace(l); t != "" {
			tags = append(tags, t)
		}
	}
	return tags, nil
}

// HeadExists reports whether dir has at least one commit.
func HeadExists(dir string) bool {
	_, err := runDefault(dir, "rev-parse", "--verify", "HEAD")
	return err == nil
}

// ShortSHA returns the short HEAD sha.
func ShortSHA(dir string) (string, error) { return runDefault(dir, "rev-parse", "--short", "HEAD") }

// FullSHA returns the full HEAD sha.
func FullSHA(dir string) (string, error) { return runDefault(dir, "rev-parse", "HEAD") }

// Count returns git rev-list --count HEAD.
func Count(dir string) (string, error) { return runDefault(dir, "rev-list", "--count", "HEAD") }

// TagsPointingAtHEAD lists tags on HEAD (empty when untagged).
func TagsPointingAtHEAD(dir string) ([]string, error) {
	out, err := runDefault(dir, "tag", "--points-at", "HEAD")
	if err != nil {
		return nil, err
	}
	if out == "" {
		return []string{}, nil
	}
	var tags []string
	for _, l := range strings.Split(out, "\n") {
		if t := strings.TrimSpace(l); t != "" {
			tags = append(tags, t)
		}
	}
	return tags, nil
}

// StatusPorcelain returns git status --porcelain (empty = clean).
func StatusPorcelain(dir string) (string, error) { return runDefault(dir, "status", "--porcelain") }

// RemoteURL returns the URL of remote or an error when missing.
func RemoteURL(dir, remote string) (string, error) {
	out, err := runDefault(dir, "remote", "get-url", remote)
	if err != nil {
		return "", fmt.Errorf("no remote %q", remote)
	}
	return out, nil
}

// LsRemoteTags runs ls-remote with a short timeout for the offline check.
func LsRemoteTags(dir, remote string) (string, error) {
	return run(dir, 10*time.Second, "ls-remote", "--tags", remote)
}

// RemoteHasTag parses ls-remote output for refs/tags/<tag>.
func RemoteHasTag(lsRemoteOut, tag string) bool {
	needle := "refs/tags/" + tag
	for _, line := range strings.Split(lsRemoteOut, "\n") {
		if strings.Contains(strings.TrimSpace(line), needle) {
			// peel ^{} lines still count as having the tag
			return true
		}
	}
	return false
}

// CreateTag makes a lightweight tag, or annotated with -m.
func CreateTag(dir, tag, message string, force bool) error {
	args := []string{"tag"}
	if force {
		args = append(args, "-f")
	}
	if message != "" {
		args = append(args, "-a", tag, "-m", message)
	} else {
		args = append(args, tag)
	}
	_, err := runDefault(dir, args...)
	return err
}

// PushTag pushes exactly one ref. Never --tags.
func PushTag(dir, remote, tag string, force bool) error {
	args := []string{"push", remote, tag}
	if force {
		args = []string{"push", "--force", remote, tag}
	}
	_, err := run(dir, 60*time.Second, args...)
	return err
}

// FetchTags runs git fetch --tags (doctor, read-only).
func FetchTags(dir, remote string) error {
	_, err := run(dir, 60*time.Second, "fetch", "--tags", remote)
	return err
}
