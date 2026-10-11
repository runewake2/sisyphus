---
title: "Let update change the title of an issue"
state: closed          # open | in-progress | closed. Must match the directory of the file.
resolution: completed  # completed | abandoned. Set this only when state is closed.
priority: medium       # critical | high | medium | low
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

# Let update change the title of an issue

## Summary

No command changes the title of an issue. Names never change, so the title is the only way to correct how an issue is described, and today that needs a hand edit of the frontmatter and of the `# Title` heading.

## Context

Found in a review of sisyphus 0.0.31 on 2026-10-11. `sisyphus update a-one open --title "New title"` fails with `unknown flag: --title`. The title is in two places: the `title` field and the first heading of the body. A hand edit easily changes only one.

## Acceptance criteria

- [ ] `sisyphus update <name> --title "<title>"` sets the `title` field and the first `#` heading of the body to the same text.
- [ ] The same title rules as `new` apply (see [[reject-multiline-titles]]).
- [ ] `sisyphus_update` has a `title` argument.
- [ ] Tests cover both places.

## Out of scope

- Renaming an issue. Names stay permanent.

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

Completed in [[0.0.41]]. `update --title` sets the title field and the first `#` heading, with the title rules of `new`. `sisyphus_update` has a `title` argument.

