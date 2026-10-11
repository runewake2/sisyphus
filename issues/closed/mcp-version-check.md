---
title: "Report a version difference between sisyphus-mcp and sisyphus"
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
parent: "[[agent-cooperation]]" # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
depends-on: []           # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# Report a version difference between sisyphus-mcp and sisyphus

## Summary

sisyphus-mcp runs the `sisyphus` that it finds on `PATH`, and does not check its version. When the two versions are different, a tool can pass a flag that the old binary does not know, and the agent gets `unknown flag` with no hint about the cause. Check the version at start-up, and tell the agent about a difference.

## Context

`execSisyphus` in `cmd/sisyphus-mcp/main.go` calls `exec.LookPath("sisyphus")` on each call. The two binaries are installed separately (`go install` twice), and an MCP client that starts from a desktop launcher often has a different `PATH` (see [[sisyphus-mcp#Other clients]]). So a difference is easy to get, for example after an update of only one binary.

Each new flag, for example `--reopen` or `--json` on the commands that change an issue, makes the risk larger.

Proposed: at start-up, run `sisyphus --version` once. If it is different from the version of sisyphus-mcp, add a line to every error result (`sisyphus-mcp 0.0.90 runs sisyphus 0.0.80 from /usr/local/bin/sisyphus. Install the same version of both.`), and write it to the server log. Also put the path and version of the binary in the server instructions ([[mcp-server-instructions]]).

## Acceptance criteria

- [ ] A version difference adds a clear line to the errors of the tools.
- [ ] A test covers a fake `sisyphus` with a different version.
- [ ] The troubleshooting table of [[sisyphus-mcp]] has a row for it.

## Out of scope

- One binary for both the CLI and the server.

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

Completed in [[0.0.91]]. sisyphus-mcp reads the version of the `sisyphus` binary once, and when it is different, adds a line with both versions and the path to every error result and to its log.

