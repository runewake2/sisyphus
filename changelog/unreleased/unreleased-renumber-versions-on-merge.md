# Unreleased: Give versions only when work lands, and add the hephaestus versioning tools

- Date: 2026-10-12
- Bookmark: samw/ai/ux-fixes
- Components: hephaestus, repo
- Changelog: Give versions only when the merge queue lands work, with hephaestus version assign and decisions renumber.

## Summary

A line of work no longer picks a version: it writes an unreleased changelog entry, and the merge
queue gives the version when the work lands. The new hephaestus module has the two tools for this:
`hephaestus version assign` and `hephaestus decisions renumber`.

## Changes

- New Go module `hephaestus/` (`github.com/runewake2/sisyphus/hephaestus`) with the command
  `hephaestus`. It uses sisyphus only through its command, and jj or git through theirs.
- `hephaestus version assign [<entry>...]` gives the next SHAME versions to the entries in
  `changelog/unreleased/`, in the order of the merge, moves each to `changelog/<version>.md`, adds
  the `CHANGELOG.md` entry, sets `VERSION`, and changes every wikilink to the entry (not in code).
  `hephaestus version list` lists the entries.
- `hephaestus decisions renumber --onto <base>` gives the next free number to each decision record
  that has the number of another record in its directory and that the base does not have, with
  its heading and every link to it.
- Both have `--dry-run` and `--json`. Tests use git repos, and a jj repo when jj is installed.
- CONTRIBUTING.md and AGENTS.md describe the new rule: unreleased entries, no change of `VERSION`
  or `CHANGELOG.md` in a line of work, and the tools that the merge queue runs.

## Decisions

- [[0001-versions-come-from-the-merge-queue]]

## Issues

- Closed [[renumber-versions-on-merge]].
