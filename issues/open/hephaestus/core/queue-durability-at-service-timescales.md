---
title: "Compact the message log and bound the queue state for long-lived service use"
state: open            # open | in-progress | closed. Must match the directory of the file.
resolution:            # completed | abandoned. Set this only when state is closed.
priority: low          # critical | high | medium | low
effort: medium         # small | medium | large | x-large
tags: [hephaestus]     # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-20
closed:                # YYYY-MM-DD. Set this only when state is closed.
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

# Compact the message log and bound the queue state for long-lived service use

## Summary

The message queue is built for short runs, and that is how hephaestus is operated today: the append-only log is never compacted, the state rebuild replays the whole log, Pending() scans every message ever sent, and leases and evictions are in memory only. Fine for a session of hours or days; wrong for a long-lived service. When the delegation runner makes traffic constant, or a repo accumulates a long message history, the queue needs: truncation or snapshotting of acknowledged messages, a Pending() that does not scan the full history, and a documented stance that leases and evictions are in-memory (a restart drops them, which the lease model tolerates) and that the queue is one local host per repo (the boundary belongs in the decision record of the delegation runner).

## Context

<Give the background that an agent needs to start the work without other help.
Link to design docs and other issues with wikilinks, for example `[[widget-scheduler]]`,
`[[widget-scheduler#Status API]]`, or `[[<issue-name>]]`. Do not use paths.>

## Acceptance criteria

- [ ] The log compacts or snapshots acknowledged messages, and a state rebuild does not replay the full history.
- [ ] Pending() costs by the open deliveries, not by every message ever sent.
- [ ] The in-memory stance of leases and evictions, and the one-host-per-repo boundary, are documented (the daemon decision record, or the mq package doc).
- [ ] The mq tests cover a compacted log: nothing pending or unacknowledged is lost.

## Out of scope

<List related work that this issue does not include. Create a separate issue for each item that someone must do later.>

## Notes

2026-10-20, from a human: track this in case it comes up in the future; we will typically operate hephaestus for short periods of time, so this is not needed now. Found during the review of what hephaestus needs to work as a job coordination service. It becomes relevant with [[delegation-runner]], whose constant traffic grows the log.

## Resolution

<Complete this section when you close the issue.
- Completed: tell what was done. Link to the changelog version and the decision records, for example `[[0.0.5]]`.
- Abandoned: tell why the work stopped. Link to the issue or decision that replaces it, if one exists.>
