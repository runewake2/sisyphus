---
title: "Add an engineer agent that aligns concurrent workstreams"
state: open            # open | in-progress | closed. Must match the directory of the file.
resolution:            # completed | abandoned. Set this only when state is closed.
priority: medium       # critical | high | medium | low
effort: large          # small | medium | large | x-large
tags: [llm, hephaestus] # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-11
closed:                # YYYY-MM-DD. Set this only when state is closed.
owner:                 # The person or agent working on the issue. Cleared when the issue returns to open.
approver:              # The person or agent who accepts the issue when it closes.
bookmark:              # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: []         # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
metadata: {}           # Optional. Arbitrary key-value notes, for example an AI agent session id to resume work with context: {session-id: "abc123"}.
deferred-from:         # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent: "[[core/delegate-issues-to-agents]]" # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
depends-on: ["[[core/delegation-model]]"] # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# Add an engineer agent that aligns concurrent workstreams

## Summary

Add an engineer role: an agent that takes two or more concurrent workstreams (bookmarks, workspaces, or epics), finds where they overlap or disagree, and organizes their work so that all of them go in the same direction.

## Context

[[agent-delegation#The other roles]] describes the role. Concurrent workstreams can change the same files, build the same feature twice, or make decisions that contradict each other, and nobody notices until the merge. The engineer compares the workstreams before that: the files that both change, the issues and decision records of each, and their direction. It proposes one direction: a decision record, changes to the scope of the issues, and an order for [[merge-queue-agent]].

## Acceptance criteria

- [ ] `hephaestus/roles/engineer.md` describes the role.
- [ ] Given two or more workstreams, it reports their overlaps and conflicts of direction, and proposes a decision record and issue changes.
- [ ] A test run on two real bookmarks of this repo, checked by a human.

## Out of scope

- Merging: [[merge-queue-agent]].

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

<Complete this section when you close the issue.
- Completed: tell what was done. Link to the changelog version and the decision records, for example `[[0.0.5]]`.
- Abandoned: tell why the work stopped. Link to the issue or decision that replaces it, if one exists.>
