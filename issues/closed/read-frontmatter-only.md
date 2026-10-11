---
title: "Read only the frontmatter when a command does not need the body"
state: closed          # open | in-progress | closed. Must match the directory of the file.
resolution: completed  # completed | abandoned. Set this only when state is closed.
priority: low          # critical | high | medium | low
effort: small          # small | medium | large | x-large
tags: [performance, cli] # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-11
closed: 2026-10-11     # YYYY-MM-DD. Set this only when state is closed.
owner: claude          # The person or agent working on the issue. Cleared when the issue returns to open.
approver:              # The person or agent who accepts the issue when it closes.
bookmark: samw/ai/perf-fixes # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: [perf-fixes] # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
metadata: {}           # Optional. Arbitrary key-value notes, for example an AI agent session id to resume work with context: {session-id: "abc123"}.
deferred-from:         # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent: "[[large-repo-performance]]" # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
depends-on: []           # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# Read only the frontmatter when a command does not need the body

## Summary

`list`, `graph`, `--ready`, `--blocked`, and the relationship checks need only the frontmatter, but `loadDocument` reads and splits the whole file. Read up to the closing `---` when the body is not needed.

## Context

Measured on 2026-10-11 with sisyphus 0.0.57 on a 64-core Linux machine, with generated repos of 1,000, 5,000, and 20,000 issues (10% epics, 70% sub-issues, 40% with 1-3 dependencies, 30% in subdirectories, 3 body wikilinks per issue, one design document per 20 issues). Times are the median of 3 runs. See [[large-repo-performance]] for the full table.

The generated issues are about 4 KB each, mostly the body. In a real repo, bodies with notes and logs can be
much larger. `loadDocument` (`cmd/sisyphus/document.go`) reads the file with `os.ReadFile` and splits every line.
A reader that stops after the frontmatter, and loads the body only when a command asks for it (`search`, `show`,
`edit`, `lint` links), saves reading and splitting. It matters most together with
[[read-issue-files-in-parallel]] and [[parse-frontmatter-once]].

## Acceptance criteria

- [ ] Commands that do not use the body read only the frontmatter of each issue.
- [ ] A command that writes an issue still writes the full file unchanged except for its edit.
- [ ] Every test passes.

## Out of scope

Nothing.

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

Completed in [[0.0.62]]. `loadHeader` reads up to the end of the frontmatter, and `issueIndex.header` serves the commands that do not use the body. With 40 KB bodies, `list` at 5,000 issues: 249 → 109 ms; with 4 KB bodies the gain is small.

