---
title: "Organize the issues of this repo in directories"
state: closed          # open | in-progress | closed. Must match the directory of the file.
resolution: completed  # completed | abandoned. Set this only when state is closed.
priority: medium       # critical | high | medium | low
effort: small          # small | medium | large | x-large
tags: [issues]         # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-11
closed: 2026-10-11     # YYYY-MM-DD. Set this only when state is closed.
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

# Organize the issues of this repo in directories

## Summary

All 122 issues of this repo were in the state directories with no subdirectories, so it was hard to find the issues of one area. Move each issue to a directory of its area with sisyphus rename.

## Context

Requested by a human on 2026-10-11, after [[rename-issue-command]]. A sub-issue goes to the directory of its epic. Each other issue goes to the directory of its area: agents, mcp, graph, performance, integrations, ux, commands, or issue-files.

## Acceptance criteria

- [ ] Every issue is in a directory below its state directory.
- [ ] `sisyphus lint` reports no problems.
- [ ] CONTRIBUTING.md tells where a new issue goes.

## Out of scope

New file names for issues.

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

Completed with [[rename-issue-command]] ([[0.0.105]]). All 122 issues moved to 8 area directories: agents, mcp, commands, graph, issue-files, performance, integrations, and ux. A sub-issue is in the directory of its epic. No file name changed, so only 88 frontmatter link lines changed, and no document. CONTRIBUTING.md tells where a new issue goes.

