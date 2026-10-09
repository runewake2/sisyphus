---
title: "Fill the gaps that keep sisyphus from being a complete issue manager"
state: open            # open | in-progress | closed. Must match the directory of the file.
resolution:            # completed | abandoned. Set this only when state is closed.
priority: high         # critical | high | medium | low
effort: large          # small | medium | large | x-large
tags: [cli]            # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-07
closed:                # YYYY-MM-DD. Set this only when state is closed.
owner:                 # The person or agent working on the issue. Cleared when the issue returns to open.
approver:              # The person or agent who accepts the issue when it closes.
bookmark:              # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: []         # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
metadata: {}           # Optional. Arbitrary key-value notes, for example an AI agent session id to resume work with context: {session-id: "abc123"}.
deferred-from:         # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent:                # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
---

# Fill the gaps that keep sisyphus from being a complete issue manager

## Summary

Today sisyphus can create an issue and change its state, but it has no way to list, show, edit, or
validate issues, and no way to express that one issue blocks another. An issue manager needs all of
these. This epic tracks the commands and schema additions that close that gap.

## Context

sisyphus currently has five commands: `new`, `update`, `parent`, `resolve`, `links`. Working with the
issues in `issues/` beyond creating and moving them means reading and grepping Markdown files by
hand. The sub-issues below each cover one missing piece:

- [[list-issues-command]]: list and filter issues from the CLI (no way to answer "what is open and
  high priority" today without `grep`).
- [[show-issue-command]]: print one issue without finding and opening its file by hand. Done.
- [[search-issues-command]]: find issues by frontmatter filters and free-text content, with results
  optionally scoped to one section (for example only the `## Summary`).
- [[update-editable-metadata]]: `priority`, `effort`, and `tags` can only be set at creation; there
  is no way to change them afterward except by hand-editing the frontmatter.
- [[issue-dependencies]]: `parent` only expresses a hierarchy (sub-issue of); there is no way to say
  one issue blocks another unrelated issue.
- [[lint-issues-command]]: nothing validates the whole `issues/` tree at once (state matching its
  directory, valid enum values, resolving wikilinks, no parent cycles); each check today only runs
  against the one issue a command touches.
- [[graph-view-command]]: no way to see a large task's whole family tree (an epic and its
  sub-issues, with state and depends-on edges) at a glance; only one issue at a time via `show`.
- [[migrate-command]]: a schema change means hand-editing every issue file; nothing upgrades
  existing issues to the installed sisyphus version's template.
- [[show-prints-resolution]]: the text output of `show` does not print `resolution`, so a reader
  cannot see whether a close or an abandon worked.
- [[edit-issue-body-command]]: no command edits a section of the body, so the `Summary` and
  `Resolution` sections keep the template placeholder unless someone edits the file by hand.

## Acceptance criteria

- [x] `sisyphus list` can filter and display issues ([[list-issues-command]]).
- [x] `sisyphus show` can print one issue ([[show-issue-command]]).
- [x] `sisyphus search` can find issues by filter and content, with section-scoped results
      ([[search-issues-command]]).
- [x] `sisyphus update` can change priority, effort, and tags ([[update-editable-metadata]]).
- [x] Issues can express depends-on/blocks relationships ([[issue-dependencies]]).
- [ ] `sisyphus lint` validates the whole `issues/` tree ([[lint-issues-command]]).
- [x] `sisyphus graph` shows a large task's family tree in the terminal ([[graph-view-command]]).
- [ ] `sisyphus migrate` upgrades every issue to the current schema ([[migrate-command]]).
- [x] `sisyphus show` prints the resolution ([[show-prints-resolution]]).
- [x] A command edits a section of an issue's body ([[edit-issue-body-command]]).

## Out of scope

- A web UI or TUI. sisyphus stays a CLI over Markdown files.
- Anything covered by [[github-source-of-truth]] (GitHub/Jira integration); that is a separate epic.
- Due dates, milestones, or time tracking. Not requested; split into a new issue if needed later.

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

<Complete this section when you close the issue.
- Completed: tell what was done. Link to the changelog version and the decision records, for example [[0.0.5]].
- Abandoned: tell why the work stopped. Link to the issue or decision that replaces it, if one exists.>
