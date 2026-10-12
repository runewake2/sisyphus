---
title: "Run the AI roles as Claude Code subagents and as workers"
state: closed          # open | in-progress | closed. Must match the directory of the file.
resolution: completed  # completed | abandoned. Set this only when state is closed.
priority: high         # critical | high | medium | low
effort: medium         # small | medium | large | x-large
tags: [hephaestus, llm] # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-12
closed: 2026-10-12     # YYYY-MM-DD. Set this only when state is closed.
owner: claude          # The person or agent working on the issue. Cleared when the issue returns to open.
approver:              # The person or agent who accepts the issue when it closes.
bookmark: samw/ai/ux-fixes # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: [ux-fixes] # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
metadata: {}           # Optional. Arbitrary key-value notes, for example an AI agent session id to resume work with context: {session-id: "abc123"}.
deferred-from:         # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent: "[[hephaestus/core/delegate-issues-to-agents]]" # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
depends-on: []           # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# Run the AI roles as Claude Code subagents and as workers

## Summary

The AI roles have role files, but nothing can run them. Make each AI role runnable: as a Claude Code subagent that a human starts, and as a hephaestus worker that runs a model through a runner command (Q4).

## Context

Requested by a human on 2026-10-12: implement the roles. Q4 of [[agent-delegation]] says that AI agents start as Claude Code subagents, started by a human, then hephaestus runs them.

## Acceptance criteria

- [ ] `hephaestus agent export claude` writes a subagent for each AI role, and a test keeps them in sync.
- [ ] `hephaestus agent run <ai-role>` runs a model through a runner with the prompt of the task, and never without a runner.
- [ ] An issue that the model leaves in progress, and a failed runner, go to a human.

## Out of scope

- Starting the workers on their triggers: [[delegation-runner]].

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

Completed in [[0.0.117]]. `hephaestus agent export claude` writes a Claude Code subagent for each AI role (a test keeps them in sync), and `hephaestus agent run <ai-role> --runner <command>` runs a model with the prompt of each task, never without a runner; an issue left in progress or a failed runner goes to a human. See [[0004-ai-roles-run-as-claude-code-subagents]].

