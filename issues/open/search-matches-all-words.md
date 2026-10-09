---
title: "Make search match all words in any order"
state: open            # open | in-progress | closed. Must match the directory of the file.
resolution:            # completed | abandoned. Set this only when state is closed.
priority: high         # critical | high | medium | low
effort: small          # small | medium | large | x-large
tags: [cli, search, llm] # The components that the work touches, for example [widget-scheduler, plan]
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

# Make search match all words in any order

## Summary

Make `sisyphus search` match issues that contain all of the query words, in any order, not only the exact phrase. Models search with natural phrases and get no results today.

## Context

`searchIssues` in `cmd/sisyphus/search.go` uses `strings.Contains(haystack, query)`, so the query must appear as one exact phrase. In the spike ([[ollama-local-models]]), models searched for `"docs site theme update"`. The issue is titled "Update the docs site theme", so the search returned `[]`. The close task (t08) then failed in all 6 runs of qwen3:4b, and in many runs of other models. A search for `"theme"` finds the issue.

Proposed behavior: split the query into words, and match an issue when every word appears somewhere in its title or content (case-insensitive). Keep a way to search for an exact phrase, for example with quotes.

## Acceptance criteria

- [ ] `sisyphus search "docs site theme update"` finds an issue titled "Update the docs site theme".
- [ ] An exact-phrase search is still possible, and the help text tells how.
- [ ] The `sisyphus_search` tool description says that the words can be in any order.
- [ ] Tests cover word order and exact phrases.

## Out of scope

- Fuzzy matching or ranking of results.

## Notes

## Resolution

<Complete this section when you close the issue.
- Completed: tell what was done. Link to the changelog version and the decision records, for example `[[0.0.5]]`.
- Abandoned: tell why the work stopped. Link to the issue or decision that replaces it, if one exists.>
