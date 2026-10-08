# 0003: Place issues by their longest chain of links

- Status: Accepted
- Date: 2026-10-08
- Version: [[0.0.22]]

## Context

In [[0.0.21]], an issue went in the column after the first issue that pointed to it, so all
sub-issues of an epic shared one column. From the human, 2026-10-08: the render is messy, with
many parallel lines. Each dependency between two sub-issues in one column had to leave on the
right, run back along a blank line, and enter on the left. And arrows that skipped columns shared
one line along a blank row, where they crossed and joined the dependency lines.

## Decision

An issue goes one column after the farthest issue that points to it, so a chain of dependencies
reads left to right. In each column, an issue takes the mean row of the issues in its directory
that point to it, so an arrow can run straight. An issue that an arrow reaches from two or more
columns back takes a row that is free in every column between, so that arrow is one straight line
along the row. Only an arrow in a cycle leaves on the right and comes back along a blank line.

## Alternatives considered

- **Keep sub-issues in one column, and draw a dependency between neighbors as a short vertical
  arrow.** It helps only when the two issues are next to each other in the column. Others still
  need a line that goes around.
- **Dummy rows for each arrow that skips columns, as in a full layered layout.** It gives the
  same straight lines, but adds ordering work for a gain that the free-row rule already gets.

## Consequences

- A drawing is wider, by one column for each step of the longest dependency chain.
- A drawing is taller where long arrows need free rows: such an epic draws as a staircase.
- Each arrow without a cycle has one turn at most and points to the right.
  TestGraphArrowsTakeNoDetourWithoutACycle checks it.
