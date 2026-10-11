---
title: "Add or remove one tag without replacing all tags"
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
parent: "[[commands/issue-manager-gaps]]" # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
depends-on: []           # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# Add or remove one tag without replacing all tags

## Summary

`sisyphus update --tags` replaces all tags. To add or remove one tag, the user must read the current tags and give the full list again. Add flags that add or remove single tags.

## Context

Found in a review of sisyphus 0.0.31 on 2026-10-11. A lost tag is easy to miss, and an agent that wants to add `bug` must first call `show`. Possible flags: `--add-tags` and `--remove-tags`, comma-separated.

## Acceptance criteria

- [ ] `update --add-tags x,y` adds tags that are not there yet, and keeps the order of the others.
- [ ] `update --remove-tags x` removes a tag. Removing a tag that is not there is not an error.
- [ ] `--tags` cannot be used together with the new flags.
- [ ] `sisyphus_update` has the same arguments, and tests cover them.

## Out of scope

Nothing.

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

Completed in [[0.0.42]]. `update --add-tags` and `--remove-tags` change single tags and keep the others. They cannot be used with `--tags`.

