---
title: "Document stdout, stderr, exit codes, and the JSON shapes"
state: closed          # open | in-progress | closed. Must match the directory of the file.
resolution: completed  # completed | abandoned. Set this only when state is closed.
priority: low          # critical | high | medium | low
effort: small          # small | medium | large | x-large
tags: [docs]           # The components that the work touches, for example [widget-scheduler, plan]
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

# Document stdout, stderr, exit codes, and the JSON shapes

## Summary

The commands follow conventions that are not written down in one place: results on stdout, warnings and errors on stderr, exit code 1 for an error and for `lint` problems, and the JSON shapes of `list`, `search`, `show`, `graph`, `links`, and `lint`. Document them, so that scripts can rely on them.

## Context

Found in a review of the CLI, the help text, and the documentation of sisyphus 0.0.67 on 2026-10-11, with a first session in an empty repo. Each command page describes its own output, but there is no page that says, for example, that a
warning never goes to stdout, that `--json` output is stable, or which fields can be missing (empty optional fields
are left out of `show --json`).

## Acceptance criteria

- [ ] A section of [[commands]] (or its own page) describes stdout, stderr, exit codes, and `--json`.
- [ ] Each JSON shape is listed with its fields, or linked from the command pages.
- [ ] The rule for changes to the JSON fields is stated (for example: fields are only added).

## Out of scope

Nothing.

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

Completed in [[0.0.103]]. [[output]] describes the streams, the exit codes, every JSON shape with its fields, and the rule that fields are only added. A test keeps the field lists complete.

