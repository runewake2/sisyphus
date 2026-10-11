# sisyphus claims

List the issues that each workspace of the repo has in progress, before its work merges.

## Usage

```bash
sisyphus claims [--json] [--prune]
```

## Flags

| Flag | Meaning |
|---|---|
| `--json` | Print JSON instead of a table. |
| `--prune` | Remove the stale claims, and list the others. |

## Behavior

Each jj workspace and git worktree has its own copy of `issues/`. When an agent starts an issue in
its workspace, the file moves to `issues/in-progress/` only there. The other workspaces see the
issue as open until the work merges. A **claim** closes this gap:

- [[sisyphus-start]], `update <name> in-progress`, and `new --state in-progress` record a claim in
  a directory that every workspace of the repo shares: `sisyphus/claims/` in the jj repo directory
  (`.jj/repo`) or in the git common directory. No VCS records it.
- A start of an issue that another workspace has claimed fails, also for the same owner:

  ```
  Error: Issue 'fix-login-bug' is in progress in another workspace, ../sisyphus-login, owned by a1, since 2026-10-11 09:12. Select another issue, or to take it, add --take.
  ```

  `--take` moves the claim to this workspace, with a warning.
- `list --ready` and `search --ready` leave out an open issue that another workspace has claimed.
- A claim ends when the issue leaves in-progress in its workspace: a close, an abandon, or a stop
  to open. A stop or a close in another workspace leaves the claim, with a warning.
- A claim is **stale** when its workspace no longer has the issue in progress, for example because
  the workspace was removed, or its change was abandoned. sisyphus ignores a stale claim, and a start
  replaces it with a warning. `--prune` removes all of them.

Each claim has a status:

| Status | Meaning |
|---|---|
| `here` | This workspace has the claim. |
| `active` | Another workspace has the claim. |
| `stale` | The workspace no longer has the issue in progress. |

Claims need jj or git: without them, a directory has only one working copy, and
`sisyphus list --state in-progress` shows the issues in progress. A repo on another machine does not
see the claims. An issue that started before sisyphus 0.0.89 has no claim until it starts again. See
[[0024-share-claims-between-workspaces]].

## Output

A table with the columns `name`, `owner`, `status`, `workspace` (the path from this workspace),
`bookmark`, and `since` (local time). With no claims, a message on stderr and no table.

With `--json`, an array of objects with `name`, `owner`, `workspace` (an absolute path),
`bookmark`, `since` (RFC 3339), and `status`. Empty fields are left out, except `status`.

## Examples

```bash
sisyphus claims
sisyphus claims --json | jq -r '.[] | select(.status == "active") | .name'
sisyphus claims --prune                    # remove the claims of removed workspaces
```
