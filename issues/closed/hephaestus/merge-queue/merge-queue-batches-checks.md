---
title: "Run the checks of the merge queue once for a batch of bookmarks"
state: closed          # open | in-progress | closed. Must match the directory of the file.
resolution: completed  # completed | abandoned. Set this only when state is closed.
priority: medium       # critical | high | medium | low
effort: medium         # small | medium | large | x-large
tags: [hephaestus, merge-queue] # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-11-03
closed: 2026-11-03     # YYYY-MM-DD. Set this only when state is closed.
owner: claude          # The person or agent working on the issue. Cleared when the issue returns to open.
approver:              # The person or agent who accepts the issue when it closes.
bookmark: samw/ai/merge-queue-batches # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: [merge-queue-batches] # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
metadata: {}           # Optional. Arbitrary key-value notes, for example an AI agent session id to resume work with context: {session-id: "abc123"}.
deferred-from:         # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent:                # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
depends-on: []           # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# Run the checks of the merge queue once for a batch of bookmarks

## Summary

Let the merge queue test several ready bookmarks together, with one run of the checks, so that it can catch up when the tests are long.

## Context

From the human, 2026-11-03: when A and B have no conflicts and are ready to merge, they can both be part of one check. This must help when the queue backs up because of long tests. See [[0004-run-the-checks-once-for-a-batch]].

## Acceptance criteria

- [x] --batch <n> runs the checks once for up to n bookmarks.
- [x] A batch that fails finds each bookmark that fails, and the others merge, with versions that have no gap.
- [x] A conflict with a member of a failed batch is tried again.
- [x] A dependent of a failed bookmark is left out.
- [x] The run does not run the checks again on a tree that it knows fails.
- [x] The default stays one run for each bookmark.

## Out of scope

Runs of the checks in parallel, and the use of a batch in the command of the role merge-queue. A human chooses that.

## Notes

2026-11-03: this work is based on samw/ai/merge-queue-rebase, which is not in main yet, because both change the same code.

## Resolution

Completed. See [[unreleased-merge-queue-batches-checks]] and [[0004-run-the-checks-once-for-a-batch]].

