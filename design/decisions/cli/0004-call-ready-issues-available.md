# 0004: Call a ready-to-start issue "available"

- Status: Accepted
- Date: 2026-10-08
- Version: [[0.0.23]]

## Context

`sisyphus graph` called an issue that is not closed and waits on nothing a "leaf". The name comes
from the sub-issue tree. Since [[0003-place-issues-by-their-longest-chain]], the drawing follows
work order, so such an issue sits where its work starts, often inside the drawing, not at an edge.
From the human, 2026-10-08: use "available" instead.

## Decision

The term is "available" everywhere: the legend (`┏━┓ available (ready to start)`, and
`(available)` after an available focus issue), the `--json` node field `available`, the Mermaid
class `available`, the help, the `sisyphus_graph` MCP description, and [[sisyphus-graph]]. The rule
that decides it does not change.

## Alternatives considered

- **Keep `leaf` in `--json` for compatibility.** Two names for one property would stay forever. As of
  2026-10-08, no reader of the field exists on this host.

## Consequences

- A `--json` reader that tests `leaf` must test `available`.
- [[0002-mark-focus-and-leaves-with-box-styles]] still holds; only the name of the heavy box changes.
