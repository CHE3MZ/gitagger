// Did-you-mean suggestions for mistyped commands.
// Dynamic: Damerau-Levenshtein (insert/delete/substitute/transpose)
// over the live KnownCommands, so new commands are covered for free.
package cmd

import (
	"sort"
	"strings"

	"github.com/CHE3MZ/gitagger/internal/style"
)

// SuggestCandidates are the names did-you-mean searches: every command
// except the ls/rm aliases (list/remove cover them).
func SuggestCandidates() []string {
	var out []string
	for _, c := range KnownCommands {
		if c != "ls" && c != "rm" {
			out = append(out, c)
		}
	}
	return out
}

// distance is optimal-string-alignment Damerau-Levenshtein (case-insensitive).
// Transpositions cost 1, so "hepl" still finds "help".
func distance(a, b string) int {
	ar := []rune(strings.ToLower(a))
	br := []rune(strings.ToLower(b))
	la, lb := len(ar), len(br)
	d := make([][]int, la+1)
	for i := range d {
		d[i] = make([]int, lb+1)
		d[i][0] = i
	}
	for j := 0; j <= lb; j++ {
		d[0][j] = j
	}
	for i := 1; i <= la; i++ {
		for j := 1; j <= lb; j++ {
			cost := 0
			if ar[i-1] != br[j-1] {
				cost = 1
			}
			d[i][j] = min(d[i-1][j]+1, d[i][j-1]+1, d[i-1][j-1]+cost)
			if i > 1 && j > 1 && ar[i-1] == br[j-2] && ar[i-2] == br[j-1] {
				d[i][j] = min(d[i][j], d[i-2][j-2]+1)
			}
		}
	}
	return d[la][lb]
}

// WorkflowSuggestCandidates are the workflow subcommand words did-you-mean searches.
func WorkflowSuggestCandidates() []string { return []string{"init", "list", "help"} }

// Suggest returns up to 3 candidates similar to input, best first.
// Empty when nothing is close (scales with input length, capped at 3).
func Suggest(input string, candidates []string) []string {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil
	}
	allow := len([]rune(input)) / 3
	if allow < 1 {
		allow = 1
	}
	if allow > 3 {
		allow = 3
	}
	type hit struct {
		name string
		dist int
	}
	var hits []hit
	for _, c := range candidates {
		if strings.EqualFold(input, c) {
			continue
		}
		if d := distance(input, c); d <= allow {
			hits = append(hits, hit{c, d})
		}
	}
	sort.Slice(hits, func(i, j int) bool {
		if hits[i].dist != hits[j].dist {
			return hits[i].dist < hits[j].dist
		}
		return hits[i].name < hits[j].name
	})
	if len(hits) > 3 {
		hits = hits[:3]
	}
	out := make([]string, 0, len(hits))
	for _, h := range hits {
		out = append(out, h.name)
	}
	return out
}

// SuggestionText formats matches for error output: "The most similar
// command is" (one) or "The most similar commands are" (several),
// each name green on its own line. Empty when there is nothing to show.
func SuggestionText(matches []string) string {
	if len(matches) == 0 {
		return ""
	}
	var b strings.Builder
	if len(matches) == 1 {
		b.WriteString("\n\nThe most similar command is\n")
	} else {
		b.WriteString("\n\nThe most similar commands are\n")
	}
	for _, m := range matches {
		b.WriteString("        " + style.Bold(style.White(m)) + "\n")
	}
	return strings.TrimSuffix(b.String(), "\n")
}
