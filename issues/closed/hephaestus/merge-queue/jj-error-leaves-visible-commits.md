---
title: "An error of a jj merge-queue run leaves its picked commits as a visible head"
state: closed          # open | in-progress | closed. Must match the directory of the file.
resolution: completed  # completed | abandoned. Set this only when state is closed.
priority: medium       # critical | high | medium | low
effort: medium         # small | medium | large | x-large
tags: []               # The components that the work touches, for example [widget-scheduler, plan]
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

# An error of a jj merge-queue run leaves its picked commits as a visible head

## Summary

Run promises: an error stops the run, and changes nothing outside the scratch working copy (internal/hephaestus/mergequeue/mergequeue.go:101-102). The jj line does not keep that promise. When Run returns an error after one or more picks (Pick, HasMerges, decisions.Apply, release.Apply, or CommitFiles at mergequeue.go:163-244, or the ctx.Err() return at mergequeue.go:139-141 after a bookmark merged), it does not call Rollback. The deferred jjLine.Close (internal/hephaestus/mergequeue/line.go:199-206) forgets the scratch workspace and abandons only the working-copy commit @. The duplicated commits below @ stay visible: the parent of @ becomes an anonymous head in the shared jj repo, and the user sees a stray line of commits in jj log, with no bookmark on it. A ctrl-c in the middle of a run leaks every bookmark that merged so far. Reproduced with jj 0.45.1: workspace add -r base, duplicate --insert-before @, then workspace forget plus abandon of @ only; the duplicate (commit 4fd9d373, 'feat') stays in the log of the user as an anonymous head. The git line does not have the problem: an unreferenced commit in the object store is invisible. Note also that jjLine.Rollback (line.go:188-190) and Close ignore the errors of jj abandon; a failed rollback silently leaves the commits of a left-out bookmark on the line, and the next bookmarks merge on top of them.

## Context

<Give the background that an agent needs to start the work without other help.
Link to design docs and other issues with wikilinks, for example `[[widget-scheduler]]`,
`[[widget-scheduler#Status API]]`, or `[[<issue-name>]]`. Do not use paths.>

## Acceptance criteria

- When mergequeue.Run returns an error in a jj repo, no commit that the run duplicated stays visible in the repo: jj log shows the same heads as before the run, except a bookmark that the run created.
- The ctx cancellation path (mergequeue.go:139-141) cleans up the same way.
- The successful path still keeps the merged line: the bookmark of the result survives Close.
- A test drives Run into an error after at least one successful Pick in a jj repo, and asserts that the repo has no new anonymous head.

## Out of scope

<List related work that this issue does not include. Create a separate issue for each item that someone must do later.>

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

Fixed: Run now rolls the line back to the last good head on every return that is not success (a deferred rollback guarded by a done flag), and a rollback that fails is an error of the run: line.Rollback returns an error, leave propagates it, and the jj line reports a failed abandon instead of ignoring it. A test with a failing pick (internal/hephaestus/mergequeue/mergequeue_test.go) shows the rollback to the base.

