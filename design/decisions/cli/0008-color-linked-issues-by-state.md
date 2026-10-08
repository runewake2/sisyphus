# 0008: Color the issues linked to the focus by state

- Status: Accepted
- Date: 2026-10-08
- Version: [[0.0.27]]

## Context

[[0007-color-the-focus-and-its-arrows]] made `<name>` and its arrows bright and everything else
gray. From the human, 2026-10-08: color the issues linked directly to it as well, blue for open,
yellow for in-progress, green for completed, and red for abandoned.

## Decision

Each issue at the other end of an arrow that leaves or enters `<name>` is bold in the color of its
state: blue (`ESC[1;94m`) for open, yellow (`ESC[1;93m`) for in-progress, green (`ESC[1;92m`) for
closed and completed, and red (`ESC[1;91m`) for closed and abandoned. A closed issue with no
resolution is completed, the template's default. A missing issue is bold white. A linked issue
in either direction counts: a parent, a dependency, a sub-issue, or a dependent. A legend line
names each state color that the drawing uses, in its color. `--json` nodes gain `resolution`.

## Alternatives considered

- **Only the issues that `<name>` depends on.** It leaves out its sub-issues and dependents,
  whose arrows are already bright.

## Consequences

- A canvas cell holds a color, not a strong flag. A drawing in a color sets it, and a gray drawing
  leaves it, so the last colored drawing in a cell wins: an arrow's join on a linked issue's
  border is cyan.
