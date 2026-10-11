---
title: "Run the role agents: start them on triggers and give them their work"
state: open            # open | in-progress | closed. Must match the directory of the file.
resolution:            # completed | abandoned. Set this only when state is closed.
priority: medium       # critical | high | medium | low
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
parent: "[[core/delegate-issues-to-agents]]" # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
depends-on: ["[[core/delegation-model]]"] # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# Run the role agents: start them on triggers and give them their work

## Summary

Build the system that holds the role agents: it starts a role agent when a trigger fires (a human request, a close, or a schedule), gives it its work, and keeps the record of each run. Answer Q4 and Q5 of [[agent-delegation#Open questions]].

## Context

The proposal is to start with Claude Code subagents that a human starts, made from `hephaestus/roles/<role>.md`, then a runner on the Claude Agent SDK. Local models only for roles that a task suite shows they can do, as the sisyphus spike `spike-ollama-local-models` did for issue tracking. Split this issue before the work starts: it is x-large.

## Acceptance criteria

- [ ] A decision record answers Q4 and Q5.
- [ ] A human can start each role agent with one command, on a role and a set of issues.
- [ ] Each run is recorded in the issues that it worked on.

## Out of scope

- Hosting the runner as a service.

## Notes

2026-10-11, from the human: the proposal for Q5 of [[agent-delegation]] is right for now: the merge queue starts on a human request (later on a schedule), the technical docs agent on each close, and the other roles on a human request. Q4 is still open.

2026-10-11, from the human: Q4 is correct: AI agents start as Claude Code subagents, then the Agent SDK. hephaestus must also run tool agents, which do not use AI, for large refactors, lints, and other work, to avoid large token spend ([[tool-agents]]).

## Resolution

<Complete this section when you close the issue.
- Completed: tell what was done. Link to the changelog version and the decision records, for example `[[0.0.5]]`.
- Abandoned: tell why the work stopped. Link to the issue or decision that replaces it, if one exists.>
