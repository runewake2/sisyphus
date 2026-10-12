# 0001: A local service with an append-only log

- Status: Accepted
- Date: 2026-10-12
- Version: [[unreleased-message-queue]]

## Context

Workers must send information to each other, across AI and tool agents and across the workspaces of
one repo ([[agent-delegation#Messages between workers]]). The human answered Q11-Q14 of
[[agent-delegation]]: messages are kept in an append-only log (Q11), are passed through a service
(Q12), are addressed by owner name, by role (one worker, or all with a flag), or as a broadcast that
listeners filter by kind (Q13), and have at-least-once, ordered delivery (Q14). This record also
holds the choices that the answers left open ([[message-queue]]).

## Decision

- **The service** is `hephaestus messages serve`, one for each repo. It listens on a free port of
  127.0.0.1, and writes its address and a random token to `mq.json` in the hephaestus directory
  that every workspace shares (`hephaestus/` in `.jj/repo` or the git common directory). Every
  request must have the token, so that another user of the machine cannot use it.
- **The log** is `messages.log` in the same directory: one JSON event for each line, a message, a
  worker that registers, or an acknowledgement. Nothing changes or removes a line. The service
  rebuilds its state from the log when it starts.
- **Delivery** is at least once: a message stays pending for its receiver until the receiver
  acknowledges it. A receive leases the message (default 60 seconds); after the lease, it comes
  again. A message to one worker of a role has one lease for the role, so that one worker has it at
  a time, and an acknowledgement by any worker of the role ends it. Each message has an id; a send
  with the id of a message in the log returns that message, so that a sender can repeat a send.
- **Order:** a message is not given while an earlier message of its sender to the same receiver is
  leased and not acknowledged.
- **Queues and subscriptions:** a message to a worker or to one worker of a role waits until it is
  handled, also for a worker that has not started yet. A broadcast, and a message to all workers of
  a role, reach the workers that had registered when it was sent: a worker registers with its first
  receive.
- **Tool agents** run as `hephaestus messages listen --as <worker> --kind <kind> -- <command>`: the
  command gets each message as JSON on its standard input; exit code 0 acknowledges it, and another
  exit code leaves it pending.

## Alternatives considered

- **A Unix socket.** No port and no token, but a socket path has a short length limit, and the
  directory of a repo can be deep.
- **Files in a shared directory, with no service.** The human chose a service (Q12).
- **Remove handled messages.** The log is append-only (Q11); a handled message has an
  acknowledgement event instead.
- **Every broadcast to every worker, also to workers that start later.** A new worker would get
  every old broadcast of the repo.

## Consequences

- Workers in every workspace of one machine share one queue. Messages between machines are out of
  scope.
- A crash of a receiver loses no message: the message comes again after the lease.
- The log grows with every event. Compaction would need a new decision, because the log is
  append-only.
