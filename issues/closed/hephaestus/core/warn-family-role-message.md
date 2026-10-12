---
title: "Warn when a message to one worker names a family of roles"
state: closed          # open | in-progress | closed. Must match the directory of the file.
resolution: completed  # completed | abandoned. Set this only when state is closed.
priority: medium       # critical | high | medium | low
effort: small          # small | medium | large | x-large
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

# Warn when a message to one worker names a family of roles

## Summary

hephaestus messages send warns when --role names a family of roles and --all is not given: the message reaches no worker. The warning names the roles of the family and how to reach them.

## Context

Requested by a human on 2026-10-12, after [[queue-roles-broadcasts-keepalive]]: a message to one worker of a family name (--role formatter) reaches no worker, because one worker has one role. This should warn the user when it occurs.

## Acceptance criteria

- [x] `messages send --role formatter` warns on stderr, with the roles of the family, and still sends the message.
- [x] `--role formatter --all` and `--role formatter/upper` do not warn.

## Out of scope

<List related work that this issue does not include. Create a separate issue for each item that someone must do later.>

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

Completed: messages send warns when --role names a family of roles without --all, with the roles of the family and how to reach them. See [[unreleased-warn-family-role-message]].

