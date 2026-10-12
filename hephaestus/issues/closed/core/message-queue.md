---
title: "Add a message queue tool agent that relays messages between workers"
state: closed          # open | in-progress | closed. Must match the directory of the file.
resolution: completed  # completed | abandoned. Set this only when state is closed.
priority: high         # critical | high | medium | low
effort: large          # small | medium | large | x-large
tags: [hephaestus]     # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-11
closed: 2026-10-12     # YYYY-MM-DD. Set this only when state is closed.
owner: claude          # The person or agent working on the issue. Cleared when the issue returns to open.
approver:              # The person or agent who accepts the issue when it closes.
bookmark: samw/ai/ux-fixes # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: [ux-fixes] # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
metadata: {}           # Optional. Arbitrary key-value notes, for example an AI agent session id to resume work with context: {session-id: "abc123"}.
deferred-from:         # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent: "[[core/delegate-issues-to-agents]]" # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
depends-on: ["[[core/delegation-model]]"] # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# Add a message queue tool agent that relays messages between workers

## Summary

Add the message queue: a tool agent that relays messages from any hephaestus worker to another, and public broadcasts to every worker that listens, across AI and tool agents and across workspaces. Tool agents get the context of their tasks as messages. [[agent-delegation#Messages between workers]] describes it.

## Context

Requested by a human on 2026-10-11: agents must be able to send information to each other, including across agents, and tool agents get their context as a message that other agents can broadcast. The message queue is itself a tool agent: it uses no model. A message is information, not a decision; agents that must agree use a referendum ([[referendum-voting]]). Open questions: Q11 (where messages are kept), Q12 (how they travel), Q13 (addressing), and Q14 (delivery).

## Acceptance criteria

- [ ] A decision record answers Q11-Q14.
- [ ] Any worker (a human, an AI agent, or a tool agent) can send a message to one worker, to a role, or as a broadcast, and the receivers get it.
- [ ] A tool agent can listen for broadcasts of one kind, and starts its task from the message.
- [ ] The message queue is a service that the workers connect to, and keeps every message in an append-only log.
- [ ] Each message reaches each receiver at least once and in the order that its sender sent it, across the workspaces of one repo, and a receiver can ignore a copy by its id.
- [ ] Tests cover direct messages, role messages, broadcasts, and a copy of a message.

## Out of scope

- Messages between machines.
- Decisions: they are referendums.

## Notes

2026-10-11, from the human: messages are kept in an append-only log (Q11), are passed through a service (Q12), and have at-least-once delivery (Q14). Q13 and the order of messages are still open in [[agent-delegation]].

2026-10-12, from the human: addressing is by owner name, by role, or broadcast, as proposed (Q13), and the messages of one sender are ordered (Q14).

## Resolution

Completed in [[0.0.112]]. `hephaestus messages serve` is a local service with a token and an append-only log in the shared hephaestus directory of the repo; `send`, `receive`, `ack`, `log`, and `listen` address one worker, one or all workers of a role, or a broadcast filtered by kind, with at-least-once, ordered delivery. Tool agents get their tasks with `listen`. See [[0001-a-local-service-with-an-append-only-log]].

