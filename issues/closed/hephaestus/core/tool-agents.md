---
title: "Run tool agents that do delegated work with no AI"
state: closed          # open | in-progress | closed. Must match the directory of the file.
resolution: completed  # completed | abandoned. Set this only when state is closed.
priority: high         # critical | high | medium | low
effort: medium         # small | medium | large | x-large
tags: [hephaestus]     # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-11
closed: 2026-10-12     # YYYY-MM-DD. Set this only when state is closed.
owner: claude          # The person or agent working on the issue. Cleared when the issue returns to open.
approver:              # The person or agent who accepts the issue when it closes.
bookmark: samw/ai/ux-fixes # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: [ux-fixes] # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
metadata: {}           # Optional. Arbitrary key-value notes, for example an AI agent session id to resume work with context: {session-id: "abc123"}.
deferred-from:         # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent: "[[hephaestus/core/delegate-issues-to-agents]]" # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
depends-on: ["[[hephaestus/core/delegation-model]]", "[[hephaestus/core/message-queue]]"] # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# Run tool agents that do delegated work with no AI

## Summary

hephaestus must also run tool agents: roles whose work is a program with fixed rules, not a model. Delegate large refactors, lints, formatters, and other work that needs no judgment to them, so that it costs no tokens. [[agent-delegation#AI agents and tool agents]] describes them.

## Context

Requested by a human on 2026-10-11, with the answer to Q4: AI agents start as Claude Code subagents, then the Agent SDK, and hephaestus must also support agents that do not use AI, to avoid large token spend. A tool agent is a role like any other (a role file with kind: tool and its command, an owner name, a workspace, and claims), and follows the same rules; only the worker differs. Open questions: Q8 (how it gets its task and reports the result), Q9 (tool agents as referendum reviewers), and Q10 (the merge queue as a tool agent).

## Acceptance criteria

- [ ] A role file can describe a tool agent: `kind: tool`, its command, and what it does when the command fails.
- [ ] hephaestus runs a tool agent on a delegated task, from a message that gives its context ([[message-queue]]): it claims the issue, runs the command in the workspace of the task, and closes or stops the issue from the result, with no model call.
- [ ] A tool agent can be a reviewer in a referendum: its exit code is its ballot, and a formatter can submit a modification.
- [ ] A decision record answers Q8, Q9, and Q10.
- [ ] One real tool agent, for example a lint or a codemod, resolves a real task of this repo.

## Out of scope

- Writing the codemods and lints themselves: each is a tool of its own, which a tool agent runs.

## Notes

2026-10-11, from the human: tool agents can be referendum reviewers (Q9). Tool agents get their context as a message, which other agents can broadcast (Q8, the context part); the message queue is another tool agent ([[message-queue]]). The merge queue is a tool agent (Q10).

2026-10-12, from the human: the proposal for the result of a tool agent is accepted (Q8): exit code 0 closes the issue with the summary of the command; another exit code stops it with the output in Notes, tagged needs-human, unless the role file names an AI role to delegate the rest to.

## Resolution

Completed in [[0.0.113]]. `hephaestus agent run <role>` runs a worker of a tool role from messages: it claims the issue, runs the command, closes the issue on exit code 0, and otherwise stops it for a human or the AI role of on-failure. `hephaestus agent review <role>` gives the ballot of a tool reviewer. The tool role repo-check did the first real task, [[check-tool-agents-change]]. See [[0003-a-tool-agent-runs-its-command-for-each-message]].

