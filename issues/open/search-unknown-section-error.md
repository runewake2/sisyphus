---
title: "Fail when search gets an unknown section"
state: open            # open | in-progress | closed. Must match the directory of the file.
resolution:            # completed | abandoned. Set this only when state is closed.
priority: low          # critical | high | medium | low
effort: small          # small | medium | large | x-large
tags: [bug, cli, search] # The components that the work touches, for example [widget-scheduler, plan]
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

# Fail when search gets an unknown section

## Summary

`sisyphus search --section nosuch foo` prints an empty line and exits 0. An unknown section is almost always a typo, so the user thinks that nothing matched.

## Context

Found in a review of sisyphus 0.0.31 on 2026-10-11. Compare `sisyphus edit <name> Nonexistent ...`, which fails with the list of sections that the issue has. `search` must do the same, with the section names of `issues/TEMPLATE.md`.

## Acceptance criteria

- [ ] `search --section` with a name that is not a section of the template fails and lists the valid section names.
- [ ] The match stays case-insensitive (`summary` is valid).
- [ ] A test covers the error.

## Out of scope

Nothing.

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

<Complete this section when you close the issue.
- Completed: tell what was done. Link to the changelog version and the decision records, for example `[[0.0.5]]`.
- Abandoned: tell why the work stopped. Link to the issue or decision that replaces it, if one exists.>
