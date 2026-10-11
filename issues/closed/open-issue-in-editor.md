---
title: "Open an issue in the editor of the user"
state: closed          # open | in-progress | closed. Must match the directory of the file.
resolution: completed  # completed | abandoned. Set this only when state is closed.
priority: low          # critical | high | medium | low
effort: small          # small | medium | large | x-large
tags: [ux, cli, feature] # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-11
closed: 2026-10-11     # YYYY-MM-DD. Set this only when state is closed.
owner: claude          # The person or agent working on the issue. Cleared when the issue returns to open.
approver:              # The person or agent who accepts the issue when it closes.
bookmark: samw/ai/ux-fixes # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: [ux-fixes] # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
metadata: {}           # Optional. Arbitrary key-value notes, for example an AI agent session id to resume work with context: {session-id: "abc123"}.
deferred-from:         # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent: "[[cli-and-docs-ux]]" # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
depends-on: []           # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# Open an issue in the editor of the user

## Summary

To write the body of an issue, a person must find its file, or pass the text to `sisyphus edit` on the command line. Add a command that opens the issue file in `$VISUAL` or `$EDITOR`.

## Context

Found in a review of the CLI, the help text, and the documentation of sisyphus 0.0.67 on 2026-10-11, with a first session in an empty repo. `sisyphus edit` is good for agents and scripts, but a person writes a Summary or Notes in an
editor. The file can be in any state directory and subdirectory, so `sisyphus open <name>` (or `edit <name>
--editor`) that finds it and starts the editor saves a lookup. After the editor exits, sisyphus can run the same
checks as `lint` on the file and warn.

## Acceptance criteria

- [ ] `sisyphus open <name>` starts `$VISUAL`, else `$EDITOR`, else a clear error, with the path of the issue.
- [ ] After the editor exits, the file is checked, and problems are warnings.
- [ ] A docs page and a test (with a fake editor) cover it.

## Out of scope

Nothing.

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

Completed in [[0.0.76]]. `open` starts `$VISUAL` or `$EDITOR` on the issue file, and checks the file when the editor exits.

