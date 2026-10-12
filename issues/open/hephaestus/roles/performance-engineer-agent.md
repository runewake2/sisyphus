---
title: "Add a performance engineer that makes the commands fast on large repos"
state: open            # open | in-progress | closed. Must match the directory of the file.
resolution:            # completed | abandoned. Set this only when state is closed.
priority: medium       # critical | high | medium | low
effort: medium         # small | medium | large | x-large
tags: [hephaestus, llm, performance] # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-12
closed:                # YYYY-MM-DD. Set this only when state is closed.
owner:                 # The person or agent working on the issue. Cleared when the issue returns to open.
approver:              # The person or agent who accepts the issue when it closes.
bookmark:              # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: []         # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
metadata: {}           # Optional. Arbitrary key-value notes, for example an AI agent session id to resume work with context: {session-id: "abc123"}.
deferred-from:         # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent: "[[hephaestus/core/delegate-issues-to-agents]]" # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
depends-on: []           # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# Add a performance engineer that makes the commands fast on large repos

## Summary

Add the AI role performance-engineer: it measures the commands on large generated repos, profiles the slow ones, and fixes one cause at a time with a benchmark before and after and no change of behavior.

## Context

Requested by a human on 2026-10-12: implement other agents that may be useful, for a government of agents who coordinate and work together. The roles may be removed after a review; a trial run shows how useful each is. The role file is [[roles/performance-engineer]]; its section Works with tells how it works with the other roles.

## Acceptance criteria

- [x] The role file describes the role, its hand-offs in Works with, and the points where it must ask a human.
- [ ] A trial run on one command, after [[add-performance-benchmarks]], finds a cause of slowness with a profile, and its fix shows a gain with the tests passing.
- [ ] A human checks the result of the trial run, and decides whether the role stays.

## Out of scope

<List related work that this issue does not include. Create a separate issue for each item that someone must do later.>

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

<Complete this section when you close the issue.
- Completed: tell what was done. Link to the changelog version and the decision records, for example `[[0.0.5]]`.
- Abandoned: tell why the work stopped. Link to the issue or decision that replaces it, if one exists.>
