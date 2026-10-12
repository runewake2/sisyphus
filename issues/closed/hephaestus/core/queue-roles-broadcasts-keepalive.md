---
title: "Support roles of families in messages, deliver broadcasts once to each role, and keep leases alive"
state: closed          # open | in-progress | closed. Must match the directory of the file.
resolution: completed  # completed | abandoned. Set this only when state is closed.
priority: medium       # critical | high | medium | low
effort: medium         # small | medium | large | x-large
tags: [hephaestus, messages] # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-12
closed: 2026-10-12     # YYYY-MM-DD. Set this only when state is closed.
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

# Support roles of families in messages, deliver broadcasts once to each role, and keep leases alive

## Summary

Allow role names with families in the message queue, deliver a broadcast once to each listening role and treat it as a notice that does not claim its issue, and add a keep-alive that renews the lease of a message while a worker handles it.

## Context

A review of hephaestus on 2026-10-12 found three defects of the message queue, each reproduced in a scratch repo: a role of a family (formatter/golang) could not receive or be sent a message; every worker that listens for bookmark-ready or issue-closed failed, because it claimed the closed issue of the broadcast; and a task longer than its lease ran again on a second worker (one message ran four times). The human asked: support roles of families in messages, fix the order of broadcasts, and give workers a keep-alive ping that tells the scheduler that they still work, so that no work is duplicated.

## Acceptance criteria

- [x] A worker of formatter/golang receives its messages, and a message to all workers of the family formatter reaches each role in it.
- [x] A broadcast reaches one worker of each listening role, also a worker that starts after the send.
- [x] A broadcast, or a message to every worker of a role, does not claim, close, or reopen its issue; the result is in its Notes.
- [x] The implementer broadcasts bookmark-ready before it closes its issue.
- [x] A worker renews its lease while it works: a task longer than the lease is not given to a second worker.
- [x] `hephaestus messages workers` shows when each worker last spoke and what it holds.

## Out of scope

<List related work that this issue does not include. Create a separate issue for each item that someone must do later.>

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

Completed: roles of families in messages, broadcasts once to each listening role as notices that do not claim their issue, the implementer broadcasts before it closes, and workers keep their leases alive. See [[unreleased-queue-roles-broadcasts-keepalive]] and [[0002-broadcasts-reach-each-role-once-and-workers-keep-their-leases-alive]].

