---
title: "Rewrite the README for people who use sisyphus in their own repo"
state: open            # open | in-progress | closed. Must match the directory of the file.
resolution:            # completed | abandoned. Set this only when state is closed.
priority: high         # critical | high | medium | low
effort: small          # small | medium | large | x-large
tags: [docs]           # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-11
closed:                # YYYY-MM-DD. Set this only when state is closed.
owner:                 # The person or agent working on the issue. Cleared when the issue returns to open.
approver:              # The person or agent who accepts the issue when it closes.
bookmark:              # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: []         # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
metadata: {}           # Optional. Arbitrary key-value notes, for example an AI agent session id to resume work with context: {session-id: "abc123"}.
deferred-from:         # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent: "[[cli-and-docs-ux]]" # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
depends-on: []           # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# Rewrite the README for people who use sisyphus in their own repo

## Summary

The README is the first page that people see, but it is written for this repo: it links to CONTRIBUTING for "the full workflow" (jj, versioning, the design log), uses wikilinks that GitHub shows as raw `[[commands]]` text, and has no walk-through with real output. Rewrite it for people who use sisyphus in their own repo.

## Context

Found in a review of the CLI, the help text, and the documentation of sisyphus 0.0.67 on 2026-10-11, with a first session in an empty repo. Points to fix:

- GitHub does not render `[[commands]]`, `[[CONTRIBUTING]]`, or `[[sisyphus-mcp]]`; a reader on GitHub sees the
  brackets and cannot click them. The README needs normal Markdown links (it can keep wikilinks where Obsidian
  is the reader, for example in `docs/`, if `links` and `lint` also accept Markdown links).
- "See [[CONTRIBUTING]] for the full workflow: the design log, decisions, versioning, the changelog,
  wikilinks, issues, and jj" describes how this repo is built, not how to use sisyphus.
- There is no example of a whole session with its output (init, new, start, list, close), and no picture of
  what an issue file looks like ([[issue-format-guide]]).
- The MCP section has one line of more than 120 characters with every tool name.

## Acceptance criteria

- [ ] The README has: what sisyphus is, install, a short session with real output, the main commands, the Obsidian and MCP sections, and links to `docs/`.
- [ ] Every link in the README works on GitHub.
- [ ] The contributor workflow of this repo is linked from one line for contributors, not as part of usage.

## Out of scope

- The "DO NOT USE" note: that is the decision of the owner.

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

<Complete this section when you close the issue.
- Completed: tell what was done. Link to the changelog version and the decision records, for example `[[0.0.5]]`.
- Abandoned: tell why the work stopped. Link to the issue or decision that replaces it, if one exists.>
