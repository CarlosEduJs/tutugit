# Configuration Reference

**tutugit** is highly configurable. When you run `tutugit init` in a repository, the tool creates a `.tutugit` folder. This directory acts as the central hub for your project's local state, workspaces, and configuration.

Inside `.tutugit`, you will find `config.yml`. This file allows you to customize metadata about your project.

## The `config.yml` File

By default, the configuration file looks like this:

```yaml
$schema: ./schemas/config.schema.json
project:
    name: "My Awesome Project"
    description: "A technical overview of the project"
```

### Supported Properties

The schema is defined locally in `.tutugit/schemas/config.schema.json`.

| Property | Type | Description |
| --- | --- | --- |
| `project.name` | String | The human-readable name of your project. This is displayed in the tutugit TUI header. |
| `project.description` | String | A short description of what your project does. This may be used when generating release summaries. |

## The `meta.json` File

While `config.yml` is meant for human editing, tutugit maintains its internal state in `.tutugit/meta.json`. 

You usually won't need to manually edit this file, but it's helpful to know what it does. It relies on `.tutugit/schemas/meta.schema.json`.

- **`workspaces`**: An array of Logical Workspaces. Each workspace tracks its ID, Name, Description, and a list of Commit Hashes associated with it.
- **`active_workspace`**: The string ID of the workspace currently selected in the TUI.
- **`tags`**: A mapping of commit hashes to semantic tags (e.g., `feat`, `fix`).
- **`impacts`**: A mapping of commit hashes to version impacts (`patch`, `minor`, `major`).

### Source Control

> [!TIP]
> **Should I commit the `.tutugit` directory?**
> Yes! me highly recommend committing `.tutugit/config.yml` and `.tutugit/meta.json` to your repository. This ensures that your entire team shares the same Logical Workspaces, Semantic Tags, and Release Summaries. The folder was designed specifically to be tracked in Git.
