---
title: "Fail list and search for a parent filter that names no issue"
state: closed          # open | in-progress | closed. Must match the directory of the file.
resolution: completed  # completed | abandoned. Set this only when state is closed.
priority: low          # critical | high | medium | low
effort: small          # small | medium | large | x-large
tags: [bug, cli]       # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-11
closed: 2026-10-11     # YYYY-MM-DD. Set this only when state is closed.
owner: claude          # The person or agent working on the issue. Cleared when the issue returns to open.
approver:              # The person or agent who accepts the issue when it closes.
bookmark: samw/ai/tracker-fixes-2 # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: [tracker-fixes-2] # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
metadata: {}           # Optional. Arbitrary key-value notes, for example an AI agent session id to resume work with context: {session-id: "abc123"}.
deferred-from:         # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent:                # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
depends-on: []           # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# Fail list and search for a parent filter that names no issue

## Summary

`sisyphus list --parent no-such-epic` prints an empty table and exits 0. A typo in the parent name looks the same as an epic with no sub-issues. `search --parent` does the same.

## Context

Found in a review of sisyphus 0.0.45 on 2026-10-11. Other filters that name an issue fail for a name that does not exist, and since 0.0.33
`search --section` fails for an unknown section. `--parent` must follow the same rule. `--owner` is free text and
can stay as it is.

## Acceptance criteria

- [ ] `list --parent` and `search --parent` fail with the usual "No issue named" error when the name gives no issue, and with the usual error when it gives more than one.
- [ ] The MCP tools give the same error.
- [ ] Tests cover both commands.

## Out of scope

Nothing.

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

Completed in [[0.0.52]]. `list --parent` and `search --parent` fail with the usual error for a name that gives no issue.

