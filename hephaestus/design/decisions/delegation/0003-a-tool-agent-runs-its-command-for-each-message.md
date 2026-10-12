# 0003: A tool agent runs its command for each message

- Status: Accepted
- Date: 2026-10-12
- Version: [[unreleased-tool-agents]]

## Context

The human answered the questions about tool agents in [[agent-delegation]]: a tool agent gets the
context of its task as a message (Q8), reports its result as proposed (Q8: exit code 0 closes the
issue, another exit code stops it), can be a referendum reviewer (Q9), and the merge queue is a
tool agent (Q10). This record holds how hephaestus does it ([[tool-agents]]).

## Decision

- `hephaestus agent run <role>` is a worker of a tool role, with the owner name `<role>-1` unless
  `--as` gives another. It gets the messages to it and to its role, and the broadcasts of the kinds
  in `listens` of its role file.
- For each message with an issue, the worker claims the issue with `sisyphus start` as its owner. A
  refused claim (another owner or workspace has the issue) is not run. It then runs the command of
  the role with the shell of the platform, in the workspace of the task (`workspace` in the data of
  the message, or `-C`), with the message as JSON on standard input.
- Exit code 0 closes the issue; its Resolution has the command and the end of its output. Another
  exit code appends the output to `Notes` and stops the issue (back to open): tagged `needs-human`,
  or, when `on-failure` names an AI role, delegated to it with `metadata delegate=<role>` and a
  `task` message to the role.
- The worker sends a `result` message to the sender, with the status, the exit code, and the end of
  the output, and acknowledges the message. A copy of a handled message is acknowledged again and
  not run.
- `hephaestus agent review <role>` gives the ballot of a tool reviewer: `no` for an exit code other
  than 0, `modify` with the changed files and the diff when the command changed the working copy,
  and `yes` otherwise. A tool reviewer never abstains.
- A worker of an AI role is refused: hephaestus does not run models yet.

## Alternatives considered

- **Claim the issue after the command.** Two workers could then run the same task.
- **Retry a failed command.** A tool does not decide: a failure goes to a human or to an AI role.
- **Commit the changes of the command.** Work goes to the merge queue through bookmarks, and the
  rules of the workspaces and bookmarks of a worker belong to the runner (the scheduler daemon), not
  to one command.

## Consequences

- A delegated task that a program can do costs no tokens.
- The merge queue (Q10) is a tool role whose command does the merge; [[merge-queue-agent]] writes it.
- The changes of a command stay in the working copy of its workspace until something commits them.
