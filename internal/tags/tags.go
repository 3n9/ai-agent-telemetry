package tags

import (
	"regexp"
	"strings"
)

// Recommended starter vocabularies.
var (
	RecommendedWorkTypes = map[string]bool{
		"coding": true, "debugging": true, "research": true, "analysis": true,
		"writing": true, "planning": true, "creative": true, "support": true, "refactor": true,
	}
	RecommendedLanguages = map[string]bool{
		"php": true, "javascript": true, "typescript": true, "python": true, "sql": true,
		"html": true, "css": true, "shell": true, "json": true, "yaml": true,
		"markdown": true, "none": true,
	}
	RecommendedDomains = map[string]bool{
		"frontend": true, "backend": true, "database": true, "devops": true,
		"documentation": true, "wordpress": true, "laravel": true, "api": true,
		"testing": true, "fiction": true, "horror": true, "email": true,
		"blog": true, "marketing": true, "none": true,
	}
)

var invalidChars = regexp.MustCompile(`[^a-z0-9\-]`)

// Normalize lowercases, trims, replaces spaces with hyphens, strips invalid chars.
func Normalize(tag string) string {
	tag = strings.ToLower(strings.TrimSpace(tag))
	tag = strings.ReplaceAll(tag, " ", "-")
	tag = invalidChars.ReplaceAllString(tag, "")
	return tag
}

// NormalizeAll normalizes a slice of tags and deduplicates them.
func NormalizeAll(raw []string) []string {
	seen := make(map[string]bool, len(raw))
	out := make([]string, 0, len(raw))
	for _, t := range raw {
		n := Normalize(t)
		if n != "" && !seen[n] {
			seen[n] = true
			out = append(out, n)
		}
	}
	return out
}

// Source returns "recommended" or "custom" for a tag value given a vocabulary map.
func Source(value string, vocab map[string]bool) string {
	if vocab[value] {
		return "recommended"
	}
	return "custom"
}
