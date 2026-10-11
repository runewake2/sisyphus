---
title: "Add setup for more clients, examples, and troubleshooting to the MCP page"
state: open            # open | in-progress | closed. Must match the directory of the file.
resolution:            # completed | abandoned. Set this only when state is closed.
priority: medium       # critical | high | medium | low
effort: small          # small | medium | large | x-large
tags: [docs, sisyphus-mcp] # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-11
closed:                # YYYY-MM-DD. Set this only when state is closed.
owner:                 # The person or agent working on the issue. Cleared when the issue returns to open.
approver:              # The person or agent who accepts the issue when it closes.
bookmark:              # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: []         # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
metadata: {}           # Optional. Arbitrary key-value notes, for example an AI agent session id to resume work with context: {session-id: "abc123"}.
deferred-from:         # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent: "[[cli-and-docs-ux]]" # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
depends-on: []           # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# Add setup for more clients, examples, and troubleshooting to the MCP page

## Summary

The `sisyphus-mcp` page has an install section and a table of tools. Add setup for more MCP clients, an example of tool calls with their results, how warnings and errors come back, and troubleshooting.

## Context

Found in a review of the CLI, the help text, and the documentation of sisyphus 0.0.67 on 2026-10-11, with a first session in an empty repo. The README shows only `claude mcp add sisyphus -- sisyphus-mcp`. A user of another client needs the
JSON configuration form, and the note that `sisyphus` must be on the `PATH` of the client process, which is a common
cause of failure (every tool then returns the "not installed" error). Useful additions:

- A configuration example for a generic MCP client (command and arguments) and for Ollama-based clients
  (see [[ollama-task-manager-guide]]).
- One short example session: `sisyphus_list`, `sisyphus_new`, `sisyphus_update`, with the text that each returns.
- How warnings come back (after the output), and how errors come back (`isError` with the message of the command).
- Troubleshooting: sisyphus not on PATH, the `dir` argument, the repo-not-found error.

## Acceptance criteria

- [ ] [[sisyphus-mcp]] has setup for Claude Code and a generic JSON configuration, an example session, the warning and error behavior, and a troubleshooting section.

## Out of scope

- Changes to the tools themselves ([[small-model-friendly-mcp]]).

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

<Complete this section when you close the issue.
- Completed: tell what was done. Link to the changelog version and the decision records, for example `[[0.0.5]]`.
- Abandoned: tell why the work stopped. Link to the issue or decision that replaces it, if one exists.>
