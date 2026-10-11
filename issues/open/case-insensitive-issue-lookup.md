---
title: "Find issues by name without regard to case, as links do"
state: open            # open | in-progress | closed. Must match the directory of the file.
resolution:            # completed | abandoned. Set this only when state is closed.
priority: low          # critical | high | medium | low
effort: small          # small | medium | large | x-large
tags: [bug, cli]       # The components that the work touches, for example [widget-scheduler, plan]
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

# Find issues by name without regard to case, as links do

## Summary

Wikilinks find an issue without regard to case, but commands do not: `sisyphus resolve "[[META-TEST]]"` finds `meta-test`, while `sisyphus show META-TEST` and `show "[[Meta-Test]]"` fail with "No issue named".

## Context

Found in a review of sisyphus 0.0.45 on 2026-10-11. The README says that sisyphus resolves wikilinks the way Obsidian does, ignoring case.
Every command takes `[[name]]` as a form of an issue name (see [[commands#Naming an issue]]), so the same link works
in a document and fails in a command. Issue names are always lowercase, so a case-insensitive lookup cannot be
ambiguous.

## Acceptance criteria

- [ ] Every command that takes an issue name finds the issue without regard to case.
- [ ] Error messages and output use the real, lowercase name.
- [ ] Tests cover `show`, `update`, and `depends-on`.

## Out of scope

- Names with uppercase letters; they stay invalid.

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

<Complete this section when you close the issue.
- Completed: tell what was done. Link to the changelog version and the decision records, for example `[[0.0.5]]`.
- Abandoned: tell why the work stopped. Link to the issue or decision that replaces it, if one exists.>
