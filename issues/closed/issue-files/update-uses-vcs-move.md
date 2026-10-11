---
title: "Use git/jj move commands when an issue changes state"
state: closed          # open | in-progress | closed. Must match the directory of the file.
resolution: completed  # completed | abandoned. Set this only when state is closed.
priority: medium       # critical | high | medium | low
effort: medium         # small | medium | large | x-large
tags: [cli]            # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-07
closed: 2026-10-07     # YYYY-MM-DD. Set this only when state is closed.
owner: claude          # The person or agent working on the issue. Cleared when the issue returns to open.
approver: samw         # The person or agent who accepts the issue when it closes.
bookmark: samw/ai/vcs-move # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: [sisyphus-vcs-move] # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
metadata: {session-id: "68fa9a09-fe4e-4ae8-9917-086c3313c5be"} # Optional. Arbitrary key-value notes, for example an AI agent session id to resume work with context: {session-id: "abc123"}.
deferred-from:         # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent:                # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
---

# Use git/jj move commands when an issue changes state

## Summary

`sisyphus update` moves an issue file between `issues/open`, `issues/in-progress`, and `issues/closed`
with a plain `os.Rename` (`saveIssue` in `issues.go`). In a git or jj working copy, the move should go
through the VCS instead, so the repo's history and status reflect it as a move, not an untracked
delete-and-add.

## Context

`saveIssue` (`cmd/sisyphus/issues.go`) currently does:

```go
if from != to {
    return os.Rename(from, to)
}
```

This works, but leaves the move for the human or agent to stage by hand later (`git add -A`, or
relying on jj's automatic working-copy snapshot). The fix should detect which VCS the repo root uses
(a colocated git+jj repo has both `.git` and `.jj`; a plain repo may have only one) and perform the
move the way that VCS expects:

- **git**: run `git mv <from> <to>` so the rename lands in the index immediately.
- **jj**: jj has no `mv` command — it detects renames from content when it snapshots the working
  copy, so a plain `os.Rename` is already correct there. Confirm this and keep the current behavior
  for jj-only repos.

See [[AGENTS]] and [[CONTRIBUTING]] for how this repo itself uses jj bookmarks and workspaces.

## Acceptance criteria

- [ ] When the repo has a `.git` directory, `sisyphus update` runs `git mv` (or equivalent) to move
      the issue file, instead of `os.Rename`.
- [ ] When the repo is jj-only (no `.git`), the move still works correctly (plain rename is fine).
- [ ] Behavior is covered by a test for both cases.
- [ ] `saveIssue` still works when `from == to` (no state change, just a content update).

## Out of scope

- Moving or renaming anything other than the issue file itself (e.g. no change to how `new` or
  `parent` write files).

## Notes

- 2026-10-07: Implemented as planned. `saveIssue` (`issues.go`) now takes `root` and checks
  `usesGit(root)` (a plain `os.Stat` of `root/.git`, so it works whether `.git` is a directory or a
  worktree file) to choose between `gitMove` (`git mv`, run with `cmd.Dir = root`) and the existing
  `os.Rename`. jj needed no change, confirming the issue's assumption.

## Resolution

Completed: `saveIssue` moves issue files with `git mv` when the repo has a `.git` directory, and
with a plain rename otherwise. See [[0.0.2]].
