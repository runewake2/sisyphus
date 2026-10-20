---
title: "PlanKick ignores QueueErr and plans a blind kick on empty queue data"
state: closed          # open | in-progress | closed. Must match the directory of the file.
resolution: completed  # completed | abandoned. Set this only when state is closed.
priority: medium       # critical | high | medium | low
effort: medium         # small | medium | large | x-large
tags: []               # The components that the work touches, for example [widget-scheduler, plan]
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

# PlanKick ignores QueueErr and plans a blind kick on empty queue data

## Summary

taskInput sets in.QueueErr when the queue answered Connect but Workers() or Pending() failed (cmd/hephaestus/doctor.go:29-39). The doctor treats that state as unknown: queueKnown is false (internal/hephaestus/doctor/doctor.go:222), and it makes no claim about workers or waiting messages. PlanKick never reads in.QueueErr (internal/hephaestus/doctor/kick.go:31-65): the failed calls leave in.Workers and in.Pending empty, so the live-worker guards at kick.go:42-50 pass vacuously, and k.Send = len(pendingTasks(in)) == 0 becomes true. Confirmed with the base() fixture of doctor_test.go plus QueueErr set and Workers/Pending nil: PlanKick(force=false) gives err=nil, Reopen=true, Send=true. So a kick during a transient queue fault (the server dies or errors between the health check of mq.Connect, internal/hephaestus/mq/client.go:39-49, and the Workers/Pending calls) reopens the issue and sends a second task message although one may already wait, or while a live worker holds the task - a double delivery that the no-force guards exist to stop. PlanKick should return an error when in.QueueErr is not empty, as it does for a closed issue.

## Context

<Give the background that an agent needs to start the work without other help.
Link to design docs and other issues with wikilinks, for example `[[widget-scheduler]]`,
`[[widget-scheduler#Status API]]`, or `[[<issue-name>]]`. Do not use paths.>

## Acceptance criteria

- PlanKick returns an error when in.QueueErr is not empty, and the error tells the user to start or check the queue and try again.
- The kick command does not reopen the issue, change its metadata, or send a message in that case.
- A test gives PlanKick an input with QueueErr set and a waiting task message hidden by the failed Pending call, and asserts the refusal.

## Out of scope

<List related work that this issue does not include. Create a separate issue for each item that someone must do later.>

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

Fixed: PlanKick refuses when QueueErr is set, also with --force: without the workers and the pending deliveries, the kick cannot tell a stuck task from a running one, and its message would not reach the queue. Tested in internal/hephaestus/doctor/doctor_test.go (TestPlanKickRefusesWhenTheQueueDoesNotAnswer).

