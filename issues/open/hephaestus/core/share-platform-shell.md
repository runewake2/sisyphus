---
title: "The merge queue checks run in the shared platform shell that stops the process group"
state: open            # open | in-progress | closed. Must match the directory of the file.
resolution:            # completed | abandoned. Set this only when state is closed.
priority: high         # critical | high | medium | low
effort: small          # small | medium | large | x-large
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
depends-on: ["[[hephaestus/core/evict-command]]"] # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# The merge queue checks run in the shared platform shell that stops the process group

## Summary

mergequeue.runChecks repeats the platform-shell choice of toolagent.Shell, without the fix that Shell has. The copies:

- internal/hephaestus/toolagent/toolagent.go:670-680 (Shell) with stopTree in shell_unix.go/shell_windows.go: sh -c or cmd /C, the command in a process group of its own, a cancel stops the whole group, WaitDelay 5s. Used by toolagent (603, 720) and referendum.RoleVoter (voter.go:238).
- internal/hephaestus/mergequeue/mergequeue.go:260-273 (runChecks, with isWindows): the same shell choice, but no process group and no WaitDelay. When a merge-queue run is stopped, the children of the checks (go test, builds) keep running and spend CPU.

The proposal: move Shell and stopTree into a small package of their own, for example internal/hephaestus/shell with Command(ctx, line string) *exec.Cmd, so that mergequeue does not have to import all of toolagent (toolagent pulls mq, taskspace, roles, watch). toolagent.Shell stays as a thin alias or its callers move too. runChecks builds its cmd from the shared function and keeps its own output wiring.

What it removes: the second copy of the platform-shell contract — two copies that must stay the same, because the process-group stop is a fix one copy got and the other did not. What it costs: one new small package (two files move; the Go build constraints come with them). The checks of the merge queue gain the process-group stop and the WaitDelay — that is the intended fix, and the only behavior change; the issue carries it on purpose, as the race fix of [[gitline-reuses-scratch]] does. Order: move the package first, point toolagent and voter at it, then runChecks.

## Context

<Give the background that an agent needs to start the work without other help.
Link to design docs and other issues with wikilinks, for example `[[widget-scheduler]]`,
`[[widget-scheduler#Status API]]`, or `[[<issue-name>]]`. Do not use paths.>

## Acceptance criteria

- [ ] One function builds the platform shell command with the process-group stop; toolagent, referendum.RoleVoter, and mergequeue.runChecks use it.
- [ ] mergequeue has no sh -c / cmd /C choice and no isWindows of its own.
- [ ] A stopped merge-queue run leaves no child of the checks running (the existing stopTree behavior, now also for the checks).
- [ ] The outputs of the checks in the report do not change.

## Out of scope

<List related work that this issue does not include. Create a separate issue for each item that someone must do later.>

## Notes

2026-10-20, from the engineer: this issue now depends on [[evict-command]]: it moves toolagent.Shell and stopTree out of toolagent, and hephaestus evict is in work in toolagent and mq now. Land it right after evict: it carries a real fix (the process-group stop for the checks). The fix of [[hephaestus/merge-queue/jj-error-leaves-visible-commits]] shifted runChecks in mergequeue.go; the function did not change.

## Resolution

<Complete this section when you close the issue.
- Completed: tell what was done. Link to the changelog version and the decision records, for example `[[0.0.5]]`.
- Abandoned: tell why the work stopped. Link to the issue or decision that replaces it, if one exists.>
