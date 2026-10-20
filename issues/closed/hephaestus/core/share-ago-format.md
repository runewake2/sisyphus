---
title: "One ago() formats how long ago a time was, for the doctor and the status"
state: closed          # open | in-progress | closed. Must match the directory of the file.
resolution: completed  # completed | abandoned. Set this only when state is closed.
priority: low          # critical | high | medium | low
effort: small          # small | medium | large | x-large
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
depends-on: ["[[hephaestus/core/evict-command]]"] # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# One ago() formats how long ago a time was, for the doctor and the status

## Summary

Two identical functions format how long ago a time was, for a human, with the same words and the same steps (never since the queue started; Ns, Nm, Nh up to 48h, Nd):

- internal/hephaestus/doctor/doctor.go:166-180 (ago(now, t)), also used by doctor/kick.go:44.
- internal/hephaestus/status/status.go:255-269 (the ago closure in Report.Text).

Both describe the records of the message queue (the LastSeen of a worker, the age of a task) to a human, so the format must stay the same: a reader of hephaestus status and hephaestus doctor should see the same words. Two copies that share a format count as one thing.

The proposal: one exported function, mq.Ago(now, t time.Time) string, in internal/hephaestus/mq — the package that holds the workers and messages whose times these words describe, and that doctor and status both import already (no new import, no cycle). Alternative home, if mq should stay pure transport: doctor exports Ago and status imports doctor (also no cycle). The mover picks one and notes it.

What it removes: one copy of a human-facing format, and the drift between the doctor and the status. What it costs: one exported name in a package that is otherwise not about text. No behavior changes.

## Context

<Give the background that an agent needs to start the work without other help.
Link to design docs and other issues with wikilinks, for example `[[widget-scheduler]]`,
`[[widget-scheduler#Status API]]`, or `[[<issue-name>]]`. Do not use paths.>

## Acceptance criteria

- [ ] One exported Ago(now, t) exists, with a test for the zero time and each step (s, m, h, d).
- [ ] doctor and status use it; neither has a copy.
- [ ] The outputs of hephaestus status and hephaestus doctor do not change.

## Out of scope

<List related work that this issue does not include. Create a separate issue for each item that someone must do later.>

## Notes

2026-10-20, from the engineer: this issue now depends on [[evict-command]], because the proposed home mq is where hephaestus evict is in work now. If the mover picks the alternative home (doctor exports Ago, status imports doctor), the dependency can be cleared. Note also: doctor/kick.go changed on 2026-10-20 with the kick fixes; ago() itself did not change.

## Resolution

Done: mq.Ago is the one human-facing ago format; the doctor and the status report use it. The text is unchanged.

