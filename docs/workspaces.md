# Logical Workspaces

One of the most powerful features of **tutugit** is the ability to group your commits into "Logical Workspaces". This section explains how workspaces differ from traditional Git branches and how you can use them to keep your workflow clean.

## The Concept

In standard Git, if you are working on a code refactor and a bug fix at the same time on the exact same branch, your staging area quickly becomes a mess. You are usually forced to either juggle multiple branches or very carefully stage files to separate the distinct changes into different commits.

**tutugit's Logical Workspaces** solve this by providing a metadata layer on top of Git. You can define a workspace (for example, "Architecture Cleanup") and associate specific commits directly with it. When it's time to generate your changelog, tutugit automatically groups these related commits together—regardless of when they were actually made in the timeline.

## Using Workspaces in the TUI

### Creating and Managing
1. Press `w` from the main interface to enter the Workspace view.
2. Press `n` to create a new workspace. You can give it a clean name and an optional description.
3. Press `a` to activate a selected workspace and make it your current context.

### Committing to a Workspace
Whenever you commit your changes (by pressing `c`), the commit is automatically linked to your currently active workspace. This link is safely stored in `.tutugit/meta.json` and does not mutate or affect the actual Git commit object. Because of this, tutugit maintains 100% compatibility with your standard Git CLI and other tools.

## Key Benefits

- **Cohesion**: Keep all related changes perfectly grouped together in your generated release notes.
- **Organization**: Focus purely on one logical task at a time, entirely eliminating the need to constantly switch branches.
- **Hygiene**: The "Multi-workspace Hygiene" check (available right in the TUI) automatically helps you identify if a workspace contains "stale" commits that have been deleted, reset, or altered in the underlying Git history.
