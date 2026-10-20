# Unreleased: one command runner and one VCS bookmark sync

- Date: 2026-10-20
- Bookmark: samw/ai/repo-plumbing
- Components: hephaestus
- Changelog: repo.Run and repo.Lines are the one command runner and line splitter of hephaestus (six copies gone, with the ToSlash fix one copy missed), and repo.ExportBookmarks/ImportBookmarks are the one place that syncs jj bookmarks with git (three copies gone, one of which dropped the error).

## Summary

The last two moves of the abstraction engineer's trial. No error text and no report output changed:
the flags that a runner prepends for itself (an identity, a --git-dir) run but stay out of the
error, so a person still reads the command that was asked.

## Changes

- `repo.Run(dir, name, hidden, args...)` returns trimmed stdout and the one error format;
  `repo.Lines` is the one splitter, with the path-separator fix the merge-queue copy missed.
  scratch, taskspace, the merge-queue lines, and revParse build on them.
- `repo.ExportBookmarks` and `repo.ImportBookmarks` guard the VCS inside (a no-op outside a jj
  repo): toolagent, the referendum, the trial, and the merge queue have no `jj git export` or
  `import` of their own. The export stays strict and the import best-effort, as every caller treated
  them; the merge queue still ignores the export error after a merge, with the reason in a comment.

## Issues

- Closed [[share-jj-export-import]] and [[export-repo-command-plumbing]], the last two proposals of
  the abstraction engineer's trial run.
