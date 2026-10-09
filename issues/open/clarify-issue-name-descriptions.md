---
title: "Remove the nested-name example and say that slug is only for new issues"
state: open            # open | in-progress | closed. Must match the directory of the file.
resolution:            # completed | abandoned. Set this only when state is closed.
priority: high         # critical | high | medium | low
effort: small          # small | medium | large | x-large
tags: [sisyphus-mcp, cli, llm] # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-09
closed:                # YYYY-MM-DD. Set this only when state is closed.
owner:                 # The person or agent working on the issue. Cleared when the issue returns to open.
approver:              # The person or agent who accepts the issue when it closes.
bookmark:              # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: []         # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
metadata: {}           # Optional. Arbitrary key-value notes, for example an AI agent session id to resume work with context: {session-id: "abc123"}.
deferred-from: "[[spike-ollama-local-models]]" # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent: "[[small-model-friendly-mcp]]" # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
depends-on: []           # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# Remove the nested-name example and say that slug is only for new issues

## Summary

The `name` descriptions in sisyphus-mcp use the example `web/auth/fix-login-bug`. Models copy this pattern and invent nested names. Remove the example from tools that take an existing issue, and tell models to use the exact name from list or search. Also say that `slug` makes names for new issues only.

## Context

`cmd/sisyphus-mcp/tools.go` describes `name` as "...for example fix-login-bug or web/auth/fix-login-bug" (`sisyphus_new`, line 93) and "The issue's full name (for example web/auth/fix-login-bug)..." (other tools, for example lines 139, 172, 248).

In the spike ([[ollama-local-models]]):
- Models guessed nested names for existing issues: `web/auth/password-reset`, `docs/theme-update`, `ai/db/fix-critical-bug`.
- When creating issues, qwen3.5:9b used a nested name in all 6 runs of t06 (`login-api/add-rate-limiting`), and ministral-3:8b and qwen3:14b put the parent into the name (`auth-epic/audit-session-tokens`) instead of setting `parent`.
- qwen3:4b called `sisyphus_slug` to look up an existing issue (2 runs in each config). `slug` only makes a new name.

## Acceptance criteria

- [ ] Tools that take an existing issue tell the model to use the exact name from `sisyphus_list` or `sisyphus_search`, with a flat example.
- [ ] The `sisyphus_new` description says that directories are optional and are not the parent, and that `parent` sets the parent.
- [ ] The `sisyphus_slug` description says that it makes a name for a new issue and does not find existing issues.
- [ ] The CLI help text agrees with the tool descriptions.

## Out of scope

- Removing support for nested names.

## Notes

## Resolution

<Complete this section when you close the issue.
- Completed: tell what was done. Link to the changelog version and the decision records, for example `[[0.0.5]]`.
- Abandoned: tell why the work stopped. Link to the issue or decision that replaces it, if one exists.>
