---
title: "Skip ignored and non-Markdown files in whole-repo walks"
state: closed          # open | in-progress | closed. Must match the directory of the file.
resolution: completed  # completed | abandoned. Set this only when state is closed.
priority: medium       # critical | high | medium | low
effort: medium         # small | medium | large | x-large
tags: [performance, cli] # The components that the work touches, for example [widget-scheduler, plan]
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

# Skip ignored and non-Markdown files in whole-repo walks

## Summary

`new`, `slug`, `resolve`, `links`, and `lint` walk every file in the repo to find Markdown names and link targets. Only `.git`, `.jj`, `bin`, `obj`, and `node_modules` are skipped. In a repo with a large `vendor/`, `target/`, `dist/`, or `.venv/`, these commands slow down a lot.

## Context

Measured on 2026-10-11 with sisyphus 0.0.57 on a 64-core Linux machine, with generated repos of 1,000, 5,000, and 20,000 issues (10% epics, 70% sub-issues, 40% with 1-3 dependencies, 30% in subdirectories, 3 body wikilinks per issue, one design document per 20 issues). Times are the median of 3 runs. See [[large-repo-performance]] for the full table.

A 1,000-issue repo with 100,000 extra `.go` files in `vendor/`:

| Command | Without vendor/ | With vendor/ |
|---|---|---|
| `slug` | 4 ms | 84 ms |
| `new` | 6 ms | 79 ms |
| `resolve` | 4 ms | 135 ms |
| `links` | 5 ms | 163 ms |
| `show`, `list` | 4 ms, 46 ms | 11 ms, 49 ms |

`repoFiles` in `cmd/sisyphus/repo.go` walks everything except `skippedDirectories`. `repoCandidates` keeps every
file, because a wikilink can point to any file, for example a YAML file. `markdownNames` (for `new` and `slug`)
needs only Markdown files but walks everything too. Each of these commands walks the repo once for this and
`loadIndex` walks `issues/` again.

Options: honor `.gitignore` (or list files with `git ls-files` / `jj file list` when the repo has them), skip
the usual build and vendor directories, and let `new` and `slug` collect only Markdown names.

## Acceptance criteria

- [ ] The whole-repo walk skips files that git ignores, or a documented list of build and vendor directories, and the list can be changed per repo.
- [ ] `new` and `slug` collect only Markdown file names.
- [ ] A test shows that a file in an ignored directory is not a link target and is not walked.
- [ ] With 100,000 ignored files, `resolve` and `links` stay within 2 times their time without them.

## Out of scope

Nothing.

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

Completed in [[0.0.66]]. The repo walk skips what git ignores (one git call), or the simple root `.gitignore` rules without git. With 100,000 ignored files, `lint` takes 88 ms instead of 9.8 s. See [[0019-skip-ignored-files-in-the-repo-walk]].

