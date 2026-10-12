---
title: "Add a merge queue agent that merges parallel work in the best order"
state: open            # open | in-progress | closed. Must match the directory of the file.
resolution:            # completed | abandoned. Set this only when state is closed.
priority: high         # critical | high | medium | low
effort: large          # small | medium | large | x-large
tags: [merge-queue, hephaestus] # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-11
closed:                # YYYY-MM-DD. Set this only when state is closed.
owner:                 # The person or agent working on the issue. Cleared when the issue returns to open.
approver:              # The person or agent who accepts the issue when it closes.
bookmark:              # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: []         # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
metadata: {}           # Optional. Arbitrary key-value notes, for example an AI agent session id to resume work with context: {session-id: "abc123"}.
deferred-from:         # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent: "[[core/delegate-issues-to-agents]]" # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
depends-on: ["[[core/tool-agents]]"] # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# Add a merge queue agent that merges parallel work in the best order

## Summary

Add a merge queue role: a tool agent, which uses no model ([[tool-agents]]), whose only job is to make one linear history from a set of tasks: it merges the bookmark of each task, one at a time, in the order that keeps the work correct and the conflicts small, onto one line based on `main`. It never moves `main`. It is the first role of [[delegate-issues-to-agents]].

## Context

In this repo, the bookmarks of several workspaces were merged by hand, and each merge had to renumber the SHAME versions that the workspaces picked at the same time ([[CONTRIBUTING#Changelog]]). [[agent-delegation#Merge queue]] describes the role: its input (a set of tasks, each with its bookmark), its order, how it handles each merge, and its output (a bookmark `merge-queue/<date>` and a report). It never moves `main` (Q3, answered by the human).

Parts:

- [[merge-order-planner]]: compute the order.
- [[renumber-versions-on-merge]]: resolve the conflicts of versions, changelogs, and decision record numbers.
- This issue: the role itself, a tool agent in `hephaestus/roles/merge-queue.md`, and its runs in a scratch workspace.

The merge queue is a tool agent, so it depends on [[tool-agents]], which runs it from a message that gives the set of tasks. Its two parts, [[merge-order-planner]] and [[renumber-versions-on-merge]], are programs that can start now, and run by hand before hephaestus can run them.

## Acceptance criteria

- [ ] Given a set of tasks, the merge queue makes `merge-queue/<date>`: one linear history on `main` with the work of every task that it could merge, in the order of [[merge-order-planner]].
- [ ] After each rebase, `gofmt`, `go vet`, the tests, and `sisyphus lint` pass; a bookmark that fails is left out, with the reason in the notes of its issues.
- [ ] It renumbers versions and decision records with [[renumber-versions-on-merge]].
- [ ] Each run records a report: the order, the renumbering, and what was left out and why.
- [ ] It never moves `main` and never pushes (Q3).
- [ ] It makes no model call (Q10). A conflict that is not mechanical stops it, and it sends a message about it to the `engineer` role or a human.
- [ ] A test run merges the work of at least three real tasks of this repo, and a human checks the result.

## Out of scope

- Conflicts that are not mechanical. The merge queue stops there and records them; [[engineer-agent]] or a human resolves them.
- Pull requests on a code host.

## Notes

2026-10-11, from the human: the merge queue never moves main. Its only job is to create a linear history from a set of tasks. This answers Q3 of [[agent-delegation]]; its input is now a set of tasks, not a list of bookmarks.

2026-10-11, from the human: the merge queue is a tool agent (Q10): it uses no model. A conflict that is not mechanical stops it, and it sends a message to the engineer or a human ([[message-queue]]).

2026-10-12, from the human: human work, and branches introduced beside the queue, land on the linear branch through the supervised merge queue (Q6).

## Resolution

<Complete this section when you close the issue.
- Completed: tell what was done. Link to the changelog version and the decision records, for example `[[0.0.5]]`.
- Abandoned: tell why the work stopped. Link to the issue or decision that replaces it, if one exists.>
