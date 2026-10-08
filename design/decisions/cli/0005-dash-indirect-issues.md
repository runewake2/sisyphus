# 0005: Draw indirectly linked issues with dashed boxes

- Status: Accepted
- Date: 2026-10-08
- Version: [[0.0.24]]

## Context

`sisyphus graph --full` also draws the issues that are linked to `<name>` but are not below it
or on its path above it, such as a sibling under the same parent. They looked the same as the
issues on the path. From the human, 2026-10-08: give them another outline.

## Decision

With `--full`, such an issue has a dashed box: light dashed (`┌╌┐`, `╎`) normally, and heavy dashed
(`┏╍┓`, `╏`) if it is available. The focus issue is never indirect. In `--json`, the node has
`indirect: true`. In `--mermaid`, the class `indirect` gives a dashed border.

## Alternatives considered

- **The three-dash lines `┄` and `┆`.** Dependency arrows use them, so a box would look like arrows.
- **Faint or colored text from ANSI codes.** It is lost when the output is piped or pasted.
- **One dashed style for all indirect issues.** It would hide which of them are available.

## Consequences

- Unicode has no dashed corners or joins, so a dashed box has solid corners, and the border cell
  where an arrow leaves it is a solid `├` or `┝`.
- The legend lists each dashed style that the drawing uses.
