// Custom templates: pinned-only formats rendered from a user template.
// Tokens use <NAME> brackets. Dots, dashes and other literal characters
// are always plain text — only bracketed names mean something. Bare words
// like SHA without brackets are literal text and can never collide.
// PRE renders with its dash (`-rc`) or empty for stable: write `<PRE>`,
// never `-<PRE>` (same suffix rule as triple tags).
package next

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// customTokens lists every valid template token.
var customTokens = []string{
	"MAJOR", "MINOR", "PATCH",
	"YEAR", "MONTH", "DAY", "DATE",
	"SHA", "FULLSHA", "COUNT",
	"PRE", "NUMBER", "TAG",
}

// versionTokens are bumped from the previous tag's semver.
var versionTokens = []string{"MAJOR", "MINOR", "PATCH"}

var semverInTag = regexp.MustCompile(`(\d+)\.(\d+)\.(\d+)`)

func validCustomToken(name string) bool {
	for _, t := range customTokens {
		if name == t {
			return true
		}
	}
	return false
}

func templateUses(tmpl string, names ...string) bool {
	for _, n := range names {
		if strings.Contains(tmpl, "<"+n+">") {
			return true
		}
	}
	return false
}

// ValidateTemplate rejects empty, token-less, or malformed templates.
func ValidateTemplate(tmpl string) error {
	if strings.TrimSpace(tmpl) == "" {
		return fmt.Errorf("custom template is empty")
	}
	seen := false
	rest := tmpl
	for {
		i := strings.Index(rest, "<")
		if i < 0 {
			break
		}
		j := strings.Index(rest[i:], ">")
		if j < 0 {
			return fmt.Errorf("custom template has unclosed < in %q", tmpl)
		}
		name := rest[i+1 : i+j]
		if !validCustomToken(name) {
			return fmt.Errorf("custom template has unknown token <%s> (want %s)", name, strings.Join(customTokens, "|"))
		}
		seen = true
		rest = rest[i+j+1:]
	}
	if strings.Contains(rest, ">") {
		return fmt.Errorf("custom template has stray > in %q", tmpl)
	}
	if !seen {
		return fmt.Errorf("custom template has no tokens — it would create the same tag every time")
	}
	stripped := tmpl
	for _, t := range customTokens {
		stripped = strings.ReplaceAll(stripped, "<"+t+">", "")
	}
	for _, r := range stripped {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '.' || r == '_' || r == '-' {
			continue
		}
		return fmt.Errorf("custom template has illegal character %q outside tokens (want [a-zA-Z0-9._-])", string(r))
	}
	return nil
}

// custom renders req.Custom. Version tokens bump from the previous tag's
// semver; without version tokens scale is ignored (like date/sha).
// NUMBER stays empty unless the render collides, then counts 1, 2, ...
func custom(req Request, scale, pre string) (string, error) {
	if err := ValidateTemplate(req.Custom); err != nil {
		return "", err
	}
	maj, min, pch := 1, 0, 0
	if templateUses(req.Custom, versionTokens...) {
		if req.Prev != "" {
			m := semverInTag.FindStringSubmatch(req.Prev)
			if m == nil {
				return "", fmt.Errorf("custom template needs MAJOR/MINOR/PATCH but previous tag %q has no version", req.Prev)
			}
			maj, _ = strconv.Atoi(m[1])
			min, _ = strconv.Atoi(m[2])
			pch, _ = strconv.Atoi(m[3])
			switch scale {
			case "major":
				maj, min, pch = maj+1, 0, 0
			case "minor":
				min, pch = min+1, 0
			default:
				pch++
			}
		}
		// Fresh repo: 1.0.0 base, scale ignored (mirrors triple InitialTag).
	}
	now := req.Now
	if now.IsZero() {
		now = time.Now()
	}
	vals := map[string]string{
		"MAJOR": strconv.Itoa(maj), "MINOR": strconv.Itoa(min), "PATCH": strconv.Itoa(pch),
		"YEAR": now.Format("2006"), "MONTH": now.Format("01"), "DAY": now.Format("02"),
		"DATE": now.Format("2006.01.02"),
		"SHA":  req.SHA, "FULLSHA": req.FullSHA, "COUNT": req.Count,
		"PRE": preSuffix(pre),
		"TAG": req.Prev,
	}
	render := func(number string) string {
		s := req.Custom
		for _, tok := range customTokens {
			v := vals[tok]
			if tok == "NUMBER" {
				v = number
			}
			s = strings.ReplaceAll(s, "<"+tok+">", v)
		}
		return s
	}
	if cand := render(""); req.Existing == nil || !req.Existing[cand] {
		return cand, nil
	}
	for n := 1; n < 1000; n++ {
		if cand := render(strconv.Itoa(n)); !req.Existing[cand] {
			return cand, nil
		}
	}
	return "", fmt.Errorf("custom template %q collides — add <NUMBER> or use -f", req.Custom)
}
