---
title: "Diagnose and kick the stuck task of an issue"
state: closed          # open | in-progress | closed. Must match the directory of the file.
resolution: completed  # completed | abandoned. Set this only when state is closed.
priority: medium       # critical | high | medium | low
effort: medium         # small | medium | large | x-large
tags: [hephaestus, llm] # The components that the work touches, for example [widget-scheduler, plan]
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

# Diagnose and kick the stuck task of an issue

## Summary

Add hephaestus doctor <issue>, which diagnoses the task of one issue that hephaestus manages (an owner that stopped, a task message that no one will take, a delegated issue with no task message, a referendum past its deadline, a missing bookmark), and hephaestus kick <issue>, which makes a stuck task run again: it reopens the issue, removes needs-human, and sends the task back to its role, for example after a worker crashed or ran out of API keys.

## Context

Requested by a human on 2026-10-20: when hephaestus manages an agent, a human needs a way to diagnose what is wrong with the task of an issue, and a way to kick a task that is stuck, for example after running out of API keys or an agent hanging. A stuck task is invisible today: the worker acknowledged its message, so the queue will not give it again, and the issue stays in progress with an owner that is gone.

## Acceptance criteria

- [x] hephaestus doctor <issue> reports the facts of the task and a finding with its fix for each problem, changes nothing, and ends with exit code 2 when there are findings.
- [x] hephaestus kick <issue> reopens an in-progress issue, removes needs-human, keeps the delegation, and sends a new task message to the role; when one already waits, it sends no second one.
- [x] The kick refuses a task that a live worker holds, without --force, and never starts a worker.
- [x] Tests cover the stuck owner, the kick, the waiting message, and the queue that does not run.

## Out of scope

<List related work that this issue does not include. Create a separate issue for each item that someone must do later.>

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

Done: hephaestus doctor and hephaestus kick, in internal/hephaestus/doctor (the diagnosis and the kick plan, tested with fakes) and cmd/hephaestus/doctor.go. The doctor reads the issue, the workers and pending deliveries of the queue, the message log, the task workspace, and the bookmark. The kick is recorded in the Notes of the issue. See the unreleased changelog entry unreleased-doctor-and-kick.

