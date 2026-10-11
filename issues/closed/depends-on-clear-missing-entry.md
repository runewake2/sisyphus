---
title: "Say when depends-on --clear removes nothing"
state: closed          # open | in-progress | closed. Must match the directory of the file.
resolution: completed  # completed | abandoned. Set this only when state is closed.
priority: low          # critical | high | medium | low
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

# Say when depends-on --clear removes nothing

## Summary

`sisyphus depends-on <name> <blocking-issue> --clear` exits 0 and prints the path of the issue even when the issue does not depend on `<blocking-issue>`. The user thinks that a dependency was removed.

## Context

Found in a review of sisyphus 0.0.45 on 2026-10-11.

```
sisyphus new dep-a
sisyphus depends-on dep-a no-such-issue --clear    # issues/open/dep-a.md, exit 0
```

`setDependsOn` in `cmd/sisyphus/issues.go` deletes matching entries and writes the file, also when nothing
matched. A typo in the name of the dependency is not visible.

## Acceptance criteria

- [ ] `depends-on <name> <blocking-issue> --clear` warns, and names the dependencies that the issue has, when `<blocking-issue>` is not one of them. The file does not change.
- [ ] A test covers it.

## Out of scope

Nothing.

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

Completed in [[0.0.53]]. `depends-on --clear` warns and lists the real dependencies when the named one is not there, and does not write the file.

