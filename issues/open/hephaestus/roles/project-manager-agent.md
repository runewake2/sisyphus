---
title: "Add a project manager agent that consolidates and focuses the issues"
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
parent: "[[hephaestus/core/delegate-issues-to-agents]]" # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
depends-on: ["[[hephaestus/core/delegation-model]]"] # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# Add a project manager agent that consolidates and focuses the issues

## Summary

Add a project manager role: an agent that reads the open issues and proposes how to consolidate and focus them: merge duplicates, split large issues, close done or obsolete ones, set priorities, and keep epics current. A human accepts the proposal.

## Context

[[agent-delegation#The other roles]] describes the role. Examples of its work in this repo: [[issue-manager-gaps]] stayed open with one criterion left until a human moved [[migrate-command]] out of it; [[small-model-friendly-mcp]] has all its sub-issues closed and waits only on a rerun. The agent would find such cases, and propose the change as one commit for a human to accept.

It uses `list`, `search`, `graph`, `claims`, and `lint`, and changes issues only with sisyphus commands.

## Acceptance criteria

- [x] `hephaestus/roles/project-manager.md` describes the role, with the points where it must ask a human.
- [ ] A run on this repo gives a proposal (duplicates, splits, closes, priorities, epics) as one change that a human can accept or reject.
- [ ] It never closes or deletes an issue without that acceptance.

## Out of scope

- Doing the work of the issues.

## Notes

2026-10-12: the role file [[roles/project-manager|project-manager]] defines the role: its charter, inputs, outputs, tools, the points where it asks a human, and its prompt ([[unreleased-define-the-ai-roles]]). The runs need [[delegation-runner]], which runs AI roles.

2026-10-12: the role runs now, as a Claude Code subagent (.claude/agents/project-manager.md) or a worker with a runner ([[run-ai-roles]]). The run of this issue still needs a human to start it, and a human to check its result.

## Resolution

<Complete this section when you close the issue.
- Completed: tell what was done. Link to the changelog version and the decision records, for example `[[0.0.5]]`.
- Abandoned: tell why the work stopped. Link to the issue or decision that replaces it, if one exists.>
