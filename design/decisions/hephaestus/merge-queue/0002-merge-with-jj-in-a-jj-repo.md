# 0002: Merge with jj in a jj repo

- Status: Accepted, amended by [[0003-rebase-the-bookmarks-of-the-tasks]]: the bookmarks of the tasks
  now move onto the line.
- Date: 2026-10-20
- Amends: [[0001-merge-with-git-in-a-scratch-worktree]]

## Context

[[0001-merge-with-git-in-a-scratch-worktree]] merged with git in every repo: in a jj repo, in a
scratch worktree of the git store of jj, between `jj git export` and `jj git import`. The tool must
prefer the VCS of the repo: jj commands in a jj repo, git commands in a git repo. The two reasons
of 0001 for git have jj answers now in hand:

- **A conflict must stop the pick.** `jj duplicate` records a conflict in the new commit instead of
  failing; the commit says so (`conflict` in a template), and `jj resolve --list -r` names the
  files. The merge queue reads the commit and leaves the bookmark out, as the git path does on a
  failed cherry-pick.
- **The new commits must be found without parsing output.** `jj duplicate --insert-before @` puts
  the copy directly under the working-copy commit of the scratch workspace, so the new commit is
  always `@-`, and the files of the working copy follow the tip for the planners and the checks.

## Decision

- `hephaestus merge-queue` builds the line in a scratch working copy of the VCS of the repo: a jj
  workspace in a jj repo, a git worktree of the git directory in a git repo. One `line` interface in
  package mergequeue; the run logic does not know which VCS works for it.
- In a jj repo, the run reads the bookmarks and the base with jj, so no `jj git export` before the
  run; the result is a jj bookmark (`jj bookmark create`), so no `jj git import` after it. One
  `jj git export` after a run that made a line lets the git side of a colocated repo see the result.
- A rollback is `jj abandon <head>..@`: it drops the picked commits, the version commit, and what
  the checks left in the working copy, and jj puts the working copy back on the head of the line.
- Closing the run forgets the scratch workspace and abandons its working-copy commit, so the repo
  keeps no trace of the run but the result.

## Alternatives considered

- **Keep merging with git in a jj repo (0001).** It works, but every run writes to the git store of
  jj twice (export, import), and the tool runs git where the repo speaks jj.
- **`jj rebase` instead of `jj duplicate`.** It moves the commits of the tasks, which must not
  change.

## Consequences

- In a jj repo, a run makes no git call at all; `git` does not have to be installed.
- The checks and the planners read the same scratch directory as before; only the VCS under it
  changed.
- 0001 still describes the git path of the line, and stays the design of a git repo.
