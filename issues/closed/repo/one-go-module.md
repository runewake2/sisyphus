---
title: "Move hephaestus into the root Go module, with its command beside sisyphus"
state: closed          # open | in-progress | closed. Must match the directory of the file.
resolution: completed  # completed | abandoned. Set this only when state is closed.
priority: medium       # critical | high | medium | low
effort: medium         # small | medium | large | x-large
tags: [hephaestus, build] # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-13
closed: 2026-10-13     # YYYY-MM-DD. Set this only when state is closed.
owner: claude          # The person or agent working on the issue. Cleared when the issue returns to open.
approver:              # The person or agent who accepts the issue when it closes.
bookmark: samw/ai/ux-fixes # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: [ux-fixes] # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
metadata: {}           # Optional. Arbitrary key-value notes, for example an AI agent session id to resume work with context: {session-id: "abc123"}.
deferred-from:         # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent:                # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
depends-on: []           # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# Move hephaestus into the root Go module, with its command beside sisyphus

## Summary

Move the hephaestus command to cmd/hephaestus beside cmd/sisyphus and cmd/sisyphus-mcp, and its packages to internal/hephaestus/, in the one Go module of the repo. A test keeps the two tools from sharing code.

## Context

Requested by a human on 2026-10-13: hephaestus and sisyphus have different cmd directories; combine them so that the tools live beside one another. This changes the answer to Q7 of [[agent-delegation]] (hephaestus as a Go module of its own).

## Acceptance criteria

- [x] cmd/hephaestus is beside cmd/sisyphus and cmd/sisyphus-mcp; hephaestus/go.mod is gone, and its packages are in internal/hephaestus/.
- [x] gofmt, go vet ./..., and go test ./... at the root build and test both tools; CI, repo-check, and the merge queue run one set of checks.
- [x] A test fails when hephaestus imports sisyphus code, or sisyphus imports hephaestus code.
- [x] The docs, the role files, and a decision record describe the layout.

## Out of scope

<List related work that this issue does not include. Create a separate issue for each item that someone must do later.>

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

Completed: the hephaestus command is cmd/hephaestus beside the sisyphus commands, in the one Go module, with its packages in internal/hephaestus/; a test keeps the tools from sharing code. See [[unreleased-one-go-module]] and [[0003-one-go-module-with-the-commands-side-by-side]].

