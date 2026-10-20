---
title: "repo-check fails on a machine where sisyphus is not on PATH"
state: open            # open | in-progress | closed. Must match the directory of the file.
resolution:            # completed | abandoned. Set this only when state is closed.
priority: medium       # critical | high | medium | low
effort: medium         # small | medium | large | x-large
tags: []               # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-20
closed:                # YYYY-MM-DD. Set this only when state is closed.
owner:                 # The person or agent working on the issue. Cleared when the issue returns to open.
approver:              # The person or agent who accepts the issue when it closes.
bookmark:              # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: []         # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
metadata: {}           # Optional. Arbitrary key-value notes, for example an AI agent session id to resume work with context: {session-id: "abc123"}.
deferred-from:         # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent:                # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
depends-on: []           # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# repo-check fails on a machine where sisyphus is not on PATH

## Summary

The command of repo-check calls sisyphus lint from PATH (hephaestus/roles/repo-check.md, line 4). On a machine where sisyphus is not installed, every task of repo-check fails and tags its issue needs-human, although the repo itself is clean. The merge-queue role does not have this problem: its checks run go run ./cmd/sisyphus lint (hephaestus/roles/merge-queue.md, command line). Evidence: in a fresh sandbox of this repo, hephaestus status stopped with: sisyphus is not installed or not on PATH (2026-10-20, operations auditor). Proposed change: the command of repo-check runs go run ./cmd/sisyphus lint, as the merge-queue checks do. Also decide whether repo-check must run go test ./...: the design says the checks of the repo are gofmt, go vet, the tests, and sisyphus lint (design/hephaestus/agent-delegation.md, Each merge), but repo-check runs no tests, and only the merge-queue checks do.

## Context

<Give the background that an agent needs to start the work without other help.
Link to design docs and other issues with wikilinks, for example `[[widget-scheduler]]`,
`[[widget-scheduler#Status API]]`, or `[[<issue-name>]]`. Do not use paths.>

## Acceptance criteria

- [ ] A human accepts the change of the command of hephaestus/roles/repo-check.md.
- [ ] repo-check passes on a clean repo on a machine with Go but without an installed sisyphus.
- [ ] A decision in the role file: repo-check runs go test ./..., or its charter says why not.

## Out of scope

<List related work that this issue does not include. Create a separate issue for each item that someone must do later.>

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

<Complete this section when you close the issue.
- Completed: tell what was done. Link to the changelog version and the decision records, for example `[[0.0.5]]`.
- Abandoned: tell why the work stopped. Link to the issue or decision that replaces it, if one exists.>
