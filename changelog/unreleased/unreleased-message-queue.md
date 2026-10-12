# Unreleased: Add the message queue of hephaestus

- Date: 2026-10-12
- Bookmark: samw/ai/ux-fixes
- Components: hephaestus
- Changelog: Add the hephaestus message queue: a service with an append-only log that relays messages between workers, with at-least-once, ordered delivery.

## Summary

`hephaestus messages` relays messages between workers (humans, AI agents, and tool agents) in every
workspace of a repo: to one worker, to one worker or every worker of a role, or as a broadcast that
listeners filter by kind. Tool agents get the context of a task as a message.

## Changes

- `hephaestus messages serve` runs the service of the repo on 127.0.0.1, with a token, and finds the
  shared hephaestus directory of the repo in `.jj/repo` or the git common directory, so that every
  jj workspace and git worktree uses one queue.
- Every event goes to the append-only log `messages.log`, and the service rebuilds its state from it
  when it starts.
- A message stays pending until its receiver acknowledges it, and comes again after a lease (default
  60 seconds). The messages of one sender keep their order. A send with the id of an earlier message
  adds nothing.
- `send`, `receive`, `ack`, `log`, and `listen`. `listen -- <command>` runs a command for each
  message, with the message as JSON on standard input, and acknowledges it when the command
  succeeds; a failed message comes again after the lease.
- Tests cover direct, role, and broadcast messages, kinds, leases, order, a repeated send, a
  restart, the token, a second worktree, and `listen`.

## Decisions

- [[0001-a-local-service-with-an-append-only-log]]

## Issues

- Closed [[message-queue]].
