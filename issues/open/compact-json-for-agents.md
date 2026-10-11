---
title: "Keep the JSON output of show and search small for agents"
state: open            # open | in-progress | closed. Must match the directory of the file.
resolution:            # completed | abandoned. Set this only when state is closed.
priority: medium       # critical | high | medium | low
effort: medium         # small | medium | large | x-large
tags: [cli, sisyphus-mcp, llm] # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-11
closed:                # YYYY-MM-DD. Set this only when state is closed.
owner:                 # The person or agent working on the issue. Cleared when the issue returns to open.
approver:              # The person or agent who accepts the issue when it closes.
bookmark:              # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: []         # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
metadata: {}           # Optional. Arbitrary key-value notes, for example an AI agent session id to resume work with context: {session-id: "abc123"}.
deferred-from:         # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent: "[[agent-cooperation]]" # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
depends-on: []           # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# Keep the JSON output of show and search small for agents

## Summary

`show --json` and `search --json` return the full body of each issue, with the template text of sections that nobody filled. Agents read this output into their context, so it costs many tokens. Leave out the template text, and let the caller ask for less.

## Context

Measured in this repo with sisyphus 0.0.84:

- `sisyphus search mcp --json`: 34,352 bytes for 12 issues, because each match has its full body in `content`.
- `sisyphus list --state all --json`: 24,868 bytes.
- The body of a new issue is about 970 bytes of template text, for example `<Record progress, findings, and open questions here while the work continues.>`. `show` in text form already hides sections that hold only template text ([[readable-show-output]]). The JSON keeps it.

The MCP tools `sisyphus_show` and `sisyphus_search` return JSON by default, so every agent call pays this cost.

Proposed:

- In JSON, leave out sections that hold only template text, as the text form does. A flag (for example `--raw`) keeps the body as it is in the file.
- `search` returns only the matched lines (with the section name) by default, and the full body with `--content`.
- `list` and `search` take `--limit`, and say how many results were left out.
- The MCP tools get the same arguments.

## Acceptance criteria

- [ ] `show --json` of a new issue has no template text.
- [ ] `search --json` returns only the matched lines unless the caller asks for the body.
- [ ] `--limit` works on `list` and `search`, and the MCP tools pass it.
- [ ] The search for `mcp` in this repo returns less than a third of the bytes of today.
- [ ] The command pages describe the changes, and [[output-conventions-doc]] records the JSON fields.

## Out of scope

- New fields in the rows: [[list-search-show-bookmark]].
- Better search matching: [[search-matches-all-words]].

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

<Complete this section when you close the issue.
- Completed: tell what was done. Link to the changelog version and the decision records, for example `[[0.0.5]]`.
- Abandoned: tell why the work stopped. Link to the issue or decision that replaces it, if one exists.>
