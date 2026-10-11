---
title: "Add sisyphus rename to rename an issue or move it to a directory"
state: closed          # open | in-progress | closed. Must match the directory of the file.
resolution: completed  # completed | abandoned. Set this only when state is closed.
priority: high         # critical | high | medium | low
effort: medium         # small | medium | large | x-large
tags: [cli, sisyphus-mcp, feature] # The components that the work touches, for example [widget-scheduler, plan]
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

# Add sisyphus rename to rename an issue or move it to a directory

## Summary

No command renames an issue or moves it to a subdirectory, and the docs say never to do it, because every link to the issue would break. Add sisyphus rename, which moves the file and changes every link to it, so that the issues of a repo can be organized in directories.

## Context

Requested by a human on 2026-10-11, to organize the issues of this repo in subdirectories. A move by hand breaks the parent, depends-on, and deferred-from links of other issues when they use the full name, and every link with the path. [[commands#Naming an issue]] said that a full name is permanent.

## Acceptance criteria

- [ ] `sisyphus rename <name> <new-name>` moves the file in its state directory, with `git mv` when git tracks it.
- [ ] Every wikilink to the issue in the repo changes, also in the frontmatter of other issues, and links in code do not.
- [ ] `--dry-run` shows the changes and changes nothing.
- [ ] The claim moves, and an issue that another workspace has in progress is refused.
- [ ] An MCP tool runs it, and the docs no longer say never to rename.

## Out of scope

- Commit messages, other branches, and other repos.
- Markdown links (`[text](path)`) to issue files.

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

Completed in [[0.0.105]]. `sisyphus rename` moves the file, changes every wikilink to it (frontmatter included, code excluded), moves the claim, and has `--dry-run`; `sisyphus_rename` runs it from MCP. See [[0027-rename-changes-every-link]].

