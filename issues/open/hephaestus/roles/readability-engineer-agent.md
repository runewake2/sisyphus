---
title: "Add a readability engineer that condenses sprawling changes"
state: open            # open | in-progress | closed. Must match the directory of the file.
resolution:            # completed | abandoned. Set this only when state is closed.
priority: medium       # critical | high | medium | low
effort: medium         # small | medium | large | x-large
tags: [hephaestus, llm] # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-20
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

# Add a readability engineer that condenses sprawling changes

## Summary

Add the AI role readability-engineer: it measures the spread of each change — the distance in lines
between its related edits, the files it touches, and the directories it spans — ranks the sprawling
changes, and proposes options for condensing the code under them, so that one behavior can be edited
from a single screen. On request, it makes one small move with no change of behavior.

## Context

Requested by a human on 2026-10-20: a role that works to make changes condensed, by measuring the
distance of line numbers between related changes, and ranking changes that span multiple files or
directories, or that leave large gaps of unchanged lines between changes. It suggests options for
condensing sprawling code, and strives for code that can be edited from a single screen. The role
file is [[roles/readability-engineer]]. The roles may be removed after a review; a trial run shows
how useful each is.

## Acceptance criteria

- [x] `hephaestus/roles/readability-engineer.md` describes the role, its measure of spread (the
  directories, the files, and the gaps of unchanged lines, against a screen of about 40 lines), and
  its options for condensing: move the helper next to its caller, group what changes together, split
  a file that mixes behaviors, join files that always change together, keep the boundaries.
- [x] Its Claude Code subagent is in `.claude/agents/`.
- [ ] A trial run on a recent bookmark reports its spread as numbers, and proposes at least one
  option to condense, or says why the spread has a reason.
- [ ] A human checks the findings of the run and decides whether the role stays.

## Out of scope

<List related work that this issue does not include. Create a separate issue for each item that someone must do later.>

## Notes

## Resolution

<Complete this section when you close the issue.
- Completed: tell what was done. Link to the changelog version and the decision records, for example `[[0.0.5]]`.
- Abandoned: tell why the work stopped. Link to the issue or decision that replaces it, if one exists.>
