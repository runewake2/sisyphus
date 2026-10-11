---
title: "sisyphus init --force resets VERSION and CHANGELOG.md to 0.0.0 on an already-versioned repo"
state: closed          # open | in-progress | closed. Must match the directory of the file.
resolution: abandoned  # completed | abandoned. Set this only when state is closed.
priority: high         # critical | high | medium | low
effort: small          # small | medium | large | x-large
tags: [cli]            # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-07
closed: 2026-10-07     # YYYY-MM-DD. Set this only when state is closed.
owner:                 # The person or agent working on the issue. Cleared when the issue returns to open.
approver: samw         # The person or agent who accepts the issue when it closes.
bookmark:              # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: []         # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
metadata: {}           # Optional. Arbitrary key-value notes, for example an AI agent session id to resume work with context: {session-id: "abc123"}.
deferred-from:         # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent:                # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
depends-on: []           # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# sisyphus init --force resets VERSION and CHANGELOG.md to 0.0.0 on an already-versioned repo

## Summary

`sisyphus init --force` overwrites `VERSION` and `CHANGELOG.md` with the kit's static template
content (`0.0.0` and a single entry), destroying a repo's real accumulated version history, because
`--force` has no "merge with what is already there" logic for these two files the way the rest of
`init` does for new-vs-existing files.

## Context

Discovered while adding the `.github/workflows/*.yml` templates to the kit: running
`sisyphus init --force` on this repo (already at a version past `0.0.0`) reset `VERSION` to `0.0.0`
and `CHANGELOG.md` to just the `[[0.0.0]]` entry, silently discarding every later version. Recovered
with `jj restore` since the repo is jj-versioned; a repo without that safety net would lose the data
outright.

`initRepo` (`cmd/sisyphus/init.go`) treats every kit file the same way: skip if it exists and
`--force` is not given, otherwise write the template verbatim. That is correct for files like
`AGENTS.md` or the new workflow files, where the template content does not change over the life of
the repo. `VERSION` and `changelog/0.0.0.md`/`CHANGELOG.md` are different: they are meant to be
mutated by every later version bump, so blindly rewriting them on `--force` is actively wrong rather
than merely redundant.

## Acceptance criteria

- [ ] `sisyphus init --force` does not touch `VERSION` or `CHANGELOG.md` when they already exist and
      already look like a real (post-`0.0.0`) history, or: `--force` skips these two files
      unconditionally once created, the way files behave today without `--force`.
- [ ] A regression test: run `init`, bump the version by hand (or via a later `init` call is not
      enough; simulate a real change), run `init --force` again, and assert `VERSION`/`CHANGELOG.md`
      are unchanged.
- [ ] `init`'s `--force` help text and `CONTRIBUTING.md`/`AGENTS.md`'s documentation of `init` note
      this exception, if one remains after the fix.

## Out of scope

- Changing how `--force` behaves for every other kit file; this is specific to the two files that
  accumulate state over time.

## Notes

- 2026-10-07: Superseded by [[narrow-sisyphus-init]]: `sisyphus init` no longer writes `VERSION` or
  `CHANGELOG.md` at all (it only sets up `issues/` now), so `--force` can no longer touch them and
  this bug's premise no longer applies.

## Resolution

Abandoned: no longer applicable. `sisyphus init` was narrowed in [[narrow-sisyphus-init]] to only
write `issues/`, so `--force` cannot reset `VERSION`/`CHANGELOG.md` anymore.
