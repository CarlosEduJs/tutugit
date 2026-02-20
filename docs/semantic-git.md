# Semantic Git Workflow

**tutugit** strongly encourages a semantic approach to version control. By attaching meaning to your commits, it becomes incredibly easy to automate releases, track progress, and maintain a highly readable project history.

## Automated Tagging

When you type a commit message, tutugit automatically analyzes it to detect a "Semantic Tag" based on standard prefixes. This categorization is incredibly useful for grouping your changes neatly in the final release summary.

### Supported Categories
- **Feature (`feat:`)**: Introducing new functionality for the user.
- **Fix (`fix:`)**: Squashing bugs and fixing issues.
- **Refactor (`refactor:`)**: Cleanups or structural changes that neither fix a bug nor add a feature.
- **Chore (`chore:`)**: Routine maintenance, dependency updates, or changes to auxiliary tools.
- **Docs (`docs:`)**: Adding or updating project documentation.
- **Experiment**: Exploring ideas or working on temporary Work-In-Progress (WIP) code.

## Impact Levels

Taking inspiration from tools like Changesets, tutugit uses a "Change Intent" system. For every commit you make, you quickly define its intended versioning impact:

- **Patch**: Small bug fixes or minor internal tweaks (increments the version patch, e.g., `0.0.x`).
- **Minor**: New features, enhancements, or significant refinements (increments the minor version, e.g., `0.x.0`).
- **Major**: Breaking changes or complete architectural overhauls (increments the major version, e.g., `x.0.0`).

### Managing Impact in the TUI
When you're in the Commit view (after pressing `c`), you can easily cycle through the impact levels using `Alt + i`. The TUI is smart enough to show a "Suggested" impact based on the semantic tag it detected, but you always have the final say and can override it manually.

## Metadata Persistence

All semantic information, including tags and impact levels, is safely stored in `.tutugit/meta.json`. This allows tutugit to analyze your complete history and accurately suggest the next version number during your release process—all without forcing you to strictly adhere to complex or rigid commit message formats like Conventional Commits.
