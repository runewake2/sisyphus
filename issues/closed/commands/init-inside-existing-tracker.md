---
title: "Warn when init runs below an existing issue tracker"
state: closed          # open | in-progress | closed. Must match the directory of the file.
resolution: completed  # completed | abandoned. Set this only when state is closed.
priority: medium       # critical | high | medium | low
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

# Warn when init runs below an existing issue tracker

## Summary

`sisyphus init` in a subdirectory of a repo that already has `issues/TEMPLATE.md` makes a second, empty tracker with no warning. Every command that runs in that subdirectory then uses the empty tracker, so the existing issues seem to be gone.

## Context

Found in a review of sisyphus 0.0.45 on 2026-10-11.

```
cd repo/sub/pkg
sisyphus init        # writes sub/pkg/issues/... with exit 0
sisyphus list        # an empty table: the issues of repo/ are not found
```

`findRoot` searches up from the current directory and stops at the first `issues/TEMPLATE.md`, so the nested
tracker hides the real one. `init` takes `--dir`, which defaults to `.`, so a user who wanted to set up the repo
from a subdirectory gets this result.

## Acceptance criteria

- [ ] `init` fails when a directory above the target directory already has `issues/TEMPLATE.md`, and names it.
- [ ] A flag (for example `--nested`) allows a nested tracker on purpose.
- [ ] Tests cover both cases.

## Out of scope

Nothing.

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

Completed in [[0.0.50]]. `init` fails below an existing tracker, and `--nested` makes a second tracker on purpose.

