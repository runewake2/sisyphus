---
title: "Add hephaestus status: one view of what needs a human"
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

# Add hephaestus status: one view of what needs a human

## Summary

hephaestus status shows the issues tagged needs-human, the referendums that run or need a decision, the live and stopped workers, the messages that wait by receiver, and the last line of the merge queue with the recommendation of the release manager.

## Context

Requested by a human on 2026-10-12, from a review of hephaestus: there was no single view of what needs a human. A person had to read sisyphus list --tags needs-human, the referendum metadata, messages workers, and the merge queue line.

## Acceptance criteria

- [x] It lists the open and in-progress issues tagged needs-human.
- [x] It lists running referendums with their deadline, overdue ones, failed ones on an open issue, and ones with no result, each with the next action.
- [x] It lists live and stopped workers (--stale), and the messages that wait for each receiver, with whether a live worker can take them.
- [x] It shows the last line of the merge queue, whether the base has it, and the last release-recommendation for it; the release manager sends that broadcast.
- [x] It works without a message queue, and --json prints the report.

## Out of scope

<List related work that this issue does not include. Create a separate issue for each item that someone must do later.>

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

Completed: hephaestus status shows what needs a human in one view. See [[unreleased-status-command]].

