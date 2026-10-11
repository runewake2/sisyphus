---
title: "Spike: cache the issue index between commands"
state: open            # open | in-progress | closed. Must match the directory of the file.
resolution:            # completed | abandoned. Set this only when state is closed.
priority: low          # critical | high | medium | low
effort: medium         # small | medium | large | x-large
tags: [performance, spike] # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-11
closed:                # YYYY-MM-DD. Set this only when state is closed.
owner:                 # The person or agent working on the issue. Cleared when the issue returns to open.
approver:              # The person or agent who accepts the issue when it closes.
bookmark:              # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: []         # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
metadata: {}           # Optional. Arbitrary key-value notes, for example an AI agent session id to resume work with context: {session-id: "abc123"}.
deferred-from:         # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent: "[[performance/large-repo-performance]]" # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
depends-on: ["[[performance/add-performance-benchmarks]]"] # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# Spike: cache the issue index between commands

## Summary

Every command walks the state directories and parses the issues it needs again, so a repo of 20,000 issues costs about 1 s for each `list`, and an agent that calls `sisyphus-mcp` many times pays this on each call. Find out whether a cache between commands is safe and worth it.

## Context

Measured on 2026-10-11 with sisyphus 0.0.57 on a 64-core Linux machine, with generated repos of 1,000, 5,000, and 20,000 issues (10% epics, 70% sub-issues, 40% with 1-3 dependencies, 30% in subdirectories, 3 body wikilinks per issue, one design document per 20 issues). Times are the median of 3 runs. See [[large-repo-performance]] for the full table.

`issueIndex` (`cmd/sisyphus/index.go`) says: "It is never kept between commands: each command must see the edits
that other sessions make in the meantime." Several agents and humans edit the same repo, often by hand, so a
cache must never show stale data. Options to compare:

- A cache file (for example in `.jj/` or a gitignored `.sisyphus/`) with the parsed frontmatter of each issue,
  keyed by path, size, and modification time; each command still stats every file, but parses only the changed
  ones.
- A long-running `sisyphus-mcp` that keeps the index in memory and watches the files (fsnotify), so agent
  calls skip the walk and the parse.
- No cache: the other sub-issues of [[large-repo-performance]] may make every command fast enough.

Measure each option against the benchmarks of [[add-performance-benchmarks]] after the other fixes are in.

## Acceptance criteria

- [ ] A short design document compares the options with measurements, and covers correctness when files change by hand, by another agent, and through jj/git operations.
- [ ] A decision: cache or no cache, recorded as a decision record.
- [ ] Follow-up issues for the chosen option, if any.

## Out of scope

- Building the cache before the decision.

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

<Complete this section when you close the issue.
- Completed: tell what was done. Link to the changelog version and the decision records, for example `[[0.0.5]]`.
- Abandoned: tell why the work stopped. Link to the issue or decision that replaces it, if one exists.>
