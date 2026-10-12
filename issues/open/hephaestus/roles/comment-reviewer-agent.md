---
title: "Add a comment reviewer that makes code comments and commit messages brief and correct"
state: open            # open | in-progress | closed. Must match the directory of the file.
resolution:            # completed | abandoned. Set this only when state is closed.
priority: medium       # critical | high | medium | low
effort: medium         # small | medium | large | x-large
tags: [hephaestus, llm, docs] # The components that the work touches, for example [widget-scheduler, plan]
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

# Add a comment reviewer that makes code comments and commit messages brief and correct

## Summary

Add the AI role technical-docs/comments, a variant of technical-docs: it reviews the code comments and the commit messages of a bookmark from only the diff, the code, and the commit descriptions, fixes the comments on its own stacked bookmark, and proposes corrected descriptions. The implementer and the engineer run it in a review loop, up to three times.

## Context

Requested by a human on 2026-10-12: a version of technical-docs whose job is to review code comments and commit messages, and make them brief, concise, and correct, in Simplified Technical English. It is often paired with an engineer in an implement -> review loop. It is separate from the implementation on purpose, so that it reviews only the produced artifacts, with no solution context. The role file is [[roles/technical-docs/comments/role]].

## Acceptance criteria

- [x] `hephaestus/roles/technical-docs/comments/role.md` describes the role: inputs limited to the produced artifacts, the checklist from CONTRIBUTING, its outputs, and its verdict.
- [x] The implementer and the engineer ask it for a review in a loop of at most three reviews, with only the change in the data of the request.
- [x] Its Claude Code subagent is in `.claude/agents/`.
- [ ] A loop with a real implementer on a small issue ends clean, and the comments and descriptions are shorter and correct.
- [ ] A human checks that the reviewer did not read the issue or the design.

## Out of scope

- Changes to code or behavior: it only changes comments.
- Rewriting the commits of another bookmark: the author applies the corrected descriptions.

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

<Complete this section when you close the issue.
- Completed: tell what was done. Link to the changelog version and the decision records, for example `[[0.0.5]]`.
- Abandoned: tell why the work stopped. Link to the issue or decision that replaces it, if one exists.>
