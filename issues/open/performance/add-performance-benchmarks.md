---
title: "Add a large-repo generator and benchmarks to catch slow code"
state: open            # open | in-progress | closed. Must match the directory of the file.
resolution:            # completed | abandoned. Set this only when state is closed.
priority: medium       # critical | high | medium | low
effort: medium         # small | medium | large | x-large
tags: [performance, tests] # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-11
closed:                # YYYY-MM-DD. Set this only when state is closed.
owner:                 # The person or agent working on the issue. Cleared when the issue returns to open.
approver:              # The person or agent who accepts the issue when it closes.
bookmark:              # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: []         # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
metadata: {}           # Optional. Arbitrary key-value notes, for example an AI agent session id to resume work with context: {session-id: "abc123"}.
deferred-from:         # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent: "[[performance/large-repo-performance]]" # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
depends-on: []           # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# Add a large-repo generator and benchmarks to catch slow code

## Summary

sisyphus has no benchmarks and no large test repo, so a quadratic loop can land without anyone seeing it, as the link check and the cycle check in `lint` show. Add a generator for large repos and Go benchmarks for the main commands.

## Context

Measured on 2026-10-11 with sisyphus 0.0.57 on a 64-core Linux machine, with generated repos of 1,000, 5,000, and 20,000 issues (10% epics, 70% sub-issues, 40% with 1-3 dependencies, 30% in subdirectories, 3 body wikilinks per issue, one design document per 20 issues). Times are the median of 3 runs. See [[large-repo-performance]] for the full table.

The measurements for [[large-repo-performance]] used a script that writes issue files directly in the format of
`sisyphus new`: epics, sub-issues, dependencies on earlier issues (no cycles), subdirectories, states, tags,
body wikilinks, and design documents with links to headings. It writes 20,000 issues in seconds. A Go
version in the test code (a `testing.B` helper that builds the repo in `b.TempDir()`) gives repeatable
benchmarks without Python.

Benchmarks for `list`, `search`, `lint`, `graph`, `show`, `new`, and `update` at 1,000 and 10,000 issues show
growth: a quadratic step multiplies the time by 100 when the size grows by 10. A test can also assert an upper
bound on the work, the way `TestEveryCommandReadsEachStateDirectoryAndIssueFileOnce` counts file reads.

## Acceptance criteria

- [ ] A test helper generates a large repo of a given size with the shape above.
- [ ] `go test -bench` covers `list`, `search`, `lint`, `graph`, `show`, `new`, and `update` at two sizes.
- [ ] A normal `go test ./...` does not run the large benchmarks.
- [ ] CONTRIBUTING or a design note tells how to run them and how to compare results.

## Out of scope

- Running benchmarks in CI with fixed thresholds: timing on shared runners is noisy.

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

<Complete this section when you close the issue.
- Completed: tell what was done. Link to the changelog version and the decision records, for example `[[0.0.5]]`.
- Abandoned: tell why the work stopped. Link to the issue or decision that replaces it, if one exists.>
