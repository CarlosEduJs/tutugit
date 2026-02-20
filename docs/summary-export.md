# Release Summaries and Export

**tutugit** makes communicating your updates effortless by generating beautifully structured Markdown reports based precisely on the work you've done.

## The Summary View

Press `L` from the main screen or the history view to enter the built-in **Release Summary** view.

When you do this, tutugit automatically analyzes:
1. The commits within your currently selected range.
2. The associated Semantic Tags (e.g., Feature, Fix, Refactor).
3. The Logical Workspaces the commits belong to.
4. Any co-authors or specific metadata attached to the commits.

The view gives you a beautiful, real-time preview of exactly what your changelog is going to look like.

## Markdown Export

When you are ready to export the summary:
1. While in the Summary view, simply press `E`.
2. tutugit will immediately generate a pristine file at `.tutugit/release.md`.

### Export Structure
The generated Markdown is designed to follow a clean, professional layout, and is strictly categorized by:
- **Workspace Name**: Groups all related commits under their respective, human-readable workspace header.
- **Semantic Type**: Creates clean sub-groups within those workspaces sorted by the type of change (Features, Fixes, Docs, etc.).
- **Impact Warning**: If any minor or major version impacts are detected, they are prominently highlighted so your users know what to expect.

## Integration with CI/CD

Because tutugit safely stores everything inside `.tutugit`, you can easily integrate it into your automated pipelines. You can use the generated `release.md` file as the exact body for your GitHub Releases, as an automated email payload, or anywhere else in your CI/CD process.
