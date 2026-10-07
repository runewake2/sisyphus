---
title: "Let \"sisyphus update\" change priority, effort, and tags on an existing issue"
state: open            # open | in-progress | closed. Must match the directory of the file.
resolution:            # completed | abandoned. Set this only when state is closed.
priority: medium       # critical | high | medium | low
effort: small          # small | medium | large | x-large
tags: [cli]            # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-07
closed:                # YYYY-MM-DD. Set this only when state is closed.
owner:                 # The person or agent working on the issue. Cleared when the issue returns to open.
approver:              # The person or agent who accepts the issue when it closes.
bookmark:              # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: []         # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
agent-session:         # AI agent session id of the current agent working on the issue, if available.
deferred-from:         # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent: "[[issue-manager-gaps]]" # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
---

# Let "sisyphus update" change priority, effort, and tags on an existing issue

## Summary

`sisyphus new` can set `priority`, `effort`, and `tags`, but `sisyphus update` cannot change them
afterward. Add `--priority`, `--effort`, and `--tags` flags to `update` so reprioritizing or
retagging an issue does not mean hand-editing its frontmatter.

## Context

Part of [[issue-manager-gaps]]. `updateIssue` and `updateOptions` (`cmd/sisyphus/issues.go`) already
follow the optional-pointer pattern used for `owner`/`approver`/`bookmark`: a flag only changes the
field when it is given (`cmd.Flags().Changed(...)`, via the existing `optional` helper in
`cli.go`). Adding `priority`, `effort`, and `tags` the same way is a small, mechanical extension of
that pattern, not a new one. `tags` should reuse the same comma-split and square-bracket formatting
`newIssue` already uses, ideally shared rather than duplicated.

Priority and effort must still validate against `priorities`/`efforts` with `oneOf`, the same as
`new` does, when the flag is given.

## Acceptance criteria

- [ ] `sisyphus update <name> <state> --priority <p>` changes `priority` and validates it against
      the same list `new --priority` uses.
- [ ] `sisyphus update <name> <state> --effort <e>` changes `effort`, validated the same way.
- [ ] `sisyphus update <name> <state> --tags <t>` replaces `tags`, using the same parsing as
      `new --tags`.
- [ ] None of the three flags is required; omitting them leaves the existing value untouched, same
      as every other optional `update` flag today.
- [ ] Covered by tests for each flag, and for state-change calls that omit them.

## Out of scope

- Changing `title` after creation (the name is the permanent id; retitling is a smaller, separate
  concern if it is ever needed).
- A generic "set any field" command; this issue only covers the three fields `new` can already set.

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

<Complete this section when you close the issue.
- Completed: tell what was done. Link to the changelog version and the decision records, for example [[0.0.5]].
- Abandoned: tell why the work stopped. Link to the issue or decision that replaces it, if one exists.>
