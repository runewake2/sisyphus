---
title: "Find dependency and parent cycles in one pass in lint"
state: closed          # open | in-progress | closed. Must match the directory of the file.
resolution: completed  # completed | abandoned. Set this only when state is closed.
priority: high         # critical | high | medium | low
effort: small          # small | medium | large | x-large
tags: [performance, lint] # The components that the work touches, for example [widget-scheduler, plan]
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

# Find dependency and parent cycles in one pass in lint

## Summary

`lint` checks for dependency cycles with a new depth-first search from every dependency of every issue (`dependsOnReaches`), so a long chain of dependencies is quadratic: 16.5 s for a chain of 3,000 issues. Find all cycles in one pass over the graph.

## Context

Measured on 2026-10-11 with sisyphus 0.0.57 on a 64-core Linux machine, with generated repos of 1,000, 5,000, and 20,000 issues (10% epics, 70% sub-issues, 40% with 1-3 dependencies, 30% in subdirectories, 3 body wikilinks per issue, one design document per 20 issues). Times are the median of 3 runs. See [[large-repo-performance]] for the full table.

```
lint on a chain where chain-step-i depends on chain-step-(i-1):
  1,000 issues:  1,849 ms
  3,000 issues: 16,528 ms
```

`lintDependsOn` in `cmd/sisyphus/lint.go` calls `dependsOnReaches(x, dependency, issue)` for each entry, and
each call starts with an empty `seen` map, so it walks the whole chain again. The parent check walks up from
each issue with `parentName`, which is linear in depth for each issue, so a deep parent tree is quadratic too.
A single depth-first search with three colors (or Tarjan's strongly connected components) finds every cycle
in O(issues + edges). The message can then name each cycle once, with its members, instead of once for each
issue in it.

`sisyphus depends-on` and `parent` check one new edge, which is linear and fast (57 ms across the 3,000
chain); they do not need to change.

## Acceptance criteria

- [ ] `lint` finds dependency cycles and parent cycles with one pass each over the whole graph.
- [ ] Each cycle is reported once, and the message names its issues in order.
- [ ] `lint` on a chain of 3,000 dependencies takes less than 500 ms.
- [ ] The lint tests for cycles pass, with the messages updated.

## Out of scope

Nothing.

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

Completed in [[0.0.59]]. `lint` finds dependency cycles with Tarjan's algorithm and parent cycles with one walk, and reports each cycle once with its issues in order. A chain of 3,000 dependencies: 15.8 s → 0.15 s.

