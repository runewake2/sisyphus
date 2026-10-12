# 0007: The watcher sends the notices of a close

- Status: Accepted
- Date: 2026-10-12
- Version: [[unreleased-watch-notices]]

## Context

Sixteen roles listen for `bookmark-ready`, and `technical-docs` listens for `issue-closed`. Only the
prompt of the implementer sent `bookmark-ready`, so the flow depended on a model that remembers a
step, and nothing sent `issue-closed`. The human asked for both to come from code.

## Decision

- `hephaestus watch` reads the issues (`sisyphus list --state all`) and the heads of their bookmarks,
  and compares them with what it saw before, in `watch.json` in the shared hephaestus directory.
- When an issue closes, it broadcasts `bookmark-ready` (when the bookmark has work that the base does
  not have), and then `issue-closed`. When the bookmark of a closed issue moves later, for example
  for the fixes of a review, it broadcasts `bookmark-ready` again, for the new head.
- The id of a notice comes from the issue (and the head), so a repeated send makes no second message.
  The state changes only after the notices are sent, so a failure loses nothing.
- The first run records what it sees, and sends nothing: the old closes of the repo are history.
- The implementer sends no broadcast. It closes its issue.

## Alternatives considered

- **Notices from sisyphus.** sisyphus does not know agents or messages
  ([[0029-sisyphus-does-not-manage-agents]]).
- **A hook in the version control.** A jj or git hook runs only in the workspace where it is set, and
  not for each close of an issue.
- **Announce the whole history on the first run.** Every old close would start the reviewers.

## Consequences

- One `hephaestus watch` must run for the repo; [[delegation-runner]] will start it with the workers.
- A close that happens while the watcher does not run is announced at its next run.
