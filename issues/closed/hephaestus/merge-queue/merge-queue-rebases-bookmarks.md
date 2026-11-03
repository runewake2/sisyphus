---
title: "Rebase the bookmarks of the tasks in the merge queue"
state: closed          # open | in-progress | closed. Must match the directory of the file.
resolution: completed  # completed | abandoned. Set this only when state is closed.
priority: medium       # critical | high | medium | low
effort: medium         # small | medium | large | x-large
tags: [hephaestus, merge-queue] # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-11-03
closed: 2026-11-03     # YYYY-MM-DD. Set this only when state is closed.
owner: claude          # The person or agent working on the issue. Cleared when the issue returns to open.
approver:              # The person or agent who accepts the issue when it closes.
bookmark: samw/ai/merge-queue-rebase # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: [merge-queue-rebase] # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
metadata: {}           # Optional. Arbitrary key-value notes, for example an AI agent session id to resume work with context: {session-id: "abc123"}.
deferred-from:         # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent:                # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
depends-on: []           # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# Rebase the bookmarks of the tasks in the merge queue

## Summary

Make the merge queue rebase each merged bookmark onto its line. Today it copies the commits, so merged bookmarks stay beside main.

## Context

From the human, 2026-11-03: the merge queue must rebase, because the old bookmarks otherwise stay drifted from main. The queue does not squash: it copies each commit with jj duplicate or git cherry-pick. [[0002-merge-with-jj-in-a-jj-repo]] rejected a rebase, and [[0003-rebase-the-bookmarks-of-the-tasks]] amends it.

## Acceptance criteria

- [x] After a human moves main to the line, main..<bookmark> is empty for each merged bookmark, in jj and in git.
- [x] In jj, the change ids stay, and no copy of a commit stays in the repo.
- [x] A bookmark that the run leaves out does not change.
- [x] A working copy on a merged bookmark follows it, and a stale workspace does not stop the cleanup of the run.
- [x] The report gives a command that puts the bookmarks back.

## Out of scope

A command that moves main. A human moves main.

## Notes

2026-11-03: the tests of the merge queue cover each criterion. Disabling the check for a moved bookmark, or the --ignore-working-copy on the forget, makes a test fail.

## Resolution

Completed. See [[unreleased-merge-queue-rebases-bookmarks]] and [[0003-rebase-the-bookmarks-of-the-tasks]].

