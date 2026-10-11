---
title: "Refuse to draw graphs too large to read, and point to JSON or Mermaid"
state: closed          # open | in-progress | closed. Must match the directory of the file.
resolution: completed  # completed | abandoned. Set this only when state is closed.
priority: high         # critical | high | medium | low
effort: small          # small | medium | large | x-large
tags: [performance, cli, graph] # The components that the work touches, for example [widget-scheduler, plan]
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

# Refuse to draw graphs too large to read, and point to JSON or Mermaid

## Summary

`graph --full` draws every issue that is linked to the focus. In a large, connected repo that is thousands of boxes: 170 MB of text and 65 s at 20,000 issues. No person can read it. Refuse to draw above a size limit, and point to `--json` or `--mermaid`, or make the limit a flag.

## Context

Measured on 2026-10-11 with sisyphus 0.0.57 on a 64-core Linux machine, with generated repos of 1,000, 5,000, and 20,000 issues (10% epics, 70% sub-issues, 40% with 1-3 dependencies, 30% in subdirectories, 3 body wikilinks per issue, one design document per 20 issues). Times are the median of 3 runs. See [[large-repo-performance]] for the full table.

At 5,000 issues, `graph <epic> --full` draws 4,445 issues: 19,429 lines of up to 22,612 characters, 170 MB.
94% of the CPU time is in `drawGraph` → `render`, mostly `canvas.line`, which marks each cell of each edge,
and `canvas.render`, which writes every cell. At 20,000 issues it takes 65 s. `graph <epic>` without `--full`
takes 1.6 s at 20,000 issues: the epic has 5 sub-issues, but the chain of their dependents puts 2,800
issues in the drawing.

A drawing is useful only while it fits on a screen or two. `--json` and `--mermaid` do not draw and are the
right output for a large graph. The node count is known after `buildGraph`, before any drawing.

## Acceptance criteria

- [ ] When a drawing would have more than a limit of issues (for example 200), `graph` does not draw. It prints the number of issues, and tells the user to use `--json`, `--mermaid`, or a narrower focus, or to raise the limit with a flag such as `--max-issues`.
- [ ] `--json` and `--mermaid` have no limit.
- [ ] `sisyphus_graph` gives the same message, so that an agent does not get 170 MB of text.
- [ ] [[sisyphus-graph]] and a test describe the limit.

## Out of scope

- A faster renderer. It may help the normal case, but a drawing of thousands of boxes stays useless.

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

Completed in [[0.0.60]]. `graph` refuses a drawing of more than `--max-issues` (default 200) issues before it draws. `graph --full` at 20,000 issues: 65 s → 1.3 s.

