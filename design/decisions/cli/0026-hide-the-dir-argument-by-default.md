# 0026: Hide the dir argument of the MCP tools by default

- Status: Accepted
- Date: 2026-10-11
- Version: [[0.0.99]]

## Context

Every sisyphus-mcp tool had an optional `dir` argument: "The repo's root directory, or a directory
below it. Defaults to sisyphus-mcp's own working directory." Small models copied words of this text
into `dir`, for example `"sisyphus-mcp"`, `"current"`, or `"docs"`, and each such call failed with
`chdir`. In the spike with local models ([[ollama-local-models]]), qwen3:14b sent `dir` in 70 of 76
calls; with `dir` hidden, its pass rate went from 39% to 86% ([[mcp-hide-dir-argument]]).

## Decision

- By default, no tool has a `dir` argument, and a call that gives one is refused by the schema.
  The tools act on the working directory of the server.
- `sisyphus-mcp --dir <path>` sets that directory, for a client that starts the server somewhere
  else.
- `sisyphus-mcp --allow-dir` gives every tool the `dir` argument again, with the description "Leave
  this out. Only to act on another repo ...", and the instructions add "Leave dir out unless the
  user names another repo."

## Alternatives considered

- **Keep `dir`, and change only its description.** A model still sees an optional argument, and
  small models fill optional arguments. The spike measured the gain of hiding it, not of a new text.
- **Remove `dir` completely.** An agent that works in more than one repo needs it, and one server
  for each repo costs a configuration entry for each repo.

## Consequences

- A small model cannot send a wrong `dir`.
- A client configuration that gave `dir` in its calls must add `--allow-dir`, or use `--dir`.
