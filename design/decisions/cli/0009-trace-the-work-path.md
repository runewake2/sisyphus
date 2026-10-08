# 0009: Trace the work path from start-now issues to the focus

- Status: Accepted
- Date: 2026-10-08
- Version: [[0.0.28]]

## Context

From the human, 2026-10-08: mark the ready-to-start work in a noticeable way, and trace the path
from it to the focus issue, so a reader can tell what to work on now and what later.

## Decision

The work path of `<name>` is what it waits on, all the way down. An issue waits on its
dependencies and its sub-issues that are not closed, which is the rule that already decides
whether an issue is available. Followed from `<name>` through issues that are not closed, the
chains end at available issues, which are the issues to start now. The arrows of the work path
have heavy lines, and a legend line `start now: …` names those issues. With color, the work
path and the borders of the start-now issues are bold magenta (`ESC[1;95m`), drawn last so they
win a shared line. The text in a start-now box keeps its state color.

## Alternatives considered

- **Only color.** The work path must also show in piped output, a file, or a terminal without
  color.
- **A mark in the box, such as `▶`.** [[0002-mark-focus-and-leaves-with-box-styles]] moved the marks
  out of the labels. Heavy lines mark the path without a label.

## Consequences

- A canvas cell holds the set of its directions that are heavy, and a table of every mix of
  light and heavy lines, made from the Unicode character names, draws it. That table also
  corrects the join where an arrow leaves an available issue's heavy box: it is `┠` (heavy
  border, light arrow), not `┝`.
- On an epic, nearly every arrow is on the work path, because the epic waits on each open
  sub-issue. The `start now` line is then what tells where to begin.
