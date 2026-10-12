# 0006: A task has a time limit and a token limit

- Status: Accepted
- Date: 2026-10-12
- Version: [[unreleased-task-limits]]

## Context

A review of hephaestus found that a task of a worker had no limit. A hung tool or a model that does
not stop held its message forever, because the keep-alive renews the lease while the task runs. A
model run could also spend tokens with no limit. The human asked for limits on a task of a worker.

## Decision

- **A time limit for each task.** `agent run --task-timeout`, else `timeout` in the role file, else
  1 hour. At the limit, the command stops with every program that it started (its process group),
  and the task is an error: the issue is stopped for a human, as for any failed task.
- **A token limit for each task of an AI role.** `agent run --max-tokens`, else `max-tokens` in the
  role file, else no limit. hephaestus reads the output of the runner as it comes. A runner that
  prints JSON lines in the form of `claude -p --output-format stream-json --verbose` reports the
  usage of each message of the model: its input, output, and cache tokens, once for each message id.
  When the sum passes the limit, the run stops. A runner that reports no usage cannot be limited, and
  the worker warns.
- **The record of each run.** The usage (tokens, the cost and the turns of the final `result` line)
  and the time go into the result of the task, and a line goes into the Notes of its issue. The text
  of the `result` line is the output of the model, for example a referendum ballot.
- Referendum reviewers keep the `max-tokens` of their role, and the referendum keeps its own
  `--timeout`.

## Alternatives considered

- **A cost limit in money.** The runner reports the cost only at the end of the run, so it cannot
  stop a run in time. Tokens are reported as the run goes.
- **A limit option of the runner.** Each runner has other options, or none; the meter works for any
  runner that reports usage.

## Consequences

- A hung or runaway task holds its message for at most its time limit.
- `economist/tokens` and `operations-auditor` read the tokens and cost of each run in the Notes.
- To use a token limit, the runner of an AI role must print stream JSON.
