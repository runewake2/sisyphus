---
title: "Index link targets by name instead of scanning every file per link"
state: open            # open | in-progress | closed. Must match the directory of the file.
resolution:            # completed | abandoned. Set this only when state is closed.
priority: high         # critical | high | medium | low
effort: small          # small | medium | large | x-large
tags: [performance, cli, lint] # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-11
closed:                # YYYY-MM-DD. Set this only when state is closed.
owner:                 # The person or agent working on the issue. Cleared when the issue returns to open.
approver:              # The person or agent who accepts the issue when it closes.
bookmark:              # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: []         # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
metadata: {}           # Optional. Arbitrary key-value notes, for example an AI agent session id to resume work with context: {session-id: "abc123"}.
deferred-from:         # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent: "[[large-repo-performance]]" # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
depends-on: []           # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# Index link targets by name instead of scanning every file per link

## Summary

Every wikilink is resolved by `matchTarget`, which loops over every file in the repo. `lint` resolves every link in every Markdown file, so it does links × files comparisons: 39 s at 20,000 issues. Build a map from link key to candidates once, and look each link up in it.

## Context

Measured on 2026-10-11 with sisyphus 0.0.57 on a 64-core Linux machine, with generated repos of 1,000, 5,000, and 20,000 issues (10% epics, 70% sub-issues, 40% with 1-3 dependencies, 30% in subdirectories, 3 body wikilinks per issue, one design document per 20 issues). Times are the median of 3 runs. See [[large-repo-performance]] for the full table.

A CPU profile of `lint` on 5,000 issues shows 80% of the time in `matchTarget` (`cmd/sisyphus/links.go`):
`strings.EqualFold` and `hasSuffixFold` in `slices.ContainsFunc` over every candidate, for every link. At
20,000 issues there are about 208,000 wikilinks and 21,000 Markdown files, so about 4 billion comparisons.

`matchTarget` has two kinds of match: an exact key, ignoring case, and a suffix match for a link that
contains `/` (`hasSuffixFold(key, "/"+target)`). A map from lowercase key to candidates answers the first.
For the second, index each candidate under every path suffix that starts after a `/` (a path of depth d gives
d keys), so that a suffix lookup is one map access too. `linkChecker` already builds the candidates once per
command; the map belongs there, and `resolveLink` and `listLinks` can use it too.

## Acceptance criteria

- [ ] Link resolution looks up candidates in a map; no code path compares a link with every file.
- [ ] Every existing test of `links`, `resolve`, and `lint` passes unchanged: same statuses, same ambiguous results, same order of resolved paths.
- [ ] `lint` at 20,000 issues (generated repo) is at least 10 times faster than 39 s.

## Out of scope

- The cycle checks of lint ([[check-cycles-once-in-lint]]).

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

<Complete this section when you close the issue.
- Completed: tell what was done. Link to the changelog version and the decision records, for example `[[0.0.5]]`.
- Abandoned: tell why the work stopped. Link to the issue or decision that replaces it, if one exists.>
