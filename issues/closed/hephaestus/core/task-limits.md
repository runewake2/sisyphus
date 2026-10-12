---
title: "Give each task of a worker a time limit and a token limit"
state: closed          # open | in-progress | closed. Must match the directory of the file.
resolution: completed  # completed | abandoned. Set this only when state is closed.
priority: medium       # critical | high | medium | low
effort: medium         # small | medium | large | x-large
tags: [hephaestus]     # The components that the work touches, for example [widget-scheduler, plan]
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

# Give each task of a worker a time limit and a token limit

## Summary

Each task of a worker stops at a time limit (--task-timeout, timeout in the role file, or 1 hour), and each task of an AI role at a token limit (--max-tokens, or max-tokens in the role file), with every program that it started. The tokens, cost, turns, and time of each run go into the result and the Notes of its issue.

## Context

A review of hephaestus on 2026-10-12 found that agent run had no time limit for a task and no token or cost limit for a model run: a runaway model or a hung tool held its lease forever, because the keep-alive renews it. The human asked to support it.

## Acceptance criteria

- [x] A task past its time limit stops, with the programs that it started, and its issue goes to a human.
- [x] A task of an AI role past its token limit stops; the tokens count each message of the model once.
- [x] The usage of a run that the runner reports goes into the result and the Notes of its issue.
- [x] A referendum reviewer keeps the token limit of its role, and its ballot is read from the result text of a stream JSON run.
- [x] The role file has timeout and max-tokens, and hephaestus roles checks them.

## Out of scope

<List related work that this issue does not include. Create a separate issue for each item that someone must do later.>

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

Completed: each task has a time limit and, for an AI role, a token limit, and each model run records its usage. See [[unreleased-task-limits]] and [[0006-a-task-has-a-time-limit-and-a-token-limit]].

