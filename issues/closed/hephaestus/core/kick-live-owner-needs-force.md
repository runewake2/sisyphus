---
title: "Doctor says a live owner needs kick --force, but kick proceeds without it"
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

# Doctor says a live owner needs kick --force, but kick proceeds without it

## Summary

The doctor and the kick disagree about the task of an in-progress issue whose owner is live but holds no task message. diagnoseInProgress (internal/hephaestus/doctor/doctor.go:286-290) reports the finding 'The owner ... is live ..., but holds no message of this task' with the fix 'hephaestus kick <issue> --force (the worker is live)'. But PlanKick without force refuses only a live owner that also holds the task (internal/hephaestus/doctor/kick.go:42-45: alive(in, w) && holdsTask(in, w)); with Holds empty, a plain kick passes, reopens the issue (which drops the live owner), and sends a new task message (confirmed: PlanKick(force=false) gives err=nil, Reopen=true, Send=true for that input). The long help of kick (cmd/hephaestus/doctor.go:118-119) says 'A task that a live worker holds is slow, not stuck: kick refuses it without --force', so a user who follows the doctor passes --force when it is not needed, and --force also turns off the leased-pending-message guard (kick.go:46-50). A worker acknowledges a message only after Handle returns (internal/hephaestus/toolagent/toolagent.go:193-201), so a live owner without a hold finished the message but may still finish the issue (close it, move the bookmark); a plain kick then reopens under it and double-delivers the task to the role. Make the two agree: either PlanKick without force refuses a live owner of an in-progress issue (and the doctor keeps its fix), or the doctor drops '--force' from the fix; the first is the safer reading. A test covers PlanKick for the live-owner-without-hold case (none does today: doctor_test.go:73-80 tests only Diagnose).

## Context

<Give the background that an agent needs to start the work without other help.
Link to design docs and other issues with wikilinks, for example `[[widget-scheduler]]`,
`[[widget-scheduler#Status API]]`, or `[[<issue-name>]]`. Do not use paths.>

## Acceptance criteria

- The fix text of diagnoseInProgress for a live owner without a hold and the force guard of PlanKick agree: the command that the doctor prints behaves as the doctor says.
- If kick refuses: PlanKick(in, role, false) returns an error for an in-progress issue whose owner is alive, also when the owner holds no task message, and PlanKick(in, role, true) proceeds.
- A test in internal/hephaestus/doctor covers PlanKick for a live owner without a hold, with and without force.

## Out of scope

<List related work that this issue does not include. Create a separate issue for each item that someone must do later.>

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

Fixed: PlanKick refuses every live owner of an in-progress issue without --force, whether or not it holds the task message, with the reason in the error. This matches the fix text of the doctor. Tested in internal/hephaestus/doctor/doctor_test.go (TestPlanKickRefusesALiveOwnerWithoutTheTask).

