---
title: "toolagent.go mixes the task lifecycle with the referendum review"
state: closed          # open | in-progress | closed. Must match the directory of the file.
resolution: completed  # completed | abandoned. Set this only when state is closed.
priority: medium       # critical | high | medium | low
effort: medium         # small | medium | large | x-large
tags: [hephaestus]     # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-20
closed: 2026-10-20     # YYYY-MM-DD. Set this only when state is closed.
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

# toolagent.go mixes the task lifecycle with the referendum review

## Summary

internal/hephaestus/toolagent/toolagent.go (793 lines) holds two behaviors: the worker task lifecycle (Run, Handle, handle, and their helpers, lines 55-710) and the referendum review of a tool role (Ballot at 713, Review at 731, fileDiff at 783). A change to one behavior must step over the other. Move the review code into its own place, so that the next change to tool-reviewer voting lands in one file of about 90 lines.

Evidence from the readability trial of 2026-10-20, on the five changes that landed that day:
- Change olxlywlx (evict): 6 hunks in toolagent.go, largest gap 87 unchanged lines (a screen is about 40).
- Change uvtlrzlq (share repeated code): 6 hunks in toolagent.go, largest gap 605 unchanged lines, between the import block and the edits to Review and the removal of contains near line 790.
In both changes, toolagent.go had the largest in-file gap of the whole change.

Options, each a proposal:
1. Same package, new file: move Ballot, Review, and fileDiff to internal/hephaestus/toolagent/review.go. No import or caller changes. The next voting change edits one small file.
2. Move to the user: move them into internal/hephaestus/referendum, next to voter.go, which is their main internal caller. toolagent does not import referendum, so no cycle. This moves the exported names toolagent.Review and toolagent.Ballot, which cmd/hephaestus/agent.go also uses: ask a human before this move, per the role rules.

Related: [[hephaestus/core/share-jj-export-import]] also shrinks toolagent.go (exportJJ and importJJ at lines 505-520).

## Context

<Give the background that an agent needs to start the work without other help.
Link to design docs and other issues with wikilinks, for example `[[widget-scheduler]]`,
`[[widget-scheduler#Status API]]`, or `[[<issue-name>]]`. Do not use paths.>

## Acceptance criteria

- [ ] Ballot, Review, and fileDiff are in one file that holds only the review behavior (option 1), or in the referendum package after a human approves the move of the exported names (option 2).
- [ ] No change of behavior: the tests of toolagent, referendum, and cmd/hephaestus pass unchanged.
- [ ] A diff of a later voting change touches only the new file (check at the next such change).

## Out of scope

<List related work that this issue does not include. Create a separate issue for each item that someone must do later.>

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

Done with option 1: Ballot, Review, and fileDiff moved from toolagent.go (793 lines, now 710) to review.go in the same package, with a file comment that names the boundary. No exported name, caller, or behavior changed; all tests pass unchanged. Option 2 (moving the review into package referendum, which moves exported names) stays open for a human to ask for; the places are in this issue.

