# 0024: Share claims between workspaces

- Status: Accepted
- Date: 2026-10-11
- Version: [[0.0.89]]

## Context

Each agent works in its own jj workspace or git worktree, based on `main`. When an agent starts an
issue, the file moves to `issues/in-progress/` only in its own working copy. Every other workspace
sees the issue as open until a human merges the work. So the check of AGENTS.md
(`sisyphus list --state in-progress`) found only the claims that were already merged, and two agents
could start the same issue ([[claims-across-workspaces]]).

## Decision

- A start records a **claim** in the shared state directory of the repo
  ([[0023-lock-and-write-issues-atomically]]): `sisyphus/claims/<tracker>/<name>.json` in the jj
  repo directory or the git common directory. A claim holds the issue name, the owner, the absolute
  path of the workspace, the bookmark, and the start time.
- A start of an issue that another workspace has claimed fails, also for the same owner, because
  many agents use one owner name. `--take` moves the claim.
- The issue file stays the truth. A claim whose workspace no longer has the issue in
  `issues/in-progress/` is stale: sisyphus ignores it, a start replaces it with a warning, and
  `sisyphus claims --prune` removes it.
- A claim ends when the issue leaves in-progress in its workspace. A change in another workspace
  does not remove it.
- `list --ready` leaves out an open issue that another workspace has claimed. `sisyphus claims`
  lists all claims.
- The lock of [[0023-lock-and-write-issues-atomically]] is in the same directory, so two starts in
  two workspaces at the same moment cannot both succeed.

## Alternatives considered

- **Read the issue files of the other workspaces and bookmarks.** jj does not record the files of a
  workspace until a jj command runs there, so a fresh start is not visible in the commits. The
  directories of the workspaces are visible, but jj and git do not list them in a way that sisyphus
  can read without running them. sisyphus uses the paths that the claims record only to find stale
  claims.
- **A git ref, for example `refs/sisyphus/claims`.** A remote could share it with other machines,
  but each change is a commit, a push is slow and can fail, and a repo without a remote gets
  nothing more than from a file.
- **Change `list --state in-progress` to include claimed issues.** The state of an issue is the
  directory of its file. A list that shows another state than the file has would confuse people and
  scripts.

## Consequences

- Agents in different workspaces of one repo on one machine cannot start the same issue.
- Agents on different machines still can: a claim is local. A shared claim for them is future work.
- A repo without jj or git has no claims, and needs none: it has one working copy.
- An issue that started before 0.0.89 has no claim until it starts again.
