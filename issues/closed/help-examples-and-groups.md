---
title: "Add examples to every command help, and group the commands"
state: closed          # open | in-progress | closed. Must match the directory of the file.
resolution: completed  # completed | abandoned. Set this only when state is closed.
priority: medium       # critical | high | medium | low
effort: small          # small | medium | large | x-large
tags: [ux, cli, docs]  # The components that the work touches, for example [widget-scheduler, plan]
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

# Add examples to every command help, and group the commands

## Summary

Only `sisyphus slug` has an example in its help, and the root help lists the commands in alphabetical order with `help` among them. Add an Example section to every command, group the root help by purpose, and give the root a clear description.

## Context

Found in a review of the CLI, the help text, and the documentation of sisyphus 0.0.67 on 2026-10-11, with a first session in an empty repo. The pages in `docs/` have good examples, but `--help` has none (cobra's `Example` field is not used).
The root description is "Manage the issues in issues/ and resolve wikilinks in the repo." cobra command groups
(`AddGroup`, `GroupID`) can show, for example:

- Work on issues: new, update, edit, show
- Find issues: list, search, graph
- Link issues: parent, depends-on, remote
- Check the repo: lint, links, resolve, slug
- Set up: init

Several short descriptions are long or mention `<name>`; they can follow one pattern.

## Acceptance criteria

- [ ] Every command has an `Example` with 2-4 lines that also appear on its docs page.
- [ ] The root help groups the commands, and has a description of what sisyphus is.
- [ ] Short descriptions fit one line of 80 characters.

## Out of scope

Nothing.

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

Completed in [[0.0.79]]. Every command has 2-4 examples in its help, which a test keeps the same as its docs page, and the root help describes sisyphus and groups the commands. Every command line of the root help fits in 80 characters.

