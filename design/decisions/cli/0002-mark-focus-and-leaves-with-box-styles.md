# 0002: Mark the focus issue and leaves with box styles

- Status: Accepted
- Date: 2026-10-08
- Version: [[0.0.21]]

## Context

The graph marked `<name>` with 📍 and each leaf with 🍃 in its box label. From the human,
2026-10-08: use different line styles for the boxes instead of emoji.

## Decision

An issue box has a double border (`╔═╗`) for `<name>`, a heavy border (`┏━┓`) for a leaf, and a light
border (`┌─┐`) otherwise. A directory box has a rounded light border (`╭─╮`), so it is not read as
an issue. No box is both double and heavy, so a focus issue that is a leaf has a double box, and
the legend adds `(a leaf)`. In `--mermaid` output, the classes `focus` and `leaf` give thicker
borders, and labels have no marks.

## Alternatives considered

- **Keep the emoji.** The human asked for line styles.
- **A fourth style for a focus leaf.** Unicode box drawing has no border that is both double and
  heavy. A dashed border would look like a dotted arrow.

## Consequences

- Labels are plain ASCII, so no wide characters need special width handling in a box label.
- Where an arrow leaves a heavy or double box, the join mixes styles (`┝`, `╟`).
