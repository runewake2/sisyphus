---
title: "Do not silently drop owner and bookmark when an issue is open"
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

# Do not silently drop owner and bookmark when an issue is open

## Summary

`sisyphus update <name> open --owner carol` exits 0 and prints the path, but the issue has no owner afterwards. sisyphus clears `owner` and `bookmark` when an issue is open, but it does this silently even when the user asked to set them.

## Context

Found in a review of sisyphus 0.0.31 on 2026-10-11.

```
sisyphus update dep-user open --owner carol   # exit 0, no warning
grep '^owner' issues/open/dep-user.md          # owner: (empty)
```

The rule that `owner` and `bookmark` are cleared when an issue returns to open is in [[CONTRIBUTING#Frontmatter]]. The defect is only that a value the user gave is thrown away with no message.

## Acceptance criteria

- [ ] When the new state is `open` and `--owner` or `--bookmark` has a value, `update` fails with an error that says these fields are only kept for in-progress and closed issues, or it warns and says that it ignored them.
- [ ] The same applies to `sisyphus new --state open --owner ...`.
- [ ] Tests cover both commands.

## Out of scope

- Changing the rule that open issues have no owner.

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

Completed in [[0.0.37]]. `new` and `update` fail for a non-empty `--owner` or `--bookmark` when the state is open. See [[0011-refuse-owner-on-open-issues]].

