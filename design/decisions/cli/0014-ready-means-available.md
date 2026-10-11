# 0014: list --ready uses the "available" rule of graph

- Status: Accepted
- Date: 2026-10-11
- Version: [[0.0.44]]

## Context

`sisyphus list --blocked` shows the issues that wait on an open dependency, but nothing answered the
main question of a task manager: what can I start now? `sisyphus graph` already computes this set
and calls it "available" ([[0004-call-ready-issues-available]]), but only inside one drawing
([[list-ready-issues]]).

## Decision

`sisyphus list --ready` and `sisyphus search --ready` show the issues that are not closed and have
no `depends-on` issue and no sub-issue that is not closed. This is the rule of `graph`, so an epic
with open sub-issues is not ready. A dependency that does not exist does not block, as in
`--blocked`. `--ready` and `--blocked` cannot be used together. The state filter still applies, so
the default lists ready open and in-progress issues, and `--state open --ready` lists only the
issues that nobody has started. The MCP tools have a `ready` argument.

## Alternatives considered

- **Count only dependencies, the exact inverse of `--blocked`.** Simpler, but an epic would be
  "ready" while all its work is still open, and `list` and `graph` would disagree.
- **Leave out in-progress issues by default.** "Ready" then means "not started", which the state
  filter can already say; one rule for both commands is clearer.

## Consequences

- A test checks that `list --ready` and the available nodes of `graph --json --full` agree.
- The flag name is "ready" and the graph term is "available": [[sisyphus-list]] links the two.
