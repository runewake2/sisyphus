---
title: "Add a triage agent that makes new issues ready for work"
state: open            # open | in-progress | closed. Must match the directory of the file.
resolution:            # completed | abandoned. Set this only when state is closed.
priority: medium       # critical | high | medium | low
effort: medium         # small | medium | large | x-large
tags: [hephaestus, llm] # The components that the work touches, for example [widget-scheduler, plan]
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

# Add a triage agent that makes new issues ready for work

## Summary

Add the AI role triage: for each new issue, it finds duplicates, the area, the parent and dependencies, writes checkable acceptance criteria, and proposes a priority and a role. It does not delegate or close.

## Context

Requested by a human on 2026-10-12: implement other agents that may be useful, for a government of agents who coordinate and work together. The roles may be removed after a review; a trial run shows how useful each is. The role file is [[roles/triage]]; its section Works with tells how it works with the other roles.

## Acceptance criteria

- [x] The role file describes the role, its hand-offs in Works with, and the points where it must ask a human.
- [ ] A trial run on three new or vague issues makes each ready for work, with each change in the Notes and no duplicate closed.
- [ ] A human checks the result of the trial run, and decides whether the role stays.

## Out of scope

<List related work that this issue does not include. Create a separate issue for each item that someone must do later.>

## Notes

2026-10-20, triage trial run on five leaf issues. One-line verdicts:
- [[small-model-friendly-mcp]]: ready; all sub-issues closed, only the harness rerun remains; proposed priority medium, role model-eval.
- [[ollama-task-manager-guide]]: ready but blocked; added depends-on [[small-model-friendly-mcp]], noted that the dir argument is now hidden by default; proposed priority medium, role technical-docs.
- [[migrate-command]]: ready; not stale (no migrate command exists); one design question for a human (per-issue or per-repo version stamp); proposed priority low, role implementer.
- [[compare-cloud-claude-models]]: ready but needs a human first (API key without rate limit, approval of 30 to 50 USD); proposed priority low, role model-eval, whose role file covers only local models today.
- [[pin-github-actions]]: ready; still valid (ci.yml unpinned); moved from issue-files/ to repo/; added one proposed criterion (verify each SHA against the release); proposed priority medium, role implementer.
No duplicate found, and no issue closed, abandoned, or delegated. Every change is recorded in a dated note of its issue.

## Resolution

<Complete this section when you close the issue.
- Completed: tell what was done. Link to the changelog version and the decision records, for example `[[0.0.5]]`.
- Abandoned: tell why the work stopped. Link to the issue or decision that replaces it, if one exists.>
