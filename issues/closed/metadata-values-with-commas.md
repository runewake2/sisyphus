---
title: "Keep metadata values that contain a comma"
state: closed          # open | in-progress | closed. Must match the directory of the file.
resolution: completed  # completed | abandoned. Set this only when state is closed.
priority: high         # critical | high | medium | low
effort: small          # small | medium | large | x-large
tags: [bug, cli]       # The components that the work touches, for example [widget-scheduler, plan]
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

# Keep metadata values that contain a comma

## Summary

sisyphus writes a metadata value that contains a comma, but reads it back wrong, and the next metadata change writes the wrong value back to the file. The data is lost without a message.

## Context

Found in a review of sisyphus 0.0.45 on 2026-10-11.

```
sisyphus new meta-test
sisyphus update meta-test --metadata "note=x, y"
grep '^metadata' issues/open/meta-test.md    # metadata: {note: "x, y"}      (correct)
sisyphus show meta-test --json               # "metadata": {"note": "\"x"}   (wrong)
sisyphus update meta-test --metadata session=abc
grep '^metadata' issues/open/meta-test.md    # metadata: {note: "\"x", session: "abc"}   (" y" is gone)
```

`formatMetadata` in `cmd/sisyphus/issues.go` quotes each value, but `parseMetadata` splits the map on every comma,
also inside a quoted value. Its comment says that "a value cannot contain a comma", but nothing stops one: `new`,
`update`, and `sisyphus_new`/`sisyphus_update` accept it. An agent that stores a short note in metadata, the use
that the template suggests, hits this easily. [[lint-issues-command]] does not report it.

## Acceptance criteria

- [ ] `parseMetadata` reads a quoted value with commas (and with `:` and escaped quotes) as one value, so that a write and a read give the same map.
- [ ] A metadata update keeps the other values unchanged.
- [ ] Tests cover a value with a comma, a colon, and a double quote, through `new`, `update`, and `show --json`.

## Out of scope

- Repair of files that already hold a corrupted value.

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

Completed in [[0.0.46]]. `parseMetadata` splits only on commas outside quotes, so a value with commas, colons, or quotes reads back as written.

