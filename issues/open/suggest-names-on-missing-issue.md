---
title: "Suggest close matches when an issue name does not exist"
state: open            # open | in-progress | closed. Must match the directory of the file.
resolution:            # completed | abandoned. Set this only when state is closed.
priority: high         # critical | high | medium | low
effort: small          # small | medium | large | x-large
tags: [cli, llm]       # The components that the work touches, for example [widget-scheduler, plan]
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

# Suggest close matches when an issue name does not exist

## Summary

When a command gets an issue name that does not exist, the error must list the closest existing names. Models guess names, and today they cannot recover from the error.

## Context

The error today is `No issue named 'docs-site-theme' in issues/open, issues/in-progress, or issues/closed.` It has no hint. In the spike ([[ollama-local-models]]), qwen3:4b got this error 23 times, and qwen3:8b got it 22 times. ministral-3:8b recovered from it in 0 of 6 runs. Models that guessed `docs/theme-update` or `docs-site-theme-update` did not find `update-docs-site`.

Proposed behavior: add up to 3 close matches to the error, by shared words in the name and title, for example `Did you mean: update-docs-site (Update the docs site theme)?`. The MCP tool result carries the same text, so the model can correct itself.

## Acceptance criteria

- [ ] A command with an unknown name lists up to 3 close existing names with their titles.
- [ ] The error still fails the command (no automatic choice).
- [ ] Tests cover a name with shared words and a name with no match.

## Out of scope

- Automatic correction of the name.

## Notes

## Resolution

<Complete this section when you close the issue.
- Completed: tell what was done. Link to the changelog version and the decision records, for example `[[0.0.5]]`.
- Abandoned: tell why the work stopped. Link to the issue or decision that replaces it, if one exists.>
