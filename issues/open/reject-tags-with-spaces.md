---
title: "Reject tags that contain spaces"
state: open            # open | in-progress | closed. Must match the directory of the file.
resolution:            # completed | abandoned. Set this only when state is closed.
priority: low          # critical | high | medium | low
effort: small          # small | medium | large | x-large
tags: [bug, cli, obsidian] # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-11
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

# Reject tags that contain spaces

## Summary

`sisyphus new --tags "needs review, bug"` writes the tag `needs review`. A tag with a space is not a valid Obsidian tag, and sisyphus says that its issues work as an Obsidian vault.

## Context

Found in a review of sisyphus 0.0.31 on 2026-10-11. The README section "Obsidian" says that a repo with sisyphus works as an Obsidian vault. Obsidian reads the `tags` frontmatter property as tags, and a tag cannot contain spaces. The `list` table also prints `needs review, bug`, which reads like three tags.

## Acceptance criteria

- [ ] `new` and `update --tags` reject a tag that contains whitespace, and suggest the kebab-case form (`needs-review`).
- [ ] Tests cover the error.

## Out of scope

- Other Obsidian tag rules, for example tags that are only digits. Add them if they cause a problem.

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

<Complete this section when you close the issue.
- Completed: tell what was done. Link to the changelog version and the decision records, for example `[[0.0.5]]`.
- Abandoned: tell why the work stopped. Link to the issue or decision that replaces it, if one exists.>
