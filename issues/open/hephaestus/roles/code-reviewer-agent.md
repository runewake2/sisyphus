---
title: "Add a code reviewer agent that reviews the code of a bookmark"
state: open            # open | in-progress | closed. Must match the directory of the file.
resolution:            # completed | abandoned. Set this only when state is closed.
priority: medium       # critical | high | medium | low
effort: large          # small | medium | large | x-large
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

# Add a code reviewer agent that reviews the code of a bookmark

## Summary

Add a code reviewer role: an agent that reviews the code of a bookmark for correctness, tests, and fit, reports each finding with a concrete case, and is normally a required reviewer in a referendum.

## Context

Requested by a human on 2026-10-12, from a review of the missing roles: no role judged the code itself, so a referendum had no reviewer for it. The role file [[roles/code-reviewer|code-reviewer]] defines the role.

## Acceptance criteria

- [x] `hephaestus/roles/code-reviewer.md` describes the role, its checklist, and the points where it must ask a human.
- [ ] A run on a real bookmark writes a review whose findings each have a concrete case.
- [ ] A human checks the review of one run.

## Out of scope

- Changing the code: the implementer or the owner of the bookmark does.

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

<Complete this section when you close the issue.
- Completed: tell what was done. Link to the changelog version and the decision records, for example `[[0.0.5]]`.
- Abandoned: tell why the work stopped. Link to the issue or decision that replaces it, if one exists.>
