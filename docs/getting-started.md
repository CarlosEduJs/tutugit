# Getting Started with tutugit

Welcome to **tutugit**! This guide will walk you through the initial setup and basic usage so you can start organizing your workflow right away.

## Installation

### From Pre-built Binaries
The easiest way to install tutugit is by downloading the latest binary for your operating system from the [Releases](https://github.com/carlosedujs/tutugit/releases) page. Just extract the archive and move the `tutugit` binary to a directory in your system's `PATH`.

### From Source
If you prefer to build it yourself, make sure you have Go 1.24 or later installed.

```bash
git clone https://github.com/carlosedujs/tutugit.git
cd tutugit
go build ./cmd/tutugit
```

## Project Initialization

Before you can use tutugit's advanced features in a repository, you need to initialize it:

```bash
tutugit init
```

This simple command sets up the following in your project:
1. Creates a `.tutugit` hidden directory in your current working directory.
2. Generates `meta.json` to keep track of your workspaces and semantic metadata.
3. Generates `config.yml` for customizable project-level settings.
4. Copies JSON schemas into `.tutugit/schemas/` to enable local configuration validation.

## Configuration

You can fully customize your project's metadata by editing the `.tutugit/config.yml` file:

```yaml
$schema: ./schemas/config.schema.json
project:
    name: "My Awesome Project"
    description: "A technical overview of the project"
```

This information is displayed in the TUI header and is also used to enrich your generated release reports.

## Basic Workflow

Ready to start? Here's how you typically use tutugit:

1. **Launch**: Simply run `tutugit` in your terminal.
2. **Stage**: Navigate through your files using `j` and `k`, and toggle staging for a file or hunk by pressing `Space`.
3. **Group**: Want to keep things organized? Press `w` to create or switch between Logical Workspaces.
4. **Commit**: Press `c` to write your commit message. Remember to use prefixes like `feat:` or `fix:` for automatic semantic tagging.
5. **Release**: When you're ready to share your work, press `L` to open the Summary view, review your progress, and export a clean Markdown report with `E`.
