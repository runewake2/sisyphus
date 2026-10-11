---
title: "Draw sisyphus graph as a real dependency graph"
state: closed          # open | in-progress | closed. Must match the directory of the file.
resolution: completed  # completed | abandoned. Set this only when state is closed.
priority: medium       # critical | high | medium | low
effort: medium         # small | medium | large | x-large
tags: [feature, cli, mcp] # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-08
closed: 2026-10-08     # YYYY-MM-DD. Set this only when state is closed.
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

# Draw sisyphus graph as a real dependency graph

## Summary

`sisyphus graph` printed only the parent tree, with `depends-on` as text next to each issue.
Draw every issue linked by `parent` or `depends-on` as boxes and arrows instead.

## Context

`sisyphus graph` walked up to the root ancestor and printed the parent tree. Each issue's
`depends-on` was listed as text. Dependencies form a graph, not a tree: an issue can depend on
issues outside its family, and several issues can depend on one. [[graph-view-command]] built the
tree view.

## Acceptance criteria

- [x] Arrows all point one way: from a parent to its sub-issue, and from a dependency to the issue
  that depends on it.
- [x] `sisyphus graph <name>` draws `<name>`, everything below it, and the path above it, as boxes
  with name and state, and notes how many other linked issues are not drawn.
- [x] `sisyphus graph <name> --full` draws every issue reachable from `<name>` through `parent`
  and `depends-on`, in both directions.
- [x] Parent edges and depends-on edges look different, and a legend says which is which.
- [x] `<name>` is marked.
- [x] Leaves are marked: issues that are not closed and have no open dependency or sub-issue.
- [x] A link to a missing issue is drawn with state `?`, and a parent cycle does not hang.
- [x] `--json` prints nodes and edges; `--mermaid` prints the Mermaid source.
- [x] `sisyphus_graph` in `sisyphus-mcp` supports `--full`, `--json`, and `--mermaid`.
- [x] [[sisyphus-graph]] documents the new output.

## Out of scope

- An ASCII-only drawing. mermaid-ascii draws dotted edges as solid in ASCII mode, so the two kinds
  of edge would look the same.
- Fitting the drawing to the terminal width.

## Notes

- 2026-10-08: Drawn with `github.com/AlexanderGrooff/mermaid-ascii` (MIT), using only its
  `pkg/graph` and `pkg/diagram` packages. It has no tagged release, so `go.mod` pins a
  pseudo-version.
- Left to right (`graph LR`) keeps typical families under 80 columns. Top down put every sibling
  in one row and ran past 200.
- The library places each node one column after the first node that points to it, and visits nodes
  in declaration order. Nodes are declared in link order (each after every issue pointing to it).
  An epic still claims all its sub-issues for one column, even when they depend on each other.
- Where two edges share a path, the library can draw one over the other. The docs point to
  `--json` and `--mermaid` for the exact links.

## Resolution

Completed: `sisyphus graph` draws an issue, everything below it, and the path above it, or with
`--full` the whole connected graph, with `--json` and `--mermaid`.
See [[0.0.16]].
