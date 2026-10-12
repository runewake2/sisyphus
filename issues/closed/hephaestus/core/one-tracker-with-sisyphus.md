---
title: "Share the issues and the design log of the repo with sisyphus"
state: closed          # open | in-progress | closed. Must match the directory of the file.
resolution: completed  # completed | abandoned. Set this only when state is closed.
priority: medium       # critical | high | medium | low
effort: small          # small | medium | large | x-large
tags: [hephaestus, issues] # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-12
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

# Share the issues and the design log of the repo with sisyphus

## Summary

hephaestus had an issues tracker and design documents of its own. Move them into the tracker and the design log of the repo, separated from sisyphus only by subdirectories.

## Context

Requested by a human on 2026-10-12: hephaestus and sisyphus should share a directory as well as issues, separated by subdirectories, not isolated by separate design and issue directories. See [[0002-one-tracker-and-one-design-log-for-both-projects]].

## Acceptance criteria

- [ ] The hephaestus issues are in issues/<state>/hephaestus/, and hephaestus/issues/ is gone.
- [ ] The hephaestus design documents and decision records are in design/hephaestus/ and design/decisions/hephaestus/.
- [ ] One sisyphus lint is clean, and CONTRIBUTING.md tells where hephaestus issues and design go.

## Out of scope

- The Go module and the role files of hephaestus: they stay in hephaestus/.

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

Completed in [[0.0.115]]. The hephaestus issues are in issues/<state>/hephaestus/, and its design documents and decision records in design/hephaestus/ and design/decisions/hephaestus/. One sisyphus lint, list, and graph covers both projects. See [[0002-one-tracker-and-one-design-log-for-both-projects]].

