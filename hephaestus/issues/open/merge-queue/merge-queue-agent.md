---
title: "Add a merge queue agent that merges parallel work in the best order"
state: open            # open | in-progress | closed. Must match the directory of the file.
resolution:            # completed | abandoned. Set this only when state is closed.
priority: high         # critical | high | medium | low
effort: large          # small | medium | large | x-large
tags: [merge-queue, llm, hephaestus] # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-11
closed:                # YYYY-MM-DD. Set this only when state is closed.
owner:                 # The person or agent working on the issue. Cleared when the issue returns to open.
approver:              # The person or agent who accepts the issue when it closes.
bookmark:              # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: []         # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
metadata: {}           # Optional. Arbitrary key-value notes, for example an AI agent session id to resume work with context: {session-id: "abc123"}.
deferred-from:         # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent: "[[core/delegate-issues-to-agents]]" # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
depends-on: []           # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# Add a merge queue agent that merges parallel work in the best order

## Summary

Add a merge queue role: an agent that takes the finished work of many parallel workspaces and merges it, one bookmark at a time, in the order that keeps the work correct and the conflicts small, onto one line based on `main`. It is the first role of [[delegate-issues-to-agents]].

## Context

In this repo, the bookmarks of several workspaces were merged by hand, and each merge had to renumber the SHAME versions that the workspaces picked at the same time (`CONTRIBUTING.md` (Changelog)). [[agent-delegation#Merge queue]] describes the role: its input (bookmarks whose issues are all closed and whose checks pass), its order, how it handles each merge, and its output (a bookmark `merge-queue/<date>` and a report). A human moves `main` (Q3).

Parts:

- [[merge-order-planner]]: compute the order.
- [[renumber-versions-on-merge]]: resolve the conflicts of versions, changelogs, and decision record numbers.
- This issue: the role itself, in `hephaestus/roles/merge-queue.md`, and its runs in a scratch workspace.

The merge queue does not depend on [[delegation-model]] to start: a human can start it by hand on a list of bookmarks. When [[delegation-model]] is done, it also takes its work from issues delegated to `merge-queue`.

## Acceptance criteria

- [ ] Given a list of bookmarks, the merge queue makes `merge-queue/<date>`: one line on `main` with every bookmark that it could merge, in the order of [[merge-order-planner]].
- [ ] After each rebase, `gofmt`, `go vet`, the tests, and `sisyphus lint` pass; a bookmark that fails is left out, with the reason in the notes of its issues.
- [ ] It renumbers versions and decision records with [[renumber-versions-on-merge]].
- [ ] Each run records a report: the order, the renumbering, and what was left out and why.
- [ ] It never moves `main` or pushes, unless a human decides otherwise (Q3).
- [ ] A test run merges at least three real bookmarks of this repo, and a human checks the result.

## Out of scope

- Conflicts that are not mechanical. The merge queue stops there and records them; [[engineer-agent]] or a human resolves them.
- Pull requests on a code host.

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

<Complete this section when you close the issue.
- Completed: tell what was done. Link to the changelog version and the decision records, for example `[[0.0.5]]`.
- Abandoned: tell why the work stopped. Link to the issue or decision that replaces it, if one exists.>
