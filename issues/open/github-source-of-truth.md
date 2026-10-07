---
title: "Use sisyphus as the source of truth for GitHub issues"
state: open            # open | in-progress | closed. Must match the directory of the file.
resolution:            # completed | abandoned. Set this only when state is closed.
priority: high         # critical | high | medium | low
effort: large          # small | medium | large | x-large
tags: [github, integration] # The components that the work touches, for example [widget-scheduler, plan]
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

- [ ] A sisyphus issue can record a remote reference ([[pin-issue-to-remote-reference]]).
- [ ] Filing a GitHub issue results in a PR that adds a matching, pinned sisyphus issue
      ([[github-action-issue-to-pr]]).
- [ ] Merging to `main` creates or updates GitHub issues for sisyphus issues that need one
      ([[github-action-sync-issues-to-github]]).
- [ ] The design is documented (in `CONTRIBUTING.md` or a decision record) as: sisyphus is always
      the source of truth for issues once a repo is initialized.

## Out of scope

- Syncing anything other than issues (no PRs, no milestones, no project boards).
- Two-way sync of issue state or comments from GitHub back into sisyphus; GitHub issues are a mirror.
- Jira-specific automation. Only the pinned reference ([[pin-issue-to-remote-reference]]) needs to
  support Jira; the two GitHub Actions are GitHub-only.

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

<Complete this section when you close the issue.
- Completed: tell what was done. Link to the changelog version and the decision records, for example [[0.0.5]].
- Abandoned: tell why the work stopped. Link to the issue or decision that replaces it, if one exists.>
