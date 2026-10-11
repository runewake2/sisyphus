---
title: "Add list --ready for issues that can start now"
state: open            # open | in-progress | closed. Must match the directory of the file.
resolution:            # completed | abandoned. Set this only when state is closed.
priority: medium       # critical | high | medium | low
effort: small          # small | medium | large | x-large
tags: [feature, cli]   # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-11
closed:                # YYYY-MM-DD. Set this only when state is closed.
owner:                 # The person or agent working on the issue. Cleared when the issue returns to open.
approver:              # The person or agent who accepts the issue when it closes.
bookmark:              # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: []         # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
metadata: {}           # Optional. Arbitrary key-value notes, for example an AI agent session id to resume work with context: {session-id: "abc123"}.
deferred-from:         # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent: "[[issue-manager-gaps]]" # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
depends-on: []           # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# Add list --ready for issues that can start now

## Summary

Add `sisyphus list --ready`: open issues whose `depends-on` issues are all closed. It is the inverse of `--blocked`, and it answers the main question of a task manager: what can I start now?

## Context

Found in a review of sisyphus 0.0.31 on 2026-10-11. `sisyphus graph` already computes this set and calls these issues "available" (see [[rename-graph-leaf-available]] and [[trace-graph-work-path]]), but only inside one drawing. `list --blocked` exists. To find work today, a user or an agent must combine `list` and `list --blocked` by hand, or read a graph for each epic.

## Acceptance criteria

- [ ] `list --ready` (and `search --ready`) shows open issues that have no `depends-on` issue that is not closed.
- [ ] The rule matches the "available" issues of `sisyphus graph`. Decide whether an epic with open sub-issues counts as ready, and record it in [[sisyphus-list]].
- [ ] `--ready` and `--blocked` together fail with a clear error.
- [ ] The MCP tools have the flag, and tests cover it.

## Out of scope

- Ranking the ready issues.

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

<Complete this section when you close the issue.
- Completed: tell what was done. Link to the changelog version and the decision records, for example `[[0.0.5]]`.
- Abandoned: tell why the work stopped. Link to the issue or decision that replaces it, if one exists.>
