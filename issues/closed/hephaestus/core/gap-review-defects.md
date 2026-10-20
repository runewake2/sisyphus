---
title: "Fix the defects of the hephaestus gap review, and add a role trial"
state: closed          # open | in-progress | closed. Must match the directory of the file.
resolution: completed  # completed | abandoned. Set this only when state is closed.
priority: medium       # critical | high | medium | low
effort: medium         # small | medium | large | x-large
tags: [hephaestus]     # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-13
closed: 2026-10-13     # YYYY-MM-DD. Set this only when state is closed.
owner: claude          # The person or agent working on the issue. Cleared when the issue returns to open.
approver:              # The person or agent who accepts the issue when it closes.
bookmark: samw/ai/ux-fixes # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: [ux-fixes] # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
metadata: {}           # Optional. Arbitrary key-value notes, for example an AI agent session id to resume work with context: {session-id: "abc123"}.
deferred-from:         # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent: "[[hephaestus/core/delegate-issues-to-agents]]" # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
depends-on: []           # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# Fix the defects of the hephaestus gap review, and add a role trial

## Summary

Fix three defects that the gap review of hephaestus found, and add a command to try a role on one issue without a change to the repo.

## Context

The gap review of [[delegate-issues-to-agents]] found these problems:

- Two bookmarks at the same commit make `merge-order` report a cycle: each is "based on" the other. This stops the whole `merge-queue` run. It occurs when an agent makes its bookmark on the head of another agent before its first commit.
- `merge-queue` exits 0 when it leaves work out, also when it merges nothing. A script cannot tell that from success, and the worker of the merge queue closes the request as done.
- `roles/x.md` and `roles/x/role.md` both give the role `x`, and no check reports it. `roles/a-b.md` and `roles/a/b/role.md` give the same worker and subagent name `a-b`.
- The trial roles need a real run each, and each run is a manual session. There is no record to compare the roles.

## Acceptance criteria

- [x] Bookmarks at the same commit merge as one, with a warning, and not as a cycle.
- [x] `merge-queue` exits 2 when it leaves work out, 0 when all work is in the line or in the base already, and 1 when it cannot run.
- [x] `hephaestus roles` reports a role with two role files, and roles with the same flat name. A command that finds a role with two role files fails.
- [x] `hephaestus agent trial <role> <issue>` runs one task of a role in a scratch copy, changes nothing in the repo or the message queue, and records its result, time, tokens, and cost. `hephaestus agent trials` lists the records.

## Out of scope

- The dispatcher, the scheduler daemon, the schedule triggers, and the configuration file: [[delegation-runner]].

## Notes

## Resolution

Completed. Bookmarks at one commit merge as one; merge-queue exits 2 when it leaves work out; hephaestus roles reports duplicate role names and flat names; hephaestus agent trial and agent trials try a role on one issue in a scratch copy and record the result and cost. See [[0.0.134]].

