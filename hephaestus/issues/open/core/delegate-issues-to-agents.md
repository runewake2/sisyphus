---
title: "Delegate sisyphus issues to agents that each have one role"
state: open            # open | in-progress | closed. Must match the directory of the file.
resolution:            # completed | abandoned. Set this only when state is closed.
priority: high         # critical | high | medium | low
effort: x-large        # small | medium | large | x-large
tags: [llm, hephaestus] # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-11
closed:                # YYYY-MM-DD. Set this only when state is closed.
owner:                 # The person or agent working on the issue. Cleared when the issue returns to open.
approver:              # The person or agent who accepts the issue when it closes.
bookmark:              # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: []         # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
metadata: {}           # Optional. Arbitrary key-value notes, for example an AI agent session id to resume work with context: {session-id: "abc123"}.
deferred-from:         # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent:                # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
depends-on: []           # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# Delegate sisyphus issues to agents that each have one role

## Summary

Add a way to delegate sisyphus issues to agents that each have one role, or to a system that holds these agents, and let the agents resolve the issues. Start with a merge queue, then a project manager, an engineer who aligns workstreams, a UX researcher, and a technical docs agent. [[agent-delegation]] is the design document.

## Context

Requested by a human on 2026-10-11. hephaestus is a separate worker manager beside sisyphus: it provisions and manages the agents, and sisyphus stays a tool to interact with issues (`hephaestus/README.md`). Today one agent works on each issue, and a human merges the bookmarks, aligns the workstreams, sets the priorities, and checks the docs. These are roles with clear inputs and outputs. The design document has the rules that all roles follow, a charter for each role, and the open questions Q1-Q7 with proposed answers.

Sub-issues:

- [[delegation-model]]: how an issue gets to a role, and where a role is described. Design first.
- [[merge-queue-agent]]: the first role, with [[renumber-versions-on-merge]] and [[merge-order-planner]].
- [[project-manager-agent]], [[engineer-agent]], [[ux-researcher-agent]], [[technical-docs-agent]]: the other roles.
- [[delegation-runner]]: the system that holds the agents and starts them.
- [[referendum-voting]]: several role agents vote on an implementation, and a modification reruns the vote.
- [[tool-agents]]: agents that do delegated work with a program, not a model, so that it costs no tokens.
- [[message-queue]]: a tool agent that relays messages between workers, and public broadcasts.

## Acceptance criteria

- [ ] All sub-issues are closed or deferred with a reason.
- [ ] A human can delegate an issue to each role, and the role agent resolves it, or stops and asks.
- [ ] The design document has no open question left (Q1-Q7), and each answer is a decision record.
- [ ] hephaestus needs no agent-specific work in sisyphus: only generic issue data.

## Out of scope

- Roles that the human did not ask for. Add them as new sub-issues when they are needed.
- Agents that push or move `main` without a human decision (`AGENTS.md`).

## Notes

2026-10-11, from the human: the referendum ([[referendum-voting]]) is how hephaestus coordinates many concurrent agents.

## Resolution

<Complete this section when you close the issue.
- Completed: tell what was done. Link to the changelog version and the decision records, for example `[[0.0.5]]`.
- Abandoned: tell why the work stopped. Link to the issue or decision that replaces it, if one exists.>
