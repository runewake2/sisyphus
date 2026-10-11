---
title: "Fit the list table to the terminal width"
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
parent: "[[ux/cli-and-docs-ux]]" # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
depends-on: []           # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# Fit the list table to the terminal width

## Summary

The `list` table is as wide as its longest title, so on a normal terminal long titles wrap and the table breaks up. Shorten the title column to fit the terminal, and keep full output when stdout is not a terminal.

## Context

Found in a review of the CLI, the help text, and the documentation of sisyphus 0.0.67 on 2026-10-11, with a first session in an empty repo. `writeTable` in `cmd/sisyphus/cli.go` pads every column to its widest cell and has no width
limit. In this repo, titles are up to 112 characters, and the widest row of `list --state open,in-progress,closed` is 211
characters. `graph`
already checks whether stdout is a terminal for its colors (`useColor`).

## Acceptance criteria

- [ ] On a terminal, `list` shortens the title column with `…` so that rows fit the terminal width.
- [ ] When stdout is not a terminal, or with `--wide`, titles are not shortened.
- [ ] `--json` does not change.
- [ ] Tests cover a narrow width.

## Out of scope

- Colors in the table. They can come later, with the same `--color` flag as `graph`.

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

Completed in [[0.0.75]]. On a terminal, `list` shortens titles with `…` to fit, down to 10 characters; pipes and `--wide` get full titles.

