# 0028: Join graph links through filtered issues only where they stay true

- Status: Accepted
- Date: 2026-10-11
- Version: [[0.0.106]]

## Context

`sisyphus graph --state` leaves out the issues in other states ([[graph-state-filter]]). A link
from or to a left-out issue then has no end. Without it, an open task below a closed sub-epic would
float with no arrow to its epic, but a simple join of every chain would draw false links: an issue
whose dependency is closed does not wait on anything before it.

## Decision

- A chain of parent links through left-out issues becomes one parent arrow: the last issue is
  still below the first.
- A chain of depends-on links through left-out issues becomes one depends-on arrow only when no
  left-out issue on it is closed: the last issue still waits on the first. A closed issue ends the
  wait, so a chain through it is not drawn.
- A chain of mixed kinds is not drawn: it is neither a sub-issue nor a wait.
- The JSON of a joined edge has `through`, the left-out issues between its ends.
- `<name>` and missing issues are always drawn. Availability and the work path come from every
  issue, so they do not change with the filter.

## Alternatives considered

- **Drop every link to a left-out issue.** Simple, but parts of an epic float with no arrow.
- **Join every chain.** It draws waits that do not exist.
- **Filter on resolution too, for example to hide only abandoned issues.** `--state` covers the
  request; a resolution filter can come later.

## Consequences

- `--state open,in-progress` shows only the work that remains, with true arrows.
- An open issue that was linked only through a closed dependency is drawn with no arrow to the rest.
