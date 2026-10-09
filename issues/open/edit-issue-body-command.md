---
title: "Add a command to edit a section of an issue's body"
state: open            # open | in-progress | closed. Must match the directory of the file.
resolution:            # completed | abandoned. Set this only when state is closed.
priority: high         # critical | high | medium | low
effort: medium         # small | medium | large | x-large
tags: [feature, cli, mcp] # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-09
closed:                # YYYY-MM-DD. Set this only when state is closed.
owner:                 # The person or agent working on the issue. Cleared when the issue returns to open.
approver:              # The person or agent who accepts the issue when it closes.
bookmark:              # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: []         # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
metadata: {}           # Optional. Arbitrary key-value notes, for example an AI agent session id to resume work with context: {session-id: "abc123"}.
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

- [ ] A command replaces the text of one section of an issue's body
- [ ] The command can append to a section, for example `Notes`, and an append to a section that holds only the template placeholder replaces the placeholder
- [ ] An unknown section is an error that lists the sections of the issue
- [ ] `sisyphus update <name> closed` warns when the `Resolution` section still holds the template placeholder
- [ ] sisyphus-mcp has a tool for the command
- [ ] The docs describe the command

## Out of scope

- A change to the title. The title is in the frontmatter and in the first heading.

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

<Complete this section when you close the issue.
- Completed: tell what was done. Link to the changelog version and the decision records, for example `[[0.0.5]]`.
- Abandoned: tell why the work stopped. Link to the issue or decision that replaces it, if one exists.>
