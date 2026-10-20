---
title: "Add hephaestus evict"
state: closed          # open | in-progress | closed. Must match the directory of the file.
resolution: completed  # completed | abandoned. Set this only when state is closed.
priority: high         # critical | high | medium | low
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
depends-on: []           # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# Add hephaestus evict

## Summary

<One or two sentences. Tell what must change and why.>

## Context

The command hephaestus evict is in work in another workspace on 2026-10-20, in internal/hephaestus/mq and internal/hephaestus/toolagent. The engineer filed this issue on 2026-10-20 to track that work, so that the abstraction moves in mq and toolagent can start after it lands. The worker of the evict command: take this issue, set the owner and the bookmark, and write the acceptance criteria.

## Acceptance criteria

<Each item is one condition that must be true before the issue is complete.>

- [ ] <Condition>

## Out of scope

<List related work that this issue does not include. Create a separate issue for each item that someone must do later.>

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

Done: hephaestus evict <worker> --force. The queue drops the worker's leases (POST /evict; the mark is in memory, consumed by the next contact, expires after one lease), the keep-alive hears the eviction and stops the task with every program it started, and the worker releases its issue (open, no needs-human, the delegation stays), does not acknowledge the message, and ends. Without --force, evict only shows what it would do and refuses with exit code 1. Tested at the queue level (internal/hephaestus/mq/evict_test.go) and end to end with a looping worker (cmd/hephaestus/evict_test.go). See the unreleased changelog entry unreleased-evict.

