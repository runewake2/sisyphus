---
title: "Add database roles for SQLite, PostgreSQL, and Meilisearch"
state: open            # open | in-progress | closed. Must match the directory of the file.
resolution:            # completed | abandoned. Set this only when state is closed.
priority: medium       # critical | high | medium | low
effort: medium         # small | medium | large | x-large
tags: [hephaestus, llm, database] # The components that the work touches, for example [widget-scheduler, plan]
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

# Add database roles for SQLite, PostgreSQL, and Meilisearch

## Summary

Add a family of AI roles in hephaestus/roles/dba/: dba/sqlite, dba/postgres, and dba/meilisearch. Each checks the schema, keys, indexes, queries, data consistency, and migrations of its database on a scratch copy, fixes one finding at a time with evidence before and after, and files the rest.

## Context

Requested by a human on 2026-10-12: a role for DBA tasks that optimizes database implementations: better indexes, data consistency, proper SQL patterns, correct key values, and other database optimization tasks, maybe one for each type of database; for now SQLite, PostgreSQL, and Meilisearch. [[roles/dba/README|The database roles]] describes the family.

## Acceptance criteria

- [x] Each database has a role file with a checklist for its engine, and the family README has the rules of every DBA (never production, evidence before and after, consistency first, safe migrations, parameters).
- [x] Their Claude Code subagents are in .claude/agents/, and the implementer and the performance engineer name them in Works with.
- [ ] A trial run of each on a small fixture with known problems (a missing index, a missing key, a string-built query, a wrong Meilisearch setting) finds them with evidence.
- [ ] A human checks the result of each trial, and decides which roles stay.

## Out of scope

<List related work that this issue does not include. Create a separate issue for each item that someone must do later.>

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

<Complete this section when you close the issue.
- Completed: tell what was done. Link to the changelog version and the decision records, for example `[[0.0.5]]`.
- Abandoned: tell why the work stopped. Link to the issue or decision that replaces it, if one exists.>
