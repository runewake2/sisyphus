---
title: "Keep the line endings of an issue file when sisyphus edits it"
state: open            # open | in-progress | closed. Must match the directory of the file.
resolution:            # completed | abandoned. Set this only when state is closed.
priority: low          # critical | high | medium | low
effort: small          # small | medium | large | x-large
tags: [bug, cli, windows] # The components that the work touches, for example [widget-scheduler, plan]
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

# Keep the line endings of an issue file when sisyphus edits it

## Summary

When sisyphus changes an issue file with CRLF line endings, it writes the whole file back with LF. One field change then shows as a change of every line.

## Context

Found in a review of sisyphus 0.0.45 on 2026-10-11.

```
sisyphus new crlf-test
sed -i 's/$/\r/' issues/open/crlf-test.md
sisyphus update crlf-test --priority high
grep -c $'\r' issues/open/crlf-test.md    # 0: every line changed
```

`splitLines` in `cmd/sisyphus/document.go` removes `\r`, and `render` joins with `\n`. The design of `document`
is to edit the frontmatter line by line so that the rest of the file stays as it is. A repo on Windows without
`core.autocrlf`, or with `* text eol=crlf`, gets a large, noisy diff.

## Acceptance criteria

- [ ] A command that writes an issue keeps the line endings that the file had (CRLF or LF).
- [ ] A new issue uses the line endings of `issues/TEMPLATE.md`.
- [ ] A test covers an update of a CRLF file.

## Out of scope

- Files with mixed line endings: use the line ending of the first line.

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

<Complete this section when you close the issue.
- Completed: tell what was done. Link to the changelog version and the decision records, for example `[[0.0.5]]`.
- Abandoned: tell why the work stopped. Link to the issue or decision that replaces it, if one exists.>
