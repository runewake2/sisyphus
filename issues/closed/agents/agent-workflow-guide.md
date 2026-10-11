---
title: "Write a guide for repos that let AI agents use sisyphus"
state: closed          # open | in-progress | closed. Must match the directory of the file.
resolution: completed  # completed | abandoned. Set this only when state is closed.
priority: medium       # critical | high | medium | low
effort: small          # small | medium | large | x-large
tags: [docs, llm]      # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-11
closed: 2026-10-11     # YYYY-MM-DD. Set this only when state is closed.
owner: claude          # The person or agent working on the issue. Cleared when the issue returns to open.
approver:              # The person or agent who accepts the issue when it closes.
bookmark: samw/ai/ux-fixes # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: [ux-fixes] # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
metadata: {}           # Optional. Arbitrary key-value notes, for example an AI agent session id to resume work with context: {session-id: "abc123"}.
deferred-from:         # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent: "[[agents/agent-cooperation]]" # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
depends-on: []           # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# Write a guide for repos that let AI agents use sisyphus

## Summary

A repo that adopts sisyphus has no instructions to give its agents. This repo has AGENTS.md, but it is about work on sisyphus itself. Write a docs page with a block that a repo can paste into its AGENTS.md or CLAUDE.md: how to find work, claim it, record progress, close it, and work beside other agents.

## Context

The README and [[sisyphus-mcp]] tell how to install and connect, but not how an agent must act. Nothing tells an agent to give an owner when it starts an issue, or to fill the Resolution section when it closes one (`close` without `-m` warns, but succeeds).

The page must cover:

- Find work: `sisyphus list --ready`, and read the issue with `show` before the start.
- Claim: `sisyphus start <name> --owner <agent> --bookmark <branch>`, and what to do when the issue is already claimed ([[start-refuses-claimed-issue]], [[claims-across-workspaces]]).
- Progress: `sisyphus edit <name> notes "<date>: <text>" --append`. Never `edit <name>` alone, which opens an editor.
- Defer work: `sisyphus new --deferred-from <name>` for work that is found but out of scope.
- Close: `sisyphus close <name> -m "<what was done>"`, then `sisyphus lint`.
- A short and a long version of the block, and both the CLI and the MCP forms.

`sisyphus init` can print a line that names the page.

## Acceptance criteria

- [ ] A docs page (for example `docs/agents.md`) has a block to paste, for the CLI and for MCP.
- [ ] The README and [[sisyphus-mcp]] link to it.
- [ ] A test checks that each command in the block exists, as for the command examples.

## Out of scope

- The server instructions: [[mcp-server-instructions]].

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

Completed in [[0.0.93]]. [[agents]] has the setup, blocks to paste for a shell (long and short) and for MCP, and how several agents work together. A test checks every command and flag on the page. The README, [[sisyphus-mcp]], and [[commands]] link to it.

