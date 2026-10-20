---
title: "One identity fallback for the commits of hephaestus"
state: open            # open | in-progress | closed. Must match the directory of the file.
resolution:            # completed | abandoned. Set this only when state is closed.
priority: medium       # critical | high | medium | low
effort: small          # small | medium | large | x-large
tags: [hephaestus]     # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-20
closed:                # YYYY-MM-DD. Set this only when state is closed.
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

# One identity fallback for the commits of hephaestus

## Summary

Three functions give a VCS a committer when the user has none, with the same shape: for each of user.name and user.email, read the config, and when it is missing or empty, add a flag with a fallback.

The copies:
- internal/hephaestus/scratch/scratch.go:120-128 (Identity): git, fallback name "hephaestus", reads the config with no working directory (so it misses the repo-local config of the worktree), does not sort.
- internal/hephaestus/mergequeue/line.go:263-274 (identity): git, fallback name "hephaestus merge queue", reads the config in a directory, sorts the args.
- internal/hephaestus/mergequeue/line.go:102-113 (jjIdentity): jj (jj config get, --config flag), fallback name "hephaestus merge queue", reads in a directory, sorts.

The proposal: one exported function in package scratch for git, scratch.Identity(dir, name string) []string, with the contract: read user.name and user.email with git config in dir (empty dir: the process directory, as today for scratch); for each missing or empty key, return -c key=fallback, with name as the fallback user.name and hephaestus@localhost as the email; sorted. The two git copies move to it. The jj copy stays in line.go: it builds jj flags with a jj command, and a flag that switches between two tools would be two functions in one; note this in the code.

What it removes: the second git copy, and the drift between them (dir-scoped lookup and sorting in one, not the other; two fallback names). What it costs: the callers of scratch.Identity today (scratch.Git) must pass a dir and a name, so the signature of an exported function changes; all callers are in this repo. The dir-scoped lookup is a small behavior change for scratch (it starts to see the repo-local config, which a configured name should win anyway — the doc of Identity already says a configured name and email win). Which fallback name each caller keeps ("hephaestus" or "hephaestus merge queue") stays as it is today, per caller. Order: after [[gitline-reuses-scratch]] or independent of it; the two do not conflict.

## Context

<Give the background that an agent needs to start the work without other help.
Link to design docs and other issues with wikilinks, for example `[[widget-scheduler]]`,
`[[widget-scheduler#Status API]]`, or `[[<issue-name>]]`. Do not use paths.>

## Acceptance criteria

- [ ] scratch.Identity(dir, name) is the one git identity fallback, with its own test (configured identity wins; missing identity gets the fallback).
- [ ] mergequeue/line.go identity() is gone; gitLine uses scratch.Identity with its name.
- [ ] jjIdentity stays, with a comment that names scratch.Identity and says why jj has its own function.
- [ ] No commit of hephaestus gets a different author or committer than today.

## Out of scope

<List related work that this issue does not include. Create a separate issue for each item that someone must do later.>

## Notes

2026-10-20, from the engineer: land this before [[export-repo-command-plumbing]], which now depends on this issue: both change scratch.go (Identity, Git) and line.go (identity, the runners), and this move is the smaller one. The fix of [[hephaestus/merge-queue/jj-error-leaves-visible-commits]] shifted the line numbers of the Summary; the three identity functions did not change.

## Resolution

<Complete this section when you close the issue.
- Completed: tell what was done. Link to the changelog version and the decision records, for example `[[0.0.5]]`.
- Abandoned: tell why the work stopped. Link to the issue or decision that replaces it, if one exists.>
