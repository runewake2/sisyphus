---
title: "Add sisyphus-mcp to provide sisyphus to agents over MCP"
state: closed          # open | in-progress | closed. Must match the directory of the file.
resolution: completed  # completed | abandoned. Set this only when state is closed.
priority: high         # critical | high | medium | low
effort: large          # small | medium | large | x-large
tags: [cli, mcp]       # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-07
closed: 2026-10-07     # YYYY-MM-DD. Set this only when state is closed.
owner: claude          # The person or agent working on the issue. Cleared when the issue returns to open.
approver: samw         # The person or agent who accepts the issue when it closes.
bookmark: samw/ai/sisyphus-mcp # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: [sisyphus-mcp-server] # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
metadata: {}           # Optional. Arbitrary key-value notes, for example an AI agent session id to resume work with context: {session-id: "abc123"}.
deferred-from:         # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent:                # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
depends-on: []           # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# Add sisyphus-mcp to provide sisyphus to agents over MCP

## Summary

Add a `cmd/sisyphus-mcp` binary: an MCP server, over stdio, that gives an agent every `sisyphus`
command as a tool, so an agent can manage issues directly instead of shelling out to the CLI itself.

## Context

Used the official `github.com/modelcontextprotocol/go-sdk` (v1.8.0), not a community one, for
long-term support. Chose a thin-wrapper design over refactoring sisyphus's domain logic into an
importable package: each tool shells out to the installed `sisyphus` binary (`exec.LookPath`) and
returns its output, the same pattern already used by the `issue-to-pr`/`sync-to-github` GitHub
Actions. This means the MCP server can never drift from the CLI's behavior, and needed no changes
to `cmd/sisyphus` itself (only the new `cmd/sisyphus-mcp` package).

One tool per CLI command: `sisyphus_new`, `sisyphus_update`, `sisyphus_show`, `sisyphus_list`,
`sisyphus_search`, `sisyphus_graph`, `sisyphus_parent`, `sisyphus_remote`, `sisyphus_depends_on`,
`sisyphus_slug`, `sisyphus_resolve`, `sisyphus_links`, `sisyphus_init`. Every tool takes an optional
`dir` (defaults to the server's own working directory, like running the CLI directly). `show`/
`list`/`search` default to JSON (a `text` flag opts into the human-readable form); `graph` defaults
to the human-readable tree (a `json` flag opts into JSON), since that command exists for visual
review.

## Acceptance criteria

- [x] `cmd/sisyphus-mcp` runs an MCP server over stdio.
- [x] Every `sisyphus` command is available as a tool.
- [x] A clear error (not a crash) when `sisyphus` is not installed or not on `PATH`.
- [x] Covered by tests that connect a real MCP client to the server (over the SDK's in-memory
      transport) and call tools end to end, not just unit tests of the argument-building.
- [x] README documents installing and registering it with an MCP client.

## Out of scope

- Refactoring sisyphus's domain logic into an importable package; the thin-wrapper design avoids
  needing that.
- Resources or prompts (MCP features beyond tools); only tools are needed here.
- Authentication or remote (non-stdio) transports; this is a local, stdio-only server.

## Notes

- 2026-10-07: Implemented as planned. `jsonschema` struct tags on each tool's argument type (no
  `description=` prefix: the whole tag value is the description) drive the SDK's automatic input
  schema and validation. Tests build the `sisyphus` binary once per test run (`sync.OnceFunc`) and
  prepend it to `PATH`, so `go test ./...` works from a clean checkout with no separate install
  step, in CI or locally.

## Resolution

Completed: `sisyphus-mcp` exposes every `sisyphus` command as an MCP tool over stdio. See
[[0.0.12]].

## Resolution

<Complete this section when you close the issue.
- Completed: tell what was done. Link to the changelog version and the decision records, for example `[[0.0.5]]`.
- Abandoned: tell why the work stopped. Link to the issue or decision that replaces it, if one exists.>
