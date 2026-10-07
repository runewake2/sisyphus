---
title: "Add sisyphus graph to view an issue's family tree in the terminal"
state: closed          # open | in-progress | closed. Must match the directory of the file.
resolution: completed  # completed | abandoned. Set this only when state is closed.
priority: high         # critical | high | medium | low
effort: medium         # small | medium | large | x-large
tags: [cli]            # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-07
closed: 2026-10-07     # YYYY-MM-DD. Set this only when state is closed.
owner: claude          # The person or agent working on the issue. Cleared when the issue returns to open.
approver: samw         # The person or agent who accepts the issue when it closes.
bookmark: samw/ai/graph-view # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: [sisyphus-graph-view] # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
metadata: {}           # Optional. Arbitrary key-value notes, for example an AI agent session id to resume work with context: {session-id: "abc123"}.
deferred-from:         # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent: "[[issue-manager-gaps]]" # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
depends-on: []           # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# Add sisyphus graph to view an issue's family tree in the terminal

## Summary

Add `sisyphus graph <name>` to print an issue's whole family tree in the terminal: its root
ancestor, every descendant with its state, and any depends-on edges, so a large task (an epic and
its sub-issues) can be reviewed visually at a glance instead of reading each issue file by hand.

## Context

Part of [[issue-manager-gaps]]. The building blocks already exist: `parentName` (walk one issue up
to its parent), `subIssues` (every direct sub-issue of one), and `parseDependsOn` (an issue's
depends-on list), all in `issues.go`. `graph` composes them: walk up from `<name>` to the top-most
ancestor with no `parent`, then recursively walk back down through `subIssues` to build the full
tree, marking `<name>`'s own node.

Rendered in the style of the Unix `tree` command (`├──`/`└──`/`│`), with each node as
`<name> [<state>]`, the requested issue marked, and a `(depends on: ...)` suffix on any node that
has one. `--json` prints the same tree as nested objects, for scripting or (later) the MCP server.

## Acceptance criteria

- [ ] `sisyphus graph <name>` prints the whole tree from `<name>`'s root ancestor down.
- [ ] Each node shows its name and state; the requested issue is visibly marked.
- [ ] A node with a `depends-on` entry shows it inline.
- [ ] `--json` prints the same tree as nested objects.
- [ ] Works for an issue with no parent and no children (prints just itself).
- [ ] A missing issue is a clear error, consistent with `show`/`update`.
- [ ] Covered by tests: a multi-level tree, the focus marker, a depends-on annotation, the
      single-node case, every reference form, the missing-issue error, and `--json`.

## Out of scope

- Rendering issues outside the tree that depend on something inside it (only outgoing depends-on
  edges from a node already in the tree are shown).
- A graphical (non-terminal) renderer.

## Notes

- 2026-10-07: Implemented as planned, in a new `graph.go` (`buildGraph`/`buildGraphNode`) plus a
  `writeGraph` renderer in `cli.go`. Guards against a cycle defensively (a shared `seen` map across
  the whole recursion) even though `parent` and `depends-on` are both validated against cycles at
  write time.

## Resolution

Completed: `sisyphus graph <name>` prints an issue's family tree, text or JSON. See [[0.0.11]].

## Resolution

<Complete this section when you close the issue.
- Completed: tell what was done. Link to the changelog version and the decision records, for example `[[0.0.5]]`.
- Abandoned: tell why the work stopped. Link to the issue or decision that replaces it, if one exists.>
