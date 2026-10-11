---
title: "Add tool annotations and structured output to the MCP tools"
state: closed          # open | in-progress | closed. Must match the directory of the file.
resolution: completed  # completed | abandoned. Set this only when state is closed.
priority: medium       # critical | high | medium | low
effort: small          # small | medium | large | x-large
tags: [sisyphus-mcp, llm] # The components that the work touches, for example [widget-scheduler, plan]
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

# Add tool annotations and structured output to the MCP tools

## Summary

The sisyphus-mcp tools have no annotations, so a client cannot tell that `sisyphus_show` only reads and that `sisyphus_update` changes files. Clients use these hints to approve read-only calls without a prompt. Add annotations, and return JSON results as structured content.

## Context

`tools/list` of sisyphus-mcp 0.0.84 has no `annotations` and no `outputSchema` on any tool, and results have no `structuredContent`: JSON comes back as text only.

Proposed annotations:

| Tools | `readOnlyHint` | `destructiveHint` | `idempotentHint` |
|---|---|---|---|
| show, list, search, graph, slug, resolve, links, lint | true | | |
| new | false | false | false |
| update, edit, parent, depends-on, remote | false | false | true for a replace, false for an append |
| init | false | true with `force` (it replaces the template) | true |

Also give each tool a `title` for the client UI.

For the tools that return JSON, also return the parsed value as `structuredContent`, with an `outputSchema`, so that a client can use the fields without a parse of the text. Keep the text content for clients that do not read structured content.

## Acceptance criteria

- [ ] Every tool has annotations and a title, and a test checks the read-only set.
- [ ] `sisyphus_show`, `sisyphus_list`, `sisyphus_search`, and `sisyphus_lint` return `structuredContent` that matches their `outputSchema`.
- [ ] [[sisyphus-mcp]] lists the read-only tools.

## Out of scope

- A change of the JSON fields themselves: [[compact-json-for-agents]], [[output-conventions-doc]].

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

Completed in [[0.0.102]]. Every tool has a title and read-only, destructive, idempotent, and open-world hints, and JSON results also come back as `structuredContent` (arrays wrapped in `results`). No `outputSchema` is declared, because the same tools return text on request; see the changelog.

