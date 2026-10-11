# 0018: Reopen a mistaken close only with an explicit flag

- Status: Accepted
- Date: 2026-10-11
- Version: [[0.0.57]]

## Context

sisyphus refuses to reopen a closed issue: work that needs more after a real close gets a new issue
that links to the old one ([[CONTRIBUTING#Issue lifecycle]]). But a close that was a mistake, for
example an agent that closed the wrong issue, could only be fixed by moving the file by hand and
clearing `resolution` and `closed`. A new issue would lose the history, the links, and the
sub-issues of the old one ([[undo-mistaken-close]]).

## Decision

`sisyphus update <name> open|in-progress --reopen` moves a closed issue back. It clears
`resolution` and `closed`, keeps everything else, and adds a dated line to `Notes` with the old
resolution and closed date, so that the history shows the reopen. The flag fails for an issue that is
not closed. Without the flag, a reopen still fails, and the error now names the flag for the case of
a mistake. `sisyphus_update` has a `reopen` argument.

## Alternatives considered

- **Allow a reopen without a flag.** It would remove the rule against reopening for real closes.
- **A separate `sisyphus undo` command.** sisyphus keeps no log of past commands, so a general undo
  cannot know what to undo; a flag on `update` states exactly what changes.

## Consequences

- A mistaken close is a one-command fix that keeps the issue's name, links, and sub-issues.
- The rule for real closes does not change: create a new issue.
