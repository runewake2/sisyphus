---
title: "Draw only the issues in selected states in sisyphus graph"
state: closed          # open | in-progress | closed. Must match the directory of the file.
resolution: completed  # completed | abandoned. Set this only when state is closed.
priority: medium       # critical | high | medium | low
effort: small          # small | medium | large | x-large
tags: [feature, cli, graph] # The components that the work touches, for example [widget-scheduler, plan]
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

# Draw only the issues in selected states in sisyphus graph

## Summary

A graph of a large epic draws every closed and abandoned issue, so the work that remains is hard to see. Let graph draw only the issues in selected states, for example only open and in-progress issues.

## Context

Requested by a human on 2026-10-11. In this repo, the graph of [[large-repo-performance]] draws 13 issues, and only 3 of them are open. A filter must not break the drawing: a link through an issue that is left out must still connect what it connects, but only where that stays true.

## Acceptance criteria

- [ ] `sisyphus graph <name> --state open,in-progress` draws only issues in those states, and always `<name>`.
- [ ] A chain of links through issues that are left out is joined where it keeps its meaning, and the JSON tells which issues it goes through.
- [ ] The legend tells how many issues are left out, and availability does not change.
- [ ] The MCP tool takes the filter.

## Out of scope

Filters on other fields, for example tags or owner.

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

Completed in [[0.0.106]]. `graph --state` draws only the issues in the given states and always the focus; chains through left-out issues are joined where they stay true, availability does not change, and the legend tells how many issues are left out. See [[0028-join-graph-links-through-filtered-issues]].

