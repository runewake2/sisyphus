# 0006: Show the issue state below the name

- Status: Accepted
- Date: 2026-10-08
- Version: [[0.0.25]]

## Context

A graph box showed `<file name> [<state>]` on one line. Since
[[0003-place-issues-by-their-longest-chain]], each dependency step adds a column, so the width of a
box label adds up across a graph, and `[in-progress]` alone is 14 columns. From the human,
2026-10-08: put the state on a second line below the name.

## Decision

An issue box is four lines high: the border, the file name, the state, and the border. Each text
line is centered. The state has no brackets, because it no longer follows the name. Arrows leave
and enter the box on the line of the name. `--mermaid` labels stay `<file name> [<state>]`,
because a Mermaid renderer wraps a label itself.

## Alternatives considered

- **Abbreviate the state, for example `wip`.** It saves less, and a reader must learn the
  abbreviations.

## Consequences

- A box is as wide as the longer of its name and its state, so most boxes get narrower. A graph
  gets about one line taller for each row of boxes.
