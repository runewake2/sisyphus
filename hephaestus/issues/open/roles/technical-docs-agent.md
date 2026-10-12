---
title: "Add a technical docs agent that checks and writes documentation for completed work"
state: open            # open | in-progress | closed. Must match the directory of the file.
resolution:            # completed | abandoned. Set this only when state is closed.
priority: medium       # critical | high | medium | low
effort: large          # small | medium | large | x-large
tags: [docs, llm, hephaestus] # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-11
closed:                # YYYY-MM-DD. Set this only when state is closed.
owner:                 # The person or agent working on the issue. Cleared when the issue returns to open.
approver:              # The person or agent who accepts the issue when it closes.
bookmark:              # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: []         # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
metadata: {}           # Optional. Arbitrary key-value notes, for example an AI agent session id to resume work with context: {session-id: "abc123"}.
deferred-from:         # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent: "[[core/delegate-issues-to-agents]]" # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
depends-on: ["[[core/delegation-model]]"] # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# Add a technical docs agent that checks and writes documentation for completed work

## Summary

Add a technical docs role: an agent that takes each piece of completed work and checks it against the documentation, fixes what no longer agrees, and writes the documentation of a new feature.

## Context

[[agent-delegation#The other roles]] describes the role. Some checks are already tests: every command page has the sections and flags of its command, every help example is on its page, [[agent-guide]] uses only commands and flags that exist, and [[output]] names every JSON field. The role covers what tests cannot: the text says what the code does, the examples give the output that the page shows, and the writing follows [[CONTRIBUTING#Writing style]]. Its trigger is a close of an issue (Q5).

## Acceptance criteria

- [ ] `hephaestus/roles/technical-docs.md` describes the role.
- [ ] For a closed issue, it reports where the docs and the behavior disagree, and fixes the docs in a change, or files an issue.
- [ ] For a new feature, it writes or completes the page, in the format of the other pages.

## Out of scope

- Changes to behavior to match the docs. It files an issue instead.

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

<Complete this section when you close the issue.
- Completed: tell what was done. Link to the changelog version and the decision records, for example `[[0.0.5]]`.
- Abandoned: tell why the work stopped. Link to the issue or decision that replaces it, if one exists.>
