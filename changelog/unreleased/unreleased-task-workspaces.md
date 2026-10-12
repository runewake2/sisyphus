# Unreleased: A workspace of its own for each task

- Date: 2026-10-12
- Bookmark: samw/ai/ux-fixes
- Components: hephaestus
- Changelog: Each task of an AI role runs in a jj workspace or git worktree of its own, with its own bookmark, which hephaestus makes and removes; a notice about a bookmark runs in a scratch copy of it.

## Summary

Workers that run at the same time no longer edit the same files, and a tool on a notice no longer
writes into the workspace of the implementer ([[0008-each-task-has-a-workspace-of-its-own]]).

## Changes

- `workspace: task|none` in a role file; `task` is the default of an AI role.
- For a task, `hephaestus agent run` makes `workspaces/<repo>-<role>-<issue>` (a jj workspace, or a
  git worktree), on the bookmark of the task or on `--onto`, with the bookmark
  `<prefix><role>-<issue>` (`--bookmark-prefix`, `HEPHAESTUS_BOOKMARK_PREFIX`, or `ai/`), and claims
  the issue there. The prompt of the model names the workspace and the bookmark.
- When the task is done, the worker announces the close (`bookmark-ready` and `issue-closed`, with
  the ids of `hephaestus watch`), and removes the workspace. A failed or stopped task, or a workspace
  with changes that no commit keeps, keeps it for a human, with a Note.
- A notice about a bookmark runs in a scratch copy of it. Changes that it makes go to the bookmark
  `<prefix>fix-<role>-<issue>`, and a Note tells the author to squash them.
- The implementer works in the workspace that hephaestus made for it.
- The scratch worktrees of the referendum and of notices share one package.

## Decisions

- [[0008-each-task-has-a-workspace-of-its-own]]

## Issues

- Closed [[task-workspaces]].
