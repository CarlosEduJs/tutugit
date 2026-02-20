<div align="center">

```text
   /\_/\
  ( o.o )
   >   <
  tutugit
```

**Because real development is chaotic. Your releases shouldn't be.**

[![Go Release](https://github.com/carlosedujs/tutugit/actions/workflows/release.yml/badge.svg)](https://github.com/carlosedujs/tutugit/actions/workflows/release.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/carlosedujs/tutugit)](https://goreportcard.com/report/github.com/carlosedujs/tutugit)

</div>

**tutugit** is a terminal-based Git power tool designed to solve a problem that traditional Git GUIs ignore: the gap between "I know what I changed" and "I can communicate this cleanly to the world." 

In the real world, developers don't write one perfect feature at a time. A typical day might involve 40 scattered commits touching bug fixes, UI tweaks, and half-finished experiments. **tutugit** accepts this chaos and helps you transform it into pristine, semantic releases—complete with organized workspaces, version impact tracking, and automated changelogs.

---

## 🌪️ The Problem: The Chaotic Reality

You make a commit fixing a typo. Then one refactoring a component. Then you start a new feature, realize another bug needs fixing, and commit that too. Your `git log` becomes a messy, chronological dump of raw thoughts.

When release day comes, figuring out what actually goes into the changelog and whether the version should be a `minor` or `patch` is a guessing game. 

## ✨ The Solution: The tutugit Workflow

With **tutugit**, you don't have to change how you work; you just change how you organize the result.

1. **Work Chaotically**: Keep coding the way you naturally do.
2. **Group Logically**: Use tutugit’s **Logical Workspaces** to group your scattered commits together (e.g., all commits related to "Auth Overhaul" go into one Workspace, even if they happened days apart).
3. **Define Impact**: At commit time, tutugit asks you to define the Semantic Versioning impact (`patch`, `minor`, `major`) and tags it automatically (`feat`, `fix`, `refactor`).
4. **Generate the Release**: When you are ready, hit a single key to export a beautifully formatted Markdown release summary. It looks as if your history was always clean and planned.

Perfect for solo developers and small teams who want professional, standard-compliant deliverables without the overhead of enterprise release boards.

---

## 🛠️ Key Features (Built for the Workflow)

While grouping your chaotic commits into logical releases is the goal, tutugit comes packed with power-user tools to help you manipulate the Git state seamlessly:

- **Logical Workspaces**: The core feature. Group staged/unstaged changes and past commits into named buckets inside your current branch.
- **Surgical Staging**: Easily stage specific files or dive deep into **Hunk Staging** to pick exactly which lines belong to which workspace.
- **Auto-Semantic Commits**: Never memorize Conventional Commits again. Write your message, define your impact, and tutugit handles the categorization.
- **Time Machine (Reflog)**: Made a mistake while organizing? Visually navigate the Git Reflog to recover lost states or undo resets safely.
- **Visual Rebase Planner**: Clean up your messy history interactively in a dedicated view (Reorder, Pick, Squash, Drop, Reword).
- **Worktree Explorer**: Manage multiple linked working trees effortlessly to work on hotfixes without losing your current uncommitted state.

---

## Quick Start

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

Initialize tutugit in the root of any existing Git repository:

```bash
tutugit init
```

This creates a lightweight `.tutugit` directory containing your local configuration and workspace metadata. *(Note: Your project must already be a git repository for tutugit to work).*

---

## 📖 Documentation

Dive deeper into how tutugit transforms your workflow:

- [Getting Started](docs/getting-started.md): Installation and basic setup.
- [Logical Workspaces](docs/workspaces.md): A deep dive into grouping your commits.
- [Semantic Git Workflow](docs/semantic-git.md): Understanding tags and version impacts.
- [Release Summaries](docs/summary-export.md): Exporting your work into clean changelogs.
- [Advanced Operations](docs/advanced-ops.md): Mastering the Reflog, Rebase Planner, and Worktrees.
- [Configuration & Keybindings](docs/configuration.md): Tailoring tutugit to your needs.


## 🤝 Development & Contributing

We welcome contributions! Please check out our [Contributing Guide](CONTRIBUTING.md) for details on setting up your environment and submitting pull requests.

If you plan to write code, make sure to run our testing tools locally:

```bash
# Run tests and generate a gorgeous HTML coverage report
./scripts/coverage.sh
```

## 📝 License

MIT License.
