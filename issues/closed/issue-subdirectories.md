---
title: "Support subdirectories below the state directories"
state: closed          # open | in-progress | closed. Must match the directory of the file.
resolution: completed  # completed | abandoned. Set this only when state is closed.
priority: high         # critical | high | medium | low
effort: medium         # small | medium | large | x-large
tags: [feature, cli]   # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-08
closed: 2026-10-08     # YYYY-MM-DD. Set this only when state is closed.
owner: claude          # The person or agent working on the issue. Cleared when the issue returns to open.
approver: samw         # The person or agent who accepts the issue when it closes.
bookmark: samw/ai/issue-subdirectories # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: [sisyphus-issue-subdirectories] # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
metadata: {session-id: "5f1f06e4-afca-4277-8f3c-b8c818789ed6"} # Optional. Arbitrary key-value notes, for example an AI agent session id to resume work with context: {session-id: "abc123"}.
deferred-from:         # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent:                # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
depends-on: []           # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# Support subdirectories below the state directories

## Summary

Let an issue sit in kebab-case subdirectories below its state directory, and let every command find it by its full name or by its file name.

## Context

A repo with many issues needs to group them, for example by project and component: issues/open/corp/cimgr/fix-capacity.md instead of issues/open/corp-cimgr-fix-capacity.md. The state directories stay as they are. The full name is the path below the state directory. The bare file name also resolves; if more than one issue has it, the command fails and lists the full names (decided 2026-10-08 by the human). An exact full name wins over a longer name that ends with it, so that a flat issue stays reachable.

## Acceptance criteria

- [x] `sisyphus new web/auth/fix-login-bug` writes `issues/open/web/auth/fix-login-bug.md`.
- [x] `sisyphus update` keeps the subdirectories when the state changes.
- [x] Every command accepts the full name, the end of the full name, a wikilink, a hashtag, and a path.
- [x] A file name that more than one issue has is refused, with the full names in the error.
- [x] `parent`, `depends-on`, and `deferred-from` store the full name; old bare links still resolve.
- [x] `list`, `search`, `show`, and `graph` print full names.
- [x] `new` warns, and does not fail, when another Markdown file has the same file name.

## Out of scope

A command that moves existing issues into subdirectories. See [[migrate-command]].

## Notes

2026-10-08, from the human: an ambiguous file name fails in every command, reads included. Consequence: a duplicate file name makes the bare form unusable for both issues, so `new` warns when it creates one.

## Resolution

Completed in [[0.0.18]].
