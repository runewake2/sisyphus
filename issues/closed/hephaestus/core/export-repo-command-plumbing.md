---
title: "repo exports Run and Lines, the shared command plumbing of hephaestus"
state: closed          # open | in-progress | closed. Must match the directory of the file.
resolution: completed  # completed | abandoned. Set this only when state is closed.
priority: medium       # critical | high | medium | low
effort: medium         # small | medium | large | x-large
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
depends-on: ["[[hephaestus/core/gitline-reuses-scratch]]", "[[hephaestus/core/one-vcs-identity-fallback]]"] # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# repo exports Run and Lines, the shared command plumbing of hephaestus

## Summary

Six places run a VCS command, capture stdout and stderr, and format the same error ("<name> <args>: %v: <trimmed stderr>"); two places split command output into trimmed non-empty lines. The copies of the runner:

- internal/hephaestus/repo/repo.go:63-79 (Repo.run): stdout not trimmed.
- internal/hephaestus/scratch/scratch.go:103-116 (Git): --git-dir or -C dir, identity prepended, trimmed.
- internal/hephaestus/taskspace/taskspace.go:122-131 (run): generic, trimmed.
- internal/hephaestus/mergequeue/line.go:89-98 (jjLine.jjIn): identity prepended, trimmed.
- internal/hephaestus/mergequeue/line.go:239-248 (gitLine.git): identity prepended, not trimmed, returns stdout also on error (no current caller uses that stdout).
- internal/hephaestus/mergequeue/line.go:251-259 (gitLine.gitDirCmd): --git-dir, trimmed.
Also cmd/hephaestus/referendum.go:372-375 (revParse) runs git --git-dir with raw Output().

The copies of the line splitter:
- internal/hephaestus/repo/repo.go:81-90 (lines): trims each line, drops empty lines, applies filepath.ToSlash.
- internal/hephaestus/mergequeue/mergequeue.go:91-99 (lines): the same, without ToSlash; line.go uses it on commit ids and on file paths (AddedSince, conflict files), where mergequeue.go:204 then joins with "/" — the ToSlash is the fix this copy did not get, visible only on Windows.

The proposal: export two functions from internal/hephaestus/repo, the lowest package that all of them import or can import: Run(dir, name string, args ...string) (string, error) with trimmed stdout and the standard error format, and Lines(output string) []string (the ToSlash version; the method run and the private lines become wrappers or go). scratch, taskspace, and mergequeue build their arg lists (identity, --git-dir) and call repo.Run; mergequeue drops its lines.

What it removes: ~50 lines, and six copies of one error format that a human reads — one reason to change. What it costs: scratch gains its first internal import (repo imports nothing from scratch, so there is no cycle); gitLine.git loses stdout-on-error (unused) and gains trimming (its callers trim or split already). Order: export Run and Lines with tests, then move one package per change: scratch, taskspace, mergequeue.

## Context

<Give the background that an agent needs to start the work without other help.
Link to design docs and other issues with wikilinks, for example `[[widget-scheduler]]`,
`[[widget-scheduler#Status API]]`, or `[[<issue-name>]]`. Do not use paths.>

## Acceptance criteria

- [ ] repo.Run and repo.Lines exist, with tests for the error format, the trimming, and the ToSlash.
- [ ] scratch.Git, taskspace.run, jjLine.jjIn, gitLine.git, and gitLine.gitDirCmd build on repo.Run; mergequeue.lines is gone.
- [ ] No error text and no report output of hephaestus changes.
- [ ] go test ./... passes after each move.

## Out of scope

<List related work that this issue does not include. Create a separate issue for each item that someone must do later.>

## Notes

2026-10-20, from the engineer: this issue now depends on [[gitline-reuses-scratch]] and [[one-vcs-identity-fallback]], because all three rewrite the same regions of line.go and scratch.go; the two small moves land first. The fix of [[hephaestus/merge-queue/jj-error-leaves-visible-commits]] changed mergequeue.go and line.go on 2026-10-20: Rollback returns an error, run() rolls back on every failure, and mergequeue_test.go is new. The claim of the Summary still holds (no caller of gitLine.git uses its stdout on error: Rollback discards it), but the line numbers shifted, and the new rollback tests must stay green through each move. This issue does not subsume [[share-jj-export-import]]: Run and Lines are plumbing, ExportBookmarks and ImportBookmarks are a VCS operation on repo.Repo; keep both, this one first.

## Resolution

Done: repo.Run(dir, name, hidden, args...) and repo.Lines are the shared command plumbing. Run returns trimmed stdout and the one error format (name, args, trimmed stderr); the hidden flags (an identity, a --git-dir) run but stay out of the error, so no error text changed. scratch.Git, taskspace.run, jjLine.jjIn, gitLine.git, gitLine.gitDirCmd, and revParse of cmd/hephaestus build on it; mergequeue.lines is gone and line.go uses repo.Lines, which carries the ToSlash that copy missed. Repo.run stays as the untrimmed wrapper over the same core, because WorkingDiff needs the last newline of a diff. Tests cover the error format, the trimming, the hidden flags, and Lines; the full suite passes.

