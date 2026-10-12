# 0008: Each task has a workspace of its own

- Status: Accepted
- Date: 2026-10-12
- Version: [[unreleased-task-workspaces]]

## Context

[[AGENTS]] asks each agent to work in a jj workspace of its own, on a bookmark of its own, and to
remove the workspace when the work is complete. A worker of hephaestus ran each task in its own
directory, or the one that the message named. So workers that run at the same time could edit the
same files, and a tool on a notice wrote into the workspace of the implementer. The human asked for
a workspace for each task.

## Decision

- **A workspace for each task.** A role file says `workspace: task` (the default of an AI role) or
  `none` (the default of a tool role). For a task message of a `task` role, the worker makes a jj
  workspace (in a jj repo) or a git worktree on a new branch (in a git repo), named
  `<role>-<issue>`, in `workspaces/<repo>-<role>-<issue>` beside the main working copy
  (`--workspaces`). It starts on the bookmark that the task names, or on the base (`--onto`). Its
  bookmark is `<prefix><role>-<issue>`, with the prefix of `--bookmark-prefix`,
  `HEPHAESTUS_BOOKMARK_PREFIX`, or `ai/`.
- **The issue goes with the work.** The worker claims the issue in the workspace, with the bookmark
  and the workspace name, and the model works and records the issue there, as [[AGENTS]] asks.
- **The end of a task.** When the task is done, the worker announces the close of the issue
  (`bookmark-ready` and `issue-closed`, with the ids that `hephaestus watch` uses, so the queue keeps
  one of each), and removes the workspace; its commits and bookmark stay. A workspace with changes
  that no commit keeps, and the workspace of a task that failed or stopped, stay for a human, and
  the Notes of the issue say where.
- **A notice about a bookmark runs in a scratch copy of it.** The changes that a role makes there
  (a formatter) are committed on the bookmark `<prefix>fix-<role>-<issue>`, on top of the bookmark,
  for its author to squash. The name is flat, because a git branch cannot be a directory of another.
- **jj names first.** In a jj repo, the worker asks jj for the commit of a bookmark, so a bookmark
  that a workspace made a moment ago counts before jj exports it to git.

## Alternatives considered

- **One workspace for each worker.** A worker does many tasks; their work would mix on one bookmark.
- **Keep the issue in the main working copy.** The issue would change in one working copy and the
  work in another, and a merge would conflict.
- **Remove every workspace at the end.** Uncommitted work, and the evidence of a failure, would go.

## Consequences

- Workers that run at the same time never edit the same files.
- `hephaestus watch` sees a close in a task workspace only after the merge; the worker announces it
  first, so the reviewers start at once.
- A human finds the workspaces of failed tasks in the Notes of their issues, and removes them.
