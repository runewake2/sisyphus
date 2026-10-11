# 0020: Do not assume jj outside a jj repo

- Status: Accepted
- Date: 2026-10-11
- Version: [[0.0.77]]

## Context

sisyphus must work in any repo, but the template that `init` writes described `bookmark` and
`workspaces` as jj concepts, and `update ... in-progress` warned about a missing jj bookmark in every
repo, also in a repo that uses only git ([[vcs-neutral-defaults]]). A repo that removed these fields
from its template got them back: `document.set` added a missing field even to clear it.

## Decision

- The warning about a missing bookmark comes only in a jj repo: a `.jj` directory at the root.
- The template keeps one form. `bookmark` is "the jj bookmark or git branch of the work", and
  `workspaces` "the jj workspaces or git worktrees where local work on the issue has happened". The
  help of `--bookmark` and `--workspace` says the same.
- `bookmark` and `workspaces` are optional in the template. Setting a field that the frontmatter does
  not have adds it only when the value is not empty, so clearing a missing field adds nothing.

## Alternatives considered

- **A second template without the jj fields for repos without jj.** Two templates to keep in step,
  and a repo that starts to use jj later would have the wrong one. The fields are useful with git
  branches too.
- **Detect the bookmark from the VCS.** Useful, but a separate feature; it can be an issue of its own.

## Consequences

- A git-only repo gets no jj warning, and can still record a branch in `bookmark`.
- A repo can remove `bookmark` and `workspaces` from its template, and every command keeps working.
