# 0001: Draw the graph without mermaid-ascii

- Status: Accepted
- Date: 2026-10-08
- Version: [[0.0.21]]

## Context

`sisyphus graph` must draw each issue subdirectory as a box around the issues below it
([[group-graph-nodes-directory-boxes]]). The drawing came from mermaid-ascii, which parses Mermaid
`subgraph` blocks. But its layout places nodes on a grid without regard to subgraphs, and then
draws each box around the area of its members. A test copy of the repo issues in five directories
gave a wrong drawing for 1 of 3 graphs: two issues of one directory sat inside the box of its
sibling, and the sibling's own box was not drawn. A wrong box tells a reader that an issue is in a
directory that does not hold it, so it is worse than no box.

## Decision

sisyphus lays out and draws the terminal graph itself, in `cmd/sisyphus/layout.go`,
`render.go`, and `canvas.go`, and does not depend on mermaid-ascii. Each directory owns a band of
lines that no other directory's issues enter, so a box holds exactly the issues below its directory.
Arrows turn in lanes between the columns, and the lanes are ordered so that two arrows do not
share a stretch of line unless they leave or enter the same issue. `--mermaid` output is still
Mermaid source, with one `subgraph` per directory, for renderers that lay out subgraphs correctly.

## Alternatives considered

- **Patch mermaid-ascii.** A patch that gave each subgraph its own grid rows fixed the membership,
  but arrows still ran along box borders and hid them. It needs a fork to maintain and an upstream
  change that sisyphus does not control.
- **Boxes only in `--mermaid` output.** Small, but the terminal drawing, which is the main view,
  does not get the boxes.

## Consequences

- The drawing changed for every graph, also one with no directories. It is more compact.
- `go-runewidth` is a direct dependency, and mermaid-ascii and its dependencies are gone.
- Layout quality is now this repo's work. The layout is simple: an issue goes in the column after
  the first issue that points to it, and in the row of that issue where the row is free.
