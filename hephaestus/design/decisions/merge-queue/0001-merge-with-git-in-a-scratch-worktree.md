# 0001: Merge with git in a scratch worktree

- Status: Accepted
- Date: 2026-10-12
- Version: [[unreleased-merge-queue-agent]]

## Context

The merge queue makes one linear history from a set of tasks, gives versions as the work lands,
never moves `main`, and makes no model call ([[agent-delegation#Merge queue]], Q3, Q6, and Q10). It
must not change the working copy of anyone, or the bookmarks of the tasks ([[merge-queue-agent]]).

## Decision

- `hephaestus merge-queue <task>...` merges in a scratch git worktree of the git directory of the
  repo, detached at the base. In a jj repo, it uses the git store of jj, after `jj git export`, so
  that the git branches are the jj bookmarks, and runs `jj git import` after, so that jj sees the
  result as a bookmark.
- It merges the bookmarks in the order of `hephaestus merge-order`. For each bookmark, it
  cherry-picks the commits that the line does not have yet (not in the base, and not in a bookmark
  that it merged already), so that a stack merges once. A bookmark with a merge commit is left out:
  the history must be linear.
- After each bookmark, it renumbers the decision records that collide with the line before it, and
  gives the unreleased entries that the bookmark added their versions, in the order of the commits
  that added them. Both go into one commit, "merge-queue: give versions to <bookmark>".
- Then it runs `--checks` in the worktree. A bookmark that conflicts, fails the checks, or must come
  after a bookmark that was left out is left out: the line goes back to its head before the
  bookmark, the tasks get a note, and a conflict a `conflict` message to `--conflicts-to` (the
  `engineer` role).
- The result is a new bookmark, `merge-queue/<date>` (`-2`, `-3`, ... when it exists), or none when
  nothing merged, and a report. It never moves the base, and never pushes.
- With `--from-message`, it reads the set of tasks from the message of a tool agent.

## Alternatives considered

- **`jj duplicate` and `jj rebase`.** jj records a conflict in the commit instead of a failure, and
  its output must be parsed to find the new commits. git cherry-pick stops at a conflict, which is
  what the merge queue needs, and works for git repos too.
- **Rebase the bookmarks of the tasks.** It changes the work of other workers, and fails when one
  of them has a working copy on it.
- **Versions in the commits of the bookmark.** Each commit would have to be changed; one commit for
  each bookmark keeps the commits of the workers as they were.

## Consequences

- A human checks `merge-queue/<date>` and moves `main` to it; the bookmarks of the tasks can then
  go.
- A failed check or a conflict costs one bookmark, not the run.
- The checks run once for each bookmark, so a run with many bookmarks takes as long as their checks.
