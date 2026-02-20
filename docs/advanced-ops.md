# Advanced Operations

Beyond standard staging and committing, **tutugit** comes packed with visual interfaces for advanced, high-level Git operations. These are the kinds of tasks that are typically error-prone or hard to manage using the standard command line.

## Time Machine (Reflog)

The Git Reflog is your safety net—it tracks absolutely every change to the HEAD of your repository, including commits, checkouts, resets, and rebases. tutugit provides a fully visual navigator for this powerful feature.

1. Press `g` from the main screen to enter the **Time Machine**.
2. Scroll through the chronological list of history entries.
3. Select any entry and press `Enter` to see its details or prepare to safely reset your state. *(Note: destructive operations like resetting will ask for confirmation so you can use them safely).*

## Visual Rebase Planner

Interactive rebasing allows you to clean up, squash, or organize your local history before you share it with others. tutugit dramatically simplifies this process with a dedicated, visual planner.

1. Press `h` to enter the **Visual History** view.
2. Find and select the commit you want to use as the base for your rebase.
3. Press `R` to launch the **Interactive Rebase Planner**.
4. Inside the planner, you have full control:
    - Use `j` and `k` to navigate through your commits.
    - Press `p` (pick), `s` (squash), `d` (drop), or `f` (fixup) to define the action for each specific commit.
    - Use uppercase `J` and `K` to physically move commits up or down in the timeline.
    - Press `Enter` to seamlessly execute the rebase using Git.

## Worktree Explorer

Git Worktrees are an incredible feature that allows you to have multiple branches checked out simultaneously in different directories, without having to clone the repository multiple times.

1. Press `t` to open the **Worktree Explorer**.
2. View a clean list of all your active worktrees, seeing their current branch and exactly where they are located on your disk.
3. Easily identify, manage, and navigate to where your parallel tasks are situated.
