---
title: "Make sisyphus safe and easy for AI agents that work together"
state: closed          # open | in-progress | closed. Must match the directory of the file.
resolution: completed  # completed | abandoned. Set this only when state is closed.
priority: high         # critical | high | medium | low
effort: large          # small | medium | large | x-large
tags: [llm, sisyphus-mcp, cli] # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-11
closed: 2026-10-11     # YYYY-MM-DD. Set this only when state is closed.
owner:                 # The person or agent working on the issue. Cleared when the issue returns to open.
approver:              # The person or agent who accepts the issue when it closes.
bookmark:              # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: []         # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
metadata: {}           # Optional. Arbitrary key-value notes, for example an AI agent session id to resume work with context: {session-id: "abc123"}.
deferred-from:         # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent:                # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
depends-on: []           # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# Make sisyphus safe and easy for AI agents that work together

## Summary

sisyphus is for repos where AI agents do much of the work, often several agents at the same time. A review of sisyphus 0.0.84 for agents found problems that let two agents take the same issue, make a tool call wait for an editor, and spend many tokens on template text. This epic holds the fixes.

## Context

Found in a review on 2026-10-11 of sisyphus 0.0.84 and sisyphus-mcp 0.0.84. The review ran the CLI as an agent runs it (no terminal, through a shell tool) and ran sisyphus-mcp with a JSON-RPC client, in an empty test repo and in this repo. Each sub-issue tells what the review saw.

The sub-issues are in three groups:

- **Cooperation between agents**: [[start-refuses-claimed-issue]], [[claims-across-workspaces]], [[atomic-issue-writes]], [[default-owner-from-environment]].
- **The MCP server**: [[edit-without-terminal]], [[mcp-cancel-and-timeout]], [[mcp-server-instructions]], [[mcp-tool-annotations]], [[mcp-lifecycle-tools]], [[mcp-version-check]], [[mcp-vcs-neutral-descriptions]].
- **Output and documentation for agents**: [[compact-json-for-agents]], [[agent-workflow-guide]].

[[small-model-friendly-mcp]] holds related work that the spike with local models found. This epic does not repeat it.

## Acceptance criteria

- [x] All sub-issues are closed.
- [x] Two agents that run `sisyphus start` on the same issue cannot both think that they have it.

## Out of scope

- The work of [[small-model-friendly-mcp]].
- Sync with GitHub or Jira ([[integration-plugins]]).

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

Completed in [[0.0.85]] to [[0.0.102]]. All 13 sub-issues are closed: claims are refused across owners and workspaces ([[0022-refuse-to-take-a-claimed-issue]], [[0024-share-claims-between-workspaces]]), writes are atomic and locked ([[0023-lock-and-write-issues-atomically]]), the MCP server has lifecycle tools, instructions, annotations, time limits, and a version check, and [[agent-guide]] tells repos how to set up agents. Claims do not cross machines; that is future work.

