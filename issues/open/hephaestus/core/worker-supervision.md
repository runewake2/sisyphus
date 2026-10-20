---
title: "Supervise the workers: restart, concurrency, and a home for the runners"
state: open            # open | in-progress | closed. Must match the directory of the file.
resolution:            # completed | abandoned. Set this only when state is closed.
priority: medium       # critical | high | medium | low
effort: medium         # small | medium | large | x-large
tags: [hephaestus, llm] # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-20
closed:                # YYYY-MM-DD. Set this only when state is closed.
owner:                 # The person or agent working on the issue. Cleared when the issue returns to open.
approver:              # The person or agent who accepts the issue when it closes.
bookmark:              # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: []         # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
metadata: {}           # Optional. Arbitrary key-value notes, for example an AI agent session id to resume work with context: {session-id: "abc123"}.
deferred-from:         # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent: "[[hephaestus/core/delegate-issues-to-agents]]" # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
depends-on: ["[[hephaestus/core/delegation-runner]]"] # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# Supervise the workers: restart, concurrency, and a home for the runners

## Summary

The queue knows live and stopped workers, but nothing acts on it: a crashed worker stays down until a human runs hephaestus agent run again, nothing bounds how many workers of a role run at once, and messages can pile up for a role with no live worker while status can only tell a human to start one. A minimal supervisor closes the loop of the delegation runner: it restarts a worker that stopped (with backoff, and never past a failure that needs a human), holds a per-role concurrency bound, and starts or alerts when deliveries wait for a role with no live worker. It also needs the answer the runner has needed all along: where the runner command and its credentials live per role, because hephaestus never starts a model without a runner, and today the runner is a flag or an environment variable of the human who starts the worker.

## Context

<Give the background that an agent needs to start the work without other help.
Link to design docs and other issues with wikilinks, for example `[[widget-scheduler]]`,
`[[widget-scheduler#Status API]]`, or `[[<issue-name>]]`. Do not use paths.>

## Acceptance criteria

- [ ] A decision record says where the runner command and its credentials of each role live, and the one-host boundary it shares with [[delegation-runner]].
- [ ] A stopped worker with pending deliveries is restarted with backoff, or reported for a human when restarting would repeat a failure (needs-human stays a hard stop).
- [ ] A per-role concurrency bound holds: no more workers of a role than it allows.
- [ ] doctor and status show what the supervisor did: the restarts, and the roles at their bound.
- [ ] The supervisor never starts a model without a configured runner, as hephaestus promises today.

## Out of scope

<List related work that this issue does not include. Create a separate issue for each item that someone must do later.>

## Notes

2026-10-20, from a human: track this; it was found in the review of what hephaestus needs to work as a job coordination service, as the gap between the queue knowing a worker stopped and anything acting on it. Fold it into the scope of [[delegation-runner]] or build it beside the daemon; the dependency records the order. The eviction and kick commands stay the manual tools; the supervisor automates only the safe cases.

## Resolution

<Complete this section when you close the issue.
- Completed: tell what was done. Link to the changelog version and the decision records, for example `[[0.0.5]]`.
- Abandoned: tell why the work stopped. Link to the issue or decision that replaces it, if one exists.>
