---
title: "Add \"sisyphus list\" to list and filter issues from the CLI"
state: open            # open | in-progress | closed. Must match the directory of the file.
resolution:            # completed | abandoned. Set this only when state is closed.
priority: high         # critical | high | medium | low
effort: medium         # small | medium | large | x-large
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

# Add "sisyphus list" to list and filter issues from the CLI

## Summary

Add `sisyphus list` to print the issues in `issues/`, filtered by state, priority, tag, owner, or
parent, as a table or JSON. This is the single biggest missing piece of sisyphus as an issue
manager: there is currently no way to answer "what is open", "what is mine", or "what is blocked"
without opening files by hand.

## Context

Part of [[issue-manager-gaps]]. `linksCommand` (`cmd/sisyphus/cli.go`) already has a table/JSON
writer (`writeTable`, and `--json` with `encoding/json`) that `list` can reuse for its own row type.
`repoFiles` and the `issues/<state>/*.md` glob pattern used by `subIssues` (`issues.go`) are the
existing building blocks for walking every issue.

Suggested shape:

```bash
sisyphus list                                   # every open and in-progress issue, state then priority
sisyphus list --state open,in-progress
sisyphus list --priority critical,high
sisyphus list --tags scheduler
sisyphus list --owner samw
sisyphus list --parent issue-manager-gaps       # direct sub-issues of an issue
sisyphus list --json
```

Default output: a table with name, title, state, priority, owner, and tags, one row per issue,
sorted by state (open, in-progress, closed) then priority (critical..low) then name. `closed` issues
are excluded by default (matching how `findIssue`/`issueMatches` already treat state); `--state`
includes them explicitly.

## Acceptance criteria

- [ ] `sisyphus list` with no flags lists every open and in-progress issue.
- [ ] `--state`, `--priority`, `--tags`, `--owner`, and `--parent` each filter the list, and combine
      with AND semantics when given together.
- [ ] `--json` prints the same rows as JSON, one array, like `sisyphus links --json`.
- [ ] Output is sorted deterministically (state, then priority, then name).
- [ ] Covered by tests for each filter and for the default (no-flag) behavior.

## Out of scope

- Free-text search over titles or bodies (a different, simpler feature: full-text grep already works
  fine for that today; revisit only if filters above are not enough).
- A tree/hierarchical view of parent/sub-issue relationships; `--parent` on a flat list covers the
  common case of "children of X" without that complexity.

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

<Complete this section when you close the issue.
- Completed: tell what was done. Link to the changelog version and the decision records, for example [[0.0.5]].
- Abandoned: tell why the work stopped. Link to the issue or decision that replaces it, if one exists.>
