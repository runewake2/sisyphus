---
title: "Give every command page the same sections"
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
parent: "[[ux/cli-and-docs-ux]]" # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
depends-on: []           # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# Give every command page the same sections

## Summary

The command pages in `docs/` have different sections: some have a Flags table, some put the flags only in the usage line (`links`, `parent`, `remote`, `depends-on`, `slug`), `lint` and `edit` have a flag table without a Flags heading, and `graph` calls its table Options. Give every page the same sections.

## Context

Found in a review of the CLI, the help text, and the documentation of sisyphus 0.0.67 on 2026-10-11, with a first session in an empty repo. Sections per page today, for example: `sisyphus-new`: Usage, Flags, What it does, Output, Examples;
`sisyphus-links`: Usage, What it checks, Output, Examples (the `--json` and `--absolute` flags are only in the
usage line); `sisyphus-graph`: Usage, What it does, Output, Options, Color, Examples. The flags in `--help` and in
the pages agree, so this is about the shape, not the content. A test could compare the flags of each command with
its page, as this review did by hand.

## Acceptance criteria

- [ ] Every command page has Usage, Flags (a table, or "None"), Behavior, Output, and Examples, in that order.
- [ ] A test checks that every flag in `--help` appears in the Flags table of the page.

## Out of scope

Nothing.

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

Completed in [[0.0.80]]. Every command page has Usage, Flags, Behavior, Output, and Examples, and a test checks the sections and that every flag is in the Flags table.

