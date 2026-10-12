# 0004: AI roles run as Claude Code subagents

- Status: Accepted
- Date: 2026-10-12
- Version: [[0.0.117]]

## Context

The human answered Q4 of [[agent-delegation]]: AI agents start as Claude Code subagents, started
by a human, then hephaestus runs them. The AI roles have role files ([[role-file]]), and a human
asked to implement them, with two more roles: a security researcher and a privacy reviewer.

## Decision

- `hephaestus agent export claude` generates a Claude Code subagent, `.claude/agents/<role>.md`, for
  each AI role, from its role file: the description from the charter, `tools` from
  `allowed-tools`, `model`, the sections, and the rules of every role. The generated files are in
  the repo, and a test fails when they differ from the role files.
- `hephaestus agent run <role>` runs a worker of an AI role like a worker of a tool role: it claims
  the issue of each message, and runs a runner command with the prompt of the task on standard
  input. `{role}` in the runner becomes the role, so that `claude -p --agent {role}` uses the
  subagent.
- The runner comes only from `--runner` or `HEPHAESTUS_AI_RUNNER`. A model costs tokens, so
  hephaestus never starts one by itself.
- The model closes or stops its issue. An issue that it leaves in progress, and a runner that fails,
  go to a human, also when the role names another role in `on-failure`: one model does not hand a
  failure to another.
- A role file is linked as `[[roles/<role>]]`, because its subagent has the same file name.

## Alternatives considered

- **The Claude Agent SDK first.** The human chose subagents first (Q4). The runner command can call
  a program on the SDK later, with no change to hephaestus.
- **A default runner.** A worker that a person starts by mistake would spend tokens.
- **Subagent files written by hand.** They would drift from the role files.

## Consequences

- A person in Claude Code can give a task to any AI role now.
- A worker of an AI role costs the tokens of its model for each message.
- [[delegation-runner]] still has to start the workers on their triggers.
