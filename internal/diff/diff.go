package diff

import (
	"strings"
)

// Hunk -> represents a single change block in a diff.
type Hunk struct {
	Header  string
	Content string
}

// FileDiff -> represents all changes in a single file.
type FileDiff struct {
	Path  string
	Hunks []Hunk
}

// ParseDiff -> splits a raw diff string into FileDiffs and Hunks.
func ParseDiff(raw string) []FileDiff {
	var files []FileDiff
	var currentFile *FileDiff
	var currentHunk *Hunk

	lines := strings.Split(raw, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "diff --git") {
			if currentFile != nil {
				if currentHunk != nil {
					currentFile.Hunks = append(currentFile.Hunks, *currentHunk)
					currentHunk = nil
				}
				files = append(files, *currentFile)
			}
			parts := strings.Split(line, " ")
			path := ""
			if len(parts) > 2 {
				path = strings.TrimPrefix(parts[2], "a/")
			}
			currentFile = &FileDiff{Path: path}
			continue
		}

		if strings.HasPrefix(line, "@@") {
			if currentHunk != nil && currentFile != nil {
				currentFile.Hunks = append(currentFile.Hunks, *currentHunk)
			}
			currentHunk = &Hunk{Header: line, Content: line + "\n"}
			continue
		}

		if currentHunk != nil {
			currentHunk.Content += line + "\n"
		}
	}

	if currentFile != nil {
		if currentHunk != nil {
			currentFile.Hunks = append(currentFile.Hunks, *currentHunk)
		}
		files = append(files, *currentFile)
	}

	return files
}

// ToPatch returns a string that can be used with `git apply`.
func (h *Hunk) ToPatch(filePath string) string {
	var b strings.Builder
	b.WriteString("diff --git a/" + filePath + " b/" + filePath + "\n")
	b.WriteString("--- a/" + filePath + "\n")
	b.WriteString("+++ b/" + filePath + "\n")
	b.WriteString(h.Content)
	return b.String()
}
