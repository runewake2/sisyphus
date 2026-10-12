# Unreleased: Time and token limits for each task

- Date: 2026-10-12
- Bookmark: samw/ai/ux-fixes
- Components: hephaestus
- Changelog: Each task of a hephaestus worker stops at a time limit (1 hour by default) and, for an AI role, at a token limit; the tokens, cost, and time of each model run go into the Notes of its issue.

## Summary

A hung tool or a runaway model held its message forever. Now each task has limits, and each model
run leaves a record of what it used ([[0006-a-task-has-a-time-limit-and-a-token-limit]]).

## Changes

- `hephaestus agent run --task-timeout <time>`, and `timeout` in a role file: the longest time of one
  task, 1 hour by default. Then the task stops, with every program that it started, and its issue
  goes to a human.
- `hephaestus agent run --max-tokens <n>`, and `max-tokens` in the role file of an AI role: the most
  tokens of one task. hephaestus counts the usage that the runner prints as JSON lines (for example
  `claude -p --output-format stream-json --verbose`), and stops the run when it passes the limit. A
  runner that reports no usage gets a warning.
- The tokens, the cost, the turns, and the time of each model run go into the result of the task and
  a line in the Notes of its issue.
- A referendum reviewer keeps the token limit of its role, and its ballot is read from the final
  result of a stream JSON run.

## Decisions

- [[0006-a-task-has-a-time-limit-and-a-token-limit]]

## Issues

- Closed [[task-limits]].
