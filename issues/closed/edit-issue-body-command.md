---
title: "Add a command to edit a section of an issue's body"
state: closed          # open | in-progress | closed. Must match the directory of the file.
resolution: completed  # completed | abandoned. Set this only when state is closed.
priority: high         # critical | high | medium | low
effort: medium         # small | medium | large | x-large
tags: [feature, cli, mcp] # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-09
closed: 2026-10-09     # YYYY-MM-DD. Set this only when state is closed.
owner: claude          # The person or agent working on the issue. Cleared when the issue returns to open.
approver:              # The person or agent who accepts the issue when it closes.
bookmark: samw/ai/edit-issue-sections # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: [edit-issue-sections] # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
metadata: {session-id: "e70e78a5-5e37-4f8f-aaa8-6dad55126d53"} # Optional. Arbitrary key-value notes, for example an AI agent session id to resume work with context: {session-id: "abc123"}.
deferred-from: "[[show-prints-resolution]]" # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent: "[[issue-manager-gaps]]" # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
depends-on: []           # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# Add a command to edit a section of an issue's body

## Summary

Add a command that replaces or appends to one section of an issue's body, for example `Resolution`, so that an agent can complete an issue without a hand edit.

## Context

From the human, 2026-10-09: a test run of Claude as a sisyphus task manager found this gap. Six times, after Claude created or closed an issue, it reported that the Summary or Resolution section still held the template placeholder, and that no command could fill it in.

`sisyphus new --context` fills only the Context section, and only at creation. Every other change to the body means a hand edit of the Markdown file. An agent that uses only sisyphus-mcp cannot edit the file.

## Acceptance criteria

- [x] A command replaces the text of one section of an issue's body
- [x] The command can append to a section, for example `Notes`, and an append to a section that holds only the template placeholder replaces the placeholder
- [x] An unknown section is an error that lists the sections of the issue
- [x] `sisyphus update <name> closed` warns when the `Resolution` section still holds the template placeholder
- [x] sisyphus-mcp has a tool for the command
- [x] The docs describe the command

## Out of scope

- A change to the title. The title is in the frontmatter and in the first heading.

## Notes

2026-10-09: added `sisyphus edit` and the `sisyphus_edit` MCP tool. `update` warns on close when `Resolution` holds the placeholder. sisyphus-mcp now returns warnings after the output, because it dropped them on success before. [[0010-edit-one-section-of-the-body]] records the design.

## Resolution

Completed in sisyphus [[0.0.31]] on bookmark samw/ai/edit-issue-sections. It waits for a human merge. See [[0010-edit-one-section-of-the-body]].

