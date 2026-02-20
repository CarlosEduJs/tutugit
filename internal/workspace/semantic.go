package workspace

import (
	"regexp"
	"strings"
)

// commitPrefixRegex matches conventional commit prefixes like "feat:", "fix(scope):", "feat!:", etc.
var commitPrefixRegex = regexp.MustCompile(`(?i)^(feat|feature|fix|bugfix|refactor|experiment|exp)(\([^)]*\))?(!)?:\s*`)

// prefixToTag maps recognized prefixes to their semantic tag.
var prefixToTag = map[string]string{
	"feat":       "feature",
	"feature":    "feature",
	"fix":        "fix",
	"bugfix":     "fix",
	"refactor":   "refactor",
	"experiment": "experiment",
	"exp":        "experiment",
}

// DetectTag parses a commit message and returns the semantic tag based on
// its conventional commit prefix. Returns "none" if no known prefix is found.
func DetectTag(message string) string {
	msg := strings.TrimSpace(message)
	match := commitPrefixRegex.FindStringSubmatch(msg)
	if len(match) < 2 {
		return "none"
	}
	prefix := strings.ToLower(match[1])
	if tag, ok := prefixToTag[prefix]; ok {
		return tag
	}
	return "none"
}

// DetectImpact suggests the impact level (patch, minor, major) based on the commit message.
// - major: if it contains a breaking change indicator (!) in the prefix or 'BREAKING CHANGE' in body.
// - minor: if it's a 'feat' or 'feature'.
// - patch: for 'fix', 'refactor', etc.
func DetectImpact(message string) string {
	msg := strings.TrimSpace(message)
	lines := strings.Split(msg, "\n")
	header := lines[0]

	// MAJOR: breaking change indicator (!) in prefix, e.g., "feat!:"
	match := commitPrefixRegex.FindStringSubmatch(header)
	if len(match) > 0 {
		// Group 3 is the optional !
		if match[3] == "!" {
			return "major"
		}
	}

	// MAJOR: "BREAKING CHANGE" in body
	if strings.Contains(msg, "BREAKING CHANGE") {
		return "major"
	}

	if len(match) > 0 {
		prefix := strings.ToLower(match[1])
		// MINOR: features
		if prefix == "feat" || prefix == "feature" {
			return "minor"
		}
	}

	// PATCH: default for everything else (fix, refactor, or non-prefix)
	return "patch"
}
