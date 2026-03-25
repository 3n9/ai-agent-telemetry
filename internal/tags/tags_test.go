package tags

import (
	"testing"
)

func TestNormalize_Lowercase(t *testing.T) {
	if got := Normalize("HELLO"); got != "hello" {
		t.Errorf("expected %q, got %q", "hello", got)
	}
	if got := Normalize("MixedCase"); got != "mixedcase" {
		t.Errorf("expected %q, got %q", "mixedcase", got)
	}
}

func TestNormalize_TrimWhitespace(t *testing.T) {
	if got := Normalize("  hello  "); got != "hello" {
		t.Errorf("expected %q, got %q", "hello", got)
	}
	if got := Normalize("\thello\t"); got != "hello" {
		t.Errorf("expected %q, got %q", "hello", got)
	}
}

func TestNormalize_SpacesToHyphens(t *testing.T) {
	if got := Normalize("hello world"); got != "hello-world" {
		t.Errorf("expected %q, got %q", "hello-world", got)
	}
	if got := Normalize("a b c"); got != "a-b-c" {
		t.Errorf("expected %q, got %q", "a-b-c", got)
	}
}

func TestNormalize_StripInvalidChars(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello!", "hello"},
		{"foo@bar", "foobar"},
		{"tag_name", "tagname"},
		{"tag.name", "tagname"},
		{"tag#1", "tag1"},
		{"hello-world", "hello-world"}, // hyphen preserved
	}
	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			if got := Normalize(tc.input); got != tc.want {
				t.Errorf("Normalize(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestNormalize_EmptyString(t *testing.T) {
	if got := Normalize(""); got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestNormalize_AlreadyNormalized(t *testing.T) {
	tags := []string{"hello", "hello-world", "abc123", "foo-bar-baz", "a"}
	for _, tag := range tags {
		t.Run(tag, func(t *testing.T) {
			if got := Normalize(tag); got != tag {
				t.Errorf("already-normalized tag %q changed to %q", tag, got)
			}
		})
	}
}

func TestNormalize_TagThatChanges(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Hello World", "hello-world"},
		{"  trimmed  ", "trimmed"},
		{"UPPER", "upper"},
		{"invalid!chars", "invalidchars"},
	}
	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			got := Normalize(tc.input)
			if got != tc.want {
				t.Errorf("Normalize(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestNormalizeAll_Deduplication(t *testing.T) {
	input := []string{"Hello", "hello", "HELLO"}
	got := NormalizeAll(input)
	if len(got) != 1 {
		t.Errorf("expected 1 unique tag after dedup, got %d: %v", len(got), got)
	}
	if got[0] != "hello" {
		t.Errorf("expected %q, got %q", "hello", got[0])
	}
}

func TestNormalizeAll_DedupAfterNormalization(t *testing.T) {
	// "foo bar" and "foo-bar" both normalize to "foo-bar"
	input := []string{"foo bar", "foo-bar"}
	got := NormalizeAll(input)
	if len(got) != 1 {
		t.Errorf("expected 1 tag, got %d: %v", len(got), got)
	}
}

func TestNormalizeAll_RemovesEmptyAfterNormalization(t *testing.T) {
	input := []string{"valid", "!!!", "  ", ""}
	got := NormalizeAll(input)
	if len(got) != 1 || got[0] != "valid" {
		t.Errorf("expected [\"valid\"], got %v", got)
	}
}

func TestNormalizeAll_PreservesOrder(t *testing.T) {
	input := []string{"zebra", "apple", "mango"}
	got := NormalizeAll(input)
	if len(got) != 3 || got[0] != "zebra" || got[1] != "apple" || got[2] != "mango" {
		t.Errorf("expected order preserved, got %v", got)
	}
}

func TestNormalizeAll_MoreThanFiveTags(t *testing.T) {
	// NormalizeAll does not enforce a 5-tag cap; it only deduplicates.
	// The cap is enforced separately by validate.Check.
	input := []string{"a", "b", "c", "d", "e", "f"}
	got := NormalizeAll(input)
	if len(got) != 6 {
		t.Errorf("NormalizeAll should not cap at 5, expected 6 got %d", len(got))
	}
}

func TestSource_WorkType(t *testing.T) {
	recognized := []string{
		"coding", "debugging", "research", "analysis",
		"writing", "planning", "creative", "support", "refactor",
	}
	for _, wt := range recognized {
		t.Run(wt, func(t *testing.T) {
			if got := Source(wt, RecommendedWorkTypes); got != "recommended" {
				t.Errorf("Source(%q, WorkTypes) = %q, want %q", wt, got, "recommended")
			}
		})
	}
	for _, wt := range []string{"unknown", "custom-type", ""} {
		t.Run("unrecognized_"+wt, func(t *testing.T) {
			if got := Source(wt, RecommendedWorkTypes); got != "custom" {
				t.Errorf("Source(%q, WorkTypes) = %q, want %q", wt, got, "custom")
			}
		})
	}
}

func TestSource_Language(t *testing.T) {
	recognized := []string{
		"php", "javascript", "typescript", "python", "sql",
		"html", "css", "shell", "json", "yaml", "markdown", "none",
	}
	for _, lang := range recognized {
		t.Run(lang, func(t *testing.T) {
			if got := Source(lang, RecommendedLanguages); got != "recommended" {
				t.Errorf("Source(%q, Languages) = %q, want %q", lang, got, "recommended")
			}
		})
	}
	if got := Source("cobol", RecommendedLanguages); got != "custom" {
		t.Errorf("unrecognized language should return %q, got %q", "custom", got)
	}
}

func TestSource_Domain(t *testing.T) {
	recognized := []string{
		"frontend", "backend", "database", "devops",
		"documentation", "wordpress", "laravel", "api",
		"testing", "fiction", "horror", "email", "blog", "marketing", "none",
	}
	for _, d := range recognized {
		t.Run(d, func(t *testing.T) {
			if got := Source(d, RecommendedDomains); got != "recommended" {
				t.Errorf("Source(%q, Domains) = %q, want %q", d, got, "recommended")
			}
		})
	}
	if got := Source("gaming", RecommendedDomains); got != "custom" {
		t.Errorf("unrecognized domain should return %q, got %q", "custom", got)
	}
}
