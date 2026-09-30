// Suggestion tests: typos find their command, gibberish finds nothing.
package tests

import (
	"strings"
	"testing"

	"github.com/CHE3MZ/gitagger/internal/cmd"
)

func TestSuggestTypos(t *testing.T) {
	cands := cmd.SuggestCandidates()
	cases := map[string]string{
		"hlp": "help", "hepl": "help", "HLP": "help",
		"chek": "check", "lst": "list", "vie": "view",
		"remov": "remove", "docter": "doctor", "versoin": "version",
		"handbok": "handbook", "minro": "minor", "mjor": "major",
		"inti": "init", "remtoe": "remote",
	}
	for in, want := range cases {
		got := cmd.Suggest(in, cands)
		if len(got) == 0 || got[0] != want {
			t.Errorf("Suggest(%q) = %v, want first %q", in, got, want)
		}
	}
}

func TestSuggestNothingClose(t *testing.T) {
	cands := cmd.SuggestCandidates()
	for _, in := range []string{"", "xyz", "foo", "bogus", "--", "12345"} {
		if got := cmd.Suggest(in, cands); len(got) != 0 {
			t.Errorf("Suggest(%q) = %v, want none", in, got)
		}
	}
	// Exact matches need no suggestion.
	for _, in := range []string{"help", "list", "HELP"} {
		if got := cmd.Suggest(in, cands); len(got) != 0 {
			t.Errorf("Suggest(%q) = %v, want none (exact)", in, got)
		}
	}
}

func TestSuggestProviders(t *testing.T) {
	if got := cmd.Suggest("hg", []string{"gh", "jenkins"}); len(got) == 0 || got[0] != "gh" {
		t.Errorf(`Suggest("hg") = %v, want [gh]`, got)
	}
	if got := cmd.Suggest("jenkin", []string{"gh", "jenkins"}); len(got) == 0 || got[0] != "jenkins" {
		t.Errorf(`Suggest("jenkin") = %v, want [jenkins]`, got)
	}
	if got := cmd.Suggest("zzz", []string{"gh", "jenkins"}); len(got) != 0 {
		t.Errorf(`Suggest("zzz") = %v, want none`, got)
	}
}

func TestSuggestCandidatesCoverCommands(t *testing.T) {
	cands := cmd.SuggestCandidates()
	for _, c := range cmd.KnownCommands {
		if c == "ls" || c == "rm" {
			continue
		}
		found := false
		for _, s := range cands {
			if s == c {
				found = true
			}
		}
		if !found {
			t.Errorf("candidates miss %q", c)
		}
	}
	for _, s := range cands {
		if s == "ls" || s == "rm" {
			t.Errorf("candidates should not suggest aliases, got %q", s)
		}
	}
}

func TestSuggestionText(t *testing.T) {
	if got := cmd.SuggestionText(nil); got != "" {
		t.Errorf("empty matches = %q, want empty", got)
	}
	one := cmd.SuggestionText([]string{"help"})
	if !strings.Contains(one, "The most similar command is") || !strings.Contains(one, "help") {
		t.Errorf("single = %q, want header + help", one)
	}
	if strings.Contains(one, "commands are") {
		t.Errorf("single = %q, must not use plural", one)
	}
	two := cmd.SuggestionText([]string{"view", "remove"})
	if !strings.Contains(two, "The most similar commands are") {
		t.Errorf("multi = %q, want plural header", two)
	}
}

func TestParseScaleSuggests(t *testing.T) {
	_, err := cmd.ParseScale([]string{"hlp"})
	if err == nil {
		t.Fatalf("hlp should fail")
	}
	if !strings.Contains(err.Error(), "most similar") || !strings.Contains(err.Error(), "help") {
		t.Errorf("hlp error = %q, want suggestion of help", err.Error())
	}
	_, err = cmd.ParseScale([]string{"bogus"})
	if err == nil {
		t.Fatalf("bogus should fail")
	}
	if strings.Contains(err.Error(), "similar") {
		t.Errorf("bogus error = %q, want no suggestion", err.Error())
	}
}

func TestWorkflowInitSuggestsProvider(t *testing.T) {
	err := cmd.RunWorkflowInit(t.TempDir(), "hg", false, false)
	if err == nil {
		t.Fatalf("hg provider should fail")
	}
	if !strings.Contains(err.Error(), "gh") {
		t.Errorf("hg error = %q, want suggestion of gh", err.Error())
	}
}
