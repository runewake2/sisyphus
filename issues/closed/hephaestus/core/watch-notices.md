---
title: "Send bookmark-ready and issue-closed from code"
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

# Send bookmark-ready and issue-closed from code

## Summary

hephaestus watch broadcasts bookmark-ready and issue-closed from the state of the issues and their bookmarks, once each, and the implementer no longer sends them.

## Context

A review of hephaestus on 2026-10-12 found that bookmark-ready (16 roles listen) was sent only if an AI implementer remembered the step of its prompt, and issue-closed (technical-docs) was never sent. The human asked for both to come from code.

## Acceptance criteria

- [x] When an issue closes, watch broadcasts bookmark-ready (if its bookmark has work that the base does not have) and then issue-closed.
- [x] When the bookmark of a closed issue moves, watch broadcasts bookmark-ready again; each notice is sent once.
- [x] The first run sends nothing, and a failure loses no notice.
- [x] The implementer and the Works with sections of the roles say that watch sends the notices.

## Out of scope

<List related work that this issue does not include. Create a separate issue for each item that someone must do later.>

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

Completed: hephaestus watch broadcasts bookmark-ready and issue-closed from the state of the issues, once each. See [[unreleased-watch-notices]] and [[0007-the-watcher-sends-the-notices-of-a-close]].

