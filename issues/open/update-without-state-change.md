---
title: "Let update change fields without a state argument"
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
parent: "[[issue-manager-gaps]]" # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
depends-on: []           # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# Let update change fields without a state argument

## Summary

`sisyphus update` needs a state argument even when the user only wants to change a field. Make the state optional, so that `sisyphus update <name> --priority high` keeps the current state.

## Context

Found in a review of sisyphus 0.0.31 on 2026-10-11.

```
sisyphus update a-one --priority high   # Error: accepts 2 arg(s), received 1
```

Today the user must look up and repeat the current state. The `sisyphus_update` MCP tool has the same rule (`state` is required). In the spike [[ollama-local-models]], small models had to be told this rule in the system prompt, and passing the wrong state moves the issue by mistake.

## Acceptance criteria

- [ ] `sisyphus update <name> [<state>]`: without a state, the state does not change and the other flags apply.
- [ ] `update <name>` with no state and no flags fails with a clear error.
- [ ] In `sisyphus_update`, `state` becomes optional.
- [ ] [[sisyphus-update]] and the tests cover the new form.

## Out of scope

Nothing.

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

<Complete this section when you close the issue.
- Completed: tell what was done. Link to the changelog version and the decision records, for example `[[0.0.5]]`.
- Abandoned: tell why the work stopped. Link to the issue or decision that replaces it, if one exists.>
