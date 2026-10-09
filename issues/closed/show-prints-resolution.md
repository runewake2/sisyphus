---
title: "Print the resolution in sisyphus show"
state: closed          # open | in-progress | closed. Must match the directory of the file.
resolution: completed  # completed | abandoned. Set this only when state is closed.
priority: high         # critical | high | medium | low
effort: small          # small | medium | large | x-large
tags: [bug, cli]       # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-09
closed: 2026-10-09     # YYYY-MM-DD. Set this only when state is closed.
owner: claude          # The person or agent working on the issue. Cleared when the issue returns to open.
approver:              # The person or agent who accepts the issue when it closes.
bookmark: samw/ai/edit-issue-sections # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: [edit-issue-sections] # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
metadata: {session-id: "e70e78a5-5e37-4f8f-aaa8-6dad55126d53"} # Optional. Arbitrary key-value notes, for example an AI agent session id to resume work with context: {session-id: "abc123"}.
deferred-from:         # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent: "[[issue-manager-gaps]]" # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
depends-on: []           # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# Print the resolution in sisyphus show

## Summary

Make the text output of `sisyphus show` print `resolution`, so that a reader can see whether a close or an abandon worked.

## Context

From the human, 2026-10-09: a test run of Claude as a sisyphus task manager found this gap. After `sisyphus update <name> closed --resolution abandoned`, Claude could not confirm the abandon with `sisyphus show`, because the text output did not print `resolution`. The `--json` output already had it.

The text output also did not print `created` and `closed`.

## Acceptance criteria

- [x] `sisyphus show` prints `resolution`, `created`, and `closed`
- [x] A test shows the resolution of an abandoned issue
- [x] [[sisyphus-show]] shows the new fields

## Out of scope

- A command to edit the `Resolution` section of the body. [[edit-issue-body-command]] covers it.

## Notes

2026-10-09: added the three fields after `state` and `tags`. The `--json` output does not change.

## Resolution

Completed in sisyphus [[0.0.30]] on bookmark samw/ai/edit-issue-sections. It waits for a human merge.
