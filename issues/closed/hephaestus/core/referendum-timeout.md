---
title: "Fail a referendum that times out, and keep the issue open"
state: closed          # open | in-progress | closed. Must match the directory of the file.
resolution: completed  # completed | abandoned. Set this only when state is closed.
priority: medium       # critical | high | medium | low
effort: small          # small | medium | large | x-large
tags: [hephaestus, referendum] # The components that the work touches, for example [widget-scheduler, plan]
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

# Fail a referendum that times out, and keep the issue open

## Summary

Give a referendum a time limit (--timeout, 1 hour by default) and stop it on Ctrl-C and SIGTERM: it then fails, its reviewers stop with every program that they started, and the issue stays open. A deadline in the metadata of the issue lets the next call record a referendum that stopped with no record as failed.

## Context

A review of hephaestus on 2026-10-12 found that a referendum that is stopped (Ctrl-C, SIGTERM, a crash, or the time limit of a shell tool) left referendum=round-<n> on its issue, which refused every later call until a human ran decide, and left its reviewers and scratch worktrees behind. The human decided: a referendum that times out fails, and the issue remains open.

## Acceptance criteria

- [x] A referendum that times out fails, with the reason in the Notes and the record; the issue keeps its state and gets no needs-human tag.
- [x] Its reviewers stop, also the programs that their shell started, and no scratch worktree stays.
- [x] Ctrl-C and SIGTERM stop a referendum the same way.
- [x] A call after the deadline of a referendum that stopped with no record records it as failed, and goes on; a call before the deadline is refused with the deadline.

## Out of scope

<List related work that this issue does not include. Create a separate issue for each item that someone must do later.>

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

Completed: a referendum that times out or is stopped fails, its reviewers stop with their child programs, and the issue stays open; a deadline in the metadata lets the next call record a crashed referendum as failed. See [[unreleased-referendum-timeout]].

