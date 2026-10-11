---
title: "Add a UX researcher agent that finds problems in how people and agents use sisyphus"
state: open            # open | in-progress | closed. Must match the directory of the file.
resolution:            # completed | abandoned. Set this only when state is closed.
priority: medium       # critical | high | medium | low
effort: large          # small | medium | large | x-large
tags: [ux, llm, hephaestus] # The components that the work touches, for example [widget-scheduler, plan]
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

# Add a UX researcher agent that finds problems in how people and agents use sisyphus

## Summary

Add a UX researcher role: an agent that uses what is being built as a new person and as an agent would, and files one issue for each problem or improvement that it finds, with what it saw.

## Context

[[agent-delegation#The other roles]] describes the role. This repo already has two such reviews: the review that filed `cli-and-docs-ux` (a first session, every `--help`, every docs page, the README on GitHub), and the review that filed `agent-cooperation` (the CLI through a shell tool with no terminal, and sisyphus-mcp through a JSON-RPC client). The spike the sisyphus spike `spike-ollama-local-models` measured how small models use the tools. The role makes these reviews repeatable, for example after each epic closes.

## Acceptance criteria

- [ ] `hephaestus/roles/ux-researcher.md` describes the role and its review methods (new person, agent through a shell, MCP client, small model).
- [ ] A run files issues with what it saw, and does not file an issue that already exists.
- [ ] A human checks the issues of one run.

## Out of scope

- Fixing the problems that it finds.

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

<Complete this section when you close the issue.
- Completed: tell what was done. Link to the changelog version and the decision records, for example `[[0.0.5]]`.
- Abandoned: tell why the work stopped. Link to the issue or decision that replaces it, if one exists.>
