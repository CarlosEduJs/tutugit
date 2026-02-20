# Command Line Interface (CLI) Reference

While **tutugit** is primarily a Text User Interface (TUI) application, it provides a few essential command-line arguments to help you manage your project setup and local environment.

## Available Commands

### `tutugit` 
*(No arguments)*

Launches the main Text User Interface (TUI). This is the default mode you will use for your daily Git operations.
- **Requirement**: Must be run inside a Git repository.

---

### `tutugit init`

Initializes **tutugit** in the current directory. 

- Creates the `.tutugit` hidden directory.
- Bootstraps `meta.json` to manage workspaces and semantic data.
- Generates a default `config.yml`.
- Copies JSON schemas into `.tutugit/schemas/` to provide autocompletion and validation in your IDE.

> [!NOTE]
> If you run `tutugit init` in a directory that has already been initialized, the command will safely abort and notify you that it is already set up.

---

### `tutugit demo`

Launches the application in an isolated, simulated environment. 

This mode is heavily used for development, testing UI changes, or just learning how to use the tool without risking changes to your actual Git repository. It loads a mock Git environment with simulated commits, files, and workspaces.

---

### `tutugit --version`

Prints the currently installed version of the binary.

```bash
$ tutugit --version
tutugit version 1.0.0
```
