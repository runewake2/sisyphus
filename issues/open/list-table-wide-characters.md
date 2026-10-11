---
title: "Align the list table for wide characters"
state: open            # open | in-progress | closed. Must match the directory of the file.
resolution:            # completed | abandoned. Set this only when state is closed.
priority: low          # critical | high | medium | low
effort: small          # small | medium | large | x-large
tags: [bug, cli]       # The components that the work touches, for example [widget-scheduler, plan]
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

# Align the list table for wide characters

## Summary

The text table of `sisyphus list` does not align its columns when a title has wide characters, for example CJK text. The graph drawing already measures width with `go-runewidth`; the table does not.

## Context

Found in a review of sisyphus 0.0.31 on 2026-10-11.

```
sisyphus new uni-code --title "Fix café ✓ 日本"
sisyphus list
```

The columns after the title move to the left on that row, because each CJK character takes two terminal cells but is counted as one. `cmd/sisyphus/canvas.go` uses `runewidth.RuneWidth`.

## Acceptance criteria

- [ ] `list` pads columns by display width, not by the number of runes.
- [ ] A test with a CJK title checks the alignment.

## Out of scope

Nothing.

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

<Complete this section when you close the issue.
- Completed: tell what was done. Link to the changelog version and the decision records, for example `[[0.0.5]]`.
- Abandoned: tell why the work stopped. Link to the issue or decision that replaces it, if one exists.>
