# Unreleased: the merge queue rebases the bookmarks of the tasks

- Date: 2026-11-03
- Bookmark: samw/ai/merge-queue-rebase
- Components: hephaestus
- Changelog: `hephaestus merge-queue` now rebases each merged bookmark onto the line, so that `main..<bookmark>` is empty after a human moves `main`. Bookmarks that it leaves out do not change, and the report gives the command that puts the bookmarks back.

## Summary

The merge queue copied commits and left the bookmarks of the tasks on the originals, so merged
bookmarks stayed beside `main`. Now the bookmarks move onto the line at the end of a run.

## Changes

- `internal/hephaestus/mergequeue`: a new step, `Land`, at the end of a run. In jj, it rebases the
  commits of each merged bookmark and drops the trial copies. In git, it moves each branch to its
  copy.
- A new reason to leave a bookmark out: `immutable`. A new report field: `undo`.
- A jj bookmark that moves during a run now stops the run. A git branch that moved, or that a
  worktree has checked out, stays with a warning.
- The run no longer fails to forget its scratch workspace when the rebase makes the workspace of
  the user stale.

## Decisions

- [[0003-rebase-the-bookmarks-of-the-tasks]]

## Issues

- Closed [[merge-queue-rebases-bookmarks]].

## Notes

This changes what the merge queue does to bookmarks: a script that reads a bookmark after a run sees
a new commit. A human who rejects a line must run the `undo` command of the report.
