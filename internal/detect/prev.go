// MaxPrev picks the highest-versioned tag of a format for bumping.
// Newest-by-date is wrong when history is back-tagged (a newer date can
// carry an older version), so compare version numbers instead. Ties keep
// the newest tag. Tags are oldest -> newest. Returns "" when no tag has
// that format (sha/sha-num have no ordering — use newest overall then).
package detect

import (
	"strconv"
	"strings"
)

// versionKey parses a tag of the given format into comparable numbers.
func versionKey(tag string, format Format) ([]int, bool) {
	base, _ := StripPre(strings.TrimSpace(tag))
	base = StripV(base)
	var parts []string
	switch format {
	case Triple:
		if !reTriple.MatchString(base) {
			return nil, false
		}
		parts = strings.Split(base, ".")
	case Double:
		if !reDouble.MatchString(base) {
			return nil, false
		}
		parts = strings.Split(base, ".")
	case Single:
		if !reSingle.MatchString(base) {
			return nil, false
		}
		parts = []string{base}
	case Date:
		if !reDate.MatchString(base) {
			return nil, false
		}
		parts = strings.Split(base, ".")
	default:
		return nil, false
	}
	nums := make([]int, len(parts))
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil {
			return nil, false
		}
		nums[i] = n
	}
	return nums, true
}

func versionGreater(a, b []int) bool {
	for i := 0; i < len(a) && i < len(b); i++ {
		if a[i] != b[i] {
			return a[i] > b[i]
		}
	}
	return len(a) > len(b)
}

// MaxPrev returns the highest-version tag of format, newest wins ties.
// Only tags classified as that format count: a date like v2020.01.02 is
// numerically a triple but belongs to the date family, not triple history.
func MaxPrev(tags []string, format Format) string {
	var best string
	var bestKey []int
	var bestSet bool
	for _, t := range tags {
		if Classify(t) != format {
			continue
		}
		k, ok := versionKey(t, format)
		if !ok {
			continue
		}
		if !bestSet || !versionGreater(bestKey, k) {
			best, bestKey, bestSet = t, k, true
		}
	}
	return best
}
