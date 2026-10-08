---
title: "Build the issue index once per command"
state: closed          # open | in-progress | closed. Must match the directory of the file.
resolution: completed  # completed | abandoned. Set this only when state is closed.
priority: medium       # critical | high | medium | low
effort: medium         # small | medium | large | x-large
tags: [cli, performance] # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-08
closed: 2026-10-08     # YYYY-MM-DD. Set this only when state is closed.
owner: claude          # The person or agent working on the issue. Cleared when the issue returns to open.
approver: samw         # The person or agent who accepts the issue when it closes.
bookmark: samw/ai/index-issues-once # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: [sisyphus-index-issues-once] # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
metadata: {session-id: "5f1f06e4-afca-4277-8f3c-b8c818789ed6"} # Optional. Arbitrary key-value notes, for example an AI agent session id to resume work with context: {session-id: "abc123"}.
deferred-from: "[[issue-subdirectories]]" # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent:                # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
depends-on: []           # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# Build the issue index once per command

## Summary

Read the state directories once per command and answer every lookup from that index, so that a command reads each issue file at most once.

## Context

Since 0.0.18, every issue lookup walked all three state directories again, and some lookups walked them twice. A lookup for each parent or depends-on reference made list --blocked, list --parent, graph, the close-time warnings, and the cycle check cost O(n^2) directory entries for n issues. Each command also parsed some issue files more than once. sisyphus slug walked the whole repo once for each candidate name.

## Acceptance criteria

- [x] One command reads each state directory once. Lookups, full names, sub-issues, dependents, list, search, and graph use the index.
- [x] One command parses each issue file at most once.
- [x] Name resolution does not change, and the existing tests pass unchanged.
- [x] `sisyphus slug` reads the repo once.
- [x] A benchmark with 3000 issues in subdirectories measures list --blocked, list --parent, graph --full, and update before and after the change.

## Out of scope

A cache that lasts longer than one command, for example on disk or in sisyphus-mcp. Each command must see the edits that other sessions make between commands.

## Notes

2026-10-08: benchmark with 3000 issues in project and component directories, each with a bare-name parent and depends-on link (cmd/sisyphus/bench_test.go):

| Command | 0.0.18 | 0.0.19 |
|---|---|---|
| list --blocked | 30.0 s | 65 ms |
| list --parent | 29.7 s | 54 ms |
| graph --full | 73.3 s | 434 ms |
| update, without a close | 11.6 ms | 7.4 ms |

Most of the remaining graph --full time is the layout of 3000 nodes, not the lookups.

## Resolution

Completed in [[0.0.19]].
