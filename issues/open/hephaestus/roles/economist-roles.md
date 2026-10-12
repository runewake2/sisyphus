---
title: "Add economist roles that analyze cost against a budget"
state: open            # open | in-progress | closed. Must match the directory of the file.
resolution:            # completed | abandoned. Set this only when state is closed.
priority: medium       # critical | high | medium | low
effort: large          # small | medium | large | x-large
tags: [hephaestus, llm, cost] # The components that the work touches, for example [widget-scheduler, plan]
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

# Add economist roles that analyze cost against a budget

## Summary

Add three AI roles in `hephaestus/roles/economist/`: `economist/tokens` (the tokens and money of AI agents), `economist/cloud` (cloud services such as AWS), and `economist/compute` (CPU, RAM, GPU and VRAM, and storage). Each compares the use with a budget, reports it with evidence, and files issues that propose savings. None changes a resource, a role, or a model.

## Context

Requested by a human on 2026-10-12: an economist who specializes in budget analysis, as two economist agents for AI agent token use and for cloud cost (such as AWS), and a third for raw computer resources: CPU, RAM, GPU, and storage. They are a family of roles in `hephaestus/roles/economist/` ([[0005-roles-of-a-family-live-in-a-directory]]); [[roles/economist/README|The economist roles]] has the rules that they share.

## Acceptance criteria

- [x] The three role files describe the role, its sources, its outputs, and the points where it must ask a human, and the family README has the rules of every economist: budget first, evidence with each number, read and do not spend, ask before a cost, no data about people, and savings with their cost.
- [x] Their Claude Code subagents are in `.claude/agents/`.
- [ ] A run of `economist/compute` on the checks of the merge queue reports the time, CPU, and peak RAM of each, with evidence, and files issues with no duplicates.
- [ ] A run of `economist/tokens` on one referendum with real AI reviewers reports the tokens and cost of each round and gate.
- [ ] `economist/cloud` is checked with read-only access to a real account, when a human gives one.
- [ ] A human checks the report of each run.

## Out of scope

- A budget file or budget command in hephaestus: for now, the budget comes in the data of the task.
- Any action that changes a resource, a plan, or a role.

## Notes

2026-10-12, from the human: economist/compute also optimizes network bandwidth and the messages that the agents send. Its role file now measures network bytes, downloads and their caches, and the traffic of the message queue: messages by kind and task, broadcasts and the workers that each starts, deliveries that came again, and the growth of the log.

## Resolution

<Complete this section when you close the issue.
- Completed: tell what was done. Link to the changelog version and the decision records, for example `[[0.0.5]]`.
- Abandoned: tell why the work stopped. Link to the issue or decision that replaces it, if one exists.>
