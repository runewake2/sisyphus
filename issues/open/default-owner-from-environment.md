---
title: "Take the default owner from the environment"
state: open            # open | in-progress | closed. Must match the directory of the file.
resolution:            # completed | abandoned. Set this only when state is closed.
priority: low          # critical | high | medium | low
effort: small          # small | medium | large | x-large
tags: [cli, llm]       # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-11
closed:                # YYYY-MM-DD. Set this only when state is closed.
owner:                 # The person or agent working on the issue. Cleared when the issue returns to open.
approver:              # The person or agent who accepts the issue when it closes.
bookmark:              # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: []         # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
metadata: {}           # Optional. Arbitrary key-value notes, for example an AI agent session id to resume work with context: {session-id: "abc123"}.
deferred-from:         # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent: "[[agent-cooperation]]" # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
depends-on: []           # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# Take the default owner from the environment

## Summary

`start` without `--owner` makes an issue that is in progress with no owner, so no one can tell who has it. Let a `SISYPHUS_OWNER` environment variable give the default owner, so that an agent harness sets the identity once.

## Context

An agent must remember to add `--owner` to each `start`. If it forgets, the claim has no owner, and [[start-refuses-claimed-issue]] cannot compare owners. A harness can set an environment variable for every agent session, for example `SISYPHUS_OWNER=claude-<session>`, and for an MCP server in its `env`.

Proposed:

- `start`, `update ... in-progress`, and `new --state in-progress` use `SISYPHUS_OWNER` when `--owner` is not given.
- `list --owner me` (or `--mine`) filters on the same value.
- `start` with no owner and no `SISYPHUS_OWNER` prints a warning.

## Acceptance criteria

- [ ] `SISYPHUS_OWNER` sets the owner when `--owner` is not given.
- [ ] `list` can filter on it.
- [ ] A start with no owner warns.
- [ ] [[sisyphus-start]] and [[sisyphus-mcp]] document it.

## Out of scope

- Other fields from the environment, for example `bookmark`.

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

<Complete this section when you close the issue.
- Completed: tell what was done. Link to the changelog version and the decision records, for example `[[0.0.5]]`.
- Abandoned: tell why the work stopped. Link to the issue or decision that replaces it, if one exists.>
