---
title: "Read issue files that start with a UTF-8 byte order mark"
state: closed          # open | in-progress | closed. Must match the directory of the file.
resolution: completed  # completed | abandoned. Set this only when state is closed.
priority: medium       # critical | high | medium | low
effort: small          # small | medium | large | x-large
tags: [bug, cli, windows] # The components that the work touches, for example [widget-scheduler, plan]
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

# Read issue files that start with a UTF-8 byte order mark

## Summary

An issue file that starts with a UTF-8 byte order mark (BOM) cannot be read by any command. The error says that the file does not start with YAML frontmatter, although it does when you look at it.

## Context

Found in a review of sisyphus 0.0.45 on 2026-10-11.

```
sisyphus new bom-test
printf '\xef\xbb\xbf' | cat - issues/open/bom-test.md > x && mv x issues/open/bom-test.md
sisyphus show bom-test      # Error: The file does not start with YAML frontmatter ('---').
sisyphus update bom-test in-progress --bookmark b   # the same error
```

`parseDocument` in `cmd/sisyphus/document.go` compares the first line with `---`, and the BOM makes the first line
`\ufeff---`. Some Windows editors write a BOM, and Obsidian and git read such files without a problem. `list` and
`lint` report the file as unreadable, with the same misleading message.

## Acceptance criteria

- [ ] sisyphus reads a file that starts with a UTF-8 BOM like the same file without it.
- [ ] A command that writes the file keeps the BOM, or removes it consistently; decide which, and record it.
- [ ] A test covers `show`, `update`, and `lint` for a file with a BOM.

## Out of scope

- Other encodings than UTF-8.

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

Completed in [[0.0.47]]. sisyphus reads a leading UTF-8 BOM and writes it back. See [[0016-write-files-back-as-they-were-read]].

