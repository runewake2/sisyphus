---
title: "Add a security researcher agent that investigates security practices and data protection"
state: open            # open | in-progress | closed. Must match the directory of the file.
resolution:            # completed | abandoned. Set this only when state is closed.
priority: medium       # critical | high | medium | low
effort: large          # small | medium | large | x-large
tags: [hephaestus, llm, security] # The components that the work touches, for example [widget-scheduler, plan]
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

# Add a security researcher agent that investigates security practices and data protection

## Summary

Add a security researcher role: an agent that investigates the security practices of the repo and how it protects data (secrets, file permissions, input that reaches a shell or a path, services, dependencies, and what agents can do), and files one issue for each weakness, with its class and place, never an exploit.

## Context

Requested by a human on 2026-10-12, with the request to implement the AI roles. The role file [[roles/security-researcher|security-researcher]] defines the role; it runs as a Claude Code subagent or a worker with a runner ([[role-file#Running an AI role]]). Its scope as a reviewer in a referendum is the security of the change.

## Acceptance criteria

- [x] `hephaestus/roles/security-researcher.md` describes the role, its checklist, and the points where it must ask a human.
- [ ] A run on a scope of this repo files issues with evidence, and no issue that exists already.
- [ ] A human checks the issues of one run.

## Out of scope

- Fixing what it finds: other roles and humans do.

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

<Complete this section when you close the issue.
- Completed: tell what was done. Link to the changelog version and the decision records, for example `[[0.0.5]]`.
- Abandoned: tell why the work stopped. Link to the issue or decision that replaces it, if one exists.>
