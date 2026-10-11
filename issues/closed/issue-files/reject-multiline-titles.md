---
title: "Reject titles with line breaks or control characters"
state: closed          # open | in-progress | closed. Must match the directory of the file.
resolution: completed  # completed | abandoned. Set this only when state is closed.
priority: medium       # critical | high | medium | low
effort: small          # small | medium | large | x-large
tags: [bug, cli]       # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-11
closed: 2026-10-11     # YYYY-MM-DD. Set this only when state is closed.
owner: claude          # The person or agent working on the issue. Cleared when the issue returns to open.
approver:              # The person or agent who accepts the issue when it closes.
bookmark: samw/ai/tracker-fixes # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: [tracker-fixes] # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
metadata: {}           # Optional. Arbitrary key-value notes, for example an AI agent session id to resume work with context: {session-id: "abc123"}.
deferred-from:         # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent:                # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
depends-on: []           # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# Reject titles with line breaks or control characters

## Summary

`sisyphus new --title` accepts a line break (and other control characters) in the title. The frontmatter stays valid, but the `# Title` heading of the body splits over two lines and the `list` table breaks.

## Context

Found in a review of sisyphus 0.0.31 on 2026-10-11.

```
sisyphus new nl-title --title $'line1\nline2'
```

The frontmatter gets `title: "line1\nline2"`. The body heading becomes `# line1` followed by a loose `line2` line. `sisyphus list` prints `line2` on its own row, which breaks the table. An agent that builds a title from multi-line text can cause this.

## Acceptance criteria

- [ ] `new` (and any later command that sets a title, see [[retitle-issue]]) rejects a title that contains a line break or another control character, with a clear error.
- [ ] Tests cover `\n`, `\r`, and `\t`.

## Out of scope

- Repair of existing titles. [[lint-issues-command]] can report them.

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

Completed in [[0.0.34]]. `new --title` rejects an empty title and any control character. The check is in `validateTitle`, for reuse by [[retitle-issue]].

