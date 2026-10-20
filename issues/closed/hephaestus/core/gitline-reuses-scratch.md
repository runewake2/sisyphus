---
title: "The git line of the merge queue reuses package scratch for its worktree"
state: closed          # open | in-progress | closed. Must match the directory of the file.
resolution: completed  # completed | abandoned. Set this only when state is closed.
priority: high         # critical | high | medium | low
effort: small          # small | medium | large | x-large
tags: [hephaestus]     # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-20
closed: 2026-10-20     # YYYY-MM-DD. Set this only when state is closed.
owner:                 # The person or agent working on the issue. Cleared when the issue returns to open.
approver:              # The person or agent who accepts the issue when it closes.
bookmark:              # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: []         # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
metadata: {}           # Optional. Arbitrary key-value notes, for example an AI agent session id to resume work with context: {session-id: "abc123"}.
deferred-from:         # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent:                # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
depends-on: []           # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# The git line of the merge queue reuses package scratch for its worktree

## Summary

gitLine in internal/hephaestus/mergequeue/line.go adds and removes a scratch git worktree itself, and repeats what package scratch does. scratch got the fix for the race of two worktrees (a lock per git directory, because one git worktree prune can remove the half-made worktree of another); gitLine did not. A merge-queue run and a referendum reviewer on the same git directory can still race.

The copies:
- internal/hephaestus/mergequeue/line.go:217-233 (newGitLine: MkdirTemp, os.Remove, worktree add --detach --quiet) repeats internal/hephaestus/scratch/scratch.go:40-54 (scratch.New), without the lock.
- internal/hephaestus/mergequeue/line.go:332-336 (Close: worktree remove --force, RemoveAll, worktree prune) repeats scratch.go:57-65 (Worktree.Remove), without the lock, and prunes always instead of only on failure.

The abstraction exists: package scratch. The move: gitLine holds a *scratch.Worktree; newGitLine calls scratch.New(gitDir, baseID, "hephaestus-merge-queue-"); Close calls Worktree.Remove plus os.RemoveAll. gitLine keeps its own git runner (git, gitDirCmd), because its identity fallback differs from scratch.Identity (see [[one-vcs-identity-fallback]]); only the add and remove of the worktree move.

What it removes: ~25 lines, and the second copy of a lifecycle that must stay one (the lock is the fix one copy got and the other did not). What it costs: mergequeue already imports nothing from scratch; the new import is one package with no further dependencies. The prune-on-every-Close behavior becomes prune-on-failed-remove; under the lock that is safe. Order: this move first; it does not depend on the identity issue.

## Context

<Give the background that an agent needs to start the work without other help.
Link to design docs and other issues with wikilinks, for example `[[widget-scheduler]]`,
`[[widget-scheduler#Status API]]`, or `[[<issue-name>]]`. Do not use paths.>

## Acceptance criteria

- [ ] newGitLine and gitLine.Close use scratch.New and scratch.Worktree.Remove, under the per-git-directory lock.
- [ ] mergequeue has no worktree add or worktree remove of its own for the git line.
- [ ] The mergequeue tests pass unchanged, and a test or the scratch package doc shows the lock covers the merge queue too.
- [ ] No output of hephaestus merge changes.

## Out of scope

<List related work that this issue does not include. Create a separate issue for each item that someone must do later.>

## Notes

2026-10-20, from the operations auditor: sisyphus lint reports that the link [[one-vcs-identity-fallback]] of the Summary is missing: no issue of that name exists. For triage: make the link point at the issue that tracks the identity fallback, or remove it. This is the only lint problem of the tracker.

2026-10-20, from the engineer: the link [[one-vcs-identity-fallback]] resolves now. The abstraction engineer filed that issue after the lint run of the operations auditor, so the two reviews raced. sisyphus lint is clean today (187 issues, no problems). The fixer must not change the link; disregard the note above. Also: the fix of [[hephaestus/merge-queue/jj-error-leaves-visible-commits]] changed line.go and mergequeue.go on 2026-10-20 (Rollback returns an error; run rolls back on every failure), so the line numbers of the Summary shifted; the functions newGitLine and gitLine.Close did not change. Order: this issue stays the first abstraction move; [[export-repo-command-plumbing]] now depends on it, because both rewrite line.go.

## Resolution

Done: gitLine holds a *scratch.Worktree; newGitLine uses scratch.New and Close uses Worktree.Remove, both under the per-git-directory lock, so a merge-queue run and a referendum reviewer on one git directory no longer race in worktree prune. mergequeue has no worktree add or remove of its own. The merge-queue tests pass unchanged.

