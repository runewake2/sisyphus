---
title: "Pass free text after -- in sisyphus-mcp"
state: closed          # open | in-progress | closed. Must match the directory of the file.
resolution: completed  # completed | abandoned. Set this only when state is closed.
priority: high         # critical | high | medium | low
effort: small          # small | medium | large | x-large
tags: [bug, sisyphus-mcp] # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-11
closed: 2026-10-11     # YYYY-MM-DD. Set this only when state is closed.
owner: claude          # The person or agent working on the issue. Cleared when the issue returns to open.
approver:              # The person or agent who accepts the issue when it closes.
bookmark: samw/ai/tracker-fixes # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: [tracker-fixes] # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
metadata: {}           # Optional. Arbitrary key-value notes, for example an AI agent session id to resume work with context: {session-id: "abc123"}.
deferred-from:         # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent:                # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
depends-on: []           # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# Pass free text after -- in sisyphus-mcp

## Summary

sisyphus-mcp passes free text to the sisyphus command as a plain argument. When the text starts with `-`, the command reads it as a flag and fails. This breaks `sisyphus_edit` for every Markdown list, so an agent cannot write acceptance criteria or notes as a list. `sisyphus_search` and `sisyphus_slug` fail in the same way.

## Context

Found in a review of sisyphus 0.0.31 on 2026-10-11.

```
sisyphus_edit   {"name":"a-one","section":"Notes","text":"- [ ] a list item"}
  -> Error: unknown shorthand flag: ' ' in - [ ] a list item
sisyphus_search {"query":"-flag-like"}
  -> Error: unknown shorthand flag: 'f' in -flag-like
sisyphus_slug   {"text":"-dash first words"}
  -> Error: unknown shorthand flag: 'd' in -dash first words
```

`cmd/sisyphus-mcp/tools.go` builds the arguments without a `--` separator, for example `args := []string{"edit", in.Name, in.Section, in.Text}` (line 341) and `runSisyphus(in.Dir, "slug", in.Text)` (line 356). On the CLI, `sisyphus edit <name> <section> -- "- item"` works, but an MCP client cannot add the `--`.

The same problem can hit a human on the CLI: `sisyphus edit x "Acceptance criteria" "- [ ] y"` fails with the same message, and the help text of `edit` does not mention `--`.

## Acceptance criteria

- [ ] Every sisyphus-mcp tool puts `--` before free-text positional arguments (edit text, search query, slug text, and any others), after all flags.
- [ ] Tests in `cmd/sisyphus-mcp` cover text that starts with `-` for `sisyphus_edit`, `sisyphus_search`, and `sisyphus_slug`.
- [ ] [[sisyphus-edit]] tells CLI users to use `--` (or `-` for standard input) for text that starts with `-`.

## Out of scope

Nothing.

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

Completed in [[0.0.32]]. Every sisyphus-mcp tool puts its positional arguments after `--` and passes each flag as one `--flag=value` argument. [[sisyphus-edit]] tells CLI users to use `--`.

