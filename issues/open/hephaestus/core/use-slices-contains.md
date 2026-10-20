---
title: "toolagent uses slices.Contains instead of its own contains"
state: open            # open | in-progress | closed. Must match the directory of the file.
resolution:            # completed | abandoned. Set this only when state is closed.
priority: low          # critical | high | medium | low
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

# toolagent uses slices.Contains instead of its own contains

## Summary

internal/hephaestus/toolagent/toolagent.go:759-766 (contains) repeats slices.Contains from the standard library, which 15 other places in internal/hephaestus already use (for example referendum/voter.go:171, mergequeue/mergequeue.go:204). Reuse first: delete the function and call slices.Contains(items, item) at toolagent.go:752. No behavior changes. (The test helpers contains(t, s, substring) in three _test.go files are a different thing and stay: see the Notes of [[hephaestus/roles/abstraction-engineer-agent]].)

## Context

<Give the background that an agent needs to start the work without other help.
Link to design docs and other issues with wikilinks, for example `[[widget-scheduler]]`,
`[[widget-scheduler#Status API]]`, or `[[<issue-name>]]`. Do not use paths.>

## Acceptance criteria

- [ ] toolagent has no contains function; the call site uses slices.Contains.
- [ ] go test ./internal/hephaestus/toolagent passes unchanged.

## Out of scope

<List related work that this issue does not include. Create a separate issue for each item that someone must do later.>

## Notes

2026-10-20, from the engineer: this issue now depends on [[evict-command]]: the one-line change is in toolagent.go, where hephaestus evict is in work now. It is not urgent; land it with or after the other toolagent moves.

## Resolution

<Complete this section when you close the issue.
- Completed: tell what was done. Link to the changelog version and the decision records, for example `[[0.0.5]]`.
- Abandoned: tell why the work stopped. Link to the issue or decision that replaces it, if one exists.>
