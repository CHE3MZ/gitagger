// Package next computes the next tag name from the previous one.
// It is pure numeration: no git, no network, no filesystem.
// That makes it easy to test without touching any .git directory.
package next

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/CHE3MZ/gitagger/internal/detect"
)

// Request describes how to bump.
type Request struct {
	Prev     string          // previous tag, e.g. "v1.2.3" (empty when no tags)
	Scale    string          // major|minor|patch
	Pre      string          // stable|rc|beta|build|nightly
	Format   detect.Format   // concrete format (never auto)
	VPrefix  bool            // keep leading "v"
	Now      time.Time       // used for date format
	Existing map[string]bool // known tags, for date collision counter
	SHA      string          // short SHA, for sha formats
	Count    string          // rev-list count, for sha-num
}

// InitialTag is the fallback for repos with no tags.
func InitialTag(vPrefix bool) string {
	if vPrefix {
		return "v0.0.1"
	}
	return "0.0.1"
}

func prefix(vPrefix bool) string {
	if vPrefix {
		return "v"
	}
	return ""
}

func preSuffix(pre string) string {
	if pre == "" || pre == "stable" {
		return ""
	}
	return "-" + pre
}

// Compute returns the next tag for req. Prev may be empty.
func Compute(req Request) (string, error) {
	scale := strings.ToLower(strings.TrimSpace(req.Scale))
	if scale == "" {
		scale = "patch"
	}
	pre := strings.ToLower(strings.TrimSpace(req.Pre))
	if pre == "" {
		pre = "stable"
	}
	if !detect.ValidPreCheck(pre) {
		return "", fmt.Errorf("unknown pre %q (want stable|rc|beta|build|nightly)", req.Pre)
	}

	if req.Format == detect.Auto {
		return "", fmt.Errorf("format is still auto — detect it first")
	}
	if req.Format == detect.Custom {
		return "", fmt.Errorf("custom format is planned for v2")
	}
	if req.Prev == "" {
		// Fresh repo: date/sha still derive from now/sha, semver starts over.
		switch req.Format {
		case detect.Date:
			return dateTag(req.Now, req.VPrefix, pre, req.Existing), nil
		case detect.SHA, detect.SHANum:
			return shaTag(req.Format, req.SHA, req.Count, pre)
		default:
			base := InitialTag(req.VPrefix)
			if pre != "stable" {
				// v0.0.1-rc style start
				b, _ := detect.StripPre(base)
				return b + preSuffix(pre), nil
			}
			return base, nil
		}
	}

	switch req.Format {
	case detect.Triple:
		return triple(req.Prev, scale, pre, req.VPrefix)
	case detect.Double:
		return double(req.Prev, scale, pre, req.VPrefix)
	case detect.Single:
		return single(pre, req.Prev, req.VPrefix)
	case detect.Date:
		return dateTag(req.Now, req.VPrefix, pre, req.Existing), nil
	case detect.SHA, detect.SHANum:
		return shaTag(req.Format, req.SHA, req.Count, pre)
	default:
		return "", fmt.Errorf("unsupported format %q in v1", string(req.Format))
	}
}

// triple bumps v1.2.3. Handles promotion: v1.2.4-rc -> v1.2.4 (no bump).
func triple(prev, scale, pre string, vPrefix bool) (string, error) {
	base, prevPre := detect.StripPre(prev)
	nums := strings.TrimPrefix(base, "v")
	parts := strings.Split(nums, ".")
	if len(parts) != 3 {
		return "", fmt.Errorf("tag %q is not triple (want v1.2.3)", prev)
	}
	maj, err1 := strconv.Atoi(parts[0])
	min, err2 := strconv.Atoi(parts[1])
	pch, err3 := strconv.Atoi(parts[2])
	if err1 != nil || err2 != nil || err3 != nil {
		return "", fmt.Errorf("tag %q is not numeric triple", prev)
	}
	if prevPre != "stable" {
		if pre == "stable" {
			// promote: strip suffix, keep numbers
			return fmt.Sprintf("%s%d.%d.%d", prefix(vPrefix), maj, min, pch), nil
		}
		// pre -> pre: swap suffix, no bump
		return fmt.Sprintf("%s%d.%d.%d%s", prefix(vPrefix), maj, min, pch, preSuffix(pre)), nil
	}
	switch scale {
	case "major":
		maj, min, pch = maj+1, 0, 0
	case "minor":
		min, pch = min+1, 0
	default:
		pch++
	}
	return fmt.Sprintf("%s%d.%d.%d%s", prefix(vPrefix), maj, min, pch, preSuffix(pre)), nil
}

// double bumps v1.2: patch/minor bump the minor part, major bumps major.
func double(prev, scale, pre string, vPrefix bool) (string, error) {
	base, prevPre := detect.StripPre(prev)
	nums := strings.TrimPrefix(base, "v")
	parts := strings.Split(nums, ".")
	if len(parts) != 2 {
		return "", fmt.Errorf("tag %q is not double (want v1.2)", prev)
	}
	maj, err1 := strconv.Atoi(parts[0])
	min, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil {
		return "", fmt.Errorf("tag %q is not numeric double", prev)
	}
	if prevPre != "stable" {
		if pre == "stable" {
			return fmt.Sprintf("%s%d.%d", prefix(vPrefix), maj, min), nil
		}
		return fmt.Sprintf("%s%d.%d%s", prefix(vPrefix), maj, min, preSuffix(pre)), nil
	}
	if scale == "major" {
		maj, min = maj+1, 0
	} else {
		min++
	}
	return fmt.Sprintf("%s%d.%d%s", prefix(vPrefix), maj, min, preSuffix(pre)), nil
}

// single bumps v1: any scale bumps the number.
func single(pre, prev string, vPrefix bool) (string, error) {
	base, prevPre := detect.StripPre(prev)
	nums := strings.TrimPrefix(base, "v")
	n, err := strconv.Atoi(nums)
	if err != nil {
		return "", fmt.Errorf("tag %q is not single (want v1)", prev)
	}
	if prevPre != "stable" {
		if pre == "stable" {
			return fmt.Sprintf("%s%d", prefix(vPrefix), n), nil
		}
		return fmt.Sprintf("%s%d%s", prefix(vPrefix), n, preSuffix(pre)), nil
	}
	return fmt.Sprintf("%s%d%s", prefix(vPrefix), n+1, preSuffix(pre)), nil
}

// dateTag renders vYYYY.MM.DD with .N collision counter before the pre suffix.
func dateTag(now time.Time, vPrefix bool, pre string, existing map[string]bool) string {
	if now.IsZero() {
		now = time.Now()
	}
	day := now.Format("2006.01.02")
	cand := prefix(vPrefix) + day + preSuffix(pre)
	if existing == nil || !existing[cand] {
		return cand
	}
	for n := 1; n < 1000; n++ {
		c := fmt.Sprintf("%s%s.%d%s", prefix(vPrefix), day, n, preSuffix(pre))
		if !existing[c] {
			return c
		}
	}
	return fmt.Sprintf("%s%s.%d%s", prefix(vPrefix), day, 999, preSuffix(pre))
}

// shaTag renders master-<sha> or master-<count>-<sha>.
func shaTag(format detect.Format, sha, count, pre string) (string, error) {
	sha = strings.TrimSpace(sha)
	if sha == "" {
		return "", fmt.Errorf("need current commit SHA for sha format (is HEAD missing?)")
	}
	short := sha
	if len(short) > 7 {
		short = short[:7]
	}
	if format == detect.SHANum {
		count = strings.TrimSpace(count)
		if count == "" {
			return "", fmt.Errorf("need commit count for sha-num format")
		}
		return "master-" + count + "-" + short + preSuffix(pre), nil
	}
	return "master-" + short + preSuffix(pre), nil
}
