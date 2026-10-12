# 0002: One tracker and one design log for both projects

- Status: Accepted
- Date: 2026-10-12
- Version: [[unreleased-one-tracker-with-sisyphus]]

## Context

hephaestus started with an issues tracker of its own (`hephaestus/issues/`, made with
`sisyphus init --nested`) and design documents of its own (`hephaestus/design/`), as
[[0029-sisyphus-does-not-manage-agents]] said. A human decided that hephaestus and sisyphus share
one directory, with their issues and design separated only by subdirectories, not isolated in
trackers and design logs of their own ([[one-tracker-with-sisyphus]]).

## Decision

- The issues of hephaestus are in the tracker of the repo, in `issues/<state>/hephaestus/` (for
  example `hephaestus/core/message-queue`). Links in the frontmatter use these full names.
- Its design documents are in `design/hephaestus/`, and its decision records in
  `design/decisions/hephaestus/<component>/`, with the numbers they had.
- Its code stays a Go module of its own in `hephaestus/` (Q7 of [[agent-delegation]]), and its
  role files stay in `hephaestus/roles/` (Q2).
- The boundary of [[0029-sisyphus-does-not-manage-agents]] does not change: sisyphus does not manage
  agents. Only the layout of the issues and the design changes.

## Alternatives considered

- **Keep the nested tracker.** `list`, `graph`, and `lint` would show only half of the work, and an
  issue of one project could not have a parent or a dependency in the other.

## Consequences

- One `sisyphus lint`, `list`, and `graph` covers both projects; the epic
  [[delegate-issues-to-agents]] is in the same graph as the sisyphus issues.
- The checks of the role files lint one tracker.
- The fix of [[nested-tracker-links]] stays: other repos can still have nested trackers.
