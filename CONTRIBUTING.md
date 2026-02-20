# Contributing to tutugit

First of all, thank you for your interest in contributing to **tutugit**! This document outlines the guidelines and best practices for contributing to the project to ensure a smooth and collaborative experience for everyone.

## Development Setup

### Prerequisites
- Go 1.24 or later
- Git

### Initializing the Environment
1. Clone the repository to your local machine:
   ```bash
   git clone https://github.com/carlosedujs/tutugit.git
   cd tutugit
   ```
2. Install the required dependencies:
   ```bash
   go mod download
   ```

## Testing Your Changes

### Automated Tests
Before submitting a Pull Request, please make sure to run the test suite:
```bash
go test ./...
```
Verify that all tests pass. If you're adding new functionality, try to include corresponding tests in the `_test.go` files to maintain our code quality.

We also have a handy coverage script that generates a detailed HTML report and terminal summary:
```bash
./scripts/coverage.sh
```
This script will print a clean summary to your terminal and create a `coverage_report/index.html` file. You can open it in your browser to inspect coverage line by line.

### TUI Testing (Demo Mode)
Testing Text User Interface (TUI) changes inside a real repository can be risky. To make things safer, use our built-in demo mode to test the interface with mock data:
```bash
go run ./cmd/tutugit demo
```
This launches the application in an isolated environment with simulated commits, worktrees, and rebase scenarios, allowing you to test freely without fear of breaking anything.

## Development Workflow

### Coding Standards
- Write idiomatic Go and follow standard naming conventions.
- Always run `go fmt ./...` before committing to ensure consistent formatting.
- Try to avoid adding external dependencies unless they are absolutely necessary.

### Commit Messages
We rely on semantic categorization for our commits. Please prefix your commit messages using the following standard formats:
- `feat:` for introducing new features.
- `fix:` for squashing bugs.
- `refactor:` for code changes that neither fix a bug nor add a feature.
- `docs:` for updates to documentation.
- `chore:` for maintenance tasks, dependency updates, etc.

## Submission Process

1. Fork the repository and create a new, descriptively named branch for your feature or fix.
2. Implement your changes, following the coding standards, and add tests if applicable.
3. Verify that all tests pass and your code is properly formatted.
4. Open a Pull Request with a clear, concise description of the problem you're solving and how your solution works.

## Licensing
By contributing to tutugit, you agree that your contributions will be licensed under the MIT License. We appreciate your help in making this tool better for everyone!
