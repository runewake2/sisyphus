---
title: "Keep sisyphus fast in repos with many issues"
state: open            # open | in-progress | closed. Must match the directory of the file.
resolution:            # completed | abandoned. Set this only when state is closed.
priority: medium       # critical | high | medium | low
effort: large          # small | medium | large | x-large
tags: [performance, cli] # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-11
closed:                # YYYY-MM-DD. Set this only when state is closed.
owner:                 # The person or agent working on the issue. Cleared when the issue returns to open.
approver:              # The person or agent who accepts the issue when it closes.
bookmark:              # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: []         # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
metadata: {}           # Optional. Arbitrary key-value notes, for example an AI agent session id to resume work with context: {session-id: "abc123"}.
deferred-from:         # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent:                # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
depends-on: []           # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# Keep sisyphus fast in repos with many issues

## Summary

sisyphus slows down much faster than the number of issues grows. At 20,000 issues, `lint` takes 39 s and `graph --full` takes 65 s, and every command pays for a full walk and parse of the repo. This epic tracks the changes that keep sisyphus fast in a large repo.

## Context

Measured on 2026-10-11 with sisyphus 0.0.57 on a 64-core Linux machine, with generated repos of 1,000, 5,000, and 20,000 issues (10% epics, 70% sub-issues, 40% with 1-3 dependencies, 30% in subdirectories, 3 body wikilinks per issue, one design document per 20 issues). Times are the median of 3 runs. See [[large-repo-performance]] for the full table.

| Command | 1,000 issues | 5,000 issues | 20,000 issues | Growth |
|---|---|---|---|---|
| `lint` | 212 ms | 2,912 ms | 39,258 ms | quadratic |
| `graph <epic> --full` | 232 ms | 4,269 ms | 64,763 ms | quadratic |
| `graph <epic>` | 38 ms | 275 ms | 1,636 ms | faster than linear |
| `list --json` (all states) | 53 ms | 257 ms | 1,011 ms | linear |
| `list`, `list --ready`, `list --blocked` | 45 ms | 210 ms | 800 ms | linear |
| `search` | 50 ms | 228 ms | 880 ms | linear |
| `update <epic> closed` | — | 133 ms | 498 ms | linear |
| `new` | 6 ms | 20 ms | 66 ms | linear |
| `show`, `update`, `edit`, `depends-on`, `resolve`, `slug` | 4 ms | 12 ms | 38 ms | linear |

Extra measurements:

- `lint` on one chain of dependencies (each issue depends on the one before): 1,849 ms for 1,000 issues,
  16,528 ms for 3,000 ([[check-cycles-once-in-lint]]).
- 100,000 non-Markdown files in a `vendor/` directory of a 1,000-issue repo: `slug` 4 → 84 ms, `resolve`
  4 → 135 ms, `links` 5 → 163 ms, `new` 6 → 79 ms ([[skip-ignored-files-in-repo-walk]]).

CPU profiles:

- `lint` (5,000 issues): 80% in `matchTarget`, which compares each link with every file in the repo
  ([[index-link-targets]]).
- `graph --full` (5,000 issues): 94% in drawing. The drawing has 4,445 issues and is 170 MB of text, 19,429
  lines of up to 22,612 characters ([[limit-large-graph-drawings]]).
- `list` (20,000 issues): about 45% in regular expressions (the frontmatter checks of `issueProblems` and
  each `document.get`), about 30% in reading files one at a time ([[parse-frontmatter-once]],
  [[read-issue-files-in-parallel]]).

The sub-issues are in order of measured effect. Do [[add-performance-benchmarks]] first or together with the
first fix, so that each fix can show its effect and later changes cannot make it slow again.

## Acceptance criteria

- [ ] All sub-issues are closed or deferred with a reason.
- [ ] At 20,000 issues (generated as above): `lint` under 2 s, `list` under 300 ms, and `show` of a full name under 10 ms, on the same machine class.
- [ ] No command is quadratic in the number of issues or files, as the benchmarks show.

## Out of scope

- A database or a server process. Plain files stay the source of truth.

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

<Complete this section when you close the issue.
- Completed: tell what was done. Link to the changelog version and the decision records, for example `[[0.0.5]]`.
- Abandoned: tell why the work stopped. Link to the issue or decision that replaces it, if one exists.>
