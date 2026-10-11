---
title: "Parse the frontmatter of each issue once, without a regex for each field"
state: open            # open | in-progress | closed. Must match the directory of the file.
resolution:            # completed | abandoned. Set this only when state is closed.
priority: high         # critical | high | medium | low
effort: medium         # small | medium | large | x-large
tags: [performance, cli] # The components that the work touches, for example [widget-scheduler, plan]
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

# Parse the frontmatter of each issue once, without a regex for each field

## Summary

Each `document.get` runs a regular expression over the frontmatter lines until it finds the key, and `issueProblems` runs it again on every line of every issue. In `list` at 20,000 issues this is about 45% of the time. Parse each frontmatter once into fields, and read and check those.

## Context

Measured on 2026-10-11 with sisyphus 0.0.57 on a 64-core Linux machine, with generated repos of 1,000, 5,000, and 20,000 issues (10% epics, 70% sub-issues, 40% with 1-3 dependencies, 30% in subdirectories, 3 body wikilinks per issue, one design document per 20 issues). Times are the median of 3 runs. See [[large-repo-performance]] for the full table.

A CPU profile of `list` at 20,000 issues (900 ms):

| Part | Share |
|---|---|
| `regexp.FindStringSubmatch` (all callers) | 45% |
| `issueProblems` (the read warnings of [[0012-warn-about-invalid-issues-on-read]]) | 41% |
| `document.get` | 18% |
| reading files | 29% |

`fieldPattern` in `cmd/sisyphus/document.go` matches one line at a time. `get` scans `front` from the start for
each key, and `list` calls it about 6 times for each issue; `issueProblems` calls `FindStringSubmatch` for
every line and `get` again. A parse that splits each line once at the first `:` and finds the value and the
comment with simple string code, kept in a map from key to line index, makes `get` a map access. `set` must
keep its behavior: the inline comment and its column stay.

## Acceptance criteria

- [ ] Each frontmatter line is parsed at most once per command; `get` does not run a regular expression.
- [ ] `issueProblems` uses the parsed fields.
- [ ] All tests pass, including the ones that check inline comments after `set`.
- [ ] `list` at 20,000 issues is at least 30% faster.

## Out of scope

- Reading files in parallel ([[read-issue-files-in-parallel]]).

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

<Complete this section when you close the issue.
- Completed: tell what was done. Link to the changelog version and the decision records, for example `[[0.0.5]]`.
- Abandoned: tell why the work stopped. Link to the issue or decision that replaces it, if one exists.>
