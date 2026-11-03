# 0003: Rebase the bookmarks of the tasks

- Status: Accepted
- Date: 2026-11-03
- Version: [[unreleased-merge-queue-rebases-bookmarks]]
- Amends: [[0001-merge-with-git-in-a-scratch-worktree]], [[0002-merge-with-jj-in-a-jj-repo]]

## Context

The merge queue put copies of the commits of a task on the line, with `jj duplicate` or
`git cherry-pick`, and did not change the bookmark of the task. After a human moved `main` to the
line, the bookmark stayed beside `main` on the original commits, with work that `main` already had.
`main..<bookmark>` did not become empty, so the check of [[CONTRIBUTING#When work is complete]]
did not pass for merged work. A human asked for a rebase, so that the bookmarks follow `main`.

0001 and 0002 rejected a rebase because it changes the work of other workers. A left-out bookmark
must still not change, and a rebase during the run would change a bookmark before the checks pass.

## Decision

- During the run, the line stays a trial: copies of the commits, and the checks run on them. A
  bookmark that the run leaves out does not change.
- At the end of the run, each merged bookmark moves onto the line:
  - In a jj repo, `jj rebase -s` moves the commits of the bookmark and what is on top of them. The
    version commit goes directly above the commits of its bookmark. The copies are then dropped.
    The change ids stay, and the working copy of a worker on the bookmark follows it.
  - In a git repo, the cherry-picks are the rebase: `git branch -f` points the branch at its copy.
- In a jj repo, the run compares the rebased line with the trial, and stops if the trees differ.
- A jj bookmark that moved during the run stops the run before any bookmark moves. A git branch
  that moved, or that a worktree has checked out, stays with a warning: the line has its work.
- A jj bookmark with immutable commits is left out, with the reason `immutable`.
- The report has the command that puts the bookmarks back, in the field `undo`.

## Alternatives considered

- **`jj rebase` in place of `jj duplicate` during the run.** A failed check must then put a bookmark
  back with new commit ids, and other workers see untested and conflicted commits during the checks.
- **Move each bookmark to its copy with `jj bookmark set`.** It is less code, but it gives each
  commit a new change id, and a working copy on the old commit does not follow.
- **Rebase when a human moves `main`.** This is the best moment, but no command moves `main`: a human
  moves it with jj. A new command for it is a larger change, and is not necessary for this result.

## Consequences

- When a human moves `main` to the line, `main..<bookmark>` is empty for each merged bookmark.
- A human who rejects a line must run the `undo` command before others use the bookmarks: a bookmark
  on a rejected line brings that line with it into the next run. In jj, the command also undoes what
  others did in the repo after the start of the run.
- A jj workspace with a working copy on top of a merged bookmark becomes stale. Run
  `jj workspace update-stale` in it.
- A left-out jj bookmark on top of a merged one follows the rebase, and can get a conflict.
- A bookmark that was pushed needs a force push after the rebase, with the normal lease of
  `jj git push`.
- `jj rebase` is not one transaction. An error part of the way leaves some bookmarks moved, and the
  message gives the undo command.
