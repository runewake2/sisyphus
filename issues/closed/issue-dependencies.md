---
title: "Add explicit depends-on/blocks relationships between issues"
state: closed          # open | in-progress | closed. Must match the directory of the file.
resolution: completed  # completed | abandoned. Set this only when state is closed.
priority: high         # critical | high | medium | low
effort: medium         # small | medium | large | x-large
tags: [cli, schema]    # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-07
closed: 2026-10-07     # YYYY-MM-DD. Set this only when state is closed.
owner: claude          # The person or agent working on the issue. Cleared when the issue returns to open.
approver: samw         # The person or agent who accepts the issue when it closes.
bookmark: samw/ai/issue-dependencies # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: [sisyphus-issue-dependencies] # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
metadata: {session-id: "68fa9a09-fe4e-4ae8-9917-086c3313c5be"} # Optional. Arbitrary key-value notes, for example an AI agent session id to resume work with context: {session-id: "abc123"}.
deferred-from:         # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent: "[[issue-manager-gaps]]" # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
---

# Add explicit depends-on/blocks relationships between issues

## Summary

Add a `depends-on` field (and its inverse, derived rather than stored) so one issue can say it is
blocked by another unrelated issue. `parent` already expresses "is a sub-issue of"; this is a
different relationship: two issues at the same level, where one cannot start until the other
closes.

## Context

Part of [[issue-manager-gaps]]. This is the one gap that was already anticipated: the test fixture
and documentation example name `explicit-step-dependencies` (see `cmd/sisyphus/new_test.go`,
`CONTRIBUTING.md`) was carried over from before this epic existed, as a placeholder for exactly this
feature.

Follow the existing `parent` field's shape and rules (`setParent`, `parentValueFor` in
`cmd/sisyphus/issues.go`):

- `depends-on` holds a list of quoted wikilinks, for example `depends-on: ["[[faster-startup]]"]`,
  since an issue can be blocked by more than one other issue (unlike `parent`, which is singular).
- Every referenced issue must exist, the way `parentValueFor` checks today.
- An issue cannot depend on itself or on a cycle of issues that depend on it, mirroring the
  ancestor-walk cycle check `parentValueFor` already does for `parent`.
- Closing an issue that other open issues depend on should warn, the same way closing an issue with
  open sub-issues warns today (`updateIssue`'s sub-issue check).
- A CLI entry point, either a new `sisyphus depends-on <name> <blocking-issue>` command (mirroring
  `sisyphus parent`) or a `--depends-on`/`--clear-depends-on` pair on `update`. Pick whichever fits
  a list field better; `parent`'s single-value command may not translate directly.

## Acceptance criteria

- [ ] An issue's frontmatter can record one or more `depends-on` issues.
- [ ] Depending on a non-existent issue, on itself, or on a cycle is rejected with a clear error.
- [ ] Closing an issue warns if other open or in-progress issues still depend on it.
- [ ] [[list-issues-command]] can filter for issues that are blocked (have an open/in-progress
      dependency) once that command exists.
- [ ] Tests cover adding, clearing, cycle detection, and the close-time warning.

## Out of scope

- A dependency graph visualization; a flag on [[list-issues-command]] (if it lands first) or a
  simple text listing is enough for now.
- Automatically blocking `sisyphus update <name> in-progress` when a dependency is still open; start
  with a warning, not a hard error, consistent with how the sub-issue check works today.

## Notes

- 2026-10-07: Implemented as planned, with one design choice: picked the dedicated-command option
  (`sisyphus depends-on <name> [<blocking-issue>] [--clear]`) since `depends-on` is a list, not a
  single value like `parent`. `--clear` with a second argument removes just that one dependency;
  `--clear` alone removes all of them; a second argument alone adds it. Cycle detection
  (`dependsOnReaches` in `issues.go`) is a DFS over the depends-on graph from the proposed blocking
  issue, checking whether it already (transitively) reaches the dependent issue — simple ancestor-walk
  cycle detection (as `parent` uses) does not apply here since depends-on is a DAG, not a tree.
  Extracted `document.go`'s per-value unquoting into `unquote`, reused to parse list-of-quoted-
  wikilinks fields. Also added `sisyphus list --blocked` (and the same flag on `search`), per the
  acceptance criteria.

## Resolution

Completed: issues can record `depends-on`, validated against existence, self-reference, and cycles
(direct and transitive); closing a depended-on issue warns if dependents are still open; `list`/
`search --blocked` filters to issues with an open dependency. See [[0.0.7]].
