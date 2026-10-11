---
title: "Stop assuming jj in the template and in warnings"
state: closed          # open | in-progress | closed. Must match the directory of the file.
resolution: completed  # completed | abandoned. Set this only when state is closed.
priority: medium       # critical | high | medium | low
effort: medium         # small | medium | large | x-large
tags: [ux, cli]        # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-11
closed: 2026-10-11     # YYYY-MM-DD. Set this only when state is closed.
owner: claude          # The person or agent working on the issue. Cleared when the issue returns to open.
approver:              # The person or agent who accepts the issue when it closes.
bookmark: samw/ai/ux-fixes # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: [ux-fixes] # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
metadata: {}           # Optional. Arbitrary key-value notes, for example an AI agent session id to resume work with context: {session-id: "abc123"}.
deferred-from:         # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent: "[[ux/cli-and-docs-ux]]" # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
depends-on: []           # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# Stop assuming jj in the template and in warnings

## Summary

sisyphus is meant to work in any repo, but the template that `init` writes has jj-only fields (`bookmark`, `workspaces`), and `update ... in-progress` warns about a missing jj bookmark in every repo. Make these depend on the repo, or on the template.

## Context

Found in a review of the CLI, the help text, and the documentation of sisyphus 0.0.67 on 2026-10-11, with a first session in an empty repo.

```
$ sisyphus update fix-login-bug in-progress        # in a repo without jj
Warning: The issue is in-progress but has no bookmark. Set it with --bookmark.
```

The embedded `issues/TEMPLATE.md` says "The jj bookmark of the work on the issue" and "jj workspaces where local
work on the issue has happened". CONTRIBUTING says that sisyphus must not know about any particular repo. A git
user gets warnings about a tool they do not use. Options: warn about the bookmark only when the repo uses jj (a
`.jj` directory), or only when the template has a `bookmark` field; and let `init` write a template without the
jj fields when the repo does not use jj.

## Acceptance criteria

- [ ] In a repo without jj, `update ... in-progress` gives no bookmark warning.
- [ ] `init` in a repo without jj writes a template without the jj-only fields, or the fields are described without jj; decide, and record it as a decision.
- [ ] Commands work with a template that has no `bookmark` or `workspaces` field.
- [ ] Tests cover a repo with and without jj.

## Out of scope

- Support for git branches as bookmarks. That can be a separate issue.

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

Completed in [[0.0.77]]. The bookmark warning is only for jj repos, the template describes git branches and worktrees too, and a template without the jj fields works. See [[0020-do-not-assume-jj]].

