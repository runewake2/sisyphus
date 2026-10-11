---
title: "Read each issue file once in lint"
state: closed          # open | in-progress | closed. Must match the directory of the file.
resolution: completed  # completed | abandoned. Set this only when state is closed.
priority: medium       # critical | high | medium | low
effort: small          # small | medium | large | x-large
tags: [performance, lint] # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-11
closed: 2026-10-11     # YYYY-MM-DD. Set this only when state is closed.
owner: claude          # The person or agent working on the issue. Cleared when the issue returns to open.
approver:              # The person or agent who accepts the issue when it closes.
bookmark: samw/ai/perf-fixes # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: [perf-fixes] # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
metadata: {}           # Optional. Arbitrary key-value notes, for example an AI agent session id to resume work with context: {session-id: "abc123"}.
deferred-from:         # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent: "[[performance/large-repo-performance]]" # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
depends-on: []           # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# Read each issue file once in lint

## Summary

`sisyphus lint` reads every issue file twice: once for the issue checks (`issueIndex.header`), and once more, in full, for the link check (`linkChecker.rows`). Read each issue once, and give the link check the lines that the index already has.

## Context

Found on 2026-10-11 after [[0.0.65]], in the follow-up review of [[large-repo-performance]]. `lintRepo` in `cmd/sisyphus/lint.go` preloads the headers of all issues, then `linkChecker.rows` reads every Markdown file with `os.ReadFile`, which includes every issue file again. At 20,000 issues `lint` takes about 1.3 s, and the second read of 20,000 files is part of it. The lines of the whole file are the frontmatter lines and the body lines of a full document, so the link check can use them.

## Acceptance criteria

- [ ] `lint` reads each issue file once, as a test shows with the read counter.
- [ ] The link problems and their line numbers do not change.

## Out of scope

Nothing.

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

Completed in [[0.0.67]]. `lint` loads each issue once and gives its lines to the link check. `lint` at 20,000 issues: 1,247 → 1,051 ms, with the same output.

