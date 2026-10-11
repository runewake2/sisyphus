---
title: "Give a way to undo a close that was a mistake"
state: closed          # open | in-progress | closed. Must match the directory of the file.
resolution: completed  # completed | abandoned. Set this only when state is closed.
priority: low          # critical | high | medium | low
effort: small          # small | medium | large | x-large
tags: [feature, cli]   # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-11
closed: 2026-10-11     # YYYY-MM-DD. Set this only when state is closed.
owner: claude          # The person or agent working on the issue. Cleared when the issue returns to open.
approver:              # The person or agent who accepts the issue when it closes.
bookmark: samw/ai/tracker-fixes-2 # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: [tracker-fixes-2] # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
metadata: {}           # Optional. Arbitrary key-value notes, for example an AI agent session id to resume work with context: {session-id: "abc123"}.
deferred-from:         # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent:                # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
depends-on: []           # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# Give a way to undo a close that was a mistake

## Summary

sisyphus refuses to reopen a closed issue, so an issue that someone closed by mistake (for example an agent that closed the wrong issue) can only be fixed by moving the file by hand.

## Context

Found in a review of sisyphus 0.0.45 on 2026-10-11.

```
sisyphus update meta-test closed
sisyphus update meta-test open   # Error: ... Do not reopen a closed issue. Create a new issue and link to [[meta-test]].
```

The rule against reopening is right for work that was done and later needs more work. But a close that was a mistake
is not that case: a new issue would lose the history, the links, and the sub-issues of the old one, and a hand move
must also clear `resolution` and `closed`. In the Ollama spike ([[ollama-local-models]]), models changed the wrong
issue several times.

## Acceptance criteria

- [ ] A clear, explicit way undoes a close, for example `sisyphus update <name> open --undo-close`. It clears `resolution` and `closed`, and moves the file back.
- [ ] The normal error message for reopening mentions it.
- [ ] A decision record says when to undo a close and when to create a new issue.
- [ ] Tests cover it.

## Out of scope

- Changing the rule against reopening for completed work.

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

Completed in [[0.0.57]]. `update --reopen` moves a closed issue back, clears its resolution and closed date, and notes the reopen in Notes. See [[0018-undo-a-mistaken-close]].

