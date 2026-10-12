---
title: "Give each task a workspace and a bookmark of its own"
state: closed          # open | in-progress | closed. Must match the directory of the file.
resolution: completed  # completed | abandoned. Set this only when state is closed.
priority: medium       # critical | high | medium | low
effort: large          # small | medium | large | x-large
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

# Give each task a workspace and a bookmark of its own

## Summary

Each task of a role with workspace: task runs in a jj workspace or git worktree of its own, with its own bookmark, which hephaestus makes, and removes when the task is done. A notice about a bookmark runs in a scratch copy of it, and its changes go to a fix bookmark.

## Context

A review of hephaestus on 2026-10-12 found that an AI worker ran in its own directory or the one that the message named, while the role prompts and AGENTS ask for a workspace of its own. Parallel workers could edit the same files, and tool agents on a notice wrote into the workspace of the implementer. The human asked to support a workspace per task.

## Acceptance criteria

- [x] A task of a task role runs in workspaces/<repo>-<role>-<issue>, with the bookmark <prefix><role>-<issue>, in a git repo and in a jj repo.
- [x] The worker claims the issue in the workspace; the model gets the workspace and the bookmark in its prompt.
- [x] A done task announces its close and removes its workspace; a failed or stopped task, or one with changes that no commit keeps, keeps it, with a Note.
- [x] A notice about a bookmark runs in a scratch copy; its changes go to <prefix>fix-<role>-<issue>, and the working copy of the worker does not change.
- [x] The scratch worktrees of the referendum use the same package.

## Out of scope

<List related work that this issue does not include. Create a separate issue for each item that someone must do later.>

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

Completed: each task of a task role runs in a workspace and on a bookmark of its own, which hephaestus makes and removes, and a notice about a bookmark runs in a scratch copy of it. See [[unreleased-task-workspaces]] and [[0008-each-task-has-a-workspace-of-its-own]].

