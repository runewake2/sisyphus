---
title: "Find an issue by its full name without walking every state directory"
state: closed          # open | in-progress | closed. Must match the directory of the file.
resolution: completed  # completed | abandoned. Set this only when state is closed.
priority: medium       # critical | high | medium | low
effort: small          # small | medium | large | x-large
tags: [performance, cli] # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-11
closed: 2026-10-11     # YYYY-MM-DD. Set this only when state is closed.
owner: claude          # The person or agent working on the issue. Cleared when the issue returns to open.
approver:              # The person or agent who accepts the issue when it closes.
bookmark: samw/ai/perf-fixes # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: [perf-fixes] # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
metadata: {}           # Optional. Arbitrary key-value notes, for example an AI agent session id to resume work with context: {session-id: "abc123"}.
deferred-from:         # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent: "[[performance/large-repo-performance]]" # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
depends-on: []           # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# Find an issue by its full name without walking every state directory

## Summary

`show`, `update`, `edit`, and the other commands that take one issue walk every state directory first, also for a full name such as `web/fix-login-bug`. At 20,000 issues that is 37 ms for `show`. Look up `issues/<state>/<name>.md` directly first.

## Context

Measured on 2026-10-11 with sisyphus 0.0.57 on a 64-core Linux machine, with generated repos of 1,000, 5,000, and 20,000 issues (10% epics, 70% sub-issues, 40% with 1-3 dependencies, 30% in subdirectories, 3 body wikilinks per issue, one design document per 20 issues). Times are the median of 3 runs. See [[large-repo-performance]] for the full table.

`loadIndex` walks `issues/open`, `issues/in-progress`, and `issues/closed` in full for every command
(`cmd/sisyphus/index.go`). For a full name, three `stat` calls find the file and also show whether it is in
more than one state directory. Only a shorter name (a bare file name of an issue in a subdirectory), a name in
another case ([[case-insensitive-issue-lookup]]), or a command that needs other issues (closing, which warns
about sub-issues and dependents) needs the walk. The index could walk lazily: on the first lookup that the
direct path does not answer.

## Acceptance criteria

- [ ] A command that gets a full name and needs no other issue finds it without walking the state directories.
- [ ] Shorter names, other cases, and duplicates in two state directories behave as before.
- [ ] `show <full-name>` at 20,000 issues takes less than 10 ms.

## Out of scope

Nothing.

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

Completed in [[0.0.64]]. The index walks the state directories only when a command needs them; a full name is found with a stat in each state directory. `show` of a full name at 20,000 issues: 42 → 5 ms.

