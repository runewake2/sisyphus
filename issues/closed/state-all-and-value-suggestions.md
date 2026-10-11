---
title: "Accept --state all, and suggest the nearest valid value"
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

# Accept --state all, and suggest the nearest valid value

## Summary

`sisyphus list --state all` fails, so listing every issue needs `--state open,in-progress,closed`. Accept `all`, and when an enum value is wrong, suggest the nearest valid one.

## Context

Found in a review of the CLI, the help text, and the documentation of sisyphus 0.0.67 on 2026-10-11, with a first session in an empty repo.

```
$ sisyphus list --state all
Error: Invalid state 'all'. Use one of: open, in-progress, closed.
$ sisyphus update fix-login-bug --priority hi
Error: Invalid priority 'hi'. Use one of: critical, high, medium, low.
```

`oneOf` in `cmd/sisyphus/cli.go` gives the valid values but no suggestion. A short edit distance or a prefix
check (`hi` → `high`, `inprogress` → `in-progress`, `done` → `closed` as a known synonym) makes a typo a one-step
fix.

## Acceptance criteria

- [ ] `--state all` in `list` and `search` means every state.
- [ ] An invalid state, priority, effort, or resolution error adds `Did you mean '<value>'?` when one valid value is near.
- [ ] Tests cover `all` and two suggestions.

## Out of scope

- Accepting the synonyms silently.

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

Completed in [[0.0.70]]. `--state all` works in `list` and `search`, and invalid values get a `Did you mean` suggestion from synonyms, prefixes, or edit distance.

