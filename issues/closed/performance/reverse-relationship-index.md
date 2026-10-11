---
title: "Build sub-issue and dependent lookups once per command"
state: closed          # open | in-progress | closed. Must match the directory of the file.
resolution: abandoned  # completed | abandoned. Set this only when state is closed.
priority: low          # critical | high | medium | low
effort: small          # small | medium | large | x-large
tags: [performance, cli] # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-11
closed: 2026-10-11     # YYYY-MM-DD. Set this only when state is closed.
owner:                 # The person or agent working on the issue. Cleared when the issue returns to open.
approver:              # The person or agent who accepts the issue when it closes.
bookmark:              # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: []         # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
metadata: {}           # Optional. Arbitrary key-value notes, for example an AI agent session id to resume work with context: {session-id: "abc123"}.
deferred-from:         # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent: "[[performance/large-repo-performance]]" # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
depends-on: []           # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# Build sub-issue and dependent lookups once per command

## Summary

`subIssues` and `dependents` scan every issue for each call, and `list --ready` builds its parent set separately. Closing an epic takes 0.5 s at 20,000 issues. Build the reverse maps (parent → sub-issues, issue → dependents) once per command, when a command first needs them.

## Context

Measured on 2026-10-11 with sisyphus 0.0.57 on a 64-core Linux machine, with generated repos of 1,000, 5,000, and 20,000 issues (10% epics, 70% sub-issues, 40% with 1-3 dependencies, 30% in subdirectories, 3 body wikilinks per issue, one design document per 20 issues). Times are the median of 3 runs. See [[large-repo-performance]] for the full table.

`updateIssue` calls `subIssues(x, name)` and `dependents(x, name)` when an issue closes; each loops over
`x.issues`, parses each document, and calls `x.canonical` on its parent or dependencies. `graph` builds its own
edges, and `lint` and `list --ready` scan again. One lazily built pair of maps in `issueIndex` serves all of
them. Most of the cost today is the parse ([[parse-frontmatter-once]]); this change removes the repeated
scans.

## Acceptance criteria

- [ ] `issueIndex` has lazily built maps of sub-issues and dependents, used by `update`, `list --ready`, `lint`, and `graph`.
- [ ] Every test passes unchanged.

## Out of scope

Nothing.

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

Abandoned on 2026-10-11, after a measured attempt. After [[0.0.61]] (no regular expression), [[0.0.62]] (frontmatter only), and [[0.0.63]] (parallel reads), the repeated scans only walk cached headers, so a reverse index no longer saves time. At 20,000 generated issues, a lazily built parent map and dependents map gave: closing an epic 222 → 220 ms (noise), and `list --ready` 297 → 313 ms (slower, because the map must cover closed issues too, which the old scan skipped). The change was not kept. `lint` and `graph` already build their links in one pass each.

