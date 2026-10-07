---
title: "Use git/jj move commands when an issue changes state"
state: open            # open | in-progress | closed. Must match the directory of the file.
resolution:            # completed | abandoned. Set this only when state is closed.
priority: medium       # critical | high | medium | low
effort: medium         # small | medium | large | x-large
tags: [cli]            # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-07
closed:                # YYYY-MM-DD. Set this only when state is closed.
owner:                 # The person or agent working on the issue. Cleared when the issue returns to open.
approver:              # The person or agent who accepts the issue when it closes.
bookmark:              # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: []         # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
agent-session:         # AI agent session id of the current agent working on the issue, if available.
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

<Record progress, findings, and open questions here while the work continues.>

## Resolution

<Complete this section when you close the issue.
- Completed: tell what was done. Link to the changelog version and the decision records, for example [[0.0.5]].
- Abandoned: tell why the work stopped. Link to the issue or decision that replaces it, if one exists.>
