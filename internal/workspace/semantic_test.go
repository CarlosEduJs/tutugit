package workspace

import "testing"

func TestDetectTag(t *testing.T) {
	tests := []struct {
		message  string
		expected string
	}{
		{"feat: add new feature", "feature"},
		{"fix: fix a bug", "fix"},
		{"refactor: clean up code", "refactor"},
		{"experiment: try something new", "experiment"},
		{"chore: update deps", "none"},
		{"docs: update readme", "none"},
		{"FEAT: uppercase works", "feature"},
		{"feat!: breaking feature", "feature"},
		{"random message", "none"},
	}

	for _, tt := range tests {
		got := DetectTag(tt.message)
		if got != tt.expected {
			t.Errorf("DetectTag(%q) = %q; want %q", tt.message, got, tt.expected)
		}
	}
}

func TestDetectImpact(t *testing.T) {
	tests := []struct {
		message  string
		expected string
	}{
		{"feat: a new feature", "minor"},
		{"fix: a small fix", "patch"},
		{"refactor: some code cleanup", "patch"},
		{"feat!: a breaking feature", "major"},
		{"fix!: a breaking fix", "major"},
		{"fix: some fix\n\nBREAKING CHANGE: this changes everything", "major"},
		{"feat: a feature\n\nSome body text", "minor"},
		{"random message", "patch"},
	}

	for _, tt := range tests {
		got := DetectImpact(tt.message)
		if got != tt.expected {
			t.Errorf("DetectImpact(%q) = %q; want %q", tt.message, got, tt.expected)
		}
	}
}
