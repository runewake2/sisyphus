# 0001: Versions come from the merge queue

- Status: Accepted
- Date: 2026-10-12
- Version: [[0.0.108]]

## Context

Each line of work picked the next SHAME version, and wrote `changelog/<version>.md`, an entry at the
top of `CHANGELOG.md`, and `VERSION`. Parallel lines of work picked the same version, and each merge
had to renumber the files, the entries, and every link to them by hand. A human decided that the
changelogs and versions of a line of work are not numbered until the merge queue gives them a
version ([[merge-queue-agent]]).

## Decision

- A line of work writes an unreleased entry, `changelog/unreleased/unreleased-<topic>.md`, with
  the heading `# Unreleased: <title>` and a line `- Changelog: <one line for CHANGELOG.md>`, and
  links to it as `[[unreleased-<topic>]]`. It never changes `VERSION` or `CHANGELOG.md`.
- The merge queue is the source of truth for versions. When it lands the work, it runs
  `hephaestus version assign`, which gives each entry the next SHAME version in the order of the
  merge, moves it to `changelog/<version>.md`, adds the `CHANGELOG.md` entry, sets `VERSION`, and
  changes every link to the entry.
- Decision record numbers can still collide. The merge queue runs
  `hephaestus decisions renumber --onto main`, which renumbers the record that the base does not
  have.

## Alternatives considered

- **Renumber versions on each merge, as before, but with a tool.** It works, but every line of work
  still writes versions that can be wrong, and every link to them must change at the merge.
- **Unnumbered decision records too.** It would remove the last collision, but the numbers order
  the decisions as people read them. Renumbering a collision is rare and mechanical.

## Consequences

- No two lines of work collide on `VERSION`, `CHANGELOG.md`, or a changelog file name.
- Until the work lands, its version is unknown: an issue links to `[[unreleased-<topic>]]`.
- `sisyphus --version` reports the version of the last landed work.
- A human who lands work without the merge queue must run `hephaestus version assign` too.
