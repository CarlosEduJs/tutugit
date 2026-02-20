# Keybindings Reference

**tutugit** uses a Vim-inspired keyboard design, enabling you to keep your hands on the home row for maximum efficiency. 

Most views share basic navigation, while advanced features have their own specialized shortcuts.

## Global Navigation
These shortcuts generally work across multiple views (Main, History, Reflog, etc.).

| Key | Action |
| --- | --- |
| `q` or `Ctrl+c` | Quit tutugit (from Main screen) |
| `Esc` or `q` | Go back / Return to the Main screen |
| `j` or `Down` | Move cursor down |
| `k` or `Up` | Move cursor up |
| `r` | Refresh the current view |

---

## 🖥 Main Screen & Staging
Manage your working tree and staging area.

| Key | Action |
| --- | --- |
| `Space` or `s` | Toggle stage/unstage for the selected file |
| `p` | Open Hunk Staging for the selected file |
| `Enter` or `e` | Expand/Collapse inline diff for the selected file |
| `d` | Open full Diff View for the selected file |
| `c` | Open the Commit form |
| `w` | Open the Workspaces panel |
| `h` | View Commit History |
| `t` | Open Worktrees Explorer |
| `g` | Open the Time Machine (Reflog) |
| `L` | View Release Summary (Changelog Preview) |

---

## ✂️ Hunk Staging (`p`)
Stage specific parts of a file.

| Key | Action |
| --- | --- |
| `j` / `k` | Navigate through diff hunks |
| `Space` or `s` | Apply (stage) the selected hunk |
| `Esc` or `q` | Return to Main screen |

---

## 📝 Commit View (`c`)
Write your commit message and define semantic impact.

| Key | Action |
| --- | --- |
| `Enter` | Confirm and commit (if message is not empty) |
| `Alt+i` | Cycle Version Impact (`patch` ➔ `minor` ➔ `major` ➔ `auto`) |
| `Esc` | Cancel commit and return |

---

## 📁 Workspaces (`w`)
Group related changes together.

| Key | Action |
| --- | --- |
| `j` / `k` | Navigate workspaces |
| `a` | Activate the selected workspace |
| `n` | Create a New Workspace (asks for Name and Description; use `Tab` to switch fields) |
| `Esc` or `w` | Return to Main screen |

---

## 📜 History View (`h`)
Browse your repository's commit timeline.

| Key | Action |
| --- | --- |
| `j` / `k` | Navigate commits |
| `Enter` | Expand/Collapse commit details |
| `R` | Start Interactive Rebase Planner, using the selected commit as the base |
| `L` | View Release Summary |
| `Esc`, `q`, or `h` | Return to Main screen |

---

## 🔄 Rebase Planner (`R` from History)
Interactively rewrite your local history.

| Key | Action |
| --- | --- |
| `j` / `k` | Navigate rebase steps / commits |
| `J` / `K` (Shift) | **Move** the selected commit down (`J`) or up (`K`) in the timeline |
| `p` | Mark as **Pick** (Keep) |
| `r` | Mark as **Reword** (Edit message) |
| `e` | Mark as **Edit** (Stop to amend) |
| `s` | Mark as **Squash** (Meld into previous commit) |
| `f` | Mark as **Fixup** (Meld, but discard message) |
| `d` | Mark as **Drop** (Remove commit) |
| `Enter` | Execute the rebase sequence |
| `Esc` or `q` | Cancel planner and return to history |

### During an Ongoing Rebase (Conflict / Edit)
If an interactive rebase stops for conflicts or edits, **tutugit** enters an `Ongoing` state.
| Key | Action |
| --- | --- |
| `c` | Continue rebase (`git rebase --continue`) |
| `s` | Skip current patch (`git rebase --skip`) |
| `a` | Abort rebase (`git rebase --abort`) |

---

## ⏳ Time Machine / Reflog (`g`)
Recover lost commits and visually browse the Reflog.

| Key | Action |
| --- | --- |
| `j` / `k` | Navigate reflog entries |
| `Enter` | Select entry and prepare to Reset |
| `y` or `Enter` | (In confirmation prompt) Verify and Execute the Hard Reset |
| `n` or `Esc` | (In confirmation prompt) Cancel reset |
| `Esc`, `q`, or `g` | Return to Main screen |

---

## Release Summary (`L`)
Preview and export your generated changelog.

| Key | Action |
| --- | --- |
| `E` | Export summary to `.tutugit/release.md` |
| `Esc`, `q`, or `L` | Return to previous screen |

---

## 🌳 Worktrees Explorer (`t`)
Manage multiple branch checkouts.

| Key | Action |
| --- | --- |
| `j` / `k` | Scroll through worktrees |
| `Esc`, `q`, or `t` | Return to Main screen |
