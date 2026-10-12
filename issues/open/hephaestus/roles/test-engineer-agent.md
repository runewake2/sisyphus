---
title: "Add a test engineer agent that makes sure the tests show each acceptance criterion"
state: open            # open | in-progress | closed. Must match the directory of the file.
resolution:            # completed | abandoned. Set this only when state is closed.
priority: medium       # critical | high | medium | low
effort: large          # small | medium | large | x-large
tags: [hephaestus, llm, tests] # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-12
closed:                # YYYY-MM-DD. Set this only when state is closed.
owner:                 # The person or agent working on the issue. Cleared when the issue returns to open.
approver:              # The person or agent who accepts the issue when it closes.
bookmark:              # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: []         # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
metadata: {}           # Optional. Arbitrary key-value notes, for example an AI agent session id to resume work with context: {session-id: "abc123"}.
deferred-from:         # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent: "[[hephaestus/core/delegate-issues-to-agents]]" # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
depends-on: []           # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# Add a test engineer agent that makes sure the tests show each acceptance criterion

## Summary

Add a test engineer role: an agent that maps each acceptance criterion of an issue to the test or evidence that shows it, and writes the missing tests on its own bookmark.

## Context

Requested by a human on 2026-10-12, from a review of the missing roles: technical-docs checks the docs of completed work, but no role checked its tests. The role file [[roles/test-engineer|test-engineer]] defines the role.

## Acceptance criteria

- [x] `hephaestus/roles/test-engineer.md` describes the role, its outputs, and the points where it must ask a human.
- [ ] A run on a real bookmark writes the table of criteria and evidence, and adds a test that fails without the change.
- [ ] A human checks the tests of one run.

## Out of scope

- Fixing the code when a test fails: it files an issue.

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

<Complete this section when you close the issue.
- Completed: tell what was done. Link to the changelog version and the decision records, for example `[[0.0.5]]`.
- Abandoned: tell why the work stopped. Link to the issue or decision that replaces it, if one exists.>
