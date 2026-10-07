---
title: "Add sisyphus pull, push, and sync on top of plugins"
state: open            # open | in-progress | closed. Must match the directory of the file.
resolution:            # completed | abandoned. Set this only when state is closed.
priority: high         # critical | high | medium | low
effort: medium         # small | medium | large | x-large
tags: [plugins, cli]   # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-07
closed:                # YYYY-MM-DD. Set this only when state is closed.
owner:                 # The person or agent working on the issue. Cleared when the issue returns to open.
approver:              # The person or agent who accepts the issue when it closes.
bookmark:              # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: []         # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
metadata: {}           # Optional. Arbitrary key-value notes, for example an AI agent session id to resume work with context: {session-id: "abc123"}.
deferred-from:         # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent: "[[integration-plugins]]" # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
depends-on: ["[[plugin-runner]]"] # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# Add sisyphus pull, push, and sync on top of plugins

## Summary

Add the commands that use plugins: `sisyphus pull <remote-url>`, `sisyphus push <name>`, and
`sisyphus sync`.

## Context

Part of [[integration-plugins]]; builds on [[plugin-runner]]. sisyphus stays the source of truth
([[github-source-of-truth]]).

- **`pull <remote-url>`**: create a sisyphus issue from a remote one, pinned with `remote` (name
  from `sisyphus slug`). Open question: when the URL is already pinned to an issue, does `pull`
  refresh any fields, or refuse?
- **`push <name>`**: mirror one issue to its remote, creating it if the issue has no `remote` yet
  (with the default plugin) and recording the new URL in `remote`.
- **`sync`**: `push` every issue. Idempotent: a second run with no changes makes no remote changes.
  Open question: are closed issues with no `remote` mirrored, or skipped?
- `--dry-run` on every command that changes anything.
- Each command also becomes a tool in `sisyphus-mcp` ([[sisyphus-mcp-server]]).

## Acceptance criteria

- [ ] `pull`, `push`, and `sync` work through a configured plugin.
- [ ] `push` and `sync` record a newly created remote's URL in `remote`.
- [ ] `sync` is idempotent.
- [ ] `--dry-run` reports the changes and changes nothing, locally or remotely.
- [ ] The open questions above are answered.
- [ ] Available as `sisyphus-mcp` tools.
- [ ] Tested against the fake plugin from [[plugin-runner]].

## Out of scope

- Reading remote edits back into sisyphus automatically.

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

<Complete this section when you close the issue.
- Completed: tell what was done. Link to the changelog version and the decision records, for example `[[0.0.5]]`.
- Abandoned: tell why the work stopped. Link to the issue or decision that replaces it, if one exists.>
