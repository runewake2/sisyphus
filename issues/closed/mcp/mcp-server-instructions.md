---
title: "Give the MCP server instructions that tell agents the workflow"
state: closed          # open | in-progress | closed. Must match the directory of the file.
resolution: completed  # completed | abandoned. Set this only when state is closed.
priority: medium       # critical | high | medium | low
effort: small          # small | medium | large | x-large
tags: [sisyphus-mcp, llm, docs] # The components that the work touches, for example [widget-scheduler, plan]
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

# Give the MCP server instructions that tell agents the workflow

## Summary

sisyphus-mcp sends no `instructions` when a client connects. An agent sees 15 tools but not the order to use them in: find ready work, claim it with an owner, record notes, and close it with a resolution. Add short server instructions.

## Context

The `initialize` result of sisyphus-mcp 0.0.84 has no `instructions` field (`mcp.NewServer` gets `nil` options in `cmd/sisyphus-mcp/main.go`). MCP clients such as Claude Code show the server instructions to the model. Today the workflow is only in AGENTS.md of this repo and in [[sisyphus-mcp]], and an agent in another repo sees neither.

The instructions must be short (each client adds them to every session) and must include:

- Use `sisyphus_list` with `ready` to find work, and do not take an issue that is in progress.
- Start an issue with `state: in-progress` and an `owner` (or the tool of [[mcp-lifecycle-tools]]).
- Record progress with `sisyphus_edit` `append` in Notes.
- Close with a `message` that fills Resolution, and run `sisyphus_lint`.
- Lists are comma-separated strings. Leave `dir` empty (see [[mcp-hide-dir-argument]]).

## Acceptance criteria

- [ ] The `initialize` result has `instructions` with the workflow above, in less than 200 words.
- [ ] A test checks that the instructions name only tools that exist.
- [ ] [[sisyphus-mcp]] shows the instructions.

## Out of scope

- The guide for repos: [[agent-workflow-guide]].

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

Completed in [[0.0.92]]. The server sends instructions with the workflow (list ready work, show, start, notes, defer, close, lint) in less than 200 words, plus a warning when the binary has another version. Tests keep them short, correct, and in step with [[sisyphus-mcp]].

