---
title: "Read and parse issue files in parallel"
state: closed          # open | in-progress | closed. Must match the directory of the file.
resolution: completed  # completed | abandoned. Set this only when state is closed.
priority: medium       # critical | high | medium | low
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

# Read and parse issue files in parallel

## Summary

`loadIndex` users read and parse issue files one at a time. Reading is about 30% of `list` at 20,000 issues, and the parse work grows with it. Read and parse the files with a small pool of goroutines.

## Context

Measured on 2026-10-11 with sisyphus 0.0.57 on a 64-core Linux machine, with generated repos of 1,000, 5,000, and 20,000 issues (10% epics, 70% sub-issues, 40% with 1-3 dependencies, 30% in subdirectories, 3 body wikilinks per issue, one design document per 20 issues). Times are the median of 3 runs. See [[large-repo-performance]] for the full table.

`issueIndex.document` (`cmd/sisyphus/index.go`) reads a file on the first request, so `matchingIssues`, `lint`,
`subIssues`, and `dependents` read the files in sequence. A command that needs every document (`list`, `search`,
`lint`, `graph`, closing an issue) can load them all at the start with a worker pool (for example `GOMAXPROCS`
workers, at most 16 or so), and store the results in `docs`. The index test that counts reads must still see each
file read once.

Use a bounded pool: an unbounded number of goroutines would open thousands of files at once.

## Acceptance criteria

- [ ] Commands that read every issue load the documents with a bounded worker pool.
- [ ] Output order and every test stay the same; each file is still read once per command.
- [ ] `list` at 20,000 issues is measurably faster on a multi-core machine, as the benchmarks show.

## Out of scope

Nothing.

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

Completed in [[0.0.63]]. `issueIndex.preload` reads many issues with at most 16 workers, each file once. `list` at 20,000 issues: 414 → 243 ms. The tests pass with the race detector.

