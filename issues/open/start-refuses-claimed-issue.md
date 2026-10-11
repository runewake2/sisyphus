---
title: "Refuse to start an issue that another owner has"
state: open            # open | in-progress | closed. Must match the directory of the file.
resolution:            # completed | abandoned. Set this only when state is closed.
priority: high         # critical | high | medium | low
effort: small          # small | medium | large | x-large
tags: [bug, cli, llm]  # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-11
closed:                # YYYY-MM-DD. Set this only when state is closed.
owner:                 # The person or agent working on the issue. Cleared when the issue returns to open.
approver:              # The person or agent who accepts the issue when it closes.
bookmark:              # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: []         # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
metadata: {}           # Optional. Arbitrary key-value notes, for example an AI agent session id to resume work with context: {session-id: "abc123"}.
deferred-from:         # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent: "[[agent-cooperation]]" # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
depends-on: []           # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# Refuse to start an issue that another owner has

## Summary

`sisyphus start` and `update <name> in-progress` replace the owner of an issue that is already in progress, with no error and no warning. A second agent can take an issue from the first agent and not know it. Refuse the change, and tell who has the issue.

## Context

AGENTS.md tells agents to check `sisyphus list --state in-progress` and then start the issue. Between the check and the start, another agent can start it. Also, an agent that does not check gets no protection at all.

In a test repo:

```console
$ sisyphus start fix-login-bug --owner a1
issues/in-progress/fix-login-bug.md
$ sisyphus start fix-login-bug --owner a2
issues/in-progress/fix-login-bug.md        # exit 0, the owner is now a2
$ sisyphus update fix-login-bug in-progress --owner a3
issues/in-progress/fix-login-bug.md        # exit 0, the owner is now a3
```

Proposed behavior:

- `start` of an issue that is in progress with a different owner is an error: `Issue 'fix-login-bug' is in progress, owned by a1. To take it, add --take.`
- `start` by the same owner is a success that changes nothing, so an agent can retry a call.
- `--take` (CLI) and `take` (MCP) change the owner, and print a warning that names the old owner.
- The same rule for `update <name> in-progress --owner <other>` and for `update --owner` on an issue in progress.

This fix does not stop two processes that start the issue at the same moment. [[atomic-issue-writes]] covers that.

## Acceptance criteria

- [ ] `start` with a different owner of an in-progress issue fails, and the error names the owner and `--take`.
- [ ] `start` with the same owner succeeds and changes nothing.
- [ ] `--take` changes the owner and warns with the old owner.
- [ ] `update` follows the same rules, and the MCP tools pass `take`.
- [ ] Tests cover each case, and [[sisyphus-start]] and [[sisyphus-update]] describe them.

## Out of scope

- Claims that are in other workspaces or branches: [[claims-across-workspaces]].

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

<Complete this section when you close the issue.
- Completed: tell what was done. Link to the changelog version and the decision records, for example `[[0.0.5]]`.
- Abandoned: tell why the work stopped. Link to the issue or decision that replaces it, if one exists.>
