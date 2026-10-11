---
title: "List all among the valid states of list and search"
state: closed          # open | in-progress | closed. Must match the directory of the file.
resolution: completed  # completed | abandoned. Set this only when state is closed.
priority: low          # critical | high | medium | low
effort: small          # small | medium | large | x-large
tags: [bug, cli]       # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-11
closed: 2026-10-11     # YYYY-MM-DD. Set this only when state is closed.
owner: claude          # The person or agent working on the issue. Cleared when the issue returns to open.
approver:              # The person or agent who accepts the issue when it closes.
bookmark: samw/ai/ux-fixes # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: [ux-fixes] # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
metadata: {}           # Optional. Arbitrary key-value notes, for example an AI agent session id to resume work with context: {session-id: "abc123"}.
deferred-from:         # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent:                # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
depends-on: []           # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# List all among the valid states of list and search

## Summary

An invalid --state of list or search gives an error that lists only open, in-progress, and closed. all is valid there too, so the error must list it and suggest it.

## Context

Found on 2026-10-11 while documenting the output of sisyphus 0.0.103:

```console
$ sisyphus list --state bogus
Error: Invalid state 'bogus'. Use one of: open, in-progress, closed.
```

`stateFilter` in `cmd/sisyphus/cli.go` turns `all` into the three states before `validateEach` checks the values, so the check never sees `all`, and its choices do not have it. A value near `all`, for example `al`, gets no suggestion.

## Acceptance criteria

- [ ] The error of an invalid `--state` of `list` and `search` lists `all`.
- [ ] `--state al` suggests `all`.
- [ ] `update` and `new`, where `all` is not valid, do not list it.

## Out of scope

Nothing.

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

Completed in [[0.0.104]]. `list` and `search` check `--state` against the states and `all`, so the error lists `all` and suggests it for a near value.

