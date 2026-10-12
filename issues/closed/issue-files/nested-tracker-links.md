---
title: "Resolve the links of a nested tracker in the whole repo"
state: closed          # open | in-progress | closed. Must match the directory of the file.
resolution: completed  # completed | abandoned. Set this only when state is closed.
priority: medium       # critical | high | medium | low
effort: small          # small | medium | large | x-large
tags: [bug, cli, lint] # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-11
closed: 2026-10-12     # YYYY-MM-DD. Set this only when state is closed.
owner: claude          # The person or agent working on the issue. Cleared when the issue returns to open.
approver:              # The person or agent who accepts the issue when it closes.
bookmark: samw/ai/ux-fixes # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: [ux-fixes] # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
metadata: {}           # Optional. Arbitrary key-value notes, for example an AI agent session id to resume work with context: {session-id: "abc123"}.
deferred-from:         # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent:                # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
depends-on: []           # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# Resolve the links of a nested tracker in the whole repo

## Summary

In a nested tracker (init --nested), lint and links resolve wikilinks only below the nested tracker, so a link to a document of the repo, for example [[AGENTS]], is reported missing. Obsidian resolves it, because its vault is the repo. Decide the scope of links in a nested tracker, and make lint agree with Obsidian.

## Context

Found on 2026-10-11, when the hephaestus tracker was created with `sisyphus init --nested` in `hephaestus/`. `sisyphus lint` there reported 25 problems: every link to a sisyphus document or issue, for example `[[CONTRIBUTING#Changelog]]` and `[[sisyphus-claims]]`, was "missing". `sisyphus lint` at the root of the repo resolves the same links, because it walks the whole repo. The hephaestus files now use Markdown links and names in code text instead, which is a workaround.

Options: resolve links in the whole repo of the VCS (the vault of Obsidian), or in the directory of the nearest enclosing tracker too; keep the issues of each tracker separate.

## Acceptance criteria

- [ ] A link from a file of a nested tracker to a file of the enclosing repo resolves in `lint` and `links`, as in Obsidian.
- [ ] The issues of the nested tracker stay separate: `list` and the other issue commands do not see the issues of the enclosing tracker.
- [ ] A decision record states the scope.

## Out of scope

- One tracker for both projects.

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

Completed in [[0.0.107]]. Links of a nested tracker resolve against every file of the working copy that holds it, as in Obsidian; its issues stay separate, and its lint checks only its own documents. The fix found that [[agent-guide]] (then docs/agents.md) was hidden by AGENTS.md, so the guide has a new name.

