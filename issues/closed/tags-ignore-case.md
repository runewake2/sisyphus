---
title: "Match and store tags without regard to case"
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

# Match and store tags without regard to case

## Summary

Tags are compared and stored with their case, so `--tags Bug` and `list --tags bug` do not match. The built-in tags are `bug` and `feature`, so a tag `Bug` silently drops out of a filter for bugs.

## Context

Found in a review of sisyphus 0.0.45 on 2026-10-11.

```
sisyphus new tagcase-test --tags Bug
sisyphus list --tags bug     # empty
```

`anyOf` in `cmd/sisyphus/list.go` compares with `==`. Obsidian treats tags without regard to case.

## Acceptance criteria

- [ ] `list --tags` and `search --tags` match tags without regard to case.
- [ ] `new`, `update --tags`, and `update --add-tags` store tags in lowercase, or keep the case and compare without it; decide, and record it.
- [ ] Tests cover both.

## Out of scope

Nothing.

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

Completed in [[0.0.55]]. Tags are written in lowercase, and filters and tag changes compare without regard to case. See [[0017-tags-are-lowercase]].

