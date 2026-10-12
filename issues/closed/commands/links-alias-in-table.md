---
title: "Report wikilink aliases in Markdown table rows"
state: closed          # open | in-progress | closed. Must match the directory of the file.
resolution: completed  # completed | abandoned. Set this only when state is closed.
priority: medium       # critical | high | medium | low
effort: small          # small | medium | large | x-large
tags: [cli, links]     # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-12
closed: 2026-10-12     # YYYY-MM-DD. Set this only when state is closed.
owner: claude          # The person or agent working on the issue. Cleared when the issue returns to open.
approver:              # The person or agent who accepts the issue when it closes.
bookmark: samw/ai/ux-fixes # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: [ux-fixes] # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
metadata: {}           # Optional. Arbitrary key-value notes, for example an AI agent session id to resume work with context: {session-id: "abc123"}.
deferred-from:         # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent:                # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
depends-on: []           # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# Report wikilink aliases in Markdown table rows

## Summary

sisyphus links and sisyphus lint report a wikilink with an alias in a table row, as the status alias-in-table, so that no table of the repo breaks.

## Context

Requested by a human on 2026-10-12. Prettier, which formatter/markdown runs, split the table cells of [[agent-delegation]] at the | of aliased wikilinks, because GitHub-flavored Markdown reads that | as the end of a cell. The human decided: avoid aliases in tables.

## Acceptance criteria

- [x] `sisyphus links` gives the status `alias-in-table` to a link with an alias in a table row, and `ok` to the same link outside a table.
- [x] `sisyphus lint` reports it, with how to fix it.
- [x] The tables of the repo have no alias: the government table of [[agent-delegation]] links without one.
- [x] [[CONTRIBUTING#Links between documents]], [[sisyphus-links]], and [[output]] describe the rule and the status.

## Out of scope

<List related work that this issue does not include. Create a separate issue for each item that someone must do later.>

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

Completed: sisyphus links and lint report a wikilink with an alias in a table row (alias-in-table), and the tables of the repo have no alias. See [[unreleased-links-alias-in-table]].

