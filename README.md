<div align="center">

# tutugit

</div>

**tutugit** is a <span style="color: cyan;">terminal-based Git power tool</span> built to bring order to your development workflow. It combines surgical staging, logical commit grouping, and automated semantic releases—all within a clean, unified Text User Interface (TUI).

## <span style="color: cyan;">Core Philosophy</span>

Traditional Git workflows often result in messy, fragmented commit histories or complicated branch management. **tutugit** solves this by introducing **Logical Workspaces**. This concept allows you to group related changes together without ever leaving your current branch. The result? A perfectly clean, semantically tagged history that is ready for automated changelog generation.

## <span style="color: cyan;">Key Features</span>

### Logical Workspaces
Group both staged and unstaged changes into named workspaces (for example, "UI Refactor" or "Bug Fix #123"). This flexibility lets you juggle multiple features simultaneously and commit them as cohesive, independent units.

### Semantic Staging
- **Interactive Hunk Staging**: Stage specific lines or hunks of a file with surgical precision.
- **Automated Tagging**: Commits are automatically categorized (e.g., `feat`, `fix`, `refactor`) based on your commit message prefix.
- **Impact Tracking**: Define the versioning impact (`patch`, `minor`, `major`) precisely when you make the commit.

### Advanced Git Operations
- **Time Machine (Reflog)**: Visually navigate the Git Reflog to recover lost states or undo mistakes.
- **Visual Rebase Planner**: Plan interactive rebases in a dedicated view by easily reordering, picking, squashing, or dropping commits.
- **Worktree Explorer**: Effortlessly manage multiple linked working trees so you can work on several branches in parallel.

### Automated Release Summaries
Automatically generate professional Markdown release summaries based on your workspace metadata and semantic history. Export your progress directly to `.tutugit/release.md` and share your updates with ease.

## <span style="color: cyan;">Quick Start</span>

### Prerequisites
- Go 1.24+ (if building from source)
- Git installed on your system

### Installation

```bash
# Clone the repository
git clone https://github.com/carlosedujs/tutugit.git
cd tutugit

# Build the binary
go build ./cmd/tutugit

# Move it to your path
mv tutugit /usr/local/bin/
```

### Initialization

Initialize tutugit in the root of your project:

```bash
tutugit init
```

This creates a `.tutugit` directory containing your local configuration and workspace metadata.

> <span style="color: red;">Your repo need git initialized to use tutugit</span>

## <span style="color: cyan;">Documentation</span>

For more detailed guides, check out our documentation suite:

- [Getting Started](docs/getting-started.md): Installation and basic setup.
- [Configuration](docs/configuration.md): How to configure tutugit.
- [Keybindings](docs/keybindings.md): A comprehensive guide to all keyboard shortcuts.
- [CLI Reference](docs/cli-reference.md): A comprehensive guide to all command-line arguments.
- [Logical Workspaces](docs/workspaces.md): A deep dive into grouping your commits.
- [Semantic Git Workflow](docs/semantic-git.md): Understanding tags and version impacts.
- [Advanced Operations](docs/advanced-ops.md): Mastering the Reflog, Rebase Planner, and Worktrees.
- [Release Summaries](docs/summary-export.md): Exporting your work into clean changelogs.


## <span style="color: cyan;">Development & Contributing</span>

Please check out our [Contributing Guide](CONTRIBUTING.md) for details on setting up your environment, following our coding standards, and submitting pull requests.

If you plan to write code, make sure to run our testing tools:

```bash
# Run tests and generate a gorgeous HTML coverage report
./scripts/coverage.sh
```

## <span style="color: cyan;">License</span>

MIT License.
