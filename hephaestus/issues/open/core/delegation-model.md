---
title: "Decide how an issue is delegated to a role, and where a role is described"
state: open            # open | in-progress | closed. Must match the directory of the file.
resolution:            # completed | abandoned. Set this only when state is closed.
priority: high         # critical | high | medium | low
effort: medium         # small | medium | large | x-large
tags: [design, hephaestus] # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-11
closed:                # YYYY-MM-DD. Set this only when state is closed.
owner:                 # The person or agent working on the issue. Cleared when the issue returns to open.
approver:              # The person or agent who accepts the issue when it closes.
bookmark:              # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: []         # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
metadata: {}           # Optional. Arbitrary key-value notes, for example an AI agent session id to resume work with context: {session-id: "abc123"}.
deferred-from:         # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent: "[[core/delegate-issues-to-agents]]" # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
depends-on: []           # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# Decide how an issue is delegated to a role, and where a role is described

## Summary

Decide how a human or an agent delegates an issue to a role, how a role agent finds its work, and where each role is described. The other issues of [[delegate-issues-to-agents]] build on this. Answer Q1 and Q2 of [[agent-delegation#Open questions]].

## Context

The draft in [[agent-delegation#How an issue gets to an agent]] proposes:

- Delegate an issue to a role with the generic `metadata` field of sisyphus: `sisyphus update <name> --metadata delegate=merge-queue`. sisyphus does not know what the key means, so it needs no field for agents ([[agent-delegation#What hephaestus needs from sisyphus]]).
- Find the work of a role with a sisyphus filter on `metadata`, which the sisyphus issue `filter-by-metadata` adds for any client.
- One file for each role, `hephaestus/roles/<role>.md`: the charter, the inputs, the outputs, the allowed tools, the points where the agent must ask a human, and the prompt.
- Owner names of role agents: the role id and an instance, for example `merge-queue-1`, in `SISYPHUS_OWNER`.

Compare at least two drafts (for example the field against a tag), as `CONTRIBUTING.md` (How to do design work) asks.

## Acceptance criteria

- [ ] A decision record answers Q1 and Q2.
- [ ] hephaestus records the role of an issue in `metadata`, and finds the issues of a role with the filter of the sisyphus issue `filter-by-metadata`.
- [ ] A role description format exists, with one example: `hephaestus/roles/merge-queue.md`.

## Out of scope

- Running the agents: [[delegation-runner]].

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

<Complete this section when you close the issue.
- Completed: tell what was done. Link to the changelog version and the decision records, for example `[[0.0.5]]`.
- Abandoned: tell why the work stopped. Link to the issue or decision that replaces it, if one exists.>
