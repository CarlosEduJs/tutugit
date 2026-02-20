package git

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const (
	// Git log parsing constants
	minLogPartsCount = 7 // Minimum number of fields in a parsed log entry
)

// GitProvider defines the operations needed by tutugit's semantic layer.
type GitProvider interface {
	Run(ctx context.Context, args ...string) (string, error)
	GetCurrentBranch(ctx context.Context) (string, error)
	GetStatus(ctx context.Context) (string, error)
	StageFile(ctx context.Context, path string) error
	UnstageFile(ctx context.Context, path string) error
	Commit(ctx context.Context, message string) error
	GetLastCommitHash(ctx context.Context) (string, error)
	GetRemoteURL(ctx context.Context) (string, error)
	GetDiff(ctx context.Context, path string, staged bool) (string, error)
	ApplyHunk(ctx context.Context, patch string) error
	GetLog(ctx context.Context, n int) ([]Commit, error)
	GetReflog(ctx context.Context, n int) ([]ReflogEntry, error)
	ListWorktrees(ctx context.Context) ([]Worktree, error)
	AddWorktree(ctx context.Context, path, branch string) error
	RemoveWorktree(ctx context.Context, path string, force bool) error
	ResetToHash(ctx context.Context, target string) error
	IsRebasing(ctx context.Context) bool
	RebaseContinue(ctx context.Context) error
	RebaseAbort(ctx context.Context) error
	RebaseSkip(ctx context.Context) error
	GetCommitsInRange(ctx context.Context, base, head string) ([]Commit, error)
	ParseStatus(ctx context.Context) ([]FileStatus, error)
	GetTags(ctx context.Context) ([]string, error)
	ValidateHash(ctx context.Context, hash string) bool
	RunInteractiveRebase(ctx context.Context, base string, steps []RebaseStep) error
}

// Runner -> handles the execution of git commands.
type Runner struct {
	Cwd string
}

// NewRunner -> creates a new git runner in the specified directory.
func NewRunner(cwd string) GitProvider {
	return &Runner{Cwd: cwd}
}

// gitCommand -> creates a git command with the correct directory and context.
func (r *Runner) gitCommand(ctx context.Context, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = r.Cwd
	cmd.Env = append(os.Environ(),
		"GIT_TERMINAL_PROMPT=0",
		"GIT_PAGER=cat",
		"PAGER=cat",
	)
	return cmd
}

// Run -> executes a git command and returns the stdout strings.
func (r *Runner) Run(ctx context.Context, args ...string) (string, error) {
	cmd := r.gitCommand(ctx, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git %s failed: %w\nstderr: %s\nstdout: %s",
			strings.Join(args, " "), err,
			strings.TrimSpace(stderr.String()),
			strings.TrimSpace(stdout.String()))
	}

	return stdout.String(), nil
}

// GetCurrentBranch -> returns the current active branch name.
func (r *Runner) GetCurrentBranch(ctx context.Context) (string, error) {
	branch, err := r.Run(ctx, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", fmt.Errorf("could not get branch: %w", err)
	}
	return strings.TrimSpace(branch), nil
}

// GetStatus -> returns the porcelain status of the repository.
func (r *Runner) GetStatus(ctx context.Context) (string, error) {
	status, err := r.Run(ctx, "status", "--porcelain")
	if err != nil {
		return "", fmt.Errorf("could not get status: %w", err)
	}
	return status, nil
}

// StageFile -> adds a file to the staging area.
func (r *Runner) StageFile(ctx context.Context, path string) error {
	_, err := r.Run(ctx, "add", path)
	if err != nil {
		return fmt.Errorf("could not add file %s: %w", path, err)
	}
	return nil
}

// UnstageFile -> removes a file from the staging area.
func (r *Runner) UnstageFile(ctx context.Context, path string) error {
	// this uses 'reset' to unstage.
	// for new files it works as well.
	_, err := r.Run(ctx, "reset", "HEAD", "--", path)
	if err != nil {
		return fmt.Errorf("could not unstage file %s: %w", path, err)
	}
	return nil
}

// Commit -> creates a new commit with the given message.
func (r *Runner) Commit(ctx context.Context, message string) error {
	_, err := r.Run(ctx, "commit", "-m", message)
	if err != nil {
		return fmt.Errorf("commit failed: %w", err)
	}
	return nil
}

// GetLastCommitHash -> returns the SHA of the HEAD commit.
func (r *Runner) GetLastCommitHash(ctx context.Context) (string, error) {
	hash, err := r.Run(ctx, "rev-parse", "HEAD")
	if err != nil {
		return "", fmt.Errorf("could not get last commit hash: %w", err)
	}
	return strings.TrimSpace(hash), nil
}

// GetRemoteURL -> returns the URL of the "origin" remote, or empty string if not set.
func (r *Runner) GetRemoteURL(ctx context.Context) (string, error) {
	url, err := r.Run(ctx, "remote", "get-url", "origin")
	if err != nil {
		return "", nil // no remote is not an error
	}
	return strings.TrimSpace(url), nil
}

// GetDiff -> returns the diff of a file. If staged is true, it shows staged changes.
func (r *Runner) GetDiff(ctx context.Context, path string, staged bool) (string, error) {
	args := []string{"diff"}
	if staged {
		args = append(args, "--cached")
	}
	args = append(args, "--", path)

	diff, err := r.Run(ctx, args...)
	if err != nil {
		return "", fmt.Errorf("could not get diff for %s: %w", path, err)
	}
	return diff, nil
}

// ApplyHunk -> stages a specific hunk using git apply --cached.
func (r *Runner) ApplyHunk(ctx context.Context, patch string) error {
	cmd := r.gitCommand(ctx, "apply", "--cached", "-")
	cmd.Stdin = strings.NewReader(patch)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("could not apply hunk: %w", err)
	}
	return nil
}

// Commit -> represents a single git commit.
type Commit struct {
	Hash      string
	ShortHash string
	Parents   []string
	Author    string
	Email     string
	Date      string
	Message   string
	Body      string
}

// GetLog -> returns the commit history.
func (r *Runner) GetLog(ctx context.Context, n int) ([]Commit, error) {
	// %x1f = Unit Separator
	// %x1e = Record Separator
	format := "%H%x1f%h%x1f%P%x1f%an%x1f%ae%x1f%cr%x1f%s%x1f%B%x1f%x1e"
	args := []string{"log", fmt.Sprintf("-n%d", n), "--pretty=format:" + format}
	return r.parseLog(ctx, args)
}

func (r *Runner) parseLog(ctx context.Context, args []string) ([]Commit, error) {
	output, err := r.Run(ctx, args...)
	if err != nil {
		return nil, fmt.Errorf("could not get git log: %w", err)
	}

	if output == "" {
		return nil, nil
	}

	records := strings.Split(output, "\x1e")
	var commits []Commit
	for _, record := range records {
		record = strings.TrimSpace(record)
		if record == "" {
			continue
		}
		parts := strings.Split(record, "\x1f")
		if len(parts) < minLogPartsCount {
			continue
		}

		var parents []string
		if parts[2] != "" {
			parents = strings.Split(parts[2], " ")
		}

		c := Commit{
			Hash:      parts[0],
			ShortHash: parts[1],
			Parents:   parents,
			Author:    parts[3],
			Email:     parts[4],
			Date:      parts[5],
			Message:   parts[6],
		}
		if len(parts) > 7 {
			c.Body = parts[7]
		}
		commits = append(commits, c)
	}
	return commits, nil
}

// ReflogEntry represents an entry in the git reflog.
type ReflogEntry struct {
	Hash     string
	Selector string // e.g. HEAD@{0}
	Action   string // e.g. commit, rebase, reset
	Message  string
	Date     string
}

// GetReflog -> returns the git reflog.
func (r *Runner) GetReflog(ctx context.Context, n int) ([]ReflogEntry, error) {
	// using --date=relative and custom format
	format := "%H|%gD|%gs|%gd"
	args := []string{"reflog", fmt.Sprintf("-n%d", n), "--pretty=format:" + format}

	output, err := r.Run(ctx, args...)
	if err != nil {
		return nil, fmt.Errorf("could not get git reflog: %w", err)
	}

	if output == "" {
		return nil, nil
	}

	lines := strings.Split(output, "\n")
	var entries []ReflogEntry
	for _, line := range lines {
		parts := strings.Split(line, "|")
		if len(parts) < 3 {
			continue
		}

		// parts: hash | selector | subject | (sometimes date if format is different)
		// subject often contains "action: message"
		subject := parts[2]
		action := "unknown"
		message := subject

		if strings.Contains(subject, ": ") {
			subParts := strings.SplitN(subject, ": ", 2)
			action = subParts[0]
			message = subParts[1]
		} else if strings.HasPrefix(subject, "commit") {
			action = "commit"
		}

		entries = append(entries, ReflogEntry{
			Hash:     parts[0],
			Selector: parts[1],
			Action:   action,
			Message:  message,
		})
	}
	return entries, nil
}

// Worktree represents a git worktree.
type Worktree struct {
	Path   string
	Branch string
	Hash   string
	IsMain bool
}

// ListWorktrees -> returns the list of registered worktrees.
func (r *Runner) ListWorktrees(ctx context.Context) ([]Worktree, error) {
	output, err := r.Run(ctx, "worktree", "list", "--porcelain")
	if err != nil {
		return nil, fmt.Errorf("could not list worktrees: %w", err)
	}

	if output == "" {
		return nil, nil
	}

	var worktrees []Worktree
	var current Worktree
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		if line == "" {
			if current.Path != "" {
				worktrees = append(worktrees, current)
				current = Worktree{}
			}
			continue
		}

		parts := strings.SplitN(line, " ", 2)
		if len(parts) < 2 {
			continue
		}

		key := parts[0]
		value := parts[1]

		switch key {
		case "worktree":
			current.Path = value
		case "branch":
			current.Branch = strings.TrimPrefix(value, "refs/heads/")
		case "HEAD":
			current.Hash = value
		}
	}

	// final one if no trailing newline
	if current.Path != "" {
		worktrees = append(worktrees, current)
	}
	// identify main worktree (usually the one containing .git as a directory)
	for i := range worktrees {
		if i == 0 {
			worktrees[i].IsMain = true // first in porcelain is main
		}
	}

	return worktrees, nil
}

// AddWorktree -> adds a new worktree.
func (r *Runner) AddWorktree(ctx context.Context, path, branch string) error {
	args := []string{"worktree", "add", path, branch}
	_, err := r.Run(ctx, args...)
	if err != nil {
		return fmt.Errorf("could not add worktree at %s: %w", path, err)
	}
	return nil
}

// RemoveWorktree -> takes a worktree path and removes it.
func (r *Runner) RemoveWorktree(ctx context.Context, path string, force bool) error {
	args := []string{"worktree", "remove", path}
	if force {
		args = append(args, "--force")
	}
	_, err := r.Run(ctx, args...)
	if err != nil {
		return fmt.Errorf("could not remove worktree %s: %w", path, err)
	}
	return nil
}

// ResetToHash -> performs a reset to a specific commit or reflog selector.
func (r *Runner) ResetToHash(ctx context.Context, target string) error {
	_, err := r.Run(ctx, "reset", "--hard", target)
	if err != nil {
		return fmt.Errorf("could not reset to %s: %w", target, err)
	}
	return nil
}

// RebaseAction represents a git rebase command.
type RebaseAction string

const (
	ActionPick   RebaseAction = "pick"
	ActionSquash RebaseAction = "squash"
	ActionFixup  RebaseAction = "fixup"
	ActionEdit   RebaseAction = "edit"
	ActionDrop   RebaseAction = "drop"
	ActionReword RebaseAction = "reword"
)

// RebaseStep -> represents a single step in a rebase todo list.
type RebaseStep struct {
	Action  RebaseAction
	Hash    string
	Message string
}

// IsRebasing -> checks if a rebase is currently in progress.
func (r *Runner) IsRebasing(ctx context.Context) bool {
	// this simple check is to see if the .git/rebase-merge or .git/rebase-apply directory exists.
	// but checking the output of a rebase command is more reliable and direct.
	_, err := r.Run(ctx, "rebase", "--show-current-patch")
	return err == nil
}

// RebaseContinue -> continues an ongoing rebase.
func (r *Runner) RebaseContinue(ctx context.Context) error {
	_, err := r.Run(ctx, "rebase", "--continue")
	return err
}

// RebaseAbort -> aborts an ongoing rebase.
func (r *Runner) RebaseAbort(ctx context.Context) error {
	_, err := r.Run(ctx, "rebase", "--abort")
	return err
}

// RebaseSkip -> skips the current patch in a rebase.
func (r *Runner) RebaseSkip(ctx context.Context) error {
	_, err := r.Run(ctx, "rebase", "--skip")
	return err
}

// GetCommitsInRange returns commits between base and head (excluding base).
func (r *Runner) GetCommitsInRange(ctx context.Context, base, head string) ([]Commit, error) {
	format := "%H%x1f%h%x1f%P%x1f%an%x1f%ae%x1f%cr%x1f%s%x1f%B%x1f%x1e"
	rangeSpec := fmt.Sprintf("%s..%s", base, head)
	if base == "" {
		rangeSpec = head
	}
	args := []string{"log", "--pretty=format:" + format, "--reverse", rangeSpec}
	return r.parseLog(ctx, args)
}

// RunInteractiveRebase -> executes an interactive rebase using the provided steps.
func (r *Runner) RunInteractiveRebase(ctx context.Context, base string, steps []RebaseStep) error {
	var b strings.Builder
	for _, s := range steps {
		b.WriteString(fmt.Sprintf("%s %s %s\n", s.Action, s.Hash, s.Message))
	}

	// create a temp file with the rebase todo content
	tmpFile, err := os.CreateTemp("", "tutugit-rebase-todo-*")
	if err != nil {
		return fmt.Errorf("could not create temp rebase todo: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	if _, err := tmpFile.Write([]byte(b.String())); err != nil {
		return fmt.Errorf("could not write rebase todo: %w", err)
	}

	// Close explicitly to ensure flush before git reads it
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("could not close temp file: %w", err)
	}

	// Use printf to safely escape the file path
	editorCmd := fmt.Sprintf("cp %q", tmpFile.Name())

	cmd := r.gitCommand(ctx, "rebase", "-i", base)
	cmd.Env = append(os.Environ(),
		"GIT_SEQUENCE_EDITOR="+editorCmd,
		"GIT_EDITOR=true",
	)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		// note: Rebase might fail with conflicts, which is "expected" if conflicts occur !!!!
		return fmt.Errorf("rebase failed: %w (stderr: %s)", err, stderr.String())
	}

	return nil
}

// FileStatus -> represents the status of a single file.
type FileStatus struct {
	Path     string
	Staged   bool
	Modified bool
	New      bool
	Deleted  bool
}

// ParseStatus -> parses the git status --short output.
func (r *Runner) ParseStatus(ctx context.Context) ([]FileStatus, error) {
	output, err := r.GetStatus(ctx)
	if err != nil {
		return nil, err
	}

	if output == "" {
		return nil, nil
	}

	lines := strings.Split(output, "\n")
	var files []FileStatus
	for _, line := range lines {
		if len(line) < 4 { // Minimum: "XY P"
			continue
		}

		// git porcelain format: "XY PATH"
		// X: status of the index
		// Y: status of the work tree
		stagedStatus := line[0]
		unstagedStatus := line[1]
		path := strings.TrimSpace(line[2:]) // path starts after "XY", skip the index 2 space and trim others

		f := FileStatus{Path: path}

		// Logic for Staged (X column)
		// If X is not ' ' and not '?', it's staged in some way.
		if stagedStatus != ' ' && stagedStatus != '?' {
			f.Staged = true
		}

		// Logic for Flags
		if stagedStatus == 'M' || unstagedStatus == 'M' {
			f.Modified = true
		}
		if stagedStatus == 'A' || stagedStatus == '?' || unstagedStatus == '?' {
			f.New = true
		}
		if stagedStatus == 'D' || unstagedStatus == 'D' {
			f.Deleted = true
		}

		files = append(files, f)
	}
	return files, nil
}

// GetTags -> returns a list of all tags in the repository.
func (r *Runner) GetTags(ctx context.Context) ([]string, error) {
	output, err := r.Run(ctx, "tag", "-l", "--sort=-v:refname")
	if err != nil {
		return nil, fmt.Errorf("could not list tags: %w", err)
	}
	if output == "" {
		return nil, nil
	}
	return strings.Split(strings.TrimSpace(output), "\n"), nil
}

// ValidateHash -> checks if a commit hash exists and is reachable from any branch.
func (r *Runner) ValidateHash(ctx context.Context, hash string) bool {
	output, err := r.Run(ctx, "branch", "-a", "--contains", hash)
	return err == nil && strings.TrimSpace(output) != ""
}
