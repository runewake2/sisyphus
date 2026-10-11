---
title: "Let an agent see the claims of agents in other workspaces"
state: closed          # open | in-progress | closed. Must match the directory of the file.
resolution: completed  # completed | abandoned. Set this only when state is closed.
priority: high         # critical | high | medium | low
effort: large          # small | medium | large | x-large
tags: [llm, cli, spike] # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-11
closed: 2026-10-11     # YYYY-MM-DD. Set this only when state is closed.
owner: claude          # The person or agent working on the issue. Cleared when the issue returns to open.
approver:              # The person or agent who accepts the issue when it closes.
bookmark: samw/ai/ux-fixes # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: [ux-fixes] # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
metadata: {}           # Optional. Arbitrary key-value notes, for example an AI agent session id to resume work with context: {session-id: "abc123"}.
deferred-from:         # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent: "[[agent-cooperation]]" # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
depends-on: []           # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# Let an agent see the claims of agents in other workspaces

## Summary

An agent sees only the claims in its own working copy. Agents that work in separate jj workspaces or git worktrees cannot see that another agent started an issue until that work merges. Find a way for `sisyphus` to show, and check, the claims of all workspaces of a repo.

## Context

AGENTS.md tells each agent to work in its own jj workspace, based on `main`, and to run `sisyphus list --state in-progress` before it takes an issue. But the issue file moves to `issues/in-progress/` only in the working copy of the agent that started it. In every other workspace, the issue stays in `issues/open/` until a human merges the work into `main`. So the check of AGENTS.md finds only the claims that are already merged, and two agents can start the same issue.

[[start-refuses-claimed-issue]] helps only in one working copy.

Options to examine:

- Read the issue files of every workspace or worktree of the repo (`jj workspace list`, `git worktree list`) and of every bookmark or branch, and report an issue as claimed if any of them has it in progress.
- Keep claims outside the working copy, in a place that all workspaces share, for example a file in `.jj/` or `.git/`, or a git ref such as `refs/sisyphus/claims`.
- A claim that a remote also shares, for agents on different machines.

Each option must work with git alone, with jj, and with no VCS ([[0020-do-not-assume-jj]]).

## Acceptance criteria

- [ ] A decision record compares the options and selects one.
- [ ] `sisyphus list --state in-progress` (or a new command) shows the claims of the other workspaces of the repo.
- [ ] `sisyphus start` refuses an issue that an agent in another workspace has started.
- [ ] AGENTS.md and [[agent-workflow-guide]] tell agents how to use it.

## Out of scope

- Claims between different clones on different machines, unless the selected option gives them at no extra cost.

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

Completed in [[0.0.89]]. A start records a claim in the jj repo directory or the git common directory, and a start of an issue that another workspace has claimed fails unless `--take` is given. Stale claims are ignored. `sisyphus claims` and `sisyphus_claims` list them, and AGENTS.md uses them. See [[0024-share-claims-between-workspaces]].

