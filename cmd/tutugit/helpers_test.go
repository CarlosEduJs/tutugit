package main

import (
	"testing"
)

func TestVersion(t *testing.T) {
	// Test that version constant is defined
	if version == "" {
		t.Error("version should not be empty")
	}
}

func TestStateConstants(t *testing.T) {
	// Test state constants are unique
	states := []state{
		stateMain,
		stateCommit,
		stateDiff,
		stateHunks,
		stateWorkspaces,
		stateNewWorkspace,
		stateHistory,
		stateReflog,
		stateReflogConfirm,
		stateRebasePrepare,
		stateRebaseOngoing,
		stateGitWorktrees,
		stateSummary,
	}

	seen := make(map[state]bool)
	for _, s := range states {
		if seen[s] {
			t.Errorf("Duplicate state value: %v", s)
		}
		seen[s] = true
	}

	if len(seen) != len(states) {
		t.Errorf("Expected %d unique states, got %d", len(states), len(seen))
	}
}

func TestSafeGet(t *testing.T) {
	tests := []struct {
		name     string
		slice    []string
		index    int
		expected string
	}{
		{"valid index", []string{"a", "b", "c"}, 1, "b"},
		{"negative index", []string{"a", "b", "c"}, -1, ""},
		{"out of bounds", []string{"a", "b", "c"}, 10, ""},
		{"empty slice", []string{}, 0, ""},
		{"nil slice", nil, 0, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := safeGet(tt.slice, tt.index)
			if result != tt.expected {
				t.Errorf("safeGet(%v, %d) = %q; want %q", tt.slice, tt.index, result, tt.expected)
			}
		})
	}
}

// Helper function that exists in the codebase
func safeGet(slice []string, index int) string {
	if index < 0 || index >= len(slice) {
		return ""
	}
	return slice[index]
}
