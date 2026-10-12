# Unreleased: Run tool agents, and let them review in referendums

- Date: 2026-10-12
- Bookmark: samw/ai/ux-fixes
- Components: hephaestus
- Changelog: Run tool agents with hephaestus agent run, which do delegated tasks with a program and no model, and add hephaestus agent review.

## Summary

hephaestus runs tool agents: workers of a tool role that get each task as a message, claim its
issue, run the command of the role, and close or stop the issue from the exit code, with no model
call. A tool agent can also vote in a referendum.

## Changes

- `hephaestus agent run <role>`: a worker of a tool role (`<role>-1` unless `--as`). It claims the
  issue of each message, runs the command in the workspace of the task with the message on standard
  input, closes the issue on exit code 0, and otherwise records the output in `Notes` and stops the
  issue, tagged `needs-human` or delegated to the AI role of `on-failure` with a task message. It
  sends a result message to the sender. `--once` handles the messages that wait now.
- `hephaestus agent review <role>`: the ballot of a tool reviewer, `yes`, `no`, or `modify` with the
  changed files and the diff.
- Role files take `listens`, the kinds of broadcast that a role reacts to. A quoted value in a role
  file keeps the quotes inside it, so that a command can quote.
- In git, the working diff of a review includes new files that git does not track yet.
- The tool role `repo-check` runs the checks of this repo. As the first real task of a tool agent, it
  checked the working copy of this change through the message queue and closed
  [[check-tool-agents-change]].

## Decisions

- [[0003-a-tool-agent-runs-its-command-for-each-message]]

## Issues

- Closed [[tool-agents]].
- Created and closed [[check-tool-agents-change]].
