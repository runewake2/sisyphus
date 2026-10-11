---
title: "Filter list and search on a metadata key"
state: open            # open | in-progress | closed. Must match the directory of the file.
resolution:            # completed | abandoned. Set this only when state is closed.
priority: medium       # critical | high | medium | low
effort: small          # small | medium | large | x-large
tags: [feature, cli, sisyphus-mcp] # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-11
closed:                # YYYY-MM-DD. Set this only when state is closed.
owner:                 # The person or agent working on the issue. Cleared when the issue returns to open.
approver:              # The person or agent who accepts the issue when it closes.
bookmark:              # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: []         # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
metadata: {}           # Optional. Arbitrary key-value notes, for example an AI agent session id to resume work with context: {session-id: "abc123"}.
deferred-from:         # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent:                # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
depends-on: []           # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# Filter list and search on a metadata key

## Summary

list and search cannot filter on the metadata field. Add --metadata key=value to list and search (and the MCP tools), so that any client can find the issues that it marked with a key, for example hephaestus with delegate=merge-queue.

## Context

hephaestus records the role of an issue in metadata, for example `sisyphus update fix-login-bug --metadata delegate=merge-queue`, and must find the issues of each role ([[0029-sisyphus-does-not-manage-agents]]). Today it must read every issue with `show --json`. The filter is generic: sisyphus does not know what a key means. `metadata` is a map of strings in the frontmatter, for example `{session-id: "abc123"}`.

## Acceptance criteria

- [ ] `list` and `search` take `--metadata key=value`, repeatable; every pair must match (AND).
- [ ] `--metadata key` with no value matches an issue that has the key with any value.
- [ ] The list and search rows of `--json` include `metadata` when it is set.
- [ ] `sisyphus_list` and `sisyphus_search` take the filter, and the docs and tests cover it.

## Out of scope

- Any meaning of a key in sisyphus, for example a role.

## Notes

2026-10-11, from the human: hephaestus records the role that owns a task as the metadata key delegate, so this filter is how it finds the work of a role, for example sisyphus list --metadata delegate=merge-queue --ready.

## Resolution

<Complete this section when you close the issue.
- Completed: tell what was done. Link to the changelog version and the decision records, for example `[[0.0.5]]`.
- Abandoned: tell why the work stopped. Link to the issue or decision that replaces it, if one exists.>
