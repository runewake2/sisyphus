---
title: "referendum reuses the prompt indent of airole"
state: open            # open | in-progress | closed. Must match the directory of the file.
resolution:            # completed | abandoned. Set this only when state is closed.
priority: low          # critical | high | medium | low
effort: small          # small | medium | large | x-large
tags: [hephaestus]     # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-20
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

# referendum reuses the prompt indent of airole

## Summary

Two identical functions indent quoted text in the prompt of an AI role with four spaces, and show "(no text)" for empty text:

- internal/hephaestus/airole/airole.go:95-100 (indent), used at airole.go:73.
- internal/hephaestus/referendum/prompt.go:101-106 (indent), used at prompt.go:57 and 82.

Both format the same thing for the same reader (the prompt of a model), so they must stay the same: when one changes how quoted text looks, the other must follow. referendum already imports airole (prompt.go builds on it), so the reuse is one export: airole.Indent, and the copy in prompt.go goes.

What it removes: one copy of a prompt format. What it costs: one exported name in airole; nothing else — no new import, no behavior change.

## Context

<Give the background that an agent needs to start the work without other help.
Link to design docs and other issues with wikilinks, for example `[[widget-scheduler]]`,
`[[widget-scheduler#Status API]]`, or `[[<issue-name>]]`. Do not use paths.>

## Acceptance criteria

- [ ] airole exports Indent, with its doc comment; referendum/prompt.go uses it and has no indent of its own.
- [ ] The prompts do not change (the existing prompt tests pass unchanged).

## Out of scope

<List related work that this issue does not include. Create a separate issue for each item that someone must do later.>

## Notes

2026-10-20, from the engineer: priority lowered to low, to match [[share-ago-format]]: both are small dedups of an identical function with no fix carried. No dependency: airole and referendum are not touched by the evict work or by the other moves; it can land at any time.

## Resolution

<Complete this section when you close the issue.
- Completed: tell what was done. Link to the changelog version and the decision records, for example `[[0.0.5]]`.
- Abandoned: tell why the work stopped. Link to the issue or decision that replaces it, if one exists.>
