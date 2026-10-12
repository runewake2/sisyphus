---
title: "Add an abstraction engineer that promotes code reuse and clear abstractions"
state: open            # open | in-progress | closed. Must match the directory of the file.
resolution:            # completed | abandoned. Set this only when state is closed.
priority: medium       # critical | high | medium | low
effort: medium         # small | medium | large | x-large
tags: [hephaestus, llm] # The components that the work touches, for example [widget-scheduler, plan]
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

# Add an abstraction engineer that promotes code reuse and clear abstractions

## Summary

Add the AI role abstraction-engineer: it reviews each new change for code that the repo already has, scans the repo for code that repeats, and proposes abstractions as issues, with their boundary, contract, and cost. On request, it makes one small abstraction with no change of behavior.

## Context

Requested by a human on 2026-10-12: an agent whose entire objective is clear abstractions and code reuse. It promotes reuse of code and shared implementation where the boundary makes sense, and builds sensible abstractions when code reuse occurs often. The role file is [[roles/abstraction-engineer]]. The roles may be removed after a review; a trial run shows how useful each is.

## Acceptance criteria

- [x] `hephaestus/roles/abstraction-engineer.md` describes the role and its rules: reuse first, three or more copies, one reason to change, respect the boundaries, clearer and not only shorter, behavior stays.
- [x] Its Claude Code subagent is in `.claude/agents/`.
- [ ] A trial run on `hephaestus/internal` finds the known copies below, and files issues with no duplicates.
- [ ] A human checks the issues of the run and decides whether the role stays.

## Out of scope

<List related work that this issue does not include. Create a separate issue for each item that someone must do later.>

## Notes

2026-10-12, from claude: known copies in hephaestus, for the trial run to find: the scratch git worktree add and remove (mergequeue.Run and referendum.RoleVoter; only the second got the fix for the race of two reviewers), the git committer identity fallback (mergequeue.identity and referendum.identity), tail (toolagent and mergequeue), lines (repo and mergequeue), indent (airole and referendum), the shell of a check (mergequeue.runChecks repeats toolagent.Shell, without the stop of the process group), and toolagent.contains (slices.Contains).

## Resolution

<Complete this section when you close the issue.
- Completed: tell what was done. Link to the changelog version and the decision records, for example `[[0.0.5]]`.
- Abandoned: tell why the work stopped. Link to the issue or decision that replaces it, if one exists.>
