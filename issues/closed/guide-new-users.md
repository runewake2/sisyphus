---
title: "Point new users to the next step after init and when no repo is found"
state: closed          # open | in-progress | closed. Must match the directory of the file.
resolution: completed  # completed | abandoned. Set this only when state is closed.
priority: low          # critical | high | medium | low
effort: small          # small | medium | large | x-large
tags: [ux, cli]        # The components that the work touches, for example [widget-scheduler, plan]
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

# Point new users to the next step after init and when no repo is found

## Summary

A first command in a new repo fails with "Cannot find the repo root" and no hint about `init`, and `init` prints four paths and no next step. Tell the user what to do next in both places.

## Context

Found in a review of the CLI, the help text, and the documentation of sisyphus 0.0.67 on 2026-10-11, with a first session in an empty repo.

```
$ sisyphus list
Error: Cannot find the repo root. Run sisyphus in a directory below the one that contains issues/TEMPLATE.md.
$ sisyphus init
issues/TEMPLATE.md
issues/closed/.gitkeep
...
```

A user who does not know about `init` cannot tell from the error how to set up a repo.

## Acceptance criteria

- [ ] The repo-not-found error ends with "To set up issue tracking here, run: sisyphus init".
- [ ] `init` writes a next step to stderr, for example `Next: sisyphus new <name> --title "..."`, so that stdout stays the list of paths.
- [ ] Tests cover both messages.

## Out of scope

Nothing.

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

Completed in [[0.0.69]]. The repo-not-found error names `sisyphus init`, and `init` writes the next step to stderr.

