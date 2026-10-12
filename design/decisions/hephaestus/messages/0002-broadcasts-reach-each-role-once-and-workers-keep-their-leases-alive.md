# 0002: Broadcasts reach each role once, and workers keep their leases alive

- Status: Accepted
- Date: 2026-10-12
- Version: [[unreleased-queue-roles-broadcasts-keepalive]]

## Context

A review of hephaestus found three defects in the message queue
([[0001-a-local-service-with-an-append-only-log]]):

- A role of a family, for example `formatter/golang`, could not receive or be sent a message: a
  role name could not have a `/`.
- A worker claimed the issue of every message. The implementer closes its issue, and
  `issue-closed` is about a closed issue, so every worker that listens for `bookmark-ready` or
  `issue-closed` failed. Several listeners also cannot all claim one issue. And a broadcast reached
  only the workers that had registered before it was sent.
- A worker did not renew its lease, so a task longer than the lease (60 seconds) went to a second
  worker. In a test, one message to two workers of a role ran four times.

The human asked: support roles of families in messages, fix the order of broadcasts, and give the
workers a keep-alive that tells the scheduler that they still work, so that no work is duplicated.

## Decision

- **Roles of families.** A role in a message or a receive can be a path, for example
  `formatter/golang`. A message to every worker of a role (`--all`) whose role is a family, for
  example `formatter`, reaches the workers of each role in it. A message to one worker of a family
  reaches nobody: one worker must be of one role. `hephaestus messages send` sends it all the same,
  but warns, with the roles of the family and the two ways to reach them (from the human,
  2026-10-12).
- **A broadcast reaches each listening role once.** The lease and the acknowledgement of a
  broadcast are for the role of the worker, not for the worker. One worker of the role handles it.
  A broadcast reaches a role when a worker of the role registered before it, so a worker that
  starts after the send still gets it. A listener with no role, for example a person, gets it as
  before.
- **Notices change no issue state.** A broadcast, and a message to every worker of a role, are
  notices. A worker does not claim, close, or reopen their issue. A tool worker records its command,
  exit code, and output in the Notes of the issue, and on a failure tags the issue `needs-human`.
  The prompt of an AI worker tells it to record its result in the Notes. A message to one worker, or
  to one worker of a role, stays a task: the worker claims its issue.
- **The order of the implementer.** It broadcasts `bookmark-ready` before it closes its issue.
- **Keep-alive.** `POST /renew` gives each lease that a worker holds a new full lease, and tells
  which leases the worker lost. `hephaestus agent run` and `messages listen` renew at once, and then
  three times in each lease (at most every 15 seconds), while they handle a message. The service
  keeps the last time that each worker received, acknowledged, or renewed; `hephaestus messages
  workers` shows it with the messages that each worker holds. The keep-alive is not in the log:
  after a restart, it starts empty, and the leases end as before.

## Alternatives considered

- **Every broadcast to every worker of a role.** Two workers of `repo-check` would run the same
  checks.
- **Keep the broadcast for the workers that listen at the send.** A worker that restarts would miss
  work.
- **A longer lease.** A model task has no upper limit in time, and a dead worker would hold its
  message longer.
- **The keep-alive in the log.** Each worker would add a line every 15 seconds to a log that is
  never compacted.

## Consequences

- One worker of each listening role handles each `bookmark-ready` and `issue-closed`.
- A long task stays with its worker while the worker lives. When the worker stops, its lease ends,
  and another worker gets the message.
- [[delegation-runner]] can read `hephaestus messages workers` to find workers that stopped.
