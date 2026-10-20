---
title: "Draw graph nodes in linear time"
state: closed          # open | in-progress | closed. Must match the directory of the file.
resolution: completed  # completed | abandoned. Set this only when state is closed.
priority: high         # critical | high | medium | low
effort: small          # small | medium | large | x-large
tags: [performance, graph] # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-20
closed: 2026-10-20     # YYYY-MM-DD. Set this only when state is closed.
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

# Draw graph nodes in linear time

## Summary

<One or two sentences. Tell what must change and why.>

## Context

Measured on 2026-10-20 with sisyphus 0.0.117 on a generated repo of 20,000 issues (the shape of [[performance/large-repo-performance]]), graph <epic> --full --json takes 1,375 ms, while graph <epic> --json takes 250 ms. The CPU profile has 70% of the samples in buildGraph, almost all in the node placement loop (cmd/sisyphus/graph.go:171-184):

- graph.go:177-178: for each placed node, the loop scans every drawn edge to find the edges that leave it (4.9 s cumulative of 7.5 s in a 5-run profile). This is O(nodes x edges).
- graph.go:172: the loop sorts the whole ready list again on every iteration (1.5 s cumulative). This is O(n^2 log n) in the worst case.

A prototype that buckets the edges by From in a map before the loop, and keeps ready as a min-heap of names, gives the same placement order (the lexicographically smallest ready node first) and the same output: graph <epic> --full --json at 20,000 issues goes from 1,375 ms to 294 ms (4.7x), the same as the non-full graph. All behavior tests of cmd/sisyphus pass with the prototype.

The loop runs for every graph drawing, but without --full the drawn set is small, so the cost shows only with --full on a large connected component.

## Acceptance criteria

<Each item is one condition that must be true before the issue is complete.>

- [ ] <Condition>

## Out of scope

<List related work that this issue does not include. Create a separate issue for each item that someone must do later.>

## Notes

2026-10-20, from the engineer: priority raised to high. The prototype is validated (1,375 ms to 294 ms at 20,000 issues, all behavior tests pass), the effort is small, and it removes the last quadratic path of [[performance/large-repo-performance]]. Per the epic, land it together with a benchmark for graph --full at two sizes (the gap that [[performance/add-performance-benchmarks]] lists), so the gain cannot regress. No overlap with the hephaestus workstreams: the change is in cmd/sisyphus/graph.go only.

## Resolution

Done: buildGraph buckets the edges by their source and keeps the ready issues in a min-heap (nameHeap), so placement no longer scans every edge for every node and no longer sorts the ready list each iteration. The placement order is unchanged (the heap pops the smallest name, as the sorted list did); all graph tests pass, and BenchmarkGraphFull covers the path. The performance engineer validated 1375 ms to 294 ms at 20k issues on the prototype of the same change.

