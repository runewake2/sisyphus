---
title: "Check the repo before the tool agents change lands"
state: closed          # open | in-progress | closed. Must match the directory of the file.
resolution: completed  # completed | abandoned. Set this only when state is closed.
priority: medium       # critical | high | medium | low
effort: small          # small | medium | large | x-large
tags: [hephaestus]     # The components that the work touches, for example [widget-scheduler, plan]
created: 2026-10-12
closed: 2026-10-12     # YYYY-MM-DD. Set this only when state is closed.
owner: repo-check-1    # The person or agent working on the issue. Cleared when the issue returns to open.
approver:              # The person or agent who accepts the issue when it closes.
bookmark:              # samw/ai/<workspace-name> of the agent that does the work. Set this when work starts.
workspaces: []         # jj workspaces where local work on the issue has happened, for example [sisyphus-move-commands].
metadata: {delegate: "repo-check"} # Optional. Arbitrary key-value notes, for example an AI agent session id to resume work with context: {session-id: "abc123"}.
deferred-from:         # Optional. The issue (as a quoted wikilink) or bookmark that deferred this work.
parent:                # Optional. The parent issue (as a quoted wikilink), if this issue is a sub-issue.
depends-on: []           # Optional. Issues (as quoted wikilinks) that must close before this one can start, for example `["[[faster-startup]]"]`.
remote:                # Optional. A URL: the GitHub issue or Jira ticket that tracks this issue outside the repo.
---

# Check the repo before the tool agents change lands

## Summary

Run the checks of the repo on the working copy of the tool agents change, with the tool role repo-check, as the first real task of a tool agent.

## Context

<Give the background that an agent needs to start the work without other help.
Link to design docs and other issues with wikilinks, for example `[[widget-scheduler]]`,
`[[widget-scheduler#Status API]]`, or `[[<issue-name>]]`. Do not use paths.>

## Acceptance criteria

- [ ] gofmt, go vet, and sisyphus lint are clean in both modules and both trackers.

## Out of scope

<List related work that this issue does not include. Create a separate issue for each item that someone must do later.>

## Notes

<Record progress, findings, and open questions here while the work continues.>

## Resolution

Done by the tool agent repo-check-1 (role `repo-check`): `cd "$(jj root)" && test -z "$(gofmt -l cmd hephaestus)" && go vet ./... && (cd hephaestus && go vet ./...) && sisyphus lint && sisyphus -C hephaestus lint && echo "gofmt, go vet, and sisyphus lint are clean in both modules and both trackers."`.

```
126 issues and 330 Markdown files: no problems.
14 issues and 24 Markdown files: no problems.
gofmt, go vet, and sisyphus lint are clean in both modules and both trackers.
```

