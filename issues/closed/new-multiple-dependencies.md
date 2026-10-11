---
title: "Let new take more than one depends-on issue"
state: closed          # open | in-progress | closed. Must match the directory of the file.
resolution: completed  # completed | abandoned. Set this only when state is closed.
priority: low          # critical | high | medium | low
effort: small          # small | medium | large | x-large
tags: [feature, cli]   # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-11
closed: 2026-10-11     # YYYY-MM-DD. Set this only when state is closed.
owner: claude          # The person or agent working on the issue. Cleared when the issue returns to open.
approver:              # The person or agent who accepts the issue when it closes.
bookmark: samw/ai/tracker-fixes # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: [tracker-fixes] # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
metadata: {}           # Optional. Arbitrary key-value notes, for example an AI agent session id to resume work with context: {session-id: "abc123"}.
deferred-from:         # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent: "[[issue-manager-gaps]]" # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
depends-on: []           # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# Let new take more than one depends-on issue

## Summary

`sisyphus new --depends-on` takes only one issue. A new issue that waits for two issues needs a second command. Accept a comma-separated list, like `--tags`.

## Context

Found in a review of sisyphus 0.0.31 on 2026-10-11.

```
sisyphus new c-three --depends-on a-one,b-two
# Error: The issue 'a-one,b-two' does not exist ...
```

The frontmatter field is a list, and `sisyphus depends-on` adds one entry at a time.

## Acceptance criteria

- [ ] `new --depends-on a-one,b-two` (and the flag given more than once) adds each issue.
- [ ] Each issue must exist, and the error names the one that does not.
- [ ] `dependsOn` in `sisyphus_new` accepts a list.
- [ ] Tests cover two dependencies and one missing issue.

## Out of scope

Nothing.

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

Completed in [[0.0.43]]. `new --depends-on` takes a comma-separated list or a repeated flag, and checks each issue. `dependsOn` of `sisyphus_new` takes a comma-separated list.

