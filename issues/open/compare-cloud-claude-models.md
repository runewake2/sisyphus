---
title: "Benchmark the cloud Claude models on the sisyphus task suite"
state: open            # open | in-progress | closed. Must match the directory of the file.
resolution:            # completed | abandoned. Set this only when state is closed.
priority: low          # critical | high | medium | low
effort: small          # small | medium | large | x-large
tags: [spike, llm]     # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-09
closed:                # YYYY-MM-DD. Set this only when state is closed.
owner:                 # The person or agent working on the issue. Cleared when the issue returns to open.
approver:              # The person or agent who accepts the issue when it closes.
bookmark:              # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: []         # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
metadata: {}           # Optional. Arbitrary key-value notes, for example an AI agent session id to resume work with context: {session-id: "abc123"}.
deferred-from: "[[spike-ollama-local-models]]" # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent:                # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
depends-on: []           # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# Benchmark the cloud Claude models on the sisyphus task suite

## Summary

Run the spike task suite against claude-opus-5-5, claude-opus-5, and claude-sonnet-5 through the Claude API, to compare them with the local models.

## Context

The spike harness has an Anthropic backend (`--backend anthropic --effort medium`), but the API key of the test sandbox returned HTTP 429 for every model, so the runs could not happen. The spike compared one Claude model instead: Claude Opus 5.5 as a Claude Code subagent that used the sisyphus CLI. It passed 28 of 28 (see [[ollama-local-models]]).

Run configs A and B for each model, and B at `low` effort. The estimated cost is 30 to 50 USD. Do not use refusal fallbacks, so that each result comes from the named model.

## Acceptance criteria

- [ ] Results for the three models are in `design/ollama-local-models/results/`.
- [ ] The comparison table in [[ollama-local-models]] includes them.

## Out of scope

Nothing.

## Notes

## Resolution

<Complete this section when you close the issue.
- Completed: tell what was done. Link to the changelog version and the decision records, for example `[[0.0.5]]`.
- Abandoned: tell why the work stopped. Link to the issue or decision that replaces it, if one exists.>
