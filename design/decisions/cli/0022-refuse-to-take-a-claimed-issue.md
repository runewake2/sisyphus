# 0022: Refuse to take a claimed issue

- Status: Accepted
- Date: 2026-10-11
- Version: [[0.0.87]]

## Context

The owner of an issue in progress has claimed it. Before this change, `sisyphus start` and
`update --owner` replaced the owner with no error and no warning
([[start-refuses-claimed-issue]]). An agent that started an issue that another agent had took it
from the other agent, and neither knew it. Agents also repeat a call when they are not sure that it
worked.

## Decision

- A start of an issue in progress, or an `--owner` change of it, by a different owner fails. The
  error names the owner and `--take`.
- `--take` changes the owner, and a warning names the old owner.
- A start by the same owner changes nothing and succeeds, so a repeated call is safe. The owner of a
  start is `--owner`, or else `SISYPHUS_OWNER`, or else no owner: a start with no owner of an issue
  that has an owner fails.
- An issue in progress with no owner can get one without `--take`: nobody can be told about the
  change.
- Only a claim checks the owner. Changes of other fields, a close, and a stop to open do not.

## Alternatives considered

- **Only a warning.** An agent that reads only the exit code, or that does not read the warnings,
  takes the issue anyway.
- **Check closes too.** A human often closes the issue of an agent, and the close moves the file
  to `closed/`, which nobody can miss.
- **A separate `claim` command.** `start` is already the claim, and the documentation tells agents
  to use it.

## Consequences

- Two agents in one working copy cannot both start an issue, unless both run at the same moment
  ([[atomic-issue-writes]]) or they work in different workspaces ([[claims-across-workspaces]]).
- An agent harness must give each agent its own owner name, for example with `SISYPHUS_OWNER`.
