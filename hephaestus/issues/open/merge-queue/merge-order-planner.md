---
title: "Compute the order in which the merge queue merges bookmarks"
state: open            # open | in-progress | closed. Must match the directory of the file.
resolution:            # completed | abandoned. Set this only when state is closed.
priority: high         # critical | high | medium | low
effort: medium         # small | medium | large | x-large
tags: [merge-queue, hephaestus] # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-11
closed:                # YYYY-MM-DD. Set this only when state is closed.
owner:                 # The person or agent working on the issue. Cleared when the issue returns to open.
approver:              # The person or agent who accepts the issue when it closes.
bookmark:              # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: []         # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
metadata: {}           # Optional. Arbitrary key-value notes, for example an AI agent session id to resume work with context: {session-id: "abc123"}.
deferred-from:         # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent: "[[merge-queue/merge-queue-agent]]" # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
depends-on: []           # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# Compute the order in which the merge queue merges bookmarks

## Summary

Compute the merge order of a set of bookmarks: dependencies between their issues first, stacked bookmarks in their order, then the bookmark that conflicts with the fewest others, then the oldest. Output the order with the reason for each place, so that a human can check it.

## Context

The rules are in [[agent-delegation#Merge queue]]. The data comes from the repo:

- The issues of a bookmark: the issues whose `bookmark` field is the bookmark, and the `Fixes #<name>` trailers of its changes.
- Dependencies: `depends-on` between those issues (`docs/sisyphus-graph.md` has the edges).
- Stacks: a bookmark whose base is another bookmark of the set.
- Overlap: the files that two bookmarks both change, from the VCS.

A cycle of dependencies between bookmarks is an error that names the issues. This is hephaestus code, for example `hephaestus merge-order <bookmark>...`. It reads the issues with `sisyphus list --json` and `sisyphus graph --json`; it does not need a new sisyphus command.

## Acceptance criteria

- [ ] Given bookmarks, it prints the order and the reason for each place, as text and as JSON.
- [ ] Dependencies and stacks always come first; the overlap only orders the rest.
- [ ] A cycle is an error that names the issues.
- [ ] Tests cover each rule with fake bookmarks.

## Out of scope

- The merge itself: [[merge-queue-agent]].

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

<Complete this section when you close the issue.
- Completed: tell what was done. Link to the changelog version and the decision records, for example `[[0.0.5]]`.
- Abandoned: tell why the work stopped. Link to the issue or decision that replaces it, if one exists.>
