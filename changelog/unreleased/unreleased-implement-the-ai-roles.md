# Unreleased: Run the AI roles, and add the security researcher and the privacy reviewer

- Date: 2026-10-12
- Bookmark: samw/ai/ux-fixes
- Components: hephaestus
- Changelog: Run the AI roles as Claude Code subagents and as workers with a runner, and add the roles security-researcher and privacy-reviewer.

## Summary

The AI roles of hephaestus run: as Claude Code subagents, which `hephaestus agent export claude`
makes from the role files, and as hephaestus workers that run a model through a runner command.
Two more AI roles join the government: the security researcher and the privacy reviewer.

## Changes

- `hephaestus agent export claude` writes `.claude/agents/<role>.md` for each AI role, and `--check`
  tells whether they are in sync. The six subagents are in the repo, and a test keeps them in sync
  with the role files.
- `hephaestus agent run <ai-role> --runner <command>` (or `HEPHAESTUS_AI_RUNNER`) runs a worker of
  an AI role: the runner, for example `claude -p --agent {role}`, gets the prompt of each task (the
  role file, the rules of every role, and the message). The model closes or stops the issue; an
  issue left in progress, and a failed runner, go to a human. Without a runner, hephaestus refuses.
- Role files of AI roles take `allowed-tools` and `model`, which go into their subagents.
- `hephaestus/roles/security-researcher.md`: secrets, file permissions, input that reaches a shell
  or a path, services, dependencies, and agents; it describes the class and place of a weakness,
  never an exploit.
- `hephaestus/roles/privacy-reviewer.md`: follows each piece of data about a person through the
  tools, and files a way to collect, keep, show, or send less.
- Links to a role file are `[[roles/<role>]]`, because its subagent has the same name.
- [[agent-delegation#The government]] has the two new roles, and [[role-file]] tells how to run an
  AI role.

## Decisions

- [[0004-ai-roles-run-as-claude-code-subagents]]

## Issues

- Closed [[run-ai-roles]].
- Created [[security-researcher-agent]] and [[privacy-reviewer-agent]].
