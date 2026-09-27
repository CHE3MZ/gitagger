// Package detect figures out which tag style a repo already uses,
// so bare `gitagger` can just follow the convention.
package detect

import (
	"regexp"
	"strings"
)

// Format is a tag archetype.
type Format string

const (
	Auto    Format = "auto"
	Triple  Format = "triple"
	Double  Format = "double"
	Single  Format = "single"
	Date    Format = "date"
	SHA     Format = "sha"
	SHANum  Format = "sha-num"
	Custom  Format = "custom"
	Unknown Format = "unknown"
)

// ValidPre lists accepted prerelease channels.
var ValidPre = []string{"stable", "rc", "beta", "build", "nightly"}

// preSuffixes are stripped before classification.
var preSuffixes = []string{"-rc", "-beta", "-build", "-nightly"}

var (
	reTriple = regexp.MustCompile(`^\d+\.\d+\.\d+$`)
	reDouble = regexp.MustCompile(`^\d+\.\d+$`)
	reSingle = regexp.MustCompile(`^\d+$`)
	reDate   = regexp.MustCompile(`^\d{4}\.\d{2}\.\d{2}(\.\d+)?$`)
	reSHA    = regexp.MustCompile(`^master-[0-9a-fA-F]{4,40}$`)
	reSHANum = regexp.MustCompile(`^master-\d+-[0-9a-fA-F]{4,40}$`)
)

// StripPre splits "v1.2.4-rc" into ("v1.2.4", "rc").
// Returns pre "stable" when there is no suffix.
func StripPre(tag string) (base, pre string) {
	for _, s := range preSuffixes {
		if strings.HasSuffix(tag, s) {
			return strings.TrimSuffix(tag, s), strings.TrimPrefix(s, "-")
		}
	}
	return tag, "stable"
}

// StripV removes a leading "v" for classification.
func StripV(tag string) string { return strings.TrimPrefix(tag, "v") }

// Classify maps one tag to its archetype (pre suffix ignored).
func Classify(tag string) Format {
	t := strings.TrimSpace(tag)
	if t == "" {
		return Unknown
	}
	base, _ := StripPre(t)
	base = StripV(base)
	switch {
	case reSHANum.MatchString(base):
		return SHANum
	case reSHA.MatchString(base):
		return SHA
	case reDate.MatchString(base):
		return Date
	case reTriple.MatchString(base):
		return Triple
	case reDouble.MatchString(base):
		return Double
	case reSingle.MatchString(base):
		return Single
	default:
		return Unknown
	}
}

// Detect picks the repo's format by majority vote over tags.
// tags should be oldest -> newest (git for-each-ref --sort=creatordate).
// Unknown tags are ignored. No tags (or all unknown) -> triple.
// Ties go to the most recent tag. vPrefix is true when >=50% use "v".
func Detect(tags []string) (Format, bool) {
	counts := map[Format]int{}
	vCount, known := 0, 0
	var lastKnown Format

	for _, t := range tags {
		f := Classify(t)
		if f == Unknown {
			continue
		}
		counts[f]++
		known++
		lastKnown = f
		if strings.HasPrefix(strings.TrimSpace(t), "v") {
			vCount++
		}
	}
	if known == 0 {
		return Triple, true // fresh repo default
	}
	best, bestN := lastKnown, -1
	for f, n := range counts {
		if n > bestN {
			best, bestN = f, n
		}
	}
	// Tie-break: most recent known format wins when tied.
	for i := len(tags) - 1; i >= 0; i-- {
		f := Classify(tags[i])
		if f == Unknown {
			continue
		}
		if counts[f] == bestN {
			best = f
			break
		}
	}
	return best, vCount*2 >= known
}

// ParseFormat parses --format flag values.
func ParseFormat(s string) (Format, bool) {
	switch Format(strings.ToLower(strings.TrimSpace(s))) {
	case Auto, Triple, Double, Single, Date, SHA, SHANum:
		return Format(strings.ToLower(strings.TrimSpace(s))), true
	default:
		return Unknown, false
	}
}

// ValidPreCheck reports whether a --pre value is allowed.
func ValidPreCheck(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	for _, v := range ValidPre {
		if s == v {
			return true
		}
	}
	return false
}
