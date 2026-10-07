---
title: "Use sisyphus as the source of truth for GitHub issues"
state: closed          # open | in-progress | closed. Must match the directory of the file.
resolution: completed  # completed | abandoned. Set this only when state is closed.
priority: high         # critical | high | medium | low
effort: large          # small | medium | large | x-large
tags: [github, integration] # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-07
closed: 2026-10-07     # YYYY-MM-DD. Set this only when state is closed.
owner: claude          # The person or agent working on the issue. Cleared when the issue returns to open.
approver: samw         # The person or agent who accepts the issue when it closes.
bookmark: samw/ai/github-source-of-truth # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: [sisyphus-github-source-of-truth] # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
metadata: {session-id: "68fa9a09-fe4e-4ae8-9917-086c3313c5be"} # Optional. Arbitrary key-value notes, for example an AI agent session id to resume work with context: {session-id: "abc123"}.
deferred-from:         # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent:                # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
---

# Use sisyphus as the source of truth for GitHub issues

## Summary

Let a repo use `issues/` as the single source of truth for its work, while still giving humans a
GitHub issue to file, read, and discuss. Once a repo is initialized with sisyphus, sisyphus issues
win over GitHub issues whenever the two disagree: GitHub issues are a mirror and an intake point,
never an independent record.

## Context

Today `issues/` and GitHub issues are two unconnected systems: filing a GitHub issue does not create
a sisyphus issue, and a sisyphus issue has no link to any GitHub issue (or Jira ticket) that tracks
it for people who live in those tools. This epic covers the three pieces that close that gap:

- [[pin-issue-to-remote-reference]]: a field on a sisyphus issue that points at the GitHub issue or
  Jira ticket it corresponds to.
- [[github-action-issue-to-pr]]: a GitHub Action that reacts to a new GitHub issue by opening a PR
  that adds the matching sisyphus issue (pinned to that GitHub issue).
- [[github-action-sync-issues-to-github]]: a GitHub Action that, on `main`, creates or updates a
  GitHub issue for every sisyphus issue that does not already have one.

Together these make sisyphus the source of truth: every sisyphus issue has at most one matching
GitHub issue, every GitHub issue either came from, or was turned into, a sisyphus issue, and content
always flows sisyphus -> GitHub, never the other way, once a repo is initialized.

## Acceptance criteria

- [x] A sisyphus issue can record a remote reference ([[pin-issue-to-remote-reference]]).
- [x] Filing a GitHub issue results in a PR that adds a matching, pinned sisyphus issue
      ([[github-action-issue-to-pr]]).
- [x] Merging to `main` creates or updates GitHub issues for sisyphus issues that need one
      ([[github-action-sync-issues-to-github]]).
- [x] The design is documented (in `CONTRIBUTING.md` or a decision record) as: sisyphus is always
      the source of truth for issues once a repo is initialized.

## Out of scope

- Syncing anything other than issues (no PRs, no milestones, no project boards).
- Two-way sync of issue state or comments from GitHub back into sisyphus; GitHub issues are a mirror.
- Jira-specific automation. Only the pinned reference ([[pin-issue-to-remote-reference]]) needs to
  support Jira; the two GitHub Actions are GitHub-only.

## Notes

- 2026-10-07: All three sub-issues landed: [[pin-issue-to-remote-reference]] (earlier),
  [[github-action-issue-to-pr]], and [[github-action-sync-issues-to-github]]. Documented the
  one-way "sisyphus wins" design as its own `### GitHub issues` subsection in `CONTRIBUTING.md`
  (and the kit template), rather than a decision record, since it is describing the workflow this
  epic just built rather than a design choice among alternatives.
- While adding the two workflow kit templates, found and filed [[init-force-resets-version]]: a
  real, unrelated bug where `sisyphus init --force` destructively resets `VERSION`/`CHANGELOG.md`
  on an already-versioned repo. Caught it on this repo itself via `jj restore`; left it as a
  separate issue rather than fixing it inline here.

## Resolution

Completed: sisyphus issues can be pinned to a GitHub issue, filing a GitHub issue opens a PR with a
matching sisyphus issue, pushing to `main` mirrors sisyphus issues to GitHub, and the one-way
source-of-truth design is documented in [[CONTRIBUTING]]. See [[0.0.9]].
