# Unreleased: Roles of families in messages, broadcasts once to each role, and keep-alive

- Date: 2026-10-12
- Bookmark: samw/ai/ux-fixes
- Components: hephaestus
- Changelog: The hephaestus message queue accepts roles of families (`formatter/golang`), delivers a broadcast once to each listening role as a notice that does not claim its issue, and workers renew their leases while they work, so that no task runs twice.

## Summary

Three defects of the message queue made the government of hephaestus fail or do work twice. The
queue now accepts roles of families, a broadcast reaches each listening role once, and a worker
keeps its lease alive while it works
([[0002-broadcasts-reach-each-role-once-and-workers-keep-their-leases-alive]]).

## Changes

- A role in a message or a receive can be a path, for example `formatter/golang`. Formerly,
  `hephaestus agent run formatter/golang` failed with `role "formatter/golang" is not a role name`.
  A message to all workers of a family, for example `--role formatter --all`, reaches each role in it.
- A broadcast reaches one worker of each listening role, also a worker that starts after the send.
  Formerly, it went to every worker that listened at the send, and to no other.
- A broadcast, and a message to every worker of a role, are notices: a worker records its result in
  the Notes of the issue, and on a failure tags it `needs-human`, but does not claim, close, or
  reopen it. Formerly, a worker claimed the issue, so `repo-check`, the formatters, `code-reviewer`,
  `test-engineer`, and `technical-docs` failed on each broadcast about a closed issue.
- The implementer broadcasts `bookmark-ready` before it closes its issue.
- Keep-alive: `hephaestus agent run` and `messages listen` renew the lease of a message while they
  handle it. Formerly, a task longer than the lease ran again on a second worker.
- `hephaestus messages renew <id>...` renews leases by hand, and `hephaestus messages workers` lists
  each worker, when it last spoke to the queue, and the messages that it holds.

## Decisions

- [[0002-broadcasts-reach-each-role-once-and-workers-keep-their-leases-alive]]

## Issues

- Closed [[queue-roles-broadcasts-keepalive]].
