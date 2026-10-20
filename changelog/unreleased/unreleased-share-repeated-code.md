# Unreleased: share the code that the abstraction engineer found repeated

- Date: 2026-10-20
- Bookmark: samw/ai/share-repeated-code
- Components: hephaestus
- Changelog: The merge queue takes its scratch worktree from package scratch (under the lock that ends the prune race with referendum reviewers), one `scratch.Identity` is the git identity fallback (fixing a latent flag-order defect that broke fallback commits on machines with no configured identity), and one `shell.Command` stops every command with its whole process tree, also the checks of the merge queue.

## Summary

The abstraction engineer reviewed internal/hephaestus and cmd/hephaestus, and the engineer ordered
the moves. The five smallest landed; the command-plumbing and jj-export moves stay as issues with
their dependencies recorded.

## Changes

- `gitLine` of the merge queue holds a `scratch.Worktree`: the add and remove of its worktree run
  under the per-git-directory lock of package scratch, so a merge-queue run and a referendum
  reviewer on one git directory no longer race in `git worktree prune`.
- `scratch.Identity(dir, name)` is the one git identity fallback, dir-scoped and tested. Both old
  copies sorted the flat flag list, which put the two `-c` flags first — on a machine with no
  configured git (or jj) identity, every fallback commit would have failed with a bogus config
  parameter. The shared function and `jjIdentity` keep the pairs in a fixed order, with tests that
  pin it.
- `internal/hephaestus/shell` holds the platform shell and its process-group stop. The checks of the
  merge queue now run through it, so a stopped run stops the children of the checks too.
- `mq.Ago` is the one "how long ago" format (status and doctor); `airole.Indent` is the one prompt
  indent (airole and the referendum prompts); `toolagent.contains` gave way to `slices.Contains`.

## Issues

- Closed [[gitline-reuses-scratch]], [[one-vcs-identity-fallback]], [[share-platform-shell]],
  [[share-ago-format]], [[export-airole-indent]], and [[use-slices-contains]].
- Open, with dependencies recorded by the engineer: [[export-repo-command-plumbing]] and
  [[share-jj-export-import]].
