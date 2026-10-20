# Unreleased: doctor and kick for stuck tasks

- Date: 2026-10-20
- Bookmark: samw/ai/doctor-and-kick
- Components: hephaestus
- Changelog: `hephaestus doctor <issue>` diagnoses the task of one issue — whether a worker works on it, lost it, or stopped on it — and `hephaestus kick <issue>` makes a stuck task run again: it reopens the issue, removes `needs-human`, and sends the task back to its role.

## Summary

A task gets stuck invisibly: a worker crashes or runs out of API keys after it acknowledged its
message, so the queue will not give the task again, and the issue stays in progress with an owner
that is gone. The doctor names the problem and its fix; the kick applies the fix.

## Changes

- `hephaestus doctor <issue>` reads the issue, the workers and pending deliveries of the message
  queue, the message log, the task workspace, and the bookmark, and reports the facts and a finding
  with its fix for each problem:
  - an owner that stopped, that the queue does not know, or that is live without the task;
  - a task stopped for a human (`needs-human`), with the Notes to read;
  - a delegated issue with no task message, and a waiting task with no live worker of its role;
  - a lease held by a stopped worker, a referendum past its deadline, a bookmark that does not
    exist, and a queue that does not run.
  It changes nothing. Exit code 2 when there are findings, so a script can watch for stuck tasks.
- `hephaestus kick <issue>` makes a stuck task run again: it reopens an in-progress issue (which
  drops the owner that stopped), removes `needs-human`, keeps the delegation, and sends a new task
  message to the role (`--role`, `--reason`, `--as`). When a task message already waits, it sends
  no second one: the reopened issue lets a worker take it. The task workspace stays, and the kicked
  task continues there. The kick goes into the Notes of the issue.
- A task that a live worker holds is slow, not stuck: the kick refuses it without `--force`. The
  kick never starts a worker, and hephaestus still never starts a model by itself.
- `internal/hephaestus/doctor` holds the diagnosis and the kick plan, tested with fakes, as the
  status package is.

## Issues

- Closed [[doctor-and-kick]].
