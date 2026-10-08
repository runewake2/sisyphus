# 0007: Color the focus issue and its arrows

- Status: Accepted
- Date: 2026-10-08
- Version: [[0.0.26]]

## Context

From the human, 2026-10-08, as an experiment: in the terminal drawing, use bold colors for the
issue given as the argument and for the arrows attached directly to it, and a dimmer color for
everything else.

## Decision

`<name>`'s box and each arrow that leaves or enters it, with its arrowhead, are bold bright cyan
(`ESC[1;96m`). All other cells are bright black (`ESC[90m`), the gray that terminals show most
consistently. A canvas cell is bright if any bright drawing touches it, so where an arrow shares
a line with the arrows of other issues, only the cells that lead to or from `<name>` are bright.
`--color auto|always|never` decides; `auto` means color only on a terminal, and not when
`NO_COLOR` is set or `TERM` is `dumb`. The legend, `--json`, and `--mermaid` have no color.

## Alternatives considered

- **Faint text (`ESC[2m`) for the rest.** Some terminals show faint text as normal text.
- **Color always, for a reader to strip.** Escape codes in a pipe or a file break other tools.

## Consequences

- The MCP tool and tests write to a buffer, so `auto` gives them no color.
- Each colored line ends with a reset, so a line that is copied alone does not color what follows.
