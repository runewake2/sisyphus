---
title: "Include the bookmark in list and search results"
state: open            # open | in-progress | closed. Must match the directory of the file.
resolution:            # completed | abandoned. Set this only when state is closed.
priority: medium       # critical | high | medium | low
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

# Include the bookmark in list and search results

## Summary

Add the `bookmark` field to the JSON rows of `sisyphus list` and `sisyphus search`. Models answer questions about bookmarks from these rows and report a wrong value today.

## Context

List rows have the keys `name, owner, priority, state, tags, title`. Search rows add `content`. Neither has `bookmark`. In the spike ([[ollama-local-models]]), task t03 asks "Who owns the docs site issue, and what bookmark is it on?". With thinking off, qwen3.5:9b searched, got a row with `owner` but no `bookmark`, and answered with the issue name as the bookmark (4 of 4 runs). With thinking on, it called `sisyphus_show` and answered correctly. ministral-3:8b answered "no bookmark".

## Acceptance criteria

- [ ] The JSON rows of `list` and `search` include `bookmark` when it is set.
- [ ] The text table of `list` does not get wider by default.
- [ ] Tests cover the new field.

## Out of scope

- Other fields. Add them only if a test shows a need.

## Notes

## Resolution

<Complete this section when you close the issue.
- Completed: tell what was done. Link to the changelog version and the decision records, for example `[[0.0.5]]`.
- Abandoned: tell why the work stopped. Link to the issue or decision that replaces it, if one exists.>
