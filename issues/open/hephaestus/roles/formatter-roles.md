---
title: "Add formatter tool roles, one for each language"
state: open            # open | in-progress | closed. Must match the directory of the file.
resolution:            # completed | abandoned. Set this only when state is closed.
priority: medium       # critical | high | medium | low
effort: large          # small | medium | large | x-large
tags: [hephaestus, tool-agents] # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-12
closed:                # YYYY-MM-DD. Set this only when state is closed.
owner:                 # The person or agent working on the issue. Cleared when the issue returns to open.
approver:              # The person or agent who accepts the issue when it closes.
bookmark:              # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: []         # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
metadata: {}           # Optional. Arbitrary key-value notes, for example an AI agent session id to resume work with context: {session-id: "abc123"}.
deferred-from:         # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent: "[[hephaestus/core/delegate-issues-to-agents]]" # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
depends-on: []           # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# Add formatter tool roles, one for each language

## Summary

Add a family of formatter tool roles, one for each language, each in `hephaestus/roles/formatter/<language>/`: `golang` (gofmt), `csharp` (dotnet format whitespace), `yaml`, `json`, and `markdown` (prettier). Each formats only the files that the work changes. As a referendum reviewer, each votes yes or modify with the formatted files.

## Context

Requested by a human on 2026-10-12, from a review of the missing roles. Q9 of [[agent-delegation]] lets a formatter submit its changes as a modification in a referendum, but no role did. On the same day the human asked to label each formatter by its language, for Go, C#, YAML, JSON, and Markdown, in a directory, with a subdirectory named after the language for each implementation. [[roles/formatter/README|The formatter roles]] describes the family. They use no model.

## Acceptance criteria

- [x] Roles can live in a family directory: `roles/<family>/<variant>/role.md` is the role `<family>/<variant>`, and the other files of the directory are its implementation ([[0005-roles-of-a-family-live-in-a-directory]]).
- [x] `hephaestus/roles/formatter/<language>/` has a `role.md` and a `format.sh` for golang, csharp, yaml, json, and markdown.
- [x] Each formatter formats only the files that the work changes (`hephaestus files --changed`), and ends with exit code 0 when no file of its language changed.
- [x] In a scratch referendum, the Go, YAML, JSON, and Markdown formatters each made a modification, the gate accepted each, and the referendum passed on the formatted version.
- [ ] `formatter/csharp` formats a changed C# file: no .NET SDK is installed on the workstation yet.
- [ ] A human checks a referendum with the formatters on a real bookmark.

## Out of scope

- Formatters for other languages: each is a new directory.

## Notes

2026-10-12, from claude: a smoke run in a scratch git repo: `hephaestus referendum call` with the formatter as the only required reviewer, on a bookmark with an unformatted cmd/main.go. Round 1 voted modify, the gate accepted it, v2 had the formatted file, and round 2 passed. A human still checks a run on a real bookmark.

2026-10-12, from claude: replaced the one Go formatter with five roles in `hephaestus/roles/formatter/<language>/`. A smoke run with all five as required reviewers found two problems, now fixed. First, a tool at the gate judged the problems of other languages, so it rejected every modification; now a tool rejects only what the modification makes worse. Second, two reviewers changed git worktrees at the same time; now the changes are serialized. With four formatters that each modify, the referendum needs five rounds (R2: one modification for each round), so it passed with --max-rounds 6. Prettier puts the cells of Markdown tables in columns, and it splits a table cell at the | of a wikilink alias such as `[[a|b]]`, because GitHub-flavored Markdown reads that | as a column separator. A human must decide this before formatter/markdown runs on the docs of this repo.

2026-10-12, from claude: ran the five formatters on the whole repo in a scratch workspace (base root()). Go, C#, and YAML had nothing to change. JSON changed only whitespace in the 16 raw benchmark results; formatter/json now leaves them out. Markdown changed 57 files: mostly padded tables, but also three changes of meaning: lists of failure modes whose numbers were mode ids were renumbered, sub-bullets under a two-digit number were not nested, and two indented code blocks in [[agent-guide]] became text. The source is corrected: the modes are bullets (Mode 19: ...), and the guide has fenced code blocks.

## Resolution

<Complete this section when you close the issue.
- Completed: tell what was done. Link to the changelog version and the decision records, for example `[[0.0.5]]`.
- Abandoned: tell why the work stopped. Link to the issue or decision that replaces it, if one exists.>
